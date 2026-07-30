package voice

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// STT handles speech recognition input using OpenAI Whisper.
type STT struct {
	Enabled    bool
	GroqAPIKey string
	whisper    *WhisperSTT
}

// NewSTT creates a new STT handler with Whisper support.
func NewSTT(enabled bool, groqAPIKey string) *STT {
	return &STT{
		Enabled:    enabled,
		GroqAPIKey: groqAPIKey,
		whisper:    NewWhisperSTT(groqAPIKey),
	}
}

// Listen listens for voice input using OpenAI Whisper STT or falls back to text prompt.
func (s *STT) Listen() (string, error) {
	if !s.Enabled {
		return "", fmt.Errorf("STT disabled")
	}

	// 1. Record audio using winmm.dll (4 seconds)
	wavFile, err := RecordWAV(4)
	if err == nil {
		defer os.Remove(wavFile)

		// 2. Transcribe using OpenAI Whisper (Groq Large v3 API or local whisper.cpp)
		text, err := s.whisper.Transcribe(wavFile)
		if err == nil && text != "" {
			return text, nil
		}
	}

	// Fallback prompt if STT speech input returns empty
	fmt.Print("💬 Type command: ")
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(input), nil
}
