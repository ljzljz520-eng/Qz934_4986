package service

import (
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"membership13/audit"
	"membership13/domain"
	"membership13/notify"
	"membership13/query"
	"membership13/store"
)

var (
	ErrDuplicateRecord = errors.New("record already exists")
	ErrUserRequired    = errors.New("active user is required")
)

type Clock func() time.Time

type Service struct {
	store *store.Store
	clock Clock
	mu    sync.RWMutex
}

func New(st *store.Store) *Service {
	return &Service{store: st, clock: time.Now}
}

func (s *Service) WithClock(clock Clock) *Service {
	if clock != nil {
		s.clock = clock
	}
	return s
}

func (s *Service) RegisterUser(user domain.User) error {
	if err := user.Validate(); err != nil {
		return err
	}
	if existing, err := s.store.FindUserByMember(user.MemberID); err == nil && existing.ID != user.ID {
		return fmt.Errorf("%w for member %d", ErrDuplicateRecord, user.MemberID)
	}
	return s.store.SaveUser(user)
}

func (s *Service) RegisterRecord(record domain.Record) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	record = domain.NormalizeRecord(record)
	if err := record.Validate(); err != nil {
		return err
	}
	if _, err := s.store.GetRecord(record.ID); err == nil {
		return ErrDuplicateRecord
	}
	user, err := s.store.FindUserByMember(record.MemberID)
	if err != nil || !user.Active || !user.EligibleForBenefit(record.BenefitCode) {
		return ErrUserRequired
	}
	return s.store.SaveRecord(record)
}

func (s *Service) ProcessRecord(id string) (domain.Record, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	record, err := s.store.GetRecord(id)
	if err != nil {
		return domain.Record{}, err
	}
	processed, err := record.Process(s.clock())
	if err != nil {
		return record, err
	}
	_ = s.store.CompareAndSwapRecord(record.ID, processed.Status, processed)
	event := notify.Build(processed, s.clock())
	if err := s.store.SaveEvent(event); err != nil {
		return processed, err
	}
	trail := audit.Review(processed, "benefit-worker", s.clock())
	if err := s.store.SaveAudit(trail); err != nil {
		return processed, err
	}
	return processed, nil
}

func (s *Service) ArchiveRecord(id string) (domain.Record, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	record, err := s.store.GetRecord(id)
	if err != nil {
		return domain.Record{}, err
	}
	if record.Status == domain.StatusPending {
		record, err = record.Process(s.clock())
		if err != nil {
			return record, err
		}
	}
	record, err = record.Archive()
	if err != nil {
		return record, err
	}
	if err = s.store.SaveRecord(record); err != nil {
		return record, err
	}
	return record, s.store.SaveAudit(audit.Archive(record, "archive-worker", s.clock()))
}

func (s *Service) RejectRecord(id, reason string) (domain.Record, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	record, err := s.store.GetRecord(id)
	if err != nil {
		return domain.Record{}, err
	}
	record, err = record.Reject(reason)
	if err != nil {
		return record, err
	}
	return record, s.store.SaveRecord(record)
}

func (s *Service) QueryRecords(filter query.RecordFilter) ([]domain.Record, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	records, err := s.store.ListRecords()
	if err != nil {
		return nil, err
	}
	return query.SortRecords(query.FilterRecords(records, filter), true), nil
}

func (s *Service) Summary(filter query.RecordFilter) (query.Summary, error) {
	records, err := s.QueryRecords(filter)
	if err != nil {
		return query.Summary{}, err
	}
	return query.Summarize(records), nil
}

func (s *Service) Timeline(recordID string) ([]string, error) {
	events, err := s.store.ListEvents(recordID)
	if err != nil {
		return nil, err
	}
	audits, err := s.store.ListAudits(recordID)
	if err != nil {
		return nil, err
	}
	items := make([]string, 0, len(events)+len(audits))
	for _, event := range events {
		items = append(items, "event:"+event.Type)
	}
	for _, item := range audits {
		items = append(items, "audit:"+item.Action)
	}
	sort.Strings(items)
	return items, nil
}

func (s *Service) GetRecord(id string) (domain.Record, error) { return s.store.GetRecord(id) }

func (s *Service) Health() error { return s.store.Health() }
