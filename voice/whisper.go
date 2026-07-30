package voice

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// WhisperSTT handles Speech-To-Text using OpenAI Whisper model (via Groq API or local whisper.cpp).
type WhisperSTT struct {
	GroqAPIKey string
}

// NewWhisperSTT creates a new Whisper STT engine.
func NewWhisperSTT(groqAPIKey string) *WhisperSTT {
	return &WhisperSTT{
		GroqAPIKey: groqAPIKey,
	}
}

// RecordWAV records mic audio for durationSec and saves it to a temporary WAV file.
func RecordWAV(durationSec int) (string, error) {
	if runtime.GOOS != "windows" {
		return "", fmt.Errorf("audio recording currently supported on Windows")
	}

	tempWav := filepath.Join(os.TempDir(), fmt.Sprintf("jarvis_record_%d.wav", time.Now().UnixNano()))
	escapedWav := strings.ReplaceAll(tempWav, "'", "''")

	psScript := fmt.Sprintf(`
Add-Type -TypeDefinition @"
using System;
using System.Runtime.InteropServices;
public class MciAudioRecorder {
    [DllImport("winmm.dll", CharSet = CharSet.Auto)]
    public static extern int mciSendString(string command, string buffer, int bufferSize, IntPtr hwndCallback);
}
"@

$wavFile = '%s'
[MciAudioRecorder]::mciSendString("open new type waveaudio alias recsound", $null, 0, [IntPtr]::Zero) | Out-Null
[MciAudioRecorder]::mciSendString("set recsound time format ms bitspersample 16 channels 1 samplespersec 16000 bytespersec 32000 alignment 2", $null, 0, [IntPtr]::Zero) | Out-Null
[MciAudioRecorder]::mciSendString("record recsound", $null, 0, [IntPtr]::Zero) | Out-Null

Start-Sleep -Seconds %d

[MciAudioRecorder]::mciSendString("save recsound '" + $wavFile + "'", $null, 0, [IntPtr]::Zero) | Out-Null
[MciAudioRecorder]::mciSendString("close recsound", $null, 0, [IntPtr]::Zero) | Out-Null
`, escapedWav, durationSec)

	cmd := exec.Command("powershell", "-NoProfile", "-Command", psScript)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to record audio: %w", err)
	}

	if info, err := os.Stat(tempWav); err != nil || info.Size() == 0 {
		return "", fmt.Errorf("recorded audio file is empty or missing")
	}

	return tempWav, nil
}

// Transcribe transcribes the WAV audio file using OpenAI Whisper (Groq Whisper API or local whisper.cpp).
func (w *WhisperSTT) Transcribe(wavPath string) (string, error) {
	if w.GroqAPIKey != "" {
		text, err := w.transcribeGroqWhisper(wavPath)
		if err == nil && text != "" {
			return text, nil
		}
	}

	// Local whisper.cpp fallback if available
	return w.transcribeLocalWhisper(wavPath)
}

type groqWhisperResponse struct {
	Text  string `json:"text"`
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

// transcribeGroqWhisper calls Groq's high-speed Whisper Large v3 API.
func (w *WhisperSTT) transcribeGroqWhisper(wavPath string) (string, error) {
	file, err := os.Open(wavPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filepath.Base(wavPath))
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(part, file); err != nil {
		return "", err
	}

	_ = writer.WriteField("model", "whisper-large-v3")
	_ = writer.WriteField("language", "en")
	_ = writer.WriteField("response_format", "json")
	_ = writer.Close()

	req, err := http.NewRequest("POST", "https://api.groq.com/openai/v1/audio/transcriptions", body)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+w.GroqAPIKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var wResp groqWhisperResponse
	if err := json.Unmarshal(respBytes, &wResp); err != nil {
		return "", err
	}

	if wResp.Error.Message != "" {
		return "", fmt.Errorf("groq whisper error: %s", wResp.Error.Message)
	}

	return strings.TrimSpace(wResp.Text), nil
}

// transcribeLocalWhisper calls local whisper.cpp executable if installed.
func (w *WhisperSTT) transcribeLocalWhisper(wavPath string) (string, error) {
	whisperExe := "whisper.exe"
	if _, err := exec.LookPath(whisperExe); err != nil {
		// Look in local whisper directory
		localPath := filepath.Join("whisper", "whisper.exe")
		if _, err := os.Stat(localPath); err == nil {
			whisperExe = localPath
		} else {
			return "", fmt.Errorf("neither Groq Whisper API nor local whisper.exe is available")
		}
	}

	cmd := exec.Command(whisperExe, "-f", wavPath, "-nt")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("local whisper error: %w", err)
	}

	return strings.TrimSpace(string(out)), nil
}
