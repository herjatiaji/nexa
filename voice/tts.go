package voice

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"

	"github.com/heraji/jarvis/voice/speech"
)

// TTS handles text-to-speech synthesis using Piper TTS (British Neural Voice) with SAPI5 fallback.
type TTS struct {
	VoiceName string // e.g. "piper" or model name
	Rate      int    // -10 to 10
	Enabled   bool
	mu        sync.Mutex
}

// NewTTS creates a new TTS engine instance.
func NewTTS(voiceName string, rate int, enabled bool) *TTS {
	if voiceName == "" {
		voiceName = "piper"
	}
	return &TTS{
		VoiceName: voiceName,
		Rate:      rate,
		Enabled:   enabled,
	}
}

// Speak speaks the given text synchronously.
func (t *TTS) Speak(text string) error {
	return t.SpeakWithEmotion(EmotionPayload{Text: text, Emotion: "neutral"})
}

// SpeakWithEmotion speaks text with natural British accent voice (ChatGPT-like conversational delivery).
func (t *TTS) SpeakWithEmotion(payload EmotionPayload) error {
	if !t.Enabled {
		return nil
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	cleanText := sanitizeTextForSpeech(payload.Text)
	if cleanText == "" {
		return nil
	}

	// 1. Try Piper British Neural Voice (en_GB-alba-medium)
	if err := t.speakPiper(cleanText); err == nil {
		return nil
	}

	// 2. Fallback to Windows System.Speech (SAPI5 British Female Voice)
	if runtime.GOOS == "windows" {
		return t.speakWindows(cleanText)
	}

	return nil
}

// SpeakAsync speaks the given text in a background goroutine.
func (t *TTS) SpeakAsync(text string) {
	if !t.Enabled {
		return
	}
	go func() {
		_ = t.Speak(text)
	}()
}

// SpeakWithProsody speaks text with prosody parameters controlling speed, pitch, and pauses.
// This is used by the Speech Intelligence Layer for emotion-aware delivery.
func (t *TTS) SpeakWithProsody(text string, prosody speech.ProsodyParams) error {
	if !t.Enabled {
		return nil
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	cleanText := sanitizeTextForSpeech(text)
	if cleanText == "" {
		return nil
	}

	// Try Piper with prosody-tuned parameters
	if err := t.speakPiperWithProsody(cleanText, prosody); err == nil {
		return nil
	}

	// Fallback to standard Windows SAPI5
	if runtime.GOOS == "windows" {
		return t.speakWindows(cleanText)
	}

	return nil
}

// SynthesizeToFile generates a WAV file from text with prosody params.
// Returns the path to the generated WAV file. Used by SpeechPlanner.Execute().
func (t *TTS) SynthesizeToFile(text string, prosody speech.ProsodyParams) (string, error) {
	if !t.Enabled {
		return "", fmt.Errorf("TTS is disabled")
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	cleanText := sanitizeTextForSpeech(text)
	if cleanText == "" {
		return "", fmt.Errorf("empty text after sanitization")
	}

	piperExe, modelFile := findPiperLocation()
	if piperExe == "" || modelFile == "" {
		return "", fmt.Errorf("piper engine or model not found")
	}

	tempWav := filepath.Join(os.TempDir(), fmt.Sprintf("nexa_chunk_%d.wav", os.Getpid()))

	// Convert prosody params to Piper CLI flags
	// Piper length_scale: >1.0 = slower, <1.0 = faster
	// Our SpeedScale: >1.0 = faster. So: length_scale = 1.0 / SpeedScale
	lengthScale := 1.0 / prosody.SpeedScale
	if lengthScale < 0.5 {
		lengthScale = 0.5
	}
	if lengthScale > 2.0 {
		lengthScale = 2.0
	}

	sentenceSilence := prosody.SentenceSilence
	if sentenceSilence < 0.05 {
		sentenceSilence = 0.05
	}

	cmd := exec.Command(piperExe, "--model", modelFile, "--output_file", tempWav,
		"--length_scale", fmt.Sprintf("%.2f", lengthScale),
		"--sentence_silence", fmt.Sprintf("%.2f", sentenceSilence))

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return "", err
	}

	if err := cmd.Start(); err != nil {
		return "", err
	}

	_, _ = io.WriteString(stdin, cleanText+"\n")
	stdin.Close()

	if err := cmd.Wait(); err != nil {
		return "", err
	}

	return tempWav, nil
}

// findPiperLocation returns the path to piper.exe and the model.onnx file if installed.
func findPiperLocation() (string, string) {
	exeCandidates := []string{
		filepath.Join("piper", "piper", "piper.exe"),
		filepath.Join("piper", "piper.exe"),
		"piper.exe",
	}

	piperExe := ""
	for _, c := range exeCandidates {
		if abs, err := filepath.Abs(c); err == nil {
			if _, err := os.Stat(abs); err == nil {
				piperExe = abs
				break
			}
		}
	}

	if piperExe == "" {
		if p, err := exec.LookPath("piper.exe"); err == nil {
			piperExe = p
		}
	}

	if piperExe == "" {
		return "", ""
	}

	piperDir := filepath.Dir(piperExe)
	modelFile := ""

	// Prioritize natural British female voice (en_GB-alba-medium.onnx)
	albaPath := filepath.Join(piperDir, "en_GB-alba-medium.onnx")
	if _, err := os.Stat(albaPath); err == nil {
		modelFile = albaPath
	} else {
		_ = filepath.Walk(piperDir, func(path string, info os.FileInfo, err error) error {
			if err == nil && !info.IsDir() && strings.HasSuffix(path, ".onnx") && !strings.Contains(path, "libtashkeel") {
				modelFile = path
				return filepath.SkipAll
			}
			return nil
		})
	}

	return piperExe, modelFile
}

// speakPiper uses Rhasspy Piper ONNX neural TTS engine with natural British conversational pacing.
func (t *TTS) speakPiper(text string) error {
	piperExe, modelFile := findPiperLocation()
	if piperExe == "" || modelFile == "" {
		return fmt.Errorf("piper engine or model not found")
	}

	tempWav := filepath.Join(os.TempDir(), fmt.Sprintf("nexa_tts_%d.wav", os.Getpid()))
	defer os.Remove(tempWav)

	// Run piper with natural conversational length_scale 1.0 and smooth sentence silence 0.2
	cmd := exec.Command(piperExe, "--model", modelFile, "--output_file", tempWav, "--length_scale", "1.0", "--sentence_silence", "0.2")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	_, _ = io.WriteString(stdin, text+"\n")
	stdin.Close()

	if err := cmd.Wait(); err != nil {
		return err
	}

	// Play clean high-fidelity WAV audio file
	if runtime.GOOS == "windows" {
		psPlay := fmt.Sprintf("(New-Object System.Media.SoundPlayer '%s').PlaySync()", strings.ReplaceAll(tempWav, "'", "''"))
		playCmd := exec.Command("powershell", "-NoProfile", "-Command", psPlay)
		return playCmd.Run()
	}

	return nil
}

// speakPiperWithProsody uses Piper with emotion-tuned prosody parameters for natural delivery.
func (t *TTS) speakPiperWithProsody(text string, prosody speech.ProsodyParams) error {
	piperExe, modelFile := findPiperLocation()
	if piperExe == "" || modelFile == "" {
		return fmt.Errorf("piper engine or model not found")
	}

	tempWav := filepath.Join(os.TempDir(), fmt.Sprintf("nexa_prosody_%d.wav", os.Getpid()))
	defer os.Remove(tempWav)

	// Convert SpeedScale to Piper length_scale (inverted relationship)
	lengthScale := 1.0 / prosody.SpeedScale
	if lengthScale < 0.5 {
		lengthScale = 0.5
	}
	if lengthScale > 2.0 {
		lengthScale = 2.0
	}

	sentenceSilence := prosody.SentenceSilence
	if sentenceSilence < 0.05 {
		sentenceSilence = 0.05
	}

	cmd := exec.Command(piperExe, "--model", modelFile, "--output_file", tempWav,
		"--length_scale", fmt.Sprintf("%.2f", lengthScale),
		"--sentence_silence", fmt.Sprintf("%.2f", sentenceSilence))

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	_, _ = io.WriteString(stdin, text+"\n")
	stdin.Close()

	if err := cmd.Wait(); err != nil {
		return err
	}

	return PlayWAVFile(tempWav)
}

// PlayWAVFile plays a WAV file synchronously using Windows System.Media.SoundPlayer.
// This is used by both direct TTS and the Speech Intelligence Layer pipeline.
func PlayWAVFile(wavPath string) error {
	if runtime.GOOS == "windows" {
		psPlay := fmt.Sprintf("(New-Object System.Media.SoundPlayer '%s').PlaySync()", strings.ReplaceAll(wavPath, "'", "''"))
		playCmd := exec.Command("powershell", "-NoProfile", "-Command", psPlay)
		return playCmd.Run()
	}
	return nil
}

// PlayJarvisSample plays a natural British female voice greeting sample.
func PlayJarvisSample() error {
	sampleText := "Hello, I am NEXA. How can I assist you with your tasks today?"
	tts := NewTTS("piper", 0, true)
	return tts.Speak(sampleText)
}

// speakWindows utilizes Windows System.Speech (SAPI5) fallback tuned for natural British female speech.
func (t *TTS) speakWindows(text string) error {
	escapedText := strings.ReplaceAll(text, "'", "''")
	escapedText = strings.ReplaceAll(escapedText, "\"", "`\"")

	psScript := fmt.Sprintf(`
Add-Type -AssemblyName System.Speech
$synth = New-Object System.Speech.Synthesis.SpeechSynthesizer
$synth.Rate = 0

$voices = $synth.GetInstalledVoices()
$britishVoice = $voices | Where-Object { 
    ($_.VoiceInfo.Culture.Name -eq 'en-GB' -and $_.VoiceInfo.Gender -eq 'Female') -or 
    $_.VoiceInfo.Name -like '*Hazel*' -or 
    $_.VoiceInfo.Name -like '*Sonia*' -or 
    $_.VoiceInfo.Name -like '*Susan*'
} | Select-Object -First 1

if (-not $britishVoice) {
    $britishVoice = $voices | Where-Object { $_.VoiceInfo.Gender -eq 'Female' -or $_.VoiceInfo.Name -like '*Zira*' } | Select-Object -First 1
}

if ($britishVoice) {
    $synth.SelectVoice($britishVoice.VoiceInfo.Name)
}

$synth.Speak('%s')
`, escapedText)

	cmd := exec.Command("powershell", "-NoProfile", "-Command", psScript)
	return cmd.Run()
}

// ListVoices returns available TTS voices installed on the system.
func ListVoices() ([]string, error) {
	piperExe, modelFile := findPiperLocation()
	var voices []string
	if piperExe != "" && modelFile != "" {
		voices = append(voices, fmt.Sprintf("Piper British Neural Voice (%s)", filepath.Base(modelFile)))
	}

	if runtime.GOOS == "windows" {
		psScript := `
Add-Type -AssemblyName System.Speech
$synth = New-Object System.Speech.Synthesis.SpeechSynthesizer
$synth.GetInstalledVoices() | ForEach-Object { $_.VoiceInfo.Name + " (" + $_.VoiceInfo.Culture.Name + " - " + $_.VoiceInfo.Gender + ")" }
`
		cmd := exec.Command("powershell", "-NoProfile", "-Command", psScript)
		if out, err := cmd.Output(); err == nil {
			lines := strings.Split(strings.TrimSpace(string(out)), "\r\n")
			for _, l := range lines {
				if strings.TrimSpace(l) != "" {
					voices = append(voices, strings.TrimSpace(l))
				}
			}
		}
	}
	return voices, nil
}

var (
	codeBlockRegex  = regexp.MustCompile("(?s)```.*?```")
	inlineCodeRegex = regexp.MustCompile("`.*?`")
	urlRegex        = regexp.MustCompile(`https?://\S+`)
	symbolRegex     = regexp.MustCompile(`[\*#_~>|]`)
	cjkRegex        = regexp.MustCompile(`[\x{3000}-\x{303f}\x{3040}-\x{309f}\x{30a0}-\x{30ff}\x{ff00}-\x{ffef}\x{4e00}-\x{9faf}]`)
)

// sanitizeTextForSpeech removes markdown formatting, symbols, and non-ASCII CJK for smooth natural reading.
func sanitizeTextForSpeech(text string) string {
	text = codeBlockRegex.ReplaceAllString(text, " [code block omitted] ")
	text = inlineCodeRegex.ReplaceAllString(text, " ")
	text = urlRegex.ReplaceAllString(text, " ")
	text = symbolRegex.ReplaceAllString(text, " ")
	text = cjkRegex.ReplaceAllString(text, " ")
	lines := strings.Split(text, "\n")
	var cleanLines []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			cleanLines = append(cleanLines, trimmed)
		}
	}
	return strings.Join(cleanLines, ". ")
}
