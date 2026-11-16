# Methodology

This summarizes `docs/plan/implementation-plan.md`. Where the two conflict, §13–§14 of the plan take precedence.

## Questions (answered independently)
1. HTTP latency and sustainable throughput under fixed offered load and stepped load.
2. Time spent in `Unmarshal` alone, for identical bytes into identical types (server `dec` timing plus isolated microbenchmarks).
3. CPU, allocations, retained heap, RSS, and GC behaviour, all normalized per successful operation **within the same window**.
4. Behavioural equivalence, checked against a written v1-semantics spec and not against each other (`docs/correctness-policy.md`).
5. Whether the differences are reproducible (≥5 independent trials) and at least the **10%** practical-significance threshold.

## Arms and pairs
| Pair | Arms | GOEXPERIMENT | encoding/json implementation |
|---|---|---|---|
| D (default) | `stdlib`, `goccy` | *(empty)* | go1.27 default, **v2-backed** (`v2_*.go`) |
| L (legacy) | `stdlib-legacy`, `goccy-legacy` | `nojsonv2` | legacy v1 (`decode.go`, `scanner.go`) |

The two arms of a pair always share the toolchain, GOEXPERIMENT, build flags (`-trimpath -buildvcs=true`, `CGO_ENABLED=0`), GOMAXPROCS=4, GOGC and GOMEMLIMIT. The only code that differs between them is `internal/codec/{stdlib,goccy}`. Results from different pairs are never compared directly.

## Request path (identical in both binaries)
`POST /ingest` → method check (405) → media-type check (415) → Content-Length check (413) → bounded read with `MaxBytesReader` into a buffer preallocated as min(CL, 4 MiB)+1 (413/400) → **shared policy scan** (strict UTF-8, no duplicate keys → 400) → fresh `new(model.IngestBatch)` → **`codec.Unmarshal`** (→400) → `Validate` (→422) → 202 with an empty body. There is no pooling, no custom unmarshallers, and no library options. The decoder's own trailing-data check is used, and both libraries reject `{} {}` and `{}x` [M].

## Statistical unit and inference
- The unit is the **trial**: one fresh server process, a 60 s warm-up that is excluded from the statistics, then a measured window.
- Per-trial statistics come from k6 for that trial only. Percentiles are **never averaged** across trials or steps.
- Cell estimate: the median of the per-trial values. Uncertainty: a percentile-bootstrap 95% CI on the reduction %, with 10,000 resamples and seed 20261115.
- Verdict: fewer than 5 trials per arm → *exploratory*. Trial CV above 10% → *inconclusive*. CI excludes 0 **and** |median reduction| ≥ 10% → *X better*. Otherwise *below threshold* or *no detectable difference*.
- P99 is reported only when every trial has ≥ 10,000 measured requests.
- Order: cells are shuffled with a seeded generator, and the two arms alternate ABBA across reps. Cool-down between trials is 120 s.

## Offered load
Fixed-rate cells use **one absolute rate for both arms**: R50 or R80 of the **slower** arm's pilot knee. Rates derived from each arm's own knee would bias the comparison. Saturation cells step through 10–150% of that knee.

## Validity gates (per trial, `trial.json.status`)
A trial is **invalid** if any of these hold: dropped iterations > 0, achieved rate < 99% of offered, non-202 > 0.1%, any timeout or transport error, VUs exhausted, k6 average > 3.5 cores or peak > 4.5 cores, CPU_Speed_Limit < 90, swap page-outs > 0, or k6 vs server request counts differing by > 0.1%.
A trial is **flagged** if CPU_Speed_Limit is 90–99, unexplained host CPU is > 10%, or profiling was active.
Invalid trials are excluded from the analysis but listed in the report. Flagged trials are included and listed. Profiling trials are never used for latency claims.

## Limitations accepted by the user
The run uses a single Apple M3 (4P+4E) with client and server co-located over loopback and no CPU pinning. Every result is therefore labeled **pilot-grade**. The fixtures are synthetic and form a single-schema size sweep.
