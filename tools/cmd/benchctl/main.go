// Command benchctl orchestrates builds, trials, and whole benchmark runs.
//
//	benchctl build  [-out bin]                       build all four arms + build.json
//	benchctl env    [-out file]                      environment manifest JSON
//	benchctl trial  -arm A -fixture F -workload W …  one trial (one server process)
//	benchctl run    -matrix loadtest/matrix/<m>.json one full run (ordered trials)
//
// Every artifact lives under results/<run-id>/…; benchctl refuses to write
// into an existing directory (stale-data protection).
package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		usage()
	}
	var err error
	switch os.Args[1] {
	case "build":
		err = cmdBuild(os.Args[2:])
	case "env":
		err = cmdEnv(os.Args[2:])
	case "trial":
		err = cmdTrial(os.Args[2:])
	case "run":
		err = cmdRun(os.Args[2:])
	default:
		usage()
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "benchctl:", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: benchctl build|env|trial|run [flags]")
	os.Exit(2)
}
