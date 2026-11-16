# Plan: encoding/json vs github.com/goccy/go-json: HTTP deserialization benchmark

Status: **PLAN ONLY. Not implemented. No benchmark results exist.** Decisions in §13–§14; awaiting approval.
Target repo: https://github.com/lurodrisilva/golang-json-encoding-benchmark
Date: 2026-11-15

Each statement below is tagged **[M]** measured (a command was run here), **[A]** assumption, **[H]** hypothesis, or **[L]** limitation.

---

## 0. Orientation evidence (collected during planning)

| Fact | Evidence command | Tag |
|---|---|---|
| Go toolchain is `go1.27.4 darwin/arm64` | `go version` | M |
| k6 is `v2.4.0`, built with go1.27.4, a **major-version** change from k6 v0/v1 | `k6 version` | M |
| Host is an Apple M3 with 8 logical CPUs (**4 P-cores + 4 E-cores**) and 24 GiB RAM | `sysctl hw.ncpu hw.perflevel0.logicalcpu hw.perflevel1.logicalcpu hw.memsize` | M |
| **Default `encoding/json` under go1.27.4 is backed by the v2 implementation.** The package compiles `v2_decode.go v2_encode.go v2_indent.go v2_inject.go v2_options.go v2_scanner.go v2_stream.go`, which carry `//go:build goexperiment.jsonv2` | `go list -f '{{.GoFiles}}' encoding/json` | M |
| `GOEXPERIMENT=nojsonv2` restores the legacy v1 sources `decode.go encode.go fold.go indent.go scanner.go stream.go tables.go tags.go` | `GOEXPERIMENT=nojsonv2 go list -f '{{.GoFiles}}' encoding/json` | M |
| Latest go-json is `v0.12.0`, published 2026-10-29 (17 days old). Its go.mod declares `go 1.23` | `go list -m -json github.com/goccy/go-json@v0.12.0` | M |
| The workspace is empty and not a git repo | `ls -la` | M |

**Consequence:** "encoding/json" is ambiguous under this toolchain. The comparison must state which `encoding/json` build is the baseline (see Q1). This does not substitute `encoding/json/v2` the package, because the import path stays `encoding/json`. But the code under test is the v2 engine behind the v1 API.

---

## 1. Interpretation of the experiment

The goal is a controlled A/B test of **one variable: the JSON decoder** used to `Unmarshal` a fully buffered request body into a shared typed model. It runs inside otherwise identical `net/http` servers. The study answers five separate questions, and each needs its own evidence:

1. **HTTP outcome:** client latency and sustainable throughput at fixed offered load, plus the saturation knee. Evidence: k6.
2. **Decoder cost in situ:** server-measured `Unmarshal` duration. Evidence: Server-Timing → k6 Trend, plus server counters.
3. **Resource cost:** CPU-seconds per successful operation, allocations per operation, retained heap, RSS, and GC cycles and pauses. Evidence: `runtime/metrics` and `getrusage` window deltas, plus pprof.
4. **Behavioral equivalence:** each implementation is checked against an **independently written expected-outcome spec**, not against the other.
5. **Decision value:** whether the differences are reproducible across trials and exceed a pre-approved practical-significance threshold.

The decoder-only microbenchmark is supplementary evidence. It never stands in for the HTTP conclusions, and HTTP never stands in for it.

Non-goals: serialization, encoding/json/v2 as a separate arm, other libraries, TLS performance, HTTP/2 performance, and framework overhead.

---

## 2. Clarifying questions, ranked by impact

1. **Which `encoding/json` is the baseline?** The default go1.27.4 build is v2-backed. Options:
   (a) default build (provisional);
   (b) `GOEXPERIMENT=nojsonv2` legacy build;
   (c) both, with legacy as a separately reported third arm `stdlib-legacy`.
   The result can flip depending on this choice **[H]**. *Recommendation: (c).*
2. **Which hardware runs the authoritative runs?** One M3 laptop has heterogeneous cores, no CPU pinning, thermal drift, and a shared client/server. *Recommendation:* two dedicated Linux hosts (server and generator) with cpusets, on a ≥10 GbE link or the same rack/VPC. Use the M3 only for pilot and development.
3. **Practical-significance threshold.** What difference justifies adding a dependency? For example, ≥10% reduction in CPU-seconds/op **and** a non-overlapping 95% CI. This must be approved before data collection, so the threshold isn't chosen after seeing results.
4. **Time budget.** The full matrix at the suggested durations takes ≥15 h of fixed-load runs alone (§7.6). Approve the tiered protocol, or a larger budget.
5. **Provisional latency thresholds** used only for run validity, not as an SLO: e.g. a run is invalid if the error rate is >0.1% or `dropped_iterations > 0`. Approve them or supply real ones.
6. **go-json version pin:** v0.12.0 (newest, little field exposure) or the previous tag v0.11.2 as the conservative choice. Both must pass the Phase-0 gate.
7. **Unicode policy:** accept invalid UTF-8 with U+FFFD replacement (v1 documented behavior), or reject it with 400? This affects validation code on both sides equally.
8. **Duplicate keys:** v1 semantics (last wins) or reject? v2-backed stdlib may differ from goccy. Do we report the difference only, or also enforce a policy?
9. **Is the optional streaming (`Decoder`) experiment in scope** for this iteration?
10. **Network topology for the authoritative tier**, and whether HTTP/1.1 cleartext with keep-alive is acceptable as the only primary transport.

## 3. Provisional assumptions (so planning can proceed)

| # | Assumption | Basis | Reversibility |
|---|---|---|---|
| A1 | Primary stdlib arm is the **default go1.27.4 build** (v2-backed). `stdlib-legacy` (`nojsonv2`) is an optional third arm reported separately | It is what users get from `import "encoding/json"` today | Easy: a build flag |
| A2 | **Every pair shares the same GOEXPERIMENT.** The goccy binary is built with the same GOEXPERIMENT as its stdlib partner | Fairness. The runtime/GC differ by experiment | Easy |
| A3 | go-json pinned to **v0.12.0**, with v0.11.2 as fallback if the Phase-0 gate fails | Newest release | Easy |
| A4 | (SUPERSEDED by §13: M3 is authoritative) The M3 is **pilot tier only**. The authoritative tier is two Linux hosts | Validity (§12) | Requires hardware |
| A5 | Transport: HTTP/1.1, cleartext, keep-alive on, no compression, no TLS | Isolates the decoder | Variants kept separate |
| A6 | Fixtures are **seeded synthetic**, labeled synthetic. An optional public-domain GeoJSON subset tests generalizability | Redistributability, no PII | Easy |
| A7 | Statistical unit = **one trial** (one process restart + warm-up + measured window). n ≥ 5 per primary cell | Avoids pseudo-replication | n can grow |
| A8 | Unicode: invalid UTF-8 is **accepted and replaced with U+FFFD** (v1 documented semantics). Any deviation is a compatibility finding | Matches the documented contract | Pending Q7 |
| A9 | No practical-significance threshold is assumed. The report shows effect sizes with CIs and **declares no winner** until Q3 is answered | Avoids post-hoc thresholds | n/a |
| A10 | `GOMAXPROCS` is fixed explicitly (e.g. 4 on the pilot = P-core count). `GOGC=100` and `GOMEMLIMIT` unset (off), recorded | Reduces scheduler variance | Easy |

---

## 4. Architecture and proposed repository tree

### 4.1 Design

```
             ┌────────────── shared: internal/server ──────────────┐
request ──►  │ mux: POST /ingest  → handler                        │
             │   1 method/content-type checks (405/415)            │
             │   2 t0; bounded read of body (413 if > limit)       │
             │   3 t1; dec.Unmarshal(buf, &v)  (fresh v)           │
             │   4 t2; Validate(&v)            (400/422)           │
             │   5 t3; counters; Server-Timing header; 202, no body│
             └───────────────▲─────────────────────────────────────┘
                             │ codec.Codec interface (one method)
           ┌─────────────────┴──────────────────┐
  internal/codec/stdlib (encoding/json)   internal/codec/goccy (go-json)
           │                                    │
   cmd/stdlib-server/main.go             cmd/goccy-server/main.go
   (3 lines: server.Run(stdlib.Codec{})) (3 lines: server.Run(goccy.Codec{}))
```

- `codec.Codec` is `interface{ Unmarshal([]byte, any) error; Name() string }`.
  - Each adapter calls **only** the plain package-level `Unmarshal`.
  - Forbidden in the primary arm: `UnmarshalNoEscape`, `UnmarshalContext`, decode options, `Decoder` (except in the secondary experiment), pooling, unsafe, and custom `UnmarshalJSON` on model types.
- The adapter is a **concrete type passed at startup**. Each binary links exactly one JSON library. A build-time check (`go version -m` + a `go list -deps` assertion in CI) proves that the stdlib binary does not link go-json. The goccy binary still links `encoding/json` transitively via `net/http/pprof` / `expvar`, which is unavoidable and documented, but it never calls it on the hot path.
- **Interface call overhead:** the interface call is the same shape in both binaries, so it cancels. A microbenchmark measures it (expected to be negligible, **[H]**).
- **Trailing data:** the primary strategy is plain `Unmarshal`. Stdlib `Unmarshal` rejects trailing non-whitespace by contract. Whether goccy `Unmarshal` does the same is **tested, not assumed**, in the correctness corpus (`{} {}`, `{}x`).
  - If goccy accepts trailing data, that is a **compatibility finding**.
  - The fix, if approved, must be identical on both sides. It must not be a pre-check on one side only.
  - Never add a scanner pass to one side.
- **Body read:** the same code on both sides. It preallocates `min(Content-Length, limit)+1` bytes, reads through `http.MaxBytesReader`, and returns 413 on overflow. A missing or invalid Content-Length falls back to a `bytes.Buffer` grown to the limit, with identical code on both sides. Its allocations show up separately in profiles under `internal/server.readBody`.
- **Does goccy copy the input?** Check in the source (Phase 0) whether plain `Unmarshal` copies or pads the input buffer. That is an allocation, which counts as decoder cost, and is documented.
- **Management listener:** `127.0.0.1:6061` (configurable, loopback only). It serves `/readyz`, `/metrics` (JSON snapshot of counters + `runtime/metrics` + `getrusage`), `/manifest` (build info), and `/debug/pprof/*`. pprof is registered on its **own mux**, never on `DefaultServeMux`, and the data listener never serves it.
- **HTTP server settings (identical):**
  - `ReadHeaderTimeout 5s`, `ReadTimeout 30s`, `WriteTimeout 30s`, `IdleTimeout 120s`, `MaxHeaderBytes 1<<16`.
  - Body limit: 4 MiB (above the 1 MiB fixture). Keep-alive on, HTTP/2 off (cleartext h1 only).
- **Lifecycle:**
  - Bind the data listener first, then mark ready.
  - On SIGTERM: readiness goes false, then `Shutdown(ctx 30s)`, then a final metrics flush to stdout as JSON, then exit 0.
  - `/readyz` returns 200 only after both listeners are bound and a self-test decode of the embedded golden fixture has succeeded.

### 4.2 Error contract (identical on both sides; no body JSON encoding)

| Condition | Status | Body | Counter |
|---|---|---|---|
| Method ≠ POST | 405 + `Allow: POST` | empty | `rejected_method` |
| Content-Type media type ≠ `application/json` (params like charset=utf-8 allowed) | 415 | empty | `rejected_ctype` |
| Body > limit | 413 | empty | `rejected_size` |
| Syntax error, trailing data, type mismatch, overflow | 400 | empty | `decode_errors` |
| Decoded but fails `Validate` | 422 | empty | `validation_failures` |
| Success | 202 | empty | `decoded_ok` |

Error bodies stay empty so that error-message text can't differ. The error **class** is mapped by a shared classifier and logged only by tests. A difference between libraries in error class (400 vs 422) is a compatibility finding.

### 4.3 Repository tree

```
golang-json-encoding-benchmark/
├── go.mod / go.sum            # go 1.27 + toolchain go1.27.4; go-json pinned
├── Makefile                   # thin wrappers over scripts/ (help target)
├── cmd/
│   ├── stdlib-server/main.go
│   ├── goccy-server/main.go
│   ├── fixturegen/main.go     # deterministic seeded generator
│   └── benchctl/main.go       # orchestrator (Go, no shell-heavy logic)
├── internal/
│   ├── model/                 # shared typed model + Validate()
│   ├── codec/{codec.go,stdlib/,goccy/}
│   ├── server/                # handler, limits, lifecycle, mgmt listener
│   ├── metrics/               # atomic counters, runtime/metrics, rusage
│   └── manifest/              # build + env capture
├── fixtures/
│   ├── primary-64k.json, size-256k.json, size-1m.json, geo-*.json (opt)
│   ├── SHA256SUMS, FIXTURES.md   # source, license, seed, schema stats
│   └── expected/*.golden.go      # independently specified assertions
├── testdata/corpus/           # negative/edge cases + expected.yaml
├── test/
│   ├── correctness/           # spec-vs-impl + differential tests
│   ├── integration/           # spawn binaries, HTTP contract tests
│   └── decodebench/           # go test -bench decoder-only
├── k6/
│   ├── lib/common.js          # fixture load (init ctx), tags, checks
│   ├── smoke.js, fixed-rate.js, saturation.js, fixed-vus.js, soak.js
├── analysis/                  # report generator (Go), templates
│   └── templates/report.md.tmpl, report.html.tmpl
├── configs/                   # matrix.yaml, profiles (pilot/authoritative)
├── docs/METHODOLOGY.md, METRICS.md, COMPATIBILITY.md
└── results/ (gitignored; see §10.2)
```

The orchestrator and analysis are written in Go to avoid a Python/R toolchain dependency. benchstat (`golang.org/x/perf/cmd/benchstat`) is pinned through a `tools` directive in go.mod.

---

## 5. Payload selection and correctness strategy

### 5.1 Public fixture evaluation
| Candidate | Verdict | Reason |
|---|---|---|
| twitter.json / citm_catalog.json (simdjson/nativejson corpora) | **Excluded** | PII (real tweets/users), unclear redistribution license |
| canada.json (GeoJSON) | Optional only | Number-heavy, few strings/escapes/booleans/nulls; license to verify at acquisition |
| Public-domain GeoJSON subset (e.g. Natural Earth, stated PD) | **Optional generalizability fixture** | Structurally different (deep numeric arrays); license verified as an acquisition step, not claimed here **[A]** |
| Seeded synthetic "order/event batch" | **Primary** | Redistributable, controllable, covers every required type, no PII |

### 5.2 Synthetic primary schema (labeled SYNTHETIC in FIXTURES.md and in every report)
`IngestBatch{ batch_id string, source {host, region, tags []string}, created_at RFC3339 string, events []Event }`
`Event{ id int64, type string(enum), ts string, user {id int64, name string(Unicode), email string, verified bool, locale *string(null-able)}, amount float64 (decimals, 2–6 places), currency string, items []Item{sku, qty int32, unit_price float64, attrs map[string]string}, geo *{lat, lon float64} (null in ~20%), note string (escapes: \" \\ \n \t é 😀, raw UTF-8 CJK/emoji), flags []bool }`
- Every JSON key maps to a model field. **Coverage proof:** a test decodes each fixture with stdlib `Decoder.DisallowUnknownFields()` (test only, not hot path) and must succeed; plus a "zero-field audit" asserting no model field is left at zero value across all events where the fixture contains it.
- Sizes are a **pure size sweep**: same schema, same generator, same formatting (compact, no whitespace), same field order; only event count changes. Targets: primary ≈ 64 KiB (50–100 KiB band), ≈ 256 KiB, ≈ 1 MiB. Exact byte counts recorded (`wc -c`, SHA-256), and transmitted bytes measured from k6 `data_sent` / server `Content-Length` counters.
- Generator: `cmd/fixturegen --seed 20261115 --events N`; uses `math/rand/v2` PCG with fixed seed, writes bytes by hand-written emitter (not via either library under test) so neither library shapes the input. Re-running must reproduce identical SHA-256 (CI check).
- Fixtures are committed; measured runs only read local files and verify SHA256SUMS first.

### 5.3 Independent semantic assertions
- Generator also emits `expected/<fixture>.summary.json`-equivalent Go golden (event count, sum of `amount` as exact decimal string, count of nulls, count of `verified=true`, specific spot-checked records incl. Unicode/escaped notes) computed from the generator's own in-memory values, **not** from either decoder.
- Tests assert each decoder's output against the golden; float comparison is exact against the generator's `strconv.FormatFloat(…,'g',-1,64)` round-trip value.

### 5.4 Correctness corpus (`testdata/corpus/expected.yaml`)
Each case: input bytes, expected HTTP status, expected decoded value (if 202), and the **source of the expectation** (v1 doc / RFC 8259 / project policy). Cases:
missing fields (→ zero values, 202 unless Validate requires), unknown fields (ignored → 202), `null` into non-pointer (no-op) and pointer (nil), int64 min/max, int64+1 overflow (400), float exponent extremes, `1e400` (400), wrong types (string→int, 400), duplicate keys (v1: last wins), case-insensitive key match (v1: matches), malformed (truncated, bad escape, lone surrogate `\ud800` → U+FFFD in v1), invalid raw UTF-8 bytes (policy A8), BOM prefix, trailing `{} {}`, `{}x`, `{}` + whitespace (accept), empty body, top-level non-object, deep nesting (10k levels: must not crash; expected 400 or 202 recorded), oversized body (413), wrong method (405), wrong content-type (415).
- Run every case against: `stdlib` (default), `stdlib-legacy` (`nojsonv2`), `goccy` under both GOEXPERIMENTs. Report a matrix; **deviations from expectation are findings, never normalized away**. Expected-differences listed in `docs/COMPATIBILITY.md` (v2-backed stdlib is known to change some v1 edge semantics **[H]** — to be measured, e.g. duplicate names, invalid UTF-8).

### 5.5 Test tiers
1. Unit: model `Validate`, body reader limits, classifier, counters.
2. Spec tests: each codec vs golden + corpus (`go test ./test/correctness`).
3. Differential: both codecs on fixtures + fuzz corpus (`go test -fuzz` seeded from corpus, time-boxed); disagreements recorded, not asserted equal.
4. Integration: spawn each built binary, drive HTTP contract table, assert status codes, empty bodies, counters; assert 202 is never sent before decode+validate (handler test with a decoder stub that blocks).
5. Concurrency: `go test -race` on handler with 64 concurrent goroutines posting distinct fixtures, asserting each decoded value matches its own input (detects object reuse). Race runs are **never** used for performance.

---

## 6. Instrumentation and metric dictionary

### 6.1 Timing boundaries (server, `time.Now()` monotonic; ~20–40 ns each **[H]**, measured in Phase 2)
- `t0` handler entry (after mux dispatch); `t1` body fully read; `t2` Unmarshal returned; `t3` Validate returned. `t_end` = t3 taken **before** Server-Timing header formatting and `WriteHeader`.
- Server-Timing: `read;dur=<ms>, dec;dur=<ms>, val;dur=<ms>, hdl;dur=<ms>` with µs-precision decimals via `strconv.AppendFloat` into a stack buffer (no fmt).

| ID | Metric | Unit | Collection | Boundary / caveat |
|---|---|---|---|---|
| A | `http_req_duration` | ms | k6 built-in | = sending + waiting + receiving. **Excludes** `http_req_blocked`, `connecting`, `tls_handshaking` (DNS/connection setup). Includes client-side kernel/network time, server queueing, handler, response header write. Verify against k6 v2.4.0 docs in Phase 0 |
| A' | `http_req_waiting` (TTFB) | ms | k6 | Closest client analogue to server handling |
| B | `srv_handler_ms` | ms | Server-Timing `hdl` → k6 Trend | t0→t_end. **Excludes** mux, header parsing before handler, response delivery |
| C | `srv_decode_ms` | ms | Server-Timing `dec` → k6 Trend | Only the `Unmarshal` call |
| D | `srv_read_ms` | ms | Server-Timing `read` | Body read; includes network wait for body bytes |
| — | `srv_validate_ms` | ms | `val` | Validate only |
| E | Wall-clock window | s | orchestrator timestamps + k6 measured scenario duration | Excludes warm-up; drain interval (`gracefulStop`) reported separately |
| F | Σ handler, Σ decode | s | server atomic sums (ns) diffed over window | Cumulative across concurrent requests — **not** wall time; can exceed E |
| G | CPU-seconds | s | `getrusage(RUSAGE_SELF)` utime+stime delta via `/metrics` | Process-wide incl. GC, net, runtime |
| H | CPU utilization | % of 1 core and % of GOMAXPROCS | G / E | |
| I | CPU per op | µs/op | G / Δdecoded_ok | Same window |
| J | RSS (SUPERSEDED by §13.1: `ps -o rss` 1 Hz + getrusage maxrss) | bytes | Linux `/proc/self/statm` sampled 1 Hz; macOS `getrusage` maxrss (peak only) **[L]** | pprof does not provide RSS |
| K | Live heap | bytes | `runtime/metrics` `/gc/heap/live:bytes`, 1 Hz | |
| L | Alloc bytes / objects | bytes, count | `/gc/heap/allocs:bytes`, `/gc/heap/allocs:objects` deltas | Per op = Δ / Δdecoded_ok; objects ≠ operations |
| M | GC cycles | count | `/gc/cycles/total:gc-cycles` delta | |
| N | GC pauses | s histogram | `/sched/pauses/total/gc:seconds` (verify name on 1.27) | Histogram delta, report p50/p99/max bucket |
| O | Counters | count | atomics: `received`, `decoded_ok`, `decode_errors`, `validation_failures`, `rejected_*`, `responded` (incremented after `WriteHeader` returns — server committed, **not** client delivered), `inflight` gauge | Low overhead: padded per-counter atomics (no locks, no labels) |
| P | k6 rates | /s | `iterations`, `http_reqs`, `dropped_iterations`, checks, `http_req_failed` | offered = configured rate; started = iterations; completed = http_reqs; successful = status 202 check |
| Q | Generator health | %, bytes | k6 host: `pidstat`/`top` 1 Hz on k6 PID; NIC bytes | Run invalid if k6 CPU > 80% of its cores **[A]** |

### 6.2 "Total processing time" (explicit, reported as four separate numbers)
1. Measured window wall-clock (E) and drain interval;
2. Σ per-request handler durations (F);
3. Σ per-request decode durations (F);
4. CPU-seconds in window (G).
Never summed or equated.

### 6.3 Instrumentation overhead control
Build tag `-tags minimal` removes Server-Timing header and per-stage timestamps (keeps counters). One control run per arm on primary fixture at the 50% rate; report Δ HTTP P50/P99 and CPU/op vs instrumented.

---

## 7. Workload matrix and experiment protocol

### 7.1 Arms
`stdlib` (default GOEXPERIMENT), `goccy` (same GOEXPERIMENT). Optional pair `stdlib-legacy` + `goccy-legacy` (both `nojsonv2`), analyzed as a separate pair.

### 7.2 Fixtures
F1 primary ≈64 KiB (all workloads), F2 ≈256 KiB, F3 ≈1 MiB, F4 optional GeoJSON.

### 7.3 Workloads (all PROVISIONAL, calibrated by pilot)
| ID | Executor | Purpose | Provisional settings |
|---|---|---|---|
| W0 smoke | `shared-iterations` 1 VU, 20 iters | contract check | status 202 + Server-Timing present |
| W1 fixed-rate | `constant-arrival-rate` | primary comparison | rates R50, R80 = 50%/80% of **slower arm's** pilot knee, same absolute rate for both arms |
| W2 saturation | `ramping-arrival-rate` steps | knee discovery | 10 steps from 10% to 150% of pilot knee, 2 min each after 60 s warm-up; per-step stats only, never pooled across steps |
| W3 fixed-VUs (opt) | `constant-vus` | closed-loop throughput | VUs = 4,16,64; labeled closed-loop; not compared with W1 |
| W4 soak (opt) | `constant-arrival-rate` R50 | GC/retention stability | 2 h, F1 only, 1 rep per arm |

- Warm-up: separate k6 scenario (`startTime` 0, 60 s at target rate) tagged `phase=warmup`; thresholds/summary use `{phase:measure}` submetrics; report also from raw output filtered by tag.
- VU budget: `preAllocatedVUs` = rate × expected P99 latency × 2, `maxVUs` 4× that; memory budget: fixture bytes are loaded once via `open(path,'b')` in init context — verify in k6 v2.4.0 whether init-context data is per-VU copied (then F3 × maxVUs RAM) or shared (`SharedArray` not usable for binary) **[A, verify]**.
- Run validity gates (provisional, need approval Q5): `dropped_iterations == 0`, non-202 rate ≤ 0.1%, k6 CPU ≤ 80%, achieved rate within 1% of offered. Failing run → `INCONCLUSIVE`, kept in data, excluded from comparison, reported.

### 7.4 Per-trial protocol
1. Verify fixture SHA-256, binary SHA-256 vs manifest, empty run dir (refuse to overwrite).
2. Start server (pinned cpuset on Linux; env `GOMAXPROCS`, `GOGC`, `GOMEMLIMIT` explicit), poll `/readyz` (timeout 10 s).
3. Snapshot `/metrics` → k6 warm-up → snapshot → k6 measure → snapshot → drain → snapshot.
4. SIGTERM, wait for exit, collect final flush; verify server `decoded_ok` delta ≈ k6 202 count (mismatch > 0 → flagged).
5. Cool-down 30 s; check thermal/CPU frequency where available (Linux `turbostat` optional; macOS `powermetrics` needs sudo **[L]**).

### 7.5 Ordering
Blocked by (fixture, workload, rate); within a block, arm order follows a balanced ABBA/BAAB sequence over the 5+ reps, seeded and recorded. Process restart per trial; identical warm-up.

### 7.6 Time budget (justifies Q4)
Full: 2 arms × 3 fixtures × 2 rates × 5 reps × (1 + 25 + ~1.5 overhead) min ≈ **27.5 h**, plus saturation (2×3×~23 min ≈ 2.3 h), profiling (~2 h), soak (4 h).
**Proposed tiered protocol:** F1 at R50 and R80 × 5 reps × 25 min (≈ 9.2 h); F2/F3 at R50 × 5 reps × 10 min (≈ 4.6 h); saturation once per arm per fixture; legacy pair F1 R50 only. Total ≈ 17 h, run overnight in chunks with cross-chunk ABBA balance. **P99 sufficiency is not assumed:** at R50 for 25 min there are many samples, but trial-to-trial variance dominates; report the CI of per-trial P99s.

### 7.7 Transport and topology (topology SUPERSEDED by §13.1: loopback on M3)
HTTP/1.1, cleartext, keep-alive on (k6 default), `noConnectionReuse=false`, no `Accept-Encoding` / no compression, k6 `batch` not used, connections ≈ VUs. Authoritative: generator and server on separate hosts, same L2 segment, ≥10 GbE, RTT recorded with `ping`. Pilot: both on M3 via loopback — server limited to `GOMAXPROCS=4`, k6 sharing remaining cores, no pinning possible **[L]**.

---

## 8. Profiling and analysis methodology

### 8.1 Profiling runs (separate trials, never the primary latency trials)
- Profile mode tag `profiling=on`; same rate R50, F1–F3; 1–2 reps per arm.
- CPU: `GET /debug/pprof/profile?seconds=60` started after warm-up; record server counter snapshots immediately before and after → ops in window.
- Allocations: `GET /debug/pprof/allocs?seconds=60` (delta profile, supported by net/http/pprof for allocs/heap) → `alloc_space`, `alloc_objects` for the window; alternatively two snapshots with `pprof -base`.
- Heap: `/debug/pprof/heap` snapshot at window end → `inuse_space`, `inuse_objects`; optional post-GC snapshot `?gc=1` labeled "post-GC" taken identically in both arms only at end of a profiling run.
- MemProfileRate left default (512 KiB) in both arms; noted as sampling.
- Outputs: raw `.pb.gz`, `go tool pprof -top -nodecount=40` flat & cum text, `-svg` call graph (requires Graphviz `dot`, documented), and `-raw`/flame view via `go tool pprof -http` (manual) — report embeds SVG.
- Binaries kept next to profiles (`bin/<arm>`, SHA-256) for symbolization.
- Caveats written in report: pprof ≠ RSS; alloc object counts ≠ operations; CPU profile is sampled (100 Hz), not per-request time; heap = retained state, allocs = activity; cross-binary profile comparison is by function attribution (e.g. `encoding/json.*` vs `github.com/goccy/go-json.*` cum %), not by address; `-diff_base` across different binaries only meaningful for shared symbols (net/http, runtime, internal/server).

### 8.2 Decoder microbenchmarks (`test/decodebench`)
`BenchmarkUnmarshal/{codec}/{fixture}` with `b.SetBytes(len)`, `b.ReportAllocs()`, fixture loaded before `b.ResetTimer()`, fresh destination per iteration, `b.Loop()`. Run `go test -run=^$ -bench . -count=10 -benchtime=2s -cpu=1,4`, alternating codec order via separate invocations; compare with `benchstat` (pinned). Also a `-cpu=4` `RunParallel` variant. Reported separately; also built for both GOEXPERIMENTs.

### 8.3 Statistics
- Unit: trial. Per trial compute metrics (P50/P95/P99 from k6 raw samples of the measure phase, CPU/op, allocs/op, ...).
- Comparison: per-metric difference of trial medians, with 95% bootstrap CI (10,000 resamples, seeded) and Mann–Whitney U as secondary; with n=5 per arm, power is limited **[L]** — wide CIs are reported, not hidden.
- Percentiles: never averaged across trials or stages. Pooled percentiles (optional) computed from concatenated raw samples of all trials in a cell and labeled "pooled raw samples, equal-duration trials, mixture = all reps".
- Reduction % = (stdlib − goccy) / stdlib × 100 for lower-is-better; if stdlib = 0, print "n/a (zero baseline)" plus absolute diff. Higher-is-better (throughput) shown as (goccy − stdlib)/stdlib.
- Raw-sample storage: k6 `--out csv=…gz`? k6 CSV not natively gzipped → write CSV then `zstd` post-run; estimate ~100 bytes/sample × 3 trends → for R=2000/s × 1500 s ≈ 900 MB raw per trial before compression **[H]** → alternative: k6 `--out json` filtered to required metrics, or HDR histograms computed by a k6 extension (no extension planned; decide in pilot).

---

## 9. Ordered implementation phases with acceptance criteria

| Phase | Work | Acceptance criteria |
|---|---|---|
| P0 Gate | go mod, pin go-json; build under go1.27.4 both GOEXPERIMENTs; run go-json's own tests; source review of `Unmarshal` (input copy, trailing data); k6 v2.4.0 doc check (metric names, `count` in summaryTrendStats, `open(...,'b')`, tags) | Builds w/o special flags (or flag applied to both & recorded); findings in `docs/COMPATIBILITY.md`; pin decision recorded |
| P1 Fixtures | generator, fixtures, SHA256SUMS, goldens, coverage audit | deterministic re-gen byte-identical; DisallowUnknownFields + zero-field audit pass; sizes in bands |
| P2 Server | shared server, codecs, mgmt listener, counters, manifest, lifecycle | integration contract tests pass for both binaries; pprof not reachable on data port; `go list -deps` assertion; instrumentation overhead microbench recorded |
| P3 Correctness | corpus, spec, differential, fuzz, race | all tests run; compatibility matrix generated; expected outcomes sourced; race clean |
| P4 Microbench | decodebench + benchstat | benchstat output for both GOEXPERIMENTs |
| P5 k6 + orchestrator | k6 scripts, `benchctl` with matrix, validity gates, artifact layout | smoke passes both arms; one dry trial end-to-end produces complete artifact set; stale-dir protection tested |
| P6 Pilot | M3 pilot: knee per arm, variance, sample-storage size, generator headroom | calibrated `configs/matrix.yaml` approved by user |
| P7 Profiling | profile mode in benchctl, pprof post-processing | profiles + text + SVG with ops-in-window per arm/fixture |
| P8 Report | analysis + MD/HTML templates, charts (SVG, generated in Go) | report renders from a fixture dataset; every table cell traceable to raw files |
| P9 Authoritative run | full tiered protocol on approved hardware | all validity gates evaluated; report generated; nothing hand-edited |

---

## 10. Commands to be implemented and artifact layout

### 10.1 Commands (planned; none exist yet)
```
make setup            # go mod download && go mod verify; check k6/go versions vs configs/versions.lock
make build            # CGO_ENABLED=0 go build -trimpath -buildvcs=true -o bin/<arm> ./cmd/<arm>-server (×arms, ×GOEXPERIMENT)
make fixtures         # go run ./cmd/fixturegen --all && sha256sum -c fixtures/SHA256SUMS
make fixtures-verify  # sha256sum -c only
make test             # go test ./...
make test-race        # go test -race ./internal/... ./test/...
make test-correctness # go test ./test/correctness -v (both GOEXPERIMENTs) → results/<run>/correctness/
make serve ARM=stdlib # foreground server, loopback mgmt
make smoke            # benchctl smoke --arms stdlib,goccy
make bench-decoder    # benchctl microbench → benchstat
make run              # benchctl run --matrix configs/matrix.yaml --profile authoritative
make saturation | soak | profile
make report RUN=<id>  # go run ./analysis --run results/<id>
```
One-command: `benchctl run --matrix configs/matrix.yaml` performs: preflight (versions, fixture checksums, binaries' `go version -m`, clean tree, free ports, no stale server PIDs), creates `results/<run-id>/` atomically (`run-id = UTC yyyymmddThhmmssZ-<gitsha7>-<matrixhash8>`, refuse if exists), executes trials in recorded randomized order, traps signals and always kills children (process group), writes `trial.status` (OK/INCONCLUSIVE/FAILED) per trial, continues past failed trials, then runs `report`. Resume via `--resume <run-id>` skipping trials with `OK` status.

### 10.2 Artifact layout
```
results/<run-id>/
  manifest.json          # go env, versions, module sums, GOEXPERIMENT, GOOS/ARCH, CGO, GOMAXPROCS, GOGC, GOMEMLIMIT, ldflags, host HW, kernel, k6 version, order seed
  bin/<arm> + sha256     # exact binaries
  correctness/ microbench/
  trials/<arm>/<fixture>/<workload>-<rate>/rep-<nn>/
      k6-summary.json  k6-raw.csv.zst  server-metrics-{pre,warm,post,drain}.json
      server-stdout.log  sysstat-server.csv  sysstat-generator.csv  trial.status
      profiles/ (profiling trials only)
  analysis/  data.csv  data.json  report.md  report.html  charts/*.svg
```
No payload bodies, env secrets, or hostnames beyond a configured label are written; logs contain only counters.

---

## 11. Comparative report template (structure)
1. **Executive summary** — question-by-question answers, each with: measurement, CI, interpretation, confidence; conditional recommendation (e.g. "adopt only if …"); "no winner" when CI spans zero or effect < approved threshold.
2. Environment & reproducibility manifest (table from manifest.json; stdlib implementation = v2-backed / legacy explicitly).
3. Fixtures (bytes on disk, bytes transmitted, SHA-256, synthetic label, schema stats).
4. Correctness & compatibility matrix (case × arm × GOEXPERIMENT → outcome vs expected, source).
5. Decoder microbenchmarks (benchstat table: ns/op, B/op, allocs/op, MB/s).
6. HTTP fixed-rate results per fixture × rate: table of arms with per-trial values + median + CI for: HTTP min/P50/P95/P99/max/avg; handler & decode same set; offered/achieved rate; errors/timeouts/dropped; total processing time (4 numbers); CPU-s, util, CPU µs/op; RSS, live heap, alloc B/op, allocs/op, GC cycles/s, GC pause p99; abs and % diff columns.
7. Saturation: latency-vs-offered-load and throughput-vs-offered-load charts (per step, per trial), knee estimate.
8. Memory & CPU comparisons (charts).
9. Profile hotspots: top-20 flat & cum per arm; share of CPU in decoder vs net/http vs runtime/GC vs body read.
10. Effects decomposition: decoder vs HTTP/network/GC/allocation/validation/instrumentation (control run Δ).
11. Saturation and generator-capacity limitations.
12. Threats to validity; unresolved questions; inconclusive runs list.
Each section labels [Measured] / [Interpretation] / [Hypothesis] / [Limitation].

---

## 12. Risks, blind spots, and decisions requiring approval

### Risks / threats to validity
- **R1 stdlib identity**: v2-backed default under go1.27.4 changes both performance and semantics vs historic "encoding/json" claims; results not transferable to older Go **[L]**.
- **R2 go-json maturity/runtime coupling**: go-json uses unsafe runtime-type internals; a fresh v0.12.0 on go1.27 may have latent bugs; the Phase-0 gate + fuzz mitigates, not eliminates.
- **R3 Hardware** (SUPERSEDED by §13: M3 is authoritative; risk stands, labeled pilot-grade): M3 P/E heterogeneity, no pinning, thermal throttling, shared client/server, macOS network stack → pilot only.
- **R4 Generator bottleneck**: k6 JS VM costs per request may exceed server costs at small payloads → saturation could measure k6. Mitigation: generator CPU gate, separate host.
- **R5 Decode share small**: if decode is a small fraction of request time, HTTP differences may be undetectable; this is a valid finding, not a failure.
- **R6 GC coupling**: allocation differences change GC frequency, which affects the *other* requests' latency; attributed via GC metrics, not removed.
- **R7 Synthetic fixture bias**: schema choice (string-heavy vs number-heavy) can favor one library; mitigated by F4 and by publishing schema stats.
- **R8 Instrumentation overhead**: Server-Timing formatting and timestamps; quantified by control run.
- **R9 Statistical power**: n=5 detects only large effects; P99 noisy.
- **R10 Raw sample volume**: storage/IO from k6 output may perturb the generator; measured in pilot.
- **R11 Body read strategy**: preallocation from Content-Length is identical, but trusting Content-Length is a design choice; documented.

### Blind spots
- Encoding (responses), streaming, HTTP/2, TLS, other GOARCH (amd64) not covered in primary.
- Production heterogeneity of payloads not represented by 3 sizes of one schema.

### Decisions requiring approval (before implementation)
1. Stdlib arm choice (Q1) — default / legacy / both.
2. Hardware tier for authoritative runs (Q2).
3. Practical-significance threshold (Q3).
4. Tiered time budget (Q4, §7.6).
5. Provisional validity thresholds (Q5).
6. go-json pin (Q6).
7. Unicode and duplicate-key policy (Q7, Q8).
8. Streaming secondary experiment in/out (Q9).
9. Create/push to the GitHub repo (no push done; none will be done without approval).

---

## 13. Decisions recorded 2026-11-15 (these override any conflicting text above)

| Decision | Answer | Effect on the plan |
|---|---|---|
| Q1 stdlib arm | **Both, reported apart** | Two pairs are now in scope, and each pair shares one GOEXPERIMENT. **Pair D:** `stdlib` vs `goccy`, both on the default (v2-backed) build. **Pair L:** `stdlib-legacy` vs `goccy-legacy`, both with `GOEXPERIMENT=nojsonv2`. Pair L is now required, not optional. Pairs are never compared across each other as if they were one A/B test. |
| Q2 hardware | **M3 only** | The M3 is the authoritative tier. Every result is labeled **"single-host, Apple M3, client and server co-located: pilot-grade validity"** [L]. |

### 13.1 Mitigations on the M3 (it has no CPU pinning)
- Run on AC power with Low Power Mode off. Quit other apps and disable Spotlight indexing and Time Machine during runs. Record all of this in the manifest.
- Server `GOMAXPROCS=4` (the P-core count). k6 runs with `GOMAXPROCS=4` too. The macOS scheduler still places threads freely, and this is documented. Neither process uses `taskpolicy -b`, because background QoS forces a process onto the E-cores.
- Capture thermal state before and after each trial with `pmset -g therm`. A trial is flagged when the CPU speed limit is below 100. `powermetrics` needs sudo, so it is an optional extra.
- **RSS on macOS:** the orchestrator samples `ps -o rss= -p <pid>` at 1 Hz, and peak RSS comes from `getrusage` maxrss (bytes on darwin). Metric J no longer uses `/proc`. Live heap, allocations and GC still come from `runtime/metrics`.
- Loopback-only topology, so no real NIC is involved [L]. Network effects are therefore outside what this study can claim.
- The generator-CPU validity gate becomes "k6 CPU ≤ 3.5 cores averaged". Saturation results are treated as **"knee of the combined client+server system"** and are not attributed to the server alone.
- Because of thermal drift, ABBA ordering and cool-downs are mandatory, not optional. Add a 120 s cool-down between trials.

### 13.2 Revised time budget (tiered, still needs approval Q4)
| Block | Cells | Duration |
|---|---|---|
| Pair D, F1, R50+R80, 5 reps × 25 min | 20 trials | ≈ 9.8 h |
| Pair D, F2+F3, R50, 5 reps × 10 min | 20 trials | ≈ 4.7 h |
| Pair L, F1, R50+R80, 5 reps × 25 min | 20 trials | ≈ 9.8 h |
| Pair L, F2+F3, R50, 3 reps × 10 min — **SECONDARY/EXPLORATORY** (below the 5-rep primary rule) | 12 trials | ≈ 2.8 h |
| Saturation (4 arms × 3 fixtures) | 12 | ≈ 4.6 h |
| Profiling + control runs | — | ≈ 3 h |
| **Total** | | **≈ 35 h**, run in overnight chunks with ABBA balance across chunks |

A cheaper option for approval: cut 25 min to 15 min on F1. That brings the total to ≈ 27 h. The pilot (P6) decides whether trial-to-trial variance, rather than the sample count, dominates P99. If it does, shorter trials with more reps are the better spend.

### 13.3 Still open
Q11 Pair L scope: option (a) or (b) in §13.4. Q3 (practical-significance threshold), Q4 (time budget), Q5 (validity thresholds), Q6 (go-json pin), Q7/Q8 (Unicode and duplicate-key policy), Q9 (streaming), and GitHub repo creation/push.

### 13.4 Correction: the cost of Pair L
The decision question said Pair L adds "about 15% to runtime". **That figure was wrong.** Without Pair L the tiered plan takes about 20 h. With Pair L as written in §13.2 it takes about 35 h, which is **about +75%**.
Options, awaiting approval:
- **(a)** Keep §13.2 as written (about 35 h). Pair L's F2/F3 cells are exploratory at 3 reps, or 5 reps for about another 2 h.
- **(b)** Reduced Pair L: F1 at R50 only, 5 reps × 25 min (about 2.4 h), plus F1 saturation for 2 arms (about 0.8 h), plus profiling (about 0.5 h). Correctness matrix and microbenchmarks still run under `nojsonv2`. Total about 23.7 h, **about +19%**.

---

## 14. Decisions recorded 2026-11-15, round 2 (override §1–§13 where they conflict)

| # | Answer | Interpretation applied |
|---|---|---|
| Q11 | b | Pair L is reduced: F1 (64 KiB) at R50 only, 5 reps; F1 saturation for both legacy arms; F1 profiling. The correctness matrix and microbenchmarks also run under `nojsonv2`. |
| Q3 | yes | Adopt the proposed threshold in §14.3. **The number needs confirming.** |
| Q4 | 64 KiB | F1 steady-state trials are shortened from 25 to **15 min**. All other cells are unchanged. |
| Q5 | Claude to propose | See §14.4. |
| Q6 | latest stable | `github.com/goccy/go-json v0.12.0`, published 2026-10-29T03:09:19Z. Checked with `go list -m -versions`, which shows no newer non-prerelease tag. Phase P0 must still pass on go1.27.4. If it fails, stop and report; **do not fall back silently**. |
| Q7/Q8 | strict UTF-8, no duplicate keys | Shared, library-independent policy layer (§14.1). |
| Q9 | no | The streaming/Decoder experiment is dropped. |
| GitHub | repo exists, empty | `git ls-remote` returned no refs. Layout: monorepo with docs (§14.2). The first push happens only after plan approval. |

### 14.1 Strict UTF-8 and duplicate-key policy (affects fairness)
**[M] Fact:** both arms accept invalid UTF-8 and duplicate keys by default. Stdlib's v1 `Unmarshal` semantics replace invalid UTF-8 with U+FFFD and let the last duplicate win, and the v2-backed default build keeps those v1 semantics. go-json is expected to do the same; the P3 corpus verifies this per arm. Neither arm's v1 API can enforce the policy natively without library-specific options, which §2 forbids.

**Design:**
- Add a shared package, `internal/policy`. It is a hand-written, allocation-light single-pass scanner that uses no JSON library. It rejects:
  - (a) invalid UTF-8 in the raw bytes, including invalid UTF-8 produced by `\u` escapes such as lone surrogates `\ud800`;
  - (b) a duplicate member name within any single object, compared after unescaping and case-sensitively.
- It runs on the identical byte slice in both arms, after the body read and before `Unmarshal`. It has its own timing segment, `pol`, in Server-Timing, and its own metric (C2).
- Outcomes: policy violation → **400** with an empty body. Counter `policy_rejects_total{reason=utf8|dupkey}`.
- **Decode time (metric C) still covers only `Unmarshal`.** The policy cost is identical across arms and appears in handler time (B). The report shows B with and without `pol`.
- The decoder microbenchmarks benchmark `Unmarshal` alone, without the policy scanner. The scanner gets its own microbenchmark so its share is visible.
- **Compatibility findings stay visible.** The P3 corpus is run twice:
  - (i) through the HTTP stack, with policy on: the expected contract;
  - (ii) through raw `Codec.Unmarshal`, with no policy: each library's native behaviour on invalid UTF-8, lone surrogates and duplicate keys, documented per arm and per GOEXPERIMENT.
- **Risk:** the scanner is new code, so it gets fuzz tests (`go test -fuzz`), differential tests against `utf8.Valid`, and a test proving it never accepts input that `Unmarshal` rejects as syntax error.
- Expected outcome per case comes from the written policy, not from either library.

### 14.2 Monorepo layout (supersedes the §4 tree)
```
golang-json-encoding-benchmark/
├── go.work                       # use ./service ./tools
├── Makefile                      # thin wrapper over benchctl
├── README.md                     # what/why, quickstart, link to docs/
├── LICENSE
├── .github/workflows/ci.yml      # vet, unit, race, fuzz-smoke, fixture checksums (no perf in CI)
├── docs/
│   ├── methodology.md            # experiment design, fairness rules, statistics
│   ├── metrics.md                # metric dictionary A–Q (+C2 policy)
│   ├── correctness-policy.md     # error contract, UTF-8/dup-key policy, corpus table
│   ├── fixtures.md               # generator, schema, checksums, sizes
│   ├── runbook.md                # M3 host prep, exact commands, troubleshooting
│   ├── threats-to-validity.md
│   └── adr/                      # 0001-stdlib-two-pairs, 0002-m3-single-host, 0003-strict-policy, …
├── service/                      # module …/service — the system under test
│   ├── go.mod                    # sole third-party dep: github.com/goccy/go-json v0.12.0
│   ├── cmd/stdlib-server/main.go
│   ├── cmd/goccy-server/main.go
│   ├── internal/server/          # handler, mux, limits, error contract, lifecycle
│   ├── internal/codec/           # Codec interface; stdlib.go, goccy.go
│   ├── internal/model/           # shared typed model + Validate()
│   ├── internal/policy/          # UTF-8 + dup-key scanner (+ fuzz)
│   ├── internal/metrics/         # atomic counters, runtime/metrics, mgmt listener
│   └── bench/                    # decoder + policy microbenchmarks
├── tools/                        # module …/tools — never linked into servers
│   ├── go.mod                    # stdlib only
│   ├── cmd/benchctl/             # orchestrator
│   ├── cmd/fixturegen/           # seeded deterministic generator
│   └── cmd/report/               # CSV/JSON → Markdown + HTML (html/template, inline SVG)
├── loadtest/k6/                  # smoke.js, steady.js, saturation.js, lib/
├── fixtures/                     # f1-64k.json, f2-256k.json, f3-1m.json, corpus/, SHA256SUMS
├── reports/                      # committed published reports, one dir per run-id
└── results/                      # raw run output — .gitignored
```
Rationale for two modules: the servers' dependency graph contains only go-json, and orchestration and report code can never change what the server binaries link. `go version -m` on each binary proves that `stdlib-server` does not contain go-json.

### 14.3 Practical-significance threshold (proposed; confirm the number)
A difference is reported as **practically significant** only if all four conditions hold:
1. The 95% bootstrap CI of the difference in per-trial medians excludes 0.
2. The point estimate is a **≥ 10% reduction** in at least one decision metric:
   - CPU-seconds per successful operation (metric H);
   - P50 handler time (B);
   - P99 HTTP latency at R80 (A);
   - sustainable knee throughput (W2).
3. No decision metric regresses by ≥ 10%.
4. Zero behavioural incompatibilities on the supported-input corpus, or every one is listed and accepted.

Decode-only microbenchmarks use the same 10% bar with benchstat (p < 0.05) but **cannot alone** justify the dependency. Anything below 10% is reported as "measurable but negligible".

### 14.4 Validity thresholds (proposed by Claude, adjustable)
A trial is **invalid** (excluded and re-run, with the exclusion reported) if any of these happen in the measured window:
- `dropped_iterations > 0`;
- achieved rate < 99% of offered;
- non-202 responses > 0.1%;
- any k6 timeout;
- k6 reaches `maxVUs`;
- the k6 process averages > 3.5 cores of CPU, or any 5 s window exceeds 4.5 cores;
- `pmset -g therm` reports CPU_Speed_Limit < 90 at any sample;
- the macOS swap pageout delta (`vm_stat`) is > 0;
- server and k6 request counts differ by > 0.1% (the boundaries are documented).

A trial is **flagged** but kept if CPU_Speed_Limit is between 90 and 99, or the ambient idle CPU in the 30 s before the trial is > 10%.

A cell is **inconclusive** if:
- the per-trial median HTTP latency across 5 reps has CV > 10%. Add reps up to 10; if it remains, report it as inconclusive.
- more than 2 trials were invalid.

**Saturation knee:** the highest step where achieved ≥ 99% of offered, no drops, and P99 ≤ 3× the P99 at the lowest step.

**Readiness:** `/readyz` returns 200 within 10 s, or the trial aborts.

### 14.5 Revised time budget
This replaces §13.2 and §13.4. The §13.4(b) estimate was also low, because its Pair L trials counted only one arm. Per-trial overhead (warm-up, cool-down and restart) is about 4.4 min.

| Block | Trials | Est. |
|---|---|---|
| Pair D, F1, R50+R80, 5 reps × 15 min | 20 | 6.5 h |
| Pair D, F2+F3, R50, 5 reps × 10 min | 20 | 4.7 h |
| Pair L, F1, R50, 5 reps × 15 min | 10 | 3.2 h |
| Saturation: Pair D × 3 fixtures + Pair L × F1 | 8 | 3.1 h |
| Profiling + minimal-instrumentation control | — | 3.5 h |
| Microbenchmarks, correctness, race, fuzz | — | < 1 h |
| **Total** | | **≈ 21–22 h**, run as 3 overnight chunks with ABBA balance in each chunk |

### 14.6 Plan deltas
- **P2:** adds `internal/policy` plus fuzz tests.
- **P3:** dual corpus runs (policy on and off).
- **P5:** `benchctl` gains the validity gates in §14.4, writing them to `trial.json` as `valid|flagged|invalid` with reasons.
- **P8:** the report adds a "policy cost" panel and the §14.3 decision table.
- **New P1a:** repository bootstrap. Clone the empty repo; add `go.work`, docs skeleton, ADRs 0001–0003, CI; push to `main` after approval.
- **Removed:** the streaming experiment.
- **Remaining open items:** confirm the 10% threshold (§14.3); approve the plan as a whole.

### 14.7 Trial durations, round 3 (overrides §14.5 durations)
- F1 64 KiB steady-state trials: **10 min** (both pairs, R50 and R80).
- F2 256 KiB steady-state trials: **5 min**.
- F3 1 MiB: **10 min** (confirmed by the user).
- New validity rule: a trial must record **≥ 10,000 measured requests** for its P99 to be reported, which leaves ≥ 100 samples above P99. Otherwise P99 is marked "insufficient samples" and only P50/P95 are used. At 5 min this needs ≥ 34 req/s, and the pilot (P6) confirms the rate.
- Shorter trials have less power to detect slow drift such as GC or heap growth. Soak (W4) stays optional.

| Block | Trials | Est. |
|---|---|---|
| Pair D, F1, R50+R80, 5 reps × 10 min | 20 | 4.8 h |
| Pair D, F2, R50, 5 reps × 5 min | 10 | 1.6 h |
| Pair D, F3, R50, 5 reps × 10 min | 10 | 2.4 h |
| Pair L, F1, R50, 5 reps × 10 min | 10 | 2.4 h |
| Saturation | 8 | 3.1 h |
| Profiling + control | — | 3.5 h |
| Microbench, correctness, race, fuzz | — | < 1 h |
| **Total** | | **≈ 18–19 h** (2 overnight chunks) |

### 14.8 Confirmed
- Practical-significance threshold: **10%** (§14.3 confirmed).
- Open items: none, apart from approval of the plan as a whole.
