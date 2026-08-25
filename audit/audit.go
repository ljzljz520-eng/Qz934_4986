package audit

import (
	"fmt"
	"time"

	"membership13/domain"
)

func NewAudit(recordID, action, actor, reason string, now time.Time) domain.Audit {
	return domain.NewAudit(fmt.Sprintf("audit-%s-%d", recordID, now.UnixNano()), recordID, action, actor, reason, now)
}

func Review(record domain.Record, actor string, now time.Time) domain.Audit {
	action := "reviewed"
	reason := "record accepted"
	if record.Status == domain.StatusRejected {
		action = "rejected"
		reason = record.Notes
	}
	return NewAudit(record.ID, action, actor, reason, now)
}

func Archive(record domain.Record, actor string, now time.Time) domain.Audit {
	return NewAudit(record.ID, "archived", actor, "processing complete", now)
}

func IsFinal(a domain.Audit) bool {
	return a.Action == "archived" || a.Action == "rejected"
}

func Actions(audits []domain.Audit) []string {
	result := make([]string, 0, len(audits))
	for _, item := range audits {
		result = append(result, item.Action)
	}
	return result
}
