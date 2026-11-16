package conformance

import (
	"strings"

	"github.com/lurodrisilva/golang-json-encoding-benchmark/service/internal/model"
)

// Spec is the outcome expected from encoding/json v1 DOCUMENTED semantics
// (pkg.go.dev/encoding/json, Unmarshal) for a raw Unmarshal into
// model.IngestBatch — NOT what either library happens to do. Each codec is
// compared to this spec independently; deviations are compatibility findings.
type Spec uint8

const (
	SpecAccept Spec = iota // Unmarshal returns nil
	SpecReject             // Unmarshal returns an error
)

func (s Spec) String() string {
	if s == SpecAccept {
		return "accept"
	}
	return "reject"
}

// Case is one correctness-corpus entry.
type Case struct {
	Name string
	Body string
	// Raw is the v1-spec outcome of a direct Unmarshal (no policy layer).
	Raw Spec
	// Check, when non-nil, asserts decoded values after a raw accept and
	// returns a description of the mismatch ("" when it matches the spec).
	Check func(*model.IngestBatch) string
	// HTTP is the agreed /ingest status with the policy layer (identical for
	// both servers; docs/correctness-policy.md).
	HTTP int
	Note string
}

const ok1 = `{"batch_id":"b","events":[{"id":1}]}`

func deep(n int) string {
	return `{"batch_id":"b","events":[{"id":1}],"x":` + strings.Repeat("[", n) + strings.Repeat("]", n) + `}`
}

func want(cond bool, msg string) string {
	if cond {
		return ""
	}
	return msg
}

// Corpus is the correctness corpus. It never feeds performance runs.
var Corpus = []Case{
	{Name: "minimal_valid", Body: ok1, Raw: SpecAccept, HTTP: 202},
	{Name: "whitespace_trailing", Body: ok1 + " \n\t\r ", Raw: SpecAccept, HTTP: 202, Note: "insignificant whitespace after the value is allowed"},
	{Name: "missing_optional_fields", Body: `{"batch_id":"b","events":[{"id":7}]}`, Raw: SpecAccept,
		Check: func(b *model.IngestBatch) string {
			e := b.Events[0]
			return want(e.User.Locale == nil && e.Geo == nil && e.Items == nil && e.Note == "", "missing fields must keep zero values")
		}, HTTP: 202},
	{Name: "missing_required_batch_id", Body: `{"events":[{"id":1}]}`, Raw: SpecAccept, HTTP: 422, Note: "required-ness is validation, not decoding"},
	{Name: "missing_events", Body: `{"batch_id":"b"}`, Raw: SpecAccept, HTTP: 422},
	{Name: "unknown_fields_ignored", Body: `{"batch_id":"b","zzz":1,"events":[{"id":1,"extra":{"a":[1,{"b":null}]}}]}`, Raw: SpecAccept, HTTP: 202, Note: "v1 ignores unknown fields by default"},
	{Name: "null_into_string", Body: `{"batch_id":null,"events":[{"id":1}]}`, Raw: SpecAccept,
		Check: func(b *model.IngestBatch) string {
			return want(b.BatchID == "", "null into non-pointer must be a no-op")
		}, HTTP: 422},
	{Name: "null_into_int", Body: `{"batch_id":"b","events":[{"id":null}]}`, Raw: SpecAccept, HTTP: 422},
	{Name: "null_into_slice", Body: `{"batch_id":"b","events":null}`, Raw: SpecAccept, HTTP: 422},
	{Name: "null_pointer_fields", Body: `{"batch_id":"b","events":[{"id":1,"user":{"locale":null},"geo":null}]}`, Raw: SpecAccept,
		Check: func(b *model.IngestBatch) string {
			return want(b.Events[0].User.Locale == nil && b.Events[0].Geo == nil, "null pointers must stay nil")
		}, HTTP: 202},
	{Name: "top_level_null", Body: `null`, Raw: SpecAccept, HTTP: 422, Note: "null into struct is a no-op"},
	{Name: "int64_max", Body: `{"batch_id":"b","events":[{"id":9223372036854775807}]}`, Raw: SpecAccept,
		Check: func(b *model.IngestBatch) string {
			return want(b.Events[0].ID == 9223372036854775807, "int64 max must round-trip")
		}, HTTP: 202},
	{Name: "int64_beyond_2pow53", Body: `{"batch_id":"b","events":[{"id":9007199254740993}]}`, Raw: SpecAccept,
		Check: func(b *model.IngestBatch) string {
			return want(b.Events[0].ID == 9007199254740993, "must not round through float64")
		}, HTTP: 202},
	{Name: "int64_min", Body: `{"batch_id":"b","events":[{"id":-9223372036854775808}]}`, Raw: SpecAccept, HTTP: 422},
	{Name: "int64_overflow", Body: `{"batch_id":"b","events":[{"id":9223372036854775808}]}`, Raw: SpecReject, HTTP: 400},
	{Name: "int64_underflow", Body: `{"batch_id":"b","events":[{"id":-9223372036854775809}]}`, Raw: SpecReject, HTTP: 400},
	{Name: "int32_overflow", Body: `{"batch_id":"b","events":[{"id":1,"items":[{"qty":2147483648}]}]}`, Raw: SpecReject, HTTP: 400},
	{Name: "int_with_fraction", Body: `{"batch_id":"b","events":[{"id":1.5}]}`, Raw: SpecReject, HTTP: 400},
	{Name: "int_with_exponent", Body: `{"batch_id":"b","events":[{"id":1e3}]}`, Raw: SpecReject, HTTP: 400, Note: "v1 parses ints with ParseInt; 1e3 is not an integer literal"},
	{Name: "negative_id", Body: `{"batch_id":"b","events":[{"id":-1}]}`, Raw: SpecAccept, HTTP: 422},
	{Name: "float_exponent", Body: `{"batch_id":"b","events":[{"id":1,"amount":1.5e3}]}`, Raw: SpecAccept,
		Check: func(b *model.IngestBatch) string {
			return want(b.Events[0].Amount == 1500, "1.5e3 must decode to 1500")
		}, HTTP: 202},
	{Name: "float_overflow", Body: `{"batch_id":"b","events":[{"id":1,"amount":1e400}]}`, Raw: SpecReject, HTTP: 400},
	{Name: "float_smallest_subnormal", Body: `{"batch_id":"b","events":[{"id":1,"amount":5e-324}]}`, Raw: SpecAccept,
		Check: func(b *model.IngestBatch) string {
			return want(b.Events[0].Amount == 5e-324, "5e-324 must decode exactly")
		}, HTTP: 202},
	{Name: "float_negative_zero", Body: `{"batch_id":"b","events":[{"id":1,"amount":-0.0}]}`, Raw: SpecAccept, HTTP: 202},
	{Name: "leading_zero", Body: `{"batch_id":"b","events":[{"id":01}]}`, Raw: SpecReject, HTTP: 400},
	{Name: "nan_literal", Body: `{"batch_id":"b","events":[{"id":1,"amount":NaN}]}`, Raw: SpecReject, HTTP: 400},
	{Name: "wrong_type_string_for_int", Body: `{"batch_id":"b","events":[{"id":"1"}]}`, Raw: SpecReject, HTTP: 400},
	{Name: "wrong_type_number_for_string", Body: `{"batch_id":1,"events":[{"id":1}]}`, Raw: SpecReject, HTTP: 400},
	{Name: "wrong_type_bool_for_string", Body: `{"batch_id":true,"events":[{"id":1}]}`, Raw: SpecReject, HTTP: 400},
	{Name: "wrong_type_string_for_bool", Body: `{"batch_id":"b","events":[{"id":1,"user":{"verified":"true"}}]}`, Raw: SpecReject, HTTP: 400},
	{Name: "wrong_type_array_for_object", Body: `{"batch_id":"b","source":[],"events":[{"id":1}]}`, Raw: SpecReject, HTTP: 400},
	{Name: "wrong_type_object_for_array", Body: `{"batch_id":"b","events":{"id":1}}`, Raw: SpecReject, HTTP: 400},
	{Name: "top_level_array", Body: `[]`, Raw: SpecReject, HTTP: 400},
	{Name: "top_level_string", Body: `"x"`, Raw: SpecReject, HTTP: 400},
	{Name: "duplicate_key_top", Body: `{"batch_id":"a","batch_id":"b","events":[{"id":1}]}`, Raw: SpecAccept,
		Check: func(b *model.IngestBatch) string { return want(b.BatchID == "b", "v1: last duplicate wins") }, HTTP: 400, Note: "policy rejects duplicates"},
	{Name: "duplicate_key_nested", Body: `{"batch_id":"b","events":[{"id":1,"id":2}]}`, Raw: SpecAccept,
		Check: func(b *model.IngestBatch) string { return want(b.Events[0].ID == 2, "v1: last duplicate wins") }, HTTP: 400},
	{Name: "case_insensitive_match", Body: `{"BATCH_ID":"b","Events":[{"ID":1}]}`, Raw: SpecAccept,
		Check: func(b *model.IngestBatch) string {
			return want(b.BatchID == "b" && len(b.Events) == 1 && b.Events[0].ID == 1, "v1 matches field names case-insensitively")
		}, HTTP: 202},
	{Name: "case_exact_preferred", Body: `{"BATCH_ID":"x","batch_id":"b","events":[{"id":1}]}`, Raw: SpecAccept, HTTP: 202, Note: "value is implementation-defined in v1; only acceptance is asserted"},
	{Name: "malformed_truncated", Body: `{"batch_id":`, Raw: SpecReject, HTTP: 400},
	{Name: "malformed_missing_colon", Body: `{"batch_id" "b"}`, Raw: SpecReject, HTTP: 400},
	{Name: "malformed_trailing_comma_object", Body: `{"batch_id":"b","events":[{"id":1}],}`, Raw: SpecReject, HTTP: 400},
	{Name: "malformed_trailing_comma_array", Body: `{"batch_id":"b","events":[{"id":1},]}`, Raw: SpecReject, HTTP: 400},
	{Name: "malformed_single_quotes", Body: `{'batch_id':'b'}`, Raw: SpecReject, HTTP: 400},
	{Name: "raw_control_char_in_string", Body: "{\"batch_id\":\"a\x01b\",\"events\":[{\"id\":1}]}", Raw: SpecReject, HTTP: 400},
	{Name: "invalid_escape", Body: `{"batch_id":"\x","events":[{"id":1}]}`, Raw: SpecReject, HTTP: 400},
	{Name: "trailing_second_value", Body: `{} {}`, Raw: SpecReject, HTTP: 400},
	{Name: "trailing_garbage", Body: `{}x`, Raw: SpecReject, HTTP: 400},
	{Name: "trailing_after_valid", Body: ok1 + `]`, Raw: SpecReject, HTTP: 400},
	{Name: "empty_body", Body: ``, Raw: SpecReject, HTTP: 400},
	{Name: "whitespace_only", Body: "  \n", Raw: SpecReject, HTTP: 400},
	{Name: "utf8_bom", Body: "\xef\xbb\xbf" + ok1, Raw: SpecReject, HTTP: 400, Note: "BOM is not JSON whitespace"},
	{Name: "unicode_raw_and_escaped", Body: `{"batch_id":"José é 😀 😀 日本","events":[{"id":1}]}`, Raw: SpecAccept,
		Check: func(b *model.IngestBatch) string {
			return want(b.BatchID == "José é 😀 😀 日本", "escapes must decode to the same runes")
		}, HTTP: 202},
	{Name: "escaped_slash_and_controls", Body: `{"batch_id":"a\/b\t\n\"\\\u0001","events":[{"id":1}]}`, Raw: SpecAccept,
		Check: func(b *model.IngestBatch) string { return want(b.BatchID == "a/b\t\n\"\\\x01", "standard escapes") }, HTTP: 202},
	{Name: "lone_surrogate_escape", Body: `{"batch_id":"\ud800","events":[{"id":1}]}`, Raw: SpecAccept,
		Check: func(b *model.IngestBatch) string {
			return want(b.BatchID == "�", "v1 replaces invalid surrogates with U+FFFD")
		}, HTTP: 400, Note: "policy: strict UTF-8"},
	{Name: "invalid_raw_utf8", Body: "{\"batch_id\":\"a\xffb\",\"events\":[{\"id\":1}]}", Raw: SpecAccept,
		Check: func(b *model.IngestBatch) string {
			return want(b.BatchID == "a�b", "v1 replaces invalid UTF-8 with U+FFFD")
		}, HTTP: 400},
	{Name: "deep_nesting_ok", Body: deep(1000), Raw: SpecAccept, HTTP: 202},
	{Name: "deep_nesting_excessive", Body: deep(10001), Raw: SpecReject, HTTP: 400, Note: "v1 limits nesting depth to 10000"},
}
