package conformance_test

import (
	"bytes"
	stdjson "encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/lurodrisilva/golang-json-encoding-benchmark/service/internal/conformance"
	"github.com/lurodrisilva/golang-json-encoding-benchmark/service/internal/metrics"
	"github.com/lurodrisilva/golang-json-encoding-benchmark/service/internal/model"
	"github.com/lurodrisilva/golang-json-encoding-benchmark/service/internal/server"
)

// Encode conformance: expectations come from the documented encoding/json v1
// Marshal semantics, never from comparing the two libraries.

// encodeCase mutates a minimal valid batch and states the v1-expected
// output: either a substring that must appear, or that Marshal must fail.
type encodeCase struct {
	name    string
	mutate  func(b *model.IngestBatch)
	want    string // substring expected in the output
	wantErr bool
	note    string
}

func baseBatch() *model.IngestBatch {
	return &model.IngestBatch{
		BatchID: "b-1",
		Source:  model.Source{Host: "h", Region: "r", Tags: []string{"t"}},
		Events: []model.Event{{
			ID: 1, Type: "click", TS: "2026-01-01T00:00:00Z",
			User:  model.User{ID: 2, Name: "n", Email: "e@x"},
			Items: []model.Item{{SKU: "s", Qty: 1, UnitPrice: 1.5, Attrs: map[string]string{"k": "v"}}},
			Flags: []bool{true},
		}},
	}
}

// invalidUTF8Want is the documented encoding/json behaviour of the selected
// toolchain engine: legacy v1 (GOEXPERIMENT=nojsonv2) writes "\\ufffd";
// the go1.27 default (v2-backed) writes the replacement rune as raw UTF-8.
func invalidUTF8Want() string {
	if strings.Contains(os.Getenv("GOEXPERIMENT"), "nojsonv2") {
		return `"note":"a\ufffdb"`
	}
	return "\"note\":\"a\ufffdb\""
}

func ev(b *model.IngestBatch) *model.Event { return &b.Events[0] }

var encodeCorpus = []encodeCase{
	{"html_escape", func(b *model.IngestBatch) { ev(b).Note = "<a&b>" }, `"note":"\u003ca\u0026b\u003e"`, false, "v1 escapes <, >, & (HTMLEscape)"},
	{"line_separators", func(b *model.IngestBatch) { ev(b).Note = "x\u2028y\u2029z" }, `"note":"x\u2028y\u2029z"`, false, "v1 escapes U+2028/U+2029"},
	{"control_chars", func(b *model.IngestBatch) { ev(b).Note = "a\tb\nc\x01" }, `"note":"a\tb\nc\u0001"`, false, "short escapes for \\t \\n, \\u00XX otherwise"},
	{"quote_backslash", func(b *model.IngestBatch) { ev(b).Note = `q"\` }, `"note":"q\"\\"`, false, ""},
	{"slash_not_escaped", func(b *model.IngestBatch) { ev(b).Note = "a/b" }, `"note":"a/b"`, false, "v1 does not escape /"},
	{"unicode_raw", func(b *model.IngestBatch) { ev(b).User.Name = "José 😀 漢" }, `"name":"José 😀 漢"`, false, "non-ASCII written as UTF-8"},
	{"invalid_utf8", func(b *model.IngestBatch) { ev(b).Note = "a\xffb" }, invalidUTF8Want(), false, "replaced by U+FFFD: escaped as \\ufffd by legacy v1, written raw by go1.27 v2-backed encoding/json"},
	{"float_integral", func(b *model.IngestBatch) { ev(b).Amount = 100 }, `"amount":100,`, false, ""},
	{"float_decimal", func(b *model.IngestBatch) { ev(b).Amount = 0.1 }, `"amount":0.1,`, false, "shortest round-trip repr"},
	{"float_1e20", func(b *model.IngestBatch) { ev(b).Amount = 1e20 }, `"amount":100000000000000000000,`, false, "below 1e21 → no exponent"},
	{"float_1e21", func(b *model.IngestBatch) { ev(b).Amount = 1e21 }, `"amount":1e+21,`, false, "≥1e21 → exponent"},
	{"float_1e-6", func(b *model.IngestBatch) { ev(b).Amount = 1e-6 }, `"amount":0.000001,`, false, "≥1e-6 → no exponent"},
	{"float_1e-7", func(b *model.IngestBatch) { ev(b).Amount = 1e-7 }, `"amount":1e-7,`, false, "<1e-6 → exponent, 2-digit exp cleaned"},
	{"float_neg_zero", func(b *model.IngestBatch) { ev(b).Amount = math.Copysign(0, -1) }, `"amount":-0,`, false, "v1 strconv 'f' keeps the sign"},
	{"float_max", func(b *model.IngestBatch) { ev(b).Amount = math.MaxFloat64 }, `"amount":1.7976931348623157e+308,`, false, ""},
	{"float_smallest", func(b *model.IngestBatch) { ev(b).Amount = math.SmallestNonzeroFloat64 }, `"amount":5e-324,`, false, ""},
	{"float_nan", func(b *model.IngestBatch) { ev(b).Amount = math.NaN() }, "", true, "UnsupportedValueError"},
	{"float_inf", func(b *model.IngestBatch) { ev(b).Amount = math.Inf(1) }, "", true, "UnsupportedValueError"},
	{"float_neg_inf", func(b *model.IngestBatch) { ev(b).Items[0].UnitPrice = math.Inf(-1) }, "", true, "UnsupportedValueError"},
	{"int64_beyond_2_53", func(b *model.IngestBatch) { ev(b).ID = 1<<53 + 1 }, `"id":9007199254740993,`, false, "exact digits"},
	{"int64_max", func(b *model.IngestBatch) { ev(b).ID = math.MaxInt64 }, `"id":9223372036854775807,`, false, ""},
	{"int64_min", func(b *model.IngestBatch) { ev(b).User.ID = math.MinInt64 }, `"id":-9223372036854775808,`, false, ""},
	{"int32_min", func(b *model.IngestBatch) { ev(b).Items[0].Qty = math.MinInt32 }, `"qty":-2147483648,`, false, ""},
	{"nil_slice", func(b *model.IngestBatch) { b.Source.Tags = nil }, `"tags":null`, false, "nil slice → null"},
	{"empty_slice", func(b *model.IngestBatch) { b.Source.Tags = []string{} }, `"tags":[]`, false, ""},
	{"nil_map", func(b *model.IngestBatch) { ev(b).Items[0].Attrs = nil }, `"attrs":null`, false, "nil map → null"},
	{"empty_map", func(b *model.IngestBatch) { ev(b).Items[0].Attrs = map[string]string{} }, `"attrs":{}`, false, ""},
	{"nil_ptr_string", func(b *model.IngestBatch) { ev(b).User.Locale = nil }, `"locale":null`, false, ""},
	{"nil_ptr_struct", func(b *model.IngestBatch) { ev(b).Geo = nil }, `"geo":null`, false, ""},
	{"ptr_set", func(b *model.IngestBatch) {
		l := "pt-BR"
		ev(b).User.Locale = &l
		ev(b).Geo = &model.Geo{Lat: -23.5, Lon: -46.25}
	}, `"locale":"pt-BR"},"amount":0,"currency":"","items"`, false, ""},
	{"map_key_order", func(b *model.IngestBatch) {
		ev(b).Items[0].Attrs = map[string]string{"zeta": "1", "alpha": "2", "Beta": "3", "é": "4", "a<b": "5"}
	}, `"attrs":{"Beta":"3","a\u003cb":"5","alpha":"2","zeta":"1","é":"4"}`, false, "keys sorted by bytes; keys HTML-escaped"},
	{"bool_flags", func(b *model.IngestBatch) { ev(b).Flags = []bool{true, false} }, `"flags":[true,false]`, false, ""},
	{"field_order", func(b *model.IngestBatch) {}, `{"batch_id":"b-1","source":{"host":"h","region":"r","tags":["t"]},"created_at":"","events":[{"id":1,"type":"click"`, false, "declaration order, no whitespace"},
}

// TestEncodeCorpus records each codec's outcome against the v1 expectation.
// Deviations are written to $ENCODE_CONFORMANCE_OUT and fail only with
// CONFORMANCE_STRICT=1 (they are findings, not normalized away).
func TestEncodeCorpus(t *testing.T) {
	var tsv strings.Builder
	tsv.WriteString("goexperiment\tcodec\tcase\tspec\tactual\tconforms\tnote\n")
	gx := os.Getenv("GOEXPERIMENT")
	for _, c := range codecs {
		for _, tc := range encodeCorpus {
			b := baseBatch()
			tc.mutate(b)
			out, err := c.Marshal(b)
			spec, actual := "ok", "ok"
			if tc.wantErr {
				spec = "error"
			}
			if err != nil {
				actual = "error"
			}
			conforms := spec == actual
			if conforms && !tc.wantErr && !bytes.Contains(out, []byte(tc.want)) {
				conforms = false
				actual = "ok-different-bytes"
			}
			if !conforms {
				msg := fmt.Sprintf("%s/%s: spec %s, got %s (err=%v) out=%.200s", c.Name(), tc.name, spec, actual, err, out)
				if os.Getenv("CONFORMANCE_STRICT") == "1" {
					t.Error(msg)
				} else {
					t.Log("DEVIATION " + msg)
				}
			}
			fmt.Fprintf(&tsv, "%s\t%s\t%s\t%s\t%s\t%t\t%s\n", gx, c.Name(), tc.name, spec, actual, conforms, tc.note)
		}
	}
	if p := os.Getenv("ENCODE_CONFORMANCE_OUT"); p != "" {
		if err := os.WriteFile(p, []byte(tsv.String()), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

// TestEncodeRoundTrip: each codec's encoding of each fixture value decodes
// (with stdlib, an independent reader) back to the golden digest. Output
// sizes and byte equality between libraries are logged as findings.
func TestEncodeRoundTrip(t *testing.T) {
	for _, f := range fixtures {
		var outs [][]byte
		for _, c := range codecs {
			v, err := server.LoadEncodeValue(c, filepath.Join(fixtureDir, f+".json"))
			if err != nil {
				t.Fatalf("%s/%s: %v", c.Name(), f, err)
			}
			out, err := c.Marshal(v)
			if err != nil {
				t.Fatalf("%s/%s marshal: %v", c.Name(), f, err)
			}
			var back model.IngestBatch
			if err := stdjson.Unmarshal(out, &back); err != nil {
				t.Fatalf("%s/%s: output not decodable: %v", c.Name(), f, err)
			}
			_, g := loadFixture(t, f)
			if got := fmt.Sprintf("%016x", conformance.Digest(&back)); got != g["digest_fnv64a"] {
				t.Errorf("%s/%s: round-trip digest %s, golden %s", c.Name(), f, got, g["digest_fnv64a"])
			}
			t.Logf("%s/%s: encoded %d bytes", c.Name(), f, len(out))
			outs = append(outs, out)
		}
		t.Logf("%s: byte-identical across codecs: %t", f, bytes.Equal(outs[0], outs[1]))
	}
}

// TestEncodeHTTPContract covers GET /batch framing and method handling.
func TestEncodeHTTPContract(t *testing.T) {
	for _, c := range codecs {
		v, err := server.LoadEncodeValue(c, filepath.Join(fixtureDir, "f1-64k.json"))
		if err != nil {
			t.Fatal(err)
		}
		want, _ := c.Marshal(v)
		m := new(metrics.Set)
		ts := httptest.NewServer(server.NewHandler(c, m, server.DefaultBodyLimit, v))
		res, err := http.Get(ts.URL + "/batch?x=1")
		if err != nil {
			t.Fatal(err)
		}
		body, _ := io.ReadAll(res.Body)
		res.Body.Close()
		if res.StatusCode != 200 || !bytes.Equal(body, want) {
			t.Errorf("%s: GET status %d, body equal %t", c.Name(), res.StatusCode, bytes.Equal(body, want))
		}
		if cl := res.Header.Get("Content-Length"); cl != strconv.Itoa(len(want)) {
			t.Errorf("%s: Content-Length %q, want %d", c.Name(), cl, len(want))
		}
		if len(res.TransferEncoding) != 0 {
			t.Errorf("%s: unexpected Transfer-Encoding %v", c.Name(), res.TransferEncoding)
		}
		if ct := res.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("%s: Content-Type %q", c.Name(), ct)
		}
		if server.InstrumentationMode == "full" && !strings.HasPrefix(res.Header.Get("Server-Timing"), "enc;dur=") {
			t.Errorf("%s: Server-Timing %q", c.Name(), res.Header.Get("Server-Timing"))
		}
		for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodDelete} {
			req, _ := http.NewRequest(method, ts.URL+"/batch", strings.NewReader("{}"))
			r, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			r.Body.Close()
			if r.StatusCode != 405 || r.Header.Get("Allow") != "GET" {
				t.Errorf("%s %s: status %d Allow %q", c.Name(), method, r.StatusCode, r.Header.Get("Allow"))
			}
		}
		ts.Close()
		if m.EncOK.Load() != 1 || m.EncRejectedMethod.Load() != 3 || m.EncRespBytes.Load() != int64(len(want)) || m.Received.Load() != 0 {
			t.Errorf("%s: counters enc_ok=%d rejected=%d bytes=%d received=%d", c.Name(),
				m.EncOK.Load(), m.EncRejectedMethod.Load(), m.EncRespBytes.Load(), m.Received.Load())
		}
	}
}

// TestEncodeConcurrentSharedValue: many goroutines encode the same shared
// value; outputs must be identical and the value unchanged (run with -race).
func TestEncodeConcurrentSharedValue(t *testing.T) {
	for _, c := range codecs {
		v, err := server.LoadEncodeValue(c, filepath.Join(fixtureDir, "f1-64k.json"))
		if err != nil {
			t.Fatal(err)
		}
		before := conformance.Digest(v)
		want, _ := c.Marshal(v)
		var wg sync.WaitGroup
		errs := make(chan string, 64)
		for range 64 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for range 20 {
					out, err := c.Marshal(v)
					if err != nil || !bytes.Equal(out, want) {
						errs <- fmt.Sprintf("%s: concurrent output differs (err=%v)", c.Name(), err)
						return
					}
				}
			}()
		}
		wg.Wait()
		close(errs)
		for e := range errs {
			t.Error(e)
		}
		if conformance.Digest(v) != before {
			t.Errorf("%s: shared value mutated by Marshal", c.Name())
		}
	}
}

// TestEncodeFixtureDigestGate: a fixture whose value differs from its golden
// digest must be refused (the server never becomes ready).
func TestEncodeFixtureDigestGate(t *testing.T) {
	dir := t.TempDir()
	data, err := os.ReadFile(filepath.Join(fixtureDir, "f1-64k.json"))
	if err != nil {
		t.Fatal(err)
	}
	golden, _ := os.ReadFile(filepath.Join(fixtureDir, "f1-64k.golden.tsv"))
	tampered := bytes.Replace(data, []byte(`"batch_id":"`), []byte(`"batch_id":"X`), 1)
	os.WriteFile(filepath.Join(dir, "f.json"), tampered, 0o644)
	os.WriteFile(filepath.Join(dir, "f.golden.tsv"), golden, 0o644)
	for _, c := range codecs {
		if _, err := server.LoadEncodeValue(c, filepath.Join(dir, "f.json")); err == nil || !strings.Contains(err.Error(), "digest") {
			t.Errorf("%s: tampered fixture accepted (err=%v)", c.Name(), err)
		}
	}
}
