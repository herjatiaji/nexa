package core

import (
	"time"
)

// EmotionState defines the visual expression and mood of the NEXA Mascot Companion.
type EmotionState string

const (
	EmotionIdle      EmotionState = "idle"      // ( ◉ ◉ )
	EmotionListening EmotionState = "listening" // ( ⚡ ⚡ )
	EmotionThinking  EmotionState = "thinking"  // ( ◌ ◌ )
	EmotionExecuting EmotionState = "executing" // ( ⚙️ ⚙️ )
	EmotionSpeaking  EmotionState = "speaking"  // ( 🔊 🔊 )
	EmotionHappy     EmotionState = "happy"     // ( 😎 😎 )
	EmotionConfused  EmotionState = "confused"  // ( 😕 😕 )
	EmotionYawn      EmotionState = "yawn"      // ( 💤 💤 )
)

// MascotState represents the complete emotion context delivered to Wails frontend.
type MascotState struct {
	Emotion     EmotionState `json:"emotion"`
	Message     string       `json:"message,omitempty"`
	EyeSymbol   string       `json:"eyeSymbol"`
	AuraColor   string       `json:"auraColor"`
	Timestamp   string       `json:"timestamp"`
}

// GetMascotExpression maps an EmotionState to eye symbols and aura color.
func GetMascotExpression(state EmotionState, customMsg string) MascotState {
	eye := "( ◉ ◉ )"
	aura := "#00e5ff" // Cyan

	switch state {
	case EmotionListening:
		eye = "( ⚡ ⚡ )"
		aura = "#00e676" // Emerald Green
	case EmotionThinking:
		eye = "( ◌ ◌ )"
		aura = "#ba68c8" // Purple
	case EmotionExecuting:
		eye = "( ⚙️ ⚙️ )"
		aura = "#ffd740" // Amber Gold
	case EmotionSpeaking:
		eye = "( 🔊 🔊 )"
		aura = "#00b0ff" // Bright Blue
	case EmotionHappy:
		eye = "( 😎 😎 )"
		aura = "#76ff03" // Vibrant Lime Green
	case EmotionConfused:
		eye = "( 😕 😕 )"
		aura = "#ff5252" // Coral Red
	case EmotionYawn:
		eye = "( 💤 💤 )"
		aura = "#80deea" // Soft Aquamarine
	}

	return MascotState{
		Emotion:   state,
		Message:   customMsg,
		EyeSymbol: eye,
		AuraColor: aura,
		Timestamp: time.Now().Format("15:04:05"),
	}
}
