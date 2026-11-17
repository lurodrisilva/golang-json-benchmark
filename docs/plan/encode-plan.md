# Encode experiment plan (GET /batch)

Status: approved 2026-11-17 (decisions A1–A6 below). Implementation phases
E0–E5 proceed; the encode pilot (E6) and the authoritative encode run (E7)
each need separate approval before they start.

## Interpretation

| Item | Interpretation |
|---|---|
| Goal | Compare `Marshal` from `encoding/json` and `github.com/goccy/go-json` through `GET /batch` and in decoder-free microbenchmarks, using the same fixtures, the same kinds of tests and the same report format as the decode experiment. |
| Same payloads | The same **Go values**. Each server decodes one fixture at startup, checks it against the golden digest, and encodes that value on every request. The output is not byte-identical to the fixture file, and may differ between libraries; response size is measured and compared per library. |
| Decode experiment | `POST /ingest` is unchanged. Encoding never runs on the decode path. |
| Report | Decode and Encode are separate sections with separate verdicts. Comparisons are always between libraries for the same operation, payload and load test; decode is never compared with encode. |

## Decisions (user, 2026-11-17)

| # | Decision | Answer |
|---|---|---|
| A1 | Relation to the 2026-11-16 decode run | Run encode only; reuse the decode run. The binary change gets a threat-to-validity entry. |
| A2 | Same payload | Same Go value; no output normalization (fairness rule). Response payload size is measured and compared between libraries. |
| A3 | Pair L (`nojsonv2`) | Decode only (already measured). No Pair L encode trials. The combined report shows Pair L decode next to Pair D encode, never compared across operations. |
| A4 | Practical-significance threshold | 10 %, as for decode. |
| A5 | Encoder-based streaming | Excluded. |
| A6 | Transport | HTTP/1.1, keep-alive, no compression, loopback (as for decode). |

## Design

| # | Area | Design | Fairness rationale |
|---|---|---|---|
| D1 | Adapter | `Codec.Marshal(v any) ([]byte, error)`: stdlib `json.Marshal`, goccy plain `json.Marshal`. Forbidden: `MarshalNoEscape`, `MarshalWithOption`, `UnorderedMap`, Encoder, pooling. | Plain public API on both sides. |
| D2 | Endpoint | `GET /batch` → 200, `Content-Type: application/json`, exact `Content-Length`, encoded body. Other methods → 405 + `Allow: GET`. Query string and request body ignored. | Identical framing (no chunking). |
| D3 | Source value | `-encode-fixture <path>`: decoded at startup with the arm's own codec, verified with `conformance.Digest` against `<fixture>.golden.tsv`; mismatch → process exits, never ready. Value is shared read-only. | Both arms provably hold identical values. |
| D4 | Hot path | entry → `Marshal` (timed) → headers incl. `Server-Timing: enc;dur=,hdl;dur=` → `Write` (timed into sums only). | Encode time is known before headers; write time is not. |
| D5 | Output buffer | Fresh `[]byte` per request. | Mirrors "fresh destination". |
| D6 | Negative control | `srv_wr` (shared write code) per-op from window sums; a gap ≥ threshold downgrades the cell, like `srv_pol`. | Same host-noise guard. |
| D7 | Metrics | Counters `enc_requests`, `enc_ok`, `enc_errors`, `enc_rejected_method`, `sum_encode_ns`, `sum_write_ns`, `sum_enc_handler_ns`, `enc_resp_bytes`; k6 trends `srv_enc`, `srv_ehdl`. k6 `http_req_duration` for GET includes receiving the body. | Documented in docs/metrics.md. |
| D8 | Profiling | Same procedure as decode. | |

## Correctness (encode corpus)

| Test | Expected (documented v1 `Marshal` semantics) |
|---|---|
| Round-trip `Marshal(fixture)` → stdlib `Unmarshal` → golden digest | Must pass, both codecs, both GOEXPERIMENTs. |
| Byte equality between libraries | Reported as a finding, not required; sizes recorded. |
| HTML escaping `<>&`, U+2028/U+2029 | Escaped as `<` … ` `. |
| Invalid UTF-8 in strings | Replaced with U+FFFD. |
| Floats `1e21`, `1e-7`, `-0`, max float64 | v1 format (exponent below 1e-6 or ≥ 1e21). |
| NaN / ±Inf | Error; HTTP never 200. |
| nil slice/map/pointers | `null`. |
| Map key order | Sorted. |
| int64 > 2^53 | Exact digits. |
| Concurrency | `-race`, 64 goroutines on the shared value; value unchanged. |
| HTTP contract | 200 + exact Content-Length; 405 for POST/PUT. |

## Test matrix (Pair D only)

| Test | Encode |
|---|---|
| Microbenchmark | `BenchmarkMarshal/fixture=/codec=`, `-count=10`, benchstat. |
| Smoke | `WORKLOAD=smoke OP=encode`. |
| Pilot ramp | New, ~3 h, 1 rep (needs approval). |
| Fixed load | R50/R80 of the slower arm's encode knee, f1/f2/f3 × 5 reps (rates proposed after pilot). |
| Saturation ramp | 12 steps, 1 rep. |
| Profiling | cpu/allocs at matched encode rates. |
| Gates | §14.4 plus response-bytes check. |

## Phases

| Phase | Work | Acceptance |
|---|---|---|
| E0 | Inspect goccy `Marshal`; encode corpus TSV under both GOEXPERIMENTs. | Deviations documented, not normalized. |
| E1 | `Codec.Marshal`, `/batch`, `-encode-fixture`, digest gate, counters, Server-Timing. | `make vet test race`; decode tests unchanged. |
| E2 | `BenchmarkMarshal`; `make microbench` covers both ops. | benchstat output has both ops. |
| E3 | k6 `OP=encode`; benchctl `Cell.op`, trial path `<op>-<workload>…` for encode; response-bytes gate. | e2e matrix passes; old decode runs still parse. |
| E4 | Report: op dimension, Encode sections, `srv_wr` control, response-size table, encode corpus findings; combined report across a decode run and an encode run. | Tools tests pass; decode section unchanged on regeneration. |
| E5 | Docs: metrics, correctness policy, threat entry, ADR 0004. | |
| E6 | Encode pilot → rate proposal (approval). | |
| E7 | Authoritative encode run (approval). | |
| E8 | Combined report published, PR. | |
