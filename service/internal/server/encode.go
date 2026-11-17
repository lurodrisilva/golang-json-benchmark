package server

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/lurodrisilva/golang-json-encoding-benchmark/service/internal/codec"
	"github.com/lurodrisilva/golang-json-encoding-benchmark/service/internal/conformance"
	"github.com/lurodrisilva/golang-json-encoding-benchmark/service/internal/metrics"
	"github.com/lurodrisilva/golang-json-encoding-benchmark/service/internal/model"
)

// encodeHandler serves GET /batch: it encodes one shared, read-only value
// with the arm's codec on every request. The only intentional difference
// between the two binaries is the codec's Marshal.
type encodeHandler struct {
	codec codec.Codec
	m     *metrics.Set
	v     *model.IngestBatch
}

func (h *encodeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t0 := now()
	h.m.EncRequests.Add(1)
	h.m.Inflight.Add(1)
	defer h.m.Inflight.Add(-1)

	if r.Method != http.MethodGet {
		h.m.EncRejectedMethod.Add(1)
		w.Header()["Allow"] = []string{"GET"}
		h.respond(w, http.StatusMethodNotAllowed)
		return
	}

	// Fresh output buffer per request: Marshal returns a new slice.
	body, err := h.codec.Marshal(h.v)
	t1 := now()
	if err != nil {
		h.m.EncErrors.Add(1)
		h.respond(w, http.StatusInternalServerError)
		return
	}

	hdr := w.Header()
	hdr["Content-Type"] = []string{"application/json"}
	hdr["Content-Length"] = []string{strconv.Itoa(len(body))}
	tEnd := now()
	recordEncTimings(h.m, w, t0, t1, tEnd)
	h.respond(w, http.StatusOK)

	n, werr := w.Write(body)
	recordWrite(h.m, tEnd)
	h.m.EncRespBytes.Add(int64(n))
	if werr != nil || n != len(body) {
		h.m.EncWriteErrors.Add(1)
		return
	}
	h.m.EncOK.Add(1)
}

// respond commits the status and counts the commit (server-side boundary).
func (h *encodeHandler) respond(w http.ResponseWriter, status int) {
	w.WriteHeader(status)
	h.m.Responded.Add(1)
}

// LoadEncodeValue decodes the fixture with the arm's own codec, validates
// it, and checks the result against the fixture's golden digest
// (<name>.golden.tsv next to <name>.json). Any mismatch is fatal: the server
// must never serve a value that differs between arms.
func LoadEncodeValue(c codec.Codec, path string) (*model.IngestBatch, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	v := new(model.IngestBatch)
	if err := c.Unmarshal(data, v); err != nil {
		return nil, fmt.Errorf("decode %s: %w", path, err)
	}
	if err := v.Validate(); err != nil {
		return nil, fmt.Errorf("validate %s: %w", path, err)
	}
	golden := strings.TrimSuffix(path, filepath.Ext(path)) + ".golden.tsv"
	g, err := conformance.LoadGolden(golden)
	if err != nil {
		return nil, fmt.Errorf("golden: %w", err)
	}
	got := fmt.Sprintf("%016x", conformance.Digest(v))
	if want := g["digest_fnv64a"]; want == "" || got != want {
		return nil, fmt.Errorf("encode fixture digest %s != golden %q", got, want)
	}
	return v, nil
}
