// Package codec isolates the single intentional difference between the two
// server binaries: the JSON decoder. Implementations live in sub-packages so
// that each binary links exactly one JSON library on the hot path.
package codec

// Codec decodes a complete, already-buffered JSON document into v.
//
// Implementations MUST call only the library's plain package-level Unmarshal:
// no options, no context variants, no NoEscape variants, no pooling, no Decoder.
type Codec interface {
	Unmarshal(data []byte, v any) error
	Name() string
}
