package speech

// ProsodyParams defines concrete TTS synthesis parameters derived from emotion analysis.
// These values are tuned to produce natural-sounding speech that subtly conveys emotion
// without sounding theatrical or robotic.
type ProsodyParams struct {
	// SpeedScale controls speech rate. Piper uses length_scale where:
	// <1.0 = faster, 1.0 = normal, >1.0 = slower.
	// We invert here: our SpeedScale >1 = faster delivery, <1 = slower.
	// Converted to length_scale as: lengthScale = 1.0 / SpeedScale
	SpeedScale float64 `json:"speedScale"`

	// PitchShift in semitones (-2.0 to +2.0). Subtle shifts convey mood.
	PitchShift float64 `json:"pitchShift"`

	// SentenceSilence is pause duration between sentences in seconds (0.05–0.8).
	// Shorter = energetic, longer = contemplative.
	SentenceSilence float64 `json:"sentenceSilence"`

	// Emphasis controls intonation intensity / expressiveness (0.7–1.3).
	// Higher = more expressive pitch contour.
	Emphasis float64 `json:"emphasis"`
}

// emotionProsodyMap defines the raw prosody profile for each emotion.
// These are "ideal" values before personality damping is applied.
var emotionProsodyMap = map[EmotionTag]ProsodyParams{
	EmotionNeutral: {
		SpeedScale:      1.00,
		PitchShift:      0.0,
		SentenceSilence: 0.25,
		Emphasis:        1.00,
	},
	EmotionHappy: {
		SpeedScale:      1.05,
		PitchShift:      0.4,
		SentenceSilence: 0.20,
		Emphasis:        1.10,
	},
	EmotionExcited: {
		SpeedScale:      1.10,
		PitchShift:      0.7,
		SentenceSilence: 0.15,
		Emphasis:        1.20,
	},
	EmotionThoughtful: {
		SpeedScale:      0.92,
		PitchShift:      -0.2,
		SentenceSilence: 0.40,
		Emphasis:        0.90,
	},
	EmotionUrgent: {
		SpeedScale:      1.12,
		PitchShift:      0.3,
		SentenceSilence: 0.10,
		Emphasis:        1.15,
	},
	EmotionSad: {
		SpeedScale:      0.90,
		PitchShift:      -0.4,
		SentenceSilence: 0.45,
		Emphasis:        0.85,
	},
	EmotionConfident: {
		SpeedScale:      1.02,
		PitchShift:      0.15,
		SentenceSilence: 0.30,
		Emphasis:        1.08,
	},
}

// GetProsody returns the prosody parameters for a given emotion tag,
// modulated by the personality config to produce natural, non-theatrical speech.
func GetProsody(emotion EmotionTag, personality PersonalityConfig) ProsodyParams {
	raw, ok := emotionProsodyMap[emotion]
	if !ok {
		raw = emotionProsodyMap[EmotionNeutral]
	}

	damping := personality.EmotionDamping

	// Apply personality damping: final = base + (emotion_delta * damping)
	// This ensures emotions are felt but subtle — like a real human speaker.
	finalSpeed := personality.BaseSpeed + (raw.SpeedScale-1.0)*damping
	finalPitch := personality.BasePitch + raw.PitchShift*damping
	finalSilence := raw.SentenceSilence // Silence timing is less affected by personality
	finalEmphasis := 1.0 + (raw.Emphasis-1.0)*damping

	// Clamp to safe ranges for natural-sounding output
	finalSpeed = clamp(finalSpeed, 0.75, 1.25)
	finalPitch = clamp(finalPitch, -1.5, 1.5)
	finalSilence = clamp(finalSilence, 0.05, 0.80)
	finalEmphasis = clamp(finalEmphasis, 0.70, 1.30)

	return ProsodyParams{
		SpeedScale:      finalSpeed,
		PitchShift:      finalPitch,
		SentenceSilence: finalSilence,
		Emphasis:        finalEmphasis,
	}
}

// clamp restricts a float64 value to [min, max].
func clamp(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
