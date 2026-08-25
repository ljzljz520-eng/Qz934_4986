package domain

import (
	"fmt"
	"strings"
	"time"
)

type Audit struct {
	ID        string    `json:"id"`
	RecordID  string    `json:"record_id"`
	Action    string    `json:"action"`
	Actor     string    `json:"actor"`
	Reason    string    `json:"reason,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

func NewAudit(id, recordID, action, actor, reason string, now time.Time) Audit {
	return Audit{ID: strings.TrimSpace(id), RecordID: strings.TrimSpace(recordID), Action: strings.TrimSpace(action), Actor: strings.TrimSpace(actor), Reason: strings.TrimSpace(reason), CreatedAt: now.UTC()}
}

func (a Audit) Validate() error {
	if a.ID == "" || a.RecordID == "" || a.Action == "" || a.Actor == "" || a.CreatedAt.IsZero() {
		return fmt.Errorf("audit identity, action, actor and timestamp are required")
	}
	return nil
}

func (a Audit) Summary() string {
	if a.Reason == "" {
		return fmt.Sprintf("%s by %s", a.Action, a.Actor)
	}
	return fmt.Sprintf("%s by %s: %s", a.Action, a.Actor, a.Reason)
}
