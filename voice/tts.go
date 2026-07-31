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

// PlayJarvisSample plays a natural British female voice greeting sample.
func PlayJarvisSample() error {
	sampleText := "Hello, I am NEXA. How can I assist you with your tasks today?"
	tts := NewTTS("piper", 0, true)
	return tts.Speak(sampleText)
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
