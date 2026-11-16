// Package conformance holds the library-independent oracles used by the
// correctness tests: the canonical digest over a decoded model, the golden
// loader, and the correctness corpus with spec-derived expectations.
package conformance

import (
	"bufio"
	"bytes"
	"fmt"
	"hash"
	"hash/fnv"
	"math"
	"os"
	"sort"
	"strings"

	"github.com/lurodrisilva/golang-json-encoding-benchmark/service/internal/model"
)

// Digest walks every model field in canonical order (type-tagged,
// length-prefixed, FNV-1a 64). It MUST stay in lockstep with digestBatch in
// tools/cmd/fixturegen, which computes the same walk over generator values.
func Digest(b *model.IngestBatch) uint64 {
	d := &digest{h: fnv.New64a()}
	d.s(b.BatchID)
	d.s(b.Source.Host)
	d.s(b.Source.Region)
	d.n(len(b.Source.Tags))
	for _, t := range b.Source.Tags {
		d.s(t)
	}
	d.s(b.CreatedAt)
	d.n(len(b.Events))
	for i := range b.Events {
		e := &b.Events[i]
		d.i(e.ID)
		d.s(e.Type)
		d.s(e.TS)
		d.i(e.User.ID)
		d.s(e.User.Name)
		d.s(e.User.Email)
		d.b(e.User.Verified)
		if e.User.Locale == nil {
			d.null()
		} else {
			d.s(*e.User.Locale)
		}
		d.f(e.Amount)
		d.s(e.Currency)
		d.n(len(e.Items))
		for j := range e.Items {
			it := &e.Items[j]
			d.s(it.SKU)
			d.i(int64(it.Qty))
			d.f(it.UnitPrice)
			if it.Attrs == nil {
				d.null()
			} else {
				d.n(len(it.Attrs))
				ks := make([]string, 0, len(it.Attrs))
				for k := range it.Attrs {
					ks = append(ks, k)
				}
				sort.Strings(ks)
				for _, k := range ks {
					d.s(k)
					d.s(it.Attrs[k])
				}
			}
		}
		if e.Geo == nil {
			d.null()
		} else {
			d.f(e.Geo.Lat)
			d.f(e.Geo.Lon)
		}
		d.s(e.Note)
		d.n(len(e.Flags))
		for _, f := range e.Flags {
			d.b(f)
		}
	}
	return d.h.Sum64()
}

type digest struct{ h hash.Hash64 }

func (d *digest) tag(t byte) { d.h.Write([]byte{t}) }
func (d *digest) u(v uint64) {
	var b [8]byte
	for i := range 8 {
		b[i] = byte(v >> (8 * i))
	}
	d.h.Write(b[:])
}
func (d *digest) s(v string) {
	d.tag('s')
	d.u(uint64(len(v)))
	d.h.Write([]byte(v))
}
func (d *digest) i(v int64)   { d.tag('i'); d.u(uint64(v)) }
func (d *digest) f(v float64) { d.tag('f'); d.u(math.Float64bits(v)) }
func (d *digest) b(v bool) {
	d.tag('b')
	if v {
		d.tag(1)
	} else {
		d.tag(0)
	}
}
func (d *digest) null()   { d.tag('n') }
func (d *digest) n(c int) { d.tag('#'); d.u(uint64(c)) }

// Golden is the parsed <fixture>.golden.tsv written by fixturegen.
type Golden map[string]string

// LoadGolden parses a tab-separated key/value golden file.
func LoadGolden(path string) (Golden, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	g := Golden{}
	sc := bufio.NewScanner(bytes.NewReader(data))
	for sc.Scan() {
		line := sc.Text()
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "\t")
		if !ok {
			return nil, fmt.Errorf("%s: bad line %q", path, line)
		}
		g[k] = v
	}
	return g, sc.Err()
}
