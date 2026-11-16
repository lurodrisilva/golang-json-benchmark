// Command report turns one benchctl run directory into a comparative report:
// report.md, report.html (inline SVG charts), trials.csv, summary.json and,
// when the exact binaries are present, pprof top-N text summaries.
//
// Statistical unit: one trial (one server process lifetime). Per-trial
// statistics (e.g. that trial's P99 from k6) are never averaged and presented
// as pooled percentiles; cell-level values are the MEDIAN OF PER-TRIAL values,
// with a seeded percentile-bootstrap 95% CI on the reduction %.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"math"
	"math/rand/v2"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type trial struct {
	Spec struct {
		Arm      string `json:"arm"`
		Fixture  string `json:"fixture"`
		Workload string `json:"workload"`
		Rate     int    `json:"rate"`
		Knee     int    `json:"knee"`
		Steps    string `json:"steps"`
		StepDur  string `json:"step_dur"`
		Rep      int    `json:"rep"`
		Profile  string `json:"profile"`
		Variant  string `json:"variant"`
		Duration string `json:"duration"`
	} `json:"spec"`
	Status   string             `json:"status"`
	Reasons  []string           `json:"reasons"`
	Window   map[string]float64 `json:"window"`
	Server   map[string]float64 `json:"server"`
	K6       map[string]float64 `json:"k6"`
	Host     map[string]float64 `json:"host"`
	Profiles []struct {
		Kind      string  `json:"kind"`
		File      string  `json:"file"`
		Seconds   int     `json:"seconds"`
		DecodedOK float64 `json:"decoded_ok_in_window"`
	} `json:"profiles"`
	dir string
}

// pair maps an arm to its pair and role. Baseline is always the stdlib arm.
func pair(arm string) (name string, baseline bool) {
	switch arm {
	case "stdlib":
		return "D", true
	case "goccy":
		return "D", false
	case "stdlib-legacy":
		return "L", true
	case "goccy-legacy":
		return "L", false
	}
	return "?", false
}

type cellKey struct {
	Pair, Fixture, Workload, Point, Variant, Profile string
}

func (k cellKey) String() string {
	s := fmt.Sprintf("Pair %s · %s · %s %s", k.Pair, k.Fixture, k.Workload, k.Point)
	if k.Variant != "" {
		s += " · variant=" + k.Variant
	}
	if k.Profile != "none" {
		s += " · profile=" + k.Profile
	}
	return s
}

func keyOf(t *trial) cellKey {
	p, _ := pair(t.Spec.Arm)
	point := ""
	switch t.Spec.Workload {
	case "fixed":
		point = fmt.Sprintf("@%d rps", t.Spec.Rate)
	case "ramp":
		point = fmt.Sprintf("knee-ref %d", t.Spec.Knee)
	}
	return cellKey{p, t.Spec.Fixture, t.Spec.Workload, point, t.Spec.Variant, t.Spec.Profile}
}

// metric describes one reported value. lower=true → lower is better.
type metric struct {
	Name, Unit, Src string
	Lower           bool
	Minimum         float64 // P99 needs a minimum per-trial sample count
	Control         bool    // shared code in both arms: negative control, never a library verdict
	get             func(*trial) (float64, bool)
}

func k6(name string) func(*trial) (float64, bool) {
	return func(t *trial) (float64, bool) { v, ok := t.K6[name]; return v, ok }
}
func srv(name string) func(*trial) (float64, bool) {
	return func(t *trial) (float64, bool) { v, ok := t.Server[name]; return v, ok }
}

const p99MinSamples = 10000 // plan §14: P99 only reported with ≥10,000 requests per trial

var fixedMetrics = func() []metric {
	ms := []metric{}
	for _, tr := range []struct {
		id, label string
		control   bool
	}{
		{"http_req_duration", "HTTP request duration (k6)", false},
		{"srv_hdl", "Handler duration (server)", false},
		{"srv_dec", "Decode duration (Unmarshal only)", false},
		{"srv_read", "Body read duration (shared code, control)", true},
		{"srv_pol", "Policy scan duration (shared code, control)", true},
		{"srv_val", "Validate duration (shared code, control)", true},
	} {
		for _, st := range []string{"min", "med", "avg", "p(95)", "p(99)", "max"} {
			m := metric{Name: tr.label + " " + st, Unit: "ms", Src: "k6:" + tr.id + "{phase:measure}." + st, Lower: true, Control: tr.control, get: k6(tr.id + "." + st)}
			if st == "p(99)" {
				m.Minimum = p99MinSamples
				id := tr.id
				m.get = func(t *trial) (float64, bool) {
					if t.K6[id+".count"] < p99MinSamples {
						return 0, false
					}
					v, ok := t.K6[id+".p(99)"]
					return v, ok
				}
			}
			ms = append(ms, m)
		}
	}
	ms = append(ms,
		metric{Name: "Achieved throughput", Unit: "req/s", Src: "k6 measured reqs / nominal window", Lower: false, get: k6("achieved_rps")},
		metric{Name: "Server decoded_ok rate", Unit: "op/s", Src: "Δdecoded_ok / window wall", Lower: false, get: srv("decoded_ok_per_s")},
		metric{Name: "CPU time per successful op", Unit: "µs/op", Src: "getrusage Δ(utime+stime)/Δdecoded_ok", Lower: true, get: srv("cpu_us_per_op")},
		metric{Name: "CPU utilization", Unit: "cores", Src: "getrusage Δ/window wall", Lower: true, get: srv("cpu_util_cores")},
		metric{Name: "CPU-seconds in window", Unit: "s", Src: "getrusage Δ", Lower: true, get: srv("cpu_seconds")},
		metric{Name: "Σ handler durations", Unit: "s", Src: "Σ sum_handler_ns Δ", Lower: true, get: scale(srv("d_sum_handler_ns"), 1e-9)},
		metric{Name: "Σ decode durations", Unit: "s", Src: "Σ sum_decode_ns Δ", Lower: true, get: scale(srv("d_sum_decode_ns"), 1e-9)},
		metric{Name: "Allocated bytes per op", Unit: "B/op", Src: "runtime/metrics /gc/heap/allocs:bytes Δ / Δdecoded_ok", Lower: true, get: srv("alloc_bytes_per_op")},
		metric{Name: "Allocated objects per op", Unit: "obj/op", Src: "runtime/metrics /gc/heap/allocs:objects Δ / Δdecoded_ok", Lower: true, get: srv("alloc_objects_per_op")},
		metric{Name: "GC cycles per 1k ops", Unit: "cycles", Src: "/gc/cycles/total Δ", Lower: true, get: srv("gc_cycles_per_kop")},
		metric{Name: "GC pause total in window", Unit: "s", Src: "/sched/pauses/total/gc histogram Δ (bucket lower bounds)", Lower: true, get: srv("gc_pause_total_s_window")},
		metric{Name: "Live heap at window end", Unit: "MiB", Src: "/gc/heap/live:bytes", Lower: true, get: scale(srv("end_rt/gc/heap/live:bytes"), 1.0/(1<<20))},
		metric{Name: "Runtime-mapped memory at end", Unit: "MiB", Src: "/memory/classes/total:bytes", Lower: true, get: scale(srv("end_rt/memory/classes/total:bytes"), 1.0/(1<<20))},
		metric{Name: "Peak RSS (ps, 5 s samples)", Unit: "MiB", Src: "ps -o rss", Lower: true, get: scale(srv("rss_peak_bytes_ps"), 1.0/(1<<20))},
		metric{Name: "Max RSS (getrusage)", Unit: "MiB", Src: "ru_maxrss", Lower: true, get: scale(srv("end_rusage_maxrss_bytes"), 1.0/(1<<20))},
	)
	return ms
}()

func scale(f func(*trial) (float64, bool), k float64) func(*trial) (float64, bool) {
	return func(t *trial) (float64, bool) { v, ok := f(t); return v * k, ok }
}

type comparison struct {
	Metric        string    `json:"metric"`
	Unit          string    `json:"unit"`
	Source        string    `json:"source"`
	LowerIsBetter bool      `json:"lower_is_better"`
	Baseline      []float64 `json:"baseline_trials"`
	Candidate     []float64 `json:"candidate_trials"`
	BaseMedian    float64   `json:"baseline_median"`
	CandMedian    float64   `json:"candidate_median"`
	AbsDiff       float64   `json:"abs_diff"`       // candidate − baseline
	Reduction     *float64  `json:"reduction_pct"`  // (stdlib − goccy)/stdlib·100; nil if baseline 0
	CILow         *float64  `json:"ci95_low_pct"`   // bootstrap
	CIHigh        *float64  `json:"ci95_high_pct"`  //
	BaseCV        float64   `json:"baseline_cv"`    // sd/mean across trials
	CandCV        float64   `json:"candidate_cv"`   //
	Verdict       string    `json:"verdict"`        // see verdict()
	Note          string    `json:"note,omitempty"` //
	control       bool
}

type cell struct {
	Key       cellKey           `json:"-"`
	Name      string            `json:"cell"`
	Base      []*trial          `json:"-"`
	Cand      []*trial          `json:"-"`
	Excluded  []string          `json:"excluded_trials"`
	Compares  []comparison      `json:"comparisons"`
	Steps     map[string][]step `json:"steps,omitempty"` // arm → steps (ramp)
	Knee      map[string]string `json:"knee,omitempty"`
	Profiling []profileRow      `json:"profiles,omitempty"`
}

type step struct {
	Index    int     `json:"index"`
	Offered  float64 `json:"offered_rps"`
	Achieved float64 `json:"achieved_rps"`
	Dropped  float64 `json:"dropped"`
	Non202   float64 `json:"non_202"`
	P50      float64 `json:"p50_ms"`
	P95      float64 `json:"p95_ms"`
	P99      float64 `json:"p99_ms"`
	DecMed   float64 `json:"srv_dec_p50_ms"`
	Count    float64 `json:"count"`
}

type profileRow struct {
	Arm, Kind, File string
	Rep             int
	DecodedOK       float64
	Top             string
}

func main() {
	run := flag.String("run", "", "benchctl run directory (required)")
	out := flag.String("out", "", "output directory (default <run>/report)")
	bench := flag.String("bench", "", "benchstat output to embed (optional)")
	conf := flag.String("conformance", "", "directory with conformance-*.tsv matrices (optional)")
	threshold := flag.Float64("threshold", 10, "practical-significance threshold, percent (user-approved: 10)")
	seed := flag.Uint64("seed", 20261115, "bootstrap seed")
	resamples := flag.Int("resamples", 10000, "bootstrap resamples")
	pprof := flag.Bool("pprof", true, "render `go tool pprof -top` summaries for profile files")
	flag.Parse()
	if *run == "" {
		fmt.Fprintln(os.Stderr, "report: -run is required")
		os.Exit(2)
	}
	if *out == "" {
		*out = filepath.Join(*run, "report")
	}
	if err := os.MkdirAll(*out, 0o755); err != nil {
		fail(err)
	}
	trials, err := load(*run)
	if err != nil {
		fail(err)
	}
	if len(trials) == 0 {
		fail(fmt.Errorf("no trial.json under %s", *run))
	}
	rng := rand.New(rand.NewPCG(*seed, *seed^0x2545f4914f6cdd1d))
	cells := group(trials)
	for _, c := range cells {
		analyze(c, rng, *resamples, *threshold)
		if *pprof {
			profiles(c, *run, *out)
		}
	}
	if err := writeCSV(filepath.Join(*out, "trials.csv"), trials); err != nil {
		fail(err)
	}
	if err := writeJSON(filepath.Join(*out, "summary.json"), map[string]any{
		"run": filepath.Base(*run), "threshold_pct": *threshold, "bootstrap": map[string]any{"seed": *seed, "resamples": *resamples},
		"cells": cells,
	}); err != nil {
		fail(err)
	}
	ctx := reportCtx{Run: *run, Cells: cells, Trials: trials, Threshold: *threshold, Resamples: *resamples, Seed: *seed}
	ctx.Env, _ = os.ReadFile(filepath.Join(*run, "env.json"))
	ctx.Build, _ = os.ReadFile(filepath.Join(*run, "build.json"))
	ctx.Matrix, _ = os.ReadFile(filepath.Join(*run, "matrix.json"))
	ctx.Fixtures, _ = os.ReadFile(filepath.Join(*run, "fixtures.SHA256SUMS"))
	if *bench != "" {
		ctx.Bench, _ = os.ReadFile(*bench)
	}
	if *conf != "" {
		ctx.Conformance = loadConformance(*conf)
	}
	md := renderMarkdown(&ctx)
	if err := os.WriteFile(filepath.Join(*out, "report.md"), []byte(md), 0o644); err != nil {
		fail(err)
	}
	if err := os.WriteFile(filepath.Join(*out, "report.html"), []byte(renderHTML(&ctx, md)), 0o644); err != nil {
		fail(err)
	}
	fmt.Println(filepath.Join(*out, "report.md"))
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "report:", err)
	os.Exit(1)
}

func load(run string) ([]*trial, error) {
	var out []*trial
	err := filepath.WalkDir(filepath.Join(run, "trials"), func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.Name() != "trial.json" {
			return nil
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		t := &trial{dir: filepath.Dir(p)}
		if err := json.Unmarshal(b, t); err != nil {
			return fmt.Errorf("%s: %w", p, err)
		}
		out = append(out, t)
		return nil
	})
	sort.Slice(out, func(i, j int) bool { return out[i].dir < out[j].dir })
	return out, err
}

func group(ts []*trial) []*cell {
	m := map[cellKey]*cell{}
	var keys []cellKey
	for _, t := range ts {
		if t.Spec.Workload == "smoke" {
			continue
		}
		k := keyOf(t)
		c, ok := m[k]
		if !ok {
			c = &cell{Key: k, Name: k.String()}
			m[k] = c
			keys = append(keys, k)
		}
		// valid and flagged trials are analysed (flag reasons are listed);
		// invalid/error trials are excluded but always reported.
		if t.Status != "valid" && t.Status != "flagged" {
			c.Excluded = append(c.Excluded, fmt.Sprintf("%s rep %d: %s %v", t.Spec.Arm, t.Spec.Rep, t.Status, t.Reasons))
			continue
		}
		if t.Status == "flagged" && !onlyProfilingFlag(t) {
			c.Excluded = append(c.Excluded, fmt.Sprintf("%s rep %d: INCLUDED but flagged %v", t.Spec.Arm, t.Spec.Rep, t.Reasons))
		}
		if _, base := pair(t.Spec.Arm); base {
			c.Base = append(c.Base, t)
		} else {
			c.Cand = append(c.Cand, t)
		}
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i].String() < keys[j].String() })
	out := make([]*cell, 0, len(keys))
	for _, k := range keys {
		out = append(out, m[k])
	}
	return out
}

func onlyProfilingFlag(t *trial) bool {
	for _, r := range t.Reasons {
		if !strings.HasPrefix(r, "profiling active") {
			return false
		}
	}
	return true
}

func analyze(c *cell, rng *rand.Rand, resamples int, threshold float64) {
	switch c.Key.Workload {
	case "fixed":
		for _, m := range fixedMetrics {
			c.Compares = append(c.Compares, compare(m, c.Base, c.Cand, rng, resamples, threshold))
		}
		applyControls(c.Compares, threshold)
	case "ramp":
		c.Steps, c.Knee = map[string][]step{}, map[string]string{}
		for _, t := range append(append([]*trial{}, c.Base...), c.Cand...) {
			ss := steps(t)
			label := fmt.Sprintf("%s rep %d", t.Spec.Arm, t.Spec.Rep)
			c.Steps[label] = ss
			c.Knee[label] = knee(ss)
		}
	}
}

func compare(m metric, base, cand []*trial, rng *rand.Rand, resamples int, threshold float64) comparison {
	cmp := comparison{Metric: m.Name, Unit: m.Unit, Source: m.Src, LowerIsBetter: m.Lower, control: m.Control}
	for _, t := range base {
		if v, ok := m.get(t); ok {
			cmp.Baseline = append(cmp.Baseline, v)
		}
	}
	for _, t := range cand {
		if v, ok := m.get(t); ok {
			cmp.Candidate = append(cmp.Candidate, v)
		}
	}
	if len(cmp.Baseline) == 0 || len(cmp.Candidate) == 0 {
		cmp.Verdict = "not available"
		if m.Minimum > 0 {
			cmp.Note = fmt.Sprintf("requires ≥%.0f samples per trial", m.Minimum)
		}
		return cmp
	}
	cmp.BaseMedian, cmp.CandMedian = median(cmp.Baseline), median(cmp.Candidate)
	cmp.AbsDiff = cmp.CandMedian - cmp.BaseMedian
	cmp.BaseCV, cmp.CandCV = cv(cmp.Baseline), cv(cmp.Candidate)
	red := func(b, g float64) (float64, bool) {
		if b == 0 {
			return 0, false
		}
		if m.Lower {
			return (b - g) / b * 100, true
		}
		return (g - b) / b * 100, true // higher-is-better: improvement % of goccy over stdlib
	}
	r, ok := red(cmp.BaseMedian, cmp.CandMedian)
	if !ok {
		cmp.Verdict = "undefined (baseline is zero)"
		return cmp
	}
	cmp.Reduction = &r
	if len(cmp.Baseline) < 2 || len(cmp.Candidate) < 2 {
		cmp.Verdict = "insufficient trials (n<2)"
		return cmp
	}
	boot := make([]float64, 0, resamples)
	bs, cs := make([]float64, len(cmp.Baseline)), make([]float64, len(cmp.Candidate))
	for range resamples {
		for i := range bs {
			bs[i] = cmp.Baseline[rng.IntN(len(cmp.Baseline))]
		}
		for i := range cs {
			cs[i] = cmp.Candidate[rng.IntN(len(cmp.Candidate))]
		}
		if v, ok := red(median(bs), median(cs)); ok {
			boot = append(boot, v)
		}
	}
	sort.Float64s(boot)
	lo, hi := quantile(boot, 0.025), quantile(boot, 0.975)
	cmp.CILow, cmp.CIHigh = &lo, &hi
	cmp.Verdict = verdict(lo, hi, r, threshold, max(cmp.BaseCV, cmp.CandCV), min(len(cmp.Baseline), len(cmp.Candidate)))
	return cmp
}

const verdictControl = "control (shared code)"

// applyControls: shared-code metrics (read, policy, validate) must not differ
// between arms. If any control's median differs by ≥ threshold, the cell's
// environment is not stable enough to attribute differences to the library,
// so every library verdict that names a winner or a detectable difference is
// downgraded to inconclusive.
func applyControls(cs []comparison, threshold float64) {
	var failed []string
	for i := range cs {
		if !cs[i].control {
			continue
		}
		cs[i].Verdict = verdictControl
		if strings.HasSuffix(cs[i].Metric, " med") && cs[i].Reduction != nil && math.Abs(*cs[i].Reduction) >= threshold {
			failed = append(failed, fmt.Sprintf("%s %+.1f%%", cs[i].Metric, *cs[i].Reduction))
			cs[i].Note = "control differs ≥ threshold"
		}
	}
	if len(failed) == 0 {
		return
	}
	for i := range cs {
		if cs[i].control {
			continue
		}
		v := cs[i].Verdict
		if strings.HasPrefix(v, "goccy better") || strings.HasPrefix(v, "stdlib better") || strings.HasPrefix(v, "detectable") {
			cs[i].Verdict = "inconclusive (shared-code control differs ≥ threshold)"
			cs[i].Note = "was: " + v + "; controls: " + strings.Join(failed, ", ")
		}
	}
}

// verdict never names a winner on noisy, small or sub-threshold evidence.
func verdict(lo, hi, point, threshold, worstCV float64, n int) string {
	switch {
	case n < 5:
		return "exploratory (n<5 trials per arm)"
	case worstCV > 0.10:
		return "inconclusive (trial CV > 10%)"
	case lo > 0 && point >= threshold:
		return "goccy better (CI excludes 0, ≥ threshold)"
	case hi < 0 && point <= -threshold:
		return "stdlib better (CI excludes 0, ≥ threshold)"
	case lo > 0 || hi < 0:
		return "detectable but below threshold"
	default:
		return "no detectable difference"
	}
}

func steps(t *trial) []step {
	pcts := strings.Split(t.Spec.Steps, ",")
	var out []step
	for i, p := range pcts {
		pct, _ := strconv.Atoi(strings.TrimSpace(p))
		sel := fmt.Sprintf("{step:%d}", i)
		dur := parseSeconds(t.Spec.StepDur)
		s := step{Index: i, Offered: float64(t.Spec.Knee) * float64(pct) / 100,
			Count:   t.K6["http_reqs"+sel+".count"],
			Dropped: t.K6[fmt.Sprintf("step_%02d_dropped", i)],
			Non202:  t.K6["non_202"+sel+".count"],
			P50:     t.K6["http_req_duration"+sel+".med"],
			P95:     t.K6["http_req_duration"+sel+".p(95)"],
			P99:     t.K6["http_req_duration"+sel+".p(99)"],
			DecMed:  t.K6["srv_dec"+sel+".med"],
		}
		if dur > 0 {
			s.Achieved = s.Count / dur
		}
		out = append(out, s)
	}
	return out
}

// knee: highest step with achieved ≥ 99% of offered, no drops, no non-202,
// and P95 ≤ 3× the lowest step's P95. P95 replaces the plan's P99 criterion
// (§14.4) because low steps of large fixtures never reach the 10,000-sample
// minimum required for P99 (plan §14); P95 needs ≥ kneeMinSamples per step.
const kneeMinSamples = 200

func knee(ss []step) string {
	if len(ss) == 0 {
		return "n/a"
	}
	if ss[0].Count < kneeMinSamples {
		return fmt.Sprintf("undetermined (lowest step < %d samples)", kneeMinSamples)
	}
	ref := ss[0].P95
	best := ""
	for _, s := range ss {
		ok := s.Offered > 0 && s.Count >= kneeMinSamples && s.Achieved >= 0.99*s.Offered && s.Dropped == 0 && s.Non202 == 0 && (ref == 0 || s.P95 <= 3*ref)
		if !ok {
			break
		}
		best = fmt.Sprintf("%.0f rps (step %d)", s.Offered, s.Index)
	}
	if best == "" {
		return "below lowest step"
	}
	return best
}

func profiles(c *cell, run, out string) {
	for _, t := range append(append([]*trial{}, c.Base...), c.Cand...) {
		for _, p := range t.Profiles {
			row := profileRow{Arm: t.Spec.Arm, Kind: p.Kind, File: filepath.Join(t.dir, p.File), Rep: t.Spec.Rep, DecodedOK: p.DecodedOK}
			bin := filepath.Join(run, "bin", binName(t.Spec.Arm, t.Spec.Variant))
			args := []string{"tool", "pprof", "-top", "-nodecount=20"}
			if p.Kind == "allocs" {
				args = append(args, "-sample_index=alloc_space")
			}
			if _, err := os.Stat(bin); err == nil {
				args = append(args, bin)
			}
			args = append(args, row.File)
			if b, err := exec.Command("go", args...).Output(); err == nil {
				row.Top = string(b)
				name := fmt.Sprintf("pprof-%s-%s-%s-rep%02d-%s.txt", c.Key.Pair, t.Spec.Fixture, t.Spec.Arm, t.Spec.Rep, p.Kind)
				os.WriteFile(filepath.Join(out, name), b, 0o644)
			} else {
				row.Top = "go tool pprof failed: " + err.Error()
			}
			c.Profiling = append(c.Profiling, row)
		}
		if heap := filepath.Join(t.dir, "heap-end.pprof"); fileExists(heap) {
			for _, idx := range []string{"inuse_space", "inuse_objects"} {
				bin := filepath.Join(run, "bin", binName(t.Spec.Arm, t.Spec.Variant))
				args := []string{"tool", "pprof", "-top", "-nodecount=15", "-sample_index=" + idx}
				if fileExists(bin) {
					args = append(args, bin)
				}
				if b, err := exec.Command("go", append(args, heap)...).Output(); err == nil {
					name := fmt.Sprintf("pprof-%s-%s-%s-rep%02d-heap-%s.txt", c.Key.Pair, t.Spec.Fixture, t.Spec.Arm, t.Spec.Rep, idx)
					os.WriteFile(filepath.Join(out, name), b, 0o644)
				}
			}
		}
	}
}

func binName(arm, variant string) string {
	if variant != "" {
		return arm + "-" + variant + "-server"
	}
	return arm + "-server"
}

func fileExists(p string) bool { _, err := os.Stat(p); return err == nil }

func parseSeconds(d string) float64 {
	if d == "" {
		return 0
	}
	n := len(d)
	mult := 1.0
	switch {
	case strings.HasSuffix(d, "ms"):
		mult, n = 1e-3, n-2
	case strings.HasSuffix(d, "s"):
		n--
	case strings.HasSuffix(d, "m"):
		mult, n = 60, n-1
	case strings.HasSuffix(d, "h"):
		mult, n = 3600, n-1
	}
	v, err := strconv.ParseFloat(d[:n], 64)
	if err != nil {
		return 0
	}
	return v * mult
}

func median(xs []float64) float64 {
	s := append([]float64(nil), xs...)
	sort.Float64s(s)
	n := len(s)
	if n == 0 {
		return math.NaN()
	}
	if n%2 == 1 {
		return s[n/2]
	}
	return (s[n/2-1] + s[n/2]) / 2
}

func quantile(sorted []float64, q float64) float64 {
	if len(sorted) == 0 {
		return math.NaN()
	}
	pos := q * float64(len(sorted)-1)
	lo := int(math.Floor(pos))
	hi := int(math.Ceil(pos))
	return sorted[lo] + (sorted[hi]-sorted[lo])*(pos-float64(lo))
}

func cv(xs []float64) float64 {
	if len(xs) < 2 {
		return 0
	}
	var mean float64
	for _, x := range xs {
		mean += x
	}
	mean /= float64(len(xs))
	if mean == 0 {
		return 0
	}
	var ss float64
	for _, x := range xs {
		ss += (x - mean) * (x - mean)
	}
	return math.Sqrt(ss/float64(len(xs)-1)) / math.Abs(mean)
}

func writeJSON(path string, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(b, '\n'), 0o644)
}

func loadConformance(dir string) map[string][]byte {
	out := map[string][]byte{}
	ms, _ := filepath.Glob(filepath.Join(dir, "conformance-*.tsv"))
	for _, m := range ms {
		if b, err := os.ReadFile(m); err == nil {
			out[filepath.Base(m)] = b
		}
	}
	return out
}
