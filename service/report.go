package service

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"membership13/domain"
	"membership13/query"
)

type MemberReport struct {
	MemberID       int            `json:"member_id"`
	RecordCount    int            `json:"record_count"`
	ProcessedCount int            `json:"processed_count"`
	PendingCount   int            `json:"pending_count"`
	FinalCount     int            `json:"final_count"`
	Benefits       []string       `json:"benefits"`
	Latest         *domain.Record `json:"latest,omitempty"`
	GeneratedAt    time.Time      `json:"generated_at"`
}

func (s *Service) MemberReport(memberID int) (MemberReport, error) {
	records, err := s.QueryRecords(query.RecordFilter{MemberID: memberID})
	if err != nil {
		return MemberReport{}, err
	}
	report := MemberReport{MemberID: memberID, RecordCount: len(records), Benefits: query.DistinctBenefits(records), GeneratedAt: s.clock().UTC()}
	for _, record := range records {
		switch record.Status {
		case domain.StatusPending:
			report.PendingCount++
		case domain.StatusProcessed:
			report.ProcessedCount++
		case domain.StatusArchived, domain.StatusRejected:
			report.FinalCount++
		}
	}
	if latest, ok := query.Latest(records); ok {
		report.Latest = &latest
	}
	return report, nil
}

func (s *Service) MemberDigest(memberID int) (string, error) {
	report, err := s.MemberReport(memberID)
	if err != nil {
		return "", err
	}
	latest := "none"
	if report.Latest != nil {
		latest = report.Latest.ID
	}
	return fmt.Sprintf("member=%d records=%d processed=%d pending=%d latest=%s", report.MemberID, report.RecordCount, report.ProcessedCount, report.PendingCount, latest), nil
}

func (s *Service) StatusBreakdown() (map[string]int, error) {
	records, err := s.QueryRecords(query.RecordFilter{})
	if err != nil {
		return nil, err
	}
	return query.StatusCounts(records), nil
}

func (s *Service) BenefitBreakdown() (map[string]int, error) {
	records, err := s.QueryRecords(query.RecordFilter{})
	if err != nil {
		return nil, err
	}
	counts := map[string]int{}
	for _, record := range records {
		counts[strings.ToUpper(record.BenefitCode)]++
	}
	return counts, nil
}

func (s *Service) MembersWithPending() ([]int, error) {
	records, err := s.QueryRecords(query.RecordFilter{Status: domain.StatusPending})
	if err != nil {
		return nil, err
	}
	seen := map[int]bool{}
	result := make([]int, 0)
	for _, record := range records {
		if !seen[record.MemberID] {
			seen[record.MemberID] = true
			result = append(result, record.MemberID)
		}
	}
	sort.Ints(result)
	return result, nil
}

func (s *Service) ReconcileMember(memberID int) error {
	user, err := s.store.FindUserByMember(memberID)
	if err != nil {
		return err
	}
	records, err := s.QueryRecords(query.RecordFilter{MemberID: memberID})
	if err != nil {
		return err
	}
	for _, record := range records {
		if !user.EligibleForBenefit(record.BenefitCode) && record.Status == domain.StatusPending {
			if _, err := s.RejectRecord(record.ID, "member eligibility changed"); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Service) ResolveStatusLabel(id string) (string, error) {
	record, err := s.GetRecord(id)
	if err != nil {
		return "", err
	}
	return domain.StatusLabel(record.Status), nil
}
