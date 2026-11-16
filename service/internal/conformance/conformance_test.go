package conformance_test

import (
	"bytes"
	stdjson "encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"

	goccyjson "github.com/goccy/go-json"

	"github.com/lurodrisilva/golang-json-encoding-benchmark/service/internal/codec"
	"github.com/lurodrisilva/golang-json-encoding-benchmark/service/internal/codec/goccy"
	"github.com/lurodrisilva/golang-json-encoding-benchmark/service/internal/codec/stdlib"
	"github.com/lurodrisilva/golang-json-encoding-benchmark/service/internal/conformance"
	"github.com/lurodrisilva/golang-json-encoding-benchmark/service/internal/metrics"
	"github.com/lurodrisilva/golang-json-encoding-benchmark/service/internal/model"
	"github.com/lurodrisilva/golang-json-encoding-benchmark/service/internal/server"
)

var codecs = []codec.Codec{stdlib.Codec{}, goccy.Codec{}}

var fixtures = []string{"f1-64k", "f2-256k", "f3-1m"}

const fixtureDir = "../../../fixtures"

func loadFixture(t testing.TB, name string) ([]byte, conformance.Golden) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(fixtureDir, name+".json"))
	if err != nil {
		t.Fatal(err)
	}
	g, err := conformance.LoadGolden(filepath.Join(fixtureDir, name+".golden.tsv"))
	if err != nil {
		t.Fatal(err)
	}
	return data, g
}

// TestFixturesMatchGolden: every codec decodes every fixture to exactly the
// generator's values (digest over all fields) plus readable spot checks.
func TestFixturesMatchGolden(t *testing.T) {
	for _, fx := range fixtures {
		data, g := loadFixture(t, fx)
		if strconv.Itoa(len(data)) != g["bytes"] {
			t.Fatalf("%s: %d bytes on disk, golden says %s", fx, len(data), g["bytes"])
		}
		for _, c := range codecs {
			t.Run(fx+"/"+c.Name(), func(t *testing.T) {
				var b model.IngestBatch
				if err := c.Unmarshal(data, &b); err != nil {
					t.Fatal(err)
				}
				if err := b.Validate(); err != nil {
					t.Fatal(err)
				}
				if got := fmt.Sprintf("%016x", conformance.Digest(&b)); got != g["digest_fnv64a"] {
					t.Errorf("digest %s, golden %s", got, g["digest_fnv64a"])
				}
				checks := map[string]string{
					"batch_id":        b.BatchID,
					"events":          strconv.Itoa(len(b.Events)),
					"first_id":        strconv.FormatInt(b.Events[0].ID, 10),
					"first_user_name": b.Events[0].User.Name,
					"first_note":      strconv.Quote(b.Events[0].Note),
					"last_id":         strconv.FormatInt(b.Events[len(b.Events)-1].ID, 10),
					"last_amount":     strconv.FormatFloat(b.Events[len(b.Events)-1].Amount, 'g', -1, 64),
				}
				for k, v := range checks {
					if g[k] != v {
						t.Errorf("%s = %q, golden %q", k, v, g[k])
					}
				}
			})
		}
	}
}

// TestFixturesHaveNoUnknownFields proves the model consumes every field in
// the fixtures (size is not inflated by silently ignored data). Both
// libraries' strict decoders are used as independent witnesses.
func TestFixturesHaveNoUnknownFields(t *testing.T) {
	for _, fx := range fixtures {
		data, _ := loadFixture(t, fx)
		d1 := stdjson.NewDecoder(bytes.NewReader(data))
		d1.DisallowUnknownFields()
		var a model.IngestBatch
		if err := d1.Decode(&a); err != nil {
			t.Errorf("%s stdlib strict: %v", fx, err)
		}
		if _, err := d1.Token(); err != io.EOF {
			t.Errorf("%s: document not fully consumed: %v", fx, err)
		}
		d2 := goccyjson.NewDecoder(bytes.NewReader(data))
		d2.DisallowUnknownFields()
		var b model.IngestBatch
		if err := d2.Decode(&b); err != nil {
			t.Errorf("%s goccy strict: %v", fx, err)
		}
	}
}

// TestCorpusRaw compares each codec's raw Unmarshal to the v1-documented spec.
// Deviations are COMPATIBILITY FINDINGS: they are written to the matrix (set
// CONFORMANCE_OUT=path.tsv) and logged; they fail only with CONFORMANCE_STRICT=1.
func TestCorpusRaw(t *testing.T) {
	var out bytes.Buffer
	fmt.Fprintf(&out, "goexperiment\tcodec\tcase\tspec\tactual\tvalue_check\tconforms\terror\n")
	exp := os.Getenv("GOEXPERIMENT")
	if exp == "" {
		exp = "default"
	}
	deviations := 0
	for _, c := range codecs {
		for _, k := range conformance.Corpus {
			var b model.IngestBatch
			err := c.Unmarshal([]byte(k.Body), &b)
			actual := conformance.SpecAccept
			if err != nil {
				actual = conformance.SpecReject
			}
			check := ""
			if err == nil && k.Check != nil {
				check = k.Check(&b)
			}
			conforms := actual == k.Raw && check == ""
			if !conforms {
				deviations++
				t.Logf("DEVIATION %s/%s: spec=%v actual=%v check=%q err=%v", c.Name(), k.Name, k.Raw, actual, check, err)
				if os.Getenv("CONFORMANCE_STRICT") == "1" {
					t.Fail()
				}
			}
			es := ""
			if err != nil {
				es = strings.NewReplacer("\t", " ", "\n", " ").Replace(err.Error())
			}
			fmt.Fprintf(&out, "%s\t%s\t%s\t%v\t%v\t%s\t%t\t%s\n", exp, c.Name(), k.Name, k.Raw, actual, check, conforms, es)
		}
	}
	t.Logf("%d deviations from v1 spec across %d codecs × %d cases", deviations, len(codecs), len(conformance.Corpus))
	if p := os.Getenv("CONFORMANCE_OUT"); p != "" {
		if err := os.WriteFile(p, out.Bytes(), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func post(t testing.TB, h http.Handler, method, ctype string, body []byte) *httptest.ResponseRecorder {
	t.Helper()
	r := httptest.NewRequest(method, "/ingest", bytes.NewReader(body))
	if ctype != "" {
		r.Header.Set("Content-Type", ctype)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

// TestCorpusHTTP: through the full handler (with policy) both servers must
// produce the agreed status and an empty body for every corpus case.
func TestCorpusHTTP(t *testing.T) {
	for _, c := range codecs {
		h := server.NewIngestHandler(c, new(metrics.Set), server.DefaultBodyLimit)
		for _, k := range conformance.Corpus {
			w := post(t, h, http.MethodPost, "application/json", []byte(k.Body))
			if w.Code != k.HTTP {
				t.Errorf("%s/%s: status %d, want %d", c.Name(), k.Name, w.Code, k.HTTP)
			}
			if w.Body.Len() != 0 {
				t.Errorf("%s/%s: non-empty body %q", c.Name(), k.Name, w.Body.String())
			}
		}
	}
}

func TestHTTPContract(t *testing.T) {
	big := bytes.Repeat([]byte(" "), int(server.DefaultBodyLimit)+1)
	cases := []struct {
		name, method, ctype string
		body                []byte
		want                int
	}{
		{"valid", "POST", "application/json", []byte(`{"batch_id":"b","events":[{"id":1}]}`), 202},
		{"ctype_with_charset", "POST", "application/json; charset=utf-8", []byte(`{"batch_id":"b","events":[{"id":1}]}`), 202},
		{"ctype_uppercase", "POST", "Application/JSON", []byte(`{"batch_id":"b","events":[{"id":1}]}`), 202},
		{"get", "GET", "application/json", nil, 405},
		{"put", "PUT", "application/json", []byte(`{}`), 405},
		{"text_plain", "POST", "text/plain", []byte(`{}`), 415},
		{"no_ctype", "POST", "", []byte(`{}`), 415},
		{"json_suffix_not_accepted", "POST", "application/problem+json", []byte(`{}`), 415},
		{"oversize", "POST", "application/json", big, 413},
	}
	for _, c := range codecs {
		h := server.NewIngestHandler(c, new(metrics.Set), server.DefaultBodyLimit)
		for _, k := range cases {
			w := post(t, h, k.method, k.ctype, k.body)
			if w.Code != k.want {
				t.Errorf("%s/%s: %d want %d", c.Name(), k.name, w.Code, k.want)
			}
			if k.want == 405 && w.Header().Get("Allow") != "POST" {
				t.Errorf("%s/%s: Allow=%q", c.Name(), k.name, w.Header().Get("Allow"))
			}
			if k.want == 202 && server.InstrumentationMode == "full" && !strings.Contains(w.Header().Get("Server-Timing"), "dec;dur=") {
				t.Errorf("%s/%s: missing Server-Timing", c.Name(), k.name)
			}
		}
	}
}

// TestOversizeChunked covers a body without Content-Length exceeding the limit.
func TestOversizeChunked(t *testing.T) {
	for _, c := range codecs {
		srv := httptest.NewServer(server.NewIngestHandler(c, new(metrics.Set), 1024))
		pr, pw := io.Pipe()
		go func() {
			pw.Write([]byte(`{"batch_id":"`))
			pw.Write(bytes.Repeat([]byte("a"), 4096))
			pw.Write([]byte(`"}`))
			pw.Close()
		}()
		resp, err := http.Post(srv.URL+"/ingest", "application/json", pr)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if resp.StatusCode != 413 {
			t.Errorf("%s: chunked oversize got %d", c.Name(), resp.StatusCode)
		}
		srv.Close()
	}
}

// TestCountersConsistent checks that counters partition received requests.
func TestCountersConsistent(t *testing.T) {
	for _, c := range codecs {
		m := new(metrics.Set)
		h := server.NewIngestHandler(c, m, server.DefaultBodyLimit)
		for _, k := range conformance.Corpus {
			post(t, h, "POST", "application/json", []byte(k.Body))
		}
		post(t, h, "GET", "", nil)
		post(t, h, "POST", "text/plain", nil)
		sum := m.DecodedOK.Load() + m.DecodeErrors.Load() + m.ValidationFailures.Load() + m.PolicyUTF8.Load() +
			m.PolicyDupKey.Load() + m.RejectedMethod.Load() + m.RejectedCType.Load() + m.RejectedSize.Load() + m.ReadErrors.Load()
		if sum != m.Received.Load() || m.Responded.Load() != m.Received.Load() || m.Inflight.Load() != 0 {
			t.Errorf("%s: received=%d outcome-sum=%d responded=%d inflight=%d", c.Name(), m.Received.Load(), sum, m.Responded.Load(), m.Inflight.Load())
		}
	}
}

// TestConcurrentIsolation runs mixed fixtures concurrently through a real
// server; run with -race. A fresh destination per request means each
// response must reflect only its own payload (verified via decode digests in
// a parallel direct path as well).
func TestConcurrentIsolation(t *testing.T) {
	type fx struct {
		data   []byte
		digest string
	}
	var all []fx
	for _, name := range fixtures {
		d, g := loadFixture(t, name)
		all = append(all, fx{d, g["digest_fnv64a"]})
	}
	for _, c := range codecs {
		m := new(metrics.Set)
		srv := httptest.NewServer(server.NewIngestHandler(c, m, server.DefaultBodyLimit))
		var wg sync.WaitGroup
		const workers, per = 8, 6
		for w := range workers {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := range per {
					f := all[(w+i)%len(all)]
					resp, err := http.Post(srv.URL+"/ingest", "application/json", bytes.NewReader(f.data))
					if err != nil {
						t.Error(err)
						return
					}
					io.Copy(io.Discard, resp.Body)
					resp.Body.Close()
					if resp.StatusCode != 202 {
						t.Errorf("%s: status %d", c.Name(), resp.StatusCode)
					}
					var b model.IngestBatch
					if err := c.Unmarshal(f.data, &b); err != nil || fmt.Sprintf("%016x", conformance.Digest(&b)) != f.digest {
						t.Errorf("%s: concurrent direct decode mismatch (err=%v)", c.Name(), err)
					}
				}
			}()
		}
		wg.Wait()
		srv.Close()
		if got := m.DecodedOK.Load(); got != workers*per {
			t.Errorf("%s: decoded_ok=%d want %d", c.Name(), got, workers*per)
		}
	}
}

// TestCodecDoesNotRetainInput: mutating the input after Unmarshal must not
// change decoded strings (guards against zero-copy aliasing in either codec).
func TestCodecDoesNotRetainInput(t *testing.T) {
	for _, c := range codecs {
		data, g := loadFixture(t, "f1-64k")
		buf := bytes.Clone(data)
		var b model.IngestBatch
		if err := c.Unmarshal(buf, &b); err != nil {
			t.Fatal(err)
		}
		for i := range buf {
			buf[i] = 'Z'
		}
		if got := fmt.Sprintf("%016x", conformance.Digest(&b)); got != g["digest_fnv64a"] {
			t.Errorf("%s: decoded value aliases input buffer", c.Name())
		}
	}
}
