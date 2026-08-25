package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	"go.etcd.io/bbolt"
	"membership13/domain"
)

var ErrTransitionConflict = errors.New("record status transition conflict")

func (s *Store) SaveRecord(record domain.Record) error {
	record = domain.NormalizeRecord(record)
	if err := validateEntity(record); err != nil {
		return err
	}
	data, err := json.Marshal(record)
	if err != nil {
		return err
	}
	return s.Update(func(tx *bbolt.Tx) error {
		return tx.Bucket(bucketRecords).Put([]byte(record.ID), data)
	})
}

func (s *Store) GetRecord(id string) (domain.Record, error) {
	var record domain.Record
	err := s.View(func(tx *bbolt.Tx) error {
		value := tx.Bucket(bucketRecords).Get([]byte(id))
		if value == nil {
			return ErrNotFound
		}
		return json.Unmarshal(append([]byte(nil), value...), &record)
	})
	return record, err
}

func (s *Store) ListRecords() ([]domain.Record, error) {
	records := make([]domain.Record, 0)
	err := s.View(func(tx *bbolt.Tx) error {
		return tx.Bucket(bucketRecords).ForEach(func(_, value []byte) error {
			if value == nil {
				return nil
			}
			var item domain.Record
			if err := json.Unmarshal(value, &item); err != nil {
				return err
			}
			records = append(records, item)
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(records, func(i, j int) bool {
		if records[i].ReceivedAt.Equal(records[j].ReceivedAt) {
			return records[i].ID < records[j].ID
		}
		return records[i].ReceivedAt.Before(records[j].ReceivedAt)
	})
	return records, nil
}

func (s *Store) UpdateRecordStatus(id string, status string) error {
	record, err := s.GetRecord(id)
	if err != nil {
		return err
	}
	if !domain.ValidStatus(status) {
		return fmt.Errorf("invalid status %s", status)
	}
	record.Status = status
	return s.SaveRecord(record)
}

func (s *Store) DeleteRecord(id string) error {
	return s.Update(func(tx *bbolt.Tx) error {
		if tx.Bucket(bucketRecords).Get([]byte(id)) == nil {
			return ErrNotFound
		}
		return tx.Bucket(bucketRecords).Delete([]byte(id))
	})
}

func (s *Store) CompareAndSwapRecord(id, expectedStatus string, next domain.Record) error {
	if !domain.ValidStatus(expectedStatus) {
		return fmt.Errorf("invalid expected status %s", expectedStatus)
	}
	if err := next.Validate(); err != nil {
		return err
	}
	data, err := json.Marshal(next)
	if err != nil {
		return err
	}
	return s.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(bucketRecords)
		value := bucket.Get([]byte(id))
		if value == nil {
			return ErrNotFound
		}
		var current domain.Record
		if err := json.Unmarshal(value, &current); err != nil {
			return err
		}
		if current.Status != expectedStatus {
			return fmt.Errorf("%w: expected %s, found %s", ErrTransitionConflict, expectedStatus, current.Status)
		}
		return bucket.Put([]byte(id), data)
	})
}
