#!/bin/zsh
# Reruns of the 5 invalid trials of run 20261116T115723Z-af6c2f3-06dccad4 (user-approved 2026-11-17).
# Invalid trials are kept; reruns get the next rep number so both stay visible in the report.
set -u
cd ~/src/sdk-benchmark/golang-json-encoding-benchmark
R=results/20261116T115723Z-af6c2f3-06dccad4
B=$R/bin
C=(-run-dir $R -bin $B -gomaxprocs 4 -pre-vus 100 -max-vus 400)
S=10,20,35,50,65,80,90,100,120,150,175,200
run() { echo "[$(date +%H:%M:%S)] $*"; bin/benchctl trial "${C[@]}" "$@"; echo "exit $?"; sleep 120; }
run -arm goccy         -fixture f1-64k  -workload fixed -rate 2000 -duration 6m -warmup 60s -rep 6
run -arm stdlib        -fixture f1-64k  -workload ramp  -knee 5000 -steps $S -step-dur 2m -rep 2
run -arm stdlib        -fixture f2-256k -workload fixed -rate 1000 -duration 5m -warmup 60s -rep 6
run -arm stdlib        -fixture f2-256k -workload ramp  -knee 1250 -steps $S -step-dur 2m -rep 2
run -arm stdlib-legacy -fixture f1-64k  -workload fixed -rate 1625 -duration 6m -warmup 60s -rep 6
echo DONE
