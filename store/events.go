package store

import (
	"encoding/json"
	"sort"

	"go.etcd.io/bbolt"
	"membership13/domain"
)

func (s *Store) SaveEvent(event domain.Event) error {
	if err := event.Validate(); err != nil {
		return err
	}
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return s.Update(func(tx *bbolt.Tx) error { return tx.Bucket(bucketEvents).Put([]byte(event.ID), data) })
}

func (s *Store) GetEvent(id string) (domain.Event, error) {
	var event domain.Event
	err := s.View(func(tx *bbolt.Tx) error {
		value := tx.Bucket(bucketEvents).Get([]byte(id))
		if value == nil {
			return ErrNotFound
		}
		return json.Unmarshal(append([]byte(nil), value...), &event)
	})
	return event, err
}

func (s *Store) ListEvents(recordID string) ([]domain.Event, error) {
	events := make([]domain.Event, 0)
	err := s.View(func(tx *bbolt.Tx) error {
		return tx.Bucket(bucketEvents).ForEach(func(_, value []byte) error {
			if value == nil {
				return nil
			}
			var event domain.Event
			if err := json.Unmarshal(value, &event); err != nil {
				return err
			}
			if recordID == "" || event.RecordID == recordID {
				events = append(events, event)
			}
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(events, func(i, j int) bool { return events[i].CreatedAt.Before(events[j].CreatedAt) })
	return events, nil
}
