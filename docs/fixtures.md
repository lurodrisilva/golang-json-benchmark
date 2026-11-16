# Fixtures (SYNTHETIC)

All three primary fixtures are **synthetic**. `tools/cmd/fixturegen` generates them deterministically from seed `20261115` (`math/rand/v2` PCG, hand-written emitter, compact output, fixed field order). They are committed to the repo and verified by `SHA256SUMS`. Nothing is downloaded or generated during a measured run.

**Why synthetic.** Common public corpora such as twitter.json and citm_catalog contain real personal data or have unclear redistribution terms. A seeded generator is redistributable, its schema can match the typed model exactly (so no bytes go to fields the model ignores), and anyone can regenerate the fixtures byte for byte (`make fixtures` produces no diff, and CI enforces this).

**Workload modelled.** A batch-ingest API: one `IngestBatch` holding many commerce/activity `Event`s, each with a user, line items, optional geo, and free-text notes.

| Fixture | Bytes (on disk = transmitted, no compression) | KiB | Events | Items | Golden digest (FNV-64a) | SHA-256 |
|---|---:|---:|---:|---:|---|---|
| `f1-64k.json` | 65,741 | 64.2 | 125 | 302 | `047c0499f1308bdb` | `f453f305…c7b8dcec` |
| `f2-256k.json` | 262,181 | 256.0 | 495 | 1,243 | `cc7b436bf29d9007` | `1df166ff…5c1dcda5` |
| `f3-1m.json` | 1,048,958 | 1024.4 | 1,969 | 4,968 | `bef50589014305ee` | `7adc3830…d3452a6d` |

F2 and F3 extend the same event stream. They are a **pure size sweep**, not a test of schema generality, which remains an open item (a structurally different fixture is optional in the plan).

**Value coverage.** The fixtures contain nested objects (`source`, `user`, `geo`), arrays (`events`, `items`, `tags`, `flags`), `map[string]string` (`attrs`), strings, int64 (including IDs > 2⁵³ in every 97th event), int32, decimals (`amount`, `unit_price`, `lat`/`lon`), booleans, nulls (`locale`, `geo`, `attrs`), escapes (`\"`, `\\`, `\n`, `\t`, `\/`, `\uXXXX` including surrogate pairs in every 5th string), and raw Unicode (Latin-1 accents, CJK, Arabic, emoji).

Example event (from f1):
```json
{"id":9007199254740993,"type":"cart_update","ts":"2026-11-01T05:00:50.820Z","user":{"id":2349622,"name":"Ægir Çelik","email":"user2349622@example.com","verified":true,"locale":"ar-EG"},"amount":1787,"currency":"EUR","items":[{"sku":"SKU-67746-A","qty":20,"unit_price":89.35,"attrs":{"región":"50% off","size":"XL"}}],"geo":null,"note":"Café — naïve résumé","flags":[]}
```

**Independent semantic assertions.** `*.golden.tsv` is written by the generator from the values it emitted, not from either decoder. It holds a structural digest and spot values (counts, null counts, first and last IDs, names, notes). `TestFixturesMatchGolden` decodes each fixture with **each** library and checks it against the golden file. `TestFixturesHaveNoUnknownFields` decodes with `DisallowUnknownFields` in both libraries, which proves the model consumes every key.

**Email addresses** all use `example.com` (RFC 2606). No data refers to real people.
