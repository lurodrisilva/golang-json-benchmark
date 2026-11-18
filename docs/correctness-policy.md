# Correctness and compatibility policy

## Two layers
1. **Raw codec layer** (`TestCorpusRaw`). Each library's `Unmarshal` is run on the 57-case corpus in `service/internal/conformance/corpus.go`. The result is compared with an expectation **written from the encoding/json v1 documented semantics**, never with the other library. This runs under both `GOEXPERIMENT` settings. Deviations are recorded in `reports/conformance/*.tsv` and reported as findings rather than normalized away. `CONFORMANCE_STRICT=1` makes them fatal.
2. **HTTP layer** (`TestCorpusHTTP`, `TestHTTPContract`). The full handler runs **with** the shared policy scanner, and the expected status codes are listed below.

## Shared, library-independent policy (user decision Q7/Q8)
`internal/policy.Check` runs before decoding in **both** binaries. The code and its cost are identical in each, and its time is reported separately as `srv_pol`.
- **Strict UTF-8.** Invalid raw UTF-8, and lone or invalid surrogate escapes (`\ud800`), are rejected with **400**. By default both libraries would silently replace these with U+FFFD.
- **No duplicate object keys**, compared after unescaping and case-sensitively at every depth: **400**. By default both libraries accept duplicates and keep the last value.
- The scanner does **not** judge syntax. On malformed input it returns OK, and the decoder rejects the input.

## Status contract
| Condition | Status |
|---|---|
| Method ≠ POST | 405 + `Allow: POST` |
| Media type ≠ `application/json` (parameters and case are ignored) | 415 |
| Body > 4 MiB (by Content-Length or while reading) | 413 |
| Body read failure | 400 |
| Invalid UTF-8 / duplicate key (policy) | 400 |
| Syntax error, trailing value or garbage, wrong type, numeric overflow, depth > 10000, empty body, BOM | 400 |
| Decoded but fails `Validate` (empty batch_id, no events, id ≤ 0, qty ≤ 0, geo out of range) | 422 |
| Success | 202, empty body, `Server-Timing` (non-minimal builds) |

Every response has an empty body. Negative-test traffic is never part of performance trials.

## Results [M] (go1.27.4, go-json v0.12.0)
- 0 deviations from the v1 spec, across 2 codecs × 57 cases × {default, nojsonv2}.
- Both libraries: reject trailing values or garbage, `{} {}` and `{}x`; reject a UTF-8 BOM; reject int64/int32 overflow, fractional numbers into ints, and `1e400`; keep the last duplicate key; match keys case-insensitively and prefer an exact match; replace invalid UTF-8 with U+FFFD (strict policy turns this into 400); accept nesting depth 10000 and reject deeper.
- Fixtures: every fixture decodes to its golden digest with both libraries, and decodes with `DisallowUnknownFields`. This proves the model consumes every field in the fixture.
- Concurrency: `TestConcurrentIsolation` and `TestCodecDoesNotRetainInput` pass under `-race`.

## Known difference that is not a correctness issue
go-json copies the input into a pooled, padded buffer before decoding. Decoded strings are copied out, so the result does not alias the request buffer (this is verified by a test). The pooled buffer is internal to the library and is a legitimate part of what is being compared.

## Encode conformance (GET /batch)
Expectations come from the documented `encoding/json` v1 `Marshal` semantics, not from comparing the two libraries (`service/internal/conformance/encode_test.go`, 33 cases):
HTML escaping of `<`, `>`, `&` and escaping of U+2028/U+2029; short and `\u00XX` control escapes; `/` not escaped; float formatting (shortest round-trip, exponent at ≥1e21 and <1e-6, `-0`, extremes); NaN/±Inf → error; exact int64/int32 extremes; nil slice/map/pointer → `null` and empty → `[]`/`{}`; map keys sorted by bytes and HTML-escaped; struct field declaration order; no whitespace.

The HTTP contract is: `GET /batch` → 200, `Content-Type: application/json`, an exact `Content-Length`, and no chunking. Any other method → 405 with `Allow: GET`. A Marshal error → 500. The encoded value is loaded once and refused at startup if its digest differs from the golden digest. It is shared read-only, and concurrent Marshal must neither change it nor produce different output (run under `-race`).

### Results [M] (go1.27.4, go-json v0.12.0)
- 0 deviations, across 2 codecs × 33 cases × {default, nojsonv2} (`reports/conformance/encode-conformance-*.tsv`).
- The two libraries produce **byte-identical** output for every fixture: f1 65,008 B, f2 260,564 B, f3 1,044,267 B. Each output decodes back to the golden digest.
- Invalid UTF-8 in a string depends on **GOEXPERIMENT**, not on the library. The default (v2-backed) build writes a raw U+FFFD. `nojsonv2` writes the escape `\ufffd`. goccy matches whichever one the build uses (it selects the behaviour with build tags). Fixtures contain only valid UTF-8, so this does not affect the benchmark.
