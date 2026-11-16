// Package goccy adapts github.com/goccy/go-json.
//
// Note (Phase-0 source review, v0.12.0): plain Unmarshal copies the input into a
// per-call context buffer taken from the library's own internal sync.Pool
// (decoder.TakeRuntimeContext / SetInput) and validates trailing bytes after the
// value (validateEndBuf). That internal pooling is part of the library under test;
// this adapter adds nothing on top.
package goccy

import json "github.com/goccy/go-json"

// Codec is the go-json adapter.
type Codec struct{}

// Unmarshal calls json.Unmarshal and nothing else.
func (Codec) Unmarshal(data []byte, v any) error { return json.Unmarshal(data, v) }

// Name identifies the arm.
func (Codec) Name() string { return "goccy" }
