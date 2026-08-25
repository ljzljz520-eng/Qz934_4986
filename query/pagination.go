package query

import (
	"strings"

	"membership13/domain"
)

type Page struct {
	Items  []domain.Record `json:"items"`
	Offset int             `json:"offset"`
	Limit  int             `json:"limit"`
	Total  int             `json:"total"`
}

func Search(records []domain.Record, term string) []domain.Record {
	term = strings.TrimSpace(term)
	if term == "" {
		return append([]domain.Record(nil), records...)
	}
	result := make([]domain.Record, 0)
	for _, record := range records {
		if record.MatchesText(term) {
			result = append(result, record)
		}
	}
	return result
}

func Paginate(records []domain.Record, offset, limit int) Page {
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 {
		limit = 20
	}
	if offset > len(records) {
		offset = len(records)
	}
	end := offset + limit
	if end > len(records) {
		end = len(records)
	}
	return Page{Items: append([]domain.Record(nil), records[offset:end]...), Offset: offset, Limit: limit, Total: len(records)}
}

func StatusCounts(records []domain.Record) map[string]int {
	counts := map[string]int{}
	for _, record := range records {
		counts[record.Status]++
	}
	return counts
}

func GroupByMember(records []domain.Record) map[int][]domain.Record {
	groups := map[int][]domain.Record{}
	for _, record := range records {
		groups[record.MemberID] = append(groups[record.MemberID], record)
	}
	return groups
}
