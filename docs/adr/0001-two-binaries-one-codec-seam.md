# ADR 0001: Two binaries with a single codec seam

**Status:** accepted (2026-11-15)

## Context
We want a comparison in which the only intentional difference between the two arms is the JSON decoder. Linking both libraries into one binary, or choosing between them at runtime, would mix their code layout, binary size, GC and inlining effects, and init costs.

## Decision
- `cmd/stdlib-server` and `cmd/goccy-server` are separate `main` packages. Each is a one-liner: `os.Exit(server.Run(<codec>.Codec{}, os.Args[1:]))`.
- All HTTP, policy, validation, metrics, and lifecycle code lives in `internal/server` and is shared.
- `internal/codec.Codec` is `interface{ Unmarshal([]byte, any) error; Name() string }`. Each implementation calls the library's plain `Unmarshal` and nothing else. Pooling, `UnmarshalNoEscape`, `UnmarshalContext`, decode options, and custom `UnmarshalJSON` are all forbidden.
- Linker checks: the stdlib binary contains zero goccy packages (`go version -m`), and `build.json` records this.

## Consequences
- The interface call adds one dynamic dispatch, the same in both arms.
- The profiles come from different binaries, so they can't be diffed symbol for symbol (T12).
