// Package model is the shared, strongly typed destination for /ingest payloads.
//
// Both server binaries decode into exactly these types. There are no custom
// UnmarshalJSON methods: each decoder uses its own reflection/compiled path.
// Every JSON key produced by tools/cmd/fixturegen maps to a field here
// (proven by TestFixtureCoverage in service/bench).
package model

import "errors"

// IngestBatch is the top-level document.
type IngestBatch struct {
	BatchID   string  `json:"batch_id"`
	Source    Source  `json:"source"`
	CreatedAt string  `json:"created_at"`
	Events    []Event `json:"events"`
}

// Source describes the emitting system.
type Source struct {
	Host   string   `json:"host"`
	Region string   `json:"region"`
	Tags   []string `json:"tags"`
}

// Event is one business event in the batch.
type Event struct {
	ID       int64   `json:"id"`
	Type     string  `json:"type"`
	TS       string  `json:"ts"`
	User     User    `json:"user"`
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
	Items    []Item  `json:"items"`
	Geo      *Geo    `json:"geo"`
	Note     string  `json:"note"`
	Flags    []bool  `json:"flags"`
}

// User is the actor of an event.
type User struct {
	ID       int64   `json:"id"`
	Name     string  `json:"name"`
	Email    string  `json:"email"`
	Verified bool    `json:"verified"`
	Locale   *string `json:"locale"`
}

// Item is a line item.
type Item struct {
	SKU       string            `json:"sku"`
	Qty       int32             `json:"qty"`
	UnitPrice float64           `json:"unit_price"`
	Attrs     map[string]string `json:"attrs"`
}

// Geo is an optional coordinate.
type Geo struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

// Validation errors. They map to HTTP 422.
var (
	ErrNoBatchID  = errors.New("batch_id is empty")
	ErrNoEvents   = errors.New("events is empty")
	ErrEventID    = errors.New("event id must be > 0")
	ErrItemQty    = errors.New("item qty must be > 0")
	ErrGeoRange   = errors.New("geo out of range")
	ErrNilPayload = errors.New("nil payload")
)

// Validate performs the minimal, library-independent post-decode checks.
// It is intentionally cheap (one pass, no allocation) and identical in both arms.
func (b *IngestBatch) Validate() error {
	if b == nil {
		return ErrNilPayload
	}
	if b.BatchID == "" {
		return ErrNoBatchID
	}
	if len(b.Events) == 0 {
		return ErrNoEvents
	}
	for i := range b.Events {
		e := &b.Events[i]
		if e.ID <= 0 {
			return ErrEventID
		}
		for j := range e.Items {
			if e.Items[j].Qty <= 0 {
				return ErrItemQty
			}
		}
		if g := e.Geo; g != nil && (g.Lat < -90 || g.Lat > 90 || g.Lon < -180 || g.Lon > 180) {
			return ErrGeoRange
		}
	}
	return nil
}
