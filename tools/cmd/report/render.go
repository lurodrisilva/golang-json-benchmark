package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"html"
	"math"
	"os"
	"sort"
	"strings"
)

type reportCtx struct {
	Run                                 string
	Cells                               []*cell
	Trials                              []*trial
	Threshold                           float64
	Resamples                           int
	Seed                                uint64
	Env, Build, Matrix, Fixtures, Bench []byte
	Conformance                         map[string][]byte
}

func f(v float64) string {
	switch {
	case math.IsNaN(v):
		return "n/a"
	case v == 0:
		return "0"
	case math.Abs(v) >= 1000:
		return fmt.Sprintf("%.0f", v)
	case math.Abs(v) >= 10:
		return fmt.Sprintf("%.1f", v)
	case math.Abs(v) >= 0.01:
		return fmt.Sprintf("%.3f", v)
	}
	return fmt.Sprintf("%.3g", v)
}

func pct(p *float64) string {
	if p == nil {
		return "n/a"
	}
	return fmt.Sprintf("%+.1f%%", *p)
}

func renderMarkdown(c *reportCtx) string {
	var b strings.Builder
	w := func(format string, a ...any) { fmt.Fprintf(&b, format, a...) }
	run := c.Run
	if i := strings.LastIndex(run, "/"); i >= 0 {
		run = run[i+1:]
	}
	w("# encoding/json vs goccy/go-json: deserialization benchmark report\n\n")
	w("Run `%s`. This report is generated from the artifacts on disk; it contains no hand-entered numbers.\n\n", run)
	w("Tags: **[M]** measured · **[A]** analysis/derived · **[H]** hypothesis · **[L]** limitation.\n\n")

	w("## 1. Executive summary\n\n")
	valid, flagged, invalid := 0, 0, 0
	for _, t := range c.Trials {
		switch t.Status {
		case "valid":
			valid++
		case "flagged":
			flagged++
		default:
			invalid++
		}
	}
	w("- [M] Trials: %d total. %d valid, %d flagged, %d invalid/error.\n", len(c.Trials), valid, flagged, invalid)
	w("- [A] Practical-significance threshold: **%.0f%%** (user-approved). A metric is called for one library only when the bootstrap 95%% CI excludes 0 **and** the median-of-trials difference is at least the threshold, with ≥5 trials per arm and trial CV ≤ 10%%.\n", c.Threshold)
	w("- [A] Baseline is always the stdlib arm of the pair. `reduction %% = (stdlib − goccy) / stdlib × 100` for lower-is-better metrics. For higher-is-better metrics (throughput) the sign is set so that positive always means goccy is better.\n")
	w("- [L] Hardware: one Apple M3 (4P+4E), with server and k6 sharing the host over loopback. These results are **pilot grade**. See §9.\n\n")
	w("**Headline comparisons (fixed-rate cells):**\n\n")
	w("| Cell | Metric | stdlib | goccy | reduction | 95%% CI | verdict |\n|---|---|---:|---:|---:|---|---|\n")
	headline := map[string]bool{"HTTP request duration (k6) med": true, "HTTP request duration (k6) p(99)": true,
		"Decode duration (Unmarshal only) med": true, "CPU time per successful op": true, "Allocated bytes per op": true}
	for _, cl := range c.Cells {
		for _, cp := range cl.Compares {
			if headline[cp.Metric] && cl.Key.Variant == "" && cl.Key.Profile == "none" {
				w("| %s | %s (%s) | %s | %s | %s | %s | %s |\n", cl.Name, cp.Metric, cp.Unit, med(cp, cp.BaseMedian), med(cp, cp.CandMedian), pct(cp.Reduction), ci(cp), cp.Verdict)
			}
		}
	}
	w("\n**Conditional recommendation [A]:** you can only adopt goccy/go-json on this evidence if the decode-time and CPU-per-op verdicts above are \"goccy better\" for the payload sizes you actually serve, **and** the correctness matrix (§3) shows no deviation that matters to your inputs. Treat cells marked exploratory, inconclusive or below threshold as \"no justification for the dependency\".\n\n")

	w("## 2. Environment and reproducibility manifest\n\n")
	w("Exact build metadata (binaries are preserved under `<run>/bin/`):\n\n")
	w("```json\n%s\n```\n\n", trimJSON(c.Build, 6000))
	w("Host and toolchain:\n\n```json\n%s\n```\n\n", trimJSON(c.Env, 6000))
	w("Fixtures (SHA-256):\n\n```\n%s```\n\n", string(c.Fixtures))
	w("Matrix:\n\n```json\n%s\n```\n\n", strings.TrimSpace(string(c.Matrix)))

	w("## 3. Correctness and compatibility\n\n")
	if len(c.Conformance) == 0 {
		w("- Conformance matrices were not supplied (`-conformance`). Run `make conformance`.\n\n")
	} else {
		names := make([]string, 0, len(c.Conformance))
		for n := range c.Conformance {
			names = append(names, n)
		}
		sort.Strings(names)
		for _, n := range names {
			rows := strings.Split(strings.TrimSpace(string(c.Conformance[n])), "\n")
			dev := 0
			for _, r := range rows[1:] {
				// columns: goexperiment codec case spec actual value_check conforms error
				if cols := strings.Split(r, "\t"); len(cols) >= 7 && cols[6] != "true" {
					dev++
				}
			}
			w("- [M] `%s`: %d rows, **%d deviations** from the v1-documented expectation.\n", n, len(rows)-1, dev)
		}
		w("\nThe full matrices are in `conformance/`. Expected outcomes are written from the encoding/json v1 documented semantics. Each library is compared against that spec, not against the other library.\n\n")
	}

	w("## 4. Fixed-offered-load results\n\n")
	w("Every value is the **median of per-trial values**. Per-trial percentiles come from k6 for that trial alone. They are never averaged across trials or stages. P99 is shown only when each trial has ≥%d samples. Read, policy and validate rows run **identical shared code** in both arms. They are **negative controls** and never get a library verdict. If any control median differs by ≥ %.0f%%, the environment is too unstable to attribute differences to the library, and every library verdict in that cell is downgraded to *inconclusive*.\n\n", p99MinSamples, c.Threshold)
	for _, cl := range c.Cells {
		if cl.Key.Workload != "fixed" {
			continue
		}
		w("### %s\n\n", cl.Name)
		w("- [M] trials analysed: stdlib-arm %d, goccy-arm %d.\n", len(cl.Base), len(cl.Cand))
		for _, e := range cl.Excluded {
			w("- [M] %s\n", e)
		}
		if len(cl.Base) > 0 {
			t := cl.Base[0]
			w("- [M] window: nominal %s s, wall %s s, drain %s s (per trial; the drain is k6 finishing in-flight requests).\n", f(t.Window["nominal_s"]), f(t.Window["wall_s"]), f(t.Window["drain_s"]))
		}
		w("\n| Metric | unit | stdlib (trials) | goccy (trials) | stdlib median | goccy median | Δ abs | reduction | 95%% CI | verdict |\n|---|---|---|---|---:|---:|---:|---:|---|---|\n")
		for _, cp := range cl.Compares {
			w("| %s | %s | %s | %s | %s | %s | %s | %s | %s | %s%s |\n", cp.Metric, cp.Unit, list(cp.Baseline), list(cp.Candidate),
				med(cp, cp.BaseMedian), med(cp, cp.CandMedian), med(cp, cp.AbsDiff), pct(cp.Reduction), ci(cp), cp.Verdict, note(cp.Note))
		}
		w("\n")
		for _, pr := range cl.Profiling {
			w("<details><summary>[M] %s rep %d %s profile: %.0f decoded ops in the window</summary>\n\n```\n%s```\n</details>\n\n", pr.Arm, pr.Rep, pr.Kind, pr.DecodedOK, pr.Top)
		}
	}

	w("## 5. Saturation (stepped offered load)\n\n")
	for _, cl := range c.Cells {
		if cl.Key.Workload != "ramp" {
			continue
		}
		w("### %s\n\n", cl.Name)
		labels := sortedKeys(cl.Steps)
		for _, l := range labels {
			w("**%s**: knee = %s [A]\n\n", l, cl.Knee[l])
			w("| step | offered rps | achieved rps | dropped | non-202 | P50 ms | P95 ms | P99 ms | decode P50 ms | samples |\n|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|\n")
			for _, s := range cl.Steps[l] {
				w("| %d | %s | %s | %s | %s | %s | %s | %s | %s | %s |\n", s.Index, f(s.Offered), f(s.Achieved), f(s.Dropped), f(s.Non202), f(s.P50), f(s.P95), stepP99(s), f(s.DecMed), f(s.Count))
			}
			w("\n")
		}
		for _, e := range cl.Excluded {
			w("- [M] %s\n", e)
		}
		w("- [L] Each step's percentiles come from that step's samples only. The knee follows plan §14.4: the highest step with ≥99%% achieved, no drops, no non-202, and P95 ≤ 3× the first step's P95 (P95 replaces the plan's P99 criterion because low steps of large fixtures cannot reach the 10,000-sample P99 minimum; each step needs ≥%d samples). Step P99 is shown only for steps with ≥%d samples. If any step's k6 CPU limit was hit, the run is marked invalid instead of reading as a knee.\n\n", kneeMinSamples, p99MinSamples)
	}

	w("## 6. Decoder microbenchmarks (isolated, not HTTP)\n\n")
	if len(c.Bench) == 0 {
		w("- Not supplied (`-bench`). Run `make microbench` (benchstat output).\n\n")
	} else {
		w("[M] benchstat output (`go test -bench` with -count≥10). These conclusions are **separate** from the HTTP conclusions:\n\n```\n%s\n```\n\n", strings.TrimSpace(string(c.Bench)))
	}

	w("## 7. Attribution: decoder vs HTTP, GC, allocation, validation, instrumentation\n\n")
	w("- [A] Handler time = read + policy scan + decode + validate (+ timer overhead). Policy and validate are shared code and should match across arms. A difference in them points to GC or scheduler interference, not to the decoder.\n")
	w("- [A] HTTP time − handler time ≈ network, kernel, and k6 client cost, plus response delivery. The library under test does not touch these.\n")
	w("- [A] The `variant=minimal` cells (no timestamps, no Server-Timing) bound the instrumentation overhead. Compare their HTTP latency and CPU/op with the matching full-instrumentation cell.\n")
	w("- [A] Profile cells are flagged by design and are not used for latency claims. CPU profiles are sampled evidence. Allocation profiles are windowed deltas (`allocs?seconds=60`). Heap profiles show retained objects at the end of the trial. pprof does not measure RSS.\n")
	w("- [L] The two arms are different binaries, so a symbol in one profile has no counterpart to diff against in the other. Compare shares of a profile (e.g. %% of CPU in the decoder), not absolute sample counts.\n\n")

	w("## 8. Generator capacity\n\n| trial | k6 avg cores | k6 peak cores | server avg cores (ps) | speed limit before/after |\n|---|---:|---:|---:|---|\n")
	for _, t := range c.Trials {
		w("| %s | %s | %s | %s | %s/%s |\n", relDir(c.Run, t.dir), f(t.Host["k6_cpu_avg_cores"]), f(t.Host["k6_cpu_peak_cores"]), f(t.Host["server_cpu_avg_cores_ps"]), f(t.Host["cpu_speed_limit_before"]), f(t.Host["cpu_speed_limit_after"]))
	}
	w("\n## 9. Threats to validity and open questions\n\n")
	for _, s := range []string{
		"[L] There is a single host with heterogeneous cores, and macOS offers no CPU pinning. The server (GOMAXPROCS=4) and k6 compete for the same 8 cores and can be scheduled onto E-cores.",
		"[L] Traffic goes over loopback, so the network path is unrealistically short. Body-read cost is therefore a lower bound.",
		"[L] The fixtures are synthetic (seeded generator). F1/F2/F3 come from one schema with different record counts, so the matrix is a size sweep and does not test schema generality.",
		"[L] Window snapshots are taken by benchctl after the warm-up ends. Their boundaries differ from k6's measure-phase boundary by up to a few requests. The k6/server count check is applied to the whole-trial totals.",
		"[L] GC pause totals come from histogram bucket lower bounds, so they understate the true pause time.",
		"[L] n=5 trials per cell gives limited statistical power. The bootstrap CI on the median of 5 values is coarse.",
		"[H] goccy's decode advantage (if any) may shrink at larger payloads if GC dominates. The F2/F3 cells test this.",
		"[L] Both stdlib arms are reported. The default go1.27 encoding/json is backed by the v2 implementation; Pair L uses GOEXPERIMENT=nojsonv2. The two pairs must not be mixed.",
	} {
		w("- %s\n", s)
	}
	return b.String()
}

func ci(cp comparison) string {
	if cp.CILow == nil {
		return "n/a"
	}
	return fmt.Sprintf("[%+.1f, %+.1f]", *cp.CILow, *cp.CIHigh)
}

func note(n string) string {
	if n == "" {
		return ""
	}
	return " (" + n + ")"
}

func list(xs []float64) string {
	if len(xs) == 0 {
		return "—"
	}
	s := make([]string, len(xs))
	for i, x := range xs {
		s[i] = f(x)
	}
	return strings.Join(s, ", ")
}

func sortedKeys[V any](m map[string]V) []string {
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	return ks
}

func relDir(run, dir string) string {
	return strings.TrimPrefix(strings.TrimPrefix(dir, run), "/trials/")
}

func trimJSON(b []byte, n int) string {
	if len(b) == 0 {
		return "{}"
	}
	var v any
	if json.Unmarshal(b, &v) == nil {
		if p, err := json.MarshalIndent(v, "", "  "); err == nil {
			b = p
		}
	}
	if len(b) > n {
		return string(b[:n]) + "\n… (truncated; see file)"
	}
	return string(b)
}

func writeCSV(path string, ts []*trial) error {
	fh, err := os.Create(path)
	if err != nil {
		return err
	}
	defer fh.Close()
	w := csv.NewWriter(fh)
	cols := map[string]bool{}
	for _, t := range ts {
		for k := range t.K6 {
			if !strings.Contains(k, "{") {
				cols["k6."+k] = true
			}
		}
		for k := range t.Server {
			cols["server."+k] = true
		}
		for k := range t.Host {
			cols["host."+k] = true
		}
	}
	names := sortedKeys(cols)
	head := append([]string{"dir", "arm", "fixture", "workload", "rate", "knee", "rep", "profile", "variant", "status", "reasons", "wall_s", "drain_s"}, names...)
	w.Write(head)
	for _, t := range ts {
		row := []string{t.dir, t.Spec.Arm, t.Spec.Fixture, t.Spec.Workload, fmt.Sprint(t.Spec.Rate), fmt.Sprint(t.Spec.Knee),
			fmt.Sprint(t.Spec.Rep), t.Spec.Profile, t.Spec.Variant, t.Status, strings.Join(t.Reasons, "; "),
			fmt.Sprint(t.Window["wall_s"]), fmt.Sprint(t.Window["drain_s"])}
		for _, n := range names {
			src, key, _ := strings.Cut(n, ".")
			var m map[string]float64
			switch src {
			case "k6":
				m = t.K6
			case "server":
				m = t.Server
			default:
				m = t.Host
			}
			if v, ok := m[key]; ok {
				row = append(row, fmt.Sprint(v))
			} else {
				row = append(row, "")
			}
		}
		w.Write(row)
	}
	w.Flush()
	return w.Error()
}

// ---------- HTML ----------

func renderHTML(c *reportCtx, md string) string {
	var b strings.Builder
	b.WriteString(`<!doctype html><html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>JSON Decode Benchmark</title><style>
:root{--bg:#fff;--fg:#1b1f24;--muted:#5b6470;--line:#d8dde3;--a:#2f6db5;--b:#c8572f;--code:#f4f6f8}
@media (prefers-color-scheme:dark){:root:not([data-theme=light]){--bg:#14171b;--fg:#e6e9ed;--muted:#9aa4b0;--line:#2c333b;--a:#6ea8ec;--b:#f08a5d;--code:#1d2228}}
:root[data-theme=dark]{--bg:#14171b;--fg:#e6e9ed;--muted:#9aa4b0;--line:#2c333b;--a:#6ea8ec;--b:#f08a5d;--code:#1d2228}
body{background:var(--bg);color:var(--fg);font:15px/1.5 system-ui,sans-serif;max-width:1100px;margin:0 auto;padding:16px}
table{border-collapse:collapse;display:block;overflow-x:auto;font-size:13px}td,th{border:1px solid var(--line);padding:3px 6px;text-align:left}
pre{background:var(--code);padding:8px;overflow-x:auto;font-size:12px}svg text{fill:var(--fg);font-size:11px}.muted{color:var(--muted)}
</style></head><body>`)
	b.WriteString(`<h1>Saturation charts</h1><p class="muted">Each line is one trial. The x axis is offered req/s. Blue lines are stdlib-arm trials and orange lines are goccy-arm trials.</p>`)
	for _, cl := range c.Cells {
		if cl.Key.Workload != "ramp" {
			continue
		}
		b.WriteString("<h2>" + html.EscapeString(cl.Name) + "</h2>")
		b.WriteString(chart(cl, "P95 latency (ms) vs offered load", func(s step) float64 { return s.P95 }))
		b.WriteString(chart(cl, "Achieved vs offered (req/s)", func(s step) float64 { return s.Achieved }))
	}
	b.WriteString("<h1>Report</h1>")
	b.WriteString(mdToHTML(md))
	b.WriteString("</body></html>\n")
	return b.String()
}

func chart(cl *cell, title string, y func(step) float64) string {
	const W, H, P = 640, 260, 44
	var maxX, maxY float64
	for _, ss := range cl.Steps {
		for _, s := range ss {
			maxX, maxY = math.Max(maxX, s.Offered), math.Max(maxY, y(s))
		}
	}
	if maxX == 0 || maxY == 0 {
		return ""
	}
	maxY *= 1.1
	var b strings.Builder
	fmt.Fprintf(&b, `<figure><figcaption>%s</figcaption><svg viewBox="0 0 %d %d" width="100%%" style="max-width:%dpx" role="img" aria-label="%s">`, html.EscapeString(title), W, H, W, html.EscapeString(title))
	fmt.Fprintf(&b, `<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="currentColor" opacity=".4"/><line x1="%d" y1="%d" x2="%d" y2="%d" stroke="currentColor" opacity=".4"/>`, P, H-P, W-10, H-P, P, 10, P, H-P)
	fmt.Fprintf(&b, `<text x="%d" y="%d">0</text><text x="%d" y="%d" text-anchor="end">%s</text><text x="4" y="16">%s</text>`, P, H-P+14, W-10, H-P+14, f(maxX), f(maxY))
	for _, label := range sortedKeys(cl.Steps) {
		color := "var(--b)"
		if arm, _, _ := strings.Cut(label, " "); func() bool { _, base := pair(arm); return base }() {
			color = "var(--a)"
		}
		var pts []string
		for _, s := range cl.Steps[label] {
			px := P + (s.Offered/maxX)*float64(W-10-P)
			py := float64(H-P) - (y(s)/maxY)*float64(H-P-10)
			pts = append(pts, fmt.Sprintf("%.1f,%.1f", px, py))
		}
		fmt.Fprintf(&b, `<polyline fill="none" stroke="%s" stroke-width="1.6" points="%s"><title>%s</title></polyline>`, color, strings.Join(pts, " "), html.EscapeString(label))
	}
	b.WriteString("</svg></figure>")
	return b.String()
}

// mdToHTML is a deliberately small converter for the subset report.md uses:
// headings, bullets, pipe tables, fenced code, <details> passthrough, **bold**.
func mdToHTML(md string) string {
	var b strings.Builder
	lines := strings.Split(md, "\n")
	inCode, inList, inTable := false, false, false
	closeBlocks := func() {
		if inList {
			b.WriteString("</ul>")
			inList = false
		}
		if inTable {
			b.WriteString("</table>")
			inTable = false
		}
	}
	for _, l := range lines {
		switch {
		case strings.HasPrefix(l, "```"):
			closeBlocks()
			if inCode {
				b.WriteString("</pre>")
			} else {
				b.WriteString("<pre>")
			}
			inCode = !inCode
		case inCode:
			b.WriteString(html.EscapeString(l) + "\n")
		case strings.HasPrefix(l, "<details>") || strings.HasPrefix(l, "</details>"):
			closeBlocks()
			b.WriteString(strings.ReplaceAll(strings.ReplaceAll(l, "<details><summary>", "<details><summary>"), "</summary>", "</summary>"))
		case strings.HasPrefix(l, "#"):
			closeBlocks()
			n := len(l) - len(strings.TrimLeft(l, "#"))
			fmt.Fprintf(&b, "<h%d>%s</h%d>", n, inline(strings.TrimSpace(l[n:])), n)
		case strings.HasPrefix(l, "- "):
			if inTable {
				b.WriteString("</table>")
				inTable = false
			}
			if !inList {
				b.WriteString("<ul>")
				inList = true
			}
			b.WriteString("<li>" + inline(l[2:]) + "</li>")
		case strings.HasPrefix(l, "|"):
			if inList {
				b.WriteString("</ul>")
				inList = false
			}
			cells := strings.Split(strings.Trim(l, "|"), "|")
			if strings.Trim(cells[0], "-: ") == "" {
				continue
			}
			tag := "td"
			if !inTable {
				b.WriteString("<table>")
				inTable = true
				tag = "th"
			}
			b.WriteString("<tr>")
			for _, cell := range cells {
				fmt.Fprintf(&b, "<%s>%s</%s>", tag, inline(strings.TrimSpace(cell)), tag)
			}
			b.WriteString("</tr>")
		case strings.TrimSpace(l) == "":
			closeBlocks()
		default:
			closeBlocks()
			b.WriteString("<p>" + inline(l) + "</p>")
		}
	}
	closeBlocks()
	return b.String()
}

func inline(s string) string {
	s = html.EscapeString(s)
	for {
		i := strings.Index(s, "**")
		if i < 0 {
			break
		}
		j := strings.Index(s[i+2:], "**")
		if j < 0 {
			break
		}
		s = s[:i] + "<strong>" + s[i+2:i+2+j] + "</strong>" + s[i+4+j:]
	}
	for {
		i := strings.Index(s, "`")
		if i < 0 {
			break
		}
		j := strings.Index(s[i+1:], "`")
		if j < 0 {
			break
		}
		s = s[:i] + "<code>" + s[i+1:i+1+j] + "</code>" + s[i+2+j:]
	}
	return s
}

// med renders a median, or "n/a" when the metric has no observations.
func med(cp comparison, v float64) string {
	if len(cp.Baseline) == 0 || len(cp.Candidate) == 0 {
		return "n/a"
	}
	return f(v)
}

func stepP99(s step) string {
	if s.Count < p99MinSamples {
		return "n/a"
	}
	return f(s.P99)
}
