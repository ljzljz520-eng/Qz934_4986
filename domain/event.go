package domain

import (
	"fmt"
	"strings"
	"time"
)

type Event struct {
	ID        string    `json:"id"`
	RecordID  string    `json:"record_id"`
	Type      string    `json:"type"`
	Payload   string    `json:"payload"`
	CreatedAt time.Time `json:"created_at"`
	Attempts  int       `json:"attempts"`
	Delivered bool      `json:"delivered"`
}

func NewEvent(id, recordID, eventType, payload string, now time.Time) Event {
	return Event{ID: strings.TrimSpace(id), RecordID: strings.TrimSpace(recordID), Type: strings.TrimSpace(eventType), Payload: payload, CreatedAt: now.UTC()}
}

func (e Event) Validate() error {
	if e.ID == "" || e.RecordID == "" || e.Type == "" || e.CreatedAt.IsZero() {
		return fmt.Errorf("event identity and timestamp are required")
	}
	if e.Attempts < 0 {
		return fmt.Errorf("event attempts cannot be negative")
	}
	return nil
}

func (e Event) MarkAttempt() Event {
	e.Attempts++
	return e
}

func (e Event) Deliver() Event {
	e.Delivered = true
	return e
}

func (e Event) Pending() bool {
	return !e.Delivered
}
