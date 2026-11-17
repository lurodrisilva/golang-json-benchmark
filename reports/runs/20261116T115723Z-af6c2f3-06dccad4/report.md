# encoding/json vs goccy/go-json: deserialization benchmark report

Run `20261116T115723Z-af6c2f3-06dccad4`. This report is generated from the artifacts on disk; it contains no hand-entered numbers.

Tags: **[M]** measured · **[A]** analysis/derived · **[H]** hypothesis · **[L]** limitation.

## 1. Executive summary

- [M] Trials: 93 total. 67 valid, 21 flagged, 5 invalid/error.
- [A] Practical-significance threshold: **10%** (user-approved). A metric is called for one library only when the bootstrap 95% CI excludes 0 **and** the median-of-trials difference is at least the threshold, with ≥5 trials per arm and trial CV ≤ 10%.
- [A] Baseline is always the stdlib arm of the pair. `reduction % = (stdlib − goccy) / stdlib × 100` for lower-is-better metrics. For higher-is-better metrics (throughput) the sign is set so that positive always means goccy is better.
- [L] Hardware: one Apple M3 (4P+4E), with server and k6 sharing the host over loopback. These results are **pilot grade**. See §9.

**Headline comparisons (fixed-rate cells):**

| Cell | Metric | stdlib | goccy | reduction | 95% CI | verdict (pre-registered) | post-hoc verdict |
|---|---|---:|---:|---:|---|---|---|
| Pair D · f1-64k · fixed @2000 rps | HTTP request duration (k6) med (ms) | 0.630 | 0.347 | +44.9% | [+44.7, +45.1] | goccy better (CI excludes 0, ≥ threshold) | goccy better (CI excludes 0, ≥ threshold) |
| Pair D · f1-64k · fixed @2000 rps | HTTP request duration (k6) p(99) (ms) | 1.407 | 0.817 | +41.9% | [+30.4, +44.8] | goccy better (CI excludes 0, ≥ threshold) | goccy better (CI excludes 0, ≥ threshold) |
| Pair D · f1-64k · fixed @2000 rps | Decode duration (Unmarshal only) med (ms) | 0.407 | 0.123 | +69.8% | [+69.7, +69.8] | goccy better (CI excludes 0, ≥ threshold) | goccy better (CI excludes 0, ≥ threshold) |
| Pair D · f1-64k · fixed @2000 rps | CPU time per successful op (µs/op) | 660.4 | 377.5 | +42.8% | [+41.6, +43.1] | goccy better (CI excludes 0, ≥ threshold) | goccy better (CI excludes 0, ≥ threshold) |
| Pair D · f1-64k · fixed @2000 rps | Allocated bytes per op (B/op) | 274486 | 265467 | +3.3% | [+3.3, +3.3] | detectable but below threshold | detectable but below threshold |
| Pair D · f1-64k · fixed @3200 rps | HTTP request duration (k6) med (ms) | 0.616 | 0.336 | +45.5% | [+45.0, +45.8] | inconclusive (shared-code control differs ≥ threshold) | goccy better (CI excludes 0, ≥ threshold) |
| Pair D · f1-64k · fixed @3200 rps | HTTP request duration (k6) p(99) (ms) | 3.742 | 1.166 | +68.8% | [+62.7, +73.6] | inconclusive (trial CV > 10%) | inconclusive (trial CV > 10%) |
| Pair D · f1-64k · fixed @3200 rps | Decode duration (Unmarshal only) med (ms) | 0.413 | 0.128 | +69.0% | [+69.0, +69.0] | inconclusive (shared-code control differs ≥ threshold) | goccy better (CI excludes 0, ≥ threshold) |
| Pair D · f1-64k · fixed @3200 rps | CPU time per successful op (µs/op) | 670.8 | 380.9 | +43.2% | [+43.1, +43.3] | inconclusive (shared-code control differs ≥ threshold) | goccy better (CI excludes 0, ≥ threshold) |
| Pair D · f1-64k · fixed @3200 rps | Allocated bytes per op (B/op) | 274373 | 264713 | +3.5% | [+3.5, +3.5] | inconclusive (shared-code control differs ≥ threshold) | detectable but below threshold |
| Pair D · f2-256k · fixed @1000 rps | HTTP request duration (k6) med (ms) | 2.293 | 1.181 | +48.5% | [+48.4, +48.7] | inconclusive (shared-code control differs ≥ threshold) | inconclusive (control differs beyond resolution) |
| Pair D · f2-256k · fixed @1000 rps | HTTP request duration (k6) p(99) (ms) | 5.419 | 2.130 | +60.7% | [+54.2, +61.9] | inconclusive (shared-code control differs ≥ threshold) | inconclusive (control differs beyond resolution) |
| Pair D · f2-256k · fixed @1000 rps | Decode duration (Unmarshal only) med (ms) | 1.625 | 0.498 | +69.4% | [+69.2, +69.5] | inconclusive (shared-code control differs ≥ threshold) | inconclusive (control differs beyond resolution) |
| Pair D · f2-256k · fixed @1000 rps | CPU time per successful op (µs/op) | 2492 | 1350 | +45.8% | [+45.4, +45.9] | inconclusive (shared-code control differs ≥ threshold) | inconclusive (control differs beyond resolution) |
| Pair D · f2-256k · fixed @1000 rps | Allocated bytes per op (B/op) | 1052761 | 1053662 | -0.1% | [-0.1, -0.0] | inconclusive (shared-code control differs ≥ threshold) | inconclusive (control differs beyond resolution) |
| Pair D · f2-256k · fixed @625 rps | HTTP request duration (k6) med (ms) | 2.260 | 1.148 | +49.2% | [+49.1, +49.5] | goccy better (CI excludes 0, ≥ threshold) | goccy better (CI excludes 0, ≥ threshold) |
| Pair D · f2-256k · fixed @625 rps | HTTP request duration (k6) p(99) (ms) | 2.942 | 1.623 | +44.8% | [+38.1, +46.9] | goccy better (CI excludes 0, ≥ threshold) | goccy better (CI excludes 0, ≥ threshold) |
| Pair D · f2-256k · fixed @625 rps | Decode duration (Unmarshal only) med (ms) | 1.596 | 0.491 | +69.2% | [+69.1, +69.3] | goccy better (CI excludes 0, ≥ threshold) | goccy better (CI excludes 0, ≥ threshold) |
| Pair D · f2-256k · fixed @625 rps | CPU time per successful op (µs/op) | 2468 | 1362 | +44.8% | [+44.5, +45.0] | goccy better (CI excludes 0, ≥ threshold) | goccy better (CI excludes 0, ≥ threshold) |
| Pair D · f2-256k · fixed @625 rps | Allocated bytes per op (B/op) | 1054013 | 1061957 | -0.8% | [-0.8, -0.7] | detectable but below threshold | detectable but below threshold |
| Pair D · f3-1m · fixed @160 rps | HTTP request duration (k6) med (ms) | 9.168 | 4.527 | +50.6% | [+49.9, +51.0] | inconclusive (shared-code control differs ≥ threshold) | goccy better (CI excludes 0, ≥ threshold) |
| Pair D · f3-1m · fixed @160 rps | HTTP request duration (k6) p(99) (ms) | 9.865 | 4.886 | +50.5% | [+49.8, +50.7] | inconclusive (shared-code control differs ≥ threshold) | goccy better (CI excludes 0, ≥ threshold) |
| Pair D · f3-1m · fixed @160 rps | Decode duration (Unmarshal only) med (ms) | 6.635 | 2.110 | +68.2% | [+68.1, +68.4] | inconclusive (shared-code control differs ≥ threshold) | goccy better (CI excludes 0, ≥ threshold) |
| Pair D · f3-1m · fixed @160 rps | CPU time per successful op (µs/op) | 9666 | 5002 | +48.2% | [+47.6, +48.5] | inconclusive (shared-code control differs ≥ threshold) | goccy better (CI excludes 0, ≥ threshold) |
| Pair D · f3-1m · fixed @160 rps | Allocated bytes per op (B/op) | 4316146 | 4720560 | -9.4% | [-9.6, -9.3] | inconclusive (shared-code control differs ≥ threshold) | detectable but below threshold |
| Pair D · f3-1m · fixed @256 rps | HTTP request duration (k6) med (ms) | 9.341 | 4.628 | +50.5% | [+50.4, +50.5] | inconclusive (shared-code control differs ≥ threshold) | inconclusive (control differs beyond resolution) |
| Pair D · f3-1m · fixed @256 rps | HTTP request duration (k6) p(99) (ms) | 10.6 | 5.058 | +52.4% | [+50.1, +53.5] | inconclusive (shared-code control differs ≥ threshold) | inconclusive (control differs beyond resolution) |
| Pair D · f3-1m · fixed @256 rps | Decode duration (Unmarshal only) med (ms) | 6.885 | 2.167 | +68.5% | [+68.5, +68.6] | inconclusive (shared-code control differs ≥ threshold) | inconclusive (control differs beyond resolution) |
| Pair D · f3-1m · fixed @256 rps | CPU time per successful op (µs/op) | 9788 | 5056 | +48.3% | [+48.2, +48.5] | inconclusive (shared-code control differs ≥ threshold) | inconclusive (control differs beyond resolution) |
| Pair D · f3-1m · fixed @256 rps | Allocated bytes per op (B/op) | 4315079 | 4755454 | -10.2% | [-10.3, -10.2] | inconclusive (shared-code control differs ≥ threshold) | inconclusive (control differs beyond resolution) |
| Pair L · f1-64k · fixed @1625 rps | HTTP request duration (k6) med (ms) | 0.871 | 0.345 | +60.4% | [+60.3, +60.4] | goccy better (CI excludes 0, ≥ threshold) | goccy better (CI excludes 0, ≥ threshold) |
| Pair L · f1-64k · fixed @1625 rps | HTTP request duration (k6) p(99) (ms) | 1.723 | 0.910 | +47.2% | [+16.0, +49.6] | inconclusive (trial CV > 10%) | inconclusive (trial CV > 10%) |
| Pair L · f1-64k · fixed @1625 rps | Decode duration (Unmarshal only) med (ms) | 0.650 | 0.121 | +81.4% | [+81.2, +81.4] | goccy better (CI excludes 0, ≥ threshold) | goccy better (CI excludes 0, ≥ threshold) |
| Pair L · f1-64k · fixed @1625 rps | CPU time per successful op (µs/op) | 912.4 | 371.8 | +59.2% | [+58.2, +59.3] | goccy better (CI excludes 0, ≥ threshold) | goccy better (CI excludes 0, ≥ threshold) |
| Pair L · f1-64k · fixed @1625 rps | Allocated bytes per op (B/op) | 289567 | 265911 | +8.2% | [+8.2, +8.2] | detectable but below threshold | detectable but below threshold |

**Conditional recommendation [A]:** you can only adopt goccy/go-json on this evidence if the decode-time and CPU-per-op verdicts above are "goccy better" for the payload sizes you actually serve, **and** the correctness matrix (§3) shows no deviation that matters to your inputs. Treat cells marked exploratory, inconclusive or below threshold as "no justification for the dependency".

## 2. Environment and reproducibility manifest

Exact build metadata (binaries are preserved under `<run>/bin/`):

```json
{
  "arms": [
    {
      "binary": "bin/stdlib-minimal-server",
      "build_flags": [
        "-trimpath",
        "-buildvcs=true",
        "-tags",
        "minimal"
      ],
      "cmd": "./cmd/stdlib-server",
      "codec": "stdlib",
      "encoding_json_gofiles": "[v2_decode.go v2_encode.go v2_indent.go v2_inject.go v2_options.go v2_scanner.go v2_stream.go]",
      "go_version_m": "bin/stdlib-minimal-server: go1.27.4\n\tpath\tgithub.com/lurodrisilva/golang-json-encoding-benchmark/service/cmd/stdlib-server\n\tmod\tgithub.com/lurodrisilva/golang-json-encoding-benchmark/service\t(devel)\t\n\tbuild\t-buildmode=exe\n\tbuild\t-compiler=gc\n\tbuild\t-tags=minimal\n\tbuild\t-trimpath=true\n\tbuild\tCGO_ENABLED=0\n\tbuild\tGOARCH=arm64\n\tbuild\tGOOS=darwin\n\tbuild\tGOARM64=v8.0\n\tbuild\tvcs=git\n\tbuild\tvcs.revision=af6c2f32a252de8d7243ce3757ab2fcc83128ff5\n\tbuild\tvcs.time=2026-11-16T11:57:10Z\n\tbuild\tvcs.modified=false\n",
      "goexperiment": "",
      "name": "stdlib",
      "pair": "D",
      "sha256": "69c182a14b6fb0d43a0fe30fb971b3e18efb73f155c48397568a7fe60aae61bc",
      "tags": "minimal"
    },
    {
      "binary": "bin/goccy-minimal-server",
      "build_flags": [
        "-trimpath",
        "-buildvcs=true",
        "-tags",
        "minimal"
      ],
      "cmd": "./cmd/goccy-server",
      "codec": "goccy",
      "encoding_json_gofiles": "[v2_decode.go v2_encode.go v2_indent.go v2_inject.go v2_options.go v2_scanner.go v2_stream.go]",
      "go_version_m": "bin/goccy-minimal-server: go1.27.4\n\tpath\tgithub.com/lurodrisilva/golang-json-encoding-benchmark/service/cmd/goccy-server\n\tmod\tgithub.com/lurodrisilva/golang-json-encoding-benchmark/service\t(devel)\t\n\tdep\tgithub.com/goccy/go-json\tv0.12.0\th1:pD+oWouAiz4yruwXIt/iT+lOMeyhr1aTtxzaMP/RgVI=\n\tbuild\t-buildmode=exe\n\tbuild\t-compiler=gc\n\tbuild\t-tags=minimal\n\tbuild\t-trimpath=true\n\tbuild\tCGO_ENABLED=0\n\tbuild\tGOARCH=arm64\n\tbuild\tGOOS=darwin\n\tbuild\tGOARM64=v8.0\n\tbuild\tvcs=git\n\tbuild\tvcs.revision=af6c2f32a252de8d7243ce3757ab2fcc83128ff5\n\tbuild\tvcs.time=2026-11-16T11:57:10Z\n\tbuild\tvcs.modified=false\n",
      "goexperiment": "",
      "name": "goccy",
      "pair": "D",
      "sha256": "50219e18488b5713412736fe4ea657467ee5a3fa9ca42ec129225587c382e859",
      "tags": "minimal"
    },
    {
      "binary": "bin/stdlib-legacy-minimal-server",
      "build_flags": [
        "-trimpath",
        "-buildvcs=true",
        "-tags",
        "minimal"
      ],
      "cmd": "./cmd/stdlib-server",
      "codec": "stdlib",
      "encoding_json_gofiles": "[decode.go encode.go fold.go indent.go scanner.go stream.go tables.go tags.go]",
      "go_version_m": "bin/stdlib-legacy-minimal-server: go1.27.4-X:nojsonv2\n\tpath\tgithub.com/lurodrisilva/golang-json-encoding-benchmark/service/cmd/stdlib-server\n\tmod\tgithub.com/lurodrisilva/golang-json-encoding-benchmark/service\t(devel)\t\n\tbuild\t-buildmode=exe\n\tbuild\t-compiler=gc\n\tbuild\t-tags=minimal\n\tbuild\t-trimpath=true\n\tbuild\tCGO_ENABLED=0\n\tbuild\tGOARCH=arm64\n\tbuild\tGOEXPERIMENT=nojsonv2\n\tbuild\tGOOS=darwin\n\tbuild\tGOARM64=v8.0\n\tbuild\tvcs=git\n\tbuild\tvcs.revision=af6c2f32a252de8d7243ce3757ab2fcc83128ff5\n\tbuild\tvcs.time=2026-11-16T11:57:10Z\n\tbuild\tvcs.modified=false\n",
      "goexperiment": "nojsonv2",
      "name": "stdlib-legacy",
      "pair": "L",
      "sha256": "e9d303600b72cdba0d6b34f03aaa5f31ec345c822ef49440717feeeead025eed",
      "tags": "minimal"
    },
    {
      "binary": "bin/goccy-legacy-minimal-server",
      "build_flags": [
        "-trimpath",
        "-buildvcs=true",
        "-tags",
        "minimal"
      ],
      "cmd": "./cmd/goccy-server",
      "codec": "goccy",
      "encoding_json_gofiles": "[decode.go encode.go fold.go indent.go scanner.go stream.go tables.go tags.go]",
      "go_version_m": "bin/goccy-legacy-minimal-server: go1.27.4-X:nojsonv2\n\tpath\tgithub.com/lurodrisilva/golang-json-encoding-benchmark/service/cmd/goccy-server\n\tmod\tgithub.com/lurodrisilva/golang-json-encoding-benchmark/service\t(devel)\t\n\tdep\tgithub.com/goccy/go-json\tv0.12.0\th1:pD+oWouAiz4yruwXIt/iT+lOMeyhr1aTtxzaMP/RgVI=\n\tbuild\t-buildmode=exe\n\tbuild\t-compiler=gc\n\tbuild\t-tags=minimal\n\tbuild\t-trimpath=true\n\tbuild\tCGO_ENABLED=0\n\tbuild\tGOARCH=arm64\n\tbuild\tGOEXPERIMENT=nojsonv2\n\tbuild\tGOOS=darwin\n\tbuild\tGOARM64=v8.0\n\tbuild\tvcs=git\n\tbuild\tvcs.revision=af6c2f32a252de8d7243ce3757ab2fcc83128ff5\n\tbuild\tvcs.time=2026-11-16T11:57:10Z\n\tbuild\tvcs.modified=false\n",
      "goexperiment": "nojsonv2",
      "name": "goccy-legacy",
      "pair": "L",
      "sha256": "ed0413b28aa84c3426267ae8da8f49aca7f7750c7f4831db4a59c37f24e9f1cd",
      "tags": "minimal"
    },
    {
      "binary": "bin/stdlib-server",
      "build_flags": [
        "-trimpath",
        "-buildvcs=true"
      ],
      "cmd": "./cmd/stdlib-server",
      "codec": "stdlib",
      "encoding_json_gofiles": "[v2_decode.go v2_encode.go v2_indent.go v2_inject.go v2_options.go v2_scanner.go v2_stream.go]",
      "go_version_m": "bin/stdlib-server: go1.27.4\n\tpath\tgithub.com/lurodrisilva/golang-json-encoding-benchmark/service/cmd/stdlib-server\n\tmod\tgithub.com/lurodrisilva/golang-json-encoding-benchmark/service\t(devel)\t\n\tbuild\t-buildmode=exe
… (truncated; see file)
```

Host and toolchain:

```json
{
  "captured_at": "2026-11-16T11:57:23Z",
  "cpu": "Apple M3",
  "env": {
    "GODEBUG": "",
    "GOGC": "",
    "GOMAXPROCS": "",
    "GOMEMLIMIT": ""
  },
  "go_env": {
    "CGO_ENABLED": "1",
    "GOAMD64": "",
    "GOARCH": "arm64",
    "GOARM64": "v8.0",
    "GOEXPERIMENT": "",
    "GOFLAGS": "",
    "GOOS": "darwin",
    "GOTOOLCHAIN": "auto",
    "GOVERSION": "go1.27.4",
    "GOWORK": "go.work"
  },
  "go_version": "go version go1.27.4 darwin/arm64",
  "goarch": "arm64",
  "goos": "darwin",
  "hostname": "redacted",
  "k6_version": "k6 v2.4.0 (commit/devel, go1.27.4, darwin/arm64)",
  "memsize": "25769803776",
  "numcpu": 8,
  "os": "ProductName:\t\tmacOS\nProductVersion:\t\t26.6.2\nBuildVersion:\t\t25G83",
  "perflevel0_logicalcpu": "4",
  "perflevel1_logicalcpu": "4",
  "pmset_therm": "Note: No thermal warning level has been recorded\nNote: No performance warning level has been recorded\nNote: No CPU power status has been recorded",
  "power_source": "Now drawing from 'AC Power'\n -InternalBattery-0 (id=23527523)\t80%; AC attached; not charging present: true",
  "uname": "Darwin Lucianos-MacBook-Pro-3106.local 25.6.0 Darwin Kernel Version 25.6.0: Fri Jul 31 19:18:29 PDT 2026; root:xnu-12377.161.14~5/RELEASE_ARM64_T8122 arm64 arm Darwin"
}
```

Fixtures (SHA-256):

```
f453f3054e43233ff6176829ae8a7b3206fee6abe61fc72fdeb80574c7b8dcec  f1-64k.json
0993b493e309e6218d553337bcab964b5d349deb779adc1e55ff849f5c1dcda5  f1-64k.golden.tsv
1df166ff4d7b862b2f6fddea91f1bb279c22e2a06ea883ede55f8c577964f3c0  f2-256k.json
85dd25abe76526aac1c6772221d52f0002998e2a2aa2d07483dfbbf91135f86f  f2-256k.golden.tsv
7adc3830b32528412a4b12792a41ad4600f498e036e13cdfc011dcb9d3452a6d  f3-1m.json
ca9664fc43e082db65727353f946775ffed595212086ece8f4a1f633d3d9a43b  f3-1m.golden.tsv
```

Matrix:

```json
{
  "name": "authoritative-trimmed",
  "seed": 20261116,
  "gomaxprocs": 4,
  "cooldown": "120s",
  "pre_vus": 100,
  "max_vus": 400,
  "cells": [
    {
      "pair": [
        "stdlib",
        "goccy"
      ],
      "fixture": "f1-64k",
      "workload": "fixed",
      "rate": 2000,
      "duration": "6m",
      "warmup": "60s",
      "reps": 5
    },
    {
      "pair": [
        "stdlib",
        "goccy"
      ],
      "fixture": "f1-64k",
      "workload": "fixed",
      "rate": 3200,
      "duration": "6m",
      "warmup": "60s",
      "reps": 5
    },
    {
      "pair": [
        "stdlib",
        "goccy"
      ],
      "fixture": "f2-256k",
      "workload": "fixed",
      "rate": 625,
      "duration": "5m",
      "warmup": "60s",
      "reps": 5
    },
    {
      "pair": [
        "stdlib",
        "goccy"
      ],
      "fixture": "f2-256k",
      "workload": "fixed",
      "rate": 1000,
      "duration": "5m",
      "warmup": "60s",
      "reps": 5
    },
    {
      "pair": [
        "stdlib",
        "goccy"
      ],
      "fixture": "f3-1m",
      "workload": "fixed",
      "rate": 160,
      "duration": "6m",
      "warmup": "60s",
      "reps": 5
    },
    {
      "pair": [
        "stdlib",
        "goccy"
      ],
      "fixture": "f3-1m",
      "workload": "fixed",
      "rate": 256,
      "duration": "6m",
      "warmup": "60s",
      "reps": 5
    },
    {
      "pair": [
        "stdlib-legacy",
        "goccy-legacy"
      ],
      "fixture": "f1-64k",
      "workload": "fixed",
      "rate": 1625,
      "duration": "6m",
      "warmup": "60s",
      "reps": 5
    },
    {
      "pair": [
        "stdlib",
        "goccy"
      ],
      "fixture": "f1-64k",
      "workload": "ramp",
      "knee": 5000,
      "steps": "10,20,35,50,65,80,90,100,120,150,175,200",
      "step_dur": "2m",
      "reps": 1
    },
    {
      "pair": [
        "stdlib",
        "goccy"
      ],
      "fixture": "f2-256k",
      "workload": "ramp",
      "knee": 1250,
      "steps": "10,20,35,50,65,80,90,100,120,150,175,200",
      "step_dur": "2m",
      "reps": 1
    },
    {
      "pair": [
        "stdlib",
        "goccy"
      ],
      "fixture": "f3-1m",
      "workload": "ramp",
      "knee": 320,
      "steps": "10,20,35,50,65,80,90,100,120,150,175,200",
      "step_dur": "2m",
      "reps": 1
    },
    {
      "pair": [
        "stdlib-legacy",
        "goccy-legacy"
      ],
      "fixture": "f1-64k",
      "workload": "ramp",
      "knee": 5000,
      "steps": "10,20,35,50,65,80,90,100,120,150,175,200",
      "step_dur": "2m",
      "reps": 1
    },
    {
      "pair": [
        "stdlib",
        "goccy"
      ],
      "fixture": "f1-64k",
      "workload": "fixed",
      "rate": 3200,
      "duration": "3m",
      "warmup": "60s",
      "reps": 1,
      "profile": "cpu"
    },
    {
      "pair": [
        "stdlib",
        "goccy"
      ],
      "fixture": "f1-64k",
      "workload": "fixed",
      "rate": 3200,
      "duration": "3m",
      "warmup": "60s",
      "reps": 1,
      "profile": "allocs"
    },
    {
      "pair": [
        "stdlib",
        "goccy"
      ],
      "fixture": "f3-1m",
      "workload": "fixed",
      "rate": 256,
      "duration": "3m",
      "warmup": "60s",
      "reps": 1,
      "profile": "cpu"
    },
    {
      "pair": [
        "stdlib",
        "goccy"
      ],
      "fixture": "f3-1m",
      "workload": "fixed",
      "rate": 256,
      "duration": "3m",
      "warmup": "60s",
      "reps": 1,
      "profile": "allocs"
    },
    {
      "pair": [
        "stdlib-legacy",
        "goccy-legacy"
      ],
      "fixture": "f1-64k",
      "workload": "fixed",
      "rate": 1625,
      "duration": "3m",
      "warmup": "60s",
      "reps": 1,
      "profile": "cpu"
    }
  ]
}
```

## 3. Correctness and compatibility

- [M] `conformance-default.tsv`: 114 rows, **0 deviations** from the v1-documented expectation.
- [M] `conformance-nojsonv2.tsv`: 114 rows, **0 deviations** from the v1-documented expectation.

The full matrices are in `conformance/`. Expected outcomes are written from the encoding/json v1 documented semantics. Each library is compared against that spec, not against the other library.

## 4. Fixed-offered-load results

Every value is the **median of per-trial values**. Per-trial percentiles come from k6 for that trial alone. They are never averaged across trials or stages. P99 is shown only when each trial has ≥10000 samples. Read, policy and validate rows run **identical shared code** in both arms. They are **negative controls** and never get a library verdict. If any control median differs by ≥ 10%, the environment is too unstable to attribute differences to the library, and every library verdict in that cell is downgraded to *inconclusive*. **Post-hoc column [A]:** this rule was added after the authoritative results were seen, and it is user-approved. It applies the same downgrade, but a control counts only when its absolute median difference also exceeds 0.002 ms. Server-Timing has 0.001 ms resolution, so smaller gaps are quantization. Read the pre-registered verdict as primary.

### Pair D · f1-64k · fixed @2000 rps

- [M] trials analysed: stdlib-arm 5, goccy-arm 5.
- [M] goccy rep 1: INCLUDED but flagged [host busy 16.4% before trial (top: corespotlightd 124%)]
- [M] goccy rep 2: invalid [dropped_iterations=22]
- [M] goccy rep 4: INCLUDED but flagged [host busy 14.4% before trial (top: duetexpertd 97%)]
- [M] goccy rep 6: INCLUDED but flagged [host busy 13.8% before trial (top: XprotectService 36%, node 20%, claude 16%, WindowServer 8%, stable 5%)]
- [M] stdlib rep 1: INCLUDED but flagged [host busy 11.9% before trial (top: WindowServer 36%, loginwindow 33%, WallpaperAerialsExtension 9%)]
- [M] window: nominal 360.0 s, wall 360.4 s, drain 0.444 s (per trial; the drain is k6 finishing in-flight requests).

| Metric | unit | stdlib (trials) | goccy (trials) | stdlib median | goccy median | Δ abs | reduction | 95% CI | verdict (pre-registered) | post-hoc verdict |
|---|---|---|---|---:|---:|---:|---:|---|---|---|
| HTTP request duration (k6) min | ms | 0.020, 0.560, 0.553, 0.552, 0.559 | 0.280, 0.281, 0.283, 0.281, 0.285 | 0.553 | 0.281 | -0.272 | +49.2% | [-1305.0, +49.8] | inconclusive (trial CV > 10%) | inconclusive (trial CV > 10%) |
| HTTP request duration (k6) med | ms | 0.632, 0.631, 0.628, 0.629, 0.630 | 0.348, 0.347, 0.347, 0.346, 0.347 | 0.630 | 0.347 | -0.283 | +44.9% | [+44.7, +45.1] | goccy better (CI excludes 0, ≥ threshold) | goccy better (CI excludes 0, ≥ threshold) |
| HTTP request duration (k6) avg | ms | 0.684, 0.675, 0.670, 0.663, 0.666 | 0.379, 0.376, 0.372, 0.371, 0.399 | 0.670 | 0.376 | -0.294 | +43.8% | [+40.5, +45.0] | goccy better (CI excludes 0, ≥ threshold) | goccy better (CI excludes 0, ≥ threshold) |
| HTTP request duration (k6) p(95) | ms | 0.858, 0.838, 0.828, 0.807, 0.808 | 0.499, 0.478, 0.477, 0.479, 0.520 | 0.828 | 0.479 | -0.349 | +42.1% | [+37.2, +44.2] | goccy better (CI excludes 0, ≥ threshold) | goccy better (CI excludes 0, ≥ threshold) |
| HTTP request duration (k6) p(99) | ms | 1.480, 1.407, 1.448, 1.373, 1.382 | 0.856, 0.806, 0.805, 0.817, 0.979 | 1.407 | 0.817 | -0.590 | +41.9% | [+30.4, +44.8] | goccy better (CI excludes 0, ≥ threshold) | goccy better (CI excludes 0, ≥ threshold) |
| HTTP request duration (k6) max | ms | 61.2, 70.8, 50.3, 10.8, 17.7 | 18.4, 52.0, 12.4, 12.1, 83.6 | 50.3 | 18.4 | -31.9 | +63.5% | [-373.7, +80.2] | inconclusive (trial CV > 10%) | inconclusive (trial CV > 10%) |
| Handler duration (server) min | ms | 0.522, 0.520, 0.520, 0.519, 0.520 | 0.245, 0.240, 0.245, 0.246, 0.239 | 0.520 | 0.245 | -0.275 | +52.9% | [+52.7, +54.0] | goccy better (CI excludes 0, ≥ threshold) | goccy better (CI excludes 0, ≥ threshold) |
| Handler duration (server) med | ms | 0.547, 0.547, 0.546, 0.545, 0.546 | 0.259, 0.257, 0.257, 0.257, 0.258 | 0.546 | 0.257 | -0.289 | +52.9% | [+52.6, +53.0] | goccy better (CI excludes 0, ≥ threshold) | goccy better (CI excludes 0, ≥ threshold) |
| Handler duration (server) avg | ms | 0.575, 0.573, 0.572, 0.570, 0.571 | 0.275, 0.273, 0.272, 0.273, 0.278 | 0.572 | 0.273 | -0.299 | +52.3% | [+51.5, +52.5] | goccy better (CI excludes 0, ≥ threshold) | goccy better (CI excludes 0, ≥ threshold) |
| Handler duration (server) p(95) | ms | 0.686, 0.685, 0.683, 0.678, 0.678 | 0.351, 0.347, 0.346, 0.349, 0.364 | 0.683 | 0.349 | -0.334 | +48.9% | [+46.7, +49.4] | goccy better (CI excludes 0, ≥ threshold) | goccy better (CI excludes 0, ≥ threshold) |
| Handler duration (server) p(99) | ms | 1.166, 1.150, 1.162, 1.147, 1.150 | 0.570, 0.557, 0.560, 0.564, 0.580 | 1.150 | 0.564 | -0.586 | +51.0% | [+49.6, +52.0] | goccy better (CI excludes 0, ≥ threshold) | goccy better (CI excludes 0, ≥ threshold) |
| Handler duration (server) max | ms | 50.7, 50.3, 15.0, 8.978, 13.7 | 12.1, 47.6, 8.689, 8.277, 73.4 | 15.0 | 12.1 | -2.921 | +19.4% | [-429.9, +82.9] | inconclusive (trial CV > 10%) | inconclusive (trial CV > 10%) |
| Decode duration (Unmarshal only) min | ms | 0.390, 0.390, 0.389, 0.388, 0.388 | 0.116, 0.112, 0.116, 0.116, 0.110 | 0.389 | 0.116 | -0.273 | +70.2% | [+70.1, +71.7] | goccy better (CI excludes 0, ≥ threshold) | goccy better (CI excludes 0, ≥ threshold) |
| Decode duration (Unmarshal only) med | ms | 0.407, 0.407, 0.407, 0.406, 0.406 | 0.123, 0.123, 0.123, 0.123, 0.123 | 0.407 | 0.123 | -0.284 | +69.8% | [+69.7, +69.8] | goccy better (CI excludes 0, ≥ threshold) | goccy better (CI excludes 0, ≥ threshold) |
| Decode duration (Unmarshal only) avg | ms | 0.428, 0.428, 0.427, 0.425, 0.425 | 0.134, 0.132, 0.132, 0.132, 0.134 | 0.427 | 0.132 | -0.295 | +69.0% | [+68.5, +69.1] | goccy better (CI excludes 0, ≥ threshold) | goccy better (CI excludes 0, ≥ threshold) |
| Decode duration (Unmarshal only) p(95) | ms | 0.527, 0.526, 0.524, 0.523, 0.522 | 0.192, 0.190, 0.190, 0.192, 0.196 | 0.524 | 0.192 | -0.332 | +63.4% | [+62.6, +63.9] | goccy better (CI excludes 0, ≥ threshold) | goccy better (CI excludes 0, ≥ threshold) |
| Decode duration (Unmarshal only) p(99) | ms | 0.873, 0.862, 0.871, 0.860, 0.862 | 0.281, 0.275, 0.277, 0.279, 0.287 | 0.862 | 0.279 | -0.583 | +67.6% | [+66.7, +68.3] | goccy better (CI excludes 0, ≥ threshold) | goccy better (CI excludes 0, ≥ threshold) |
| Decode duration (Unmarshal only) max | ms | 43.8, 50.1, 14.8, 8.836, 13.3 | 11.8, 27.5, 6.936, 8.141, 67.6 | 14.8 | 11.8 | -2.961 | +20.0% | [-357.5, +83.7] | inconclusive (trial CV > 10%) | inconclusive (trial CV > 10%) |
| Body read duration (shared code, control) min | ms | 0.005, 0.004, 0.004, 0.004, 0.004 | 0.005, 0.004, 0.004, 0.004, 0.005 | 0.004 | 0.004 | 0 | +0.0% | [-25.0, +20.0] | control (shared code) | control (shared code) |
| Body read duration (shared code, control) med | ms | 0.009, 0.009, 0.009, 0.009, 0.009 | 0.009, 0.009, 0.009, 0.009, 0.010 | 0.009 | 0.009 | 0 | +0.0% | [-11.1, +0.0] | control (shared code) | control (shared code) |
| Body read duration (shared code, control) avg | ms | 0.013, 0.013, 0.012, 0.012, 0.012 | 0.013, 0.012, 0.012, 0.012, 0.014 | 0.012 | 0.012 | -0.000273 | +2.2% | [-9.8, +8.8] | control (shared code) | control (shared code) |
| Body read duration (shared code, control) p(95) | ms | 0.027, 0.026, 0.025, 0.023, 0.023 | 0.026, 0.021, 0.020, 0.020, 0.024 | 0.025 | 0.021 | -0.004 | +16.0% | [-4.3, +23.1] | control (shared code) | control (shared code) |
| Body read duration (shared code, control) p(99) | ms | 0.083, 0.089, 0.085, 0.085, 0.083 | 0.064, 0.061, 0.057, 0.058, 0.081 | 0.085 | 0.061 | -0.024 | +28.2% | [+4.7, +32.9] | control (shared code) | control (shared code) |
| Body read duration (shared code, control) max | ms | 50.0, 20.0, 9.879, 6.821, 6.937 | 10.1, 34.2, 5.977, 5.467, 60.3 | 9.879 | 10.1 | 0.244 | -2.5% | [-510.8, +79.7] | control (shared code) | control (shared code) |
| Policy scan duration (shared code, control) min | ms | 0.119, 0.119, 0.119, 0.119, 0.119 | 0.120, 0.118, 0.119, 0.120, 0.116 | 0.119 | 0.119 | 0 | +0.0% | [-0.8, +2.5] | control (shared code) | control (shared code) |
| Policy scan duration (shared code, control) med | ms | 0.129, 0.129, 0.128, 0.129, 0.129 | 0.125, 0.124, 0.125, 0.125, 0.124 | 0.129 | 0.125 | -0.004 | +3.1% | [+2.3, +3.9] | control (shared code) | control (shared code) |
| Policy scan duration (shared code, control) avg | ms | 0.133, 0.133, 0.133, 0.132, 0.133 | 0.128, 0.128, 0.128, 0.128, 0.129 | 0.133 | 0.128 | -0.00451 | +3.4% | [+2.7, +3.7] | control (shared code) | control (shared code) |
| Policy scan duration (shared code, control) p(95) | ms | 0.136, 0.136, 0.136, 0.136, 0.136 | 0.131, 0.130, 0.130, 0.130, 0.131 | 0.136 | 0.130 | -0.006 | +4.4% | [+3.7, +4.4] | control (shared code) | control (shared code) |
| Policy scan duration (shared code, control) p(99) | ms | 0.261, 0.261, 0.261, 0.260, 0.261 | 0.259, 0.257, 0.257, 0.258, 0.260 | 0.261 | 0.258 | -0.003 | +1.1% | [+0.4, +1.5] | control (shared code) | control (shared code) |
| Policy scan duration (shared code, control) max | ms | 9.795, 26.9, 6.424, 5.155, 5.254 | 5.958, 28.4, 5.242, 6.730, 69.5 | 6.424 | 6.730 | 0.306 | -4.8% | [-981.5, +75.0] | control (shared code) | control (shared code) |
| Validate duration (shared code, control) min | ms | 0, 0, 0, 0, 0 | 0, 0, 0, 0, 0 | 0 | 0 | 0 | n/a | n/a | control (shared code) | control (shared code) |
| Validate duration (shared code, control) med | ms | 0.001, 0.001, 0.001, 0.001, 0.001 | 0.001, 0.001, 0.001, 0.001, 0.001 | 0.001 | 0.001 | 0 | +0.0% | [+0.0, +0.0] | control (shared code) | control (shared code) |
| Validate duration (shared code, control) avg | ms | 0.000831, 0.000729, 0.00077, 0.000806, 0.000701 | 0.000827, 0.000743, 0.000728, 0.000753, 0.000784 | 0.00077 | 0.000753 | -1.67e-05 | +2.2% | [-11.8, +9.6] | control (shared code) | control (shared code) |
| Validate duration (shared code, control) p(95) | ms | 0.001, 0.001, 0.001, 0.001, 0.001 | 0.001, 0.001, 0.001, 0.001, 0.001 | 0.001 | 0.001 | 0 | +0.0% | [+0.0, +0.0] | control (shared code) | control (shared code) |
| Validate duration (shared code, control) p(99) | ms | 0.002, 0.002, 0.002, 0.002, 0.002 | 0.002, 0.002, 0.002, 0.002, 0.002 | 0.002 | 0.002 | 0 | +0.0% | [+0.0, +0.0] | control (shared code) | control (shared code) |
| Validate duration (shared code, control) max | ms | 0.848, 0.657, 2.031, 1.634, 0.154 | 1.568, 1.268, 1.200, 0.409, 1.101 | 0.848 | 1.200 | 0.352 | -41.5% | [-679.2, +51.8] | control (shared code) | control (shared code) |
| Achieved throughput | req/s | 2000, 2000, 2000, 2000, 2000 | 2000, 2000, 2000, 2000, 2000 | 2000 | 2000 | 0 | +0.0% | [+0.0, +0.0] | no detectable difference | no detectable difference |
| Server decoded_ok rate | op/s | 1998, 1998, 1998, 1998, 1998 | 1999, 1999, 1999, 1999, 1999 | 1998 | 1999 | 0.130 | +0.0% | [+0.0, +0.0] | detectable but below threshold | detectable but below threshold |
| CPU time per successful op | µs/op | 663.7, 661.7, 660.4, 656.4, 657.6 | 382.6, 377.0, 377.4, 377.5, 385.5 | 660.4 | 377.5 | -283.0 | +42.8% | [+41.6, +43.1] | goccy better (CI excludes 0, ≥ threshold) | goccy better (CI excludes 0, ≥ threshold) |
| CPU utilization | cores | 1.326, 1.322, 1.320, 1.312, 1.314 | 0.765, 0.754, 0.754, 0.754, 0.770 | 1.320 | 0.754 | -0.565 | +42.8% | [+41.6, +43.1] | goccy better (CI excludes 0, ≥ threshold) | goccy better (CI excludes 0, ≥ threshold) |
| CPU-seconds in window | s | 478.1, 476.7, 475.7, 472.8, 473.6 | 275.6, 271.6, 271.8, 271.9, 277.7 | 475.7 | 271.9 | -203.8 | +42.8% | [+41.6, +43.1] | goccy better (CI excludes 0, ≥ threshold) | goccy better (CI excludes 0, ≥ threshold) |
| Σ handler durations | s | 414.1, 413.1, 412.3, 410.3, 411.0 | 198.1, 196.6, 196.2, 196.3, 199.9 | 412.3 | 196.6 | -215.7 | +52.3% | [+51.5, +52.5] | goccy better (CI excludes 0, ≥ threshold) | goccy better (CI excludes 0, ≥ threshold) |
| Σ decode durations | s | 308.3, 308.0, 307.5, 305.8, 306.4 | 96.3, 95.3, 95.2, 95.3, 96.8 | 307.5 | 95.3 | -212.1 | +69.0% | [+68.5, +69.1] | goccy better (CI excludes 0, ≥ threshold) | goccy better (CI excludes 0, ≥ threshold) |
| Allocated bytes per op | B/op | 274484, 274488, 274486, 274486, 274486 | 265430, 265490, 265467, 265485, 265427 | 274486 | 265467 | -9019 | +3.3% | [+3.3, +3.3] | detectable but below threshold | detectable but below threshold |
| Allocated objects per op | obj/op | 2394, 2394, 2394, 2394, 2394 | 2172, 2172, 2172, 2172, 2172 | 2394 | 2172 | -221.5 | +9.3% | [+9.3, +9.3] | detectable but below threshold | detectable but below threshold |
| GC cycles per 1k ops | cycles | 19.1, 19.1, 19.1, 19.0, 19.1 | 19.6, 19.6, 19.5, 19.5, 19.5 | 19.1 | 19.5 | 0.463 | -2.4% | [-2.6, -2.2] | detectable but below threshold | detectable but below threshold |
| GC pause total in window | s | 0.511, 0.499, 0.497, 0.464, 0.466 | 0.361, 0.325, 0.324, 0.330, 0.438 | 0.497 | 0.330 | -0.167 | +33.7% | [+12.0, +35.5] | inconclusive (trial CV > 10%) | inconclusive (trial CV > 10%) |
| Live heap at window end | MiB | 11.2, 11.2, 10.8, 11.3, 11.4 | 11.7, 11.5, 11.5, 11.5, 11.5 | 11.2 | 11.5 | 0.282 | -2.5% | [-6.7, -1.0] | detectable but below threshold | detectable but below threshold |
| Runtime-mapped memory at end | MiB | 83.2, 79.5, 75.2, 75.2, 79.2 | 79.5, 75.2, 79.5, 75.7, 83.7 | 79.2 | 79.5 | 0.250 | -0.3% | [-5.7, +5.3] | no detectable difference | no detectable difference |
| Peak RSS (ps, 5 s samples) | MiB | 86.1, 84.9, 83.5, 82.7, 84.5 | 85.4, 83.2, 84.5, 84.3, 89.5 | 84.5 | 84.5 | -0.016 | +0.0% | [-6.0, +2.0] | no detectable difference | no detectable difference |
| Max RSS (getrusage) | MiB | 86.1, 84.9, 83.5, 82.7, 84.5 | 85.4, 83.2, 84.5, 84.3, 89.5 | 84.5 | 84.5 | -0.016 | +0.0% | [-6.0, +2.0] | no detectable difference | no detectable difference |

### Pair D · f1-64k · fixed @3200 rps

- [M] trials analysed: stdlib-arm 5, goccy-arm 5.
- [M] window: nominal 360.0 s, wall 360.7 s, drain 0.680 s (per trial; the drain is k6 finishing in-flight requests).

| Metric | unit | stdlib (trials) | goccy (trials) | stdlib median | goccy median | Δ abs | reduction | 95% CI | verdict (pre-registered) | post-hoc verdict |
|---|---|---|---|---:|---:|---:|---:|---|---|---|
| HTTP request duration (k6) min | ms | 0.564, 0.565, 0.562, 0.020, 0.561 | 0.288, 0.278, 0.280, 0.280, 0.281 | 0.562 | 0.280 | -0.282 | +50.2% | [-1300.0, +50.5] | inconclusive (trial CV > 10%) | inconclusive (trial CV > 10%) |
| HTTP request duration (k6) med | ms | 0.616, 0.617, 0.616, 0.615, 0.616 | 0.336, 0.335, 0.338, 0.339, 0.334 | 0.616 | 0.336 | -0.280 | +45.5% | [+45.0, +45.8] | inconclusive (shared-code control differs ≥ threshold) (was: goccy better (CI excludes 0, ≥ threshold); controls: Body read duration (shared code, control) med -11.1%) | goccy better (CI excludes 0, ≥ threshold) |
| HTTP request duration (k6) avg | ms | 0.767, 0.883, 0.817, 0.751, 0.779 | 0.391, 0.410, 0.406, 0.420, 0.388 | 0.779 | 0.406 | -0.372 | +47.8% | [+45.4, +54.0] | inconclusive (shared-code control differs ≥ threshold) (was: goccy better (CI excludes 0, ≥ threshold); controls: Body read duration (shared code, control) med -11.1%) | goccy better (CI excludes 0, ≥ threshold) |
| HTTP request duration (k6) p(95) | ms | 1.261, 1.342, 1.321, 1.259, 1.300 | 0.686, 0.707, 0.693, 0.682, 0.669 | 1.300 | 0.686 | -0.614 | +47.2% | [+45.0, +49.2] | inconclusive (shared-code control differs ≥ threshold) (was: goccy better (CI excludes 0, ≥ threshold); controls: Body read duration (shared code, control) med -11.1%) | goccy better (CI excludes 0, ≥ threshold) |
| HTTP request duration (k6) p(99) | ms | 3.235, 4.304, 4.227, 3.174, 3.742 | 1.120, 1.311, 1.166, 1.184, 1.117 | 3.742 | 1.166 | -2.576 | +68.8% | [+62.7, +73.6] | inconclusive (trial CV > 10%) | inconclusive (trial CV > 10%) |
| HTTP request duration (k6) max | ms | 92.0, 263.2, 141.9, 40.7, 52.7 | 31.7, 79.9, 80.9, 118.9, 31.9 | 92.0 | 79.9 | -12.1 | +13.2% | [-98.9, +77.7] | inconclusive (trial CV > 10%) | inconclusive (trial CV > 10%) |
| Handler duration (server) min | ms | 0.524, 0.524, 0.525, 0.524, 0.525 | 0.244, 0.243, 0.244, 0.244, 0.243 | 0.524 | 0.244 | -0.280 | +53.4% | [+53.4, +53.7] | inconclusive (shared-code control differs ≥ threshold) (was: goccy better (CI excludes 0, ≥ threshold); controls: Body read duration (shared code, control) med -11.1%) | goccy better (CI excludes 0, ≥ threshold) |
| Handler duration (server) med | ms | 0.547, 0.547, 0.547, 0.547, 0.547 | 0.264, 0.263, 0.264, 0.264, 0.264 | 0.547 | 0.264 | -0.283 | +51.7% | [+51.7, +51.9] | inconclusive (shared-code control differs ≥ threshold) (was: goccy better (CI excludes 0, ≥ threshold); controls: Body read duration (shared code, control) med -11.1%) | goccy better (CI excludes 0, ≥ threshold) |
| Handler duration (server) avg | ms | 0.598, 0.612, 0.606, 0.599, 0.600 | 0.287, 0.290, 0.289, 0.289, 0.287 | 0.600 | 0.289 | -0.312 | +51.9% | [+51.6, +52.8] | inconclusive (shared-code control differs ≥ threshold) (was: goccy better (CI excludes 0, ≥ threshold); controls: Body read duration (shared code, control) med -11.1%) | goccy better (CI excludes 0, ≥ threshold) |
| Handler duration (server) p(95) | ms | 0.741, 0.747, 0.746, 0.757, 0.747 | 0.387, 0.397, 0.392, 0.394, 0.391 | 0.747 | 0.392 | -0.355 | +47.5% | [+46.8, +48.2] | inconclusive (shared-code control differs ≥ threshold) (was: goccy better (CI excludes 0, ≥ threshold); controls: Body read duration (shared code, control) med -11.1%) | goccy better (CI excludes 0, ≥ threshold) |
| Handler duration (server) p(99) | ms | 1.333, 1.394, 1.372, 1.335, 1.362 | 0.622, 0.664, 0.628, 0.627, 0.619 | 1.362 | 0.627 | -0.735 | +54.0% | [+51.2, +55.0] | inconclusive (shared-code control differs ≥ threshold) (was: goccy better (CI excludes 0, ≥ threshold); controls: Body read duration (shared code, control) med -11.1%) | goccy better (CI excludes 0, ≥ threshold) |
| Handler duration (server) max | ms | 69.0, 139.7, 114.3, 23.9, 37.3 | 20.4, 79.3, 74.8, 94.8, 25.8 | 69.0 | 74.8 | 5.887 | -8.5% | [-213.6, +81.5] | inconclusive (trial CV > 10%) | inconclusive (trial CV > 10%) |
| Decode duration (Unmarshal only) min | ms | 0.392, 0.391, 0.394, 0.394, 0.394 | 0.116, 0.116, 0.116, 0.117, 0.116 | 0.394 | 0.116 | -0.278 | +70.6% | [+70.3, +70.6] | inconclusive (shared-code control differs ≥ threshold) (was: goccy better (CI excludes 0, ≥ threshold); controls: Body read duration (shared code, control) med -11.1%) | goccy better (CI excludes 0, ≥ threshold) |
| Decode duration (Unmarshal only) med | ms | 0.413, 0.413, 0.413, 0.413, 0.413 | 0.128, 0.128, 0.128, 0.128, 0.128 | 0.413 | 0.128 | -0.285 | +69.0% | [+69.0, +69.0] | inconclusive (shared-code control differs ≥ threshold) (was: goccy better (CI excludes 0, ≥ threshold); controls: Body read duration (shared code, control) med -11.1%) | goccy better (CI excludes 0, ≥ threshold) |
| Decode duration (Unmarshal only) avg | ms | 0.448, 0.452, 0.452, 0.448, 0.449 | 0.141, 0.143, 0.143, 0.143, 0.142 | 0.449 | 0.143 | -0.306 | +68.2% | [+68.1, +68.7] | inconclusive (shared-code control differs ≥ threshold) (was: goccy better (CI excludes 0, ≥ threshold); controls: Body read duration (shared code, control) med -11.1%) | goccy better (CI excludes 0, ≥ threshold) |
| Decode duration (Unmarshal only) p(95) | ms | 0.561, 0.566, 0.565, 0.572, 0.566 | 0.203, 0.208, 0.206, 0.207, 0.207 | 0.566 | 0.207 | -0.359 | +63.4% | [+63.1, +64.1] | inconclusive (shared-code control differs ≥ threshold) (was: goccy better (CI excludes 0, ≥ threshold); controls: Body read duration (shared code, control) med -11.1%) | goccy better (CI excludes 0, ≥ threshold) |
| Decode duration (Unmarshal only) p(99) | ms | 0.957, 0.980, 0.971, 0.960, 0.971 | 0.305, 0.316, 0.307, 0.307, 0.305 | 0.971 | 0.307 | -0.664 | +68.4% | [+67.5, +68.7] | inconclusive (shared-code control differs ≥ threshold) (was: goccy better (CI excludes 0, ≥ threshold); controls: Body read duration (shared code, control) med -11.1%) | goccy better (CI excludes 0, ≥ threshold) |
| Decode duration (Unmarshal only) max | ms | 67.6, 110.4, 114.1, 21.6, 37.0 | 18.6, 77.9, 74.7, 88.8, 22.3 | 67.6 | 74.7 | 7.098 | -10.5% | [-246.2, +80.5] | inconclusive (trial CV > 10%) | inconclusive (trial CV > 10%) |
| Body read duration (shared code, control) min | ms | 0.004, 0.004, 0.004, 0.004, 0.004 | 0.004, 0.004, 0.004, 0.004, 0.004 | 0.004 | 0.004 | 0 | +0.0% | [+0.0, +0.0] | control (shared code) | control (shared code) |
| Body read duration (shared code, control) med | ms | 0.009, 0.009, 0.009, 0.009, 0.009 | 0.010, 0.010, 0.010, 0.010, 0.010 | 0.009 | 0.010 | 0.001 | -11.1% | [-11.1, -11.1] | control (shared code) (control differs ≥ threshold (within timing resolution: ignored post-hoc)) | control (shared code) |
| Body read duration (shared code, control) avg | ms | 0.017, 0.024, 0.020, 0.016, 0.018 | 0.015, 0.015, 0.015, 0.016, 0.014 | 0.018 | 0.015 | -0.00231 | +13.2% | [+6.8, +37.8] | control (shared code) | control (shared code) |
| Body read duration (shared code, control) p(95) | ms | 0.028, 0.029, 0.029, 0.029, 0.028 | 0.030, 0.029, 0.030, 0.029, 0.028 | 0.029 | 0.029 | 0 | +0.0% | [-7.1, +3.4] | control (shared code) | control (shared code) |
| Body read duration (shared code, control) p(99) | ms | 0.109, 0.111, 0.111, 0.110, 0.111 | 0.094, 0.096, 0.096, 0.094, 0.092 | 0.111 | 0.094 | -0.017 | +15.3% | [+12.7, +17.1] | control (shared code) | control (shared code) |
| Body read duration (shared code, control) max | ms | 40.2, 109.1, 110.8, 22.5, 33.2 | 19.6, 39.5, 46.3, 94.4, 10.8 | 40.2 | 39.5 | -0.692 | +1.7% | [-134.8, +82.3] | control (shared code) | control (shared code) |
| Policy scan duration (shared code, control) min | ms | 0.119, 0.119, 0.119, 0.120, 0.119 | 0.118, 0.119, 0.118, 0.119, 0.118 | 0.119 | 0.118 | -0.001 | +0.8% | [+0.0, +1.7] | control (shared code) | control (shared code) |
| Policy scan duration (shared code, control) med | ms | 0.124, 0.124, 0.124, 0.124, 0.124 | 0.124, 0.124, 0.124, 0.124, 0.124 | 0.124 | 0.124 | 0 | +0.0% | [+0.0, +0.0] | control (shared code) | control (shared code) |
| Policy scan duration (shared code, control) avg | ms | 0.133, 0.135, 0.134, 0.133, 0.133 | 0.130, 0.131, 0.130, 0.130, 0.130 | 0.133 | 0.130 | -0.0032 | +2.4% | [+1.7, +3.4] | control (shared code) | control (shared code) |
| Policy scan duration (shared code, control) p(95) | ms | 0.170, 0.173, 0.173, 0.180, 0.173 | 0.136, 0.140, 0.137, 0.139, 0.137 | 0.173 | 0.137 | -0.036 | +20.8% | [+19.1, +23.9] | control (shared code) | control (shared code) |
| Policy scan duration (shared code, control) p(99) | ms | 0.273, 0.276, 0.275, 0.273, 0.275 | 0.267, 0.271, 0.268, 0.268, 0.268 | 0.275 | 0.268 | -0.007 | +2.5% | [+1.5, +2.9] | control (shared code) | control (shared code) |
| Policy scan duration (shared code, control) max | ms | 37.6, 91.1, 36.1, 14.6, 31.2 | 7.012, 40.4, 24.3, 32.8, 8.954 | 36.1 | 24.3 | -11.8 | +32.8% | [-66.3, +81.3] | control (shared code) | control (shared code) |
| Validate duration (shared code, control) min | ms | 0, 0, 0, 0, 0 | 0, 0, 0, 0, 0 | 0 | 0 | 0 | n/a | n/a | control (shared code) | control (shared code) |
| Validate duration (shared code, control) med | ms | 0.001, 0.001, 0.001, 0.001, 0.001 | 0.001, 0.001, 0.001, 0.001, 0.001 | 0.001 | 0.001 | 0 | +0.0% | [+0.0, +0.0] | control (shared code) | control (shared code) |
| Validate duration (shared code, control) avg | ms | 0.000803, 0.000821, 0.000809, 0.000851, 0.000817 | 0.00091, 0.000875, 0.000949, 0.00092, 0.000929 | 0.000817 | 0.00092 | 0.000103 | -12.6% | [-16.1, -6.9] | control (shared code) | control (shared code) |
| Validate duration (shared code, control) p(95) | ms | 0.001, 0.001, 0.001, 0.001, 0.001 | 0.001, 0.001, 0.001, 0.001, 0.001 | 0.001 | 0.001 | 0 | +0.0% | [+0.0, +0.0] | control (shared code) | control (shared code) |
| Validate duration (shared code, control) p(99) | ms | 0.002, 0.002, 0.002, 0.002, 0.002 | 0.002, 0.002, 0.002, 0.002, 0.002 | 0.002 | 0.002 | 0 | +0.0% | [+0.0, +0.0] | control (shared code) | control (shared code) |
| Validate duration (shared code, control) max | ms | 3.563, 2.075, 1.998, 2.438, 1.695 | 0.694, 0.444, 0.733, 0.409, 0.737 | 2.075 | 0.694 | -1.381 | +66.6% | [+59.1, +83.2] | control (shared code) | control (shared code) |
| Achieved throughput | req/s | 3200, 3200, 3200, 3200, 3200 | 3200, 3200, 3200, 3200, 3200 | 3200 | 3200 | 0.00278 | +0.0% | [+0.0, +0.0] | no detectable difference | no detectable difference |
| Server decoded_ok rate | op/s | 3196, 3196, 3196, 3196, 3196 | 3196, 3196, 3196, 3196, 3196 | 3196 | 3196 | 0.124 | +0.0% | [+0.0, +0.0] | inconclusive (shared-code control differs ≥ threshold) (was: detectable but below threshold; controls: Body read duration (shared code, control) med -11.1%) | detectable but below threshold |
| CPU time per successful op | µs/op | 669.8, 670.8, 670.0, 671.7, 670.8 | 380.7, 381.6, 381.2, 380.9, 380.7 | 670.8 | 380.9 | -289.8 | +43.2% | [+43.1, +43.3] | inconclusive (shared-code control differs ≥ threshold) (was: goccy better (CI excludes 0, ≥ threshold); controls: Body read duration (shared code, control) med -11.1%) | goccy better (CI excludes 0, ≥ threshold) |
| CPU utilization | cores | 2.141, 2.144, 2.141, 2.147, 2.144 | 1.217, 1.220, 1.218, 1.217, 1.217 | 2.144 | 1.217 | -0.926 | +43.2% | [+43.1, +43.3] | inconclusive (shared-code control differs ≥ threshold) (was: goccy better (CI excludes 0, ≥ threshold); controls: Body read duration (shared code, control) med -11.1%) | goccy better (CI excludes 0, ≥ threshold) |
| CPU-seconds in window | s | 772.1, 773.2, 772.3, 774.2, 773.2 | 438.9, 439.9, 439.4, 439.1, 438.8 | 773.2 | 439.1 | -334.1 | +43.2% | [+43.1, +43.3] | inconclusive (shared-code control differs ≥ threshold) (was: goccy better (CI excludes 0, ≥ threshold); controls: Body read duration (shared code, control) med -11.1%) | goccy better (CI excludes 0, ≥ threshold) |
| Σ handler durations | s | 689.3, 705.1, 698.5, 690.1, 692.1 | 330.4, 334.4, 332.6, 333.6, 330.3 | 692.1 | 332.6 | -359.5 | +51.9% | [+51.6, +52.8] | inconclusive (shared-code control differs ≥ threshold) (was: goccy better (CI excludes 0, ≥ threshold); controls: Body read duration (shared code, control) med -11.1%) | goccy better (CI excludes 0, ≥ threshold) |
| Σ decode durations | s | 516.1, 520.9, 520.7, 517.0, 517.5 | 163.0, 165.0, 164.4, 164.6, 163.3 | 517.5 | 164.4 | -353.1 | +68.2% | [+68.1, +68.7] | inconclusive (shared-code control differs ≥ threshold) (was: goccy better (CI excludes 0, ≥ threshold); controls: Body read duration (shared code, control) med -11.1%) | goccy better (CI excludes 0, ≥ threshold) |
| Allocated bytes per op | B/op | 274372, 274372, 274373, 274373, 274374 | 264715, 264700, 264701, 264713, 264713 | 274373 | 264713 | -9660 | +3.5% | [+3.5, +3.5] | inconclusive (shared-code control differs ≥ threshold) (was: detectable but below threshold; controls: Body read duration (shared code, control) med -11.1%) | detectable but below threshold |
| Allocated objects per op | obj/op | 2393, 2393, 2393, 2393, 2393 | 2172, 2172, 2172, 2172, 2172 | 2393 | 2172 | -220.7 | +9.2% | [+9.2, +9.2] | inconclusive (shared-code control differs ≥ threshold) (was: detectable but below threshold; controls: Body read duration (shared code, control) med -11.1%) | detectable but below threshold |
| GC cycles per 1k ops | cycles | 11.2, 11.2, 11.2, 11.2, 11.2 | 11.9, 11.9, 11.9, 11.9, 11.9 | 11.2 | 11.9 | 0.730 | -6.5% | [-6.8, -6.3] | inconclusive (shared-code control differs ≥ threshold) (was: detectable but below threshold; controls: Body read duration (shared code, control) med -11.1%) | detectable but below threshold |
| GC pause total in window | s | 0.733, 0.782, 0.760, 0.724, 0.749 | 0.499, 0.548, 0.520, 0.551, 0.500 | 0.749 | 0.520 | -0.229 | +30.6% | [+24.9, +34.3] | inconclusive (shared-code control differs ≥ threshold) (was: goccy better (CI excludes 0, ≥ threshold); controls: Body read duration (shared code, control) med -11.1%) | goccy better (CI excludes 0, ≥ threshold) |
| Live heap at window end | MiB | 20.3, 19.0, 19.0, 19.0, 19.1 | 18.8, 19.0, 18.2, 17.7, 18.6 | 19.0 | 18.6 | -0.432 | +2.3% | [-0.0, +8.3] | no detectable difference | no detectable difference |
| Runtime-mapped memory at end | MiB | 129.2, 128.7, 124.7, 128.9, 125.1 | 125.0, 124.7, 121.2, 133.2, 124.7 | 128.7 | 124.7 | -3.934 | +3.1% | [-3.5, +5.8] | no detectable difference | no detectable difference |
| Peak RSS (ps, 5 s samples) | MiB | 133.2, 134.7, 131.4, 133.0, 132.0 | 129.3, 128.9, 127.3, 138.5, 130.3 | 133.0 | 129.3 | -3.688 | +2.8% | [-4.1, +4.3] | no detectable difference | no detectable difference |
| Max RSS (getrusage) | MiB | 133.2, 134.7, 131.4, 133.0, 132.0 | 129.3, 128.9, 127.3, 138.5, 130.3 | 133.0 | 129.3 | -3.688 | +2.8% | [-4.1, +4.3] | no detectable difference | no detectable difference |

### Pair D · f1-64k · fixed @3200 rps · profile=allocs

- [M] trials analysed: stdlib-arm 1, goccy-arm 1.
- [M] stdlib rep 1: INCLUDED but flagged [host busy 13.3% before trial (top: node 36%, claude 27%, WindowServer 12%, benchctl 7%, stable 7%) profiling active: not a primary latency trial]
- [M] window: nominal 180.0 s, wall 180.6 s, drain 0.617 s (per trial; the drain is k6 finishing in-flight requests).

| Metric | unit | stdlib (trials) | goccy (trials) | stdlib median | goccy median | Δ abs | reduction | 95% CI | verdict (pre-registered) | post-hoc verdict |
|---|---|---|---|---:|---:|---:|---:|---|---|---|
| HTTP request duration (k6) min | ms | 0.574 | 0.282 | 0.574 | 0.282 | -0.292 | +50.9% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| HTTP request duration (k6) med | ms | 0.617 | 0.341 | 0.617 | 0.341 | -0.276 | +44.7% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| HTTP request duration (k6) avg | ms | 0.802 | 0.674 | 0.802 | 0.674 | -0.128 | +16.0% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| HTTP request duration (k6) p(95) | ms | 1.377 | 0.787 | 1.377 | 0.787 | -0.590 | +42.8% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| HTTP request duration (k6) p(99) | ms | 3.703 | 1.480 | 3.703 | 1.480 | -2.223 | +60.0% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| HTTP request duration (k6) max | ms | 72.1 | 302.2 | 72.1 | 302.2 | 230.2 | -319.4% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Handler duration (server) min | ms | 0.525 | 0.245 | 0.525 | 0.245 | -0.280 | +53.3% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Handler duration (server) med | ms | 0.548 | 0.265 | 0.548 | 0.265 | -0.283 | +51.6% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Handler duration (server) avg | ms | 0.602 | 0.309 | 0.602 | 0.309 | -0.293 | +48.7% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Handler duration (server) p(95) | ms | 0.766 | 0.405 | 0.766 | 0.405 | -0.361 | +47.1% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Handler duration (server) p(99) | ms | 1.363 | 0.675 | 1.363 | 0.675 | -0.688 | +50.5% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Handler duration (server) max | ms | 62.1 | 243.7 | 62.1 | 243.7 | 181.5 | -292.1% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Decode duration (Unmarshal only) min | ms | 0.394 | 0.117 | 0.394 | 0.117 | -0.277 | +70.3% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Decode duration (Unmarshal only) med | ms | 0.413 | 0.129 | 0.413 | 0.129 | -0.284 | +68.8% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Decode duration (Unmarshal only) avg | ms | 0.449 | 0.151 | 0.449 | 0.151 | -0.298 | +66.3% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Decode duration (Unmarshal only) p(95) | ms | 0.571 | 0.205 | 0.571 | 0.205 | -0.366 | +64.1% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Decode duration (Unmarshal only) p(99) | ms | 0.965 | 0.318 | 0.965 | 0.318 | -0.647 | +67.0% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Decode duration (Unmarshal only) max | ms | 28.8 | 139.4 | 28.8 | 139.4 | 110.6 | -383.8% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Body read duration (shared code, control) min | ms | 0.004 | 0.004 | 0.004 | 0.004 | 0 | +0.0% | n/a | control (shared code) | control (shared code) |
| Body read duration (shared code, control) med | ms | 0.009 | 0.010 | 0.009 | 0.010 | 0.001 | -11.1% | n/a | control (shared code) (control differs ≥ threshold (within timing resolution: ignored post-hoc)) | control (shared code) |
| Body read duration (shared code, control) avg | ms | 0.018 | 0.022 | 0.018 | 0.022 | 0.00367 | -19.9% | n/a | control (shared code) | control (shared code) |
| Body read duration (shared code, control) p(95) | ms | 0.033 | 0.033 | 0.033 | 0.033 | 0 | +0.0% | n/a | control (shared code) | control (shared code) |
| Body read duration (shared code, control) p(99) | ms | 0.116 | 0.101 | 0.116 | 0.101 | -0.015 | +12.9% | n/a | control (shared code) | control (shared code) |
| Body read duration (shared code, control) max | ms | 22.7 | 215.3 | 22.7 | 215.3 | 192.6 | -850.0% | n/a | control (shared code) | control (shared code) |
| Policy scan duration (shared code, control) min | ms | 0.120 | 0.119 | 0.120 | 0.119 | -0.001 | +0.8% | n/a | control (shared code) | control (shared code) |
| Policy scan duration (shared code, control) med | ms | 0.124 | 0.124 | 0.124 | 0.124 | 0 | +0.0% | n/a | control (shared code) | control (shared code) |
| Policy scan duration (shared code, control) avg | ms | 0.134 | 0.135 | 0.134 | 0.135 | 0.00092 | -0.7% | n/a | control (shared code) | control (shared code) |
| Policy scan duration (shared code, control) p(95) | ms | 0.181 | 0.144 | 0.181 | 0.144 | -0.037 | +20.4% | n/a | control (shared code) | control (shared code) |
| Policy scan duration (shared code, control) p(99) | ms | 0.278 | 0.274 | 0.278 | 0.274 | -0.004 | +1.4% | n/a | control (shared code) | control (shared code) |
| Policy scan duration (shared code, control) max | ms | 43.3 | 243.5 | 43.3 | 243.5 | 200.2 | -461.9% | n/a | control (shared code) | control (shared code) |
| Validate duration (shared code, control) min | ms | 0 | 0 | 0 | 0 | 0 | n/a | n/a | control (shared code) | control (shared code) |
| Validate duration (shared code, control) med | ms | 0.001 | 0.001 | 0.001 | 0.001 | 0 | +0.0% | n/a | control (shared code) | control (shared code) |
| Validate duration (shared code, control) avg | ms | 0.000833 | 0.000906 | 0.000833 | 0.000906 | 7.38e-05 | -8.9% | n/a | control (shared code) | control (shared code) |
| Validate duration (shared code, control) p(95) | ms | 0.001 | 0.001 | 0.001 | 0.001 | 0 | +0.0% | n/a | control (shared code) | control (shared code) |
| Validate duration (shared code, control) p(99) | ms | 0.002 | 0.002 | 0.002 | 0.002 | 0 | +0.0% | n/a | control (shared code) | control (shared code) |
| Validate duration (shared code, control) max | ms | 1.903 | 0.577 | 1.903 | 0.577 | -1.326 | +69.7% | n/a | control (shared code) | control (shared code) |
| Achieved throughput | req/s | 3200 | 3200 | 3200 | 3200 | 0 | +0.0% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Server decoded_ok rate | op/s | 3194 | 3195 | 3194 | 3195 | 0.399 | +0.0% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| CPU time per successful op | µs/op | 670.5 | 382.5 | 670.5 | 382.5 | -288.0 | +43.0% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| CPU utilization | cores | 2.142 | 1.222 | 2.142 | 1.222 | -0.920 | +43.0% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| CPU-seconds in window | s | 386.8 | 220.7 | 386.8 | 220.7 | -166.1 | +42.9% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Σ handler durations | s | 347.6 | 178.3 | 347.6 | 178.3 | -169.3 | +48.7% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Σ decode durations | s | 259.1 | 87.2 | 259.1 | 87.2 | -171.9 | +66.3% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Allocated bytes per op | B/op | 274368 | 264597 | 274368 | 264597 | -9771 | +3.6% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Allocated objects per op | obj/op | 2393 | 2172 | 2393 | 2172 | -220.6 | +9.2% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| GC cycles per 1k ops | cycles | 9.079 | 9.651 | 9.079 | 9.651 | 0.572 | -6.3% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| GC pause total in window | s | 0.308 | 0.237 | 0.308 | 0.237 | -0.071 | +23.1% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Live heap at window end | MiB | 19.1 | 18.4 | 19.1 | 18.4 | -0.686 | +3.6% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Runtime-mapped memory at end | MiB | 129.0 | 141.3 | 129.0 | 141.3 | 12.3 | -9.5% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Peak RSS (ps, 5 s samples) | MiB | 132.2 | 146.7 | 132.2 | 146.7 | 14.5 | -11.0% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Max RSS (getrusage) | MiB | 132.2 | 146.7 | 132.2 | 146.7 | 14.5 | -11.0% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |

<details><summary>[M] stdlib rep 1 allocs profile: 192011 decoded ops in the window</summary>

```
File: stdlib-server
Type: alloc_space
Time: 2026-11-16 08:59:24 -03
Duration: 60s, Total samples = 50069.59MB 
Showing nodes accounting for 49466.12MB, 98.79% of 50069.59MB total
Dropped 83 nodes (cum <= 250.35MB)
Showing top 20 nodes out of 31
      flat  flat%   sum%        cum   cum%
15957.98MB 31.87% 31.87% 15957.98MB 31.87%  reflect.growslice
13518.49MB 27.00% 58.87% 13523.49MB 27.01%  github.com/lurodrisilva/golang-json-encoding-benchmark/service/internal/server.readBody
12385.90MB 24.74% 83.61% 12385.90MB 24.74%  reflect.mapassign_faststr0
 2692.05MB  5.38% 88.98%  2692.05MB  5.38%  encoding/json/v2.makeString
 2118.10MB  4.23% 93.22%  2118.10MB  4.23%  reflect.makemap
 1936.53MB  3.87% 97.08%  1936.53MB  3.87%  reflect.unsafe_New
  785.53MB  1.57% 98.65%   785.53MB  1.57%  slices.Grow[go.shape.[]uint8,go.shape.uint8] (inline)
   41.01MB 0.082% 98.73%   270.06MB  0.54%  net/http.(*conn).readRequest
      20MB  0.04% 98.77% 49765.77MB 99.39%  github.com/lurodrisilva/golang-json-encoding-benchmark/service/internal/server.(*ingestHandler).ServeHTTP
    8.04MB 0.016% 98.79%  3485.62MB  6.96%  encoding/json/v2.makeStringArshaler.func2
    2.50MB 0.005% 98.79% 50055.09MB   100%  net/http.(*conn).serve
         0     0% 98.79% 35885.12MB 71.67%  encoding/json.Unmarshal (inline)
         0     0% 98.79%   785.53MB  1.57%  encoding/json/internal/jsonwire.AppendUnquote
         0     0% 98.79%   785.53MB  1.57%  encoding/json/internal/jsonwire.UnquoteMayCopy
         0     0% 98.79% 35885.12MB 71.67%  encoding/json/v2.Unmarshal
         0     0% 98.79% 16263.03MB 32.48%  encoding/json/v2.makeMapArshaler.func3
         0     0% 98.79%   584.51MB  1.17%  encoding/json/v2.makePointerArshaler.func3
         0     0% 98.79% 35852.09MB 71.60%  encoding/json/v2.makeSliceArshaler.func3
         0     0% 98.79% 35884.12MB 71.67%  encoding/json/v2.makeStructArshaler.func3
         0     0% 98.79%  1367.52MB  2.73%  encoding/json/v2.newAddressableValue
```
</details>

<details><summary>[M] goccy rep 1 allocs profile: 192017 decoded ops in the window</summary>

```
File: goccy-server
Type: alloc_space
Time: 2026-11-16 09:05:25 -03
Duration: 60s, Total samples = 48349.81MB 
Showing nodes accounting for 47694.76MB, 98.65% of 48349.81MB total
Dropped 82 nodes (cum <= 241.75MB)
Showing top 20 nodes out of 35
      flat  flat%   sum%        cum   cum%
14387.15MB 29.76% 29.76% 14387.15MB 29.76%  reflect.growslice
13540.46MB 28.01% 57.76% 13543.96MB 28.01%  github.com/lurodrisilva/golang-json-encoding-benchmark/service/internal/server.readBody
12541.94MB 25.94% 83.70% 12541.94MB 25.94%  reflect.mapassign_faststr0
 3912.07MB  8.09% 91.79%  3912.07MB  8.09%  github.com/goccy/go-json/internal/decoder.(*RuntimeContext).makeString (inline)
 2106.10MB  4.36% 96.15%  2106.10MB  4.36%  reflect.makemap
  542.01MB  1.12% 97.27%   542.01MB  1.12%  github.com/goccy/go-json/internal/decoder.(*RuntimeContext).unescapeString
  311.50MB  0.64% 97.91%   311.50MB  0.64%  reflect.unsafe_New
  282.50MB  0.58% 98.50%   282.50MB  0.58%  github.com/goccy/go-json/internal/decoder.allocatorOf.func18
   49.01MB   0.1% 98.60%   267.06MB  0.55%  net/http.(*conn).readRequest
   19.50MB  0.04% 98.64% 48040.83MB 99.36%  github.com/lurodrisilva/golang-json-encoding-benchmark/service/internal/server.(*ingestHandler).ServeHTTP
    2.50MB 0.0052% 98.65% 48333.80MB   100%  net/http.(*conn).serve
         0     0% 98.65% 34145.21MB 70.62%  github.com/goccy/go-json.Unmarshal (inline)
         0     0% 98.65% 34145.21MB 70.62%  github.com/goccy/go-json.unmarshal
         0     0% 98.65%   453.01MB  0.94%  github.com/goccy/go-json/internal/decoder.(*basicPtrDecoder).Decode
         0     0% 98.65% 15823.06MB 32.73%  github.com/goccy/go-json/internal/decoder.(*mapDecoder).Decode
         0     0% 98.65%   311.50MB  0.64%  github.com/goccy/go-json/internal/decoder.(*ptrDecoder).Decode
         0     0% 98.65% 34067.79MB 70.46%  github.com/goccy/go-json/internal/decoder.(*sliceDecoder).Decode
         0     0% 98.65% 15833.25MB 32.75%  github.com/goccy/go-json/internal/decoder.(*sliceDecoder).decodeLong
         0     0% 98.65% 14387.15MB 29.76%  github.com/goccy/go-json/internal/decoder.(*sliceDecoder).grow
         0     0% 98.65%  4884.32MB 10.10%  github.com/goccy/go-json/internal/decoder.(*sliceDecoder).growBuf (inline)
```
</details>

### Pair D · f1-64k · fixed @3200 rps · profile=cpu

- [M] trials analysed: stdlib-arm 1, goccy-arm 1.
- [M] window: nominal 180.0 s, wall 180.5 s, drain 0.513 s (per trial; the drain is k6 finishing in-flight requests).

| Metric | unit | stdlib (trials) | goccy (trials) | stdlib median | goccy median | Δ abs | reduction | 95% CI | verdict (pre-registered) | post-hoc verdict |
|---|---|---|---|---:|---:|---:|---:|---|---|---|
| HTTP request duration (k6) min | ms | 0.574 | 0.282 | 0.574 | 0.282 | -0.292 | +50.9% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| HTTP request duration (k6) med | ms | 0.618 | 0.337 | 0.618 | 0.337 | -0.281 | +45.5% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| HTTP request duration (k6) avg | ms | 0.782 | 0.408 | 0.782 | 0.408 | -0.375 | +47.9% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| HTTP request duration (k6) p(95) | ms | 1.336 | 0.765 | 1.336 | 0.765 | -0.571 | +42.7% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| HTTP request duration (k6) p(99) | ms | 3.423 | 1.218 | 3.423 | 1.218 | -2.205 | +64.4% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| HTTP request duration (k6) max | ms | 73.0 | 34.2 | 73.0 | 34.2 | -38.8 | +53.2% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Handler duration (server) min | ms | 0.526 | 0.245 | 0.526 | 0.245 | -0.281 | +53.4% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Handler duration (server) med | ms | 0.548 | 0.266 | 0.548 | 0.266 | -0.282 | +51.5% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Handler duration (server) avg | ms | 0.598 | 0.289 | 0.598 | 0.289 | -0.309 | +51.7% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Handler duration (server) p(95) | ms | 0.742 | 0.385 | 0.742 | 0.385 | -0.357 | +48.1% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Handler duration (server) p(99) | ms | 1.336 | 0.645 | 1.336 | 0.645 | -0.691 | +51.7% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Handler duration (server) max | ms | 65.9 | 29.5 | 65.9 | 29.5 | -36.4 | +55.2% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Decode duration (Unmarshal only) min | ms | 0.394 | 0.117 | 0.394 | 0.117 | -0.277 | +70.3% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Decode duration (Unmarshal only) med | ms | 0.413 | 0.129 | 0.413 | 0.129 | -0.284 | +68.8% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Decode duration (Unmarshal only) avg | ms | 0.447 | 0.143 | 0.447 | 0.143 | -0.304 | +68.1% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Decode duration (Unmarshal only) p(95) | ms | 0.553 | 0.197 | 0.553 | 0.197 | -0.356 | +64.4% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Decode duration (Unmarshal only) p(99) | ms | 0.956 | 0.311 | 0.956 | 0.311 | -0.645 | +67.5% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Decode duration (Unmarshal only) max | ms | 65.8 | 27.1 | 65.8 | 27.1 | -38.6 | +58.7% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Body read duration (shared code, control) min | ms | 0.005 | 0.005 | 0.005 | 0.005 | 0 | +0.0% | n/a | control (shared code) | control (shared code) |
| Body read duration (shared code, control) med | ms | 0.010 | 0.010 | 0.010 | 0.010 | 0 | +0.0% | n/a | control (shared code) | control (shared code) |
| Body read duration (shared code, control) avg | ms | 0.019 | 0.016 | 0.019 | 0.016 | -0.00313 | +16.6% | n/a | control (shared code) | control (shared code) |
| Body read duration (shared code, control) p(95) | ms | 0.033 | 0.031 | 0.033 | 0.031 | -0.002 | +6.1% | n/a | control (shared code) | control (shared code) |
| Body read duration (shared code, control) p(99) | ms | 0.116 | 0.087 | 0.116 | 0.087 | -0.029 | +25.0% | n/a | control (shared code) | control (shared code) |
| Body read duration (shared code, control) max | ms | 36.3 | 9.299 | 36.3 | 9.299 | -27.0 | +74.4% | n/a | control (shared code) | control (shared code) |
| Policy scan duration (shared code, control) min | ms | 0.119 | 0.118 | 0.119 | 0.118 | -0.001 | +0.8% | n/a | control (shared code) | control (shared code) |
| Policy scan duration (shared code, control) med | ms | 0.124 | 0.124 | 0.124 | 0.124 | 0 | +0.0% | n/a | control (shared code) | control (shared code) |
| Policy scan duration (shared code, control) avg | ms | 0.132 | 0.130 | 0.132 | 0.130 | -0.00208 | +1.6% | n/a | control (shared code) | control (shared code) |
| Policy scan duration (shared code, control) p(95) | ms | 0.166 | 0.136 | 0.166 | 0.136 | -0.030 | +18.1% | n/a | control (shared code) | control (shared code) |
| Policy scan duration (shared code, control) p(99) | ms | 0.276 | 0.271 | 0.276 | 0.271 | -0.005 | +1.8% | n/a | control (shared code) | control (shared code) |
| Policy scan duration (shared code, control) max | ms | 12.8 | 18.7 | 12.8 | 18.7 | 5.906 | -46.3% | n/a | control (shared code) | control (shared code) |
| Validate duration (shared code, control) min | ms | 0 | 0 | 0 | 0 | 0 | n/a | n/a | control (shared code) | control (shared code) |
| Validate duration (shared code, control) med | ms | 0.001 | 0.001 | 0.001 | 0.001 | 0 | +0.0% | n/a | control (shared code) | control (shared code) |
| Validate duration (shared code, control) avg | ms | 0.000853 | 0.000901 | 0.000853 | 0.000901 | 4.83e-05 | -5.7% | n/a | control (shared code) | control (shared code) |
| Validate duration (shared code, control) p(95) | ms | 0.001 | 0.001 | 0.001 | 0.001 | 0 | +0.0% | n/a | control (shared code) | control (shared code) |
| Validate duration (shared code, control) p(99) | ms | 0.002 | 0.002 | 0.002 | 0.002 | 0 | +0.0% | n/a | control (shared code) | control (shared code) |
| Validate duration (shared code, control) max | ms | 3.260 | 1.040 | 3.260 | 1.040 | -2.220 | +68.1% | n/a | control (shared code) | control (shared code) |
| Achieved throughput | req/s | 3200 | 3200 | 3200 | 3200 | 0 | +0.0% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Server decoded_ok rate | op/s | 3195 | 3195 | 3195 | 3195 | 0.142 | +0.0% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| CPU time per successful op | µs/op | 669.4 | 382.4 | 669.4 | 382.4 | -287.0 | +42.9% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| CPU utilization | cores | 2.139 | 1.222 | 2.139 | 1.222 | -0.917 | +42.9% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| CPU-seconds in window | s | 386.1 | 220.5 | 386.1 | 220.5 | -165.5 | +42.9% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Σ handler durations | s | 345.0 | 166.8 | 345.0 | 166.8 | -178.3 | +51.7% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Σ decode durations | s | 257.5 | 82.3 | 257.5 | 82.3 | -175.3 | +68.1% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Allocated bytes per op | B/op | 274369 | 264577 | 274369 | 264577 | -9792 | +3.6% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Allocated objects per op | obj/op | 2393 | 2172 | 2393 | 2172 | -220.6 | +9.2% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| GC cycles per 1k ops | cycles | 9.006 | 9.603 | 9.006 | 9.603 | 0.597 | -6.6% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| GC pause total in window | s | 0.318 | 0.215 | 0.318 | 0.215 | -0.103 | +32.5% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Live heap at window end | MiB | 17.2 | 19.1 | 17.2 | 19.1 | 1.831 | -10.6% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Runtime-mapped memory at end | MiB | 129.0 | 145.4 | 129.0 | 145.4 | 16.4 | -12.7% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Peak RSS (ps, 5 s samples) | MiB | 134.9 | 150.1 | 134.9 | 150.1 | 15.2 | -11.3% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Max RSS (getrusage) | MiB | 134.9 | 150.1 | 134.9 | 150.1 | 15.2 | -11.3% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |

<details><summary>[M] stdlib rep 1 cpu profile: 192534 decoded ops in the window</summary>

```
File: stdlib-server
Type: cpu
Time: 2026-11-16 17:36:53 -03
Duration: 60.16s, Total samples = 122.99s (204.43%)
Showing nodes accounting for 106.62s, 86.69% of 122.99s total
Dropped 418 nodes (cum <= 0.61s)
Showing top 20 nodes out of 126
      flat  flat%   sum%        cum   cum%
    84.07s 68.36% 68.36%     84.07s 68.36%  syscall.rawsyscalln
     3.75s  3.05% 71.40%      3.75s  3.05%  runtime.pthread_kill
     3.64s  2.96% 74.36%      3.64s  2.96%  runtime.kevent
     2.15s  1.75% 76.11%      2.15s  1.75%  encoding/json/internal/jsonwire.ConsumeSimpleString (inline)
     1.70s  1.38% 77.49%      1.70s  1.38%  runtime.usleep
     1.36s  1.11% 78.60%      1.37s  1.11%  encoding/json/internal/jsonwire.ConsumeWhitespace (inline)
     1.26s  1.02% 79.62%      5.21s  4.24%  encoding/json/jsontext.(*decoderState).ReadValue
     1.05s  0.85% 80.48%      1.05s  0.85%  runtime.pthread_cond_wait
     1.04s  0.85% 81.32%     20.42s 16.60%  encoding/json/v2.makeStructArshaler.func3
     1.02s  0.83% 82.15%      5.48s  4.46%  encoding/json/jsontext.(*decoderState).consumeValue
     0.96s  0.78% 82.93%      1.05s  0.85%  github.com/lurodrisilva/golang-json-encoding-benchmark/service/internal/policy.(*scanner).str
     0.79s  0.64% 83.58%      1.69s  1.37%  encoding/json/jsontext.(*decoderState).PeekKind
     0.65s  0.53% 84.10%      0.88s  0.72%  encoding/json/jsontext.(*decoderState).ReadToken
     0.54s  0.44% 84.54%      0.86s   0.7%  runtime.pcvalue
     0.53s  0.43% 84.97%      1.66s  1.35%  runtime.mapaccess2_faststr
     0.49s   0.4% 85.37%      1.31s  1.07%  encoding/json/v2.makeString
     0.45s  0.37% 85.74%      0.78s  0.63%  internal/runtime/maps.(*Map).getWithoutKeySmallFastStr
     0.43s  0.35% 86.09%      0.73s  0.59%  encoding/json/internal/jsonwire.ConsumeStringResumable
     0.41s  0.33% 86.42%      2.06s  1.67%  github.com/lurodrisilva/golang-json-encoding-benchmark/service/internal/policy.(*scanner).object
     0.33s  0.27% 86.69%      2.07s  1.68%  encoding/json/jsontext.(*decoderState).consumeObject
```
</details>

<details><summary>[M] goccy rep 1 cpu profile: 192487 decoded ops in the window</summary>

```
File: goccy-server
Type: cpu
Time: 2026-11-16 17:42:54 -03
Duration: 60.15s, Total samples = 69.67s (115.83%)
Showing nodes accounting for 62.48s, 89.68% of 69.67s total
Dropped 366 nodes (cum <= 0.35s)
Showing top 20 nodes out of 121
      flat  flat%   sum%        cum   cum%
    47.16s 67.69% 67.69%     47.17s 67.70%  syscall.rawsyscalln
     5.66s  8.12% 75.81%      5.66s  8.12%  runtime.kevent
     2.18s  3.13% 78.94%      2.18s  3.13%  runtime.pthread_kill
     1.25s  1.79% 80.74%      1.25s  1.79%  runtime.pthread_cond_wait
     1.25s  1.79% 82.53%      1.25s  1.79%  runtime.usleep
     0.67s  0.96% 83.49%      0.79s  1.13%  github.com/lurodrisilva/golang-json-encoding-benchmark/service/internal/policy.(*scanner).str
     0.62s  0.89% 84.38%      0.62s  0.89%  runtime.pthread_cond_signal
     0.53s  0.76% 85.14%      0.66s  0.95%  github.com/goccy/go-json/internal/decoder.(*structDecoder).decodeKey
     0.50s  0.72% 85.86%      1.96s  2.81%  github.com/goccy/go-json/internal/decoder.(*stringDecoder).decodeStringField
     0.37s  0.53% 86.39%      0.45s  0.65%  runtime.tryDeferToSpanScan
     0.35s   0.5% 86.90%      0.35s   0.5%  github.com/goccy/go-json/internal/decoder.skipWhiteSpace (inline)
     0.31s  0.44% 87.34%      0.68s  0.98%  runtime.pcvalue
     0.30s  0.43% 87.77%      0.57s  0.82%  runtime.mallocgcTinySC2
     0.27s  0.39% 88.16%      0.37s  0.53%  github.com/goccy/go-json/internal/jsonnum.ParseFloatTerminated
     0.25s  0.36% 88.52%      1.54s  2.21%  github.com/lurodrisilva/golang-json-encoding-benchmark/service/internal/policy.(*scanner).object
     0.21s   0.3% 88.82%      6.60s  9.47%  github.com/goccy/go-json/internal/decoder.(*structDecoder).Decode
     0.19s  0.27% 89.09%      0.41s  0.59%  runtime.scanObjectsSmall
     0.15s  0.22% 89.31%      1.05s  1.51%  runtime.slicebytetostring
     0.14s   0.2% 89.51%      1.88s  2.70%  runtime.mallocgc
     0.12s  0.17% 89.68%      0.49s   0.7%  runtime.(*stkframe).getStackMap
```
</details>

### Pair D · f2-256k · fixed @1000 rps

- [M] trials analysed: stdlib-arm 5, goccy-arm 5.
- [M] stdlib rep 5: invalid [dropped_iterations=13115 achieved 0.9563 of offered (< 0.99) non-202 fraction 0.00119 > 0.001 transport errors/timeouts=340 host busy 80.1% before trial (top: Comet Helper (Aperitif Renderer) 144%, Comet Helper (Aperitif Renderer) 86%, MTLCompilerService 67%, Comet 65%, Comet Helper (Aperitif Renderer) 45%)]
- [M] stdlib rep 6: INCLUDED but flagged [host busy 37.0% before trial (top: duetexpertd 83%, mobileassetd 59%, mds 46%, backupd-helper 45%, deleted 14%)]
- [M] window: nominal 300.0 s, wall 300.3 s, drain 0.262 s (per trial; the drain is k6 finishing in-flight requests).

| Metric | unit | stdlib (trials) | goccy (trials) | stdlib median | goccy median | Δ abs | reduction | 95% CI | verdict (pre-registered) | post-hoc verdict |
|---|---|---|---|---:|---:|---:|---:|---|---|---|
| HTTP request duration (k6) min | ms | 2.174, 2.163, 2.159, 2.166, 2.159 | 1.029, 1.044, 1.053, 1.043, 1.045 | 2.163 | 1.044 | -1.119 | +51.7% | [+51.3, +52.4] | inconclusive (shared-code control differs ≥ threshold) (was: goccy better (CI excludes 0, ≥ threshold); controls: Body read duration (shared code, control) med -15.4%) | inconclusive (control differs beyond resolution) |
| HTTP request duration (k6) med | ms | 2.297, 2.293, 2.290, 2.293, 2.290 | 1.179, 1.181, 1.181, 1.181, 1.176 | 2.293 | 1.181 | -1.112 | +48.5% | [+48.4, +48.7] | inconclusive (shared-code control differs ≥ threshold) (was: goccy better (CI excludes 0, ≥ threshold); controls: Body read duration (shared code, control) med -15.4%) | inconclusive (control differs beyond resolution) |
| HTTP request duration (k6) avg | ms | 2.442, 2.453, 2.444, 2.446, 2.423 | 1.278, 1.233, 1.234, 1.234, 1.247 | 2.444 | 1.234 | -1.210 | +49.5% | [+47.7, +49.7] | inconclusive (shared-code control differs ≥ threshold) (was: goccy better (CI excludes 0, ≥ threshold); controls: Body read duration (shared code, control) med -15.4%) | inconclusive (control differs beyond resolution) |
| HTTP request duration (k6) p(95) | ms | 2.775, 2.788, 2.774, 2.779, 2.745 | 1.449, 1.445, 1.447, 1.446, 1.481 | 2.775 | 1.447 | -1.328 | +47.9% | [+46.6, +48.1] | inconclusive (shared-code control differs ≥ threshold) (was: goccy better (CI excludes 0, ≥ threshold); controls: Body read duration (shared code, control) med -15.4%) | inconclusive (control differs beyond resolution) |
| HTTP request duration (k6) p(99) | ms | 5.216, 5.496, 5.419, 5.452, 4.987 | 2.283, 2.086, 2.130, 2.077, 2.405 | 5.419 | 2.130 | -3.289 | +60.7% | [+54.2, +61.9] | inconclusive (shared-code control differs ≥ threshold) (was: goccy better (CI excludes 0, ≥ threshold); controls: Body read duration (shared code, control) med -15.4%) | inconclusive (control differs beyond resolution) |
| HTTP request duration (k6) max | ms | 33.8, 97.0, 39.9, 28.0, 22.8 | 203.8, 13.6, 14.3, 44.0, 46.2 | 33.8 | 44.0 | 10.2 | -30.3% | [-503.4, +66.0] | inconclusive (trial CV > 10%) | inconclusive (trial CV > 10%) |
| Handler duration (server) min | ms | 2.106, 2.100, 2.092, 2.100, 2.095 | 0.981, 0.987, 0.981, 0.986, 0.982 | 2.100 | 0.982 | -1.118 | +53.2% | [+52.9, +53.4] | inconclusive (shared-code control differs ≥ threshold) (was: goccy better (CI excludes 0, ≥ threshold); controls: Body read duration (shared code, control) med -15.4%) | inconclusive (control differs beyond resolution) |
| Handler duration (server) med | ms | 2.200, 2.193, 2.191, 2.194, 2.189 | 1.063, 1.064, 1.063, 1.064, 1.065 | 2.193 | 1.064 | -1.129 | +51.5% | [+51.4, +51.6] | inconclusive (shared-code control differs ≥ threshold) (was: goccy better (CI excludes 0, ≥ threshold); controls: Body read duration (shared code, control) med -15.4%) | inconclusive (control differs beyond resolution) |
| Handler duration (server) avg | ms | 2.322, 2.322, 2.319, 2.321, 2.308 | 1.134, 1.116, 1.116, 1.116, 1.126 | 2.321 | 1.116 | -1.205 | +51.9% | [+51.1, +52.0] | inconclusive (shared-code control differs ≥ threshold) (was: goccy better (CI excludes 0, ≥ threshold); controls: Body read duration (shared code, control) med -15.4%) | inconclusive (control differs beyond resolution) |
| Handler duration (server) p(95) | ms | 2.643, 2.646, 2.640, 2.641, 2.623 | 1.343, 1.339, 1.340, 1.340, 1.362 | 2.641 | 1.340 | -1.301 | +49.3% | [+48.4, +49.4] | inconclusive (shared-code control differs ≥ threshold) (was: goccy better (CI excludes 0, ≥ threshold); controls: Body read duration (shared code, control) med -15.4%) | inconclusive (control differs beyond resolution) |
| Handler duration (server) p(99) | ms | 4.319, 4.434, 4.457, 4.463, 4.236 | 1.974, 1.778, 1.808, 1.772, 2.039 | 4.434 | 1.808 | -2.626 | +59.2% | [+53.4, +60.2] | inconclusive (shared-code control differs ≥ threshold) (was: goccy better (CI excludes 0, ≥ threshold); controls: Body read duration (shared code, control) med -15.4%) | inconclusive (control differs beyond resolution) |
| Handler duration (server) max | ms | 32.6, 94.1, 30.6, 19.2, 21.6 | 154.3, 13.0, 10.7, 33.9, 43.0 | 30.6 | 33.9 | 3.323 | -10.9% | [-403.9, +67.1] | inconclusive (trial CV > 10%) | inconclusive (trial CV > 10%) |
| Decode duration (Unmarshal only) min | ms | 1.576, 1.573, 1.573, 1.570, 1.572 | 0.467, 0.466, 0.465, 0.469, 0.466 | 1.573 | 0.466 | -1.107 | +70.4% | [+70.2, +70.4] | inconclusive (shared-code control differs ≥ threshold) (was: goccy better (CI excludes 0, ≥ threshold); controls: Body read duration (shared code, control) med -15.4%) | inconclusive (control differs beyond resolution) |
| Decode duration (Unmarshal only) med | ms | 1.632, 1.625, 1.625, 1.626, 1.622 | 0.498, 0.498, 0.497, 0.500, 0.499 | 1.625 | 0.498 | -1.127 | +69.4% | [+69.2, +69.5] | inconclusive (shared-code control differs ≥ threshold) (was: goccy better (CI excludes 0, ≥ threshold); controls: Body read duration (shared code, control) med -15.4%) | inconclusive (control differs beyond resolution) |
| Decode duration (Unmarshal only) avg | ms | 1.734, 1.732, 1.731, 1.732, 1.722 | 0.548, 0.537, 0.536, 0.539, 0.543 | 1.732 | 0.539 | -1.193 | +68.9% | [+68.4, +69.0] | inconclusive (shared-code control differs ≥ threshold) (was: goccy better (CI excludes 0, ≥ threshold); controls: Body read duration (shared code, control) med -15.4%) | inconclusive (control differs beyond resolution) |
| Decode duration (Unmarshal only) p(95) | ms | 2.047, 2.044, 2.044, 2.044, 2.030 | 0.760, 0.758, 0.757, 0.759, 0.768 | 2.044 | 0.759 | -1.285 | +62.9% | [+62.4, +63.0] | inconclusive (shared-code control differs ≥ threshold) (was: goccy better (CI excludes 0, ≥ threshold); controls: Body read duration (shared code, control) med -15.4%) | inconclusive (control differs beyond resolution) |
| Decode duration (Unmarshal only) p(99) | ms | 3.197, 3.277, 3.300, 3.291, 3.126 | 1.036, 0.969, 0.985, 0.959, 1.053 | 3.277 | 0.985 | -2.292 | +69.9% | [+67.1, +70.7] | inconclusive (shared-code control differs ≥ threshold) (was: goccy better (CI excludes 0, ≥ threshold); controls: Body read duration (shared code, control) med -15.4%) | inconclusive (control differs beyond resolution) |
| Decode duration (Unmarshal only) max | ms | 25.3, 93.5, 19.1, 17.8, 20.6 | 150.1, 11.8, 8.412, 33.3, 42.3 | 20.6 | 33.3 | 12.7 | -61.6% | [-627.9, +66.8] | inconclusive (trial CV > 10%) | inconclusive (trial CV > 10%) |
| Body read duration (shared code, control) min | ms | 0.015, 0.015, 0.014, 0.015, 0.015 | 0.015, 0.016, 0.015, 0.015, 0.015 | 0.015 | 0.015 | 0 | +0.0% | [-7.1, +0.0] | control (shared code) | control (shared code) |
| Body read duration (shared code, control) med | ms | 0.052, 0.052, 0.052, 0.051, 0.052 | 0.060, 0.061, 0.060, 0.060, 0.061 | 0.052 | 0.060 | 0.008 | -15.4% | [-17.6, -15.4] | control (shared code) (control differs ≥ threshold) | control (shared code) |
| Body read duration (shared code, control) avg | ms | 0.061, 0.062, 0.061, 0.061, 0.060 | 0.070, 0.067, 0.067, 0.066, 0.067 | 0.061 | 0.067 | 0.00598 | -9.8% | [-14.9, -8.6] | control (shared code) | control (shared code) |
| Body read duration (shared code, control) p(95) | ms | 0.112, 0.113, 0.110, 0.112, 0.110 | 0.115, 0.117, 0.118, 0.116, 0.110 | 0.112 | 0.116 | 0.004 | -3.6% | [-6.4, +1.8] | control (shared code) | control (shared code) |
| Body read duration (shared code, control) p(99) | ms | 0.235, 0.239, 0.234, 0.237, 0.220 | 0.196, 0.206, 0.206, 0.203, 0.208 | 0.235 | 0.206 | -0.029 | +12.3% | [+6.4, +16.6] | control (shared code) | control (shared code) |
| Body read duration (shared code, control) max | ms | 17.7, 40.7, 28.3, 11.3, 7.822 | 90.8, 2.289, 4.359, 13.8, 8.807 | 17.7 | 8.807 | -8.931 | +50.3% | [-412.0, +89.3] | control (shared code) | control (shared code) |
| Policy scan duration (shared code, control) min | ms | 0.484, 0.485, 0.486, 0.485, 0.485 | 0.477, 0.479, 0.479, 0.478, 0.477 | 0.485 | 0.478 | -0.007 | +1.4% | [+1.2, +1.6] | control (shared code) | control (shared code) |
| Policy scan duration (shared code, control) med | ms | 0.502, 0.503, 0.502, 0.503, 0.502 | 0.496, 0.496, 0.497, 0.496, 0.496 | 0.502 | 0.496 | -0.006 | +1.2% | [+1.0, +1.4] | control (shared code) | control (shared code) |
| Policy scan duration (shared code, control) avg | ms | 0.524, 0.526, 0.525, 0.525, 0.523 | 0.513, 0.509, 0.509, 0.508, 0.512 | 0.525 | 0.509 | -0.016 | +3.0% | [+2.1, +3.2] | control (shared code) | control (shared code) |
| Policy scan duration (shared code, control) p(95) | ms | 0.595, 0.596, 0.595, 0.595, 0.594 | 0.567, 0.563, 0.565, 0.563, 0.573 | 0.595 | 0.565 | -0.030 | +5.0% | [+3.7, +5.4] | control (shared code) | control (shared code) |
| Policy scan duration (shared code, control) p(99) | ms | 1.011, 1.017, 1.016, 1.018, 0.998 | 0.982, 0.900, 0.929, 0.880, 0.986 | 1.016 | 0.929 | -0.087 | +8.6% | [+2.5, +13.4] | control (shared code) | control (shared code) |
| Policy scan duration (shared code, control) max | ms | 9.151, 17.7, 8.870, 10.2, 9.152 | 50.9, 4.563, 4.960, 6.547, 14.1 | 9.152 | 6.547 | -2.605 | +28.5% | [-456.0, +63.0] | control (shared code) | control (shared code) |
| Validate duration (shared code, control) min | ms | 0.002, 0.002, 0.002, 0.002, 0.002 | 0.002, 0.002, 0.002, 0.002, 0.002 | 0.002 | 0.002 | 0 | +0.0% | [+0.0, +0.0] | control (shared code) | control (shared code) |
| Validate duration (shared code, control) med | ms | 0.002, 0.002, 0.002, 0.002, 0.002 | 0.002, 0.002, 0.002, 0.002, 0.003 | 0.002 | 0.002 | 0 | +0.0% | [-50.0, +0.0] | control (shared code) | control (shared code) |
| Validate duration (shared code, control) avg | ms | 0.00272, 0.00271, 0.00272, 0.00274, 0.0023 | 0.00264, 0.0026, 0.00262, 0.00263, 0.00287 | 0.00272 | 0.00263 | -9.17e-05 | +3.4% | [-14.3, +4.3] | control (shared code) | control (shared code) |
| Validate duration (shared code, control) p(95) | ms | 0.004, 0.004, 0.004, 0.004, 0.003 | 0.004, 0.004, 0.004, 0.004, 0.005 | 0.004 | 0.004 | 0 | +0.0% | [-33.3, +0.0] | control (shared code) | control (shared code) |
| Validate duration (shared code, control) p(99) | ms | 0.007, 0.007, 0.007, 0.007, 0.006 | 0.006, 0.006, 0.006, 0.006, 0.007 | 0.007 | 0.006 | -0.001 | +14.3% | [+0.0, +14.3] | control (shared code) | control (shared code) |
| Validate duration (shared code, control) max | ms | 3.754, 1.091, 2.861, 3.335, 2.692 | 1.115, 0.630, 0.674, 0.537, 2.319 | 2.861 | 0.674 | -2.187 | +76.4% | [+13.9, +83.2] | control (shared code) | control (shared code) |
| Achieved throughput | req/s | 1000, 1000, 1000, 1000, 1000 | 1000, 1000, 1000, 1000, 1000 | 1000 | 1000 | 0.00333 | +0.0% | [+0.0, +0.0] | no detectable difference | no detectable difference |
| Server decoded_ok rate | op/s | 999.5, 999.4, 999.5, 999.5, 999.5 | 999.5, 999.5, 999.5, 999.5, 999.5 | 999.5 | 999.5 | 0.033 | +0.0% | [+0.0, +0.0] | inconclusive (shared-code control differs ≥ threshold) (was: detectable but below threshold; controls: Body read duration (shared code, control) med -15.4%) | inconclusive (control differs beyond resolution) |
| CPU time per successful op | µs/op | 2497, 2492, 2492, 2494, 2480 | 1351, 1349, 1349, 1350, 1362 | 2492 | 1350 | -1142 | +45.8% | [+45.4, +45.9] | inconclusive (shared-code control differs ≥ threshold) (was: goccy better (CI excludes 0, ≥ threshold); controls: Body read duration (shared code, control) med -15.4%) | inconclusive (control differs beyond resolution) |
| CPU utilization | cores | 2.495, 2.491, 2.491, 2.493, 2.479 | 1.350, 1.348, 1.348, 1.349, 1.361 | 2.491 | 1.349 | -1.142 | +45.8% | [+45.4, +45.9] | inconclusive (shared-code control differs ≥ threshold) (was: goccy better (CI excludes 0, ≥ threshold); controls: Body read duration (shared code, control) med -15.4%) | inconclusive (control differs beyond resolution) |
| CPU-seconds in window | s | 749.3, 748.0, 747.9, 748.4, 744.4 | 405.5, 404.9, 404.8, 405.1, 408.7 | 748.0 | 405.1 | -342.8 | +45.8% | [+45.4, +45.9] | inconclusive (shared-code control differs ≥ threshold) (was: goccy better (CI excludes 0, ≥ threshold); controls: Body read duration (shared code, control) med -15.4%) | inconclusive (control differs beyond resolution) |
| Σ handler durations | s | 696.9, 696.9, 696.0, 696.5, 692.6 | 340.3, 334.8, 334.8, 335.0, 337.8 | 696.5 | 335.0 | -361.6 | +51.9% | [+51.1, +52.0] | inconclusive (shared-code control differs ≥ threshold) (was: goccy better (CI excludes 0, ≥ threshold); controls: Body read duration (shared code, control) med -15.4%) | inconclusive (control differs beyond resolution) |
| Σ decode durations | s | 520.4, 519.8, 519.3, 519.7, 516.8 | 164.5, 161.1, 161.0, 161.7, 163.1 | 519.7 | 161.7 | -358.0 | +68.9% | [+68.4, +69.0] | inconclusive (shared-code control differs ≥ threshold) (was: goccy better (CI excludes 0, ≥ threshold); controls: Body read duration (shared code, control) med -15.4%) | inconclusive (control differs beyond resolution) |
| Allocated bytes per op | B/op | 1052752, 1052757, 1052776, 1052762, 1052761 | 1054018, 1053662, 1053568, 1053909, 1053222 | 1052761 | 1053662 | 901.2 | -0.1% | [-0.1, -0.0] | inconclusive (shared-code control differs ≥ threshold) (was: detectable but below threshold; controls: Body read duration (shared code, control) med -15.4%) | inconclusive (control differs beyond resolution) |
| Allocated objects per op | obj/op | 9491, 9491, 9491, 9491, 9491 | 8483, 8483, 8483, 8483, 8483 | 9491 | 8483 | -1008 | +10.6% | [+10.6, +10.6] | inconclusive (shared-code control differs ≥ threshold) (was: goccy better (CI excludes 0, ≥ threshold); controls: Body read duration (shared code, control) med -15.4%) | inconclusive (control differs beyond resolution) |
| GC cycles per 1k ops | cycles | 121.6, 122.1, 122.1, 121.7, 121.9 | 125.7, 125.7, 125.9, 126.0, 125.6 | 121.9 | 125.7 | 3.788 | -3.1% | [-3.5, -2.9] | inconclusive (shared-code control differs ≥ threshold) (was: detectable but below threshold; controls: Body read duration (shared code, control) med -15.4%) | inconclusive (control differs beyond resolution) |
| GC pause total in window | s | 1.590, 1.600, 1.605, 1.606, 1.527 | 1.023, 1.014, 1.032, 1.034, 1.101 | 1.600 | 1.032 | -0.568 | +35.5% | [+31.2, +36.6] | inconclusive (shared-code control differs ≥ threshold) (was: goccy better (CI excludes 0, ≥ threshold); controls: Body read duration (shared code, control) med -15.4%) | inconclusive (control differs beyond resolution) |
| Live heap at window end | MiB | 6.746, 6.817, 6.621, 6.790, 7.232 | 7.627, 6.852, 7.528, 7.861, 7.424 | 6.790 | 7.528 | 0.739 | -10.9% | [-15.8, -0.9] | inconclusive (shared-code control differs ≥ threshold) (was: stdlib better (CI excludes 0, ≥ threshold); controls: Body read duration (shared code, control) med -15.4%) | inconclusive (control differs beyond resolution) |
| Runtime-mapped memory at end | MiB | 50.3, 58.3, 50.2, 50.0, 54.0 | 62.3, 50.5, 54.5, 58.1, 54.3 | 50.3 | 54.5 | 4.250 | -8.5% | [-24.0, +6.6] | no detectable difference | no detectable difference |
| Peak RSS (ps, 5 s samples) | MiB | 57.7, 63.5, 57.8, 57.4, 57.8 | 69.8, 58.5, 59.1, 61.4, 61.8 | 57.8 | 61.4 | 3.641 | -6.3% | [-20.8, +3.4] | no detectable difference | no detectable difference |
| Max RSS (getrusage) | MiB | 57.7, 63.5, 57.8, 57.4, 57.8 | 69.8, 58.5, 59.1, 61.4, 61.8 | 57.8 | 61.4 | 3.641 | -6.3% | [-20.8, +3.4] | no detectable difference | no detectable difference |

### Pair D · f2-256k · fixed @625 rps

- [M] trials analysed: stdlib-arm 5, goccy-arm 5.
- [M] stdlib rep 4: INCLUDED but flagged [host busy 16.7% before trial (top: corespotlightd 114%, WindowServer 10%)]
- [M] window: nominal 300.0 s, wall 300.2 s, drain 0.170 s (per trial; the drain is k6 finishing in-flight requests).

| Metric | unit | stdlib (trials) | goccy (trials) | stdlib median | goccy median | Δ abs | reduction | 95% CI | verdict (pre-registered) | post-hoc verdict |
|---|---|---|---|---:|---:|---:|---:|---|---|---|
| HTTP request duration (k6) min | ms | 2.145, 2.144, 2.148, 2.143, 2.152 | 1.023, 1.033, 1.027, 1.047, 1.056 | 2.145 | 1.033 | -1.112 | +51.8% | [+50.8, +52.3] | goccy better (CI excludes 0, ≥ threshold) | goccy better (CI excludes 0, ≥ threshold) |
| HTTP request duration (k6) med | ms | 2.263, 2.261, 2.260, 2.260, 2.260 | 1.148, 1.148, 1.142, 1.150, 1.150 | 2.260 | 1.148 | -1.112 | +49.2% | [+49.1, +49.5] | goccy better (CI excludes 0, ≥ threshold) | goccy better (CI excludes 0, ≥ threshold) |
| HTTP request duration (k6) avg | ms | 2.339, 2.344, 2.341, 2.339, 2.335 | 1.193, 1.214, 1.193, 1.195, 1.197 | 2.339 | 1.195 | -1.144 | +48.9% | [+48.1, +49.1] | goccy better (CI excludes 0, ≥ threshold) | goccy better (CI excludes 0, ≥ threshold) |
| HTTP request duration (k6) p(95) | ms | 2.559, 2.557, 2.561, 2.555, 2.558 | 1.379, 1.384, 1.382, 1.381, 1.379 | 2.558 | 1.381 | -1.177 | +46.0% | [+45.9, +46.1] | goccy better (CI excludes 0, ≥ threshold) | goccy better (CI excludes 0, ≥ threshold) |
| HTTP request duration (k6) p(99) | ms | 2.910, 2.986, 2.942, 2.990, 2.884 | 1.581, 1.821, 1.623, 1.634, 1.589 | 2.942 | 1.623 | -1.319 | +44.8% | [+38.1, +46.9] | goccy better (CI excludes 0, ≥ threshold) | goccy better (CI excludes 0, ≥ threshold) |
| HTTP request duration (k6) max | ms | 13.3, 66.8, 20.8, 31.8, 16.4 | 13.6, 92.6, 16.0, 11.0, 13.4 | 20.8 | 13.6 | -7.176 | +34.5% | [-345.2, +79.6] | inconclusive (trial CV > 10%) | inconclusive (trial CV > 10%) |
| Handler duration (server) min | ms | 2.077, 2.074, 2.077, 2.075, 2.068 | 0.943, 0.959, 0.959, 0.958, 0.955 | 2.075 | 0.958 | -1.117 | +53.8% | [+53.7, +54.6] | goccy better (CI excludes 0, ≥ threshold) | goccy better (CI excludes 0, ≥ threshold) |
| Handler duration (server) med | ms | 2.161, 2.160, 2.162, 2.159, 2.159 | 1.046, 1.050, 1.052, 1.049, 1.049 | 2.160 | 1.049 | -1.111 | +51.4% | [+51.3, +51.6] | goccy better (CI excludes 0, ≥ threshold) | goccy better (CI excludes 0, ≥ threshold) |
| Handler duration (server) avg | ms | 2.235, 2.236, 2.237, 2.235, 2.232 | 1.088, 1.103, 1.096, 1.093, 1.092 | 2.235 | 1.093 | -1.142 | +51.1% | [+50.7, +51.3] | goccy better (CI excludes 0, ≥ threshold) | goccy better (CI excludes 0, ≥ threshold) |
| Handler duration (server) p(95) | ms | 2.447, 2.444, 2.448, 2.443, 2.446 | 1.291, 1.299, 1.298, 1.294, 1.293 | 2.446 | 1.294 | -1.152 | +47.1% | [+46.9, +47.2] | goccy better (CI excludes 0, ≥ threshold) | goccy better (CI excludes 0, ≥ threshold) |
| Handler duration (server) p(99) | ms | 2.706, 2.741, 2.733, 2.740, 2.701 | 1.422, 1.508, 1.440, 1.444, 1.431 | 2.733 | 1.440 | -1.293 | +47.3% | [+44.8, +48.0] | goccy better (CI excludes 0, ≥ threshold) | goccy better (CI excludes 0, ≥ threshold) |
| Handler duration (server) max | ms | 12.8, 54.4, 19.7, 24.0, 8.133 | 5.633, 92.0, 12.2, 6.958, 6.838 | 19.7 | 6.958 | -12.7 | +64.6% | [-368.1, +87.2] | inconclusive (trial CV > 10%) | inconclusive (trial CV > 10%) |
| Decode duration (Unmarshal only) min | ms | 1.548, 1.548, 1.547, 1.549, 1.548 | 0.443, 0.446, 0.444, 0.447, 0.446 | 1.548 | 0.446 | -1.102 | +71.2% | [+71.1, +71.4] | goccy better (CI excludes 0, ≥ threshold) | goccy better (CI excludes 0, ≥ threshold) |
| Decode duration (Unmarshal only) med | ms | 1.598, 1.596, 1.596, 1.596, 1.595 | 0.491, 0.493, 0.491, 0.491, 0.492 | 1.596 | 0.491 | -1.105 | +69.2% | [+69.1, +69.3] | goccy better (CI excludes 0, ≥ threshold) | goccy better (CI excludes 0, ≥ threshold) |
| Decode duration (Unmarshal only) avg | ms | 1.659, 1.659, 1.659, 1.659, 1.656 | 0.525, 0.534, 0.527, 0.526, 0.527 | 1.659 | 0.527 | -1.132 | +68.3% | [+67.8, +68.4] | goccy better (CI excludes 0, ≥ threshold) | goccy better (CI excludes 0, ≥ threshold) |
| Decode duration (Unmarshal only) p(95) | ms | 1.873, 1.868, 1.870, 1.869, 1.870 | 0.731, 0.738, 0.733, 0.730, 0.730 | 1.870 | 0.731 | -1.139 | +60.9% | [+60.5, +61.0] | goccy better (CI excludes 0, ≥ threshold) | goccy better (CI excludes 0, ≥ threshold) |
| Decode duration (Unmarshal only) p(99) | ms | 2.063, 2.066, 2.072, 2.072, 2.060 | 0.824, 0.845, 0.830, 0.831, 0.830 | 2.066 | 0.830 | -1.236 | +59.8% | [+59.1, +60.1] | goccy better (CI excludes 0, ≥ threshold) | goccy better (CI excludes 0, ≥ threshold) |
| Decode duration (Unmarshal only) max | ms | 11.5, 52.0, 14.9, 23.3, 7.018 | 3.979, 91.4, 11.3, 5.621, 5.780 | 14.9 | 5.780 | -9.085 | +61.1% | [-514.7, +88.9] | inconclusive (trial CV > 10%) | inconclusive (trial CV > 10%) |
| Body read duration (shared code, control) min | ms | 0.015, 0.015, 0.015, 0.015, 0.015 | 0.017, 0.016, 0.016, 0.017, 0.016 | 0.015 | 0.016 | 0.001 | -6.7% | [-13.3, -6.7] | control (shared code) | control (shared code) |
| Body read duration (shared code, control) med | ms | 0.057, 0.057, 0.059, 0.057, 0.057 | 0.054, 0.054, 0.057, 0.054, 0.053 | 0.057 | 0.054 | -0.003 | +5.3% | [+0.0, +8.5] | control (shared code) | control (shared code) |
| Body read duration (shared code, control) avg | ms | 0.059, 0.061, 0.062, 0.059, 0.060 | 0.057, 0.059, 0.061, 0.058, 0.057 | 0.060 | 0.058 | -0.00163 | +2.7% | [-1.1, +6.2] | control (shared code) | control (shared code) |
| Body read duration (shared code, control) p(95) | ms | 0.086, 0.088, 0.103, 0.086, 0.087 | 0.075, 0.080, 0.087, 0.080, 0.078 | 0.087 | 0.080 | -0.007 | +8.0% | [+0.0, +22.3] | control (shared code) | control (shared code) |
| Body read duration (shared code, control) p(99) | ms | 0.158, 0.170, 0.153, 0.163, 0.166 | 0.098, 0.132, 0.145, 0.125, 0.124 | 0.163 | 0.125 | -0.038 | +23.3% | [+11.0, +39.9] | control (shared code) | control (shared code) |
| Body read duration (shared code, control) max | ms | 6.404, 37.7, 6.668, 14.2, 5.057 | 3.900, 27.6, 3.449, 2.936, 3.964 | 6.668 | 3.900 | -2.768 | +41.5% | [-313.4, +89.7] | control (shared code) | control (shared code) |
| Policy scan duration (shared code, control) min | ms | 0.486, 0.484, 0.486, 0.484, 0.483 | 0.455, 0.459, 0.463, 0.464, 0.461 | 0.484 | 0.461 | -0.023 | +4.8% | [+4.1, +6.0] | control (shared code) | control (shared code) |
| Policy scan duration (shared code, control) med | ms | 0.499, 0.499, 0.499, 0.500, 0.500 | 0.495, 0.497, 0.498, 0.497, 0.497 | 0.499 | 0.497 | -0.002 | +0.4% | [+0.2, +0.8] | control (shared code) | control (shared code) |
| Policy scan duration (shared code, control) avg | ms | 0.514, 0.514, 0.514, 0.514, 0.514 | 0.503, 0.507, 0.506, 0.506, 0.506 | 0.514 | 0.506 | -0.00788 | +1.5% | [+1.3, +2.1] | control (shared code) | control (shared code) |
| Policy scan duration (shared code, control) p(95) | ms | 0.593, 0.593, 0.592, 0.591, 0.590 | 0.555, 0.568, 0.561, 0.565, 0.562 | 0.592 | 0.562 | -0.030 | +5.1% | [+4.1, +6.2] | control (shared code) | control (shared code) |
| Policy scan duration (shared code, control) p(99) | ms | 0.777, 0.774, 0.777, 0.779, 0.784 | 0.596, 0.604, 0.596, 0.597, 0.595 | 0.777 | 0.596 | -0.181 | +23.3% | [+22.3, +24.0] | control (shared code) | control (shared code) |
| Policy scan duration (shared code, control) max | ms | 4.668, 12.0, 7.465, 11.1, 3.101 | 3.690, 17.2, 3.823, 3.735, 4.016 | 7.465 | 3.823 | -3.642 | +48.8% | [-130.9, +68.1] | control (shared code) | control (shared code) |
| Validate duration (shared code, control) min | ms | 0.002, 0.002, 0.002, 0.002, 0.002 | 0.002, 0.002, 0.002, 0.002, 0.002 | 0.002 | 0.002 | 0 | +0.0% | [+0.0, +0.0] | control (shared code) | control (shared code) |
| Validate duration (shared code, control) med | ms | 0.002, 0.002, 0.002, 0.002, 0.002 | 0.002, 0.002, 0.002, 0.002, 0.002 | 0.002 | 0.002 | 0 | +0.0% | [+0.0, +0.0] | control (shared code) | control (shared code) |
| Validate duration (shared code, control) avg | ms | 0.00222, 0.00222, 0.00219, 0.00223, 0.00221 | 0.00221, 0.00224, 0.00217, 0.00221, 0.0022 | 0.00222 | 0.00221 | -1.25e-05 | +0.6% | [-0.8, +2.4] | control (shared code) | control (shared code) |
| Validate duration (shared code, control) p(95) | ms | 0.003, 0.003, 0.003, 0.003, 0.003 | 0.003, 0.003, 0.003, 0.003, 0.003 | 0.003 | 0.003 | 0 | +0.0% | [+0.0, +0.0] | control (shared code) | control (shared code) |
| Validate duration (shared code, control) p(99) | ms | 0.005, 0.005, 0.005, 0.005, 0.005 | 0.005, 0.005, 0.004, 0.005, 0.005 | 0.005 | 0.005 | 0 | +0.0% | [+0.0, +20.0] | control (shared code) | control (shared code) |
| Validate duration (shared code, control) max | ms | 0.642, 0.195, 0.242, 0.278, 0.420 | 0.558, 0.341, 0.551, 0.781, 0.277 | 0.278 | 0.551 | 0.273 | -98.2% | [-186.2, +34.0] | control (shared code) | control (shared code) |
| Achieved throughput | req/s | 625.0, 625.0, 625.0, 625.0, 625.0 | 625.0, 625.0, 625.0, 625.0, 625.0 | 625.0 | 625.0 | 0 | +0.0% | [-0.0, +0.0] | no detectable difference | no detectable difference |
| Server decoded_ok rate | op/s | 624.8, 624.8, 624.8, 624.8, 624.8 | 624.8, 624.8, 624.8, 624.8, 624.8 | 624.8 | 624.8 | 0.013 | +0.0% | [+0.0, +0.0] | detectable but below threshold | detectable but below threshold |
| CPU time per successful op | µs/op | 2468, 2469, 2467, 2468, 2467 | 1363, 1368, 1356, 1361, 1362 | 2468 | 1362 | -1105 | +44.8% | [+44.5, +45.0] | goccy better (CI excludes 0, ≥ threshold) | goccy better (CI excludes 0, ≥ threshold) |
| CPU utilization | cores | 1.542, 1.543, 1.541, 1.542, 1.542 | 0.852, 0.855, 0.847, 0.850, 0.851 | 1.542 | 0.851 | -0.691 | +44.8% | [+44.5, +45.0] | goccy better (CI excludes 0, ≥ threshold) | goccy better (CI excludes 0, ≥ threshold) |
| CPU-seconds in window | s | 462.8, 463.0, 462.7, 462.8, 462.7 | 255.7, 256.7, 254.3, 255.3, 255.5 | 462.8 | 255.5 | -207.3 | +44.8% | [+44.5, +45.0] | goccy better (CI excludes 0, ≥ threshold) | goccy better (CI excludes 0, ≥ threshold) |
| Σ handler durations | s | 419.2, 419.4, 419.6, 419.2, 418.6 | 204.0, 206.8, 205.5, 204.9, 204.8 | 419.2 | 204.9 | -214.3 | +51.1% | [+50.7, +51.3] | goccy better (CI excludes 0, ≥ threshold) | goccy better (CI excludes 0, ≥ threshold) |
| Σ decode durations | s | 311.2, 311.1, 311.1, 311.2, 310.6 | 98.4, 100.2, 98.8, 98.6, 98.8 | 311.1 | 98.8 | -212.4 | +68.3% | [+67.8, +68.4] | goccy better (CI excludes 0, ≥ threshold) | goccy better (CI excludes 0, ≥ threshold) |
| Allocated bytes per op | B/op | 1054006, 1054029, 1054029, 1054013, 1054008 | 1061976, 1062464, 1061957, 1061263, 1061308 | 1054013 | 1061957 | 7945 | -0.8% | [-0.8, -0.7] | detectable but below threshold | detectable but below threshold |
| Allocated objects per op | obj/op | 9496, 9496, 9496, 9496, 9496 | 8484, 8484, 8484, 8484, 8484 | 9496 | 8484 | -1012 | +10.7% | [+10.7, +10.7] | goccy better (CI excludes 0, ≥ threshold) | goccy better (CI excludes 0, ≥ threshold) |
| GC cycles per 1k ops | cycles | 204.6, 205.4, 204.5, 204.8, 204.7 | 190.0, 190.7, 190.4, 190.4, 190.5 | 204.7 | 190.4 | -14.3 | +7.0% | [+6.9, +7.3] | detectable but below threshold | detectable but below threshold |
| GC pause total in window | s | 1.107, 1.132, 1.111, 1.106, 1.098 | 0.668, 0.713, 0.677, 0.675, 0.668 | 1.107 | 0.675 | -0.431 | +39.0% | [+35.6, +40.3] | goccy better (CI excludes 0, ≥ threshold) | goccy better (CI excludes 0, ≥ threshold) |
| Live heap at window end | MiB | 4.086, 4.122, 3.836, 4.649, 4.638 | 4.326, 4.587, 4.588, 4.931, 5.202 | 4.122 | 4.588 | 0.466 | -11.3% | [-27.3, +1.3] | no detectable difference | no detectable difference |
| Runtime-mapped memory at end | MiB | 37.8, 41.7, 37.7, 37.7, 37.7 | 41.7, 49.7, 41.8, 42.0, 42.0 | 37.7 | 42.0 | 4.250 | -11.3% | [-31.8, -0.6] | stdlib better (CI excludes 0, ≥ threshold) | stdlib better (CI excludes 0, ≥ threshold) |
| Peak RSS (ps, 5 s samples) | MiB | 41.8, 46.4, 42.2, 45.4, 41.3 | 47.8, 52.7, 46.9, 48.7, 46.5 | 42.2 | 47.8 | 5.594 | -13.2% | [-24.9, -2.5] | stdlib better (CI excludes 0, ≥ threshold) | stdlib better (CI excludes 0, ≥ threshold) |
| Max RSS (getrusage) | MiB | 41.8, 46.4, 42.2, 45.4, 41.3 | 47.8, 52.7, 46.9, 48.8, 46.5 | 42.2 | 47.8 | 5.594 | -13.2% | [-24.9, -2.5] | stdlib better (CI excludes 0, ≥ threshold) | stdlib better (CI excludes 0, ≥ threshold) |

### Pair D · f3-1m · fixed @160 rps

- [M] trials analysed: stdlib-arm 5, goccy-arm 5.
- [M] goccy rep 2: INCLUDED but flagged [host busy 26.5% before trial (top: corespotlightd 143%, Code Helper 27%, WindowServer 18%, Code Helper (Renderer) 6%)]
- [M] window: nominal 360.0 s, wall 360.1 s, drain 0.095 s (per trial; the drain is k6 finishing in-flight requests).

| Metric | unit | stdlib (trials) | goccy (trials) | stdlib median | goccy median | Δ abs | reduction | 95% CI | verdict (pre-registered) | post-hoc verdict |
|---|---|---|---|---:|---:|---:|---:|---|---|---|
| HTTP request duration (k6) min | ms | 8.722, 8.747, 8.723, 8.742, 8.747 | 4.155, 4.171, 4.120, 4.135, 4.114 | 8.742 | 4.135 | -4.607 | +52.7% | [+52.3, +52.9] | inconclusive (shared-code control differs ≥ threshold) (was: goccy better (CI excludes 0, ≥ threshold); controls: Validate duration (shared code, control) med +10.0%) | goccy better (CI excludes 0, ≥ threshold) |
| HTTP request duration (k6) med | ms | 9.160, 9.176, 9.182, 9.159, 9.168 | 4.527, 4.583, 4.506, 4.594, 4.491 | 9.168 | 4.527 | -4.641 | +50.6% | [+49.9, +51.0] | inconclusive (shared-code control differs ≥ threshold) (was: goccy better (CI excludes 0, ≥ threshold); controls: Validate duration (shared code, control) med +10.0%) | goccy better (CI excludes 0, ≥ threshold) |
| HTTP request duration (k6) avg | ms | 9.189, 9.204, 9.206, 9.189, 9.226 | 4.527, 4.582, 4.512, 4.597, 4.499 | 9.204 | 4.527 | -4.677 | +50.8% | [+50.1, +51.1] | inconclusive (shared-code control differs ≥ threshold) (was: goccy better (CI excludes 0, ≥ threshold); controls: Validate duration (shared code, control) med +10.0%) | goccy better (CI excludes 0, ≥ threshold) |
| HTTP request duration (k6) p(95) | ms | 9.442, 9.460, 9.451, 9.432, 9.449 | 4.706, 4.754, 4.698, 4.755, 4.691 | 9.449 | 4.706 | -4.743 | +50.2% | [+49.6, +50.4] | inconclusive (shared-code control differs ≥ threshold) (was: goccy better (CI excludes 0, ≥ threshold); controls: Validate duration (shared code, control) med +10.0%) | goccy better (CI excludes 0, ≥ threshold) |
| HTTP request duration (k6) p(99) | ms | 9.871, 9.826, 9.859, 9.865, 9.866 | 4.866, 4.886, 4.862, 4.954, 4.926 | 9.865 | 4.886 | -4.979 | +50.5% | [+49.8, +50.7] | inconclusive (shared-code control differs ≥ threshold) (was: goccy better (CI excludes 0, ≥ threshold); controls: Validate duration (shared code, control) med +10.0%) | goccy better (CI excludes 0, ≥ threshold) |
| HTTP request duration (k6) max | ms | 26.0, 26.0, 29.4, 27.6, 172.8 | 19.9, 16.0, 19.2, 64.7, 20.0 | 27.6 | 19.9 | -7.685 | +27.8% | [-134.4, +88.5] | inconclusive (trial CV > 10%) | inconclusive (trial CV > 10%) |
| Handler duration (server) min | ms | 8.594, 8.626, 8.605, 8.621, 8.652 | 4.021, 4.063, 4.006, 4.051, 3.986 | 8.621 | 4.021 | -4.600 | +53.4% | [+52.9, +53.8] | inconclusive (shared-code control differs ≥ threshold) (was: goccy better (CI excludes 0, ≥ threshold); controls: Validate duration (shared code, control) med +10.0%) | goccy better (CI excludes 0, ≥ threshold) |
| Handler duration (server) med | ms | 9.037, 9.052, 9.056, 9.036, 9.042 | 4.408, 4.464, 4.388, 4.475, 4.367 | 9.042 | 4.408 | -4.634 | +51.2% | [+50.5, +51.7] | inconclusive (shared-code control differs ≥ threshold) (was: goccy better (CI excludes 0, ≥ threshold); controls: Validate duration (shared code, control) med +10.0%) | goccy better (CI excludes 0, ≥ threshold) |
| Handler duration (server) avg | ms | 9.064, 9.079, 9.079, 9.064, 9.084 | 4.406, 4.462, 4.392, 4.474, 4.375 | 9.079 | 4.406 | -4.672 | +51.5% | [+50.7, +51.8] | inconclusive (shared-code control differs ≥ threshold) (was: goccy better (CI excludes 0, ≥ threshold); controls: Validate duration (shared code, control) med +10.0%) | goccy better (CI excludes 0, ≥ threshold) |
| Handler duration (server) p(95) | ms | 9.314, 9.332, 9.321, 9.305, 9.320 | 4.583, 4.633, 4.577, 4.632, 4.566 | 9.320 | 4.583 | -4.737 | +50.8% | [+50.3, +51.0] | inconclusive (shared-code control differs ≥ threshold) (was: goccy better (CI excludes 0, ≥ threshold); controls: Validate duration (shared code, control) med +10.0%) | goccy better (CI excludes 0, ≥ threshold) |
| Handler duration (server) p(99) | ms | 9.682, 9.656, 9.681, 9.657, 9.692 | 4.729, 4.755, 4.728, 4.798, 4.778 | 9.681 | 4.755 | -4.926 | +50.9% | [+50.4, +51.2] | inconclusive (shared-code control differs ≥ threshold) (was: goccy better (CI excludes 0, ≥ threshold); controls: Validate duration (shared code, control) med +10.0%) | goccy better (CI excludes 0, ≥ threshold) |
| Handler duration (server) max | ms | 25.8, 25.8, 28.0, 27.5, 172.7 | 19.7, 15.9, 18.4, 54.0, 19.1 | 27.5 | 19.1 | -8.332 | +30.3% | [-96.4, +88.9] | inconclusive (trial CV > 10%) | inconclusive (trial CV > 10%) |
| Decode duration (Unmarshal only) min | ms | 6.313, 6.311, 6.308, 6.320, 6.309 | 1.840, 1.870, 1.858, 1.864, 1.825 | 6.311 | 1.858 | -4.453 | +70.6% | [+70.4, +71.1] | inconclusive (shared-code control differs ≥ threshold) (was: goccy better (CI excludes 0, ≥ threshold); controls: Validate duration (shared code, control) med +10.0%) | goccy better (CI excludes 0, ≥ threshold) |
| Decode duration (Unmarshal only) med | ms | 6.630, 6.644, 6.646, 6.626, 6.635 | 2.110, 2.118, 2.110, 2.112, 2.096 | 6.635 | 2.110 | -4.525 | +68.2% | [+68.1, +68.4] | inconclusive (shared-code control differs ≥ threshold) (was: goccy better (CI excludes 0, ≥ threshold); controls: Validate duration (shared code, control) med +10.0%) | goccy better (CI excludes 0, ≥ threshold) |
| Decode duration (Unmarshal only) avg | ms | 6.646, 6.661, 6.659, 6.643, 6.659 | 2.101, 2.110, 2.102, 2.111, 2.091 | 6.659 | 2.102 | -4.557 | +68.4% | [+68.2, +68.6] | inconclusive (shared-code control differs ≥ threshold) (was: goccy better (CI excludes 0, ≥ threshold); controls: Validate duration (shared code, control) med +10.0%) | goccy better (CI excludes 0, ≥ threshold) |
| Decode duration (Unmarshal only) p(95) | ms | 6.908, 6.929, 6.923, 6.900, 6.922 | 2.244, 2.250, 2.246, 2.244, 2.242 | 6.922 | 2.244 | -4.678 | +67.6% | [+67.4, +67.6] | inconclusive (shared-code control differs ≥ threshold) (was: goccy better (CI excludes 0, ≥ threshold); controls: Validate duration (shared code, control) med +10.0%) | goccy better (CI excludes 0, ≥ threshold) |
| Decode duration (Unmarshal only) p(99) | ms | 7.078, 7.097, 7.098, 7.085, 7.085 | 2.354, 2.358, 2.355, 2.387, 2.379 | 7.085 | 2.358 | -4.727 | +66.7% | [+66.3, +66.8] | inconclusive (shared-code control differs ≥ threshold) (was: goccy better (CI excludes 0, ≥ threshold); controls: Validate duration (shared code, control) med +10.0%) | goccy better (CI excludes 0, ≥ threshold) |
| Decode duration (Unmarshal only) max | ms | 23.3, 23.4, 24.9, 25.1, 170.5 | 14.1, 13.6, 11.7, 46.0, 9.888 | 24.9 | 13.6 | -11.3 | +45.4% | [-84.8, +92.0] | inconclusive (trial CV > 10%) | inconclusive (trial CV > 10%) |
| Body read duration (shared code, control) min | ms | 0.058, 0.072, 0.068, 0.059, 0.057 | 0.105, 0.094, 0.078, 0.101, 0.111 | 0.059 | 0.101 | 0.042 | -71.2% | [-88.1, -30.6] | control (shared code) | control (shared code) |
| Body read duration (shared code, control) med | ms | 0.331, 0.333, 0.336, 0.338, 0.337 | 0.363, 0.338, 0.346, 0.361, 0.341 | 0.336 | 0.346 | 0.01 | -3.0% | [-9.0, -0.6] | control (shared code) | control (shared code) |
| Body read duration (shared code, control) avg | ms | 0.320, 0.321, 0.324, 0.323, 0.328 | 0.344, 0.330, 0.336, 0.342, 0.335 | 0.323 | 0.336 | 0.013 | -3.9% | [-6.7, -2.1] | control (shared code) | control (shared code) |
| Body read duration (shared code, control) p(95) | ms | 0.406, 0.403, 0.403, 0.403, 0.403 | 0.408, 0.399, 0.404, 0.408, 0.406 | 0.403 | 0.406 | 0.003 | -0.7% | [-1.2, +1.0] | control (shared code) | control (shared code) |
| Body read duration (shared code, control) p(99) | ms | 0.487, 0.478, 0.469, 0.478, 0.484 | 0.426, 0.417, 0.423, 0.427, 0.426 | 0.478 | 0.426 | -0.052 | +10.9% | [+9.2, +13.1] | control (shared code) | control (shared code) |
| Body read duration (shared code, control) max | ms | 4.295, 4.495, 13.9, 3.977, 40.4 | 3.898, 4.474, 4.593, 24.7, 4.610 | 4.495 | 4.593 | 0.098 | -2.2% | [-449.5, +88.6] | control (shared code) | control (shared code) |
| Policy scan duration (shared code, control) min | ms | 1.977, 1.970, 1.964, 1.976, 1.969 | 1.839, 1.860, 1.839, 1.845, 1.838 | 1.970 | 1.839 | -0.131 | +6.6% | [+5.6, +7.0] | control (shared code) | control (shared code) |
| Policy scan duration (shared code, control) med | ms | 2.082, 2.079, 2.078, 2.078, 2.077 | 1.940, 2.009, 1.933, 2.008, 1.925 | 2.078 | 1.940 | -0.138 | +6.6% | [+3.3, +7.4] | control (shared code) | control (shared code) |
| Policy scan duration (shared code, control) avg | ms | 2.088, 2.086, 2.085, 2.087, 2.087 | 1.952, 2.013, 1.945, 2.012, 1.939 | 2.087 | 1.952 | -0.135 | +6.5% | [+3.5, +7.1] | control (shared code) | control (shared code) |
| Policy scan duration (shared code, control) p(95) | ms | 2.189, 2.187, 2.184, 2.187, 2.185 | 2.041, 2.072, 2.034, 2.072, 2.029 | 2.187 | 2.041 | -0.146 | +6.7% | [+5.2, +7.2] | control (shared code) | control (shared code) |
| Policy scan duration (shared code, control) p(99) | ms | 2.287, 2.287, 2.276, 2.279, 2.277 | 2.102, 2.118, 2.099, 2.126, 2.111 | 2.279 | 2.111 | -0.168 | +7.4% | [+6.7, +8.1] | control (shared code) | control (shared code) |
| Policy scan duration (shared code, control) max | ms | 13.3, 5.650, 15.1, 14.7, 47.2 | 12.3, 10.5, 15.1, 21.0, 13.5 | 14.7 | 13.5 | -1.195 | +8.1% | [-139.2, +71.4] | control (shared code) | control (shared code) |
| Validate duration (shared code, control) min | ms | 0.007, 0.007, 0.007, 0.007, 0.007 | 0.007, 0.007, 0.007, 0.007, 0.007 | 0.007 | 0.007 | 0 | +0.0% | [+0.0, +0.0] | control (shared code) | control (shared code) |
| Validate duration (shared code, control) med | ms | 0.010, 0.010, 0.010, 0.010, 0.010 | 0.009, 0.009, 0.009, 0.009, 0.009 | 0.010 | 0.009 | -0.001 | +10.0% | [+10.0, +10.0] | control (shared code) (control differs ≥ threshold (within timing resolution: ignored post-hoc)) | control (shared code) |
| Validate duration (shared code, control) avg | ms | 0.010, 0.010, 0.010, 0.010, 0.010 | 0.00928, 0.00926, 0.00927, 0.00955, 0.00935 | 0.010 | 0.00928 | -0.000938 | +9.2% | [+6.5, +9.6] | control (shared code) | control (shared code) |
| Validate duration (shared code, control) p(95) | ms | 0.014, 0.014, 0.014, 0.014, 0.014 | 0.013, 0.013, 0.013, 0.014, 0.013 | 0.014 | 0.013 | -0.001 | +7.1% | [+0.0, +7.1] | control (shared code) | control (shared code) |
| Validate duration (shared code, control) p(99) | ms | 0.021, 0.021, 0.021, 0.021, 0.021 | 0.019, 0.019, 0.019, 0.021, 0.020 | 0.021 | 0.019 | -0.002 | +9.5% | [+0.0, +9.5] | control (shared code) | control (shared code) |
| Validate duration (shared code, control) max | ms | 0.205, 0.093, 0.176, 0.240, 0.170 | 0.102, 0.135, 0.095, 0.493, 0.804 | 0.176 | 0.135 | -0.041 | +23.3% | [-372.9, +53.7] | control (shared code) | control (shared code) |
| Achieved throughput | req/s | 160.0, 160.0, 160.0, 160.0, 160.0 | 160.0, 160.0, 160.0, 160.0, 160.0 | 160.0 | 160.0 | 0 | +0.0% | [+0.0, +0.0] | no detectable difference | no detectable difference |
| Server decoded_ok rate | op/s | 160.0, 160.0, 160.0, 160.0, 160.0 | 160.0, 160.0, 160.0, 160.0, 160.0 | 160.0 | 160.0 | 0.00123 | +0.0% | [-0.0, +0.0] | no detectable difference | no detectable difference |
| CPU time per successful op | µs/op | 9657, 9679, 9675, 9650, 9666 | 5002, 5069, 4992, 5055, 4980 | 9666 | 5002 | -4664 | +48.2% | [+47.6, +48.5] | inconclusive (shared-code control differs ≥ threshold) (was: goccy better (CI excludes 0, ≥ threshold); controls: Validate duration (shared code, control) med +10.0%) | goccy better (CI excludes 0, ≥ threshold) |
| CPU utilization | cores | 1.545, 1.549, 1.548, 1.544, 1.546 | 0.800, 0.811, 0.799, 0.809, 0.797 | 1.546 | 0.800 | -0.746 | +48.2% | [+47.6, +48.5] | inconclusive (shared-code control differs ≥ threshold) (was: goccy better (CI excludes 0, ≥ threshold); controls: Validate duration (shared code, control) med +10.0%) | goccy better (CI excludes 0, ≥ threshold) |
| CPU-seconds in window | s | 556.3, 557.6, 557.3, 556.0, 556.9 | 288.2, 292.0, 287.6, 291.2, 286.9 | 556.9 | 288.2 | -268.7 | +48.2% | [+47.6, +48.5] | inconclusive (shared-code control differs ≥ threshold) (was: goccy better (CI excludes 0, ≥ threshold); controls: Validate duration (shared code, control) med +10.0%) | goccy better (CI excludes 0, ≥ threshold) |
| Σ handler durations | s | 522.2, 523.0, 523.0, 522.2, 523.3 | 253.9, 257.1, 253.0, 257.7, 252.0 | 523.0 | 253.9 | -269.2 | +51.5% | [+50.7, +51.8] | inconclusive (shared-code control differs ≥ threshold) (was: goccy better (CI excludes 0, ≥ threshold); controls: Validate duration (shared code, control) med +10.0%) | goccy better (CI excludes 0, ≥ threshold) |
| Σ decode durations | s | 382.8, 383.8, 383.6, 382.7, 383.6 | 121.1, 121.6, 121.1, 121.6, 120.5 | 383.6 | 121.1 | -262.5 | +68.4% | [+68.2, +68.6] | inconclusive (shared-code control differs ≥ threshold) (was: goccy better (CI excludes 0, ≥ threshold); controls: Validate duration (shared code, control) med +10.0%) | goccy better (CI excludes 0, ≥ threshold) |
| Allocated bytes per op | B/op | 4316165, 4316146, 4316181, 4316114, 4316131 | 4720849, 4720560, 4719128, 4730544, 4719017 | 4316146 | 4720560 | 404414 | -9.4% | [-9.6, -9.3] | inconclusive (shared-code control differs ≥ threshold) (was: detectable but below threshold; controls: Validate duration (shared code, control) med +10.0%) | detectable but below threshold |
| Allocated objects per op | obj/op | 38097, 38096, 38097, 38096, 38096 | 33948, 33948, 33948, 33949, 33948 | 38096 | 33948 | -4148 | +10.9% | [+10.9, +10.9] | inconclusive (shared-code control differs ≥ threshold) (was: goccy better (CI excludes 0, ≥ threshold); controls: Validate duration (shared code, control) med +10.0%) | goccy better (CI excludes 0, ≥ threshold) |
| GC cycles per 1k ops | cycles | 954.5, 956.8, 955.2, 954.9, 954.5 | 899.7, 899.7, 898.6, 900.3, 898.9 | 954.9 | 899.7 | -55.3 | +5.8% | [+5.7, +6.0] | inconclusive (shared-code control differs ≥ threshold) (was: detectable but below threshold; controls: Validate duration (shared code, control) med +10.0%) | detectable but below threshold |
| GC pause total in window | s | 1.533, 1.549, 1.538, 1.541, 1.552 | 0.992, 0.997, 0.994, 1.024, 1.007 | 1.541 | 0.997 | -0.543 | +35.3% | [+33.5, +35.9] | inconclusive (shared-code control differs ≥ threshold) (was: goccy better (CI excludes 0, ≥ threshold); controls: Validate duration (shared code, control) med +10.0%) | goccy better (CI excludes 0, ≥ threshold) |
| Live heap at window end | MiB | 2.700, 5.390, 5.198, 2.735, 4.460 | 4.558, 6.439, 4.425, 4.446, 4.403 | 4.460 | 4.446 | -0.014 | +0.3% | [-68.8, +17.5] | inconclusive (trial CV > 10%) | inconclusive (trial CV > 10%) |
| Runtime-mapped memory at end | MiB | 33.0, 33.3, 33.5, 33.3, 69.6 | 33.5, 61.6, 33.5, 61.6, 53.5 | 33.3 | 53.5 | 20.2 | -60.9% | [-85.3, +23.2] | inconclusive (trial CV > 10%) | inconclusive (trial CV > 10%) |
| Peak RSS (ps, 5 s samples) | MiB | 37.2, 37.1, 41.0, 39.5, 73.9 | 39.8, 64.7, 39.2, 67.2, 59.9 | 39.5 | 59.9 | 20.4 | -51.7% | [-74.4, +18.9] | inconclusive (trial CV > 10%) | inconclusive (trial CV > 10%) |
| Max RSS (getrusage) | MiB | 37.2, 37.1, 41.0, 39.5, 73.9 | 39.8, 64.7, 39.2, 67.2, 59.9 | 39.5 | 59.9 | 20.4 | -51.7% | [-74.4, +18.9] | inconclusive (trial CV > 10%) | inconclusive (trial CV > 10%) |

### Pair D · f3-1m · fixed @256 rps

- [M] trials analysed: stdlib-arm 5, goccy-arm 5.
- [M] goccy rep 2: INCLUDED but flagged [host busy 34.7% before trial (top: corespotlightd 130%, duetexpertd 76%, WindowServer 24%, corebrightnessd 12%, ControlCenter 7%)]
- [M] window: nominal 360.0 s, wall 360.1 s, drain 0.118 s (per trial; the drain is k6 finishing in-flight requests).

| Metric | unit | stdlib (trials) | goccy (trials) | stdlib median | goccy median | Δ abs | reduction | 95% CI | verdict (pre-registered) | post-hoc verdict |
|---|---|---|---|---:|---:|---:|---:|---|---|---|
| HTTP request duration (k6) min | ms | 8.934, 8.924, 8.923, 8.870, 8.962 | 0.193, 4.135, 4.189, 4.184, 4.177 | 8.924 | 4.177 | -4.747 | +53.2% | [+52.9, +97.8] | inconclusive (trial CV > 10%) | inconclusive (trial CV > 10%) |
| HTTP request duration (k6) med | ms | 9.327, 9.331, 9.341, 9.358, 9.349 | 4.630, 4.629, 4.627, 4.628, 4.628 | 9.341 | 4.628 | -4.713 | +50.5% | [+50.4, +50.5] | inconclusive (shared-code control differs ≥ threshold) (was: goccy better (CI excludes 0, ≥ threshold); controls: Body read duration (shared code, control) med -47.6%) | inconclusive (control differs beyond resolution) |
| HTTP request duration (k6) avg | ms | 9.390, 9.393, 9.402, 9.426, 9.418 | 4.655, 4.627, 4.622, 4.684, 4.622 | 9.402 | 4.627 | -4.775 | +50.8% | [+50.2, +50.9] | inconclusive (shared-code control differs ≥ threshold) (was: goccy better (CI excludes 0, ≥ threshold); controls: Body read duration (shared code, control) med -47.6%) | inconclusive (control differs beyond resolution) |
| HTTP request duration (k6) p(95) | ms | 9.711, 9.708, 9.705, 9.751, 9.727 | 4.876, 4.859, 4.850, 4.855, 4.853 | 9.711 | 4.855 | -4.856 | +50.0% | [+49.8, +50.2] | inconclusive (shared-code control differs ≥ threshold) (was: goccy better (CI excludes 0, ≥ threshold); controls: Body read duration (shared code, control) med -47.6%) | inconclusive (control differs beyond resolution) |
| HTTP request duration (k6) p(99) | ms | 10.6, 10.6, 10.6, 10.8, 10.9 | 5.305, 5.058, 5.043, 5.093, 5.040 | 10.6 | 5.058 | -5.569 | +52.4% | [+50.1, +53.5] | inconclusive (shared-code control differs ≥ threshold) (was: goccy better (CI excludes 0, ≥ threshold); controls: Body read duration (shared code, control) med -47.6%) | inconclusive (control differs beyond resolution) |
| HTTP request duration (k6) max | ms | 33.6, 35.8, 35.0, 29.7, 29.0 | 40.5, 21.7, 18.7, 201.5, 15.6 | 33.6 | 21.7 | -11.9 | +35.4% | [-500.2, +53.4] | inconclusive (trial CV > 10%) | inconclusive (trial CV > 10%) |
| Handler duration (server) min | ms | 8.813, 8.799, 8.792, 8.765, 8.833 | 4.010, 4.029, 4.015, 4.019, 4.028 | 8.799 | 4.019 | -4.780 | +54.3% | [+54.1, +54.5] | inconclusive (shared-code control differs ≥ threshold) (was: goccy better (CI excludes 0, ≥ threshold); controls: Body read duration (shared code, control) med -47.6%) | inconclusive (control differs beyond resolution) |
| Handler duration (server) med | ms | 9.218, 9.221, 9.223, 9.241, 9.230 | 4.520, 4.515, 4.513, 4.513, 4.512 | 9.223 | 4.513 | -4.710 | +51.1% | [+51.0, +51.2] | inconclusive (shared-code control differs ≥ threshold) (was: goccy better (CI excludes 0, ≥ threshold); controls: Body read duration (shared code, control) med -47.6%) | inconclusive (control differs beyond resolution) |
| Handler duration (server) avg | ms | 9.273, 9.276, 9.276, 9.299, 9.290 | 4.530, 4.504, 4.497, 4.529, 4.498 | 9.276 | 4.504 | -4.772 | +51.4% | [+51.2, +51.6] | inconclusive (shared-code control differs ≥ threshold) (was: goccy better (CI excludes 0, ≥ threshold); controls: Body read duration (shared code, control) med -47.6%) | inconclusive (control differs beyond resolution) |
| Handler duration (server) p(95) | ms | 9.583, 9.581, 9.568, 9.612, 9.590 | 4.764, 4.740, 4.725, 4.730, 4.733 | 9.583 | 4.733 | -4.850 | +50.6% | [+50.3, +50.8] | inconclusive (shared-code control differs ≥ threshold) (was: goccy better (CI excludes 0, ≥ threshold); controls: Body read duration (shared code, control) med -47.6%) | inconclusive (control differs beyond resolution) |
| Handler duration (server) p(99) | ms | 10.4, 10.4, 10.4, 10.5, 10.6 | 5.079, 4.925, 4.909, 4.936, 4.910 | 10.4 | 4.925 | -5.441 | +52.5% | [+51.0, +53.5] | inconclusive (shared-code control differs ≥ threshold) (was: goccy better (CI excludes 0, ≥ threshold); controls: Body read duration (shared code, control) med -47.6%) | inconclusive (control differs beyond resolution) |
| Handler duration (server) max | ms | 33.3, 35.7, 34.8, 29.5, 28.3 | 39.8, 20.9, 17.8, 200.7, 10.9 | 33.3 | 20.9 | -12.4 | +37.1% | [-502.4, +67.2] | inconclusive (trial CV > 10%) | inconclusive (trial CV > 10%) |
| Decode duration (Unmarshal only) min | ms | 6.449, 6.460, 6.459, 6.456, 6.451 | 1.902, 1.908, 1.905, 1.902, 1.913 | 6.456 | 1.905 | -4.551 | +70.5% | [+70.4, +70.6] | inconclusive (shared-code control differs ≥ threshold) (was: goccy better (CI excludes 0, ≥ threshold); controls: Body read duration (shared code, control) med -47.6%) | inconclusive (control differs beyond resolution) |
| Decode duration (Unmarshal only) med | ms | 6.878, 6.886, 6.879, 6.897, 6.885 | 2.164, 2.167, 2.165, 2.167, 2.168 | 6.885 | 2.167 | -4.718 | +68.5% | [+68.5, +68.6] | inconclusive (shared-code control differs ≥ threshold) (was: goccy better (CI excludes 0, ≥ threshold); controls: Body read duration (shared code, control) med -47.6%) | inconclusive (control differs beyond resolution) |
| Decode duration (Unmarshal only) avg | ms | 6.929, 6.935, 6.928, 6.948, 6.938 | 2.175, 2.164, 2.158, 2.173, 2.162 | 6.935 | 2.164 | -4.770 | +68.8% | [+68.6, +68.9] | inconclusive (shared-code control differs ≥ threshold) (was: goccy better (CI excludes 0, ≥ threshold); controls: Body read duration (shared code, control) med -47.6%) | inconclusive (control differs beyond resolution) |
| Decode duration (Unmarshal only) p(95) | ms | 7.204, 7.210, 7.193, 7.228, 7.209 | 2.386, 2.364, 2.353, 2.355, 2.358 | 7.209 | 2.358 | -4.851 | +67.3% | [+66.9, +67.4] | inconclusive (shared-code control differs ≥ threshold) (was: goccy better (CI excludes 0, ≥ threshold); controls: Body read duration (shared code, control) med -47.6%) | inconclusive (control differs beyond resolution) |
| Decode duration (Unmarshal only) p(99) | ms | 7.684, 7.650, 7.644, 7.776, 7.779 | 2.628, 2.521, 2.506, 2.518, 2.509 | 7.684 | 2.518 | -5.166 | +67.2% | [+65.8, +67.7] | inconclusive (shared-code control differs ≥ threshold) (was: goccy better (CI excludes 0, ≥ threshold); controls: Body read duration (shared code, control) med -47.6%) | inconclusive (control differs beyond resolution) |
| Decode duration (Unmarshal only) max | ms | 23.5, 33.4, 32.4, 25.4, 26.0 | 34.5, 9.738, 13.1, 107.1, 8.502 | 26.0 | 13.1 | -12.9 | +49.6% | [-312.6, +70.8] | inconclusive (trial CV > 10%) | inconclusive (trial CV > 10%) |
| Body read duration (shared code, control) min | ms | 0.055, 0.051, 0.057, 0.057, 0.062 | 0.055, 0.059, 0.068, 0.053, 0.065 | 0.057 | 0.059 | 0.002 | -3.5% | [-23.6, +7.0] | control (shared code) | control (shared code) |
| Body read duration (shared code, control) med | ms | 0.226, 0.224, 0.229, 0.227, 0.228 | 0.347, 0.335, 0.336, 0.329, 0.334 | 0.227 | 0.335 | 0.108 | -47.6% | [-52.9, -44.9] | control (shared code) (control differs ≥ threshold) | control (shared code) |
| Body read duration (shared code, control) avg | ms | 0.239, 0.237, 0.244, 0.243, 0.244 | 0.323, 0.317, 0.316, 0.324, 0.316 | 0.243 | 0.317 | 0.075 | -30.9% | [-35.4, -29.5] | control (shared code) | control (shared code) |
| Body read duration (shared code, control) p(95) | ms | 0.368, 0.366, 0.377, 0.380, 0.377 | 0.413, 0.414, 0.411, 0.413, 0.413 | 0.377 | 0.413 | 0.036 | -9.5% | [-12.8, -8.7] | control (shared code) | control (shared code) |
| Body read duration (shared code, control) p(99) | ms | 0.529, 0.527, 0.547, 0.546, 0.545 | 0.460, 0.458, 0.453, 0.455, 0.453 | 0.545 | 0.455 | -0.090 | +16.5% | [+13.1, +17.1] | control (shared code) | control (shared code) |
| Body read duration (shared code, control) max | ms | 8.163, 7.410, 8.770, 8.660, 8.751 | 19.3, 5.962, 7.021, 98.7, 4.559 | 8.660 | 7.021 | -1.639 | +18.9% | [-1039.1, +47.4] | control (shared code) | control (shared code) |
| Policy scan duration (shared code, control) min | ms | 1.994, 1.993, 2.000, 1.999, 1.996 | 1.927, 1.943, 1.945, 1.927, 1.939 | 1.996 | 1.939 | -0.057 | +2.9% | [+2.5, +3.6] | control (shared code) | control (shared code) |
| Policy scan duration (shared code, control) med | ms | 2.048, 2.047, 2.048, 2.049, 2.049 | 1.994, 1.992, 1.993, 1.993, 1.991 | 2.048 | 1.993 | -0.055 | +2.7% | [+2.6, +2.8] | control (shared code) | control (shared code) |
| Policy scan duration (shared code, control) avg | ms | 2.093, 2.093, 2.092, 2.096, 2.095 | 2.020, 2.009, 2.010, 2.018, 2.007 | 2.093 | 2.010 | -0.083 | +4.0% | [+3.5, +4.2] | control (shared code) | control (shared code) |
| Policy scan duration (shared code, control) p(95) | ms | 2.218, 2.219, 2.218, 2.228, 2.222 | 2.102, 2.093, 2.093, 2.094, 2.091 | 2.219 | 2.093 | -0.126 | +5.7% | [+5.3, +6.1] | control (shared code) | control (shared code) |
| Policy scan duration (shared code, control) p(99) | ms | 2.433, 2.437, 2.426, 2.473, 2.462 | 2.216, 2.164, 2.165, 2.176, 2.160 | 2.437 | 2.165 | -0.272 | +11.2% | [+9.1, +12.5] | control (shared code) | control (shared code) |
| Policy scan duration (shared code, control) max | ms | 12.3, 15.3, 7.450, 10.5, 15.9 | 14.4, 13.7, 9.323, 73.2, 6.304 | 12.3 | 13.7 | 1.417 | -11.6% | [-497.2, +48.6] | control (shared code) | control (shared code) |
| Validate duration (shared code, control) min | ms | 0.007, 0.007, 0.008, 0.007, 0.008 | 0.007, 0.007, 0.007, 0.007, 0.007 | 0.007 | 0.007 | 0 | +0.0% | [+0.0, +12.5] | control (shared code) | control (shared code) |
| Validate duration (shared code, control) med | ms | 0.010, 0.010, 0.012, 0.012, 0.012 | 0.011, 0.013, 0.013, 0.013, 0.013 | 0.012 | 0.013 | 0.001 | -8.3% | [-30.0, +8.3] | control (shared code) | control (shared code) |
| Validate duration (shared code, control) avg | ms | 0.011, 0.011, 0.013, 0.013, 0.013 | 0.013, 0.014, 0.013, 0.014, 0.013 | 0.013 | 0.013 | 0.000795 | -6.3% | [-19.4, -0.7] | control (shared code) | control (shared code) |
| Validate duration (shared code, control) p(95) | ms | 0.016, 0.016, 0.017, 0.017, 0.017 | 0.020, 0.019, 0.019, 0.019, 0.019 | 0.017 | 0.019 | 0.002 | -11.8% | [-18.7, -11.8] | control (shared code) | control (shared code) |
| Validate duration (shared code, control) p(99) | ms | 0.025, 0.025, 0.023, 0.024, 0.025 | 0.032, 0.026, 0.025, 0.026, 0.025 | 0.025 | 0.026 | 0.001 | -4.0% | [-28.0, +0.0] | control (shared code) | control (shared code) |
| Validate duration (shared code, control) max | ms | 0.535, 0.643, 1.731, 2.436, 5.340 | 1.556, 0.447, 0.284, 0.362, 0.222 | 1.731 | 0.362 | -1.369 | +79.1% | [+10.1, +93.2] | control (shared code) | control (shared code) |
| Achieved throughput | req/s | 256.0, 256.0, 256.0, 256.0, 256.0 | 256.0, 256.0, 256.0, 256.0, 256.0 | 256.0 | 256.0 | 0 | +0.0% | [-0.0, +0.0] | no detectable difference | no detectable difference |
| Server decoded_ok rate | op/s | 256.0, 256.0, 256.0, 256.0, 256.0 | 256.0, 256.0, 256.0, 256.0, 256.0 | 256.0 | 256.0 | 0.00163 | +0.0% | [-0.0, +0.0] | no detectable difference | no detectable difference |
| CPU time per successful op | µs/op | 9780, 9778, 9788, 9815, 9805 | 5062, 5061, 5053, 5054, 5056 | 9788 | 5056 | -4732 | +48.3% | [+48.2, +48.5] | inconclusive (shared-code control differs ≥ threshold) (was: goccy better (CI excludes 0, ≥ threshold); controls: Body read duration (shared code, control) med -47.6%) | inconclusive (control differs beyond resolution) |
| CPU utilization | cores | 2.503, 2.503, 2.505, 2.512, 2.510 | 1.296, 1.295, 1.293, 1.294, 1.294 | 2.505 | 1.294 | -1.211 | +48.3% | [+48.2, +48.5] | inconclusive (shared-code control differs ≥ threshold) (was: goccy better (CI excludes 0, ≥ threshold); controls: Body read duration (shared code, control) med -47.6%) | inconclusive (control differs beyond resolution) |
| CPU-seconds in window | s | 901.5, 901.3, 902.2, 904.8, 903.8 | 466.6, 466.5, 465.8, 465.9, 466.1 | 902.2 | 466.1 | -436.2 | +48.3% | [+48.2, +48.5] | inconclusive (shared-code control differs ≥ threshold) (was: goccy better (CI excludes 0, ≥ threshold); controls: Body read duration (shared code, control) med -47.6%) | inconclusive (control differs beyond resolution) |
| Σ handler durations | s | 854.7, 855.0, 855.0, 857.2, 856.3 | 417.6, 415.2, 414.5, 417.4, 414.6 | 855.0 | 415.2 | -439.9 | +51.4% | [+51.2, +51.6] | inconclusive (shared-code control differs ≥ threshold) (was: goccy better (CI excludes 0, ≥ threshold); controls: Body read duration (shared code, control) med -47.6%) | inconclusive (control differs beyond resolution) |
| Σ decode durations | s | 638.7, 639.2, 638.6, 640.4, 639.5 | 200.5, 199.5, 199.0, 200.3, 199.3 | 639.2 | 199.5 | -439.7 | +68.8% | [+68.6, +68.9] | inconclusive (shared-code control differs ≥ threshold) (was: goccy better (CI excludes 0, ≥ threshold); controls: Body read duration (shared code, control) med -47.6%) | inconclusive (control differs beyond resolution) |
| Allocated bytes per op | B/op | 4315101, 4315084, 4315049, 4315043, 4315079 | 4754247, 4755518, 4755454, 4754134, 4759494 | 4315079 | 4755454 | 440375 | -10.2% | [-10.3, -10.2] | inconclusive (shared-code control differs ≥ threshold) (was: stdlib better (CI excludes 0, ≥ threshold); controls: Body read duration (shared code, control) med -47.6%) | inconclusive (control differs beyond resolution) |
| Allocated objects per op | obj/op | 38091, 38091, 38091, 38091, 38091 | 33947, 33947, 33948, 33947, 33947 | 38091 | 33947 | -4144 | +10.9% | [+10.9, +10.9] | inconclusive (shared-code control differs ≥ threshold) (was: goccy better (CI excludes 0, ≥ threshold); controls: Body read duration (shared code, control) med -47.6%) | inconclusive (control differs beyond resolution) |
| GC cycles per 1k ops | cycles | 739.4, 737.5, 734.1, 735.7, 738.9 | 777.0, 789.6, 791.9, 791.2, 793.5 | 737.5 | 791.2 | 53.7 | -7.3% | [-7.9, -5.4] | inconclusive (shared-code control differs ≥ threshold) (was: detectable but below threshold; controls: Body read duration (shared code, control) med -47.6%) | inconclusive (control differs beyond resolution) |
| GC pause total in window | s | 2.540, 2.547, 2.566, 2.594, 2.562 | 1.755, 1.721, 1.687, 1.798, 1.705 | 2.562 | 1.721 | -0.841 | +32.8% | [+29.8, +34.3] | inconclusive (shared-code control differs ≥ threshold) (was: goccy better (CI excludes 0, ≥ threshold); controls: Body read duration (shared code, control) med -47.6%) | inconclusive (control differs beyond resolution) |
| Live heap at window end | MiB | 3.562, 5.910, 6.004, 5.133, 4.779 | 7.101, 6.655, 5.052, 5.831, 4.713 | 5.133 | 5.831 | 0.698 | -13.6% | [-63.7, +15.9] | inconclusive (trial CV > 10%) | inconclusive (trial CV > 10%) |
| Runtime-mapped memory at end | MiB | 53.8, 45.8, 45.5, 41.5, 41.5 | 70.2, 49.6, 53.6, 90.2, 41.5 | 45.5 | 53.6 | 8.129 | -17.9% | [-98.5, +8.8] | inconclusive (trial CV > 10%) | inconclusive (trial CV > 10%) |
| Peak RSS (ps, 5 s samples) | MiB | 58.5, 52.7, 49.2, 46.2, 50.8 | 74.5, 54.2, 58.1, 96.6, 49.0 | 50.8 | 58.1 | 7.234 | -14.2% | [-90.1, +6.9] | inconclusive (trial CV > 10%) | inconclusive (trial CV > 10%) |
| Max RSS (getrusage) | MiB | 58.5, 52.7, 49.2, 46.2, 50.8 | 74.5, 54.2, 58.1, 96.6, 49.0 | 50.8 | 58.1 | 7.234 | -14.2% | [-90.1, +6.9] | inconclusive (trial CV > 10%) | inconclusive (trial CV > 10%) |

### Pair D · f3-1m · fixed @256 rps · profile=allocs

- [M] trials analysed: stdlib-arm 1, goccy-arm 1.
- [M] window: nominal 180.0 s, wall 180.1 s, drain 0.095 s (per trial; the drain is k6 finishing in-flight requests).

| Metric | unit | stdlib (trials) | goccy (trials) | stdlib median | goccy median | Δ abs | reduction | 95% CI | verdict (pre-registered) | post-hoc verdict |
|---|---|---|---|---:|---:|---:|---:|---|---|---|
| HTTP request duration (k6) min | ms | 8.878 | 4.191 | 8.878 | 4.191 | -4.687 | +52.8% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| HTTP request duration (k6) med | ms | 9.335 | 4.629 | 9.335 | 4.629 | -4.706 | +50.4% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| HTTP request duration (k6) avg | ms | 9.432 | 4.632 | 9.432 | 4.632 | -4.800 | +50.9% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| HTTP request duration (k6) p(95) | ms | 9.810 | 4.891 | 9.810 | 4.891 | -4.919 | +50.1% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| HTTP request duration (k6) p(99) | ms | 11.6 | 5.178 | 11.6 | 5.178 | -6.448 | +55.5% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| HTTP request duration (k6) max | ms | 98.0 | 15.5 | 98.0 | 15.5 | -82.5 | +84.2% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Handler duration (server) min | ms | 8.790 | 4.004 | 8.790 | 4.004 | -4.786 | +54.4% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Handler duration (server) med | ms | 9.218 | 4.513 | 9.218 | 4.513 | -4.705 | +51.0% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Handler duration (server) avg | ms | 9.300 | 4.507 | 9.300 | 4.507 | -4.793 | +51.5% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Handler duration (server) p(95) | ms | 9.664 | 4.774 | 9.664 | 4.774 | -4.890 | +50.6% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Handler duration (server) p(99) | ms | 11.3 | 5.001 | 11.3 | 5.001 | -6.271 | +55.6% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Handler duration (server) max | ms | 21.7 | 15.3 | 21.7 | 15.3 | -6.357 | +29.3% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Decode duration (Unmarshal only) min | ms | 6.383 | 1.908 | 6.383 | 1.908 | -4.475 | +70.1% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Decode duration (Unmarshal only) med | ms | 6.879 | 2.176 | 6.879 | 2.176 | -4.703 | +68.4% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Decode duration (Unmarshal only) avg | ms | 6.949 | 2.166 | 6.949 | 2.166 | -4.783 | +68.8% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Decode duration (Unmarshal only) p(95) | ms | 7.244 | 2.395 | 7.244 | 2.395 | -4.849 | +66.9% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Decode duration (Unmarshal only) p(99) | ms | 8.316 | 2.592 | 8.316 | 2.592 | -5.724 | +68.8% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Decode duration (Unmarshal only) max | ms | 19.2 | 10.5 | 19.2 | 10.5 | -8.701 | +45.4% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Body read duration (shared code, control) min | ms | 0.058 | 0.055 | 0.058 | 0.055 | -0.003 | +5.2% | n/a | control (shared code) | control (shared code) |
| Body read duration (shared code, control) med | ms | 0.231 | 0.312 | 0.231 | 0.312 | 0.081 | -35.1% | n/a | control (shared code) (control differs ≥ threshold) | control (shared code) |
| Body read duration (shared code, control) avg | ms | 0.247 | 0.310 | 0.247 | 0.310 | 0.063 | -25.6% | n/a | control (shared code) | control (shared code) |
| Body read duration (shared code, control) p(95) | ms | 0.380 | 0.416 | 0.380 | 0.416 | 0.036 | -9.5% | n/a | control (shared code) | control (shared code) |
| Body read duration (shared code, control) p(99) | ms | 0.538 | 0.480 | 0.538 | 0.480 | -0.058 | +10.8% | n/a | control (shared code) | control (shared code) |
| Body read duration (shared code, control) max | ms | 5.906 | 7.055 | 5.906 | 7.055 | 1.149 | -19.5% | n/a | control (shared code) | control (shared code) |
| Policy scan duration (shared code, control) min | ms | 1.996 | 1.946 | 1.996 | 1.946 | -0.050 | +2.5% | n/a | control (shared code) | control (shared code) |
| Policy scan duration (shared code, control) med | ms | 2.047 | 1.995 | 2.047 | 1.995 | -0.052 | +2.5% | n/a | control (shared code) | control (shared code) |
| Policy scan duration (shared code, control) avg | ms | 2.092 | 2.016 | 2.092 | 2.016 | -0.075 | +3.6% | n/a | control (shared code) | control (shared code) |
| Policy scan duration (shared code, control) p(95) | ms | 2.220 | 2.104 | 2.220 | 2.104 | -0.116 | +5.2% | n/a | control (shared code) | control (shared code) |
| Policy scan duration (shared code, control) p(99) | ms | 2.600 | 2.192 | 2.600 | 2.192 | -0.408 | +15.7% | n/a | control (shared code) | control (shared code) |
| Policy scan duration (shared code, control) max | ms | 10.2 | 9.564 | 10.2 | 9.564 | -0.634 | +6.2% | n/a | control (shared code) | control (shared code) |
| Validate duration (shared code, control) min | ms | 0.008 | 0.007 | 0.008 | 0.007 | -0.001 | +12.5% | n/a | control (shared code) | control (shared code) |
| Validate duration (shared code, control) med | ms | 0.012 | 0.014 | 0.012 | 0.014 | 0.002 | -16.7% | n/a | control (shared code) (control differs ≥ threshold (within timing resolution: ignored post-hoc)) | control (shared code) |
| Validate duration (shared code, control) avg | ms | 0.013 | 0.014 | 0.013 | 0.014 | 0.00159 | -12.5% | n/a | control (shared code) | control (shared code) |
| Validate duration (shared code, control) p(95) | ms | 0.018 | 0.020 | 0.018 | 0.020 | 0.002 | -11.1% | n/a | control (shared code) | control (shared code) |
| Validate duration (shared code, control) p(99) | ms | 0.026 | 0.027 | 0.026 | 0.027 | 0.001 | -3.8% | n/a | control (shared code) | control (shared code) |
| Validate duration (shared code, control) max | ms | 0.858 | 2.939 | 0.858 | 2.939 | 2.081 | -242.5% | n/a | control (shared code) | control (shared code) |
| Achieved throughput | req/s | 256.0 | 256.0 | 256.0 | 256.0 | 0.00556 | +0.0% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Server decoded_ok rate | op/s | 256.0 | 256.0 | 256.0 | 256.0 | -0.00217 | -0.0% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| CPU time per successful op | µs/op | 9846 | 5105 | 9846 | 5105 | -4740 | +48.1% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| CPU utilization | cores | 2.520 | 1.307 | 2.520 | 1.307 | -1.213 | +48.1% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| CPU-seconds in window | s | 453.8 | 235.3 | 453.8 | 235.3 | -218.5 | +48.1% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Σ handler durations | s | 428.7 | 207.8 | 428.7 | 207.8 | -221.0 | +51.5% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Σ decode durations | s | 320.3 | 99.8 | 320.3 | 99.8 | -220.5 | +68.8% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Allocated bytes per op | B/op | 4314970 | 4733136 | 4314970 | 4733136 | 418166 | -9.7% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Allocated objects per op | obj/op | 38091 | 33946 | 38091 | 33946 | -4144 | +10.9% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| GC cycles per 1k ops | cycles | 684.3 | 710.3 | 684.3 | 710.3 | 26.0 | -3.8% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| GC pause total in window | s | 1.247 | 0.803 | 1.247 | 0.803 | -0.444 | +35.6% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Live heap at window end | MiB | 6.912 | 5.668 | 6.912 | 5.668 | -1.244 | +18.0% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Runtime-mapped memory at end | MiB | 41.8 | 45.7 | 41.8 | 45.7 | 3.934 | -9.4% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Peak RSS (ps, 5 s samples) | MiB | 47.5 | 53.8 | 47.5 | 53.8 | 6.312 | -13.3% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Max RSS (getrusage) | MiB | 47.5 | 53.8 | 47.5 | 53.8 | 6.312 | -13.3% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |

<details><summary>[M] stdlib rep 1 allocs profile: 15362 decoded ops in the window</summary>

```
File: stdlib-server
Type: alloc_space
Time: 2026-11-16 19:15:58 -03
Duration: 60.09s, Total samples = 63403.70MB 
Showing nodes accounting for 63305.09MB, 99.84% of 63403.70MB total
Dropped 68 nodes (cum <= 317.02MB)
Showing top 20 nodes out of 30
      flat  flat%   sum%        cum   cum%
22402.41MB 35.33% 35.33% 22402.41MB 35.33%  reflect.growslice
15792.34MB 24.91% 60.24% 15792.34MB 24.91%  reflect.mapassign_faststr0
15550.37MB 24.53% 84.77% 15550.37MB 24.53%  github.com/lurodrisilva/golang-json-encoding-benchmark/service/internal/server.readBody
 3460.57MB  5.46% 90.22%  3460.57MB  5.46%  encoding/json/v2.makeString
 2617.12MB  4.13% 94.35%  2617.12MB  4.13%  reflect.makemap
 2507.04MB  3.95% 98.31%  2507.04MB  3.95%  reflect.unsafe_New
  928.53MB  1.46% 99.77%   928.53MB  1.46%  slices.Grow[go.shape.[]uint8,go.shape.uint8] (inline)
   46.21MB 0.073% 99.84%  4435.31MB  7.00%  encoding/json/v2.makeStringArshaler.func2
    0.50MB 0.00079% 99.84% 63342.11MB 99.90%  github.com/lurodrisilva/golang-json-encoding-benchmark/service/internal/server.(*ingestHandler).ServeHTTP
         0     0% 99.84% 47763.23MB 75.33%  encoding/json.Unmarshal (inline)
         0     0% 99.84%   928.53MB  1.46%  encoding/json/internal/jsonwire.AppendUnquote
         0     0% 99.84%   928.53MB  1.46%  encoding/json/internal/jsonwire.UnquoteMayCopy
         0     0% 99.84% 47763.23MB 75.33%  encoding/json/v2.Unmarshal
         0     0% 99.84% 20773.99MB 32.76%  encoding/json/v2.makeMapArshaler.func3
         0     0% 99.84%   758.01MB  1.20%  encoding/json/v2.makePointerArshaler.func3
         0     0% 99.84% 47704.01MB 75.24%  encoding/json/v2.makeSliceArshaler.func3
         0     0% 99.84% 47754.22MB 75.32%  encoding/json/v2.makeStructArshaler.func3
         0     0% 99.84%  1768.53MB  2.79%  encoding/json/v2.newAddressableValue
         0     0% 99.84% 47754.22MB 75.32%  encoding/json/v2.unmarshalDecode
         0     0% 99.84% 47763.23MB 75.33%  github.com/lurodrisilva/golang-json-encoding-benchmark/service/internal/codec/stdlib.Codec.Unmarshal
```
</details>

<details><summary>[M] goccy rep 1 allocs profile: 15362 decoded ops in the window</summary>

```
File: goccy-server
Type: alloc_space
Time: 2026-11-16 19:09:58 -03
Duration: 60s, Total samples = 69217.23MB 
Showing nodes accounting for 68761.96MB, 99.34% of 69217.23MB total
Dropped 61 nodes (cum <= 346.09MB)
Showing top 20 nodes out of 35
      flat  flat%   sum%        cum   cum%
18805.17MB 27.17% 27.17% 18805.17MB 27.17%  reflect.growslice
15901.37MB 22.97% 50.14% 15901.37MB 22.97%  reflect.mapassign_faststr0
15525.95MB 22.43% 72.57% 15525.95MB 22.43%  github.com/lurodrisilva/golang-json-encoding-benchmark/service/internal/server.readBody
 9876.24MB 14.27% 86.84%  9876.24MB 14.27%  github.com/goccy/go-json/internal/decoder.(*RuntimeContext).SetInput (inline)
 5013.59MB  7.24% 94.08%  5013.59MB  7.24%  github.com/goccy/go-json/internal/decoder.(*RuntimeContext).makeString (inline)
 2590.12MB  3.74% 97.83%  2590.12MB  3.74%  reflect.makemap
  659.02MB  0.95% 98.78%   659.02MB  0.95%  github.com/goccy/go-json/internal/decoder.(*RuntimeContext).unescapeString
  387.01MB  0.56% 99.34%   387.01MB  0.56%  reflect.unsafe_New
       3MB 0.0043% 99.34%    69146MB 99.90%  github.com/lurodrisilva/golang-json-encoding-benchmark/service/internal/server.(*ingestHandler).ServeHTTP
    0.50MB 0.00072% 99.34% 69215.73MB   100%  net/http.(*conn).serve
         0     0% 99.34% 53593.03MB 77.43%  github.com/goccy/go-json.Unmarshal (inline)
         0     0% 99.34% 53593.03MB 77.43%  github.com/goccy/go-json.unmarshal
         0     0% 99.34%   495.01MB  0.72%  github.com/goccy/go-json/internal/decoder.(*basicPtrDecoder).Decode
         0     0% 99.34% 20013.01MB 28.91%  github.com/goccy/go-json/internal/decoder.(*mapDecoder).Decode
         0     0% 99.34%   386.51MB  0.56%  github.com/goccy/go-json/internal/decoder.(*ptrDecoder).Decode
         0     0% 99.34% 43699.28MB 63.13%  github.com/goccy/go-json/internal/decoder.(*sliceDecoder).Decode
         0     0% 99.34% 13091.98MB 18.91%  github.com/goccy/go-json/internal/decoder.(*sliceDecoder).bufElem
         0     0% 99.34% 42223.64MB 61.00%  github.com/goccy/go-json/internal/decoder.(*sliceDecoder).decodeLong
         0     0% 99.34% 18805.17MB 27.17%  github.com/goccy/go-json/internal/decoder.(*sliceDecoder).grow
         0     0% 99.34% 13504.01MB 19.51%  github.com/goccy/go-json/internal/decoder.(*sliceDecoder).growBuf (inline)
```
</details>

### Pair D · f3-1m · fixed @256 rps · profile=cpu

- [M] trials analysed: stdlib-arm 1, goccy-arm 1.
- [M] window: nominal 180.0 s, wall 180.1 s, drain 0.106 s (per trial; the drain is k6 finishing in-flight requests).

| Metric | unit | stdlib (trials) | goccy (trials) | stdlib median | goccy median | Δ abs | reduction | 95% CI | verdict (pre-registered) | post-hoc verdict |
|---|---|---|---|---:|---:|---:|---:|---|---|---|
| HTTP request duration (k6) min | ms | 8.867 | 4.153 | 8.867 | 4.153 | -4.714 | +53.2% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| HTTP request duration (k6) med | ms | 9.270 | 4.601 | 9.270 | 4.601 | -4.669 | +50.4% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| HTTP request duration (k6) avg | ms | 9.326 | 4.597 | 9.326 | 4.597 | -4.728 | +50.7% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| HTTP request duration (k6) p(95) | ms | 9.631 | 4.851 | 9.631 | 4.851 | -4.780 | +49.6% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| HTTP request duration (k6) p(99) | ms | 10.6 | 5.035 | 10.6 | 5.035 | -5.576 | +52.6% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| HTTP request duration (k6) max | ms | 23.9 | 12.5 | 23.9 | 12.5 | -11.4 | +47.8% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Handler duration (server) min | ms | 8.749 | 3.972 | 8.749 | 3.972 | -4.777 | +54.6% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Handler duration (server) med | ms | 9.169 | 4.491 | 9.169 | 4.491 | -4.678 | +51.0% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Handler duration (server) avg | ms | 9.217 | 4.480 | 9.217 | 4.480 | -4.737 | +51.4% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Handler duration (server) p(95) | ms | 9.516 | 4.743 | 9.516 | 4.743 | -4.773 | +50.2% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Handler duration (server) p(99) | ms | 10.4 | 4.914 | 10.4 | 4.914 | -5.442 | +52.6% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Handler duration (server) max | ms | 23.8 | 12.3 | 23.8 | 12.3 | -11.4 | +48.1% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Decode duration (Unmarshal only) min | ms | 6.421 | 1.908 | 6.421 | 1.908 | -4.513 | +70.3% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Decode duration (Unmarshal only) med | ms | 6.842 | 2.158 | 6.842 | 2.158 | -4.684 | +68.5% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Decode duration (Unmarshal only) avg | ms | 6.876 | 2.147 | 6.876 | 2.147 | -4.729 | +68.8% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Decode duration (Unmarshal only) p(95) | ms | 7.146 | 2.370 | 7.146 | 2.370 | -4.776 | +66.8% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Decode duration (Unmarshal only) p(99) | ms | 7.640 | 2.520 | 7.640 | 2.520 | -5.120 | +67.0% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Decode duration (Unmarshal only) max | ms | 21.3 | 9.266 | 21.3 | 9.266 | -12.0 | +56.4% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Body read duration (shared code, control) min | ms | 0.056 | 0.054 | 0.056 | 0.054 | -0.002 | +3.6% | n/a | control (shared code) | control (shared code) |
| Body read duration (shared code, control) med | ms | 0.230 | 0.336 | 0.230 | 0.336 | 0.106 | -46.1% | n/a | control (shared code) (control differs ≥ threshold) | control (shared code) |
| Body read duration (shared code, control) avg | ms | 0.238 | 0.311 | 0.238 | 0.311 | 0.073 | -30.4% | n/a | control (shared code) | control (shared code) |
| Body read duration (shared code, control) p(95) | ms | 0.362 | 0.404 | 0.362 | 0.404 | 0.042 | -11.6% | n/a | control (shared code) | control (shared code) |
| Body read duration (shared code, control) p(99) | ms | 0.478 | 0.436 | 0.478 | 0.436 | -0.042 | +8.7% | n/a | control (shared code) | control (shared code) |
| Body read duration (shared code, control) max | ms | 4.844 | 5.589 | 4.844 | 5.589 | 0.745 | -15.4% | n/a | control (shared code) | control (shared code) |
| Policy scan duration (shared code, control) min | ms | 1.993 | 1.944 | 1.993 | 1.944 | -0.049 | +2.5% | n/a | control (shared code) | control (shared code) |
| Policy scan duration (shared code, control) med | ms | 2.047 | 1.992 | 2.047 | 1.992 | -0.055 | +2.7% | n/a | control (shared code) | control (shared code) |
| Policy scan duration (shared code, control) avg | ms | 2.093 | 2.011 | 2.093 | 2.011 | -0.082 | +3.9% | n/a | control (shared code) | control (shared code) |
| Policy scan duration (shared code, control) p(95) | ms | 2.222 | 2.100 | 2.222 | 2.100 | -0.122 | +5.5% | n/a | control (shared code) | control (shared code) |
| Policy scan duration (shared code, control) p(99) | ms | 2.451 | 2.171 | 2.451 | 2.171 | -0.280 | +11.4% | n/a | control (shared code) | control (shared code) |
| Policy scan duration (shared code, control) max | ms | 9.874 | 9.317 | 9.874 | 9.317 | -0.557 | +5.6% | n/a | control (shared code) | control (shared code) |
| Validate duration (shared code, control) min | ms | 0.007 | 0.007 | 0.007 | 0.007 | 0 | +0.0% | n/a | control (shared code) | control (shared code) |
| Validate duration (shared code, control) med | ms | 0.009 | 0.010 | 0.009 | 0.010 | 0.001 | -11.1% | n/a | control (shared code) (control differs ≥ threshold (within timing resolution: ignored post-hoc)) | control (shared code) |
| Validate duration (shared code, control) avg | ms | 0.010 | 0.011 | 0.010 | 0.011 | 0.000575 | -5.7% | n/a | control (shared code) | control (shared code) |
| Validate duration (shared code, control) p(95) | ms | 0.014 | 0.015 | 0.014 | 0.015 | 0.001 | -7.1% | n/a | control (shared code) | control (shared code) |
| Validate duration (shared code, control) p(99) | ms | 0.021 | 0.021 | 0.021 | 0.021 | 0 | +0.0% | n/a | control (shared code) | control (shared code) |
| Validate duration (shared code, control) max | ms | 0.207 | 0.252 | 0.207 | 0.252 | 0.045 | -21.7% | n/a | control (shared code) | control (shared code) |
| Achieved throughput | req/s | 256.0 | 256.0 | 256.0 | 256.0 | 0.00556 | +0.0% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Server decoded_ok rate | op/s | 256.0 | 256.0 | 256.0 | 256.0 | 0.00317 | +0.0% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| CPU time per successful op | µs/op | 9707 | 5026 | 9707 | 5026 | -4681 | +48.2% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| CPU utilization | cores | 2.484 | 1.286 | 2.484 | 1.286 | -1.198 | +48.2% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| CPU-seconds in window | s | 447.5 | 231.7 | 447.5 | 231.7 | -215.8 | +48.2% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Σ handler durations | s | 424.9 | 206.5 | 424.9 | 206.5 | -218.4 | +51.4% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Σ decode durations | s | 317.0 | 99.0 | 317.0 | 99.0 | -218.0 | +68.8% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Allocated bytes per op | B/op | 4314598 | 4708823 | 4314598 | 4708823 | 394225 | -9.1% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Allocated objects per op | obj/op | 38089 | 33945 | 38089 | 33945 | -4144 | +10.9% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| GC cycles per 1k ops | cycles | 645.3 | 676.5 | 645.3 | 676.5 | 31.2 | -4.8% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| GC pause total in window | s | 1.131 | 0.746 | 1.131 | 0.746 | -0.385 | +34.0% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Live heap at window end | MiB | 5.070 | 6.882 | 5.070 | 6.882 | 1.812 | -35.7% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Runtime-mapped memory at end | MiB | 45.5 | 45.5 | 45.5 | 45.5 | 0 | +0.0% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Peak RSS (ps, 5 s samples) | MiB | 49.8 | 49.7 | 49.8 | 49.7 | -0.062 | +0.1% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Max RSS (getrusage) | MiB | 49.8 | 49.7 | 49.8 | 49.7 | -0.062 | +0.1% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |

<details><summary>[M] stdlib rep 1 cpu profile: 15400 decoded ops in the window</summary>

```
File: stdlib-server
Type: cpu
Time: 2026-11-16 21:55:13 -03
Duration: 60.15s, Total samples = 123.56s (205.43%)
Showing nodes accounting for 108.69s, 87.97% of 123.56s total
Dropped 322 nodes (cum <= 0.62s)
Showing top 20 nodes out of 127
      flat  flat%   sum%        cum   cum%
    34.20s 27.68% 27.68%     34.20s 27.68%  runtime.madvise
    22.52s 18.23% 45.90%     22.52s 18.23%  syscall.rawsyscalln
    19.99s 16.18% 62.08%        20s 16.19%  runtime.usleep
    13.82s 11.18% 73.27%     13.82s 11.18%  runtime.pthread_kill
     4.02s  3.25% 76.52%      4.02s  3.25%  runtime.procyieldAsm
     2.22s  1.80% 78.32%      2.22s  1.80%  encoding/json/internal/jsonwire.ConsumeSimpleString (inline)
     1.75s  1.42% 79.73%      1.75s  1.42%  runtime.kevent
     1.50s  1.21% 80.95%      1.50s  1.21%  runtime.pthread_cond_wait
     1.12s  0.91% 81.85%      1.12s  0.91%  encoding/json/internal/jsonwire.ConsumeWhitespace (inline)
     1.07s  0.87% 82.72%      5.90s  4.78%  encoding/json/jsontext.(*decoderState).consumeValue
     1.01s  0.82% 83.54%      4.81s  3.89%  encoding/json/jsontext.(*decoderState).ReadValue
     0.79s  0.64% 84.18%      1.51s  1.22%  encoding/json/jsontext.(*decoderState).PeekKind
     0.74s   0.6% 84.78%      0.97s  0.79%  encoding/json/jsontext.(*decoderState).ReadToken
     0.73s  0.59% 85.37%     18.93s 15.32%  encoding/json/v2.makeStructArshaler.func3
     0.69s  0.56% 85.93%      0.80s  0.65%  github.com/lurodrisilva/golang-json-encoding-benchmark/service/internal/policy.(*scanner).str
     0.68s  0.55% 86.48%      0.68s  0.55%  runtime.pthread_cond_signal
     0.57s  0.46% 86.94%      0.94s  0.76%  internal/runtime/maps.(*Map).getWithoutKeySmallFastStr
     0.43s  0.35% 87.29%      1.26s  1.02%  encoding/json/v2.makeString
     0.42s  0.34% 87.63%      0.85s  0.69%  encoding/json/internal/jsonwire.ConsumeStringResumable
     0.42s  0.34% 87.97%      2.51s  2.03%  encoding/json/jsontext.(*decoderState).consumeObject
```
</details>

<details><summary>[M] goccy rep 1 cpu profile: 15403 decoded ops in the window</summary>

```
File: goccy-server
Type: cpu
Time: 2026-11-16 22:01:14 -03
Duration: 60.16s, Total samples = 69.95s (116.27%)
Showing nodes accounting for 64.44s, 92.12% of 69.95s total
Dropped 279 nodes (cum <= 0.35s)
Showing top 20 nodes out of 124
      flat  flat%   sum%        cum   cum%
    21.92s 31.34% 31.34%     21.92s 31.34%  syscall.rawsyscalln
    16.02s 22.90% 54.24%     16.02s 22.90%  runtime.madvise
     9.21s 13.17% 67.41%      9.21s 13.17%  runtime.kevent
     4.21s  6.02% 73.42%      4.21s  6.02%  runtime.usleep
     3.51s  5.02% 78.44%      3.51s  5.02%  runtime.pthread_kill
     2.62s  3.75% 82.19%      2.62s  3.75%  runtime.pthread_cond_signal
     2.59s  3.70% 85.89%      2.59s  3.70%  runtime.pthread_cond_wait
     1.09s  1.56% 87.45%      1.20s  1.72%  github.com/lurodrisilva/golang-json-encoding-benchmark/service/internal/policy.(*scanner).str
     0.55s  0.79% 88.23%      0.55s  0.79%  runtime.pthread_cond_timedwait_relative_np
     0.47s  0.67% 88.91%      1.81s  2.59%  github.com/goccy/go-json/internal/decoder.(*stringDecoder).decodeStringField
     0.40s  0.57% 89.48%      0.49s   0.7%  github.com/goccy/go-json/internal/decoder.(*structDecoder).decodeKey
     0.36s  0.51% 89.99%      0.52s  0.74%  runtime.mallocgcTinySC2
     0.31s  0.44% 90.44%      0.35s   0.5%  runtime.tryDeferToSpanScan
     0.25s  0.36% 90.79%      1.46s  2.09%  runtime.mallocgc
     0.20s  0.29% 91.08%      5.24s  7.49%  github.com/goccy/go-json/internal/decoder.(*structDecoder).Decode
     0.20s  0.29% 91.37%      0.43s  0.61%  runtime.pcvalue
     0.14s   0.2% 91.57%      1.93s  2.76%  github.com/lurodrisilva/golang-json-encoding-benchmark/service/internal/policy.(*scanner).object
     0.13s  0.19% 91.75%      1.94s  2.77%  github.com/goccy/go-json/internal/decoder.(*mapDecoder).Decode
     0.13s  0.19% 91.94%      1.93s  2.76%  github.com/lurodrisilva/golang-json-encoding-benchmark/service/internal/policy.(*scanner).value
     0.13s  0.19% 92.12%      0.93s  1.33%  runtime.slicebytetostring
```
</details>

### Pair L · f1-64k · fixed @1625 rps

- [M] trials analysed: stdlib-arm 5, goccy-arm 5.
- [M] goccy-legacy rep 5: INCLUDED but flagged [host busy 20.9% before trial (top: WindowServer 41%, Comet Helper (Aperitif Renderer) 32%, Comet Helper (Aperitif) 14%, Comet 13%, Comet Helper (Aperitif Renderer) 11%)]
- [M] stdlib-legacy rep 5: invalid [dropped_iterations=10 host busy 12.9% before trial (top: spotlightknowledged 100%)]
- [M] window: nominal 360.0 s, wall 360.3 s, drain 0.345 s (per trial; the drain is k6 finishing in-flight requests).

| Metric | unit | stdlib (trials) | goccy (trials) | stdlib median | goccy median | Δ abs | reduction | 95% CI | verdict (pre-registered) | post-hoc verdict |
|---|---|---|---|---:|---:|---:|---:|---|---|---|
| HTTP request duration (k6) min | ms | 0.814, 0.795, 0.796, 0.789, 0.790 | 0.285, 0.283, 0.280, 0.280, 0.271 | 0.795 | 0.280 | -0.515 | +64.8% | [+64.1, +65.9] | goccy better (CI excludes 0, ≥ threshold) | goccy better (CI excludes 0, ≥ threshold) |
| HTTP request duration (k6) med | ms | 0.871, 0.872, 0.871, 0.871, 0.870 | 0.345, 0.345, 0.345, 0.345, 0.346 | 0.871 | 0.345 | -0.526 | +60.4% | [+60.3, +60.4] | goccy better (CI excludes 0, ≥ threshold) | goccy better (CI excludes 0, ≥ threshold) |
| HTTP request duration (k6) avg | ms | 0.919, 0.953, 0.920, 0.920, 0.920 | 0.375, 0.375, 0.378, 0.374, 0.414 | 0.920 | 0.375 | -0.544 | +59.2% | [+55.0, +60.6] | goccy better (CI excludes 0, ≥ threshold) | goccy better (CI excludes 0, ≥ threshold) |
| HTTP request duration (k6) p(95) | ms | 1.223, 1.237, 1.227, 1.230, 1.223 | 0.462, 0.463, 0.465, 0.461, 0.512 | 1.227 | 0.463 | -0.764 | +62.3% | [+58.3, +62.6] | goccy better (CI excludes 0, ≥ threshold) | goccy better (CI excludes 0, ≥ threshold) |
| HTTP request duration (k6) p(99) | ms | 1.720, 1.763, 1.717, 1.743, 1.723 | 0.901, 0.910, 0.974, 0.868, 1.448 | 1.723 | 0.910 | -0.813 | +47.2% | [+16.0, +49.6] | inconclusive (trial CV > 10%) | inconclusive (trial CV > 10%) |
| HTTP request duration (k6) max | ms | 15.4, 169.4, 15.6, 16.3, 49.4 | 12.7, 9.581, 12.7, 11.0, 130.7 | 16.3 | 12.7 | -3.609 | +22.2% | [-703.6, +92.5] | inconclusive (trial CV > 10%) | inconclusive (trial CV > 10%) |
| Handler duration (server) min | ms | 0.756, 0.756, 0.756, 0.754, 0.756 | 0.235, 0.233, 0.235, 0.234, 0.235 | 0.756 | 0.235 | -0.521 | +68.9% | [+68.8, +69.2] | goccy better (CI excludes 0, ≥ threshold) | goccy better (CI excludes 0, ≥ threshold) |
| Handler duration (server) med | ms | 0.784, 0.785, 0.785, 0.784, 0.783 | 0.255, 0.256, 0.256, 0.256, 0.257 | 0.784 | 0.256 | -0.528 | +67.3% | [+67.2, +67.5] | goccy better (CI excludes 0, ≥ threshold) | goccy better (CI excludes 0, ≥ threshold) |
| Handler duration (server) avg | ms | 0.824, 0.830, 0.825, 0.824, 0.823 | 0.269, 0.269, 0.270, 0.269, 0.276 | 0.824 | 0.269 | -0.556 | +67.4% | [+66.5, +67.6] | goccy better (CI excludes 0, ≥ threshold) | goccy better (CI excludes 0, ≥ threshold) |
| Handler duration (server) p(95) | ms | 1.115, 1.121, 1.118, 1.121, 1.114 | 0.332, 0.332, 0.334, 0.332, 0.346 | 1.118 | 0.332 | -0.786 | +70.3% | [+69.1, +70.4] | goccy better (CI excludes 0, ≥ threshold) | goccy better (CI excludes 0, ≥ threshold) |
| Handler duration (server) p(99) | ms | 1.525, 1.525, 1.519, 1.530, 1.521 | 0.550, 0.546, 0.546, 0.546, 0.580 | 1.525 | 0.546 | -0.979 | +64.2% | [+62.0, +64.3] | goccy better (CI excludes 0, ≥ threshold) | goccy better (CI excludes 0, ≥ threshold) |
| Handler duration (server) max | ms | 12.2, 134.5, 14.1, 13.7, 15.4 | 9.374, 6.749, 11.2, 9.687, 54.9 | 14.1 | 9.687 | -4.396 | +31.2% | [-289.8, +92.8] | inconclusive (trial CV > 10%) | inconclusive (trial CV > 10%) |
| Decode duration (Unmarshal only) min | ms | 0.621, 0.620, 0.624, 0.621, 0.622 | 0.109, 0.108, 0.108, 0.108, 0.108 | 0.621 | 0.108 | -0.513 | +82.6% | [+82.4, +82.7] | goccy better (CI excludes 0, ≥ threshold) | goccy better (CI excludes 0, ≥ threshold) |
| Decode duration (Unmarshal only) med | ms | 0.650, 0.651, 0.650, 0.649, 0.648 | 0.121, 0.121, 0.121, 0.121, 0.122 | 0.650 | 0.121 | -0.529 | +81.4% | [+81.2, +81.4] | goccy better (CI excludes 0, ≥ threshold) | goccy better (CI excludes 0, ≥ threshold) |
| Decode duration (Unmarshal only) avg | ms | 0.683, 0.688, 0.684, 0.683, 0.681 | 0.130, 0.130, 0.130, 0.129, 0.133 | 0.683 | 0.130 | -0.554 | +81.0% | [+80.6, +81.2] | goccy better (CI excludes 0, ≥ threshold) | goccy better (CI excludes 0, ≥ threshold) |
| Decode duration (Unmarshal only) p(95) | ms | 0.962, 0.966, 0.965, 0.967, 0.961 | 0.184, 0.185, 0.187, 0.184, 0.189 | 0.965 | 0.185 | -0.780 | +80.8% | [+80.4, +81.0] | goccy better (CI excludes 0, ≥ threshold) | goccy better (CI excludes 0, ≥ threshold) |
| Decode duration (Unmarshal only) p(99) | ms | 1.246, 1.244, 1.241, 1.250, 1.243 | 0.271, 0.268, 0.269, 0.269, 0.284 | 1.244 | 0.269 | -0.975 | +78.4% | [+77.2, +78.5] | goccy better (CI excludes 0, ≥ threshold) | goccy better (CI excludes 0, ≥ threshold) |
| Decode duration (Unmarshal only) max | ms | 12.0, 133.2, 13.8, 9.742, 13.5 | 8.992, 6.391, 7.653, 8.849, 46.6 | 13.5 | 8.849 | -4.637 | +34.4% | [-245.5, +93.4] | inconclusive (trial CV > 10%) | inconclusive (trial CV > 10%) |
| Body read duration (shared code, control) min | ms | 0.005, 0.004, 0.005, 0.004, 0.005 | 0.005, 0.004, 0.004, 0.004, 0.005 | 0.005 | 0.004 | -0.001 | +20.0% | [-25.0, +20.0] | control (shared code) | control (shared code) |
| Body read duration (shared code, control) med | ms | 0.009, 0.009, 0.009, 0.009, 0.009 | 0.009, 0.009, 0.009, 0.009, 0.010 | 0.009 | 0.009 | 0 | +0.0% | [-11.1, +0.0] | control (shared code) | control (shared code) |
| Body read duration (shared code, control) avg | ms | 0.012, 0.013, 0.012, 0.012, 0.013 | 0.012, 0.012, 0.012, 0.012, 0.014 | 0.012 | 0.012 | -0.000613 | +4.9% | [-10.4, +8.0] | control (shared code) | control (shared code) |
| Body read duration (shared code, control) p(95) | ms | 0.026, 0.030, 0.029, 0.031, 0.030 | 0.015, 0.016, 0.015, 0.017, 0.019 | 0.030 | 0.016 | -0.014 | +46.7% | [+34.6, +50.0] | control (shared code) | control (shared code) |
| Body read duration (shared code, control) p(99) | ms | 0.071, 0.072, 0.072, 0.072, 0.073 | 0.050, 0.051, 0.051, 0.051, 0.060 | 0.072 | 0.051 | -0.021 | +29.2% | [+16.7, +30.6] | control (shared code) | control (shared code) |
| Body read duration (shared code, control) max | ms | 5.653, 105.7, 7.221, 6.321, 14.6 | 5.746, 5.054, 9.240, 9.020, 26.0 | 7.221 | 9.020 | 1.799 | -24.9% | [-260.3, +91.5] | control (shared code) | control (shared code) |
| Policy scan duration (shared code, control) min | ms | 0.120, 0.120, 0.119, 0.120, 0.119 | 0.116, 0.116, 0.116, 0.116, 0.116 | 0.120 | 0.116 | -0.004 | +3.3% | [+2.5, +3.3] | control (shared code) | control (shared code) |
| Policy scan duration (shared code, control) med | ms | 0.124, 0.124, 0.124, 0.125, 0.124 | 0.124, 0.124, 0.124, 0.124, 0.124 | 0.124 | 0.124 | 0 | +0.0% | [+0.0, +0.8] | control (shared code) | control (shared code) |
| Policy scan duration (shared code, control) avg | ms | 0.128, 0.128, 0.128, 0.129, 0.128 | 0.127, 0.127, 0.127, 0.127, 0.129 | 0.128 | 0.127 | -0.00139 | +1.1% | [-0.4, +1.3] | control (shared code) | control (shared code) |
| Policy scan duration (shared code, control) p(95) | ms | 0.133, 0.133, 0.133, 0.134, 0.133 | 0.129, 0.130, 0.130, 0.131, 0.132 | 0.133 | 0.130 | -0.003 | +2.3% | [+0.8, +3.0] | control (shared code) | control (shared code) |
| Policy scan duration (shared code, control) p(99) | ms | 0.258, 0.259, 0.258, 0.259, 0.258 | 0.245, 0.212, 0.204, 0.217, 0.251 | 0.258 | 0.217 | -0.041 | +15.9% | [+2.7, +20.9] | control (shared code) | control (shared code) |
| Policy scan duration (shared code, control) max | ms | 4.834, 23.8, 4.034, 4.002, 5.690 | 4.838, 4.767, 9.316, 5.101, 48.2 | 4.834 | 5.101 | 0.267 | -5.5% | [-896.3, +78.6] | control (shared code) | control (shared code) |
| Validate duration (shared code, control) min | ms | 0, 0, 0, 0, 0 | 0, 0, 0, 0, 0 | 0 | 0 | 0 | n/a | n/a | control (shared code) | control (shared code) |
| Validate duration (shared code, control) med | ms | 0.001, 0.001, 0.001, 0.001, 0.001 | 0.001, 0.001, 0.001, 0.001, 0.001 | 0.001 | 0.001 | 0 | +0.0% | [+0.0, +0.0] | control (shared code) | control (shared code) |
| Validate duration (shared code, control) avg | ms | 0.000628, 0.000615, 0.00066, 0.00063, 0.000703 | 0.000683, 0.000594, 0.000656, 0.000617, 0.000621 | 0.00063 | 0.000621 | -8.45e-06 | +1.3% | [-8.5, +11.6] | control (shared code) | control (shared code) |
| Validate duration (shared code, control) p(95) | ms | 0.001, 0.001, 0.001, 0.001, 0.001 | 0.001, 0.001, 0.001, 0.001, 0.001 | 0.001 | 0.001 | 0 | +0.0% | [+0.0, +0.0] | control (shared code) | control (shared code) |
| Validate duration (shared code, control) p(99) | ms | 0.001, 0.002, 0.001, 0.002, 0.001 | 0.001, 0.001, 0.001, 0.001, 0.001 | 0.001 | 0.001 | 0 | +0.0% | [+0.0, +50.0] | control (shared code) | control (shared code) |
| Validate duration (shared code, control) max | ms | 0.490, 1.148, 0.526, 0.420, 0.237 | 1.158, 0.440, 0.666, 0.340, 1.284 | 0.490 | 0.666 | 0.176 | -35.9% | [-205.7, +42.0] | control (shared code) | control (shared code) |
| Achieved throughput | req/s | 1625, 1625, 1625, 1625, 1625 | 1625, 1625, 1625, 1625, 1625 | 1625 | 1625 | 0 | +0.0% | [+0.0, +0.0] | no detectable difference | no detectable difference |
| Server decoded_ok rate | op/s | 1624, 1624, 1624, 1624, 1624 | 1624, 1624, 1624, 1624, 1624 | 1624 | 1624 | 0.106 | +0.0% | [-0.0, +0.0] | no detectable difference | no detectable difference |
| CPU time per successful op | µs/op | 912.1, 913.3, 913.1, 912.4, 910.6 | 371.7, 371.8, 371.9, 371.8, 381.3 | 912.4 | 371.8 | -540.5 | +59.2% | [+58.2, +59.3] | goccy better (CI excludes 0, ≥ threshold) | goccy better (CI excludes 0, ≥ threshold) |
| CPU utilization | cores | 1.481, 1.483, 1.483, 1.482, 1.479 | 0.604, 0.604, 0.604, 0.604, 0.619 | 1.482 | 0.604 | -0.878 | +59.2% | [+58.2, +59.3] | goccy better (CI excludes 0, ≥ threshold) | goccy better (CI excludes 0, ≥ threshold) |
| CPU-seconds in window | s | 533.8, 534.5, 534.4, 533.9, 532.9 | 217.5, 217.6, 217.6, 217.6, 223.2 | 533.9 | 217.6 | -316.3 | +59.2% | [+58.2, +59.3] | goccy better (CI excludes 0, ≥ threshold) | goccy better (CI excludes 0, ≥ threshold) |
| Σ handler durations | s | 482.3, 485.8, 482.8, 482.4, 481.4 | 157.2, 157.3, 157.8, 157.3, 161.6 | 482.4 | 157.3 | -325.1 | +67.4% | [+66.5, +67.6] | goccy better (CI excludes 0, ≥ threshold) | goccy better (CI excludes 0, ≥ threshold) |
| Σ decode durations | s | 399.8, 402.8, 400.1, 399.5, 398.7 | 75.8, 75.8, 76.2, 75.7, 77.8 | 399.8 | 75.8 | -324.0 | +81.0% | [+80.6, +81.2] | goccy better (CI excludes 0, ≥ threshold) | goccy better (CI excludes 0, ≥ threshold) |
| Allocated bytes per op | B/op | 289567, 289567, 289567, 289567, 289568 | 265913, 265911, 265900, 265921, 265866 | 289567 | 265911 | -23655 | +8.2% | [+8.2, +8.2] | detectable but below threshold | detectable but below threshold |
| Allocated objects per op | obj/op | 3160, 3160, 3160, 3160, 3160 | 2173, 2173, 2173, 2173, 2173 | 3160 | 2173 | -987.7 | +31.3% | [+31.3, +31.3] | goccy better (CI excludes 0, ≥ threshold) | goccy better (CI excludes 0, ≥ threshold) |
| GC cycles per 1k ops | cycles | 24.5, 24.6, 24.5, 24.6, 24.6 | 23.9, 23.9, 23.9, 23.9, 23.8 | 24.6 | 23.9 | -0.653 | +2.7% | [+2.4, +2.9] | detectable but below threshold | detectable but below threshold |
| GC pause total in window | s | 0.563, 0.630, 0.565, 0.572, 0.566 | 0.274, 0.269, 0.273, 0.268, 0.378 | 0.566 | 0.273 | -0.294 | +51.8% | [+33.3, +56.7] | inconclusive (trial CV > 10%) | inconclusive (trial CV > 10%) |
| Live heap at window end | MiB | 9.022, 9.056, 9.128, 9.063, 9.254 | 9.427, 9.266, 9.328, 9.327, 9.050 | 9.063 | 9.327 | 0.264 | -2.9% | [-4.0, +0.1] | no detectable difference | no detectable difference |
| Runtime-mapped memory at end | MiB | 63.2, 67.0, 67.2, 63.2, 66.9 | 63.2, 67.2, 67.0, 67.2, 71.0 | 66.9 | 67.2 | 0.316 | -0.5% | [-6.3, +5.5] | no detectable difference | no detectable difference |
| Peak RSS (ps, 5 s samples) | MiB | 71.0, 75.4, 71.0, 70.6, 71.2 | 72.3, 72.9, 71.1, 72.2, 78.8 | 71.0 | 72.3 | 1.266 | -1.8% | [-11.0, +4.2] | no detectable difference | no detectable difference |
| Max RSS (getrusage) | MiB | 71.0, 75.4, 71.0, 70.6, 71.2 | 72.3, 72.9, 71.1, 72.2, 78.8 | 71.0 | 72.3 | 1.266 | -1.8% | [-11.0, +4.2] | no detectable difference | no detectable difference |

### Pair L · f1-64k · fixed @1625 rps · profile=cpu

- [M] trials analysed: stdlib-arm 1, goccy-arm 1.
- [M] goccy-legacy rep 1: INCLUDED but flagged [host busy 20.3% before trial (top: Comet Helper (Aperitif Renderer) 96%, Comet 16%, WindowServer 15%, loginwindow 11%, WallpaperAerialsExtension 6%) profiling active: not a primary latency trial]
- [M] window: nominal 180.0 s, wall 180.3 s, drain 0.299 s (per trial; the drain is k6 finishing in-flight requests).

| Metric | unit | stdlib (trials) | goccy (trials) | stdlib median | goccy median | Δ abs | reduction | 95% CI | verdict (pre-registered) | post-hoc verdict |
|---|---|---|---|---:|---:|---:|---:|---|---|---|
| HTTP request duration (k6) min | ms | 0.812 | 0.295 | 0.812 | 0.295 | -0.517 | +63.7% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| HTTP request duration (k6) med | ms | 0.873 | 0.350 | 0.873 | 0.350 | -0.523 | +59.9% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| HTTP request duration (k6) avg | ms | 0.927 | 0.387 | 0.927 | 0.387 | -0.540 | +58.3% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| HTTP request duration (k6) p(95) | ms | 1.267 | 0.505 | 1.267 | 0.505 | -0.762 | +60.1% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| HTTP request duration (k6) p(99) | ms | 1.795 | 1.061 | 1.795 | 1.061 | -0.734 | +40.9% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| HTTP request duration (k6) max | ms | 14.2 | 8.553 | 14.2 | 8.553 | -5.624 | +39.7% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Handler duration (server) min | ms | 0.756 | 0.236 | 0.756 | 0.236 | -0.520 | +68.8% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Handler duration (server) med | ms | 0.786 | 0.258 | 0.786 | 0.258 | -0.528 | +67.2% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Handler duration (server) avg | ms | 0.827 | 0.272 | 0.827 | 0.272 | -0.554 | +67.1% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Handler duration (server) p(95) | ms | 1.161 | 0.337 | 1.161 | 0.337 | -0.824 | +71.0% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Handler duration (server) p(99) | ms | 1.540 | 0.566 | 1.540 | 0.566 | -0.974 | +63.2% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Handler duration (server) max | ms | 9.196 | 7.202 | 9.196 | 7.202 | -1.994 | +21.7% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Decode duration (Unmarshal only) min | ms | 0.625 | 0.110 | 0.625 | 0.110 | -0.515 | +82.4% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Decode duration (Unmarshal only) med | ms | 0.650 | 0.122 | 0.650 | 0.122 | -0.528 | +81.2% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Decode duration (Unmarshal only) avg | ms | 0.685 | 0.131 | 0.685 | 0.131 | -0.554 | +80.9% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Decode duration (Unmarshal only) p(95) | ms | 1.002 | 0.185 | 1.002 | 0.185 | -0.817 | +81.5% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Decode duration (Unmarshal only) p(99) | ms | 1.252 | 0.275 | 1.252 | 0.275 | -0.977 | +78.0% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Decode duration (Unmarshal only) max | ms | 8.029 | 5.820 | 8.029 | 5.820 | -2.209 | +27.5% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Body read duration (shared code, control) min | ms | 0.005 | 0.005 | 0.005 | 0.005 | 0 | +0.0% | n/a | control (shared code) | control (shared code) |
| Body read duration (shared code, control) med | ms | 0.010 | 0.010 | 0.010 | 0.010 | 0 | +0.0% | n/a | control (shared code) | control (shared code) |
| Body read duration (shared code, control) avg | ms | 0.013 | 0.013 | 0.013 | 0.013 | 0.000245 | -1.9% | n/a | control (shared code) | control (shared code) |
| Body read duration (shared code, control) p(95) | ms | 0.022 | 0.016 | 0.022 | 0.016 | -0.006 | +27.3% | n/a | control (shared code) | control (shared code) |
| Body read duration (shared code, control) p(99) | ms | 0.069 | 0.059 | 0.069 | 0.059 | -0.010 | +14.5% | n/a | control (shared code) | control (shared code) |
| Body read duration (shared code, control) max | ms | 6.210 | 5.558 | 6.210 | 5.558 | -0.652 | +10.5% | n/a | control (shared code) | control (shared code) |
| Policy scan duration (shared code, control) min | ms | 0.120 | 0.117 | 0.120 | 0.117 | -0.003 | +2.5% | n/a | control (shared code) | control (shared code) |
| Policy scan duration (shared code, control) med | ms | 0.124 | 0.125 | 0.124 | 0.125 | 0.001 | -0.8% | n/a | control (shared code) | control (shared code) |
| Policy scan duration (shared code, control) avg | ms | 0.129 | 0.128 | 0.129 | 0.128 | -0.0011 | +0.9% | n/a | control (shared code) | control (shared code) |
| Policy scan duration (shared code, control) p(95) | ms | 0.132 | 0.130 | 0.132 | 0.130 | -0.002 | +1.5% | n/a | control (shared code) | control (shared code) |
| Policy scan duration (shared code, control) p(99) | ms | 0.262 | 0.249 | 0.262 | 0.249 | -0.013 | +5.0% | n/a | control (shared code) | control (shared code) |
| Policy scan duration (shared code, control) max | ms | 3.475 | 4.773 | 3.475 | 4.773 | 1.298 | -37.4% | n/a | control (shared code) | control (shared code) |
| Validate duration (shared code, control) min | ms | 0 | 0 | 0 | 0 | 0 | n/a | n/a | control (shared code) | control (shared code) |
| Validate duration (shared code, control) med | ms | 0.001 | 0.001 | 0.001 | 0.001 | 0 | +0.0% | n/a | control (shared code) | control (shared code) |
| Validate duration (shared code, control) avg | ms | 0.000723 | 0.00072 | 0.000723 | 0.00072 | -2.77e-06 | +0.4% | n/a | control (shared code) | control (shared code) |
| Validate duration (shared code, control) p(95) | ms | 0.001 | 0.001 | 0.001 | 0.001 | 0 | +0.0% | n/a | control (shared code) | control (shared code) |
| Validate duration (shared code, control) p(99) | ms | 0.002 | 0.001 | 0.002 | 0.001 | -0.001 | +50.0% | n/a | control (shared code) | control (shared code) |
| Validate duration (shared code, control) max | ms | 1.306 | 0.411 | 1.306 | 0.411 | -0.895 | +68.5% | n/a | control (shared code) | control (shared code) |
| Achieved throughput | req/s | 1625 | 1625 | 1625 | 1625 | 0 | +0.0% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Server decoded_ok rate | op/s | 1624 | 1624 | 1624 | 1624 | 0.125 | +0.0% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| CPU time per successful op | µs/op | 917.1 | 378.9 | 917.1 | 378.9 | -538.2 | +58.7% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| CPU utilization | cores | 1.489 | 0.615 | 1.489 | 0.615 | -0.874 | +58.7% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| CPU-seconds in window | s | 268.5 | 110.9 | 268.5 | 110.9 | -157.5 | +58.7% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Σ handler durations | s | 242.1 | 79.7 | 242.1 | 79.7 | -162.3 | +67.1% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Σ decode durations | s | 200.4 | 38.4 | 200.4 | 38.4 | -162.1 | +80.9% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Allocated bytes per op | B/op | 289571 | 265500 | 289571 | 265500 | -24071 | +8.3% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Allocated objects per op | obj/op | 3160 | 2173 | 3160 | 2173 | -987.8 | +31.3% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| GC cycles per 1k ops | cycles | 19.5 | 19.1 | 19.5 | 19.1 | -0.398 | +2.0% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| GC pause total in window | s | 0.250 | 0.135 | 0.250 | 0.135 | -0.115 | +45.9% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Live heap at window end | MiB | 9.458 | 9.362 | 9.458 | 9.362 | -0.096 | +1.0% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Runtime-mapped memory at end | MiB | 71.2 | 67.2 | 71.2 | 67.2 | -4.000 | +5.6% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Peak RSS (ps, 5 s samples) | MiB | 75.0 | 74.8 | 75.0 | 74.8 | -0.109 | +0.1% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |
| Max RSS (getrusage) | MiB | 75.0 | 74.9 | 75.0 | 74.9 | -0.094 | +0.1% | n/a | insufficient trials (n<2) | insufficient trials (n<2) |

<details><summary>[M] stdlib-legacy rep 1 cpu profile: 97743 decoded ops in the window</summary>

```
File: stdlib-legacy-server
Type: cpu
Time: 2026-11-16 19:20:59 -03
Duration: 60.14s, Total samples = 85.35s (141.91%)
Showing nodes accounting for 70.10s, 82.13% of 85.35s total
Dropped 337 nodes (cum <= 0.43s)
Showing top 20 nodes out of 131
      flat  flat%   sum%        cum   cum%
    43.25s 50.67% 50.67%     43.25s 50.67%  syscall.rawsyscalln
     4.52s  5.30% 55.97%      4.52s  5.30%  runtime.kevent
     3.08s  3.61% 59.58%      3.49s  4.09%  encoding/json.unquoteBytes
     2.99s  3.50% 63.08%      3.75s  4.39%  encoding/json.(*decodeState).rescanLiteral
     1.82s  2.13% 65.21%     27.70s 32.45%  encoding/json.(*decodeState).object
     1.74s  2.04% 67.25%      1.74s  2.04%  runtime.pthread_kill
     1.28s  1.50% 68.75%      1.28s  1.50%  runtime.usleep
     1.23s  1.44% 70.19%      3.46s  4.05%  encoding/json.checkValid
     1.20s  1.41% 71.60%      1.88s  2.20%  reflect.(*rtype).Name
     1.15s  1.35% 72.95%      5.35s  6.27%  encoding/json.indirect
     1.15s  1.35% 74.29%      1.49s  1.75%  encoding/json.stateEndValue
     1.15s  1.35% 75.64%      1.15s  1.35%  runtime.pthread_cond_wait
     1.05s  1.23% 76.87%      1.09s  1.28%  encoding/json.stateInString
     0.90s  1.05% 77.93%      1.10s  1.29%  encoding/json.stateBeginValue
     0.78s  0.91% 78.84%      1.84s  2.16%  encoding/json.(*decodeState).scanWhile
     0.73s  0.86% 79.70%      0.73s  0.86%  internal/strconv.readFloat
     0.60s   0.7% 80.40%     10.11s 11.85%  encoding/json.(*decodeState).literalStore
     0.53s  0.62% 81.02%      0.53s  0.62%  runtime.memmove
     0.48s  0.56% 81.58%     27.70s 32.45%  encoding/json.(*decodeState).value
     0.47s  0.55% 82.13%      0.47s  0.55%  internal/runtime/maps.memHashAES
```
</details>

<details><summary>[M] goccy-legacy rep 1 cpu profile: 97732 decoded ops in the window</summary>

```
File: goccy-legacy-server
Type: cpu
Time: 2026-11-16 19:27:00 -03
Duration: 60.14s, Total samples = 33.77s (56.15%)
Showing nodes accounting for 31.77s, 94.08% of 33.77s total
Dropped 198 nodes (cum <= 0.17s)
Showing top 20 nodes out of 104
      flat  flat%   sum%        cum   cum%
    23.69s 70.15% 70.15%     23.69s 70.15%  syscall.rawsyscalln
     3.25s  9.62% 79.77%      3.25s  9.62%  runtime.kevent
     1.06s  3.14% 82.91%      1.06s  3.14%  runtime.pthread_cond_wait
     0.93s  2.75% 85.67%      0.93s  2.75%  runtime.pthread_cond_signal
     0.87s  2.58% 88.24%      0.87s  2.58%  runtime.pthread_kill
     0.78s  2.31% 90.55%      0.78s  2.31%  runtime.usleep
     0.36s  1.07% 91.62%      0.39s  1.15%  github.com/lurodrisilva/golang-json-encoding-benchmark/service/internal/policy.(*scanner).str
     0.18s  0.53% 92.15%      0.23s  0.68%  github.com/goccy/go-json/internal/decoder.(*structDecoder).decodeKey
     0.18s  0.53% 92.69%      0.32s  0.95%  runtime.pcvalue
     0.11s  0.33% 93.01%      0.24s  0.71%  runtime.scanObjectsSmall
     0.08s  0.24% 93.25%      0.74s  2.19%  github.com/lurodrisilva/golang-json-encoding-benchmark/service/internal/policy.(*scanner).object
     0.05s  0.15% 93.40%      0.19s  0.56%  github.com/goccy/go-json/internal/decoder.(*stringDecoder).decodeStringField
     0.05s  0.15% 93.54%      0.24s  0.71%  runtime.(*unwinder).resolveInternal
     0.04s  0.12% 93.66%      0.98s  2.90%  github.com/goccy/go-json/internal/decoder.(*structDecoder).Decode
     0.03s 0.089% 93.75%      0.74s  2.19%  github.com/lurodrisilva/golang-json-encoding-benchmark/service/internal/policy.(*scanner).value
     0.03s 0.089% 93.84%      0.21s  0.62%  runtime.(*stkframe).getStackMap
     0.02s 0.059% 93.90%      0.26s  0.77%  github.com/goccy/go-json/internal/decoder.(*mapDecoder).Decode
     0.02s 0.059% 93.96%      0.98s  2.90%  github.com/goccy/go-json/internal/decoder.(*sliceDecoder).Decode
     0.02s 0.059% 94.02%      0.22s  0.65%  github.com/goccy/go-json/internal/decoder.(*stringDecoder).Decode
     0.02s 0.059% 94.08%      0.31s  0.92%  runtime.(*unwinder).next
```
</details>

## 5. Saturation (stepped offered load)

### Pair D · f1-64k · ramp knee-ref 5000

**goccy rep 1**: knee = 6000 rps (step 8) [A]

| step | offered rps | achieved rps | dropped | non-202 | P50 ms | P95 ms | P99 ms | decode P50 ms | samples |
|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| 0 | 500.0 | 500.0 | 0 | 0 | 0.951 | 1.631 | 1.935 | 0.365 | 60001 |
| 1 | 1000 | 1000 | 0 | 0 | 0.578 | 0.768 | 1.055 | 0.234 | 120001 |
| 2 | 1750 | 1750 | 0 | 0 | 0.339 | 0.460 | 0.993 | 0.119 | 210001 |
| 3 | 2500 | 2500 | 0 | 0 | 0.342 | 0.462 | 0.884 | 0.122 | 300001 |
| 4 | 3250 | 3250 | 0 | 0 | 0.328 | 0.510 | 1.080 | 0.126 | 390001 |
| 5 | 4000 | 4000 | 0 | 0 | 0.332 | 0.668 | 1.698 | 0.132 | 480000 |
| 6 | 4500 | 4500 | 0 | 0 | 0.305 | 0.650 | 1.306 | 0.132 | 540001 |
| 7 | 5000 | 5000 | 0 | 0 | 0.306 | 0.757 | 2.097 | 0.131 | 600001 |
| 8 | 6000 | 6000 | 0 | 0 | 0.322 | 1.181 | 4.333 | 0.132 | 720004 |
| 9 | 7500 | 7496 | 533.0 | 0 | 0.481 | 26.1 | 91.4 | 0.138 | 899470 |
| 10 | 8750 | 8320 | 51608 | 0 | 160.9 | 469.0 | 1081 | 0.180 | 998399 |
| 11 | 10000 | 7978 | 242647 | 0 | 238.3 | 367.5 | 944.8 | 0.172 | 957352 |

**stdlib rep 2**: knee = 2500 rps (step 3) [A]

| step | offered rps | achieved rps | dropped | non-202 | P50 ms | P95 ms | P99 ms | decode P50 ms | samples |
|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| 0 | 500.0 | 500.0 | 0 | 0 | 1.344 | 1.583 | 1.793 | 0.914 | 60001 |
| 1 | 1000 | 1000 | 0 | 0 | 0.663 | 0.805 | 1.337 | 0.436 | 120000 |
| 2 | 1750 | 1750 | 0 | 0 | 0.608 | 0.792 | 1.399 | 0.397 | 210001 |
| 3 | 2500 | 2500 | 0 | 0 | 0.615 | 0.878 | 1.862 | 0.414 | 300001 |
| 4 | 3250 | 3249 | 82.0 | 0 | 0.614 | 2.219 | 33.4 | 0.412 | 389919 |
| 5 | 4000 | 4000 | 0 | 0 | 0.605 | 2.252 | 10.6 | 0.412 | 480001 |
| 6 | 4500 | 4500 | 0 | 0 | 0.630 | 5.633 | 20.5 | 0.411 | 540001 |
| 7 | 5000 | 5000 | 0 | 0 | 0.908 | 31.7 | 68.4 | 0.416 | 600003 |
| 8 | 6000 | 5422 | 69358 | 0 | 137.0 | 591.9 | 1291 | 0.429 | 650645 |
| 9 | 7500 | 5297 | 264318 | 0 | 258.7 | 568.8 | 1570 | 0.428 | 635685 |
| 10 | 8750 | 5219 | 423734 | 239.0 | 320.7 | 516.3 | 1434 | 0.429 | 626274 |
| 11 | 10000 | 5162 | 580516 | 700.0 | 373.6 | 613.6 | 1384 | 0.429 | 619485 |

- [M] stdlib rep 1: invalid [swapouts during trial=8005]
- [L] Each step's percentiles come from that step's samples only. The knee follows plan §14.4: the highest step with ≥99% achieved, no drops, no non-202, and P95 ≤ 3× the first step's P95 (P95 replaces the plan's P99 criterion because low steps of large fixtures cannot reach the 10,000-sample P99 minimum; each step needs ≥200 samples). Step P99 is shown only for steps with ≥10000 samples. If any step's k6 CPU limit was hit, the run is marked invalid instead of reading as a knee.

### Pair D · f2-256k · ramp knee-ref 1250

**goccy rep 1**: knee = 1000 rps (step 5) [A]

| step | offered rps | achieved rps | dropped | non-202 | P50 ms | P95 ms | P99 ms | decode P50 ms | samples |
|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| 0 | 125.0 | 125.0 | 0 | 0 | 2.752 | 3.288 | 3.514 | 1.219 | 15000 |
| 1 | 250.0 | 250.0 | 0 | 0 | 1.879 | 2.332 | 2.709 | 0.830 | 30001 |
| 2 | 437.5 | 438.0 | 0 | 0 | 1.565 | 1.822 | 2.221 | 0.686 | 52561 |
| 3 | 625.0 | 625.0 | 0 | 0 | 1.151 | 1.339 | 1.922 | 0.491 | 75001 |
| 4 | 812.5 | 813.0 | 0 | 0 | 1.132 | 1.361 | 2.631 | 0.484 | 97561 |
| 5 | 1000 | 1000 | 0 | 0 | 1.183 | 1.392 | 2.673 | 0.493 | 120001 |
| 6 | 1125 | 1124 | 140.0 | 0 | 1.181 | 1.413 | 3.134 | 0.495 | 134861 |
| 7 | 1250 | 1250 | 0 | 0 | 1.200 | 1.527 | 4.210 | 0.499 | 150001 |
| 8 | 1500 | 1492 | 947.0 | 0 | 1.211 | 3.280 | 45.4 | 0.507 | 179054 |
| 9 | 1875 | 1875 | 0 | 0 | 1.216 | 3.450 | 15.6 | 0.512 | 225003 |
| 10 | 2188 | 2187 | 79.0 | 0 | 1.251 | 6.198 | 21.8 | 0.518 | 262482 |
| 11 | 2500 | 2493 | 838.0 | 0 | 1.358 | 20.1 | 76.5 | 0.523 | 299162 |

**stdlib rep 2**: knee = 1250 rps (step 7) [A]

| step | offered rps | achieved rps | dropped | non-202 | P50 ms | P95 ms | P99 ms | decode P50 ms | samples |
|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| 0 | 125.0 | 125.0 | 0 | 0 | 3.359 | 3.544 | 3.741 | 2.400 | 15001 |
| 1 | 250.0 | 250.0 | 0 | 0 | 2.521 | 2.701 | 2.783 | 1.781 | 30000 |
| 2 | 437.5 | 438.0 | 0 | 0 | 2.208 | 2.419 | 2.577 | 1.550 | 52561 |
| 3 | 625.0 | 625.0 | 0 | 0 | 2.343 | 2.510 | 3.099 | 1.598 | 75000 |
| 4 | 812.5 | 813.0 | 0 | 0 | 2.348 | 2.515 | 4.166 | 1.614 | 97561 |
| 5 | 1000 | 1000 | 0 | 0 | 2.393 | 2.658 | 5.632 | 1.713 | 120000 |
| 6 | 1125 | 1125 | 0 | 0 | 2.433 | 3.243 | 7.349 | 1.771 | 135001 |
| 7 | 1250 | 1250 | 0 | 0 | 2.473 | 5.070 | 11.9 | 1.784 | 150001 |
| 8 | 1500 | 1500 | 0 | 0 | 2.617 | 17.0 | 32.8 | 1.845 | 180001 |
| 9 | 1875 | 1675 | 23998 | 0 | 218.5 | 517.7 | 737.6 | 1.651 | 201003 |
| 10 | 2188 | 1661 | 63264 | 0 | 243.6 | 563.1 | 786.9 | 1.654 | 199297 |
| 11 | 2500 | 1657 | 101164 | 4.000 | 280.6 | 636.8 | 881.4 | 1.654 | 198837 |

- [M] goccy rep 1: INCLUDED but flagged [host busy 11.7% before trial (top: mediaanalysisd 28%, Code Helper 21%, WindowServer 15%, Microsoft Outlook 7%, Code Helper (Renderer) 5%)]
- [M] stdlib rep 1: invalid [swapouts during trial=18402 host busy 15.0% before trial (top: corespotlightd 109%)]
- [L] Each step's percentiles come from that step's samples only. The knee follows plan §14.4: the highest step with ≥99% achieved, no drops, no non-202, and P95 ≤ 3× the first step's P95 (P95 replaces the plan's P99 criterion because low steps of large fixtures cannot reach the 10,000-sample P99 minimum; each step needs ≥200 samples). Step P99 is shown only for steps with ≥10000 samples. If any step's k6 CPU limit was hit, the run is marked invalid instead of reading as a knee.

### Pair D · f3-1m · ramp knee-ref 320

**goccy rep 1**: knee = 480 rps (step 9) [A]

| step | offered rps | achieved rps | dropped | non-202 | P50 ms | P95 ms | P99 ms | decode P50 ms | samples |
|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| 0 | 32.0 | 32.0 | 0 | 0 | 8.936 | 9.585 | n/a | 4.044 | 3841 |
| 1 | 64.0 | 64.0 | 0 | 0 | 6.111 | 6.453 | n/a | 2.809 | 7681 |
| 2 | 112.0 | 112.0 | 0 | 0 | 4.783 | 5.036 | 5.406 | 2.204 | 13441 |
| 3 | 160.0 | 160.0 | 0 | 0 | 4.387 | 4.608 | 5.059 | 2.026 | 19201 |
| 4 | 208.0 | 208.0 | 0 | 0 | 4.474 | 4.722 | 4.962 | 2.089 | 24961 |
| 5 | 256.0 | 256.0 | 0 | 0 | 4.627 | 4.882 | 5.452 | 2.178 | 30721 |
| 6 | 288.0 | 288.0 | 0 | 0 | 4.622 | 4.907 | 5.436 | 2.185 | 34561 |
| 7 | 320.0 | 320.0 | 0 | 0 | 4.662 | 4.911 | 6.590 | 2.192 | 38401 |
| 8 | 384.0 | 384.0 | 0 | 0 | 4.644 | 4.921 | 6.100 | 2.199 | 46081 |
| 9 | 480.0 | 480.0 | 0 | 0 | 4.684 | 5.183 | 7.699 | 2.260 | 57601 |
| 10 | 560.0 | 559.8 | 22.0 | 0 | 4.818 | 14.1 | 156.4 | 2.382 | 67179 |
| 11 | 640.0 | 640.0 | 0 | 0 | 4.893 | 7.056 | 20.8 | 2.436 | 76801 |

**stdlib rep 1**: knee = 320 rps (step 7) [A]

| step | offered rps | achieved rps | dropped | non-202 | P50 ms | P95 ms | P99 ms | decode P50 ms | samples |
|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| 0 | 32.0 | 32.0 | 0 | 0 | 11.9 | 12.3 | n/a | 8.587 | 3841 |
| 1 | 64.0 | 64.0 | 0 | 0 | 8.749 | 8.948 | n/a | 6.262 | 7681 |
| 2 | 112.0 | 112.0 | 0 | 0 | 8.708 | 9.337 | 10.1 | 6.254 | 13441 |
| 3 | 160.0 | 160.0 | 0 | 0 | 9.129 | 9.459 | 10.7 | 6.636 | 19201 |
| 4 | 208.0 | 208.0 | 0 | 0 | 9.200 | 9.636 | 12.4 | 6.712 | 24961 |
| 5 | 256.0 | 256.0 | 0 | 0 | 9.379 | 10.1 | 14.1 | 6.909 | 30721 |
| 6 | 288.0 | 288.0 | 0 | 0 | 9.441 | 10.2 | 21.6 | 6.950 | 34561 |
| 7 | 320.0 | 320.0 | 0 | 0 | 9.506 | 10.7 | 19.0 | 7.020 | 38400 |
| 8 | 384.0 | 383.7 | 35.0 | 0 | 9.995 | 40.0 | 227.9 | 7.292 | 46045 |
| 9 | 480.0 | 421.3 | 7048 | 0 | 776.4 | 1999 | 2844 | 6.674 | 50552 |
| 10 | 560.0 | 417.7 | 17074 | 0 | 794.6 | 2087 | 3107 | 6.680 | 50127 |
| 11 | 640.0 | 418.9 | 26530 | 1.000 | 802.7 | 2047 | 2893 | 6.681 | 50271 |

- [L] Each step's percentiles come from that step's samples only. The knee follows plan §14.4: the highest step with ≥99% achieved, no drops, no non-202, and P95 ≤ 3× the first step's P95 (P95 replaces the plan's P99 criterion because low steps of large fixtures cannot reach the 10,000-sample P99 minimum; each step needs ≥200 samples). Step P99 is shown only for steps with ≥10000 samples. If any step's k6 CPU limit was hit, the run is marked invalid instead of reading as a knee.

### Pair L · f1-64k · ramp knee-ref 5000

**goccy-legacy rep 1**: knee = 6000 rps (step 8) [A]

| step | offered rps | achieved rps | dropped | non-202 | P50 ms | P95 ms | P99 ms | decode P50 ms | samples |
|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| 0 | 500.0 | 500.0 | 0 | 0 | 1.232 | 1.783 | 2.157 | 0.473 | 60000 |
| 1 | 1000 | 1000 | 0 | 0 | 0.581 | 0.823 | 1.269 | 0.231 | 120001 |
| 2 | 1750 | 1750 | 0 | 0 | 0.332 | 0.456 | 0.875 | 0.117 | 210001 |
| 3 | 2500 | 2500 | 0 | 0 | 0.337 | 0.449 | 0.861 | 0.121 | 300001 |
| 4 | 3250 | 3250 | 0 | 0 | 0.325 | 0.496 | 0.967 | 0.125 | 390001 |
| 5 | 4000 | 4000 | 0 | 0 | 0.330 | 0.608 | 1.405 | 0.130 | 480001 |
| 6 | 4500 | 4500 | 0 | 0 | 0.302 | 0.619 | 1.163 | 0.130 | 540003 |
| 7 | 5000 | 5000 | 0 | 0 | 0.303 | 0.691 | 1.456 | 0.130 | 600000 |
| 8 | 6000 | 6000 | 0 | 0 | 0.323 | 1.415 | 7.612 | 0.131 | 720003 |
| 9 | 7500 | 7500 | 0 | 0 | 0.458 | 17.6 | 63.7 | 0.135 | 900004 |
| 10 | 8750 | 8471 | 33464 | 0 | 112.4 | 486.8 | 1031 | 0.177 | 1016544 |
| 11 | 10000 | 8086 | 229716 | 0 | 233.0 | 437.5 | 1010 | 0.166 | 970285 |

**stdlib-legacy rep 1**: knee = 2500 rps (step 3) [A]

| step | offered rps | achieved rps | dropped | non-202 | P50 ms | P95 ms | P99 ms | decode P50 ms | samples |
|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| 0 | 500.0 | 500.0 | 0 | 0 | 1.590 | 1.815 | 1.905 | 1.226 | 60001 |
| 1 | 1000 | 1000 | 0 | 0 | 0.843 | 1.027 | 1.161 | 0.625 | 120001 |
| 2 | 1750 | 1750 | 0 | 0 | 0.850 | 1.121 | 2.182 | 0.646 | 210001 |
| 3 | 2500 | 2500 | 0 | 0 | 0.843 | 1.472 | 4.389 | 0.646 | 300000 |
| 4 | 3250 | 3250 | 0 | 0 | 0.866 | 6.206 | 20.2 | 0.655 | 390001 |
| 5 | 4000 | 4000 | 0 | 0 | 1.777 | 40.0 | 84.0 | 0.652 | 480001 |
| 6 | 4500 | 4272 | 27401 | 0 | 114.7 | 650.6 | 1379 | 0.655 | 512600 |
| 7 | 5000 | 4202 | 95806 | 0 | 139.2 | 792.2 | 1557 | 0.656 | 504197 |
| 8 | 6000 | 3997 | 240393 | 0 | 238.3 | 831.8 | 1814 | 0.658 | 479611 |
| 9 | 7500 | 3978 | 422592 | 0 | 285.2 | 1265 | 2107 | 0.658 | 477411 |
| 10 | 8750 | 4080 | 560436 | 307.0 | 275.1 | 1577 | 2708 | 0.657 | 489571 |
| 11 | 10000 | 4081 | 710339 | 136.0 | 424.0 | 1395 | 2773 | 0.656 | 489662 |

- [M] stdlib-legacy rep 1: INCLUDED but flagged [host busy 14.7% before trial (top: spotlightknowledged 98%)]
- [L] Each step's percentiles come from that step's samples only. The knee follows plan §14.4: the highest step with ≥99% achieved, no drops, no non-202, and P95 ≤ 3× the first step's P95 (P95 replaces the plan's P99 criterion because low steps of large fixtures cannot reach the 10,000-sample P99 minimum; each step needs ≥200 samples). Step P99 is shown only for steps with ≥10000 samples. If any step's k6 CPU limit was hit, the run is marked invalid instead of reading as a knee.

## 6. Decoder microbenchmarks (isolated, not HTTP)

[M] benchstat output (`go test -bench` with -count≥10). These conclusions are **separate** from the HTTP conclusions:

```
goos: darwin
goarch: arm64
pkg: github.com/lurodrisilva/golang-json-encoding-benchmark/service/bench
cpu: Apple M3
                            │    stdlib    │                goccy                │                    │
                            │    sec/op    │   sec/op     vs base                │   sec/op     vs base   │
Unmarshal/fixture=f1-64k-8     365.9µ ± 2%   117.5µ ± 2%  -67.89% (p=0.000 n=10)
Unmarshal/fixture=f2-256k-8   1519.0µ ± 1%   501.2µ ± 0%  -67.00% (p=0.000 n=10)
Unmarshal/fixture=f3-1m-8      6.330m ± 0%   2.008m ± 1%  -68.27% (p=0.000 n=10)
Policy/fixture=f1-64k-8                                                            117.6µ ± 0%
Policy/fixture=f2-256k-8                                                           478.7µ ± 0%
Policy/fixture=f3-1m-8                                                             1.934m ± 1%
geomean                        1.521m        490.9µ       -67.73%                  477.5µ       ? ¹ ²
¹ benchmark set differs from baseline; geomeans may not be comparable
² ratios must be >0 to compute geomean

                            │    stdlib    │                 goccy                 │                     │
                            │     B/s      │     B/s       vs base                 │     B/s       vs base   │
Unmarshal/fixture=f1-64k-8    171.3Mi ± 2%   533.6Mi ± 2%  +211.46% (p=0.000 n=10)
Unmarshal/fixture=f2-256k-8   164.6Mi ± 1%   498.8Mi ± 0%  +203.06% (p=0.000 n=10)
Unmarshal/fixture=f3-1m-8     158.0Mi ± 0%   498.1Mi ± 1%  +215.16% (p=0.000 n=10)
Policy/fixture=f1-64k-8                                                              533.1Mi ± 0%
Policy/fixture=f2-256k-8                                                             522.3Mi ± 0%
Policy/fixture=f3-1m-8                                                               517.3Mi ± 1%
geomean                       164.6Mi        509.9Mi       +209.85%                  524.2Mi       ? ¹ ²
¹ benchmark set differs from baseline; geomeans may not be comparable
² ratios must be >0 to compute geomean

                            │    stdlib    │                goccy                │                   │
                            │     B/op     │     B/op      vs base               │    B/op     vs base   │
Unmarshal/fixture=f1-64k-8    192.6Ki ± 0%   182.9Ki ± 0%  -5.04% (p=0.000 n=10)
Unmarshal/fixture=f2-256k-8   759.9Ki ± 0%   718.7Ki ± 0%  -5.42% (p=0.000 n=10)
Unmarshal/fixture=f3-1m-8     3.098Mi ± 0%   2.836Mi ± 0%  -8.47% (p=0.000 n=10)
Policy/fixture=f1-64k-8                                                            888.0 ± 0%
Policy/fixture=f2-256k-8                                                           888.0 ± 0%
Policy/fixture=f3-1m-8                                                             888.0 ± 0%
geomean                       774.3Ki        725.4Ki       -6.32%                  888.0       ? ¹ ²
¹ benchmark set differs from baseline; geomeans may not be comparable
² ratios must be >0 to compute geomean

                            │   stdlib    │                goccy                │                   │
                            │  allocs/op  │  allocs/op   vs base                │ allocs/op   vs base   │
Unmarshal/fixture=f1-64k-8    2.539k ± 0%   2.869k ± 0%  +13.00% (p=0.000 n=10)
Unmarshal/fixture=f2-256k-8   10.15k ± 0%   11.28k ± 0%  +11.12% (p=0.000 n=10)
Unmarshal/fixture=f3-1m-8     40.90k ± 0%   45.34k ± 0%  +10.86% (p=0.000 n=10)
Policy/fixture=f1-64k-8                                                           5.000 ± 0%
Policy/fixture=f2-256k-8                                                          5.000 ± 0%
Policy/fixture=f3-1m-8                                                            5.000 ± 0%
geomean                       10.18k        11.36k       +11.65%                  5.000       ? ¹ ²
¹ benchmark set differs from baseline; geomeans may not be comparable
² ratios must be >0 to compute geomean
```

## 7. Attribution: decoder vs HTTP, GC, allocation, validation, instrumentation

- [A] Handler time = read + policy scan + decode + validate (+ timer overhead). Policy and validate are shared code and should match across arms. A difference in them points to GC or scheduler interference, not to the decoder.
- [A] HTTP time − handler time ≈ network, kernel, and k6 client cost, plus response delivery. The library under test does not touch these.
- [A] The `variant=minimal` cells (no timestamps, no Server-Timing) bound the instrumentation overhead. Compare their HTTP latency and CPU/op with the matching full-instrumentation cell.
- [A] Profile cells are flagged by design and are not used for latency claims. CPU profiles are sampled evidence. Allocation profiles are windowed deltas (`allocs?seconds=60`). Heap profiles show retained objects at the end of the trial. pprof does not measure RSS.
- [L] The two arms are different binaries, so a symbol in one profile has no counterpart to diff against in the other. Compare shares of a profile (e.g. % of CPU in the decoder), not absolute sample counts.

## 8. Generator capacity

| trial | k6 avg cores | k6 peak cores | server avg cores (ps) | speed limit before/after |
|---|---:|---:|---:|---|
| goccy-legacy/f1-64k/fixed-r1625-prof-cpu/rep-01 | 0.398 | 0.609 | 0.614 | 100.0/100.0 |
| goccy-legacy/f1-64k/fixed-r1625/rep-01 | 0.382 | 0.594 | 0.603 | 100.0/100.0 |
| goccy-legacy/f1-64k/fixed-r1625/rep-02 | 0.380 | 0.596 | 0.603 | 100.0/100.0 |
| goccy-legacy/f1-64k/fixed-r1625/rep-03 | 0.380 | 0.598 | 0.604 | 100.0/100.0 |
| goccy-legacy/f1-64k/fixed-r1625/rep-04 | 0.378 | 0.590 | 0.603 | 100.0/100.0 |
| goccy-legacy/f1-64k/fixed-r1625/rep-05 | 0.402 | 0.656 | 0.630 | 100.0/100.0 |
| goccy-legacy/f1-64k/ramp-k5000/rep-01 | 1.229 | 2.777 | 1.785 | 100.0/100.0 |
| goccy/f1-64k/fixed-r2000/rep-01 | 0.485 | 0.740 | 0.763 | 100.0/100.0 |
| goccy/f1-64k/fixed-r2000/rep-02 | 0.512 | 0.742 | 0.792 | 100.0/100.0 |
| goccy/f1-64k/fixed-r2000/rep-03 | 0.475 | 0.732 | 0.752 | 100.0/100.0 |
| goccy/f1-64k/fixed-r2000/rep-04 | 0.474 | 0.724 | 0.752 | 100.0/100.0 |
| goccy/f1-64k/fixed-r2000/rep-05 | 0.477 | 0.726 | 0.753 | 100.0/100.0 |
| goccy/f1-64k/fixed-r2000/rep-06 | 0.490 | 0.759 | 0.770 | 100.0/100.0 |
| goccy/f1-64k/fixed-r3200-prof-allocs/rep-01 | 0.808 | 1.118 | 1.221 | 100.0/100.0 |
| goccy/f1-64k/fixed-r3200-prof-cpu/rep-01 | 0.808 | 1.114 | 1.220 | 100.0/100.0 |
| goccy/f1-64k/fixed-r3200/rep-01 | 0.784 | 1.118 | 1.216 | 100.0/100.0 |
| goccy/f1-64k/fixed-r3200/rep-02 | 0.792 | 1.098 | 1.219 | 100.0/100.0 |
| goccy/f1-64k/fixed-r3200/rep-03 | 0.781 | 1.096 | 1.217 | 100.0/100.0 |
| goccy/f1-64k/fixed-r3200/rep-04 | 0.783 | 1.085 | 1.217 | 100.0/100.0 |
| goccy/f1-64k/fixed-r3200/rep-05 | 0.787 | 1.090 | 1.216 | 100.0/100.0 |
| goccy/f1-64k/ramp-k5000/rep-01 | 1.234 | 2.856 | 1.786 | 100.0/100.0 |
| goccy/f2-256k/fixed-r1000/rep-01 | 0.338 | 0.462 | 1.348 | 100.0/100.0 |
| goccy/f2-256k/fixed-r1000/rep-02 | 0.335 | 0.476 | 1.346 | 100.0/100.0 |
| goccy/f2-256k/fixed-r1000/rep-03 | 0.338 | 0.492 | 1.346 | 100.0/100.0 |
| goccy/f2-256k/fixed-r1000/rep-04 | 0.336 | 0.474 | 1.347 | 100.0/100.0 |
| goccy/f2-256k/fixed-r1000/rep-05 | 0.347 | 0.530 | 1.363 | 100.0/100.0 |
| goccy/f2-256k/fixed-r625/rep-01 | 0.200 | 0.278 | 0.851 | 100.0/100.0 |
| goccy/f2-256k/fixed-r625/rep-02 | 0.200 | 0.292 | 0.854 | 100.0/100.0 |
| goccy/f2-256k/fixed-r625/rep-03 | 0.199 | 0.298 | 0.848 | 100.0/100.0 |
| goccy/f2-256k/fixed-r625/rep-04 | 0.200 | 0.292 | 0.850 | 100.0/100.0 |
| goccy/f2-256k/fixed-r625/rep-05 | 0.199 | 0.282 | 0.851 | 100.0/100.0 |
| goccy/f2-256k/ramp-k1250/rep-01 | 0.433 | 1.081 | 1.618 | 100.0/100.0 |
| goccy/f3-1m/fixed-r160/rep-01 | 0.106 | 0.136 | 0.797 | 100.0/100.0 |
| goccy/f3-1m/fixed-r160/rep-02 | 0.103 | 0.138 | 0.806 | 100.0/100.0 |
| goccy/f3-1m/fixed-r160/rep-03 | 0.104 | 0.136 | 0.795 | 100.0/100.0 |
| goccy/f3-1m/fixed-r160/rep-04 | 0.105 | 0.136 | 0.804 | 100.0/100.0 |
| goccy/f3-1m/fixed-r160/rep-05 | 0.106 | 0.138 | 0.793 | 100.0/100.0 |
| goccy/f3-1m/fixed-r256-prof-allocs/rep-01 | 0.156 | 0.202 | 1.299 | 100.0/100.0 |
| goccy/f3-1m/fixed-r256-prof-cpu/rep-01 | 0.149 | 0.190 | 1.281 | 100.0/100.0 |
| goccy/f3-1m/fixed-r256/rep-01 | 0.156 | 0.198 | 1.291 | 100.0/100.0 |
| goccy/f3-1m/fixed-r256/rep-02 | 0.156 | 0.200 | 1.293 | 100.0/100.0 |
| goccy/f3-1m/fixed-r256/rep-03 | 0.154 | 0.196 | 1.289 | 100.0/100.0 |
| goccy/f3-1m/fixed-r256/rep-04 | 0.155 | 0.190 | 1.291 | 100.0/100.0 |
| goccy/f3-1m/fixed-r256/rep-05 | 0.154 | 0.194 | 1.290 | 100.0/100.0 |
| goccy/f3-1m/ramp-k320/rep-01 | 0.206 | 0.536 | 1.505 | 100.0/100.0 |
| stdlib-legacy/f1-64k/fixed-r1625-prof-cpu/rep-01 | 0.399 | 0.628 | 1.487 | 100.0/100.0 |
| stdlib-legacy/f1-64k/fixed-r1625/rep-01 | 0.382 | 0.596 | 1.480 | 100.0/100.0 |
| stdlib-legacy/f1-64k/fixed-r1625/rep-02 | 0.385 | 0.592 | 1.481 | 100.0/100.0 |
| stdlib-legacy/f1-64k/fixed-r1625/rep-03 | 0.384 | 0.601 | 1.482 | 100.0/100.0 |
| stdlib-legacy/f1-64k/fixed-r1625/rep-04 | 0.387 | 0.604 | 1.480 | 100.0/100.0 |
| stdlib-legacy/f1-64k/fixed-r1625/rep-05 | 0.403 | 0.616 | 1.502 | 100.0/100.0 |
| stdlib-legacy/f1-64k/fixed-r1625/rep-06 | 0.386 | 0.601 | 1.477 | 100.0/100.0 |
| stdlib-legacy/f1-64k/ramp-k5000/rep-01 | 1.103 | 1.793 | 2.963 | 100.0/100.0 |
| stdlib/f1-64k/fixed-r2000/rep-01 | 0.486 | 0.698 | 1.325 | 100.0/100.0 |
| stdlib/f1-64k/fixed-r2000/rep-02 | 0.484 | 0.692 | 1.325 | 100.0/100.0 |
| stdlib/f1-64k/fixed-r2000/rep-03 | 0.479 | 0.726 | 1.319 | 100.0/100.0 |
| stdlib/f1-64k/fixed-r2000/rep-04 | 0.473 | 0.730 | 1.310 | 100.0/100.0 |
| stdlib/f1-64k/fixed-r2000/rep-05 | 0.474 | 0.722 | 1.312 | 100.0/100.0 |
| stdlib/f1-64k/fixed-r3200-prof-allocs/rep-01 | 0.826 | 1.140 | 2.143 | 100.0/100.0 |
| stdlib/f1-64k/fixed-r3200-prof-cpu/rep-01 | 0.816 | 1.127 | 2.138 | 100.0/100.0 |
| stdlib/f1-64k/fixed-r3200/rep-01 | 0.803 | 1.138 | 2.140 | 100.0/100.0 |
| stdlib/f1-64k/fixed-r3200/rep-02 | 0.810 | 1.130 | 2.143 | 100.0/100.0 |
| stdlib/f1-64k/fixed-r3200/rep-03 | 0.802 | 1.106 | 2.141 | 100.0/100.0 |
| stdlib/f1-64k/fixed-r3200/rep-04 | 0.802 | 1.132 | 2.145 | 100.0/100.0 |
| stdlib/f1-64k/fixed-r3200/rep-05 | 0.806 | 1.122 | 2.142 | 100.0/100.0 |
| stdlib/f1-64k/ramp-k5000/rep-01 | 1.142 | 2.008 | 2.536 | 100.0/100.0 |
| stdlib/f1-64k/ramp-k5000/rep-02 | 1.137 | 2.093 | 2.556 | 100.0/100.0 |
| stdlib/f2-256k/fixed-r1000/rep-01 | 0.344 | 0.474 | 2.494 | 100.0/100.0 |
| stdlib/f2-256k/fixed-r1000/rep-02 | 0.345 | 0.464 | 2.490 | 100.0/100.0 |
| stdlib/f2-256k/fixed-r1000/rep-03 | 0.346 | 0.468 | 2.490 | 100.0/100.0 |
| stdlib/f2-256k/fixed-r1000/rep-04 | 0.346 | 0.462 | 2.492 | 100.0/100.0 |
| stdlib/f2-256k/fixed-r1000/rep-05 | 0.404 | 0.636 | 2.568 | 100.0/100.0 |
| stdlib/f2-256k/fixed-r1000/rep-06 | 0.335 | 0.468 | 2.478 | 100.0/100.0 |
| stdlib/f2-256k/fixed-r625/rep-01 | 0.205 | 0.300 | 1.541 | 100.0/100.0 |
| stdlib/f2-256k/fixed-r625/rep-02 | 0.205 | 0.300 | 1.542 | 100.0/100.0 |
| stdlib/f2-256k/fixed-r625/rep-03 | 0.205 | 0.296 | 1.541 | 100.0/100.0 |
| stdlib/f2-256k/fixed-r625/rep-04 | 0.205 | 0.296 | 1.541 | 100.0/100.0 |
| stdlib/f2-256k/fixed-r625/rep-05 | 0.205 | 0.300 | 1.541 | 100.0/100.0 |
| stdlib/f2-256k/ramp-k1250/rep-01 | 0.460 | 0.928 | 2.535 | 100.0/100.0 |
| stdlib/f2-256k/ramp-k1250/rep-02 | 0.436 | 0.960 | 2.501 | 100.0/100.0 |
| stdlib/f3-1m/fixed-r160/rep-01 | 0.097 | 0.134 | 1.544 | 100.0/100.0 |
| stdlib/f3-1m/fixed-r160/rep-02 | 0.096 | 0.132 | 1.547 | 100.0/100.0 |
| stdlib/f3-1m/fixed-r160/rep-03 | 0.096 | 0.130 | 1.546 | 100.0/100.0 |
| stdlib/f3-1m/fixed-r160/rep-04 | 0.096 | 0.130 | 1.542 | 100.0/100.0 |
| stdlib/f3-1m/fixed-r160/rep-05 | 0.096 | 0.130 | 1.545 | 100.0/100.0 |
| stdlib/f3-1m/fixed-r256-prof-allocs/rep-01 | 0.174 | 0.216 | 2.519 | 100.0/100.0 |
| stdlib/f3-1m/fixed-r256-prof-cpu/rep-01 | 0.164 | 0.204 | 2.486 | 100.0/100.0 |
| stdlib/f3-1m/fixed-r256/rep-01 | 0.168 | 0.202 | 2.502 | 100.0/100.0 |
| stdlib/f3-1m/fixed-r256/rep-02 | 0.167 | 0.212 | 2.502 | 100.0/100.0 |
| stdlib/f3-1m/fixed-r256/rep-03 | 0.171 | 0.212 | 2.504 | 100.0/100.0 |
| stdlib/f3-1m/fixed-r256/rep-04 | 0.172 | 0.210 | 2.511 | 100.0/100.0 |
| stdlib/f3-1m/fixed-r256/rep-05 | 0.171 | 0.212 | 2.509 | 100.0/100.0 |
| stdlib/f3-1m/ramp-k320/rep-01 | 0.234 | 0.563 | 2.492 | 100.0/100.0 |

## 9. Threats to validity and open questions

- [L] There is a single host with heterogeneous cores, and macOS offers no CPU pinning. The server (GOMAXPROCS=4) and k6 compete for the same 8 cores and can be scheduled onto E-cores.
- [L] Traffic goes over loopback, so the network path is unrealistically short. Body-read cost is therefore a lower bound.
- [L] The fixtures are synthetic (seeded generator). F1/F2/F3 come from one schema with different record counts, so the matrix is a size sweep and does not test schema generality.
- [L] Window snapshots are taken by benchctl after the warm-up ends. Their boundaries differ from k6's measure-phase boundary by up to a few requests. The k6/server count check is applied to the whole-trial totals.
- [L] GC pause totals come from histogram bucket lower bounds, so they understate the true pause time.
- [L] n=5 trials per cell gives limited statistical power. The bootstrap CI on the median of 5 values is coarse.
- [H] goccy's decode advantage (if any) may shrink at larger payloads if GC dominates. The F2/F3 cells test this.
- [L] Both stdlib arms are reported. The default go1.27 encoding/json is backed by the v2 implementation; Pair L uses GOEXPERIMENT=nojsonv2. The two pairs must not be mixed.
