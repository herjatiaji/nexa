package speech

// PersonalityConfig defines the voice character and emotional expression style
// for NEXA's speech synthesis. This is the "soul" of how NEXA sounds.
//
// Design philosophy: NEXA speaks like a competent British professional —
// warm, articulate, composed. Emotions are conveyed through subtle prosody
// shifts, not dramatic voice changes. Think BBC newsreader meets helpful colleague.
type PersonalityConfig struct {
	// Name of this personality profile.
	Name string `json:"name"`

	// BaseSpeed is the default speech rate multiplier (1.0 = normal).
	// Slightly below 1.0 gives a measured, thoughtful delivery.
	BaseSpeed float64 `json:"baseSpeed"`

	// BasePitch is the default pitch offset in semitones (0.0 = natural register).
	BasePitch float64 `json:"basePitch"`

	// Warmth controls how "friendly" the voice sounds (0.0–1.0).
	// Higher values add slightly more pause variation and softer onset.
	Warmth float64 `json:"warmth"`

	// Formality controls register (0.0=casual, 1.0=formal).
	// Higher values produce more measured pacing and less pitch variation.
	Formality float64 `json:"formality"`

	// EmotionDamping attenuates raw emotion prosody changes (0.0–1.0).
	// 0.0 = completely flat (no emotion), 1.0 = full theatrical expression.
	// 0.5–0.7 = natural human range — emotions are felt but not exaggerated.
	EmotionDamping float64 `json:"emotionDamping"`
}

// DefaultPersonality returns the standard NEXA voice persona:
// British professional, warm, articulate, with subtle emotional expression.
func DefaultPersonality() PersonalityConfig {
	return PersonalityConfig{
		Name:           "NEXA",
		BaseSpeed:      1.00,
		BasePitch:      0.0,
		Warmth:         0.70,
		Formality:      0.75,
		EmotionDamping: 0.55,
	}
}

// CasualPersonality returns a more relaxed, conversational persona.
func CasualPersonality() PersonalityConfig {
	return PersonalityConfig{
		Name:           "NEXA Casual",
		BaseSpeed:      1.03,
		BasePitch:      0.1,
		Warmth:         0.85,
		Formality:      0.40,
		EmotionDamping: 0.70,
	}
}

// ProfessionalPersonality returns a more formal, measured persona.
func ProfessionalPersonality() PersonalityConfig {
	return PersonalityConfig{
		Name:           "NEXA Professional",
		BaseSpeed:      0.97,
		BasePitch:      -0.1,
		Warmth:         0.55,
		Formality:      0.90,
		EmotionDamping: 0.40,
	}
}
