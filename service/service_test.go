package service

import (
	"path/filepath"
	"testing"
	"time"

	"membership13/domain"
	"membership13/query"
	"membership13/store"
)

func testService(t *testing.T) *Service {
	t.Helper()
	s, err := store.Open(filepath.Join(t.TempDir(), "service.db"))
	if err != nil {
		t.Fatal(err)
	}
	svc := New(s).WithClock(func() time.Time { return time.Unix(100, 0).UTC() })
	if err := svc.RegisterUser(domain.NewUser("u13", 13, "Member 13", "gold", time.Unix(1, 0))); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	return svc
}

func TestProcessRecord(t *testing.T) {
	svc := testService(t)
	if err := svc.RegisterRecord(domain.NewRecord("r1", 13, "basic-gift", time.Unix(2, 0))); err != nil {
		t.Fatal(err)
	}
	got, err := svc.ProcessRecord("r1")
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != domain.StatusProcessed {
		t.Fatalf("returned status %s", got.Status)
	}
	items, err := svc.QueryRecords(query.RecordFilter{MemberID: 13})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("expected one record, got %d", len(items))
	}
}

func TestQueryRecords(t *testing.T) {
	svc := testService(t)
	for _, id := range []string{"r1", "r2"} {
		if err := svc.RegisterRecord(domain.NewRecord(id, 13, "basic-gift", time.Unix(2, 0))); err != nil {
			t.Fatal(err)
		}
	}
	items, err := svc.QueryRecords(query.RecordFilter{MemberID: 13})
	if err != nil || len(items) != 2 {
		t.Fatalf("query failed: %v", err)
	}
}
