package query

import (
	"sort"

	"membership13/domain"
)

type MemberStats struct {
	MemberID  int
	Total     int
	Pending   int
	Processed int
	Archived  int
	Rejected  int
}

func AggregateMembers(records []domain.Record) []MemberStats {
	groups := map[int]*MemberStats{}
	for _, record := range records {
		stats := groups[record.MemberID]
		if stats == nil {
			stats = &MemberStats{MemberID: record.MemberID}
			groups[record.MemberID] = stats
		}
		stats.Total++
		switch record.Status {
		case domain.StatusPending:
			stats.Pending++
		case domain.StatusProcessed:
			stats.Processed++
		case domain.StatusArchived:
			stats.Archived++
		case domain.StatusRejected:
			stats.Rejected++
		}
	}
	result := make([]MemberStats, 0, len(groups))
	for _, stats := range groups {
		result = append(result, *stats)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].MemberID < result[j].MemberID })
	return result
}

func CompletionRate(stats MemberStats) float64 {
	if stats.Total == 0 {
		return 0
	}
	return float64(stats.Processed+stats.Archived) / float64(stats.Total)
}

func PendingRate(stats MemberStats) float64 {
	if stats.Total == 0 {
		return 0
	}
	return float64(stats.Pending) / float64(stats.Total)
}

func TopMembers(records []domain.Record, limit int) []MemberStats {
	result := AggregateMembers(records)
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].Total == result[j].Total {
			return result[i].MemberID < result[j].MemberID
		}
		return result[i].Total > result[j].Total
	})
	if limit <= 0 || limit >= len(result) {
		return result
	}
	return result[:limit]
}

func ByStatus(records []domain.Record, status string) []domain.Record {
	return FilterRecords(records, RecordFilter{Status: status})
}

func CountByBenefit(records []domain.Record) map[string]int {
	counts := map[string]int{}
	for _, record := range records {
		counts[record.BenefitCode]++
	}
	return counts
}

func CountByDay(records []domain.Record) map[string]int {
	counts := map[string]int{}
	for _, record := range records {
		counts[record.ReceivedAt.UTC().Format("2006-01-02")]++
	}
	return counts
}

func FirstByStatus(records []domain.Record, status string) (domain.Record, bool) {
	items := ByStatus(records, status)
	return Earliest(items)
}

func LastByStatus(records []domain.Record, status string) (domain.Record, bool) {
	items := ByStatus(records, status)
	return Latest(items)
}

func CopyRecords(records []domain.Record) []domain.Record {
	result := make([]domain.Record, len(records))
	copy(result, records)
	return result
}
