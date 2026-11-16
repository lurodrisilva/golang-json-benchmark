package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// TrialSpec fully determines one trial. It is written verbatim into
// trial.json so every number can be traced to its inputs.
type TrialSpec struct {
	RunDir     string `json:"-"`
	Arm        string `json:"arm"`
	BinDir     string `json:"bin_dir"`
	Fixture    string `json:"fixture"`  // f1-64k | f2-256k | f3-1m
	Workload   string `json:"workload"` // smoke | fixed | ramp
	Rate       int    `json:"rate,omitempty"`
	Duration   string `json:"duration,omitempty"`
	Warmup     string `json:"warmup,omitempty"`
	Knee       int    `json:"knee,omitempty"`
	Steps      string `json:"steps,omitempty"`
	StepDur    string `json:"step_dur,omitempty"`
	Rep        int    `json:"rep"`
	Profile    string `json:"profile"` // none | cpu | allocs
	GOMAXPROCS int    `json:"gomaxprocs"`
	PreVUs     int    `json:"pre_vus"`
	MaxVUs     int    `json:"max_vus"`
	Addr       string `json:"addr"`
	Mgmt       string `json:"mgmt"`
	Raw        bool   `json:"raw_samples"`
	Variant    string `json:"variant"` // "" | minimal (instrumentation control)
}

func (s TrialSpec) dir() string {
	load := s.Workload
	switch s.Workload {
	case "fixed":
		load = fmt.Sprintf("fixed-r%d", s.Rate)
	case "ramp":
		load = fmt.Sprintf("ramp-k%d", s.Knee)
	}
	arm := s.Arm
	if s.Variant != "" {
		arm += "+" + s.Variant
	}
	if s.Profile != "none" {
		load += "-prof-" + s.Profile
	}
	return filepath.Join(s.RunDir, "trials", arm, s.Fixture, load, fmt.Sprintf("rep-%02d", s.Rep))
}

func (s TrialSpec) binary() string {
	name := s.Arm + "-server"
	if s.Variant != "" {
		name = s.Arm + "-" + s.Variant + "-server"
	}
	return filepath.Join(s.BinDir, name)
}

func cmdTrial(args []string) error {
	fs := flag.NewFlagSet("trial", flag.ExitOnError)
	var s TrialSpec
	fs.StringVar(&s.RunDir, "run-dir", "", "results/<run-id> (required; created if absent)")
	fs.StringVar(&s.Arm, "arm", "", "stdlib|goccy|stdlib-legacy|goccy-legacy")
	fs.StringVar(&s.BinDir, "bin", "bin", "directory with built binaries")
	fs.StringVar(&s.Fixture, "fixture", "f1-64k", "fixture name")
	fs.StringVar(&s.Workload, "workload", "smoke", "smoke|fixed|ramp")
	fs.IntVar(&s.Rate, "rate", 0, "fixed: offered requests/s")
	fs.StringVar(&s.Duration, "duration", "10m", "fixed: measured duration")
	fs.StringVar(&s.Warmup, "warmup", "60s", "fixed: warm-up (excluded)")
	fs.IntVar(&s.Knee, "knee", 0, "ramp: reference knee rate")
	fs.StringVar(&s.Steps, "steps", "10,20,35,50,65,80,90,100,120,150", "ramp: percent steps")
	fs.StringVar(&s.StepDur, "step-dur", "2m", "ramp: duration per step")
	fs.IntVar(&s.Rep, "rep", 1, "repetition number")
	fs.StringVar(&s.Profile, "profile", "none", "none|cpu|allocs")
	fs.IntVar(&s.GOMAXPROCS, "gomaxprocs", 4, "server GOMAXPROCS (identical for all arms)")
	fs.IntVar(&s.PreVUs, "pre-vus", 50, "k6 preAllocatedVUs")
	fs.IntVar(&s.MaxVUs, "max-vus", 400, "k6 maxVUs")
	fs.StringVar(&s.Addr, "addr", "127.0.0.1:8080", "server listen address")
	fs.StringVar(&s.Mgmt, "mgmt", "127.0.0.1:6061", "server management address")
	fs.BoolVar(&s.Raw, "raw", true, "keep gzipped raw k6 samples (needed for pooled percentiles)")
	fs.StringVar(&s.Variant, "variant", "", "binary variant suffix, e.g. minimal")
	fs.Parse(args)
	if s.RunDir == "" || s.Arm == "" {
		return errors.New("-run-dir and -arm are required")
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	res, err := runTrial(ctx, s)
	if res != nil {
		fmt.Fprintf(os.Stderr, "trial %s: %s %v\n", s.dir(), res.Status, res.Reasons)
	}
	return err
}

// TrialResult is the machine-readable outcome consumed by the report tool.
type TrialResult struct {
	Spec       TrialSpec           `json:"spec"`
	Status     string              `json:"status"` // valid | flagged | invalid | error
	Reasons    []string            `json:"reasons"`
	StartedAt  time.Time           `json:"started_at"`
	EndedAt    time.Time           `json:"ended_at"`
	Window     Window              `json:"window"`
	Server     map[string]float64  `json:"server"` // window deltas + derived
	K6         map[string]float64  `json:"k6"`     // gate inputs from the summary
	Host       map[string]float64  `json:"host"`
	Samples    map[string][]Sample `json:"samples"`
	BinarySHA  string              `json:"binary_sha256"`
	FixtureSHA string              `json:"fixture_sha256"`
	Profiles   []ProfileWindow     `json:"profiles,omitempty"`
	// HostTopBefore lists processes using >=5% of one core before the trial.
	HostTopBefore []string `json:"host_top_before,omitempty"`
}

// Window is the measured interval, bounded by two /metrics snapshots taken
// by benchctl: after warm-up ends and after k6 exits (includes k6's drain of
// in-flight requests; drain length = end − (start + duration)).
type Window struct {
	StartUnixNano int64   `json:"start_unix_nano"`
	EndUnixNano   int64   `json:"end_unix_nano"`
	WallSeconds   float64 `json:"wall_s"`
	Nominal       float64 `json:"nominal_s"`
	DrainSeconds  float64 `json:"drain_s"`
}

type ProfileWindow struct {
	Kind        string  `json:"kind"`
	File        string  `json:"file"`
	Seconds     int     `json:"seconds"`
	DecodedOK   float64 `json:"decoded_ok_in_window"`
	StartOffset float64 `json:"start_offset_s"`
}

func runTrial(ctx context.Context, s TrialSpec) (*TrialResult, error) {
	dir := s.dir()
	if _, err := os.Stat(dir); err == nil {
		return nil, fmt.Errorf("%s exists: refusing to overwrite (stale data protection)", dir)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	root, err := repoRoot()
	if err != nil {
		return nil, err
	}
	fixturePath := filepath.Join(root, "fixtures", s.Fixture+".json")
	res := &TrialResult{Spec: s, Status: "valid", StartedAt: time.Now().UTC(), Server: map[string]float64{},
		K6: map[string]float64{}, Host: map[string]float64{}}
	if res.BinarySHA, err = fileSHA256(s.binary()); err != nil {
		return nil, err
	}
	if res.FixtureSHA, err = fileSHA256(fixturePath); err != nil {
		return nil, err
	}
	finish := func(e error) (*TrialResult, error) {
		res.EndedAt = time.Now().UTC()
		if e != nil {
			res.Status = "error"
			res.Reasons = append(res.Reasons, e.Error())
		}
		if werr := writeJSON(filepath.Join(dir, "trial.json"), res); werr != nil && e == nil {
			e = werr
		}
		return res, e
	}
	for _, a := range []string{s.Addr, s.Mgmt} {
		if err := portFree(a); err != nil {
			return finish(err)
		}
	}
	res.Host["idle_cpu_pct_before"], res.HostTopBefore = hostBusyPct()
	res.Host["cpu_speed_limit_before"] = float64(thermal())
	swap0 := swapouts()

	// --- server ---
	srvOut, err := mustCreate(filepath.Join(dir, "server-final-metrics.json"))
	if err != nil {
		return finish(err)
	}
	defer srvOut.Close()
	srvLog, err := mustCreate(filepath.Join(dir, "server.log"))
	if err != nil {
		return finish(err)
	}
	defer srvLog.Close()
	srv := exec.Command(s.binary(), "-addr", s.Addr, "-mgmt-addr", s.Mgmt)
	srv.Env = serverEnv(s.GOMAXPROCS)
	srv.Stdout, srv.Stderr = srvOut, srvLog
	if err := srv.Start(); err != nil {
		return finish(err)
	}
	srvDone := make(chan error, 1)
	go func() { srvDone <- srv.Wait() }()
	killServer := func() {
		srv.Process.Signal(syscall.SIGTERM)
		select {
		case <-srvDone:
		case <-time.After(40 * time.Second):
			srv.Process.Kill()
			<-srvDone
		}
	}
	if err := waitReady(s.Mgmt, 10*time.Second); err != nil {
		killServer()
		return finish(err)
	}
	if b, _, err := httpGet("http://"+s.Mgmt+"/manifest", 5*time.Second); err == nil {
		os.WriteFile(filepath.Join(dir, "server-manifest.json"), b, 0o644)
	}

	// --- k6 ---
	k6Args := []string{"run", "-q", "--no-color",
		"--summary-export", filepath.Join(dir, "k6-summary.json"),
		"--tag", "impl=" + s.Arm, "--tag", "fixture=" + s.Fixture, "--tag", "workload=" + s.Workload,
		"--tag", "rep=" + strconv.Itoa(s.Rep), "--tag", "profile=" + s.Profile,
		"-e", "WORKLOAD=" + s.Workload, "-e", "FIXTURE=" + fixturePath, "-e", "BASE_URL=http://" + s.Addr,
		"-e", "RATE=" + strconv.Itoa(s.Rate), "-e", "DURATION=" + s.Duration, "-e", "WARMUP=" + s.Warmup,
		"-e", "KNEE=" + strconv.Itoa(s.Knee), "-e", "STEPS=" + s.Steps, "-e", "STEP_DUR=" + s.StepDur,
		"-e", "PRE_VUS=" + strconv.Itoa(s.PreVUs), "-e", "MAX_VUS=" + strconv.Itoa(s.MaxVUs)}
	if s.Raw && s.Workload != "smoke" {
		k6Args = append(k6Args, "--out", "csv="+filepath.Join(dir, "k6-samples.csv.gz"))
	}
	k6Args = append(k6Args, filepath.Join(root, "loadtest", "k6", "ingest.js"))
	k6Log, err := mustCreate(filepath.Join(dir, "k6.log"))
	if err != nil {
		killServer()
		return finish(err)
	}
	defer k6Log.Close()
	k6 := exec.CommandContext(ctx, "k6", k6Args...)
	k6.Env = append(os.Environ(), "K6_NO_USAGE_REPORT=true")
	k6.Stdout, k6.Stderr = k6Log, k6Log
	k6Start := time.Now()
	if err := k6.Start(); err != nil {
		killServer()
		return finish(err)
	}
	smp := startSampler(k6Start, 5*time.Second, map[string]int{"server": srv.Process.Pid, "k6": k6.Process.Pid})
	k6Done := make(chan error, 1)
	go func() { k6Done <- k6.Wait() }()

	warm, nominal := windowPlan(s)
	snapErr := make(chan error, 1)
	var m0 map[string]any
	var t0 time.Time
	go func() {
		select {
		case <-time.After(time.Until(k6Start.Add(warm))):
		case <-ctx.Done():
			snapErr <- ctx.Err()
			return
		}
		var b []byte
		var e error
		m0, b, e = snapshot(s.Mgmt)
		t0 = time.Now()
		if e == nil {
			os.WriteFile(filepath.Join(dir, "metrics-window-start.json"), b, 0o644)
			if s.Profile != "none" {
				pw, perr := collectProfile(ctx, s, dir, t0.Sub(k6Start).Seconds())
				if perr != nil {
					e = perr
				} else {
					res.Profiles = append(res.Profiles, pw...)
				}
			}
		}
		snapErr <- e
	}()

	k6Err := <-k6Done
	samples := smp.Stop()
	res.Samples = samples
	if e := <-snapErr; e != nil {
		killServer()
		return finish(fmt.Errorf("window start snapshot/profile: %w", e))
	}
	m1, b1, err := snapshot(s.Mgmt)
	t1 := time.Now()
	if err != nil {
		killServer()
		return finish(fmt.Errorf("window end snapshot: %w", err))
	}
	os.WriteFile(filepath.Join(dir, "metrics-window-end.json"), b1, 0o644)
	if s.Profile != "none" {
		if b, _, err := httpGet("http://"+s.Mgmt+"/debug/pprof/heap", 30*time.Second); err == nil {
			os.WriteFile(filepath.Join(dir, "heap-end.pprof"), b, 0o644)
		}
	}
	killServer()
	if k6Err != nil {
		var ee *exec.ExitError
		if !errors.As(k6Err, &ee) {
			return finish(fmt.Errorf("k6: %w", k6Err))
		}
		// k6 exit codes are not used for gating (no thresholds); record it.
		res.K6["exit_code"] = float64(ee.ExitCode())
	}
	res.Window = Window{StartUnixNano: t0.UnixNano(), EndUnixNano: t1.UnixNano(), WallSeconds: t1.Sub(t0).Seconds(),
		Nominal: nominal.Seconds(), DrainSeconds: t1.Sub(t0).Seconds() - nominal.Seconds()}
	res.Host["cpu_speed_limit_after"] = float64(thermal())
	if swap0 >= 0 {
		res.Host["swapouts_delta"] = float64(swapouts() - swap0)
	}
	deltas(res, m0, m1)
	if err := readK6Summary(res, filepath.Join(dir, "k6-summary.json"), s); err != nil {
		return finish(err)
	}
	finalCounts(res, filepath.Join(dir, "server-final-metrics.json"))
	gate(res)
	return finish(nil)
}

// windowPlan returns when the measured window starts after k6 launch and
// its nominal length.
func windowPlan(s TrialSpec) (time.Duration, time.Duration) {
	d := func(x string) time.Duration {
		v, err := time.ParseDuration(x)
		if err != nil {
			return 0
		}
		return v
	}
	switch s.Workload {
	case "fixed":
		return d(s.Warmup), d(s.Duration)
	case "ramp":
		return 0, time.Duration(len(strings.Split(s.Steps, ","))) * d(s.StepDur)
	}
	return 0, 0
}

// serverEnv: identical runtime environment for every arm. GOGC/GOMEMLIMIT
// are left at defaults (recorded in the manifest) unless set by the caller.
func serverEnv(procs int) []string {
	env := []string{}
	for _, kv := range os.Environ() {
		k, _, _ := strings.Cut(kv, "=")
		if k == "GOMAXPROCS" || k == "GODEBUG" {
			continue
		}
		env = append(env, kv)
	}
	return append(env, "GOMAXPROCS="+strconv.Itoa(procs))
}

func collectProfile(ctx context.Context, s TrialSpec, dir string, offset float64) ([]ProfileWindow, error) {
	const secs = 60
	path := map[string]string{"cpu": "/debug/pprof/profile", "allocs": "/debug/pprof/allocs"}[s.Profile]
	if path == "" {
		return nil, fmt.Errorf("unknown profile %q", s.Profile)
	}
	a, _, err := snapshot(s.Mgmt)
	if err != nil {
		return nil, err
	}
	// allocs?seconds=N returns a delta profile over the window (no manual
	// baseline subtraction needed); profile?seconds=N is the CPU window.
	b, code, err := httpGet(fmt.Sprintf("http://%s%s?seconds=%d", s.Mgmt, path, secs), (secs+30)*time.Second)
	if err != nil || code != 200 {
		return nil, fmt.Errorf("pprof %s: code=%d err=%v", s.Profile, code, err)
	}
	z, _, err := snapshot(s.Mgmt)
	if err != nil {
		return nil, err
	}
	f := s.Profile + ".pprof"
	if err := os.WriteFile(filepath.Join(dir, f), b, 0o644); err != nil {
		return nil, err
	}
	return []ProfileWindow{{Kind: s.Profile, File: f, Seconds: secs, StartOffset: offset,
		DecodedOK: num(z, "decoded_ok") - num(a, "decoded_ok")}}, nil
}

func num(m map[string]any, k string) float64 {
	if m == nil {
		return 0
	}
	switch v := m[k].(type) {
	case float64:
		return v
	}
	return 0
}

// deltas computes window deltas from the two in-process snapshots.
func deltas(r *TrialResult, a, b map[string]any) {
	keys := []string{"received", "decoded_ok", "decode_errors", "validation_failures", "policy_rejects_utf8",
		"policy_rejects_dupkey", "rejected_method", "rejected_ctype", "rejected_size", "read_errors", "responded",
		"sum_read_ns", "sum_policy_ns", "sum_decode_ns", "sum_validate_ns", "sum_handler_ns",
		"rusage_utime_ns", "rusage_stime_ns", "rusage_nvcsw", "rusage_nivcsw",
		"rt/gc/heap/allocs:bytes", "rt/gc/heap/allocs:objects", "rt/gc/heap/frees:objects", "rt/gc/cycles/total:gc-cycles"}
	for _, k := range keys {
		r.Server["d_"+k] = num(b, k) - num(a, k)
	}
	for _, k := range []string{"rt/gc/heap/live:bytes", "rt/gc/heap/goal:bytes", "rt/memory/classes/total:bytes", "rusage_maxrss_bytes", "inflight"} {
		r.Server["end_"+k] = num(b, k)
	}
	r.Server["gc_pause_total_s_window"] = histSum(b, "rt/sched/pauses/total/gc:seconds") - histSum(a, "rt/sched/pauses/total/gc:seconds")
	r.Server["gc_pause_count_window"] = histCount(b, "rt/sched/pauses/total/gc:seconds") - histCount(a, "rt/sched/pauses/total/gc:seconds")
	cpu := (r.Server["d_rusage_utime_ns"] + r.Server["d_rusage_stime_ns"]) / 1e9
	r.Server["cpu_seconds"] = cpu
	if w := r.Window.WallSeconds; w > 0 {
		r.Server["cpu_util_cores"] = cpu / w
		r.Server["decoded_ok_per_s"] = r.Server["d_decoded_ok"] / w
	}
	if ok := r.Server["d_decoded_ok"]; ok > 0 {
		r.Server["cpu_us_per_op"] = cpu * 1e6 / ok
		r.Server["alloc_bytes_per_op"] = r.Server["d_rt/gc/heap/allocs:bytes"] / ok
		r.Server["alloc_objects_per_op"] = r.Server["d_rt/gc/heap/allocs:objects"] / ok
		r.Server["gc_cycles_per_kop"] = r.Server["d_rt/gc/cycles/total:gc-cycles"] * 1000 / ok
	}
}

// histSum approximates the sum of a runtime/metrics histogram using bucket
// lower bounds (conservative); exact per-bucket counts are kept raw.
func histSum(m map[string]any, k string) float64 {
	h, ok := m[k].(map[string]any)
	if !ok {
		return 0
	}
	counts, _ := h["counts"].([]any)
	buckets, _ := h["buckets"].([]any)
	s := 0.0
	for i, c := range counts {
		if i >= len(buckets) {
			break
		}
		lo, err := strconv.ParseFloat(fmt.Sprint(buckets[i]), 64)
		if err != nil || lo < 0 {
			continue
		}
		cv, _ := c.(float64)
		s += cv * lo
	}
	return s
}

func histCount(m map[string]any, k string) float64 {
	h, ok := m[k].(map[string]any)
	if !ok {
		return 0
	}
	counts, _ := h["counts"].([]any)
	s := 0.0
	for _, c := range counts {
		cv, _ := c.(float64)
		s += cv
	}
	return s
}

func finalCounts(r *TrialResult, path string) {
	b, err := os.ReadFile(path)
	if err != nil {
		return
	}
	var m map[string]any
	if json.Unmarshal(b, &m) == nil {
		r.Server["final_received"] = num(m, "received")
		r.Server["final_decoded_ok"] = num(m, "decoded_ok")
		r.Server["final_responded"] = num(m, "responded")
	}
}

// hostBusyPct sums %cpu over all processes, as a fraction of total capacity.
// It also returns the top consumers (command basename only, no arguments, so
// no paths or secrets reach the artifacts) to explain a "host busy" flag.
func hostBusyPct() (float64, []string) {
	out, err := exec.Command("ps", "-A", "-o", "%cpu=,comm=").Output()
	if err != nil {
		return -1, nil
	}
	type p struct {
		cpu  float64
		name string
	}
	var ps []p
	s := 0.0
	for _, line := range strings.Split(string(out), "\n") {
		f := strings.Fields(line)
		if len(f) < 2 {
			continue
		}
		v, _ := strconv.ParseFloat(f[0], 64)
		s += v
		ps = append(ps, p{v, filepath.Base(strings.Join(f[1:], " "))})
	}
	sort.Slice(ps, func(i, j int) bool { return ps[i].cpu > ps[j].cpu })
	var top []string
	for i := 0; i < len(ps) && i < 5 && ps[i].cpu >= 5; i++ {
		top = append(top, fmt.Sprintf("%s %.0f%%", ps[i].name, ps[i].cpu))
	}
	return s / float64(runtime.NumCPU()), top
}
