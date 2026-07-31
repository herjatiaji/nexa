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
	"strings"
	"time"
)

// EmotionPayload holds emotion metadata passed to VOICEVOX synthesis.
type EmotionPayload struct {
	Text    string  `json:"text"`
	Emotion string  `json:"emotion"` // happy, sad, thinking, excited, neutral
	Speed   float64 `json:"speed,omitempty"`
	Pitch   float64 `json:"pitch,omitempty"`
}

// VoicevoxStyle represents a style variant for a speaker (e.g. normal, cheerful).
type VoicevoxStyle struct {
	Name string `json:"name"`
	ID   int    `json:"id"`
}

// VoicevoxSpeaker represents a character from /speakers endpoint.
type VoicevoxSpeaker struct {
	Name   string          `json:"name"`
	Styles []VoicevoxStyle `json:"styles"`
}

// VoicevoxClient interacts with local VOICEVOX engine REST API (http://localhost:50021).
type VoicevoxClient struct {
	BaseURL string
	Speaker int // e.g. 3 (Zundamon Normal), 1 (Shikoku Metan)
	client  *http.Client
}

// NewVoicevoxClient creates a new VOICEVOX API client with dynamic speaker discovery.
func NewVoicevoxClient(baseURL string, speaker int) *VoicevoxClient {
	if baseURL == "" {
		baseURL = "http://localhost:50021"
	}
	client := &VoicevoxClient{
		BaseURL: baseURL,
		Speaker: speaker,
		client:  &http.Client{Timeout: 10 * time.Second},
	}

	if speaker <= 0 && client.IsAvailable() {
		// Auto-discover Shikoku Metan (mature female character voice)
		if id, err := client.FindSpeakerID("四国めたん", "ノーマル"); err == nil {
			client.Speaker = id
		} else {
			client.Speaker = 2 // Fallback to Shikoku Metan Normal default ID
		}
	} else if speaker <= 0 {
		client.Speaker = 2
	}

	return client
}

// IsAvailable checks if the local VOICEVOX engine is running on http://localhost:50021.
func (vc *VoicevoxClient) IsAvailable() bool {
	resp, err := vc.client.Get(vc.BaseURL + "/version")
	if err != nil || resp.StatusCode != http.StatusOK {
		return false
	}
	_ = resp.Body.Close()
	return true
}

// FetchSpeakers queries http://localhost:50021/speakers to get all installed characters & styles.
func (vc *VoicevoxClient) FetchSpeakers() ([]VoicevoxSpeaker, error) {
	resp, err := vc.client.Get(vc.BaseURL + "/speakers")
	if err != nil || resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch speakers from VOICEVOX: %v", err)
	}
	defer resp.Body.Close()

	var speakers []VoicevoxSpeaker
	if err := json.NewDecoder(resp.Body).Decode(&speakers); err != nil {
		return nil, err
	}
	return speakers, nil
}

// FindSpeakerID dynamically resolves speaker ID by character name (e.g., "ずんだもん" or "Zundamon") and style.
func (vc *VoicevoxClient) FindSpeakerID(charName string, styleName string) (int, error) {
	speakers, err := vc.FetchSpeakers()
	if err != nil {
		return 0, err
	}

	for _, sp := range speakers {
		if strings.Contains(strings.ToLower(sp.Name), strings.ToLower(charName)) ||
			(charName == "Zundamon" && strings.Contains(sp.Name, "ずんだもん")) ||
			(charName == "Shikoku Metan" && strings.Contains(sp.Name, "四国めたん")) {
			for _, st := range sp.Styles {
				if styleName == "" || strings.Contains(strings.ToLower(st.Name), strings.ToLower(styleName)) || st.Name == "ノーマル" {
					return st.ID, nil
				}
			}
		}
	}

	// Return first style ID of first speaker as fallback
	if len(speakers) > 0 && len(speakers[0].Styles) > 0 {
		return speakers[0].Styles[0].ID, nil
	}
	return 3, nil
}

// SynthesizeSpeech converts text + emotion payload to WAV audio via VOICEVOX REST API.
func (vc *VoicevoxClient) SynthesizeSpeech(payload EmotionPayload) (string, error) {
	if !vc.IsAvailable() {
		return "", fmt.Errorf("VOICEVOX engine is not running on %s", vc.BaseURL)
	}

	// Step 1: Create Audio Query (POST /audio_query?speaker=ID&text=TEXT)
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

	// Step 2: Modulate Query JSON with Cheerful Audio Parameters (pitchScale, speedScale, intonationScale)
	var queryData map[string]interface{}
	if err := json.Unmarshal(queryJSON, &queryData); err == nil {
		// Cheerful Genki Default
		speed := 1.10
		pitch := 0.05
		intonation := 1.20

		switch payload.Emotion {
		case "happy", "excited":
			speed = 1.15
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

	// Step 3: Synthesize WAV Audio (POST /synthesis?speaker=ID)
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
