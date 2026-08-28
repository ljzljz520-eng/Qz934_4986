package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	StatusPending   = "pending"
	StatusProcessed = "processed"
	StatusArchived  = "archived"
	StatusRejected  = "rejected"
)

var (
	ErrInvalidRecord  = errors.New("invalid record")
	ErrInvalidStatus  = errors.New("invalid status")
	ErrNotProcessable = errors.New("record is not processable")
)

type Record struct {
	ID          string    `json:"id"`
	MemberID    int       `json:"member_id"`
	BenefitCode string    `json:"benefit_code"`
	Status      string    `json:"status"`
	ReceivedAt  time.Time `json:"received_at"`
	ProcessedAt time.Time `json:"processed_at,omitempty"`
	Notes       string    `json:"notes,omitempty"`
	Source      string    `json:"source,omitempty"`
}

func NewRecord(id string, memberID int, benefitCode string, now time.Time) Record {
	return Record{ID: strings.TrimSpace(id), MemberID: memberID, BenefitCode: strings.TrimSpace(benefitCode), Status: StatusPending, ReceivedAt: now.UTC(), Source: "membership-page"}
}

func (r Record) Validate() error {
	if strings.TrimSpace(r.ID) == "" || r.MemberID <= 0 || strings.TrimSpace(r.BenefitCode) == "" {
		return fmt.Errorf("%w: identity fields are required", ErrInvalidRecord)
	}
	if r.ReceivedAt.IsZero() {
		return fmt.Errorf("%w: received time is required", ErrInvalidRecord)
	}
	if !ValidStatus(r.Status) {
		return fmt.Errorf("%w: %s", ErrInvalidStatus, r.Status)
	}
	if r.Status == StatusProcessed && r.ProcessedAt.IsZero() {
		return fmt.Errorf("%w: processed time is required", ErrInvalidRecord)
	}
	return nil
}

func (r Record) CanProcess() bool {
	return r.Status == StatusPending && r.MemberID > 0 && strings.TrimSpace(r.BenefitCode) != ""
}

func (r Record) Process(now time.Time) (Record, error) {
	if !r.CanProcess() {
		return r, ErrNotProcessable
	}
	r.Status = StatusProcessed
	r.ProcessedAt = now.UTC()
	return r, r.Validate()
}

func (r Record) Archive() (Record, error) {
	if r.Status != StatusProcessed {
		return r, fmt.Errorf("%w: archive requires processed status", ErrNotProcessable)
	}
	r.Status = StatusArchived
	return r, nil
}

func (r Record) Reject(reason string) (Record, error) {
	if strings.TrimSpace(reason) == "" || r.Status != StatusPending {
		return r, ErrNotProcessable
	}
	r.Status = StatusRejected
	r.Notes = strings.TrimSpace(reason)
	return r, nil
}

func ValidStatus(status string) bool {
	switch status {
	case StatusPending, StatusProcessed, StatusArchived, StatusRejected:
		return true
	default:
		return false
	}
}

func StatusLabel(status string) string {
	switch status {
	case StatusPending:
		return "待处理"
	case StatusProcessed:
		return "已处理"
	case StatusArchived:
		return "已归档"
	case StatusRejected:
		return "已驳回"
	default:
		return "未知"
	}
}

func NormalizeRecord(r Record) Record {
	r.ID = strings.TrimSpace(r.ID)
	r.BenefitCode = strings.ToUpper(strings.TrimSpace(r.BenefitCode))
	r.Notes = strings.TrimSpace(r.Notes)
	r.Source = strings.TrimSpace(r.Source)
	if r.Source == "" {
		r.Source = "membership-page"
	}
	return r
}
