package membership13

import (
	"path/filepath"
	"testing"
	"time"

	"membership13/domain"
	"membership13/store"
)

func TestPersistenceSurvivesReopen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "reopen.db")
	s, err := store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	r := domain.NewRecord("persisted", 13, "basic-gift", time.Unix(1, 0))
	if err := s.SaveRecord(r); err != nil {
		t.Fatal(err)
	}
	if err := s.SaveUser(domain.NewUser("u", 13, "Member", "gold", time.Unix(1, 0))); err != nil {
		t.Fatal(err)
	}
	if err := s.SaveEvent(domain.NewEvent("e", "persisted", "created", "{}", time.Unix(1, 0))); err != nil {
		t.Fatal(err)
	}
	if err := s.SaveAudit(domain.NewAudit("a", "persisted", "created", "test", "", time.Unix(1, 0))); err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	s, err = store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if _, err := s.GetRecord("persisted"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.GetUser("u"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.GetEvent("e"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.GetAudit("a"); err != nil {
		t.Fatal(err)
	}
}
