package main

import (
	"strings"
	"testing"
)

// base returns a fixed-rate trial that passes every gate.
func base() *TrialResult {
	return &TrialResult{
		Spec:   TrialSpec{Workload: "fixed", Rate: 100, MaxVUs: 400, Profile: "none"},
		Status: "valid",
		Window: Window{Nominal: 10},
		K6:     map[string]float64{"reqs": 1000, "http_reqs_total": 1100, "vus_max": 20},
		Server: map[string]float64{"final_received": 1100},
		Host:   map[string]float64{"cpu_speed_limit_before": 100, "cpu_speed_limit_after": 100, "swapouts_delta": 0, "idle_cpu_pct_before": 1},
		Samples: map[string][]Sample{
			"k6": {{T: 0, CPUTime: 0}, {T: 5, CPUTime: 5}, {T: 10, CPUTime: 10}}, // 1 core
		},
	}
}

func TestGate(t *testing.T) {
	cases := []struct {
		name   string
		mut    func(r *TrialResult)
		status string
		reason string
	}{
		{"clean", func(*TrialResult) {}, "valid", ""},
		{"dropped", func(r *TrialResult) { r.K6["dropped_iterations"] = 1 }, "invalid", "dropped_iterations"},
		{"achieved just ok", func(r *TrialResult) { r.K6["reqs"] = 990 }, "valid", ""},
		{"achieved low", func(r *TrialResult) { r.K6["reqs"] = 989 }, "invalid", "achieved"},
		{"non202 at limit", func(r *TrialResult) { r.K6["non_202"] = 1 }, "valid", ""},
		{"non202 over", func(r *TrialResult) { r.K6["non_202"] = 2 }, "invalid", "non-202"},
		{"transport", func(r *TrialResult) { r.K6["transport_errors"] = 1 }, "invalid", "transport"},
		{"vus exhausted", func(r *TrialResult) { r.K6["vus_max"] = 400 }, "invalid", "VU pool"},
		{"k6 cpu avg", func(r *TrialResult) {
			r.Samples["k6"] = []Sample{{T: 0}, {T: 5, CPUTime: 20}, {T: 10, CPUTime: 40}} // 4 cores
		}, "invalid", "load generator CPU"},
		{"k6 cpu peak", func(r *TrialResult) {
			r.Samples["k6"] = []Sample{{T: 0}, {T: 5, CPUTime: 5}, {T: 10, CPUTime: 28}} // peak 4.6
		}, "invalid", "load generator CPU"},
		{"thermal invalid", func(r *TrialResult) { r.Host["cpu_speed_limit_after"] = 89 }, "invalid", "cpu_speed_limit_after"},
		{"thermal flag", func(r *TrialResult) { r.Host["cpu_speed_limit_before"] = 95 }, "flagged", "cpu_speed_limit_before"},
		{"thermal unknown", func(r *TrialResult) { r.Host["cpu_speed_limit_before"] = -1 }, "valid", ""},
		{"swap", func(r *TrialResult) { r.Host["swapouts_delta"] = 3 }, "invalid", "swapouts"},
		{"count mismatch", func(r *TrialResult) { r.Server["final_received"] = 1098 }, "invalid", "server received"},
		{"count within", func(r *TrialResult) { r.Server["final_received"] = 1099 }, "valid", ""},
		{"busy host", func(r *TrialResult) { r.Host["idle_cpu_pct_before"] = 11 }, "flagged", "host busy"},
		{"profiling", func(r *TrialResult) { r.Spec.Profile = "cpu" }, "flagged", "profiling active"},
		{"no reqs", func(r *TrialResult) { r.K6["reqs"] = 0 }, "invalid", "no measured requests"},
		{"invalid beats flag", func(r *TrialResult) { r.Spec.Profile = "cpu"; r.K6["dropped_iterations"] = 1 }, "invalid", "dropped"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := base()
			c.mut(r)
			gate(r)
			if r.Status != c.status {
				t.Fatalf("status %q, want %q (reasons %v)", r.Status, c.status, r.Reasons)
			}
			if c.reason != "" && !strings.Contains(strings.Join(r.Reasons, "|"), c.reason) {
				t.Fatalf("reasons %v lack %q", r.Reasons, c.reason)
			}
		})
	}
}

func TestVCSSetting(t *testing.T) {
	m := "bin: go1.27.4\n\tpath\tx\n\tbuild\tvcs.revision=abc123\n\tbuild\tvcs.modified=true\n"
	if got := vcsSetting(m, "vcs.revision"); got != "abc123" {
		t.Fatalf("revision %q", got)
	}
	if got := vcsSetting(m, "vcs.modified"); got != "true" {
		t.Fatalf("modified %q", got)
	}
	if got := vcsSetting(m, "vcs.time"); got != "" {
		t.Fatalf("missing key %q", got)
	}
}
