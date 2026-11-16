# ADR 0002: Report two stdlib pairs, default (v2-backed) and nojsonv2 (legacy)

**Status:** accepted (2026-11-15, user decision Q1, reduced Pair L per Q11b)

## Context
[M] Under go1.27.4, `go list -f '{{.GoFiles}}' encoding/json` lists `v2_*.go` by default: the `encoding/json` package is implemented on top of the v2 engine. With `GOEXPERIMENT=nojsonv2` the same command lists `decode.go` and `scanner.go` (legacy). The import is `encoding/json` in both cases, so this is not a substitution of `encoding/json/v2`.

## Decision
- **Pair D**: `stdlib` vs `goccy`, default GOEXPERIMENT. This pair carries the full matrix.
- **Pair L**: `stdlib-legacy` vs `goccy-legacy`, `GOEXPERIMENT=nojsonv2`. This pair is reduced to F1 R50 × 5 reps plus F1 saturation and profiling. The correctness tests and microbenchmarks also run under nojsonv2.
- Both arms of a pair share an identical GOEXPERIMENT, and benchctl refuses a mismatch. Results from different pairs are never mixed.

## Consequences
The time budget grows (plan §14.7). Readers on Go ≤ 1.26 or with `nojsonv2` should read Pair L.
