# Threats to validity

| # | Threat | Effect | Mitigation | Residual |
|---|---|---|---|---|
| T1 | A single Apple M3 (4P+4E) with k6 and the server co-located, and no CPU pinning on macOS | k6 competes for cores, and either process can land on E-cores, which inflates variance and caps throughput | GOMAXPROCS=4; k6 CPU gates (avg ≤ 3.5, peak ≤ 4.5 cores); ABBA order; restarts; 5 reps | **High**. Every result is labeled pilot-grade |
| T2 | Thermal throttling | Later trials run slower | `pmset -g therm` before and after every trial; invalid below 90, flagged below 100; 120 s cool-down | Medium |
| T3 | Loopback network | Body read and transfer cost are far below a real network | Read time is reported separately (`srv_read`) | Results do not transfer to real networks |
| T4 | Synthetic, single-schema fixtures | Results may not generalize to other shapes, such as many small strings or deep trees | The size sweep exercises scale. A structurally different fixture is an open item | Medium |
| T5 | stdlib arm under go1.27 is v2-backed | "stdlib" means something different from older Go versions | Pair L (`nojsonv2`) is reported separately, and `build.json` records the compiled files | Low |
| T6 | go-json v0.12.0 was 17 days old when pinned | Possible regressions | Its own test suite passed under both GOEXPERIMENTs, and the 57-case conformance suite shows 0 deviations | Low or medium |
| T7 | Instrumentation (5 timestamps, header formatting) | Adds cost equally to both arms, which slightly dilutes relative differences | `-tags minimal` control cells | Quantified by the control |
| T8 | The shared policy scanner (strict UTF-8, duplicate keys) adds the same cost to both arms | Lowers the HTTP-level relative difference compared with decoding alone | `srv_pol` is reported, and the microbenchmarks show decode alone | Documented |
| T9 | Window snapshot alignment | The server window counts (Δdecoded_ok) differ from k6's measure phase by a few requests | Per-op normalization uses server counts inside the server window. The count-consistency gate uses whole-trial totals | Small (< 1%) |
| T10 | n = 5 trials per cell | Little power to detect differences under ~10% when CV is high | CV gate (> 10% → inconclusive) and bootstrap CIs | Medium |
| T11 | The GC pause histogram has coarse buckets | Pause sums are lower bounds | The value is labeled as a lower bound | Low |
| T12 | Profiling different binaries | Absolute sample counts cannot be compared | Compare shares within each profile. Profiling trials are flagged and excluded from latency claims | Documented |
| T13 | Fixed rate taken from the slower arm's knee | The faster arm runs further below its own saturation point, which is intended | Saturation cells report each arm's own knee | Intended design |

### Observation from the smoke-scale e2e run (not a result)

- [M] In the HTTP path, the shared policy scan (`srv_pol`, identical code in both arms) had a median of about 0.56–0.65 ms per request for f1. The trial-to-trial spread on this identical code was about 39%.
- [M] `BenchmarkPolicy` measures about 65 µs for the same function, roughly 8–10× less.
- [H] At about 0.27 cores of load, macOS probably runs the server on E-cores or at low clocks. Absolute server-side timings at light load therefore reflect the scheduler as much as the code.
- [mitigation] The report treats the read, policy and validate timings as **negative controls**. If any control differs by at least the threshold, the cell's library verdicts are downgraded to inconclusive. The pilot (P6) must check whether the control's spread shrinks as offered load rises.

## Open questions
- Does the decode-time difference survive at real network latency? This cannot be answered on loopback.
- Would a structurally different payload (deep nesting, numeric-heavy, many small objects) change the conclusion?
- What is goccy's behaviour under `GOMEMLIMIT` pressure? It is not tested. The soak workload (W4) is optional and not in the matrix.
