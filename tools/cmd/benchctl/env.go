package main

import (
	"flag"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// envManifest captures host + toolchain facts. Each command's raw output is
// kept verbatim; nothing is parsed or interpreted here.
func envManifest() map[string]any {
	run := func(name string, a ...string) string {
		o, err := exec.Command(name, a...).CombinedOutput()
		if err != nil {
			return "ERROR: " + err.Error() + ": " + strings.TrimSpace(string(o))
		}
		return strings.TrimSpace(string(o))
	}
	m := map[string]any{
		"captured_at": time.Now().UTC().Format(time.RFC3339),
		"goos":        runtime.GOOS,
		"goarch":      runtime.GOARCH,
		"numcpu":      runtime.NumCPU(),
		"go_version":  run("go", "version"),
		"go_env":      goEnv(),
		"k6_version":  run("k6", "version"),
		"uname":       run("uname", "-a"),
		"hostname":    "redacted",
		"env": map[string]string{
			"GOGC": os.Getenv("GOGC"), "GOMEMLIMIT": os.Getenv("GOMEMLIMIT"),
			"GOMAXPROCS": os.Getenv("GOMAXPROCS"), "GODEBUG": os.Getenv("GODEBUG"),
		},
	}
	switch runtime.GOOS {
	case "darwin":
		m["cpu"] = run("sysctl", "-n", "machdep.cpu.brand_string")
		m["perflevel0_logicalcpu"] = run("sysctl", "-n", "hw.perflevel0.logicalcpu")
		m["perflevel1_logicalcpu"] = run("sysctl", "-n", "hw.perflevel1.logicalcpu")
		m["memsize"] = run("sysctl", "-n", "hw.memsize")
		m["os"] = run("sw_vers")
		m["pmset_therm"] = run("pmset", "-g", "therm")
		m["power_source"] = run("pmset", "-g", "batt")
	case "linux":
		m["lscpu"] = run("lscpu")
		m["meminfo"] = run("head", "-3", "/proc/meminfo")
		m["governor"] = run("sh", "-c", "cat /sys/devices/system/cpu/cpu*/cpufreq/scaling_governor 2>/dev/null | sort | uniq -c")
	}
	return m
}

func cmdEnv(args []string) error {
	fs := flag.NewFlagSet("env", flag.ExitOnError)
	out := fs.String("out", "", "write to file (default stdout)")
	fs.Parse(args)
	if *out == "" {
		return writeJSON("/dev/stdout", envManifest())
	}
	return writeJSON(*out, envManifest())
}
