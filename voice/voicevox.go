package voice

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

// EmotionPayload holds emotion metadata passed to VOICEVOX synthesis.
type EmotionPayload struct {
	Text    string  `json:"text"`
	Emotion string  `json:"emotion"` // happy, sad, thinking, excited, neutral
	Speed   float64 `json:"speed,omitempty"`
	Pitch   float64 `json:"pitch,omitempty"`
}

// VoicevoxClient interacts with local VOICEVOX engine REST API (http://localhost:50021).
type VoicevoxClient struct {
	BaseURL string
	Speaker int // 3: Zundamon (Normal), 1: Shikikou, 2: Metan, 10: Tsukuyomi
	client  *http.Client
}

// NewVoicevoxClient creates a new VOICEVOX API client.
func NewVoicevoxClient(baseURL string, speaker int) *VoicevoxClient {
	if baseURL == "" {
		baseURL = "http://localhost:50021"
	}
	if speaker <= 0 {
		speaker = 3 // Zundamon
	}
	return &VoicevoxClient{
		BaseURL: baseURL,
		Speaker: speaker,
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

// IsAvailable checks if the local VOICEVOX engine is running.
func (vc *VoicevoxClient) IsAvailable() bool {
	resp, err := vc.client.Get(vc.BaseURL + "/version")
	if err != nil || resp.StatusCode != http.StatusOK {
		return false
	}
	_ = resp.Body.Close()
	return true
}

// SynthesizeSpeech converts text + emotion payload to WAV audio via VOICEVOX REST API.
func (vc *VoicevoxClient) SynthesizeSpeech(payload EmotionPayload) (string, error) {
	if !vc.IsAvailable() {
		return "", fmt.Errorf("VOICEVOX engine is not running on %s", vc.BaseURL)
	}

	// 1. Create Audio Query
	queryURL := fmt.Sprintf("%s/audio_query?speaker=%d&text=%s", vc.BaseURL, vc.Speaker, url.QueryEscape(payload.Text))
	reqQuery, err := http.NewRequest("POST", queryURL, nil)
	if err != nil {
		return "", err
	}

	respQuery, err := vc.client.Do(reqQuery)
	if err != nil || respQuery.StatusCode != http.StatusOK {
		return "", fmt.Errorf("audio query failed with status %d", respQuery.StatusCode)
	}
	defer respQuery.Body.Close()

	queryJSON, err := io.ReadAll(respQuery.Body)
	if err != nil {
		return "", err
	}

	// 2. Adjust Audio Parameters according to Phase 2 Emotion JSON Payload
	var queryData map[string]interface{}
	if err := json.Unmarshal(queryJSON, &queryData); err == nil {
		speed := 1.05
		pitch := 0.02
		intonation := 1.1

		switch payload.Emotion {
		case "happy", "excited":
			speed = 1.12
			pitch = 0.08
			intonation = 1.25
		case "thinking":
			speed = 0.95
			pitch = -0.02
			intonation = 0.95
		case "sad", "error":
			speed = 0.88
			pitch = -0.08
			intonation = 0.85
		}

		if payload.Speed > 0 {
			speed = payload.Speed
		}
		if payload.Pitch != 0 {
			pitch = payload.Pitch
		}

		queryData["speedScale"] = speed
		queryData["pitchScale"] = pitch
		queryData["intonationScale"] = intonation
		queryJSON, _ = json.Marshal(queryData)
	}

	// 3. Synthesize WAV Audio
	synthURL := fmt.Sprintf("%s/synthesis?speaker=%d", vc.BaseURL, vc.Speaker)
	reqSynth, err := http.NewRequest("POST", synthURL, bytes.NewBuffer(queryJSON))
	if err != nil {
		return "", err
	}
	reqSynth.Header.Set("Content-Type", "application/json")

	respSynth, err := vc.client.Do(reqSynth)
	if err != nil || respSynth.StatusCode != http.StatusOK {
		return "", fmt.Errorf("voicevox synthesis failed with status %d", respSynth.StatusCode)
	}
	defer respSynth.Body.Close()

	tempWav := filepath.Join(os.TempDir(), fmt.Sprintf("nexa_voicevox_%d.wav", os.Getpid()))
	out, err := os.Create(tempWav)
	if err != nil {
		return "", err
	}
	defer out.Close()

	_, err = io.Copy(out, respSynth.Body)
	if err != nil {
		return "", err
	}

	return tempWav, nil
}
