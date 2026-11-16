package main

import (
	"math/rand/v2"
	"strings"
	"testing"
)

func trials(vals ...float64) []*trial {
	out := make([]*trial, len(vals))
	for i, v := range vals {
		out[i] = &trial{K6: map[string]float64{"x": v}}
	}
	return out
}

func lowerM() metric  { return metric{Name: "x med", Lower: true, get: k6("x")} }
func higherM() metric { return metric{Name: "tput", Lower: false, get: k6("x")} }
func rng() *rand.Rand { return rand.New(rand.NewPCG(1, 2)) }

func TestReductionSign(t *testing.T) {
	// lower is better: stdlib 100, goccy 80 → +20 %.
	c := compare(lowerM(), trials(100, 100, 100), trials(80, 80, 80), rng(), 200, 10)
	if c.Reduction == nil || *c.Reduction != 20 {
		t.Fatalf("lower reduction = %v", c.Reduction)
	}
	// higher is better: stdlib 100, goccy 120 → +20 % (improvement of goccy).
	c = compare(higherM(), trials(100, 100, 100), trials(120, 120, 120), rng(), 200, 10)
	if c.Reduction == nil || *c.Reduction != 20 {
		t.Fatalf("higher improvement = %v", c.Reduction)
	}
}

func TestZeroBaseline(t *testing.T) {
	c := compare(lowerM(), trials(0, 0, 0, 0, 0), trials(1, 1, 1, 1, 1), rng(), 200, 10)
	if c.Reduction != nil || !strings.HasPrefix(c.Verdict, "undefined") {
		t.Fatalf("zero baseline: red=%v verdict=%q", c.Reduction, c.Verdict)
	}
}

func TestMissingMetric(t *testing.T) {
	c := compare(metric{Name: "y", get: k6("absent")}, trials(1), trials(1), rng(), 10, 10)
	if c.Verdict != "not available" {
		t.Fatalf("verdict %q", c.Verdict)
	}
}

func TestVerdictBranches(t *testing.T) {
	cases := []struct {
		lo, hi, pt, cv float64
		n              int
		want           string
	}{
		{20, 30, 25, 0.01, 4, "exploratory"},
		{20, 30, 25, 0.11, 5, "inconclusive"},
		{12, 30, 25, 0.01, 5, "goccy better"},
		{-30, -12, -25, 0.01, 5, "stdlib better"},
		{2, 8, 5, 0.01, 5, "detectable but below threshold"},
		{-1, 30, 15, 0.01, 5, "no detectable difference"},
	}
	for _, c := range cases {
		if got := verdict(c.lo, c.hi, c.pt, 10, c.cv, c.n); !strings.HasPrefix(got, c.want) {
			t.Errorf("verdict(%v)=%q want prefix %q", c, got, c.want)
		}
	}
}

func TestBootstrapDeterministic(t *testing.T) {
	b, g := trials(100, 104, 98, 101, 99), trials(80, 83, 79, 81, 82)
	a := compare(lowerM(), b, g, rng(), 2000, 10)
	c := compare(lowerM(), b, g, rng(), 2000, 10)
	if *a.CILow != *c.CILow || *a.CIHigh != *c.CIHigh {
		t.Fatalf("CI not deterministic: %v..%v vs %v..%v", *a.CILow, *a.CIHigh, *c.CILow, *c.CIHigh)
	}
	if !(*a.CILow > 0 && *a.CILow <= *a.Reduction && *a.Reduction <= *a.CIHigh) {
		t.Fatalf("CI %v..%v does not bracket %v", *a.CILow, *a.CIHigh, *a.Reduction)
	}
	if !strings.HasPrefix(a.Verdict, "goccy better") {
		t.Fatalf("verdict %q", a.Verdict)
	}
}

func TestControlsDowngrade(t *testing.T) {
	p := func(v float64) *float64 { return &v }
	cs := []comparison{
		{Metric: "Decode med", Verdict: "goccy better (CI excludes 0, ≥ threshold)", Reduction: p(50)},
		{Metric: "Policy med", control: true, Verdict: "goccy better", Reduction: p(-14.9)},
		{Metric: "Read med", control: true, Verdict: "no detectable difference", Reduction: p(1)},
	}
	applyControls(cs, 10)
	if cs[1].Verdict != verdictControl || cs[2].Verdict != verdictControl {
		t.Fatalf("controls not relabeled: %q %q", cs[1].Verdict, cs[2].Verdict)
	}
	if !strings.HasPrefix(cs[0].Verdict, "inconclusive") {
		t.Fatalf("library verdict not downgraded: %q", cs[0].Verdict)
	}
	// Stable controls leave verdicts intact.
	cs = []comparison{
		{Metric: "Decode med", Verdict: "goccy better", Reduction: p(50)},
		{Metric: "Policy med", control: true, Reduction: p(3)},
	}
	applyControls(cs, 10)
	if cs[0].Verdict != "goccy better" {
		t.Fatalf("verdict changed with stable controls: %q", cs[0].Verdict)
	}
}

func TestKnee(t *testing.T) {
	mk := func(off, ach, p95, n float64) step {
		return step{Offered: off, Achieved: ach, P95: p95, Count: n}
	}
	cases := []struct {
		name string
		ss   []step
		want string
	}{
		{"empty", nil, "n/a"},
		{"few samples", []step{mk(10, 10, 1, 50)}, "undetermined"},
		{"latency knee", []step{mk(100, 100, 1, 1000), mk(200, 200, 2.9, 1000), mk(300, 300, 3.1, 1000)}, "200 rps"},
		{"throughput knee", []step{mk(100, 100, 1, 1000), mk(200, 190, 1, 1000)}, "100 rps"},
		{"drops", []step{mk(100, 100, 1, 1000), {Offered: 200, Achieved: 200, P95: 1, Count: 1000, Dropped: 1}}, "100 rps"},
		{"below lowest", []step{mk(100, 50, 1, 1000)}, "below lowest step"},
	}
	for _, c := range cases {
		if got := knee(c.ss); !strings.HasPrefix(got, c.want) {
			t.Errorf("%s: knee=%q want prefix %q", c.name, got, c.want)
		}
	}
}
