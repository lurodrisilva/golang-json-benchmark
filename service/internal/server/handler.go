package server

import (
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/lurodrisilva/golang-json-encoding-benchmark/service/internal/codec"
	"github.com/lurodrisilva/golang-json-encoding-benchmark/service/internal/metrics"
	"github.com/lurodrisilva/golang-json-encoding-benchmark/service/internal/model"
	"github.com/lurodrisilva/golang-json-encoding-benchmark/service/internal/policy"
)

// DefaultBodyLimit is the maximum accepted request body (4 MiB).
const DefaultBodyLimit int64 = 4 << 20

// ingestHandler is identical for every codec; the codec is the only variable.
type ingestHandler struct {
	codec     codec.Codec
	m         *metrics.Set
	bodyLimit int64
}

// respond commits the status with an empty body and counts the commit.
// Responded is a server-side boundary: it does not prove client delivery.
func (h *ingestHandler) respond(w http.ResponseWriter, status int) {
	w.WriteHeader(status)
	h.m.Responded.Add(1)
}

func (h *ingestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t0 := now()
	h.m.Received.Add(1)
	h.m.Inflight.Add(1)
	defer h.m.Inflight.Add(-1)

	if r.Method != http.MethodPost {
		h.m.RejectedMethod.Add(1)
		w.Header().Set("Allow", http.MethodPost)
		h.respond(w, http.StatusMethodNotAllowed)
		return
	}
	if !isJSONMediaType(r.Header.Get("Content-Type")) {
		h.m.RejectedCType.Add(1)
		h.respond(w, http.StatusUnsupportedMediaType)
		return
	}
	if r.ContentLength > h.bodyLimit {
		h.m.RejectedSize.Add(1)
		h.respond(w, http.StatusRequestEntityTooLarge)
		return
	}

	body, err := readBody(w, r, h.bodyLimit)
	if err != nil {
		var mbe *http.MaxBytesError
		if errors.As(err, &mbe) {
			h.m.RejectedSize.Add(1)
			h.respond(w, http.StatusRequestEntityTooLarge)
			return
		}
		h.m.ReadErrors.Add(1)
		h.respond(w, http.StatusBadRequest)
		return
	}
	t1 := now()

	switch policy.Check(body) {
	case policy.InvalidUTF8:
		h.m.PolicyUTF8.Add(1)
		h.respond(w, http.StatusBadRequest)
		return
	case policy.DuplicateKey:
		h.m.PolicyDupKey.Add(1)
		h.respond(w, http.StatusBadRequest)
		return
	}
	t2 := now()

	dst := new(model.IngestBatch) // fresh destination per request, no pooling
	if err := h.codec.Unmarshal(body, dst); err != nil {
		h.m.DecodeErrors.Add(1)
		h.respond(w, http.StatusBadRequest)
		return
	}
	t3 := now()

	if err := dst.Validate(); err != nil {
		h.m.ValidationFailures.Add(1)
		h.respond(w, http.StatusUnprocessableEntity)
		return
	}
	tEnd := now() // taken BEFORE any header formatting

	h.m.DecodedOK.Add(1)
	recordTimings(h.m, w, t0, t1, t2, t3, tEnd)
	h.respond(w, http.StatusAccepted)
}

// readBody reads the whole body with one identical strategy for both codecs:
// a bounded preallocation from Content-Length, falling back to growth.
func readBody(w http.ResponseWriter, r *http.Request, limit int64) ([]byte, error) {
	rd := http.MaxBytesReader(w, r.Body, limit)
	capHint := int64(512)
	if r.ContentLength > 0 {
		capHint = min(r.ContentLength, limit) + 1 // +1 lets io detect EOF without regrowth
	}
	buf := make([]byte, 0, capHint)
	for {
		n, err := rd.Read(buf[len(buf):cap(buf)])
		buf = buf[:len(buf)+n]
		if err == io.EOF {
			return buf, nil
		}
		if err != nil {
			return nil, err
		}
		if len(buf) == cap(buf) {
			buf = append(buf, 0)[:len(buf)]
		}
	}
}

// isJSONMediaType accepts "application/json" with optional parameters,
// case-insensitively, without allocating.
func isJSONMediaType(ct string) bool {
	if i := strings.IndexByte(ct, ';'); i >= 0 {
		ct = ct[:i]
	}
	return strings.EqualFold(strings.TrimSpace(ct), "application/json")
}
