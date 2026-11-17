// Package codec isolates the single intentional difference between the two
// server binaries: the JSON decoder. Implementations live in sub-packages so
// that each binary links exactly one JSON library on the hot path.
package codec

// Codec decodes a complete, already-buffered JSON document into v, and
// encodes v into a fresh byte slice.
//
// Implementations MUST call only the library's plain package-level Unmarshal
// and Marshal: no options, no context variants, no NoEscape variants, no
// pooling, no Decoder/Encoder.
type Codec interface {
	Unmarshal(data []byte, v any) error
	Marshal(v any) ([]byte, error)
	Name() string
}
