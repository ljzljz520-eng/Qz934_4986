package store

import (
	"encoding/json"
	"sort"

	"go.etcd.io/bbolt"
	"membership13/domain"
)

func (s *Store) SaveUser(user domain.User) error {
	if err := validateEntity(user); err != nil {
		return err
	}
	data, err := json.Marshal(user)
	if err != nil {
		return err
	}
	return s.Update(func(tx *bbolt.Tx) error {
		return tx.Bucket(bucketUsers).Put([]byte(user.ID), data)
	})
}

func (s *Store) GetUser(id string) (domain.User, error) {
	var user domain.User
	err := s.View(func(tx *bbolt.Tx) error {
		value := tx.Bucket(bucketUsers).Get([]byte(id))
		if value == nil {
			return ErrNotFound
		}
		return json.Unmarshal(append([]byte(nil), value...), &user)
	})
	return user, err
}

func (s *Store) ListUsers() ([]domain.User, error) {
	users := make([]domain.User, 0)
	err := s.View(func(tx *bbolt.Tx) error {
		return tx.Bucket(bucketUsers).ForEach(func(_, value []byte) error {
			if value == nil {
				return nil
			}
			var user domain.User
			if err := json.Unmarshal(value, &user); err != nil {
				return err
			}
			users = append(users, user)
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(users, func(i, j int) bool { return users[i].MemberID < users[j].MemberID })
	return users, nil
}

func (s *Store) FindUserByMember(memberID int) (domain.User, error) {
	users, err := s.ListUsers()
	if err != nil {
		return domain.User{}, err
	}
	for _, user := range users {
		if user.MemberID == memberID {
			return user, nil
		}
	}
	return domain.User{}, ErrNotFound
}
