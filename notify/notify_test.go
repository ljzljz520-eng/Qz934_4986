package notify

import (
	"testing"
	"time"

	"membership13/domain"
)

func TestNotificationEvent(t *testing.T) {
	r := domain.NewRecord("r1", 13, "basic", time.Unix(1, 0))
	processed, _ := r.Process(time.Unix(2, 0))
	e := Build(processed, time.Unix(3, 0))
	if !IsProcessEvent(e) {
		t.Fatal("wrong event type")
	}
	payload, err := ParsePayload(e)
	if err != nil || payload["member_id"] == nil {
		t.Fatal("payload missing")
	}
}
