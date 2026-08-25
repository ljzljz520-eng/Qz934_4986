package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"membership13/domain"
)

type Snapshot struct {
	CreatedAt time.Time       `json:"created_at"`
	Records   []domain.Record `json:"records"`
	Users     []domain.User   `json:"users"`
	Events    []domain.Event  `json:"events"`
	Audits    []domain.Audit  `json:"audits"`
}

func (s *Store) Snapshot() (Snapshot, error) {
	records, err := s.ListRecords()
	if err != nil {
		return Snapshot{}, err
	}
	users, err := s.ListUsers()
	if err != nil {
		return Snapshot{}, err
	}
	events, err := s.ListEvents("")
	if err != nil {
		return Snapshot{}, err
	}
	audits, err := s.ListAudits("")
	if err != nil {
		return Snapshot{}, err
	}
	return Snapshot{CreatedAt: time.Now().UTC(), Records: records, Users: users, Events: events, Audits: audits}, nil
}

func (s *Store) ExportJSON(path string) error {
	if filepath.Clean(path) == "." || path == "" {
		return fmt.Errorf("export path is required")
	}
	snapshot, err := s.Snapshot()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(filepath.Clean(path)), 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Clean(path), data, 0o600)
}

func (s *Store) ImportJSON(path string) error {
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return err
	}
	var snapshot Snapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return err
	}
	for _, user := range snapshot.Users {
		if err := s.SaveUser(user); err != nil {
			return err
		}
	}
	for _, record := range snapshot.Records {
		if err := s.SaveRecord(record); err != nil {
			return err
		}
	}
	for _, event := range snapshot.Events {
		if err := s.SaveEvent(event); err != nil {
			return err
		}
	}
	for _, audit := range snapshot.Audits {
		if err := s.SaveAudit(audit); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) DeleteByStatus(status string) (int, error) {
	records, err := s.ListRecords()
	if err != nil {
		return 0, err
	}
	removed := 0
	for _, record := range records {
		if record.Status == status {
			if err := s.DeleteRecord(record.ID); err != nil {
				return removed, err
			}
			removed++
		}
	}
	return removed, nil
}

func (s *Store) ReplaceRecord(record domain.Record) error {
	if _, err := s.GetRecord(record.ID); err != nil && err != ErrNotFound {
		return err
	}
	return s.SaveRecord(record)
}

func (s *Store) RecordExists(id string) bool {
	_, err := s.GetRecord(id)
	return err == nil
}

func (s *Store) UserExists(id string) bool {
	_, err := s.GetUser(id)
	return err == nil
}

func (s *Store) EventExists(id string) bool {
	_, err := s.GetEvent(id)
	return err == nil
}

func (s *Store) AuditExists(id string) bool {
	_, err := s.GetAudit(id)
	return err == nil
}
