package query

import (
	"strings"
	"time"

	"membership13/domain"
)

type Criteria struct {
	MemberID       int
	Terms          []string
	Statuses       []string
	BenefitCodes   []string
	IncludeFinal   bool
	ReceivedAfter  time.Time
	ReceivedBefore time.Time
}

func ApplyCriteria(records []domain.Record, criteria Criteria) []domain.Record {
	result := make([]domain.Record, 0, len(records))
	for _, record := range records {
		if criteria.MemberID > 0 && record.MemberID != criteria.MemberID {
			continue
		}
		if !criteria.IncludeFinal && record.IsFinal() {
			continue
		}
		if !matchStatus(record.Status, criteria.Statuses) {
			continue
		}
		if !matchCode(record.BenefitCode, criteria.BenefitCodes) {
			continue
		}
		if !matchTerms(record, criteria.Terms) {
			continue
		}
		if !criteria.ReceivedAfter.IsZero() && record.ReceivedAt.Before(criteria.ReceivedAfter) {
			continue
		}
		if !criteria.ReceivedBefore.IsZero() && record.ReceivedAt.After(criteria.ReceivedBefore) {
			continue
		}
		result = append(result, record)
	}
	return result
}

func matchStatus(status string, statuses []string) bool {
	if len(statuses) == 0 {
		return true
	}
	for _, candidate := range statuses {
		if strings.EqualFold(status, strings.TrimSpace(candidate)) {
			return true
		}
	}
	return false
}

func matchCode(code string, codes []string) bool {
	if len(codes) == 0 {
		return true
	}
	for _, candidate := range codes {
		if strings.EqualFold(code, strings.TrimSpace(candidate)) {
			return true
		}
	}
	return false
}

func matchTerms(record domain.Record, terms []string) bool {
	for _, term := range terms {
		if strings.TrimSpace(term) != "" && !record.MatchesText(term) {
			return false
		}
	}
	return true
}

func NormalizeCriteria(criteria Criteria) Criteria {
	criteria.Terms = cleanStrings(criteria.Terms)
	criteria.Statuses = cleanStrings(criteria.Statuses)
	criteria.BenefitCodes = cleanStrings(criteria.BenefitCodes)
	return criteria
}

func cleanStrings(values []string) []string {
	result := make([]string, 0, len(values))
	seen := map[string]bool{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		key := strings.ToLower(value)
		if value != "" && !seen[key] {
			seen[key] = true
			result = append(result, value)
		}
	}
	return result
}

func Earliest(records []domain.Record) (domain.Record, bool) {
	if len(records) == 0 {
		return domain.Record{}, false
	}
	item := records[0]
	for _, record := range records[1:] {
		if record.ReceivedAt.Before(item.ReceivedAt) {
			item = record
		}
	}
	return item, true
}

func Latest(records []domain.Record) (domain.Record, bool) {
	if len(records) == 0 {
		return domain.Record{}, false
	}
	item := records[0]
	for _, record := range records[1:] {
		if record.ReceivedAt.After(item.ReceivedAt) {
			item = record
		}
	}
	return item, true
}

func DistinctBenefits(records []domain.Record) []string {
	seen := map[string]bool{}
	result := make([]string, 0)
	for _, record := range records {
		code := strings.ToUpper(record.BenefitCode)
		if !seen[code] {
			seen[code] = true
			result = append(result, code)
		}
	}
	return result
}

func StatusOrder(status string) int {
	switch status {
	case domain.StatusPending:
		return 1
	case domain.StatusProcessed:
		return 2
	case domain.StatusArchived:
		return 3
	case domain.StatusRejected:
		return 4
	default:
		return 0
	}
}
