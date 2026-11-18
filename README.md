# golang-json-encoding-benchmark

A reproducible comparison of HTTP request **deserialization** (decode, `POST /ingest`) and response **serialization** (encode, `GET /batch`) performance between Go's `encoding/json` and [`github.com/goccy/go-json`](https://github.com/goccy/go-json).

**Status:** decode results (run `20261116T115723Z-af6c2f3-06dccad4`) and the combined decode + encode report (encode runs `20261117T144443Z-34f080e-3f7a4e65`, `20261118T102549Z-4f2c938-92dbc890`, `20261118T113550Z-65180c7-98d18137`) are published under [`reports/runs/`](reports/runs/). All results are pilot-grade: single Apple M3 host, k6 and server co-located over loopback. Libraries are compared only within the same operation, payload and load; encode and decode are never compared with each other.

## What is compared
Two separately built `net/http` servers, `stdlib-server` and `goccy-server`, run identical code apart from one `codec.Unmarshal` call. They expose `POST /ingest`, read the body into a bounded buffer, apply a shared strict-input policy, decode it into a fresh typed `IngestBatch`, validate it, and return `202` with an empty body. Under go1.27 the default `encoding/json` is v2-backed, so a second pair built with `GOEXPERIMENT=nojsonv2` is reported separately (ADR 0002).

For encode, started with `-encode-fixture`, each binary also serves `GET /batch`. It `Marshal`s a shared, digest-checked fixture value and writes it with an exact `Content-Length`. Libraries are compared only within the same operation, fixture and workload, never encode against decode.

## Results

All numbers below come from the published report [`reports/runs/20261117T144443Z-34f080e-3f7a4e65/report.md`](reports/runs/20261117T144443Z-34f080e-3f7a4e65/report.md) and are **pilot-grade**: one machine, with the load generator and server on the same host. Percentages are `(stdlib − goccy) / stdlib`, so a positive value means goccy used less. Each table cell is the median across trials. Confidence intervals, tail percentiles and per-trial data are in the full report.

### Executive summary
- On this hardware, `goccy/go-json` v0.12.0 was faster than Go 1.27.4's `encoding/json` in every measured case, for both decoding (Unmarshal) and encoding (Marshal).
- **Codec time.** Without HTTP, goccy took about 67–68% less time to Unmarshal and 59–62% less to Marshal. Under HTTP load, the codec call itself took 68–70% less (decode) and 51–62% less (encode).
- **Request latency and CPU.** End-to-end HTTP request latency was 44–58% lower, and CPU per request 40–53% lower. The network and shared server code dilute the codec gap.
- **Throughput.** goccy sustained about 2.4× the request rate on 64 KiB payloads, for both decode and encode.
- **Memory depends on the operation.** goccy allocated 7–28% fewer bytes per request when encoding. When decoding, it was about equal at 64–256 KiB but allocated 9–10% *more* at 1 MiB.
- **Behaviour.** Both libraries produced identical results on the 57-case decode and 33-case encode correctness suites, with 0 deviations. Encoded output was byte-identical.
- **Caveat.** With only one host, some differences could not be fully attributed to the library: identical shared code also ran at different speeds between arms. Treat these numbers as direction and order of magnitude, not production figures.

### Testing method, tooling and hardware
| Aspect | Setup |
|---|---|
| Applications | Two `net/http` binaries (`stdlib-server`, `goccy-server`), identical apart from one codec call. Decode: `POST /ingest` → `202`. Encode: `GET /batch` → `200`. A fresh destination per request, plain `Unmarshal`/`Marshal` only, no pooling or library options. |
| Payloads | Deterministic synthetic fixtures (seed 20261115): 64 KiB (65,741 B), 256 KiB (262,181 B), 1 MiB (1,048,958 B). Nested objects, arrays, numbers, booleans, nulls, escapes and Unicode, all decoded into a typed model. |
| Correctness first | Golden digests, a v1-spec corpus (57 decode / 33 encode cases), HTTP contract tests and the race detector, all run under both `GOEXPERIMENT` settings before any load. |
| Load model | k6 open model (constant arrival rate), HTTP/1.1 keep-alive, no TLS or compression, loopback. 60 s warm-up excluded, then 5–6 min measured per trial. Rates are 50% and 80% of the slower library's saturation point. |
| Repetitions | 5+ independent trials per library per cell. The server restarts each trial, run order is shuffled (ABBA) and there is a 120 s cool-down. The trial is the statistical unit, with a bootstrap 95% CI on the difference of per-trial medians. |
| Validity gates | A trial is invalid on dropped iterations, < 99% achieved rate, failed responses > 0.1%, transport errors, k6 CPU overload, thermal throttling or swapping. Of 176 trials, 124 were valid, 38 flagged (kept) and 14 invalid or errored (excluded, but listed). |
| Instrumentation | `Server-Timing` per phase (read, policy, codec, validate, write), recorded as k6 Trends. Server CPU via rusage, memory and GC via `runtime/metrics`, pprof CPU and allocation profiles in separate trials. |
| Tooling | Go 1.27.4 (default `encoding/json` is v2-backed; legacy engine via `GOEXPERIMENT=nojsonv2`), `github.com/goccy/go-json` v0.12.0, k6 v2.4.0, `benchstat`, `go tool pprof`, and the repo's `benchctl` and `report` tools. |
| Hardware | Apple M3 (4 performance + 4 efficiency cores), 24 GiB RAM, macOS. `GOMAXPROCS=4` for the server. k6 ran on the same machine, and macOS offers no core pinning. |

### Overall results
Ranges cover every fixed-load cell (Pair D, default toolchain) and the three payload sizes.

| Metric | Unmarshal (decode): goccy vs stdlib | Marshal (encode): goccy vs stdlib |
|---|---|---|
| Codec-only microbenchmark time (no HTTP) | 67–68% less | 59–62% less |
| Codec call time under HTTP load | 68–70% less | 51–62% less |
| HTTP request latency (k6 median) | 45–51% lower | 44–58% lower |
| CPU per request | 43–48% less | 40–53% less |
| Bytes allocated per request | 3–4% fewer at 64 KiB, about equal at 256 KiB, 9–10% more at 1 MiB | 7–28% fewer |
| Highest sustained rate, 64 KiB (single ramp) | 2,500 → 6,000 rps | 4,500 → 10,800 rps |
| Highest sustained rate, 256 KiB (single ramp) | 1,250 → 1,000 rps (ramp not trusted) | 2,250 → 3,938 rps |
| Highest sustained rate, 1 MiB (single ramp) | 320 → ≥ 480 rps (goccy not saturated) | 900 → 1,200 rps |
| Correctness deviations | 0 / 0 | 0 / 0 (byte-identical output) |

### Marshal (encode)
`GET /batch`, default toolchain. Values are stdlib → goccy, with the change in brackets.

| Payload @ rate | Trials (stdlib / goccy) | HTTP median (ms) | Marshal median (ms) | CPU per request (µs) | Bytes allocated per request |
|---|---|---|---|---|---|
| Microbench 64 KiB | 10 / 10 | — | 0.187 → 0.072 (62%) | — | 16% fewer |
| Microbench 256 KiB | 10 / 10 | — | 0.759 → 0.308 (60%) | — | 16% fewer |
| Microbench 1 MiB | 10 / 10 | — | 3.06 → 1.27 (59%) | — | 16% fewer |
| 64 KiB @ 3,600 rps | 2 / 2 | 0.286 → 0.121 (58%) | 0.200 → 0.077 (62%) | 283 → 143 (50%) | 16% fewer |
| 64 KiB @ 4,500 rps | 5 / 6 | 0.276 → 0.121 (56%) | 0.204 → 0.077 (62%) | 276 → 141 (49%) | 81.0K → 68.1K (16%) |
| 64 KiB @ 7,200 rps* | 2 / 4 | 0.249 → 0.122 (51%) | 0.201 → 0.077 (62%) | 280 → 145 (49%) | 16% fewer |
| 256 KiB @ 1,125 rps | 5 / 5 | 1.062 → 0.548 (48%) | 0.787 → 0.332 (58%) | 982 → 519 (47%) | 371K → 308K (17%) |
| 256 KiB @ 1,800 rps | 5 / 5 | 1.019 → 0.573 (44%) | 0.812 → 0.311 (62%) | 1,020 → 475 (53%) | 341K → 290K (15%) |
| 1 MiB @ 300 rps | 5 / 5 | 3.780 → 2.090 (45%) | 3.244 → 1.599 (51%) | 4,006 → 2,406 (40%) | 3.64M → 2.61M (28%) |
| 1 MiB @ 480 rps | 5 / 6 | 3.815 → 1.748 (54%) | 3.395 → 1.298 (62%) | 3,966 → 1,961 (51%) | 2.59M → 2.40M (7%) |

\* 7,200 rps is above stdlib's 64 KiB encode saturation point (4,500 rps), so most stdlib trials there were invalid. Response sizes were identical for both libraries: 65,008 / 260,564 / 1,044,267 bytes.

### Unmarshal (decode)
`POST /ingest`. Values are stdlib → goccy, with the change in brackets. A negative change means goccy used more.

| Payload @ rate | Trials (stdlib / goccy) | HTTP median (ms) | Unmarshal median (ms) | CPU per request (µs) | Bytes allocated per request |
|---|---|---|---|---|---|
| Microbench 64 KiB | 10 / 10 | — | 0.390 → 0.129 (67%) | — | 5–8% fewer |
| Microbench 256 KiB | 10 / 10 | — | 1.57 → 0.51 (68%) | — | 5–8% fewer |
| Microbench 1 MiB | 10 / 10 | — | 6.38 → 2.02 (68%) | — | 5–8% fewer |
| 64 KiB @ 2,000 rps | 5 / 5 | 0.630 → 0.347 (45%) | 0.407 → 0.123 (70%) | 660 → 378 (43%) | 274K → 265K (3%) |
| 64 KiB @ 3,200 rps | 5 / 5 | 0.616 → 0.336 (46%) | 0.413 → 0.128 (69%) | 671 → 381 (43%) | 274K → 265K (4%) |
| 256 KiB @ 625 rps | 5 / 5 | 2.260 → 1.148 (49%) | 1.596 → 0.491 (69%) | 2,468 → 1,362 (45%) | 1.054M → 1.062M (−1%) |
| 256 KiB @ 1,000 rps | 5 / 5 | 2.293 → 1.181 (49%) | 1.625 → 0.498 (69%) | 2,492 → 1,350 (46%) | about equal |
| 1 MiB @ 160 rps | 5 / 5 | 9.168 → 4.527 (51%) | 6.635 → 2.110 (68%) | 9,666 → 5,002 (48%) | 4.32M → 4.72M (−9%) |
| 1 MiB @ 256 rps | 5 / 5 | 9.341 → 4.628 (50%) | 6.885 → 2.166 (69%) | 9,788 → 5,056 (48%) | 4.32M → 4.75M (−10%) |
| 64 KiB @ 1,625 rps, legacy engine† | 5 / 5 | 0.871 → 0.345 (60%) | 0.650 → 0.121 (81%) | 912 → 372 (59%) | 290K → 266K (8%) |

† Both binaries built with `GOEXPERIMENT=nojsonv2` (the pre-v2 stdlib engine). This row is reported separately and is not comparable with the default-toolchain rows.

## Quick start
```sh
make vet test race conformance   # correctness first (both GOEXPERIMENT settings)
make microbench                  # codec-only (Unmarshal + Marshal) ns/op, B/op, allocs/op + benchstat
make smoke                       # ~11 min end-to-end harness check (not a result)
make run MATRIX=loadtest/matrix/pilot.json && make report
```
Requirements: Go 1.27.4 and k6 v2.4.0. Full procedure: [`docs/runbook.md`](docs/runbook.md).

## Layout
```
service/   go.mod (only dep: goccy/go-json v0.12.0)
  cmd/{stdlib-server,goccy-server}   one-line mains
  internal/server      shared HTTP handler, lifecycle, mgmt listener (127.0.0.1:6061: /readyz /metrics /manifest /debug/pprof)
  internal/codec       Codec seam + stdlib/goccy adapters
  internal/model       shared typed model + Validate
  internal/policy      shared strict UTF-8 / duplicate-key scanner
  internal/metrics     padded atomic counters, rusage, runtime/metrics snapshot
  internal/conformance fixture goldens, 57-case v1-spec corpus, HTTP contract tests
  bench/               decoder microbenchmarks
tools/     go.mod (stdlib only)
  cmd/fixturegen  deterministic synthetic fixtures (seed 20261115)
  cmd/benchctl    build | env | trial | run: orchestration, validity gates, artifacts
  cmd/report      Markdown/HTML/CSV report with bootstrap CIs
loadtest/k6/ingest.js, loadtest/matrix/*.json
fixtures/  f1-64k, f2-256k, f3-1m (+ golden, SHA256SUMS)
docs/      methodology, metrics, correctness-policy, fixtures, runbook, threats-to-validity, adr/
```

## Integrity rules
- Results are generated only from artifacts on disk, and nothing is hand-edited.
- Invalid trials are excluded from the analysis but always listed.
- A winner is declared only when the 95% CI excludes 0 **and** the difference is ≥ 10%, with at least 5 trials per arm.
- Every result from the single-host M3 setup is labeled pilot-grade.

## License
BSD 3-Clause; see [LICENSE](LICENSE). The fixtures are synthetic and generated by this repository.
