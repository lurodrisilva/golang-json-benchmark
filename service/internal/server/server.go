// Package server implements the shared HTTP application. Both binaries call
// Run with their codec; everything else (routing, limits, timeouts, error
// contract, instrumentation, lifecycle) is identical.
package server

import (
	"context"
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"
	rtmetrics "runtime/metrics"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/lurodrisilva/golang-json-encoding-benchmark/service/internal/codec"
	"github.com/lurodrisilva/golang-json-encoding-benchmark/service/internal/metrics"
	"github.com/lurodrisilva/golang-json-encoding-benchmark/service/internal/model"
	"github.com/lurodrisilva/golang-json-encoding-benchmark/service/internal/policy"
)

//go:embed selftest.json
var selftest []byte

// Config is the full runtime configuration. Defaults are identical for both binaries.
type Config struct {
	Addr          string
	MgmtAddr      string
	BodyLimit     int64
	ShutdownGrace time.Duration
}

func parseFlags(args []string) (Config, error) {
	fs := flag.NewFlagSet("server", flag.ContinueOnError)
	var c Config
	fs.StringVar(&c.Addr, "addr", "127.0.0.1:8080", "ingest listener address")
	fs.StringVar(&c.MgmtAddr, "mgmt-addr", "127.0.0.1:6061", "management listener (loopback only)")
	fs.Int64Var(&c.BodyLimit, "body-limit", DefaultBodyLimit, "max request body bytes")
	fs.DurationVar(&c.ShutdownGrace, "shutdown-grace", 30*time.Second, "graceful shutdown timeout")
	if err := fs.Parse(args); err != nil {
		return c, err
	}
	host, _, err := net.SplitHostPort(c.MgmtAddr)
	if err != nil {
		return c, fmt.Errorf("mgmt-addr: %w", err)
	}
	if ip := net.ParseIP(host); host != "localhost" && (ip == nil || !ip.IsLoopback()) {
		return c, fmt.Errorf("mgmt-addr %q must be a loopback address", c.MgmtAddr)
	}
	return c, nil
}

// NewIngestHandler returns the /ingest handler mux (exported for tests).
func NewIngestHandler(c codec.Codec, m *metrics.Set, bodyLimit int64) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/ingest", &ingestHandler{codec: c, m: m, bodyLimit: bodyLimit})
	return mux
}

// SelfTest decodes the embedded fixture through the full pipeline.
func SelfTest(c codec.Codec) error {
	if r := policy.Check(selftest); r != policy.OK {
		return fmt.Errorf("selftest policy: %v", r)
	}
	var b model.IngestBatch
	if err := c.Unmarshal(selftest, &b); err != nil {
		return fmt.Errorf("selftest decode: %w", err)
	}
	if err := b.Validate(); err != nil {
		return fmt.Errorf("selftest validate: %w", err)
	}
	if b.BatchID != "selftest" || len(b.Events) != 1 || b.Events[0].User.Name != "José é 😀" {
		return errors.New("selftest: unexpected decoded values")
	}
	return nil
}

// Manifest describes the running binary (build info + runtime settings).
func Manifest(c codec.Codec) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "impl\t%s\n", c.Name())
	fmt.Fprintf(&sb, "instrumentation\t%s\n", InstrumentationMode)
	fmt.Fprintf(&sb, "goversion\t%s\n", runtime.Version())
	fmt.Fprintf(&sb, "goos\t%s\ngoarch\t%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Fprintf(&sb, "gomaxprocs\t%d\nnumcpu\t%d\n", runtime.GOMAXPROCS(0), runtime.NumCPU())
	for _, k := range []string{"GOGC", "GOMEMLIMIT", "GODEBUG", "GOEXPERIMENT", "GOMAXPROCS"} {
		fmt.Fprintf(&sb, "env.%s\t%s\n", k, os.Getenv(k))
	}
	// Read effective GOGC/GOMEMLIMIT without mutating GC settings.
	rs := []rtmetrics.Sample{{Name: "/gc/gogc:percent"}, {Name: "/gc/gomemlimit:bytes"}}
	rtmetrics.Read(rs)
	fmt.Fprintf(&sb, "gc_percent\t%d\n", rs[0].Value.Uint64())
	fmt.Fprintf(&sb, "memory_limit\t%d\n", rs[1].Value.Uint64())
	if bi, ok := debug.ReadBuildInfo(); ok {
		fmt.Fprintf(&sb, "path\t%s\n", bi.Path)
		for _, d := range bi.Deps {
			fmt.Fprintf(&sb, "dep\t%s\t%s\t%s\n", d.Path, d.Version, d.Sum)
		}
		for _, s := range bi.Settings {
			fmt.Fprintf(&sb, "build\t%s\t%s\n", s.Key, s.Value)
		}
	}
	return sb.String()
}

// Run starts both listeners, becomes ready after a self-test, and shuts down
// gracefully on SIGINT/SIGTERM. It returns the process exit code.
func Run(c codec.Codec, args []string) int {
	start := time.Now()
	cfg, err := parseFlags(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	if err := SelfTest(c); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	m := new(metrics.Set)
	var ready atomic.Bool

	srv := &http.Server{
		Handler:           NewIngestHandler(c, m, cfg.BodyLimit),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    1 << 16,
	}
	var p http.Protocols
	p.SetHTTP1(true) // HTTP/1.1 only (no h2c), identical for both binaries
	srv.Protocols = &p

	manifest := Manifest(c)
	mux := http.NewServeMux()
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, _ *http.Request) {
		if ready.Load() {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusServiceUnavailable)
	})
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(m.Snapshot(make([]byte, 0, 4096), c.Name(), start))
	})
	mux.HandleFunc("/manifest", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write([]byte(manifest))
	})
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	mgmt := &http.Server{Handler: mux, ReadHeaderTimeout: 5 * time.Second}

	ln, err := net.Listen("tcp", cfg.Addr)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	mln, err := net.Listen("tcp", cfg.MgmtAddr)
	if err != nil {
		ln.Close()
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	errc := make(chan error, 2)
	go func() { errc <- srv.Serve(ln) }()
	go func() { errc <- mgmt.Serve(mln) }()
	ready.Store(true)
	fmt.Fprintf(os.Stderr, "ready impl=%s addr=%s mgmt=%s\n", c.Name(), ln.Addr(), mln.Addr())

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	code := 0
	select {
	case s := <-sig:
		fmt.Fprintf(os.Stderr, "signal %v: shutting down\n", s)
	case err := <-errc:
		fmt.Fprintln(os.Stderr, "serve:", err)
		code = 1
	}

	ready.Store(false)
	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownGrace)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "shutdown:", err)
		code = 1
	}
	// Final counters go to stdout as one JSON line (after ingest drain).
	os.Stdout.Write(append(m.Snapshot(nil, c.Name(), start), '\n'))
	mgmt.Shutdown(ctx)
	return code
}
