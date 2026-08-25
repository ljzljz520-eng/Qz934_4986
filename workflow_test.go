package membership13

import (
	"path/filepath"
	"testing"
	"time"

	"membership13/domain"
	"membership13/query"
	"membership13/service"
	"membership13/store"
)

func setupWorkflow(t *testing.T) (*service.Service, *store.Store) {
	t.Helper()
	s, err := store.Open(filepath.Join(t.TempDir(), "workflow.db"))
	if err != nil {
		t.Fatal(err)
	}
	svc := service.New(s).WithClock(func() time.Time { return time.Unix(100, 0) })
	if err := svc.RegisterUser(domain.NewUser("u13", 13, "Member 13", "gold", time.Unix(1, 0))); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	return svc, s
}

func TestWorkflowOne(t *testing.T) {
	svc, _ := setupWorkflow(t)
	if err := svc.RegisterRecord(domain.NewRecord("r13", 13, "basic-gift", time.Unix(2, 0))); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.GetRecord("r13"); err != nil {
		t.Fatal(err)
	}
}

func TestWorkflowTwo(t *testing.T) {
	svc, _ := setupWorkflow(t)
	if err := svc.RegisterRecord(domain.NewRecord("r14", 13, "basic-gift", time.Unix(2, 0))); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.ProcessRecord("r14"); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.ArchiveRecord("r14"); err != nil {
		t.Fatal(err)
	}
}

func TestWorkflowThree(t *testing.T) {
	svc, _ := setupWorkflow(t)
	if err := svc.RegisterRecord(domain.NewRecord("r15", 13, "basic-gift", time.Unix(2, 0))); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.ProcessRecord("r15"); err != nil {
		t.Fatal(err)
	}
	items, err := svc.Timeline("r15")
	if err != nil || len(items) < 2 {
		t.Fatalf("timeline failed: %v", err)
	}
}

func TestRecordFlow13(t *testing.T) {
	svc, _ := setupWorkflow(t)
	if err := svc.RegisterRecord(domain.NewRecord("member-13-benefit", 13, "basic-gift", time.Unix(2, 0))); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.ProcessRecord("member-13-benefit"); err != nil {
		t.Fatal(err)
	}
	records, err := svc.QueryRecords(query.RecordFilter{MemberID: 13})
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 1 || records[0].Status != domain.StatusProcessed {
		t.Fatalf("member 13 benefit should be processed, got %+v", records)
	}
}
