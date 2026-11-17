//go:build minimal

package server

import (
	"net/http"
	"time"

	"github.com/lurodrisilva/golang-json-encoding-benchmark/service/internal/metrics"
)

// InstrumentationMode is recorded in the manifest.
const InstrumentationMode = "minimal"

// now returns the zero time: the minimal control build takes no timestamps
// and emits no Server-Timing header; counters remain.
func now() time.Time { return time.Time{} }

func recordTimings(*metrics.Set, http.ResponseWriter, time.Time, time.Time, time.Time, time.Time, time.Time) {
}

func recordEncTimings(*metrics.Set, http.ResponseWriter, time.Time, time.Time, time.Time) {}

func recordWrite(*metrics.Set, time.Time) {}
