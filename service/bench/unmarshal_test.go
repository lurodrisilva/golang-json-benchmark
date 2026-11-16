// Package bench holds decoder-only microbenchmarks. They measure Unmarshal of
// identical fixture bytes into a fresh model.IngestBatch per iteration — the
// same call the servers time as "dec". Fixture loading is outside the timer.
// Conclusions from here are reported separately from HTTP results.
//
//	go test -run '^$' -bench . -benchmem -count 10 ./bench | tee bench.txt
//	benchstat -col /codec bench.txt
package bench

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lurodrisilva/golang-json-encoding-benchmark/service/internal/codec"
	"github.com/lurodrisilva/golang-json-encoding-benchmark/service/internal/codec/goccy"
	"github.com/lurodrisilva/golang-json-encoding-benchmark/service/internal/codec/stdlib"
	"github.com/lurodrisilva/golang-json-encoding-benchmark/service/internal/model"
	"github.com/lurodrisilva/golang-json-encoding-benchmark/service/internal/policy"
)

var fixtures = []string{"f1-64k", "f2-256k", "f3-1m"}

func load(b *testing.B, name string) []byte {
	data, err := os.ReadFile(filepath.Join("..", "..", "fixtures", name+".json"))
	if err != nil {
		b.Fatal(err)
	}
	return data
}

var sink *model.IngestBatch

// BenchmarkUnmarshal/fixture=<f>/codec=<c> — key=value names let benchstat
// pivot on /codec.
func BenchmarkUnmarshal(b *testing.B) {
	for _, fx := range fixtures {
		data := load(b, fx)
		for _, c := range []codec.Codec{stdlib.Codec{}, goccy.Codec{}} {
			b.Run("fixture="+fx+"/codec="+c.Name(), func(b *testing.B) {
				b.SetBytes(int64(len(data)))
				b.ReportAllocs()
				for b.Loop() {
					v := new(model.IngestBatch)
					if err := c.Unmarshal(data, v); err != nil {
						b.Fatal(err)
					}
					sink = v
				}
			})
		}
	}
}

// BenchmarkPolicy measures the shared, codec-independent policy scan cost
// that both servers pay before decoding (context for HTTP-vs-decoder gaps).
func BenchmarkPolicy(b *testing.B) {
	for _, fx := range fixtures {
		data := load(b, fx)
		b.Run("fixture="+fx, func(b *testing.B) {
			b.SetBytes(int64(len(data)))
			b.ReportAllocs()
			for b.Loop() {
				if policy.Check(data) != policy.OK {
					b.Fatal("policy reject")
				}
			}
		})
	}
}
