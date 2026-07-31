package cognitive

// SocialState manages attention scoring and social interruption rules.
type SocialState struct {
	AttentionScore   float64 `json:"attention_score"`   // 0.0 (idle) to 1.0 (deep focus)
	StressLevel      float64 `json:"stress_level"`      // 0.0 (relaxed) to 1.0 (high stress)
	CanInterrupt     bool    `json:"can_interrupt"`     // Is NEXA permitted to interrupt user right now?
	UrgencyThreshold float64 `json:"urgency_threshold"` // Minimum decision score required to interrupt
}

// NewSocialState creates default SocialState.
func NewSocialState() SocialState {
	return SocialState{
		AttentionScore:   0.5,
		StressLevel:      0.2,
		CanInterrupt:     true,
		UrgencyThreshold: 0.5,
	}
}

// PermissionFor returns permission multiplier (0.0 to 1.0) for a given proposed action.
func (s SocialState) PermissionFor(action string) float64 {
	if !s.CanInterrupt && action != "SILENT" && action != "OBSERVE" {
		return 0.0
	}
	if action == "SPEAK" && s.AttentionScore > 0.8 {
		// User in deep focus: penalize spoken interruptions unless critical
		return 0.2
	}
	if action == "TOAST" {
		return 0.8
	}
	return 1.0
}
