// Package policy enforces the project's input policy (docs/correctness-policy.md)
// independently of either JSON library:
//
//   - strict UTF-8: raw invalid UTF-8 and \u escapes that do not form valid
//     scalar values (lone surrogates) are rejected;
//   - no duplicate member names within one object (compared after unescaping,
//     case-sensitively).
//
// The scanner is single-pass and does not decode into Go values. It does not
// decide syntax validity: when it meets a syntax error it stops and returns OK,
// leaving the decoder under test to reject the document. This keeps the error
// classification for malformed input owned by the decoder and keeps the policy
// strictly additive and identical in both arms.
package policy

import (
	"unicode/utf8"
)

// Reason is the outcome of Check.
type Reason uint8

const (
	// OK means no policy violation was found (the document may still be syntactically invalid).
	OK Reason = iota
	// InvalidUTF8 means a string contained invalid UTF-8 or a lone surrogate escape.
	InvalidUTF8
	// DuplicateKey means an object contained the same member name twice.
	DuplicateKey
)

func (r Reason) String() string {
	switch r {
	case OK:
		return "ok"
	case InvalidUTF8:
		return "utf8"
	case DuplicateKey:
		return "dupkey"
	}
	return "unknown"
}

// MaxDepth bounds nesting. Deeper documents are left to the decoder (both
// libraries reject them); the scanner simply stops and returns OK.
const MaxDepth = 10000

// linearKeys is the object size up to which duplicate detection uses a linear
// scan; larger objects switch to a map so adversarial inputs stay O(n).
const linearKeys = 32

type span struct {
	off, end int32
	scratch  bool
}

type scanner struct {
	data    []byte
	i       int
	depth   int
	reason  Reason
	keys    []span
	scratch []byte
}

// Check scans data and reports the first policy violation, if any.
func Check(data []byte) Reason {
	var keyBuf [64]span
	s := scanner{data: data, keys: keyBuf[:0]}
	s.value()
	return s.reason
}

func (s *scanner) ws() {
	for s.i < len(s.data) {
		switch s.data[s.i] {
		case ' ', '\t', '\n', '\r':
			s.i++
		default:
			return
		}
	}
}

// value parses one value. It returns false when scanning must stop
// (policy violation recorded in s.reason, or syntax error).
func (s *scanner) value() bool {
	s.ws()
	if s.i >= len(s.data) {
		return false
	}
	switch c := s.data[s.i]; {
	case c == '{':
		return s.object()
	case c == '[':
		return s.array()
	case c == '"':
		_, ok := s.str()
		return ok
	case c == 't':
		return s.lit("true")
	case c == 'f':
		return s.lit("false")
	case c == 'n':
		return s.lit("null")
	case c == '-' || (c >= '0' && c <= '9'):
		return s.number()
	}
	return false
}

func (s *scanner) lit(w string) bool {
	if len(s.data)-s.i < len(w) || string(s.data[s.i:s.i+len(w)]) != w {
		return false
	}
	s.i += len(w)
	return true
}

func isDigit(c byte) bool { return c >= '0' && c <= '9' }

func (s *scanner) number() bool {
	d := s.data
	if d[s.i] == '-' {
		s.i++
	}
	if s.i >= len(d) {
		return false
	}
	if d[s.i] == '0' {
		s.i++
	} else if isDigit(d[s.i]) {
		for s.i < len(d) && isDigit(d[s.i]) {
			s.i++
		}
	} else {
		return false
	}
	if s.i < len(d) && d[s.i] == '.' {
		s.i++
		if s.i >= len(d) || !isDigit(d[s.i]) {
			return false
		}
		for s.i < len(d) && isDigit(d[s.i]) {
			s.i++
		}
	}
	if s.i < len(d) && (d[s.i] == 'e' || d[s.i] == 'E') {
		s.i++
		if s.i < len(d) && (d[s.i] == '+' || d[s.i] == '-') {
			s.i++
		}
		if s.i >= len(d) || !isDigit(d[s.i]) {
			return false
		}
		for s.i < len(d) && isDigit(d[s.i]) {
			s.i++
		}
	}
	return true
}

func (s *scanner) array() bool {
	if s.depth++; s.depth > MaxDepth {
		return false
	}
	s.i++ // '['
	s.ws()
	if s.i < len(s.data) && s.data[s.i] == ']' {
		s.i++
		s.depth--
		return true
	}
	for {
		if !s.value() {
			return false
		}
		s.ws()
		if s.i >= len(s.data) {
			return false
		}
		switch s.data[s.i] {
		case ',':
			s.i++
		case ']':
			s.i++
			s.depth--
			return true
		default:
			return false
		}
	}
}

func (s *scanner) keyBytes(k span) []byte {
	if k.scratch {
		return s.scratch[k.off:k.end]
	}
	return s.data[k.off:k.end]
}

func (s *scanner) object() bool {
	if s.depth++; s.depth > MaxDepth {
		return false
	}
	s.i++ // '{'
	s.ws()
	if s.i < len(s.data) && s.data[s.i] == '}' {
		s.i++
		s.depth--
		return true
	}
	mark, scratchMark := len(s.keys), len(s.scratch)
	var seen map[string]struct{}
	defer func() { s.keys, s.scratch = s.keys[:mark], s.scratch[:scratchMark] }()
	for {
		s.ws()
		if s.i >= len(s.data) || s.data[s.i] != '"' {
			return false
		}
		k, ok := s.str()
		if !ok {
			return false
		}
		kb := s.keyBytes(k)
		if seen != nil {
			if _, dup := seen[string(kb)]; dup {
				s.reason = DuplicateKey
				return false
			}
			seen[string(kb)] = struct{}{}
		} else {
			for _, prev := range s.keys[mark:] {
				if string(s.keyBytes(prev)) == string(kb) {
					s.reason = DuplicateKey
					return false
				}
			}
			s.keys = append(s.keys, k)
			if len(s.keys)-mark > linearKeys {
				seen = make(map[string]struct{}, 2*linearKeys)
				for _, prev := range s.keys[mark:] {
					seen[string(s.keyBytes(prev))] = struct{}{}
				}
			}
		}
		s.ws()
		if s.i >= len(s.data) || s.data[s.i] != ':' {
			return false
		}
		s.i++
		if !s.value() {
			return false
		}
		s.ws()
		if s.i >= len(s.data) {
			return false
		}
		switch s.data[s.i] {
		case ',':
			s.i++
		case '}':
			s.i++
			s.depth--
			return true
		default:
			return false
		}
	}
}

func hex4(b []byte) (rune, bool) {
	if len(b) < 4 {
		return 0, false
	}
	var r rune
	for _, c := range b[:4] {
		switch {
		case c >= '0' && c <= '9':
			c -= '0'
		case c >= 'a' && c <= 'f':
			c = c - 'a' + 10
		case c >= 'A' && c <= 'F':
			c = c - 'A' + 10
		default:
			return 0, false
		}
		r = r<<4 | rune(c)
	}
	return r, true
}

// str scans a string starting at '"'. It returns the span of the unescaped
// content (into data when there were no escapes, into scratch otherwise).
func (s *scanner) str() (span, bool) {
	d := s.data
	s.i++ // opening quote
	start := s.i
	escaped := false
	for s.i < len(d) {
		c := d[s.i]
		switch {
		case c == '"':
			end := s.i
			s.i++
			if !escaped {
				return span{off: int32(start), end: int32(end)}, true
			}
			return s.unescape(d[start:end]), true
		case c == '\\':
			escaped = true
			if s.i+1 >= len(d) {
				return span{}, false
			}
			switch d[s.i+1] {
			case '"', '\\', '/', 'b', 'f', 'n', 'r', 't':
				s.i += 2
			case 'u':
				r, ok := hex4(d[s.i+2:])
				if !ok {
					return span{}, false
				}
				s.i += 6
				if r >= 0xD800 && r < 0xDC00 { // high surrogate: must be followed by \u low surrogate
					if s.i+1 < len(d) && d[s.i] == '\\' && d[s.i+1] == 'u' {
						lo, ok := hex4(d[s.i+2:])
						if !ok {
							return span{}, false
						}
						if lo >= 0xDC00 && lo < 0xE000 {
							s.i += 6
							continue
						}
					}
					s.reason = InvalidUTF8
					return span{}, false
				}
				if r >= 0xDC00 && r < 0xE000 { // lone low surrogate
					s.reason = InvalidUTF8
					return span{}, false
				}
			default:
				return span{}, false
			}
		case c < 0x20:
			return span{}, false
		case c < utf8.RuneSelf:
			s.i++
		default:
			r, size := utf8.DecodeRune(d[s.i:])
			if r == utf8.RuneError && size == 1 {
				s.reason = InvalidUTF8
				return span{}, false
			}
			s.i += size
		}
	}
	return span{}, false
}

// unescape writes the decoded form of an already validated string body into scratch.
func (s *scanner) unescape(b []byte) span {
	off := len(s.scratch)
	for i := 0; i < len(b); {
		c := b[i]
		if c != '\\' {
			s.scratch = append(s.scratch, c)
			i++
			continue
		}
		switch b[i+1] {
		case 'b':
			s.scratch = append(s.scratch, '\b')
		case 'f':
			s.scratch = append(s.scratch, '\f')
		case 'n':
			s.scratch = append(s.scratch, '\n')
		case 'r':
			s.scratch = append(s.scratch, '\r')
		case 't':
			s.scratch = append(s.scratch, '\t')
		case 'u':
			r, _ := hex4(b[i+2:])
			i += 6
			if r >= 0xD800 && r < 0xDC00 {
				lo, _ := hex4(b[i+2:])
				r = (r-0xD800)<<10 | (lo - 0xDC00) + 0x10000
				i += 6
			}
			s.scratch = utf8.AppendRune(s.scratch, r)
			continue
		default: // '"', '\\', '/'
			s.scratch = append(s.scratch, b[i+1])
		}
		i += 2
	}
	return span{off: int32(off), end: int32(len(s.scratch)), scratch: true}
}
