package speech

import (
	"regexp"
	"strings"
)

// EmotionTag represents the detected emotional tone of a text segment.
type EmotionTag string

const (
	EmotionNeutral    EmotionTag = "neutral"
	EmotionHappy      EmotionTag = "happy"
	EmotionExcited    EmotionTag = "excited"
	EmotionThoughtful EmotionTag = "thoughtful"
	EmotionUrgent     EmotionTag = "urgent"
	EmotionSad        EmotionTag = "sad"
	EmotionConfident  EmotionTag = "confident"
)

// emotionKeywords maps EmotionTag to lexical keyword sets for detection.
var emotionKeywords = map[EmotionTag][]string{
	EmotionHappy: {
		"great", "awesome", "wonderful", "fantastic", "excellent", "perfect",
		"glad", "happy", "pleased", "delighted", "brilliant", "love", "lovely",
		"nice", "good news", "congratulations", "well done", "cheers",
		"thank you", "thanks", "appreciate", "grateful",
	},
	EmotionExcited: {
		"amazing", "incredible", "wow", "unbelievable", "extraordinary",
		"mind-blowing", "spectacular", "outstanding", "superb", "phenomenal",
		"absolutely", "definitely", "exciting", "thrilling",
	},
	EmotionSad: {
		"unfortunately", "sorry", "apolog", "regret", "sadly", "bad news",
		"issue", "problem", "error", "fail", "broken", "crash", "bug",
		"trouble", "difficult", "struggling", "concern", "worry", "afraid",
		"cannot", "unable", "impossible", "lost",
	},
	EmotionThoughtful: {
		"consider", "perhaps", "maybe", "might", "could", "think about",
		"interesting", "hmm", "well", "let me", "it seems", "appears",
		"actually", "in fact", "essentially", "fundamentally", "nuance",
		"however", "although", "on the other hand", "perspective",
	},
	EmotionUrgent: {
		"urgent", "immediately", "critical", "asap", "right now", "warning",
		"danger", "alert", "emergency", "important", "must", "need to",
		"hurry", "quickly", "time-sensitive", "deadline",
	},
	EmotionConfident: {
		"certainly", "absolutely", "clearly", "obviously", "without doubt",
		"guaranteed", "sure", "confident", "precisely", "exactly",
		"no question", "undoubtedly", "indeed", "of course",
		"here's how", "the solution is", "i recommend", "the answer is",
	},
}

var (
	exclamationRegex = regexp.MustCompile(`!`)
	questionRegex    = regexp.MustCompile(`\?`)
	ellipsisRegex    = regexp.MustCompile(`\.{2,}|…`)
	capsWordRegex    = regexp.MustCompile(`\b[A-Z]{2,}\b`)
	emojiHappyRegex  = regexp.MustCompile(`[😊😄😁🎉🥳👍✅🎊💚💙🌟⭐]`)
	emojiSadRegex    = regexp.MustCompile(`[😢😔😞😟❌💔😕🥲]`)
	emojiExcitedRe   = regexp.MustCompile(`[🔥🚀💥⚡🎯💪🏆✨😍]`)
)

// EmotionScore holds per-emotion detection confidence.
type EmotionScore struct {
	Tag   EmotionTag
	Score float64
}

// AnalyzeEmotion detects the dominant emotion from a text segment using
// lexical keyword matching, punctuation analysis, and emoji detection.
// Returns the strongest detected emotion and its confidence score (0.0–1.0).
func AnalyzeEmotion(text string) (EmotionTag, float64) {
	lower := strings.ToLower(text)
	wordCount := len(strings.Fields(text))
	if wordCount == 0 {
		return EmotionNeutral, 0.0
	}

	scores := make(map[EmotionTag]float64)

	// 1. Lexical keyword matching (primary signal)
	for tag, keywords := range emotionKeywords {
		hits := 0
		for _, kw := range keywords {
			if strings.Contains(lower, kw) {
				hits++
			}
		}
		if hits > 0 {
			// Normalize by word count to avoid bias toward long text
			density := float64(hits) / float64(wordCount)
			scores[tag] += density * 3.0 // Weight lexical signals
			if hits >= 3 {
				scores[tag] += 0.2 // Bonus for multiple keyword matches
			}
		}
	}

	// 2. Punctuation analysis (secondary signal)
	exclamations := len(exclamationRegex.FindAllString(text, -1))
	questions := len(questionRegex.FindAllString(text, -1))
	ellipses := len(ellipsisRegex.FindAllString(text, -1))
	capsWords := len(capsWordRegex.FindAllString(text, -1))

	if exclamations >= 2 {
		scores[EmotionExcited] += 0.3
	} else if exclamations == 1 {
		scores[EmotionHappy] += 0.1
	}

	if questions >= 2 {
		scores[EmotionThoughtful] += 0.15
	}

	if ellipses >= 1 {
		scores[EmotionThoughtful] += 0.2
	}

	if capsWords >= 2 {
		scores[EmotionUrgent] += 0.25
	}

	// 3. Emoji detection (tertiary signal)
	if emojiHappyRegex.MatchString(text) {
		scores[EmotionHappy] += 0.25
	}
	if emojiSadRegex.MatchString(text) {
		scores[EmotionSad] += 0.25
	}
	if emojiExcitedRe.MatchString(text) {
		scores[EmotionExcited] += 0.25
	}

	// 4. Sentence structure analysis
	if wordCount <= 8 && exclamations > 0 {
		// Short exclamatory → urgent or excited
		scores[EmotionUrgent] += 0.1
	}
	if wordCount >= 40 {
		// Long explanatory → thoughtful
		scores[EmotionThoughtful] += 0.1
	}

	// Find dominant emotion
	bestTag := EmotionNeutral
	bestScore := 0.0
	for tag, score := range scores {
		if score > bestScore {
			bestTag = tag
			bestScore = score
		}
	}

	// Clamp confidence to [0, 1]
	if bestScore > 1.0 {
		bestScore = 1.0
	}

	// Require minimum threshold to avoid false positives
	if bestScore < 0.08 {
		return EmotionNeutral, 0.0
	}

	return bestTag, bestScore
}

// AnalyzeChunkEmotions analyzes emotion for each chunk in a slice of text segments.
func AnalyzeChunkEmotions(chunks []string) []EmotionTag {
	emotions := make([]EmotionTag, len(chunks))
	for i, chunk := range chunks {
		tag, _ := AnalyzeEmotion(chunk)
		emotions[i] = tag
	}
	return emotions
}
