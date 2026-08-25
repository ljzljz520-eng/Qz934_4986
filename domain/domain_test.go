package domain

import (
	"testing"
	"time"
)

func TestDomainValidation(t *testing.T) {
	now := time.Unix(100, 0).UTC()
	r := NewRecord("r1", 13, "basic-gift", now)
	if err := r.Validate(); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Process(now.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	if ValidStatus("bad") {
		t.Fatal("invalid status accepted")
	}
	u := NewUser("u1", 13, "Member 13", "gold", now)
	if err := u.Validate(); err != nil {
		t.Fatal(err)
	}
	if !u.EligibleForBenefit("premium") {
		t.Fatal("gold member should be eligible")
	}
}
