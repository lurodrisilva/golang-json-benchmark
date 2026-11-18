package bench

import (
	"path/filepath"
	"testing"

	"github.com/lurodrisilva/golang-json-encoding-benchmark/service/internal/codec"
	"github.com/lurodrisilva/golang-json-encoding-benchmark/service/internal/codec/goccy"
	"github.com/lurodrisilva/golang-json-encoding-benchmark/service/internal/codec/stdlib"
	"github.com/lurodrisilva/golang-json-encoding-benchmark/service/internal/server"
)

var outSink []byte

// BenchmarkMarshal/fixture=<f>/codec=<c> measures Marshal of the same
// digest-checked value the servers encode on GET /batch (each arm decodes
// the fixture with its own codec; the digest gate proves the values are
// identical). Value loading is outside the timer. SetBytes uses the encoded
// size, so MB/s is output throughput.
func BenchmarkMarshal(b *testing.B) {
	for _, fx := range fixtures {
		for _, c := range []codec.Codec{stdlib.Codec{}, goccy.Codec{}} {
			v, err := server.LoadEncodeValue(c, filepath.Join("..", "..", "fixtures", fx+".json"))
			if err != nil {
				b.Fatal(err)
			}
			out, err := c.Marshal(v)
			if err != nil {
				b.Fatal(err)
			}
			b.Run("fixture="+fx+"/codec="+c.Name(), func(b *testing.B) {
				b.SetBytes(int64(len(out)))
				b.ReportAllocs()
				for b.Loop() {
					o, err := c.Marshal(v)
					if err != nil {
						b.Fatal(err)
					}
					outSink = o
				}
			})
		}
	}
}
