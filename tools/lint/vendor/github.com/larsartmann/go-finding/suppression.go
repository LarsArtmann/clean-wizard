package finding

import "time"

// SuppressionKind indicates where a suppression was defined.
type SuppressionKind string

// Suppression kinds indicate where a suppression was defined.
const (
	SuppressionInSource SuppressionKind = "in-source" // e.g., //nolint, //lint:ignore
	SuppressionInConfig SuppressionKind = "in-config" // Config file rules
	SuppressionInReview SuppressionKind = "in-review" // Accepted as false positive
)

// Suppression represents a suppressed finding.
type Suppression struct {
	Kind      SuppressionKind `json:"kind"`                // Where the suppression is defined
	Rule      RuleName        `json:"rule"`                // Which rule is suppressed
	Reason    string          `json:"reason"`              // Why it's suppressed
	ExpiresAt *time.Time      `json:"expiresAt,omitempty"` // Optional expiry
}

// IsExpired returns true if the suppression has expired relative to now.
// A suppression is expired when now is strictly after ExpiresAt.
// At the exact ExpiresAt instant, the suppression is still considered active
// (valid through that moment, expired any time after).
func (s *Suppression) IsExpired(now time.Time) bool {
	if s == nil || s.ExpiresAt == nil {
		return false
	}

	return now.After(*s.ExpiresAt)
}

// IsValid returns true if the suppression has a valid kind and a non-empty rule.
// The kind must be one of the predefined SuppressionKind constants (checked
// via Kind.IsValid), not just any non-empty string. This prevents typos like
// "in-source" from silently passing validation.
func (s *Suppression) IsValid() bool {
	if s == nil {
		return false
	}

	return s.Kind.IsValid() && s.Rule != ""
}

// IsActive returns true if the suppression is valid and not expired.
// This combines IsValid and !IsExpired into a single check.
func (s *Suppression) IsActive(now time.Time) bool {
	return s.IsValid() && !s.IsExpired(now)
}

// IsValid returns true if the suppression kind is a recognized value.
func (k SuppressionKind) IsValid() bool {
	switch k {
	case SuppressionInSource, SuppressionInConfig, SuppressionInReview:
		return true
	}

	return false
}
