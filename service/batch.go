package service

import (
	"fmt"
	"sort"
	"time"

	"membership13/domain"
	"membership13/query"
)

type BatchResult struct {
	Requested int      `json:"requested"`
	Processed int      `json:"processed"`
	Skipped   int      `json:"skipped"`
	Failed    int      `json:"failed"`
	IDs       []string `json:"ids"`
	Errors    []string `json:"errors,omitempty"`
}

func (s *Service) ProcessPending(memberID int) (BatchResult, error) {
	records, err := s.QueryRecords(query.RecordFilter{MemberID: memberID, Status: domain.StatusPending})
	if err != nil {
		return BatchResult{}, err
	}
	result := BatchResult{Requested: len(records), IDs: make([]string, 0, len(records))}
	for _, record := range records {
		processed, processErr := s.ProcessRecord(record.ID)
		if processErr != nil {
			result.Failed++
			result.Errors = append(result.Errors, record.ID+": "+processErr.Error())
			continue
		}
		if processed.Status == domain.StatusProcessed {
			result.Processed++
			result.IDs = append(result.IDs, processed.ID)
		} else {
			result.Skipped++
		}
	}
	return result, nil
}

func (s *Service) ProcessIDs(ids []string) BatchResult {
	result := BatchResult{Requested: len(ids), IDs: make([]string, 0, len(ids))}
	for _, id := range ids {
		if id == "" {
			result.Skipped++
			continue
		}
		record, err := s.ProcessRecord(id)
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, id+": "+err.Error())
			continue
		}
		result.Processed++
		result.IDs = append(result.IDs, record.ID)
	}
	return result
}

func (s *Service) ArchiveProcessed(memberID int) (int, error) {
	records, err := s.QueryRecords(query.RecordFilter{MemberID: memberID, Status: domain.StatusProcessed})
	if err != nil {
		return 0, err
	}
	count := 0
	for _, record := range records {
		if _, err := s.ArchiveRecord(record.ID); err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}

func (s *Service) RegisterRecords(records []domain.Record) BatchResult {
	result := BatchResult{Requested: len(records), IDs: make([]string, 0, len(records))}
	for _, record := range records {
		if err := s.RegisterRecord(record); err != nil {
			result.Failed++
			result.Errors = append(result.Errors, record.ID+": "+err.Error())
			continue
		}
		result.Processed++
		result.IDs = append(result.IDs, record.ID)
	}
	return result
}

func (s *Service) StaleRecords(memberID int, now time.Time, age time.Duration) ([]domain.Record, error) {
	records, err := s.QueryRecords(query.RecordFilter{MemberID: memberID, Status: domain.StatusPending})
	if err != nil {
		return nil, err
	}
	result := make([]domain.Record, 0)
	for _, record := range records {
		if record.Age(now) >= age {
			result = append(result, record)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Age(now) > result[j].Age(now) })
	return result, nil
}

func (s *Service) ExplainRecord(id string) (string, error) {
	record, err := s.GetRecord(id)
	if err != nil {
		return "", err
	}
	switch record.Status {
	case domain.StatusPending:
		return fmt.Sprintf("%s is awaiting processing", id), nil
	case domain.StatusProcessed:
		return fmt.Sprintf("%s processed for member %d", id, record.MemberID), nil
	case domain.StatusArchived:
		return fmt.Sprintf("%s archived", id), nil
	default:
		return fmt.Sprintf("%s rejected: %s", id, record.Notes), nil
	}
}
