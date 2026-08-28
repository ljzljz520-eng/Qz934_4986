package domain

import (
	"fmt"
	"strings"
	"time"
)

type RecordView struct {
	ID          string `json:"id"`
	MemberID    int    `json:"member_id"`
	BenefitCode string `json:"benefit_code"`
	Status      string `json:"status"`
	StatusText  string `json:"status_text"`
	Received    string `json:"received"`
	Processed   string `json:"processed,omitempty"`
}

func (r Record) View(location *time.Location) RecordView {
	if location == nil {
		location = time.UTC
	}
	view := RecordView{ID: r.ID, MemberID: r.MemberID, BenefitCode: r.BenefitCode, Status: r.Status, StatusText: StatusLabel(r.Status), Received: r.ReceivedAt.In(location).Format(time.RFC3339)}
	if !r.ProcessedAt.IsZero() {
		view.Processed = r.ProcessedAt.In(location).Format(time.RFC3339)
	}
	return view
}

func (r Record) SearchText() string {
	return strings.ToLower(fmt.Sprintf("%s %d %s %s %s", r.ID, r.MemberID, r.BenefitCode, r.Status, r.Notes))
}

func (r Record) MatchesText(term string) bool {
	return strings.Contains(r.SearchText(), strings.ToLower(strings.TrimSpace(term)))
}

func (r Record) IsFinal() bool {
	return r.Status == StatusArchived || r.Status == StatusRejected
}

func (r Record) Age(now time.Time) time.Duration {
	if now.Before(r.ReceivedAt) {
		return 0
	}
	return now.Sub(r.ReceivedAt)
}
