package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

// Validity thresholds (plan §14.4, user-approved Q5). Changing any of these
// changes which trials count; they are recorded per trial via the binary.
const (
	gateAchievedMin     = 0.99  // achieved/offered
	gateNon202Max       = 0.001 // fraction of measured requests
	gateK6AvgCoresMax   = 3.5   // mean k6 CPU over the trial
	gateK6PeakCoresMax  = 4.5   // any 5 s sample
	gateSpeedLimitBad   = 90    // CPU_Speed_Limit < 90 → invalid
	gateSpeedLimitFlag  = 100   // 90..99 → flagged
	gateCountDiffMax    = 0.001 // |k6 reqs − server received| / received
	gateIdleBusyFlagPct = 10.0  // host busy before trial (% of capacity)
)

type k6Metric map[string]float64

func readK6Summary(r *TrialResult, path string, s TrialSpec) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("k6 summary: %w", err)
	}
	var doc struct {
		Metrics map[string]map[string]any `json:"metrics"`
	}
	if err := json.Unmarshal(b, &doc); err != nil {
		return fmt.Errorf("k6 summary: %w", err)
	}
	get := func(name, field string) float64 {
		m, ok := doc.Metrics[name]
		if !ok {
			return 0
		}
		v, _ := m[field].(float64)
		return v
	}
	r.K6["http_reqs_total"] = get("http_reqs", "count")
	r.K6["vus_max"] = get("vus_max", "value")
	sel := "{phase:measure}"
	r.K6["reqs"] = get("http_reqs"+sel, "count")
	r.K6["non_202"] = get("non_202"+sel, "count")
	if s.encode() {
		// "non_202" is the generic failed-response counter: for encode it is
		// status != 200 plus 200s whose length differs from encode_bytes.
		r.K6["non_200"] = get("non_200"+sel, "count")
		r.K6["bad_length"] = get("bad_length"+sel, "count")
		r.K6["non_202"] = r.K6["non_200"] + r.K6["bad_length"]
		r.K6["data_received"] = get("data_received"+sel, "count")
	}
	r.K6["transport_errors"] = get("transport_errors"+sel, "count")
	r.K6["dropped_iterations"] = get("dropped_iterations{scenario:measure}", "count")
	if s.Workload == "ramp" {
		r.K6["dropped_iterations"] = 0
		for name := range doc.Metrics {
			if strings.HasPrefix(name, "dropped_iterations{scenario:step_") {
				step := strings.TrimSuffix(strings.TrimPrefix(name, "dropped_iterations{scenario:"), "}")
				r.K6[step+"_dropped"] = get(name, "count")
			}
		}
	}
	trends := []string{"http_req_duration", "srv_read", "srv_pol", "srv_dec", "srv_val", "srv_hdl"}
	if s.encode() {
		trends = []string{"http_req_duration", "http_req_waiting", "http_req_receiving", "srv_enc", "srv_ehdl"}
	}
	for _, t := range trends {
		for _, f := range []string{"count", "min", "med", "avg", "p(90)", "p(95)", "p(99)", "max"} {
			if v := get(t+sel, f); v != 0 || f == "min" {
				r.K6[t+"."+f] = v
			}
		}
	}
	if s.Workload == "ramp" {
		names := make([]string, 0)
		for name := range doc.Metrics {
			if strings.HasPrefix(name, "http_reqs{step:") || strings.HasPrefix(name, "http_req_duration{step:") ||
				strings.HasPrefix(name, "srv_dec{step:") || strings.HasPrefix(name, "non_202{step:") ||
				strings.HasPrefix(name, "srv_enc{step:") || strings.HasPrefix(name, "non_200{step:") ||
				strings.HasPrefix(name, "bad_length{step:") {
				names = append(names, name)
			}
		}
		sort.Strings(names)
		for _, name := range names {
			for f, v := range doc.Metrics[name] {
				if fv, ok := v.(float64); ok {
					r.K6[name+"."+f] = fv
				}
			}
		}
	}
	return nil
}

func (r *TrialResult) invalid(format string, a ...any) {
	r.Status = "invalid"
	r.Reasons = append(r.Reasons, fmt.Sprintf(format, a...))
}

func (r *TrialResult) flag(format string, a ...any) {
	if r.Status == "valid" {
		r.Status = "flagged"
	}
	r.Reasons = append(r.Reasons, fmt.Sprintf(format, a...))
}

// gate applies the validity rules. A trial is never silently dropped: its
// status and reasons are recorded and the report decides inclusion.
func gate(r *TrialResult) {
	s := r.Spec
	reqs := r.K6["reqs"]
	if reqs == 0 {
		r.invalid("no measured requests")
		return
	}
	if d := r.K6["dropped_iterations"]; d > 0 {
		r.invalid("dropped_iterations=%.0f", d)
	}
	if s.Workload == "fixed" {
		offered := float64(s.Rate) * r.Window.Nominal
		if ach := reqs / offered; ach < gateAchievedMin {
			r.invalid("achieved %.4f of offered (< %.2f)", ach, gateAchievedMin)
		}
		r.K6["offered_rps"] = float64(s.Rate)
		r.K6["achieved_rps"] = reqs / r.Window.Nominal
	}
	if s.Workload != "ramp" {
		if f := r.K6["non_202"] / reqs; f > gateNon202Max {
			r.invalid("non-202 fraction %.5f > %.3f", f, gateNon202Max)
		}
	}
	if t := r.K6["transport_errors"]; t > 0 && s.Workload != "ramp" {
		r.invalid("transport errors/timeouts=%.0f", t)
	}
	// VU exhaustion: k6's vus_max is the total across all scenarios (warm-up,
	// measure, every ramp step each preallocate their own pool), so it cannot
	// be compared with the per-scenario MaxVUs cap. A scenario that runs out
	// of VUs emits dropped_iterations, which is what the gates above (fixed)
	// and the knee rule (ramp, per step) already act on. vus_max is recorded.
	avg, peak := cores(r.Samples["k6"])
	r.Host["k6_cpu_avg_cores"], r.Host["k6_cpu_peak_cores"] = avg, peak
	if avg > gateK6AvgCoresMax || peak > gateK6PeakCoresMax {
		r.invalid("load generator CPU avg=%.2f peak=%.2f cores", avg, peak)
	}
	srvAvg, _ := cores(r.Samples["server"])
	r.Host["server_cpu_avg_cores_ps"] = srvAvg
	for _, k := range []string{"cpu_speed_limit_before", "cpu_speed_limit_after"} {
		v := r.Host[k]
		switch {
		case v < 0:
		case v < gateSpeedLimitBad:
			r.invalid("%s=%.0f", k, v)
		case v < gateSpeedLimitFlag:
			r.flag("%s=%.0f", k, v)
		}
	}
	if v, ok := r.Host["swapouts_delta"]; ok && v > 0 {
		r.invalid("swapouts during trial=%.0f", v)
	}
	if rec := r.Server["final_received"]; rec > 0 {
		if d := abs(r.K6["http_reqs_total"]-rec) / rec; d > gateCountDiffMax {
			r.invalid("k6 reqs %.0f vs server received %.0f (diff %.4f)", r.K6["http_reqs_total"], rec, d)
		}
	}
	if v := r.Host["idle_cpu_pct_before"]; v > gateIdleBusyFlagPct {
		r.flag("host busy %.1f%% before trial (top: %s)", v, strings.Join(r.HostTopBefore, ", "))
	}
	if r.Spec.Profile != "none" {
		r.flag("profiling active: not a primary latency trial")
	}
	maxRSS := 0.0
	for _, smp := range r.Samples["server"] {
		maxRSS = max(maxRSS, float64(smp.RSSKiB)*1024)
	}
	r.Server["rss_peak_bytes_ps"] = maxRSS
}

// cores converts cumulative ps CPU time into per-interval core usage.
func cores(ss []Sample) (avg, peak float64) {
	if len(ss) < 2 {
		return 0, 0
	}
	for i := 1; i < len(ss); i++ {
		dt := ss[i].T - ss[i-1].T
		if dt <= 0 {
			continue
		}
		peak = max(peak, (ss[i].CPUTime-ss[i-1].CPUTime)/dt)
	}
	span := ss[len(ss)-1].T - ss[0].T
	if span > 0 {
		avg = (ss[len(ss)-1].CPUTime - ss[0].CPUTime) / span
	}
	return avg, peak
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
