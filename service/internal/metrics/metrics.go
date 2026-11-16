// Package metrics holds low-overhead process counters and renders snapshots
// (counters + runtime/metrics + getrusage) for the management listener.
//
// Hot-path cost: one atomic add per counter touched; each counter sits on its
// own cache line. There are no locks, labels, or per-request allocations.
// Snapshots are rendered with strconv into a flat JSON object so that the
// metrics path links no JSON library.
package metrics

import (
	"math"
	"runtime"
	"runtime/metrics"
	"strconv"
	"sync/atomic"
	"syscall"
	"time"
)

// Counter is a cache-line-padded atomic int64.
type Counter struct {
	atomic.Int64
	_ [56]byte
}

// Set is the full set of server counters. Counting boundaries:
//   - Received: handler entry for any path routed to /ingest.
//   - DecodedOK: Unmarshal and Validate both succeeded (one "operation").
//   - Responded: incremented after WriteHeader returns for ANY status. This is
//     the server-side commit boundary, NOT client-observed delivery.
type Set struct {
	Received           Counter
	DecodedOK          Counter
	DecodeErrors       Counter
	ValidationFailures Counter
	PolicyUTF8         Counter
	PolicyDupKey       Counter
	RejectedMethod     Counter
	RejectedCType      Counter
	RejectedSize       Counter
	ReadErrors         Counter
	Responded          Counter
	Inflight           Counter

	// Cumulative per-request durations (ns) for successful operations only.
	// These are sums over concurrent requests and are NOT wall-clock time.
	SumReadNs     Counter
	SumPolicyNs   Counter
	SumDecodeNs   Counter
	SumValidateNs Counter
	SumHandlerNs  Counter
}

var rtNames = []string{
	"/gc/heap/live:bytes",
	"/gc/heap/goal:bytes",
	"/gc/heap/allocs:bytes",
	"/gc/heap/allocs:objects",
	"/gc/heap/frees:objects",
	"/gc/cycles/total:gc-cycles",
	"/memory/classes/total:bytes",
	"/sched/goroutines:goroutines",
	"/sched/pauses/total/gc:seconds",
}

// Snapshot renders the current state as a flat JSON object.
// startWall is the process start time used for uptime.
func (s *Set) Snapshot(dst []byte, impl string, startWall time.Time) []byte {
	now := time.Now()
	dst = append(dst, '{')
	dst = str(dst, "impl", impl)
	dst = num(dst, "unix_nano", now.UnixNano())
	dst = num(dst, "uptime_ns", int64(now.Sub(startWall)))
	dst = num(dst, "gomaxprocs", int64(runtime.GOMAXPROCS(0)))

	counters := []struct {
		k string
		c *Counter
	}{
		{"received", &s.Received}, {"decoded_ok", &s.DecodedOK}, {"decode_errors", &s.DecodeErrors},
		{"validation_failures", &s.ValidationFailures}, {"policy_rejects_utf8", &s.PolicyUTF8},
		{"policy_rejects_dupkey", &s.PolicyDupKey}, {"rejected_method", &s.RejectedMethod},
		{"rejected_ctype", &s.RejectedCType}, {"rejected_size", &s.RejectedSize},
		{"read_errors", &s.ReadErrors}, {"responded", &s.Responded}, {"inflight", &s.Inflight},
		{"sum_read_ns", &s.SumReadNs}, {"sum_policy_ns", &s.SumPolicyNs}, {"sum_decode_ns", &s.SumDecodeNs},
		{"sum_validate_ns", &s.SumValidateNs}, {"sum_handler_ns", &s.SumHandlerNs},
	}
	for _, c := range counters {
		dst = num(dst, c.k, c.c.Load())
	}

	var ru syscall.Rusage
	if err := syscall.Getrusage(syscall.RUSAGE_SELF, &ru); err == nil {
		dst = num(dst, "rusage_utime_ns", tvNs(ru.Utime))
		dst = num(dst, "rusage_stime_ns", tvNs(ru.Stime))
		maxrss := int64(ru.Maxrss)
		if runtime.GOOS != "darwin" { // Linux reports KiB, darwin reports bytes
			maxrss *= 1024
		}
		dst = num(dst, "rusage_maxrss_bytes", maxrss)
		dst = num(dst, "rusage_nvcsw", int64(ru.Nvcsw))
		dst = num(dst, "rusage_nivcsw", int64(ru.Nivcsw))
	}

	samples := make([]metrics.Sample, len(rtNames))
	for i, n := range rtNames {
		samples[i].Name = n
	}
	metrics.Read(samples)
	for _, sm := range samples {
		key := "rt" + sm.Name
		switch sm.Value.Kind() {
		case metrics.KindUint64:
			dst = num(dst, key, int64(sm.Value.Uint64()))
		case metrics.KindFloat64:
			dst = append(dst, ',', '"')
			dst = append(dst, key...)
			dst = append(dst, '"', ':')
			dst = strconv.AppendFloat(dst, sm.Value.Float64(), 'g', -1, 64)
		case metrics.KindFloat64Histogram:
			h := sm.Value.Float64Histogram()
			dst = append(dst, ',', '"')
			dst = append(dst, key...)
			dst = append(dst, `":{"counts":[`...)
			for i, c := range h.Counts {
				if i > 0 {
					dst = append(dst, ',')
				}
				dst = strconv.AppendUint(dst, c, 10)
			}
			// Bucket boundaries are strings so that ±Inf survive JSON.
			dst = append(dst, `],"buckets":[`...)
			for i, b := range h.Buckets {
				if i > 0 {
					dst = append(dst, ',')
				}
				dst = append(dst, '"')
				switch {
				case math.IsInf(b, 1):
					dst = append(dst, "+Inf"...)
				case math.IsInf(b, -1):
					dst = append(dst, "-Inf"...)
				default:
					dst = strconv.AppendFloat(dst, b, 'g', -1, 64)
				}
				dst = append(dst, '"')
			}
			dst = append(dst, "]}"...)
		default:
			dst = append(dst, ',', '"')
			dst = append(dst, key...)
			dst = append(dst, `":null`...)
		}
	}
	return append(dst, '}')
}

func tvNs(tv syscall.Timeval) int64 {
	return int64(tv.Sec)*1e9 + int64(tv.Usec)*1e3
}

func str(dst []byte, k, v string) []byte {
	if len(dst) > 1 {
		dst = append(dst, ',')
	}
	dst = append(dst, '"')
	dst = append(dst, k...)
	dst = append(dst, `":`...)
	return strconv.AppendQuote(dst, v)
}

func num(dst []byte, k string, v int64) []byte {
	if len(dst) > 1 {
		dst = append(dst, ',')
	}
	dst = append(dst, '"')
	dst = append(dst, k...)
	dst = append(dst, `":`...)
	return strconv.AppendInt(dst, v, 10)
}
