# Metric dictionary

| ID | Metric | Unit | Source / method | Boundary | Overhead |
|---|---|---|---|---|---|
| A | `http_req_duration` | ms | k6 | Begins when k6 starts sending the request. Covers sending + waiting + receiving (k6 `http_req_sending` + `http_req_waiting` + `http_req_receiving`). **Excludes** DNS, TCP connect and TLS (`http_req_blocked`/`connecting`, which are 0 with keep-alive and reused connections) and k6 scheduling delay. Runs to the last response byte, so it **includes** response delivery. | k6 client |
| B | `srv_hdl` | ms | `Server-Timing: hdl` → k6 Trend | Starts at handler entry. Ends after `Validate`, **before** the Server-Timing header is formatted and before `WriteHeader`. **Excludes** response delivery, net/http connection handling, and header parsing. | 5 × `time.Now()` |
| C | `srv_dec` | ms | `Server-Timing: dec` | The `codec.Unmarshal` call only, including allocation of the destination. | — |
| D | `srv_read` | ms | `Server-Timing: read` | The bounded body read (`MaxBytesReader` + a read loop into a buffer preallocated from Content-Length). | — |
| D′ | `srv_pol`, `srv_val` | ms | `Server-Timing: pol`, `val` | The shared policy scan and `Validate`. These should be identical across arms. | — |
| E | Σ handler / Σ decode / Σ read / Σ policy / Σ validate | ns | atomic `Sum*Ns` counters, delta over the window | Same as B–D′. Sums of concurrent durations: these are **not** elapsed time. | 1 atomic add each |
| F | Measured wall-clock | s | benchctl | k6 measure-phase nominal duration, actual wall time, and the drain (k6 finishing in-flight requests). | — |
| G | CPU-seconds | s | `getrusage(RUSAGE_SELF)` utime+stime delta, via `/metrics` | The measurement window. `cpu_us_per_op` = CPU-s ÷ Δdecoded_ok. | syscall per scrape |
| H | Counters | count | atomics: received, decoded_ok, decode_errors, validation_failures, policy_rejects_{utf8,dupkey}, rejected_{method,ctype,size}, read_errors, responded, inflight | `responded` counts after `WriteHeader` returns. This is **server-side only** and does not prove client delivery. k6 `http_reqs` with status 202 is the client-observed count. | padded atomics, no locks |
| I | Live heap / heap goal | bytes | `runtime/metrics` `/gc/heap/live:bytes`, `/gc/heap/goal:bytes` | Point-in-time snapshot at the start and end of the window, **without** a forced GC. | — |
| J | Allocations | bytes, objects | `/gc/heap/allocs:bytes`, `/gc/heap/allocs:objects` deltas ÷ Δdecoded_ok | The window. | — |
| K | GC | cycles, s | `/gc/cycles/total:gc-cycles` delta; `/sched/pauses/total/gc:seconds` histogram delta | The window. Pause time is summed from bucket **lower bounds**, so it is a lower bound itself. Pause count is about 2 per cycle (two STW phases). | — |
| L | RSS | KiB | `ps -o rss` every 5 s (avg and peak), plus `ru_maxrss` | Whole process. pprof does **not** measure RSS. | 1 `ps` per 5 s |
| M | Process CPU utilization | cores | `ps` cputime deltas (server and k6) | 5 s samples. These feed the generator-capacity gate. | — |
| N | Offered / achieved rate | req/s | k6 arrival-rate config; `http_reqs` rate for the measure phase | Achieved < 99% of offered → invalid. | — |
| O | Dropped iterations, non-202, transport errors | count | k6 `dropped_iterations{scenario}`, `non_202`, `transport_errors` (status 0: timeouts, resets) | Never filtered out of the results. | — |
| P | Microbench | ns/op, B/op, allocs/op | `go test -bench` + benchstat | Only `Unmarshal` into a fresh destination. Fixture loading happens outside the timer. | — |
| Q | Profiles | samples | `/debug/pprof/profile?seconds=60`, `/debug/pprof/allocs?seconds=60` (windowed delta), `heap` at the end | Only in dedicated profiling trials. The operations counted in the window are recorded next to each profile. | ~5% CPU **[H]** |

**Operation** means one POST that was successfully decoded and validated (`decoded_ok`). Attempted requests (`received`) are always kept separate from operations.

**Instrumentation control.** Binaries built with `-tags minimal` take no timestamps and emit no Server-Timing header. Comparing their HTTP latency and CPU per op with the full-instrumentation build bounds the instrumentation overhead.
