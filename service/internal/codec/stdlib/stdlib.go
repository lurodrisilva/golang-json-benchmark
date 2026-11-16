// Package stdlib adapts encoding/json (whichever engine the toolchain's
// GOEXPERIMENT selects: v2-backed by default on go1.27, legacy with nojsonv2).
package stdlib

import "encoding/json"

// Codec is the encoding/json adapter.
type Codec struct{}

// Unmarshal calls json.Unmarshal and nothing else.
func (Codec) Unmarshal(data []byte, v any) error { return json.Unmarshal(data, v) }

// Name identifies the arm.
func (Codec) Name() string { return "stdlib" }
