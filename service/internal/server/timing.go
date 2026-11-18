//go:build !minimal

package server

import (
	"net/http"
	"strconv"
	"time"

	"github.com/lurodrisilva/golang-json-encoding-benchmark/service/internal/metrics"
)

// InstrumentationMode is recorded in the manifest.
const InstrumentationMode = "full"

func now() time.Time { return time.Now() }

// recordTimings accumulates per-phase sums and emits Server-Timing (ms, 3
// decimals). Values cover handler entry → just before header commit; they
// exclude response delivery and everything outside the handler.
func recordTimings(m *metrics.Set, w http.ResponseWriter, t0, t1, t2, t3, tEnd time.Time) {
	read, pol, dec, val, hdl := t1.Sub(t0), t2.Sub(t1), t3.Sub(t2), tEnd.Sub(t3), tEnd.Sub(t0)
	m.SumReadNs.Add(int64(read))
	m.SumPolicyNs.Add(int64(pol))
	m.SumDecodeNs.Add(int64(dec))
	m.SumValidateNs.Add(int64(val))
	m.SumHandlerNs.Add(int64(hdl))

	var b [128]byte
	s := b[:0]
	s = appendDur(s, "read", read)
	s = append(s, ',')
	s = appendDur(s, "pol", pol)
	s = append(s, ',')
	s = appendDur(s, "dec", dec)
	s = append(s, ',')
	s = appendDur(s, "val", val)
	s = append(s, ',')
	s = appendDur(s, "hdl", hdl)
	w.Header()["Server-Timing"] = []string{string(s)}
}

func appendDur(dst []byte, name string, d time.Duration) []byte {
	dst = append(dst, name...)
	dst = append(dst, ";dur="...)
	return strconv.AppendFloat(dst, float64(d)/1e6, 'f', 3, 64)
}

// recordEncTimings accumulates encode sums and emits Server-Timing
// "enc;dur=,ehdl;dur=" (ms). ehdl ends before the header commit, so it
// excludes writing the response body (measured separately in recordWrite).
func recordEncTimings(m *metrics.Set, w http.ResponseWriter, t0, t1, tEnd time.Time) {
	enc, hdl := t1.Sub(t0), tEnd.Sub(t0)
	m.SumEncodeNs.Add(int64(enc))
	m.SumEncHandlerNs.Add(int64(hdl))
	var b [64]byte
	s := b[:0]
	s = appendDur(s, "enc", enc)
	s = append(s, ',')
	s = appendDur(s, "ehdl", hdl)
	w.Header()["Server-Timing"] = []string{string(s)}
}

// recordWrite adds the body-write duration (header commit → Write return)
// to a server-side sum. It cannot be sent to the client in a header.
func recordWrite(m *metrics.Set, start time.Time) {
	m.SumWriteNs.Add(int64(time.Since(start)))
}
