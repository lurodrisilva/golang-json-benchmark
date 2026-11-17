# Runbook

Every command runs from the repository root. You need Go 1.27.4 (pinned in `go.work` / `toolchain`), k6 v2.4.0 on `PATH`, and macOS or Linux.

## 1. Correctness (always first)

```sh
make vet               # gofmt + go vet (service, tools, -tags minimal)
make fixtures-verify   # SHA-256 of fixtures + golden files
make test              # unit/integration/conformance, default AND GOEXPERIMENT=nojsonv2, plus -tags minimal
make race              # race detector, both GOEXPERIMENT settings (never during perf runs)
make conformance       # reports/conformance/conformance-{default,nojsonv2}.tsv
make fuzz              # 60 s fuzz of the shared policy scanner
```

`make fixtures` regenerates the fixtures from seed 20261115. It must produce **no** `git diff`, because CI enforces this.

## 2. Build

```sh
make build             # bin/{stdlib,goccy,stdlib-legacy,goccy-legacy}-server (+ -minimal-), bin/build.json
make env               # host/toolchain manifest (also written into every run)
```

`build.json` records, for each binary: sha256, `go version -m`, GOEXPERIMENT, and the `encoding/json` GoFiles that were compiled in (`v2_*.go` means v2-backed, `decode.go` means legacy). `benchctl run` refuses binaries whose hash differs from `build.json`. It also copies them into the run directory so that profiles can be symbolized later.

## 3. Decoder microbenchmarks (separate from HTTP)

```sh
make microbench COUNT=10   # reports/bench/{default,nojsonv2}.txt + benchstat-*.txt
```

## 4. HTTP experiments

```sh
make smoke                                           # ~11 min end-to-end pipeline check; NOT a result
make plan MATRIX=loadtest/matrix/pilot.json          # dry-run: order + time estimate
make run  MATRIX=loadtest/matrix/pilot.json          # P6 pilot (~3 h): ramps to find each arm's knee
```

**Calibrating the authoritative matrix (needs user approval):**

1. From the pilot report §5, take the knee of the **slower** arm for each fixture (and the legacy pair).
2. Copy `loadtest/matrix/authoritative.template.json` to `authoritative.json`. Set `rate` to round(0.5 × slower knee) for the first F1 cell and to round(0.8 × slower knee) for the second; set R50 for the other fixed cells. Set each ramp `knee` to the slower arm's knee. The profiling and minimal cells use R50.
3. Check that R50 gives ≥ 10,000 requests in the trial duration, or P99 will be reported as "insufficient samples".
4. Run `make plan MATRIX=loadtest/matrix/authoritative.json` and get the plan approved.
5. Run `make run MATRIX=loadtest/matrix/authoritative.json`. The template's zero rates are refused by design, so an uncalibrated run cannot start.

### Host preparation (M3, single host)
- Plug into AC power, close other apps, and disable Spotlight indexing / Time Machine for the run.
- Run `pmset -g therm` before you start. Every trial records `CPU_Speed_Limit` and fails or flags below 100.
- Don't use the machine while a matrix is running. Trials with more than 10% unexplained host CPU are flagged.

## 5. Report

```sh
make report                         # latest run
make report RUN=results/<run-id>    # specific run
make report RUN=results/<decode-run>,results/<encode-run>   # combined report (first run holds the output)
```

Encode cells use `"op": "encode"` in the matrix (for example `loadtest/matrix/e2e-encode.json`). benchctl starts both servers with `-encode-fixture` and sets k6 `OP=encode` and `EXPECT_BYTES` (taken from the server manifest). Trials are written to `enc-<load>` directories. The report keys every cell by op and never compares across ops.

Outputs `<run>/report/{report.md,report.html,trials.csv,cells.json,pprof-*.txt}`.

## 6. Artifact layout

```
results/<UTC yyyymmddThhmmssZ>-<gitsha7>-<matrixsha8>/
  matrix.json env.json build.json fixtures.SHA256SUMS index.jsonl bin/
  trials/<arm[+variant]>/<fixture>/<fixed-rN|ramp-kN|smoke>[-prof-cpu|-prof-allocs]/rep-NN/
    trial.json  k6-summary.json  k6-samples.csv.gz  k6.log
    server.log  server-manifest.json  server-final-metrics.json
    metrics-window-start.json  metrics-window-end.json  [cpu.pprof|allocs.pprof heap-end.pprof]
```

A run directory is never reused, and a trial directory that already exists is refused (stale-data protection). `results/` is gitignored. Publish curated reports, not raw runs. Payloads are synthetic and contain no secrets. The env manifest redacts the hostname.

## 7. Failure handling

- When a trial fails to start, isn't ready within 10 s, or k6 exits non-zero, the error is recorded on that trial's line in `index.jsonl`. After 3 consecutive errors the run aborts.
- Ctrl-C stops the run and keeps all completed trials. Servers are sent SIGTERM, and after the grace period SIGKILL.
- Ports 8080/6061 must be free before each trial, or the trial errors.
