package notify

import (
	"encoding/json"
	"fmt"
	"time"

	"membership13/domain"
)

type Envelope struct {
	Topic string       `json:"topic"`
	Event domain.Event `json:"event"`
}

func Build(record domain.Record, now time.Time) domain.Event {
	payload, _ := json.Marshal(map[string]interface{}{"member_id": record.MemberID, "benefit_code": record.BenefitCode, "status": domain.StatusLabel(record.Status)})
	return domain.NewEvent(fmt.Sprintf("event-%s", record.ID), record.ID, "benefit.processed", string(payload), now)
}

func BuildEnvelope(event domain.Event) Envelope {
	return Envelope{Topic: "membership.benefits", Event: event}
}

func ParsePayload(event domain.Event) (map[string]interface{}, error) {
	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(event.Payload), &payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func IsProcessEvent(event domain.Event) bool { return event.Type == "benefit.processed" }

func Retry(event domain.Event) domain.Event { return event.MarkAttempt() }

func Delivered(event domain.Event) domain.Event { return event.Deliver() }
