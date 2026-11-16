package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Sample is one periodic observation of a process from outside it (ps), so it
// works for both the server and k6 with zero in-process overhead.
type Sample struct {
	T       float64 `json:"t_s"`      // seconds since trial start
	CPUPct  float64 `json:"cpu_pct"`  // ps %cpu (100 = one core)
	RSSKiB  int64   `json:"rss_kib"`  // resident set size
	CPUTime float64 `json:"cpu_time"` // cumulative user+sys seconds (ps time)
}

// sampler polls `ps` for a set of pids every interval until stopped.
type sampler struct {
	mu      sync.Mutex
	samples map[string][]Sample
	stop    chan struct{}
	done    chan struct{}
}

func startSampler(start time.Time, every time.Duration, pids map[string]int) *sampler {
	s := &sampler{samples: map[string][]Sample{}, stop: make(chan struct{}), done: make(chan struct{})}
	go func() {
		defer close(s.done)
		t := time.NewTicker(every)
		defer t.Stop()
		for {
			for name, pid := range pids {
				if smp, ok := psSample(pid); ok {
					smp.T = time.Since(start).Seconds()
					s.mu.Lock()
					s.samples[name] = append(s.samples[name], smp)
					s.mu.Unlock()
				}
			}
			select {
			case <-s.stop:
				return
			case <-t.C:
			}
		}
	}()
	return s
}

func (s *sampler) Stop() map[string][]Sample {
	close(s.stop)
	<-s.done
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.samples
}

func psSample(pid int) (Sample, bool) {
	out, err := exec.Command("ps", "-o", "%cpu=,rss=,time=", "-p", strconv.Itoa(pid)).Output()
	if err != nil {
		return Sample{}, false
	}
	f := strings.Fields(string(out))
	if len(f) < 3 {
		return Sample{}, false
	}
	cpu, _ := strconv.ParseFloat(f[0], 64)
	rss, _ := strconv.ParseInt(f[1], 10, 64)
	return Sample{CPUPct: cpu, RSSKiB: rss, CPUTime: parsePsTime(f[2])}, true
}

// parsePsTime parses ps "time" ([[dd-]hh:]mm:ss[.cc]).
func parsePsTime(s string) float64 {
	days := 0.0
	if d, rest, ok := strings.Cut(s, "-"); ok {
		days, _ = strconv.ParseFloat(d, 64)
		s = rest
	}
	parts := strings.Split(s, ":")
	total := 0.0
	for _, p := range parts {
		v, _ := strconv.ParseFloat(p, 64)
		total = total*60 + v
	}
	return days*86400 + total
}

// thermal returns the macOS CPU_Speed_Limit (100 = unthrottled) or -1.
func thermal() int {
	if runtime.GOOS != "darwin" {
		return -1
	}
	out, err := exec.Command("pmset", "-g", "therm").Output()
	if err != nil {
		return -1
	}
	sc := bufio.NewScanner(strings.NewReader(string(out)))
	for sc.Scan() {
		line := sc.Text()
		if strings.Contains(line, "CPU_Speed_Limit") {
			if _, v, ok := strings.Cut(line, "="); ok {
				n, err := strconv.Atoi(strings.TrimSpace(v))
				if err == nil {
					return n
				}
			}
		}
	}
	return 100 // no warning lines printed = no limit recorded
}

// swapouts returns cumulative pageouts/swapouts (darwin vm_stat "Swapouts").
func swapouts() int64 {
	if runtime.GOOS != "darwin" {
		return -1
	}
	out, err := exec.Command("vm_stat").Output()
	if err != nil {
		return -1
	}
	for _, line := range strings.Split(string(out), "\n") {
		if strings.HasPrefix(line, "Swapouts:") {
			v := strings.TrimSuffix(strings.TrimSpace(strings.TrimPrefix(line, "Swapouts:")), ".")
			n, _ := strconv.ParseInt(v, 10, 64)
			return n
		}
	}
	return -1
}

func httpGet(url string, timeout time.Duration) ([]byte, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	return b, resp.StatusCode, err
}

func waitReady(mgmt string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if _, code, err := httpGet("http://"+mgmt+"/readyz", time.Second); err == nil && code == 200 {
			return nil
		}
		time.Sleep(50 * time.Millisecond)
	}
	return fmt.Errorf("server not ready on %s within %s", mgmt, timeout)
}

// portFree fails fast if a previous trial's process is still bound.
func portFree(addr string) error {
	if _, _, err := httpGet("http://"+addr+"/", 300*time.Millisecond); err == nil {
		return fmt.Errorf("%s already serving — stale process? refusing to start", addr)
	}
	return nil
}

func snapshot(mgmt string) (map[string]any, []byte, error) {
	b, code, err := httpGet("http://"+mgmt+"/metrics", 5*time.Second)
	if err != nil {
		return nil, nil, err
	}
	if code != 200 {
		return nil, nil, fmt.Errorf("/metrics status %d", code)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, nil, err
	}
	return m, b, nil
}

func mustCreate(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
}
