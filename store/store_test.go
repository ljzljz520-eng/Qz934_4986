package store

import (
	"path/filepath"
	"testing"
	"time"

	"membership13/domain"
)

func TestStoreRecords(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "records.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	r := domain.NewRecord("r1", 13, "BASIC-GIFT", time.Unix(1, 0))
	if err := s.SaveRecord(r); err != nil {
		t.Fatal(err)
	}
	got, err := s.GetRecord("r1")
	if err != nil {
		t.Fatal(err)
	}
	if got.BenefitCode != "BASIC-GIFT" {
		t.Fatalf("unexpected code %s", got.BenefitCode)
	}
}
