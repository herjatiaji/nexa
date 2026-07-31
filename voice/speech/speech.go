package speech

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// SpeechChunk represents a single segment of text with its analyzed emotion and prosody.
type SpeechChunk struct {
	Text     string       `json:"text"`
	Emotion  EmotionTag   `json:"emotion"`
	Prosody  ProsodyParams `json:"prosody"`
	Priority int          `json:"priority"` // 0=normal, 1=emphasis, 2=critical
}

// SpeechPlan is the output of the Speech Planner — a fully analyzed speech plan
// ready for synthesis. Each chunk has its own emotion and prosody parameters.
type SpeechPlan struct {
	Chunks      []SpeechChunk    `json:"chunks"`
	Personality PersonalityConfig `json:"personality"`
	DominantEmotion EmotionTag   `json:"dominantEmotion"`
	Confidence  float64          `json:"confidence"`
	Timestamp   string           `json:"timestamp"`
}

// SpeechPlanner orchestrates the full Speech Intelligence pipeline:
// Text → Segmentation → Emotion Analysis → Prosody Mapping → Synthesis → Post-processing.
type SpeechPlanner struct {
	personality PersonalityConfig
	mu          sync.Mutex

	// OnChunkReady is called when a speech chunk has been analyzed (for UI feedback).
	OnChunkReady func(chunk SpeechChunk, index int, total int)

	// OnEmotionDetected is called when the dominant emotion for the response is determined.
	OnEmotionDetected func(emotion EmotionTag, confidence float64)
}

// NewSpeechPlanner creates a new SpeechPlanner with the given personality configuration.
func NewSpeechPlanner(personality PersonalityConfig) *SpeechPlanner {
	return &SpeechPlanner{
		personality: personality,
	}
}

// Plan analyzes text and produces a SpeechPlan with emotion-tagged, prosody-mapped chunks.
func (sp *SpeechPlanner) Plan(text string) *SpeechPlan {
	sp.mu.Lock()
	defer sp.mu.Unlock()

	// Segment text into natural speech chunks
	chunks := segmentText(text)

	// Analyze emotion and map prosody for each chunk
	speechChunks := make([]SpeechChunk, len(chunks))
	emotionCounts := make(map[EmotionTag]float64)

	for i, chunkText := range chunks {
		emotion, confidence := AnalyzeEmotion(chunkText)
		prosody := GetProsody(emotion, sp.personality)

		speechChunks[i] = SpeechChunk{
			Text:     chunkText,
			Emotion:  emotion,
			Prosody:  prosody,
			Priority: 0,
		}

		emotionCounts[emotion] += confidence

		if sp.OnChunkReady != nil {
			sp.OnChunkReady(speechChunks[i], i, len(chunks))
		}
	}

	// Determine dominant emotion across all chunks
	dominantEmotion := EmotionNeutral
	bestScore := 0.0
	for tag, score := range emotionCounts {
		if score > bestScore {
			dominantEmotion = tag
			bestScore = score
		}
	}

	if sp.OnEmotionDetected != nil {
		conf := bestScore
		if conf > 1.0 {
			conf = 1.0
		}
		sp.OnEmotionDetected(dominantEmotion, conf)
	}

	return &SpeechPlan{
		Chunks:          speechChunks,
		Personality:     sp.personality,
		DominantEmotion: dominantEmotion,
		Confidence:      bestScore,
		Timestamp:       time.Now().Format("15:04:05"),
	}
}

// SynthesizeFunc is the function signature for TTS synthesis with prosody control.
// It takes text and prosody params, returns the path to the generated WAV file.
type SynthesizeFunc func(text string, prosody ProsodyParams) (string, error)

// Execute synthesizes all chunks in the SpeechPlan using the provided synthesis function.
// For short responses (<=100 words), synthesizes as a single chunk for natural flow.
// For longer responses, synthesizes per-chunk and concatenates with crossfade.
func (sp *SpeechPlanner) Execute(plan *SpeechPlan, synthesize SynthesizeFunc) (string, error) {
	if len(plan.Chunks) == 0 {
		return "", fmt.Errorf("empty speech plan")
	}

	// Count total words to decide single vs multi-chunk synthesis
	totalWords := 0
	for _, chunk := range plan.Chunks {
		totalWords += len(strings.Fields(chunk.Text))
	}

	// For short responses, use single-chunk mode for most natural delivery
	if totalWords <= 100 || len(plan.Chunks) == 1 {
		// Merge all text, use dominant emotion's prosody
		fullText := ""
		for _, chunk := range plan.Chunks {
			fullText += chunk.Text + " "
		}
		fullText = strings.TrimSpace(fullText)

		dominantProsody := GetProsody(plan.DominantEmotion, plan.Personality)

		wavPath, err := synthesize(fullText, dominantProsody)
		if err != nil {
			return "", err
		}

		// Post-process: normalize volume and trim silence
		_ = NormalizeVolume(wavPath, -3.0)
		_ = TrimSilence(wavPath, 150)

		return wavPath, nil
	}

	// Multi-chunk synthesis for longer responses
	var wavPaths []string
	for _, chunk := range plan.Chunks {
		wavPath, err := synthesize(chunk.Text, chunk.Prosody)
		if err != nil {
			// Skip failed chunks but continue with others
			continue
		}

		// Post-process each chunk
		_ = NormalizeVolume(wavPath, -3.0)
		_ = TrimSilence(wavPath, 150)

		wavPaths = append(wavPaths, wavPath)
	}

	if len(wavPaths) == 0 {
		return "", fmt.Errorf("all speech chunks failed synthesis")
	}

	if len(wavPaths) == 1 {
		return wavPaths[0], nil
	}

	// Concatenate chunks with natural crossfade (30ms)
	outputPath := filepath.Join(os.TempDir(), fmt.Sprintf("nexa_speech_%d.wav", time.Now().UnixNano()))
	if err := ConcatWAVs(wavPaths, outputPath, 30); err != nil {
		// Fallback: return first chunk only
		return wavPaths[0], nil
	}

	// Clean up individual chunk files
	for _, p := range wavPaths {
		_ = os.Remove(p)
	}

	return outputPath, nil
}

// segmentText splits text into natural speech segments at sentence boundaries.
// Groups short consecutive sentences together to avoid choppy delivery.
func segmentText(text string) []string {
	// Split on sentence-ending punctuation
	var segments []string
	current := ""

	sentences := splitSentences(text)

	for _, sentence := range sentences {
		sentence = strings.TrimSpace(sentence)
		if sentence == "" {
			continue
		}

		wordCount := len(strings.Fields(current + " " + sentence))

		// Group sentences into chunks of 15–40 words for natural phrasing
		if current == "" {
			current = sentence
		} else if wordCount <= 40 {
			current += " " + sentence
		} else {
			segments = append(segments, strings.TrimSpace(current))
			current = sentence
		}
	}

	if strings.TrimSpace(current) != "" {
		segments = append(segments, strings.TrimSpace(current))
	}

	if len(segments) == 0 && strings.TrimSpace(text) != "" {
		segments = []string{strings.TrimSpace(text)}
	}

	return segments
}

// splitSentences splits text into sentences using common delimiters.
func splitSentences(text string) []string {
	var sentences []string
	current := ""

	runes := []rune(text)
	for i, r := range runes {
		current += string(r)

		// Check for sentence boundary: period, exclamation, question mark
		// followed by space or end of text
		if (r == '.' || r == '!' || r == '?') && (i == len(runes)-1 || runes[i+1] == ' ' || runes[i+1] == '\n') {
			trimmed := strings.TrimSpace(current)
			if trimmed != "" {
				sentences = append(sentences, trimmed)
			}
			current = ""
		}

		// Also split on newlines (paragraph boundaries)
		if r == '\n' {
			trimmed := strings.TrimSpace(current)
			if trimmed != "" {
				sentences = append(sentences, trimmed)
			}
			current = ""
		}
	}

	// Add remaining text
	trimmed := strings.TrimSpace(current)
	if trimmed != "" {
		sentences = append(sentences, trimmed)
	}

	return sentences
}

// GetPersonality returns the current personality config.
func (sp *SpeechPlanner) GetPersonality() PersonalityConfig {
	return sp.personality
}

// SetPersonality updates the personality config.
func (sp *SpeechPlanner) SetPersonality(p PersonalityConfig) {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	sp.personality = p
}
