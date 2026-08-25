package audit

import (
	"testing"
	"time"

	"membership13/domain"
)

func TestAuditTrail(t *testing.T) {
	r := domain.NewRecord("r1", 13, "basic", time.Unix(1, 0))
	processed, _ := r.Process(time.Unix(2, 0))
	a := Review(processed, "worker", time.Unix(3, 0))
	if err := a.Validate(); err != nil {
		t.Fatal(err)
	}
	if a.Action != "reviewed" {
		t.Fatal("unexpected action")
	}
}
