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

// VoicevoxClient interacts with local VOICEVOX engine REST API (http://localhost:50021).
type VoicevoxClient struct {
	BaseURL string
	Speaker int // e.g. 3 (Zundamon Normal), 1 (Shikikou), 2 (Metan)
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

// SynthesizeSpeech converts text to WAV audio via VOICEVOX REST API.
func (vc *VoicevoxClient) SynthesizeSpeech(text string) (string, error) {
	if !vc.IsAvailable() {
		return "", fmt.Errorf("VOICEVOX engine is not running on %s", vc.BaseURL)
	}

	// 1. Create Audio Query
	queryURL := fmt.Sprintf("%s/audio_query?speaker=%d&text=%s", vc.BaseURL, vc.Speaker, url.QueryEscape(text))
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

	// Tune query JSON for cute pace and pitch
	var queryData map[string]interface{}
	if err := json.Unmarshal(queryJSON, &queryData); err == nil {
		queryData["speedScale"] = 1.08
		queryData["pitchScale"] = 0.05
		queryJSON, _ = json.Marshal(queryData)
	}

	// 2. Synthesize WAV Audio
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
