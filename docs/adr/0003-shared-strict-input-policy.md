# ADR 0003: Shared library-independent input policy (strict UTF-8, no duplicate keys)

**Status:** accepted (2026-11-15, user decisions Q7/Q8)

## Context
By default both libraries accept invalid UTF-8 (replacing it with U+FFFD) and duplicate keys (last one wins). The service policy requires rejecting both. If either library were configured for this, the arms would no longer be symmetric, because the options differ and goccy has no equivalent for some of them.

## Decision
`internal/policy.Check` is a recursive-descent scanner with no dependencies. It runs before `Unmarshal` in both binaries and returns 400 on invalid UTF-8, on lone or invalid surrogate escapes, and on duplicate keys (compared after unescaping). It doesn't judge syntax, which stays the decoder's job. Its cost is timed separately (`srv_pol`) and is the same in both arms.

## Consequences
- The policy adds a full extra pass over the body to every request (≈ 65 µs for 64 KiB on the M3, measured in a microbenchmark). This dilutes relative HTTP-level differences, which is why decode-only timings are reported as well.
- The policy is fuzzed against `json.Valid` + `utf8.Valid` as an oracle.
