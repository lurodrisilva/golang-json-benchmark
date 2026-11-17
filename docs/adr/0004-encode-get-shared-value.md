# ADR 0004: Encode experiment via GET /batch over a shared, digest-checked value

**Status:** accepted (2026-11-17, user decisions A1–A6, docs/plan/encode-plan.md)

## Context
The user asked for `Marshal` to be measured with the same payloads and the same test structure as the decode experiment. Encoding needs an in-memory value. Building it per request would put decode or construction work inside the encode path. Separately, `GET` responses carry the encoded body, which brings response delivery into the client-side timing.

## Decision
- When started with `-encode-fixture <path>`, both binaries register `GET /batch`. At startup they decode the fixture with their own codec and require the result to match the golden digest; otherwise they refuse to become ready. A selftest `Marshal` then records `encode_bytes` in the manifest. The value is shared read-only across requests.
- The handler times `Marshal` only (`enc`) and handler entry → marshalled (`ehdl`). It then writes headers, an exact `Content-Length` and the body. The write is timed into a counter that serves as a shared-code negative control.
- `Codec.Marshal` is plain `json.Marshal` / `gojson.Marshal`. There are no options, no pooling of our own, and no `MarshalNoEscape`. Default HTML escaping stays on in both.
- benchctl `op` (`decode` | `encode`) selects the operation, using the same fixtures, workloads, gates and statistics. Encode trials live under `enc-<load>` directories, so decode paths are unchanged. The report keys cells by op and compares libraries only within the same op, fixture and workload (A3). Response sizes are compared but never normalized (A2).
- Only Pair D is measured for encode. Pair L remains a decode-only arm (A3). HTTP/1.1 with keep-alive and no compression (A6). Streaming encoders are excluded (A5).

## Consequences
- Encode results describe the cost of serializing a resident value. They do not cover building that value (T14).
- Client latency includes response transfer over loopback (T15). Server-side `srv_enc` is the decoder-free figure.
- The decode run of 2026-11-16 is reused as is. Encode and decode cells come from different runs, so the report shows both manifests and never compares across operations (T16).
