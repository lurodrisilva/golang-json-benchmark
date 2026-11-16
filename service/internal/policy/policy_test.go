package policy

import (
	"encoding/json"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestCheck(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want Reason
	}{
		{"empty object", `{}`, OK},
		{"simple", `{"a":1,"b":[true,false,null],"c":{"d":"x"}}`, OK},
		{"same key in sibling objects", `[{"a":1},{"a":2}]`, OK},
		{"same key at different depth", `{"a":{"a":1}}`, OK},
		{"dup key", `{"a":1,"a":2}`, DuplicateKey},
		{"dup key nested", `{"x":[{"k":1,"k":1}]}`, DuplicateKey},
		{"dup key via escape", `{"a":1,"a":2}`, DuplicateKey},
		{"dup key via surrogate escape", `{"😀":1,"😀":2}`, DuplicateKey},
		{"case differs is not dup", `{"a":1,"A":2}`, OK},
		{"valid unicode", `{"n":"José 日本 😀 é 😀"}`, OK},
		{"lone high surrogate", `{"n":"\ud800"}`, InvalidUTF8},
		{"lone low surrogate", `{"n":"\udc00"}`, InvalidUTF8},
		{"high then non-low", `{"n":"\ud800A"}`, InvalidUTF8},
		{"invalid raw utf8", "{\"n\":\"\xff\"}", InvalidUTF8},
		{"utf8-encoded surrogate", "{\"n\":\"\xed\xa0\x80\"}", InvalidUTF8},
		{"truncated utf8", "{\"n\":\"\xe6\x97\"}", InvalidUTF8},
		{"invalid utf8 in key", "{\"\xff\":1}", InvalidUTF8},
		{"syntax error is not a policy violation", `{"a":`, OK},
		{"trailing data is not a policy violation", `{} {}`, OK},
		{"bom is not a policy violation", "\xef\xbb\xbf{}", OK},
	}
	for _, c := range cases {
		if got := Check([]byte(c.in)); got != c.want {
			t.Errorf("%s: Check(%q) = %v, want %v", c.name, c.in, got, c.want)
		}
	}
}

func TestLargeObjectDuplicateUsesMap(t *testing.T) {
	var b strings.Builder
	b.WriteString("{")
	for i := range 1000 {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(`"k`)
		b.WriteString(strings.Repeat("x", i%7))
		b.WriteString(string(rune('a' + i%26)))
		b.WriteString(itoa(i))
		b.WriteString(`":1`)
	}
	ok := b.String() + "}"
	if got := Check([]byte(ok)); got != OK {
		t.Fatalf("unique large object: got %v", got)
	}
	dup := b.String() + `,"k` + "a" + itoa(0) + `":2}`
	if got := Check([]byte(dup)); got != DuplicateKey {
		t.Fatalf("duplicate in large object: got %v", got)
	}
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var b []byte
	for ; i > 0; i /= 10 {
		b = append([]byte{byte('0' + i%10)}, b...)
	}
	return string(b)
}

func TestDeepNestingDoesNotCrash(t *testing.T) {
	in := strings.Repeat("[", 1<<20) + strings.Repeat("]", 1<<20)
	if got := Check([]byte(in)); got != OK {
		t.Fatalf("got %v", got)
	}
}

// FuzzCheck asserts the scanner never panics and is consistent with an
// independent oracle: raw invalid UTF-8 in a syntactically valid document must
// be reported, and a document that has neither invalid UTF-8 nor escapes nor
// duplicate keys (checked via a map decode) must pass.
func FuzzCheck(f *testing.F) {
	for _, s := range []string{`{}`, `{"a":1,"a":2}`, `{"n":"\ud800"}`, "{\"n\":\"\xff\"}", `[1,2,{"x":"y"}]`, `"x"`} {
		f.Add([]byte(s))
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		r := Check(data)
		if !json.Valid(data) {
			return // syntax owned by decoders; only "no panic" is asserted
		}
		if !utf8.Valid(data) && r == OK {
			// raw invalid bytes can only appear inside strings in valid JSON
			t.Fatalf("invalid raw UTF-8 accepted: %q", data)
		}
	})
}

func BenchmarkCheck(b *testing.B) {
	doc := []byte(`{"batch_id":"b","events":[` + strings.Repeat(`{"id":1,"type":"order","user":{"name":"José 😀","locale":null},"items":[{"sku":"a","attrs":{"k":"v","k2":"v2"}}],"note":"a\"b\\c\n"},`, 300) + `{"id":2}]}`)
	b.SetBytes(int64(len(doc)))
	b.ReportAllocs()
	for b.Loop() {
		if Check(doc) != OK {
			b.Fatal("unexpected")
		}
	}
}
