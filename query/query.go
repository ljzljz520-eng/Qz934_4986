package query

import (
	"sort"
	"strings"
	"time"

	"membership13/domain"
)

type RecordFilter struct {
	MemberID    int
	BenefitCode string
	Status      string
	From        time.Time
	To          time.Time
}

type Summary struct {
	Total    int            `json:"total"`
	ByStatus map[string]int `json:"by_status"`
	Latest   *domain.Record `json:"latest,omitempty"`
}

func FilterRecords(records []domain.Record, filter RecordFilter) []domain.Record {
	result := make([]domain.Record, 0, len(records))
	for _, record := range records {
		if filter.MemberID > 0 && record.MemberID != filter.MemberID {
			continue
		}
		if filter.BenefitCode != "" && !strings.EqualFold(record.BenefitCode, filter.BenefitCode) {
			continue
		}
		if filter.Status != "" && record.Status != filter.Status {
			continue
		}
		if !filter.From.IsZero() && record.ReceivedAt.Before(filter.From) {
			continue
		}
		if !filter.To.IsZero() && record.ReceivedAt.After(filter.To) {
			continue
		}
		result = append(result, record)
	}
	return result
}

func SortRecords(records []domain.Record, descending bool) []domain.Record {
	result := append([]domain.Record(nil), records...)
	sort.SliceStable(result, func(i, j int) bool {
		if descending {
			return result[i].ReceivedAt.After(result[j].ReceivedAt)
		}
		return result[i].ReceivedAt.Before(result[j].ReceivedAt)
	})
	return result
}

func Summarize(records []domain.Record) Summary {
	summary := Summary{Total: len(records), ByStatus: map[string]int{}}
	for i := range records {
		summary.ByStatus[records[i].Status]++
		if summary.Latest == nil || records[i].ReceivedAt.After(summary.Latest.ReceivedAt) {
			item := records[i]
			summary.Latest = &item
		}
	}
	return summary
}

func PendingOnly(records []domain.Record) []domain.Record {
	return FilterRecords(records, RecordFilter{Status: domain.StatusPending})
}

func ProcessedOnly(records []domain.Record) []domain.Record {
	return FilterRecords(records, RecordFilter{Status: domain.StatusProcessed})
}

func MemberRecords(records []domain.Record, memberID int) []domain.Record {
	return SortRecords(FilterRecords(records, RecordFilter{MemberID: memberID}), false)
}

func HasStatus(records []domain.Record, status string) bool {
	for _, record := range records {
		if record.Status == status {
			return true
		}
	}
	return false
}
