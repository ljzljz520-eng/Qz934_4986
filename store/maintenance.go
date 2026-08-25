package store

import (
	"fmt"
	"time"

	"go.etcd.io/bbolt"
)

func (s *Store) VacuumMarker(now time.Time) error {
	if now.IsZero() {
		return fmt.Errorf("marker time is required")
	}
	return s.Update(func(tx *bbolt.Tx) error {
		return tx.Bucket(bucketMeta).Put([]byte("last_vacuum_marker"), []byte(now.UTC().Format(time.RFC3339Nano)))
	})
}

func (s *Store) DatabaseReady() bool { return s.Health() == nil }

func (s *Store) TouchMetadata(key, value string) error {
	if key == "" {
		return fmt.Errorf("metadata key is required")
	}
	return s.Update(func(tx *bbolt.Tx) error { return tx.Bucket(bucketMeta).Put([]byte(key), []byte(value)) })
}

func (s *Store) ReadMetadata(key string) (string, error) {
	var value string
	err := s.View(func(tx *bbolt.Tx) error {
		data := tx.Bucket(bucketMeta).Get([]byte(key))
		if data == nil {
			return ErrNotFound
		}
		value = string(data)
		return nil
	})
	return value, err
}
