package store

import (
	"encoding/json"
	"sort"

	"go.etcd.io/bbolt"
	"membership13/domain"
)

func (s *Store) SaveAudit(audit domain.Audit) error {
	if err := audit.Validate(); err != nil {
		return err
	}
	data, err := json.Marshal(audit)
	if err != nil {
		return err
	}
	return s.Update(func(tx *bbolt.Tx) error { return tx.Bucket(bucketAudits).Put([]byte(audit.ID), data) })
}

func (s *Store) GetAudit(id string) (domain.Audit, error) {
	var audit domain.Audit
	err := s.View(func(tx *bbolt.Tx) error {
		value := tx.Bucket(bucketAudits).Get([]byte(id))
		if value == nil {
			return ErrNotFound
		}
		return json.Unmarshal(append([]byte(nil), value...), &audit)
	})
	return audit, err
}

func (s *Store) ListAudits(recordID string) ([]domain.Audit, error) {
	audits := make([]domain.Audit, 0)
	err := s.View(func(tx *bbolt.Tx) error {
		return tx.Bucket(bucketAudits).ForEach(func(_, value []byte) error {
			if value == nil {
				return nil
			}
			var audit domain.Audit
			if err := json.Unmarshal(value, &audit); err != nil {
				return err
			}
			if recordID == "" || audit.RecordID == recordID {
				audits = append(audits, audit)
			}
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(audits, func(i, j int) bool { return audits[i].CreatedAt.Before(audits[j].CreatedAt) })
	return audits, nil
}
