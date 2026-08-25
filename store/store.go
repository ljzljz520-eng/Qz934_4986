package store

import (
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"go.etcd.io/bbolt"
	"membership13/domain"
)

var (
	ErrNotFound = errors.New("entity not found")
	ErrClosed   = errors.New("store is closed")
)

var bucketRecords = []byte("records")
var bucketUsers = []byte("users")
var bucketEvents = []byte("events")
var bucketAudits = []byte("audits")
var bucketMeta = []byte("meta")

type Store struct {
	db   *bbolt.DB
	path string
}

func Open(path string) (*Store, error) {
	if path == "" {
		return nil, fmt.Errorf("database path is required")
	}
	db, err := bbolt.Open(filepath.Clean(path), 0o600, &bbolt.Options{Timeout: 2 * time.Second, NoSync: true})
	if err != nil {
		return nil, err
	}
	s := &Store{db: db, path: filepath.Clean(path)}
	if err := s.initialize(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) initialize() error {
	if s == nil || s.db == nil {
		return ErrClosed
	}
	return s.db.Update(func(tx *bbolt.Tx) error {
		for _, name := range [][]byte{bucketRecords, bucketUsers, bucketEvents, bucketAudits, bucketMeta} {
			if _, err := tx.CreateBucketIfNotExists(name); err != nil {
				return err
			}
		}
		return tx.Bucket(bucketMeta).Put([]byte("schema"), []byte("1"))
	})
}

func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	err := s.db.Close()
	s.db = nil
	return err
}

func (s *Store) Path() string { return s.path }

func (s *Store) View(fn func(*bbolt.Tx) error) error {
	if s == nil || s.db == nil {
		return ErrClosed
	}
	return s.db.View(fn)
}

func (s *Store) Update(fn func(*bbolt.Tx) error) error {
	if s == nil || s.db == nil {
		return ErrClosed
	}
	return s.db.Update(fn)
}

func (s *Store) Health() error {
	if s == nil || s.db == nil {
		return ErrClosed
	}
	return s.db.View(func(tx *bbolt.Tx) error {
		if tx.Bucket(bucketRecords) == nil || tx.Bucket(bucketUsers) == nil || tx.Bucket(bucketEvents) == nil || tx.Bucket(bucketAudits) == nil {
			return fmt.Errorf("storage schema incomplete")
		}
		return nil
	})
}

func (s *Store) CountAll() (map[string]int, error) {
	counts := map[string]int{}
	err := s.View(func(tx *bbolt.Tx) error {
		for key, name := range map[string][]byte{"records": bucketRecords, "users": bucketUsers, "events": bucketEvents, "audits": bucketAudits} {
			b := tx.Bucket(name)
			if b == nil {
				return fmt.Errorf("missing bucket %s", key)
			}
			counts[key] = b.Stats().KeyN
		}
		return nil
	})
	return counts, err
}

func validateEntity(v interface{ Validate() error }) error {
	if v == nil {
		return fmt.Errorf("entity cannot be nil")
	}
	return v.Validate()
}

var _ = domain.StatusPending
