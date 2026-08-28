package domain

import (
	"fmt"
	"strings"
	"time"
)

type EligibilityDecision struct {
	Allowed   bool
	Code      string
	Reason    string
	CheckedAt time.Time
}

func CheckEligibility(user User, record Record, now time.Time) EligibilityDecision {
	decision := EligibilityDecision{Code: record.BenefitCode, CheckedAt: now.UTC()}
	if !user.Active {
		decision.Reason = "user inactive"
		return decision
	}
	if user.MemberID != record.MemberID {
		decision.Reason = "member mismatch"
		return decision
	}
	if strings.TrimSpace(record.BenefitCode) == "" {
		decision.Reason = "benefit code missing"
		return decision
	}
	if record.ReceivedAt.After(now) {
		decision.Reason = "received time is in the future"
		return decision
	}
	decision.Allowed = user.EligibleForBenefit(record.BenefitCode)
	if !decision.Allowed {
		decision.Reason = fmt.Sprintf("tier %s is not eligible", user.Tier)
		return decision
	}
	decision.Reason = "eligible"
	return decision
}

func (d EligibilityDecision) Explain() string {
	if d.Allowed {
		return "eligible: " + d.Code
	}
	return "ineligible: " + d.Reason
}

func (d EligibilityDecision) IsFresh(now time.Time, maxAge time.Duration) bool {
	if now.Before(d.CheckedAt) {
		return false
	}
	return now.Sub(d.CheckedAt) <= maxAge
}

func (d EligibilityDecision) Require() error {
	if !d.Allowed {
		return fmt.Errorf("eligibility denied: %s", d.Reason)
	}
	return nil
}

func TierRank(tier string) int {
	switch strings.ToLower(strings.TrimSpace(tier)) {
	case "standard":
		return 1
	case "silver":
		return 2
	case "gold":
		return 3
	case "platinum":
		return 4
	default:
		return 0
	}
}

func HigherTier(left, right string) string {
	if TierRank(left) >= TierRank(right) {
		return strings.ToLower(strings.TrimSpace(left))
	}
	return strings.ToLower(strings.TrimSpace(right))
}

func EligibleTier(tier string, allowed []string) bool {
	for _, candidate := range allowed {
		if strings.EqualFold(strings.TrimSpace(tier), strings.TrimSpace(candidate)) {
			return true
		}
	}
	return false
}

func NormalizeTier(tier string) string {
	tier = strings.ToLower(strings.TrimSpace(tier))
	if ValidTier(tier) {
		return tier
	}
	return "standard"
}

func NormalizeBenefitCode(code string) string {
	return strings.ToUpper(strings.TrimSpace(code))
}

func ValidateMemberID(memberID int) error {
	if memberID <= 0 {
		return fmt.Errorf("member id must be positive")
	}
	return nil
}

func ValidateBenefitCode(code string) error {
	code = NormalizeBenefitCode(code)
	if code == "" {
		return fmt.Errorf("benefit code must not be empty")
	}
	if len(code) > 64 {
		return fmt.Errorf("benefit code is too long")
	}
	return nil
}

func ValidateWindow(start, end time.Time) error {
	if start.IsZero() || end.IsZero() {
		return fmt.Errorf("window endpoints are required")
	}
	if end.Before(start) {
		return fmt.Errorf("window end precedes start")
	}
	return nil
}

func IsPremiumTier(tier string) bool { return TierRank(tier) >= 3 }

func IsBasicCode(code string) bool { return strings.HasPrefix(NormalizeBenefitCode(code), "BASIC-") }

func IsGoldCode(code string) bool { return strings.HasPrefix(NormalizeBenefitCode(code), "GOLD-") }

func IsPlatinumCode(code string) bool {
	return strings.HasPrefix(NormalizeBenefitCode(code), "PLATINUM-")
}

func HasKnownPrefix(code string) bool {
	code = NormalizeBenefitCode(code)
	return IsBasicCode(code) || IsGoldCode(code) || IsPlatinumCode(code)
}

func IsRecognizedTier(tier string) bool { return TierRank(tier) > 0 }
