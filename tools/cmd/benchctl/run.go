package main

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"math/rand/v2"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

// Matrix is a run definition (loadtest/matrix/*.json). Rates are ABSOLUTE
// and shared by both arms of a pair — derived from the slower arm's pilot
// knee — so neither library is tested at a rate relative to its own capacity.
type Matrix struct {
	Name       string `json:"name"`
	Seed       uint64 `json:"seed"`
	GOMAXPROCS int    `json:"gomaxprocs"`
	Cooldown   string `json:"cooldown"`
	PreVUs     int    `json:"pre_vus"`
	MaxVUs     int    `json:"max_vus"`
	Cells      []Cell `json:"cells"`
}

type Cell struct {
	Pair     []string `json:"pair"` // exactly two arms sharing GOEXPERIMENT
	Fixture  string   `json:"fixture"`
	Workload string   `json:"workload"`
	Rate     int      `json:"rate,omitempty"`
	Duration string   `json:"duration,omitempty"`
	Warmup   string   `json:"warmup,omitempty"`
	Knee     int      `json:"knee,omitempty"`
	Steps    string   `json:"steps,omitempty"`
	StepDur  string   `json:"step_dur,omitempty"`
	Reps     int      `json:"reps"`
	Profile  string   `json:"profile,omitempty"`
	Variant  string   `json:"variant,omitempty"`
	Op       string   `json:"op,omitempty"` // "" | decode | encode
}

func cmdRun(args []string) error {
	fs := flag.NewFlagSet("run", flag.ExitOnError)
	matrixPath := fs.String("matrix", "", "matrix JSON (required)")
	binDir := fs.String("bin", "bin", "built binaries + build.json")
	resultsDir := fs.String("results", "results", "results root")
	dry := fs.Bool("dry-run", false, "print the ordered trial plan and estimated duration only")
	allowDirty := fs.Bool("allow-dirty", false, "permit a run from an uncommitted working tree (recorded in the run manifest; run-id gets a -dirty suffix)")
	fs.Parse(args)
	if *matrixPath == "" {
		return errors.New("-matrix is required")
	}
	raw, err := os.ReadFile(*matrixPath)
	if err != nil {
		return err
	}
	var mx Matrix
	if err := json.Unmarshal(raw, &mx); err != nil {
		return fmt.Errorf("matrix: %w", err)
	}
	if mx.GOMAXPROCS == 0 {
		mx.GOMAXPROCS = 4
	}
	if mx.Cooldown == "" {
		mx.Cooldown = "120s"
	}
	cool, err := time.ParseDuration(mx.Cooldown)
	if err != nil {
		return err
	}
	plan, err := order(mx)
	if err != nil {
		return err
	}
	est := estimate(plan, cool)
	if *dry {
		for i, s := range plan {
			fmt.Printf("%3d  %-14s %-8s %-6s rate=%-5d knee=%-5d rep=%d profile=%s variant=%s\n",
				i+1, s.Arm, s.Fixture, s.Workload, s.Rate, s.Knee, s.Rep, s.Profile, s.Variant)
		}
		fmt.Printf("%d trials, estimated %s (incl. cooldowns)\n", len(plan), est.Round(time.Minute))
		return nil
	}

	root, err := repoRoot()
	if err != nil {
		return err
	}
	if err := verifyFixtures(root); err != nil {
		return err
	}
	bins, _ := filepath.Abs(*binDir)
	if err := verifyBinaries(bins, plan); err != nil {
		return err
	}
	sum := sha256.Sum256(raw)
	git := gitInfo(root)
	sha := git["commit"]
	if len(sha) > 7 {
		sha = sha[:7]
	}
	if sha == "" {
		sha = "nogit"
	}
	// Provenance: the run-id names a commit, so the code that ran must be
	// that commit. A dirty tree is refused unless explicitly allowed.
	if git["dirty"] == "true" {
		if !*allowDirty {
			return errors.New("working tree has uncommitted changes: commit first, or pass -allow-dirty (recorded; results are not attributable to a commit)")
		}
		sha += "-dirty"
	}
	if err := checkBinaryVCS(bins, git["commit"], *allowDirty); err != nil {
		return err
	}
	runID := fmt.Sprintf("%s-%s-%s", time.Now().UTC().Format("20060102T150405Z"), sha, hex.EncodeToString(sum[:])[:8])
	runDir := filepath.Join(*resultsDir, runID)
	if err := os.MkdirAll(filepath.Dir(runDir), 0o755); err != nil {
		return err
	}
	if err := os.Mkdir(runDir, 0o755); err != nil {
		return fmt.Errorf("run dir: %w", err)
	}
	fmt.Fprintf(os.Stderr, "run %s: %d trials, estimated %s\n", runID, len(plan), est.Round(time.Minute))
	os.WriteFile(filepath.Join(runDir, "matrix.json"), raw, 0o644)
	writeJSON(filepath.Join(runDir, "env.json"), envManifest())
	writeJSON(filepath.Join(runDir, "provenance.json"), map[string]any{"git": git, "allow_dirty": *allowDirty, "matrix_sha256": hex.EncodeToString(sum[:])})
	copyFile(filepath.Join(bins, "build.json"), filepath.Join(runDir, "build.json"))
	copyFile(filepath.Join(root, "fixtures", "SHA256SUMS"), filepath.Join(runDir, "fixtures.SHA256SUMS"))
	// Preserve the exact binaries used (profile symbolization) and run the
	// trials from those copies, so a rebuild mid-run cannot leak in.
	runBins := filepath.Join(runDir, "bin")
	if err := os.Mkdir(runBins, 0o755); err != nil {
		return err
	}
	copied := map[string]bool{}
	for _, s := range plan {
		s.BinDir = bins
		p := s.binary()
		if copied[p] {
			continue
		}
		copied[p] = true
		if err := copyExec(p, filepath.Join(runBins, filepath.Base(p))); err != nil {
			return err
		}
	}
	copyFile(filepath.Join(bins, "build.json"), filepath.Join(runBins, "build.json"))
	if err := verifyBinaries(runBins, plan); err != nil {
		return err
	}
	bins = runBins
	idx, err := os.OpenFile(filepath.Join(runDir, "index.jsonl"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer idx.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	consecutiveErr := 0
	for i, s := range plan {
		if ctx.Err() != nil {
			return fmt.Errorf("interrupted after %d/%d trials; partial run kept in %s", i, len(plan), runDir)
		}
		s.RunDir, s.BinDir = runDir, bins
		fmt.Fprintf(os.Stderr, "[%d/%d] %s %s %s%s rep %d\n", i+1, len(plan), s.Arm, s.Fixture, opLabel(s), s.Workload, s.Rep)
		res, err := runTrial(ctx, s)
		line := map[string]any{"seq": i + 1, "dir": strings.TrimPrefix(s.dir(), runDir+"/")}
		if res != nil {
			line["status"], line["reasons"] = res.Status, res.Reasons
		}
		if err != nil {
			line["error"] = err.Error()
			consecutiveErr++
		} else {
			consecutiveErr = 0
		}
		b, _ := json.Marshal(line)
		idx.Write(append(b, '\n'))
		if consecutiveErr >= 3 {
			return fmt.Errorf("3 consecutive trial errors; aborting (last: %v)", err)
		}
		if i < len(plan)-1 {
			select {
			case <-time.After(cool):
			case <-ctx.Done():
			}
		}
	}
	fmt.Println(runDir)
	return nil
}

// order expands cells into trials. Within each cell, repetitions alternate
// arm order ABBA-style (rep1 A,B; rep2 B,A; …) with a seeded random choice of
// which arm goes first; cells themselves are shuffled with the same seed.
func order(mx Matrix) ([]TrialSpec, error) {
	rng := rand.New(rand.NewPCG(mx.Seed, mx.Seed^0x5851f42d4c957f2d))
	cells := append([]Cell(nil), mx.Cells...)
	rng.Shuffle(len(cells), func(i, j int) { cells[i], cells[j] = cells[j], cells[i] })
	var out []TrialSpec
	for _, c := range cells {
		if len(c.Pair) != 2 {
			return nil, fmt.Errorf("cell %s/%s: pair must have exactly 2 arms", c.Fixture, c.Workload)
		}
		a, err := armByName(c.Pair[0])
		if err != nil {
			return nil, err
		}
		b, err := armByName(c.Pair[1])
		if err != nil {
			return nil, err
		}
		if a.GOEXPERIMENT != b.GOEXPERIMENT {
			return nil, fmt.Errorf("pair %v mixes GOEXPERIMENT settings", c.Pair)
		}
		switch c.Workload {
		case "fixed":
			if c.Rate <= 0 || c.Duration == "" {
				return nil, fmt.Errorf("cell %s/fixed: rate and duration are required (absolute rate from the slower arm's pilot knee)", c.Fixture)
			}
		case "ramp":
			if c.Knee <= 0 {
				return nil, fmt.Errorf("cell %s/ramp: knee is required", c.Fixture)
			}
		case "smoke":
		default:
			return nil, fmt.Errorf("cell %s: unknown workload %q", c.Fixture, c.Workload)
		}
		if c.Reps == 0 {
			c.Reps = 5
		}
		if c.Profile == "" {
			c.Profile = "none"
		}
		switch c.Op {
		case "", "decode":
			c.Op = ""
		case "encode":
		default:
			return nil, fmt.Errorf("cell %s/%s: unknown op %q", c.Fixture, c.Workload, c.Op)
		}
		first := rng.IntN(2)
		for rep := 1; rep <= c.Reps; rep++ {
			pair := []string{c.Pair[first], c.Pair[1-first]}
			if rep%2 == 0 {
				pair[0], pair[1] = pair[1], pair[0]
			}
			for _, arm := range pair {
				out = append(out, TrialSpec{Arm: arm, Fixture: c.Fixture, Workload: c.Workload, Rate: c.Rate,
					Duration: c.Duration, Warmup: c.Warmup, Knee: c.Knee, Steps: c.Steps, StepDur: c.StepDur,
					Rep: rep, Profile: c.Profile, GOMAXPROCS: mx.GOMAXPROCS, PreVUs: or(mx.PreVUs, 50),
					MaxVUs: or(mx.MaxVUs, 400), Addr: "127.0.0.1:8080", Mgmt: "127.0.0.1:6061", Raw: true,
					Variant: c.Variant, Op: c.Op})
			}
		}
	}
	return out, nil
}

func or(v, d int) int {
	if v == 0 {
		return d
	}
	return v
}

func estimate(plan []TrialSpec, cool time.Duration) time.Duration {
	var t time.Duration
	for _, s := range plan {
		w, n := windowPlan(s)
		if s.Workload == "smoke" {
			n = 5 * time.Second
		}
		t += w + n + 15*time.Second + cool // 15 s: start, readiness, drain, shutdown
	}
	return t
}

// verifyFixtures checks fixtures against SHA256SUMS before any trial.
func verifyFixtures(root string) error {
	f, err := os.Open(filepath.Join(root, "fixtures", "SHA256SUMS"))
	if err != nil {
		return err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		want, name, ok := strings.Cut(sc.Text(), "  ")
		if !ok {
			continue
		}
		got, err := fileSHA256(filepath.Join(root, "fixtures", name))
		if err != nil {
			return err
		}
		if got != want {
			return fmt.Errorf("fixture %s checksum mismatch: %s != %s", name, got, want)
		}
	}
	return sc.Err()
}

// verifyBinaries refuses to run if any binary differs from build.json
// (protects against a stale or hand-rebuilt binary).
func verifyBinaries(bins string, plan []TrialSpec) error {
	b, err := os.ReadFile(filepath.Join(bins, "build.json"))
	if err != nil {
		return fmt.Errorf("build.json: %w (run `benchctl build` first)", err)
	}
	var m struct {
		Arms []BuiltArm `json:"arms"`
	}
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}
	want := map[string]string{}
	for _, a := range m.Arms {
		want[filepath.Base(a.Binary)] = a.SHA256
	}
	seen := map[string]bool{}
	for _, s := range plan {
		s.BinDir = bins
		p := s.binary()
		if seen[p] {
			continue
		}
		seen[p] = true
		got, err := fileSHA256(p)
		if err != nil {
			return err
		}
		if want[filepath.Base(p)] != got {
			return fmt.Errorf("%s does not match build.json (stale binary?)", p)
		}
	}
	return nil
}

func copyExec(src, dst string) error {
	b, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, b, 0o755)
}

func copyFile(src, dst string) {
	if b, err := os.ReadFile(src); err == nil {
		os.WriteFile(dst, b, 0o644)
	}
}

// checkBinaryVCS requires every built binary to embed vcs.revision == HEAD
// and vcs.modified=false (from `go version -m`), so the run-id's commit is
// the code that was compiled. -allow-dirty relaxes both (recorded).
func checkBinaryVCS(bins, head string, allowDirty bool) error {
	b, err := os.ReadFile(filepath.Join(bins, "build.json"))
	if err != nil {
		return err
	}
	var m struct {
		Arms []BuiltArm `json:"arms"`
	}
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}
	for _, a := range m.Arms {
		rev, mod := vcsSetting(a.GoVersionM, "vcs.revision"), vcsSetting(a.GoVersionM, "vcs.modified")
		if allowDirty {
			continue
		}
		if mod != "false" {
			return fmt.Errorf("%s built from a modified tree (vcs.modified=%q): rebuild after committing, or pass -allow-dirty", a.Binary, mod)
		}
		if head != "" && rev != head {
			return fmt.Errorf("%s built at %s, HEAD is %s: rebuild (stale binary)", a.Binary, rev, head)
		}
	}
	return nil
}

func vcsSetting(goVersionM, key string) string {
	for _, l := range strings.Split(goVersionM, "\n") {
		f := strings.Fields(l)
		if len(f) >= 2 && f[0] == "build" && strings.HasPrefix(f[1], key+"=") {
			return strings.TrimPrefix(f[1], key+"=")
		}
	}
	return ""
}

func opLabel(s TrialSpec) string {
	if s.encode() {
		return "encode "
	}
	return ""
}
