// Command fixturegen deterministically generates the SYNTHETIC benchmark
// fixtures and their golden expectations.
//
// Values are drawn from math/rand/v2 PCG seeded with -seed. JSON is produced
// by a hand-written compact emitter (no JSON library), with fixed field order.
// The golden file is computed from the generator's own values (never from a
// decoder), so tests compare each decoder against an independent oracle.
//
// All fixtures share one event stream: a larger fixture is the smaller one
// with more events appended, so F1→F2→F3 is a pure size sweep.
//
//	fixturegen -out ../fixtures -seed 20261115
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"hash/fnv"
	"math"
	"math/rand/v2"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"unicode/utf8"
)

type item struct {
	sku       string
	qty       int32
	unitPrice float64
	attrs     map[string]string
}

type geo struct{ lat, lon float64 }

type user struct {
	id       int64
	name     string
	email    string
	verified bool
	locale   *string
}

type event struct {
	id       int64
	typ      string
	ts       string
	user     user
	amount   float64
	currency string
	items    []item
	geo      *geo
	note     string
	flags    []bool
}

type batch struct {
	batchID   string
	host      string
	region    string
	tags      []string
	createdAt string
	events    []event
}

var (
	types      = []string{"order", "refund", "view", "signup", "payment", "cart_update"}
	firstNames = []string{"José", "Zoë", "Ægir", "Björk", "李雷", "さくら", "Мария", "محمد", "Søren", "Nguyễn", "François", "Ana", "John", "Priya", "Łukasz", "Ömer"}
	lastNames  = []string{"Silva", "Müller", "O'Brien", "王", "山田", "Иванова", "Jørgensen", "García", "Smith", "Kowalski", "Đặng", "Çelik"}
	locales    = []string{"pt-BR", "en-US", "ja-JP", "de-DE", "fr-FR", "zh-CN", "ar-EG"}
	currencies = []string{"EUR", "USD", "BRL", "JPY", "GBP"}
	attrKeys   = []string{"color", "size", "material", "región", "variant", "warehouse"}
	attrVals   = []string{"red", "XL", "cotton", "azul-marinho", "日本製", "v2", "A/B", "50% off", "tab\there"}
	notes      = []string{
		`Customer said "deliver after 6pm"`,
		"Line one\nLine two",
		`Path C:\orders\2026\11`,
		"Gift wrap 🎁 please",
		"配達は午前中にお願いします",
		"Доставка до двери",
		"Emoji mix 😀👍🏽🇧🇷",
		"Tab\tseparated\tvalues",
		"</script> & <b>bold</b>",
		"Café — naïve résumé",
		"",
		"control \u0001 char",
	}
)

type gen struct{ r *rand.Rand }

func (g *gen) pick(s []string) string { return s[g.r.IntN(len(s))] }

func (g *gen) money(maxCents int) float64 {
	return float64(g.r.IntN(maxCents)+1) / 100
}

func (g *gen) event(i int) event {
	e := event{
		id:  int64(1_000_000 + i),
		typ: g.pick(types),
		ts:  fmt.Sprintf("2026-11-%02dT%02d:%02d:%02d.%03dZ", 1+i%28, g.r.IntN(24), g.r.IntN(60), g.r.IntN(60), g.r.IntN(1000)),
	}
	if i%97 == 0 {
		e.id = 9_007_199_254_740_993 + int64(i) // beyond 2^53: must decode exactly as int64
	}
	e.user = user{
		id:       int64(g.r.IntN(5_000_000) + 1),
		name:     g.pick(firstNames) + " " + g.pick(lastNames),
		verified: g.r.IntN(2) == 0,
	}
	e.user.email = fmt.Sprintf("user%d@example.com", e.user.id)
	if g.r.IntN(10) >= 3 {
		l := g.pick(locales)
		e.user.locale = &l
	}
	e.currency = g.pick(currencies)
	n := 1 + g.r.IntN(4)
	for j := range n {
		it := item{
			sku:       fmt.Sprintf("SKU-%05d-%c", g.r.IntN(100000), 'A'+rune(j)),
			qty:       int32(1 + g.r.IntN(20)),
			unitPrice: g.money(50000),
		}
		if na := g.r.IntN(4); na > 0 {
			it.attrs = make(map[string]string, na)
			for range na {
				it.attrs[g.pick(attrKeys)] = g.pick(attrVals)
			}
		}
		e.amount += float64(it.qty) * it.unitPrice
		e.items = append(e.items, it)
	}
	e.amount = math.Round(e.amount*100) / 100
	if g.r.IntN(5) != 0 {
		e.geo = &geo{
			lat: math.Round((g.r.Float64()*180-90)*1e6) / 1e6,
			lon: math.Round((g.r.Float64()*360-180)*1e6) / 1e6,
		}
	}
	e.note = g.pick(notes)
	for range g.r.IntN(5) {
		e.flags = append(e.flags, g.r.IntN(2) == 0)
	}
	return e
}

// ---- emitter ----

type emitter struct {
	buf  bytes.Buffer
	nstr int // string counter: drives deterministic escaping style
}

// str writes a JSON string. Every 5th string escapes all non-ASCII runes as
// \uXXXX (surrogate pairs above the BMP); every 11th escapes '/' as "\/".
func (w *emitter) str(s string) {
	w.nstr++
	escapeNonASCII := w.nstr%5 == 0
	escapeSlash := w.nstr%11 == 0
	b := &w.buf
	b.WriteByte('"')
	for _, r := range s {
		switch {
		case r == '"':
			b.WriteString(`\"`)
		case r == '\\':
			b.WriteString(`\\`)
		case r == '\n':
			b.WriteString(`\n`)
		case r == '\t':
			b.WriteString(`\t`)
		case r == '/' && escapeSlash:
			b.WriteString(`\/`)
		case r < 0x20:
			fmt.Fprintf(b, `\u%04x`, r)
		case r >= 0x80 && escapeNonASCII:
			if r > 0xFFFF {
				r -= 0x10000
				fmt.Fprintf(b, `\u%04x\u%04x`, 0xD800+(r>>10), 0xDC00+(r&0x3FF))
			} else {
				fmt.Fprintf(b, `\u%04x`, r)
			}
		default:
			b.WriteRune(r)
		}
	}
	b.WriteByte('"')
}

func (w *emitter) key(k string, first bool) {
	if !first {
		w.buf.WriteByte(',')
	}
	w.str(k)
	w.buf.WriteByte(':')
}

func (w *emitter) num(v int64)    { w.buf.WriteString(strconv.FormatInt(v, 10)) }
func (w *emitter) flt(v float64)  { w.buf.WriteString(strconv.FormatFloat(v, 'g', -1, 64)) }
func (w *emitter) boolean(v bool) { w.buf.WriteString(strconv.FormatBool(v)) }

func (w *emitter) strs(s []string) {
	w.buf.WriteByte('[')
	for i, v := range s {
		if i > 0 {
			w.buf.WriteByte(',')
		}
		w.str(v)
	}
	w.buf.WriteByte(']')
}

func (w *emitter) event(e *event) {
	b := &w.buf
	b.WriteByte('{')
	w.key("id", true)
	w.num(e.id)
	w.key("type", false)
	w.str(e.typ)
	w.key("ts", false)
	w.str(e.ts)
	w.key("user", false)
	b.WriteByte('{')
	w.key("id", true)
	w.num(e.user.id)
	w.key("name", false)
	w.str(e.user.name)
	w.key("email", false)
	w.str(e.user.email)
	w.key("verified", false)
	w.boolean(e.user.verified)
	w.key("locale", false)
	if e.user.locale == nil {
		b.WriteString("null")
	} else {
		w.str(*e.user.locale)
	}
	b.WriteByte('}')
	w.key("amount", false)
	w.flt(e.amount)
	w.key("currency", false)
	w.str(e.currency)
	w.key("items", false)
	b.WriteByte('[')
	for i := range e.items {
		it := &e.items[i]
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteByte('{')
		w.key("sku", true)
		w.str(it.sku)
		w.key("qty", false)
		w.num(int64(it.qty))
		w.key("unit_price", false)
		w.flt(it.unitPrice)
		w.key("attrs", false)
		if it.attrs == nil {
			b.WriteString("null")
		} else {
			b.WriteByte('{')
			for j, k := range sortedKeys(it.attrs) {
				w.key(k, j == 0)
				w.str(it.attrs[k])
			}
			b.WriteByte('}')
		}
		b.WriteByte('}')
	}
	b.WriteByte(']')
	w.key("geo", false)
	if e.geo == nil {
		b.WriteString("null")
	} else {
		b.WriteByte('{')
		w.key("lat", true)
		w.flt(e.geo.lat)
		w.key("lon", false)
		w.flt(e.geo.lon)
		b.WriteByte('}')
	}
	w.key("note", false)
	w.str(e.note)
	w.key("flags", false)
	if e.flags == nil {
		b.WriteString("[]")
	} else {
		b.WriteByte('[')
		for i, f := range e.flags {
			if i > 0 {
				b.WriteByte(',')
			}
			w.boolean(f)
		}
		b.WriteByte(']')
	}
	b.WriteByte('}')
}

func sortedKeys(m map[string]string) []string {
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	return ks
}

// ---- golden digest ----
//
// Canonical walk over every model field, type-tagged, length-prefixed, FNV-1a
// 64. service/internal/model/digest_test.go implements the identical walk over
// the DECODED Go value. Keep the two in lockstep (docs/fixtures.md).

type digest struct {
	h interface {
		Write([]byte) (int, error)
		Sum64() uint64
	}
}

func (d *digest) tag(t byte) { d.h.Write([]byte{t}) }
func (d *digest) s(v string) {
	d.tag('s')
	d.u(uint64(len(v)))
	d.h.Write([]byte(v))
}
func (d *digest) u(v uint64) {
	var b [8]byte
	for i := range 8 {
		b[i] = byte(v >> (8 * i))
	}
	d.h.Write(b[:])
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

func digestBatch(bt *batch) uint64 {
	d := &digest{h: fnv.New64a()}
	d.s(bt.batchID)
	d.s(bt.host)
	d.s(bt.region)
	d.n(len(bt.tags))
	for _, t := range bt.tags {
		d.s(t)
	}
	d.s(bt.createdAt)
	d.n(len(bt.events))
	for i := range bt.events {
		e := &bt.events[i]
		d.i(e.id)
		d.s(e.typ)
		d.s(e.ts)
		d.i(e.user.id)
		d.s(e.user.name)
		d.s(e.user.email)
		d.b(e.user.verified)
		if e.user.locale == nil {
			d.null()
		} else {
			d.s(*e.user.locale)
		}
		d.f(e.amount)
		d.s(e.currency)
		d.n(len(e.items))
		for j := range e.items {
			it := &e.items[j]
			d.s(it.sku)
			d.i(int64(it.qty))
			d.f(it.unitPrice)
			if it.attrs == nil {
				d.null()
			} else {
				d.n(len(it.attrs))
				for _, k := range sortedKeys(it.attrs) {
					d.s(k)
					d.s(it.attrs[k])
				}
			}
		}
		if e.geo == nil {
			d.null()
		} else {
			d.f(e.geo.lat)
			d.f(e.geo.lon)
		}
		d.s(e.note)
		d.n(len(e.flags))
		for _, f := range e.flags {
			d.b(f)
		}
	}
	return d.h.Sum64()
}

type stats struct {
	events, items, attrs, nullLocale, nullGeo, nullAttrs, flags, bigIDs int
	nonASCIIRunes                                                       int
	sumQty                                                              int64
}

func collect(bt *batch) stats {
	var s stats
	count := func(v string) {
		for _, r := range v {
			if r >= 0x80 {
				s.nonASCIIRunes++
			}
		}
	}
	s.events = len(bt.events)
	for i := range bt.events {
		e := &bt.events[i]
		count(e.user.name)
		count(e.note)
		if e.id > 1<<53 {
			s.bigIDs++
		}
		if e.user.locale == nil {
			s.nullLocale++
		}
		if e.geo == nil {
			s.nullGeo++
		}
		s.flags += len(e.flags)
		for _, it := range e.items {
			s.items++
			s.sumQty += int64(it.qty)
			if it.attrs == nil {
				s.nullAttrs++
			}
			s.attrs += len(it.attrs)
			for k, v := range it.attrs {
				count(k)
				count(v)
			}
		}
	}
	return s
}

// ---- main ----

type target struct {
	name  string
	bytes int
}

func main() {
	out := flag.String("out", "fixtures", "output directory")
	seed := flag.Uint64("seed", 20261115, "PCG seed")
	flag.Parse()

	targets := []target{{"f1-64k", 64 << 10}, {"f2-256k", 256 << 10}, {"f3-1m", 1 << 20}}
	if err := run(*out, *seed, targets); err != nil {
		fmt.Fprintln(os.Stderr, "fixturegen:", err)
		os.Exit(1)
	}
}

func run(out string, seed uint64, targets []target) error {
	if err := os.MkdirAll(out, 0o755); err != nil {
		return err
	}
	var sums bytes.Buffer
	for _, t := range targets {
		data, bt, err := build(seed, t)
		if err != nil {
			return err
		}
		if !utf8.Valid(data) {
			return fmt.Errorf("%s: emitter produced invalid UTF-8", t.name)
		}
		if err := os.WriteFile(filepath.Join(out, t.name+".json"), data, 0o644); err != nil {
			return err
		}
		st := collect(bt)
		var g bytes.Buffer
		fmt.Fprintf(&g, "# SYNTHETIC fixture golden — generated by tools/cmd/fixturegen; do not edit\n")
		fmt.Fprintf(&g, "fixture\t%s\nseed\t%d\nbytes\t%d\n", t.name, seed, len(data))
		fmt.Fprintf(&g, "digest_fnv64a\t%016x\n", digestBatch(bt))
		fmt.Fprintf(&g, "batch_id\t%s\nevents\t%d\nitems\t%d\nattrs\t%d\nnull_locale\t%d\nnull_geo\t%d\nnull_attrs\t%d\nflags\t%d\nbig_ids\t%d\nsum_qty\t%d\nnon_ascii_runes\t%d\n",
			bt.batchID, st.events, st.items, st.attrs, st.nullLocale, st.nullGeo, st.nullAttrs, st.flags, st.bigIDs, st.sumQty, st.nonASCIIRunes)
		e0, en := &bt.events[0], &bt.events[len(bt.events)-1]
		fmt.Fprintf(&g, "first_id\t%d\nfirst_user_name\t%s\nfirst_note\t%s\nlast_id\t%d\nlast_amount\t%s\n",
			e0.id, e0.user.name, strconv.Quote(e0.note), en.id, strconv.FormatFloat(en.amount, 'g', -1, 64))
		if err := os.WriteFile(filepath.Join(out, t.name+".golden.tsv"), g.Bytes(), 0o644); err != nil {
			return err
		}
		for _, f := range []string{t.name + ".json", t.name + ".golden.tsv"} {
			b, _ := os.ReadFile(filepath.Join(out, f))
			h := sha256.Sum256(b)
			fmt.Fprintf(&sums, "%s  %s\n", hex.EncodeToString(h[:]), f)
		}
		fmt.Printf("%s: %d bytes (%.1f KiB), %d events\n", t.name, len(data), float64(len(data))/1024, len(bt.events))
	}
	return os.WriteFile(filepath.Join(out, "SHA256SUMS"), sums.Bytes(), 0o644)
}

// build emits events from a freshly seeded stream until the document reaches
// the target size (the first event that crosses it is the last one).
func build(seed uint64, t target) ([]byte, *batch, error) {
	g := &gen{r: rand.New(rand.NewPCG(seed, seed^0x9e3779b97f4a7c15))}
	bt := &batch{
		batchID:   "batch-" + strconv.FormatUint(seed, 10),
		host:      "ingest-01.eu-west.example.internal",
		region:    "eu-west-1",
		tags:      []string{"synthetic", "benchmark", "região-sul", "v1"},
		createdAt: "2026-11-15T00:00:00Z",
	}
	w := &emitter{}
	head := func() {
		w.buf.WriteByte('{')
		w.key("batch_id", true)
		w.str(bt.batchID)
		w.key("source", false)
		w.buf.WriteByte('{')
		w.key("host", true)
		w.str(bt.host)
		w.key("region", false)
		w.str(bt.region)
		w.key("tags", false)
		w.strs(bt.tags)
		w.buf.WriteByte('}')
		w.key("created_at", false)
		w.str(bt.createdAt)
		w.key("events", false)
		w.buf.WriteByte('[')
	}
	head()
	for i := 0; w.buf.Len()+2 < t.bytes; i++ {
		if i > 0 {
			w.buf.WriteByte(',')
		}
		e := g.event(i)
		w.event(&e)
		bt.events = append(bt.events, e)
		if i > 1_000_000 {
			return nil, nil, fmt.Errorf("%s: runaway generation", t.name)
		}
	}
	w.buf.WriteString("]}")
	return w.buf.Bytes(), bt, nil
}
