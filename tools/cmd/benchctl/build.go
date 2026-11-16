package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// Arm is one built binary. Pair D = default toolchain build; Pair L = the
// same sources built with GOEXPERIMENT=nojsonv2 (legacy encoding/json). Both
// members of a pair always share GOEXPERIMENT.
type Arm struct {
	Name         string `json:"name"`
	Cmd          string `json:"cmd"`
	Pair         string `json:"pair"`
	Codec        string `json:"codec"`
	GOEXPERIMENT string `json:"goexperiment"`
}

var arms = []Arm{
	{"stdlib", "./cmd/stdlib-server", "D", "stdlib", ""},
	{"goccy", "./cmd/goccy-server", "D", "goccy", ""},
	{"stdlib-legacy", "./cmd/stdlib-server", "L", "stdlib", "nojsonv2"},
	{"goccy-legacy", "./cmd/goccy-server", "L", "goccy", "nojsonv2"},
}

func armByName(n string) (Arm, error) {
	for _, a := range arms {
		if a.Name == n {
			return a, nil
		}
	}
	return Arm{}, fmt.Errorf("unknown arm %q", n)
}

type BuiltArm struct {
	Arm
	Binary      string   `json:"binary"`
	SHA256      string   `json:"sha256"`
	BuildFlags  []string `json:"build_flags"`
	GoVersionM  string   `json:"go_version_m"`
	JSONGoFiles string   `json:"encoding_json_gofiles"`
	Tags        string   `json:"tags"`
}

// buildFlags are identical for every arm; the only per-arm difference is the
// cmd package (codec) and, between pairs, GOEXPERIMENT.
var buildFlags = []string{"-trimpath", "-buildvcs=true"}

func cmdBuild(args []string) error {
	fs := flag.NewFlagSet("build", flag.ExitOnError)
	out := fs.String("out", "bin", "output directory")
	tags := fs.String("tags", "", "build tags (e.g. minimal) applied to ALL arms")
	only := fs.String("arms", "", "comma-separated subset of arms (default all)")
	fs.Parse(args)
	root, err := repoRoot()
	if err != nil {
		return err
	}
	outDir, _ := filepath.Abs(*out)
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	var built []BuiltArm
	for _, a := range arms {
		if *only != "" && !contains(strings.Split(*only, ","), a.Name) {
			continue
		}
		name := a.Name + "-server"
		if *tags != "" {
			name = a.Name + "-" + strings.ReplaceAll(*tags, ",", "_") + "-server"
		}
		bin := filepath.Join(outDir, name)
		flags := append([]string{"build"}, buildFlags...)
		if *tags != "" {
			flags = append(flags, "-tags", *tags)
		}
		flags = append(flags, "-o", bin, a.Cmd)
		cmd := exec.Command("go", flags...)
		cmd.Dir = filepath.Join(root, "service")
		cmd.Env = armEnv(a)
		cmd.Stdout, cmd.Stderr = os.Stderr, os.Stderr
		fmt.Fprintf(os.Stderr, "build %s (GOEXPERIMENT=%q) → %s\n", a.Name, a.GOEXPERIMENT, bin)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("build %s: %w", a.Name, err)
		}
		vm, _ := exec.Command("go", "version", "-m", bin).Output()
		gl := exec.Command("go", "list", "-f", "{{.GoFiles}}", "encoding/json")
		gl.Env = armEnv(a)
		gf, _ := gl.Output()
		sum, err := fileSHA256(bin)
		if err != nil {
			return err
		}
		built = append(built, BuiltArm{a, bin, sum, flags[1 : len(flags)-3], string(vm), strings.TrimSpace(string(gf)), *tags})
	}
	// Merge with an existing build.json in the same directory (e.g. a later
	// `-tags minimal` build) so it always describes every binary present.
	if old, err := os.ReadFile(filepath.Join(outDir, "build.json")); err == nil {
		var prev struct {
			Arms []BuiltArm `json:"arms"`
		}
		if json.Unmarshal(old, &prev) == nil {
			fresh := map[string]bool{}
			for _, b := range built {
				fresh[b.Binary] = true
			}
			for _, p := range prev.Arms {
				if !fresh[p.Binary] {
					built = append(built, p)
				}
			}
		}
	}
	manifest := map[string]any{
		"built_at":   time.Now().UTC().Format(time.RFC3339),
		"go_env":     goEnv(),
		"host_goos":  runtime.GOOS,
		"arms":       built,
		"git":        gitInfo(root),
		"cgo_enable": "0",
	}
	return writeJSON(filepath.Join(outDir, "build.json"), manifest)
}

// armEnv pins the build environment: CGO off, GOFLAGS cleared, explicit
// GOEXPERIMENT (empty string = toolchain default).
func armEnv(a Arm) []string {
	env := []string{}
	for _, kv := range os.Environ() {
		k, _, _ := strings.Cut(kv, "=")
		switch k {
		case "GOEXPERIMENT", "CGO_ENABLED", "GOFLAGS", "GOOS", "GOARCH":
			continue
		}
		env = append(env, kv)
	}
	return append(env, "CGO_ENABLED=0", "GOFLAGS=", "GOEXPERIMENT="+a.GOEXPERIMENT)
}

func goEnv() map[string]string {
	out, err := exec.Command("go", "env", "-json", "GOVERSION", "GOOS", "GOARCH", "GOAMD64", "GOARM64",
		"GOEXPERIMENT", "CGO_ENABLED", "GOFLAGS", "GOTOOLCHAIN", "GOWORK").Output()
	m := map[string]string{}
	if err == nil {
		json.Unmarshal(out, &m)
	}
	return m
}

func gitInfo(root string) map[string]string {
	run := func(a ...string) string {
		c := exec.Command("git", a...)
		c.Dir = root
		o, _ := c.Output()
		return strings.TrimSpace(string(o))
	}
	return map[string]string{
		"commit": run("rev-parse", "HEAD"),
		"dirty":  fmt.Sprint(run("status", "--porcelain") != ""),
		"branch": run("rev-parse", "--abbrev-ref", "HEAD"),
	}
}

func repoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.work")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.work not found above working directory")
		}
		dir = parent
	}
}

func fileSHA256(p string) (string, error) {
	f, err := os.Open(p)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func writeJSON(path string, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(b, '\n'), 0o644)
}

func contains(xs []string, s string) bool {
	for _, x := range xs {
		if x == s {
			return true
		}
	}
	return false
}
