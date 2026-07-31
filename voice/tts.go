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

// TTS handles text-to-speech synthesis using Piper TTS with SAPI5 fallback.
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
	if !t.Enabled {
		return nil
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	cleanText := sanitizeTextForSpeech(text)
	if cleanText == "" {
		return nil
	}

	// 1. Try Piper TTS first if available
	if err := t.speakPiper(cleanText); err == nil {
		return nil
	}

	// 2. Fallback to Windows System.Speech (SAPI5)
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
	_ = filepath.Walk(piperDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && strings.HasSuffix(path, ".onnx") && !strings.Contains(path, "libtashkeel") {
			modelFile = path
			return filepath.SkipAll
		}
		return nil
	})

	return piperExe, modelFile
}

// speakPiper uses Rhasspy Piper ONNX neural TTS engine with clean cheerful pacing.
func (t *TTS) speakPiper(text string) error {
	piperExe, modelFile := findPiperLocation()
	if piperExe == "" || modelFile == "" {
		return fmt.Errorf("piper engine or model not found")
	}

	tempWav := filepath.Join(os.TempDir(), fmt.Sprintf("nexa_tts_%d.wav", os.Getpid()))
	defer os.Remove(tempWav)

	// Run piper with length_scale 0.92 for natural, cheerful pace
	cmd := exec.Command(piperExe, "--model", modelFile, "--output_file", tempWav, "--length_scale", "0.92")
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

	// Play clean WAV audio file
	if runtime.GOOS == "windows" {
		psPlay := fmt.Sprintf("(New-Object System.Media.SoundPlayer '%s').PlaySync()", strings.ReplaceAll(tempWav, "'", "''"))
		playCmd := exec.Command("powershell", "-NoProfile", "-Command", psPlay)
		return playCmd.Run()
	}

	return nil
}

// speakWindows utilizes Windows System.Speech (SAPI5) fallback tuned for a cheerful female voice.
func (t *TTS) speakWindows(text string) error {
	escapedText := strings.ReplaceAll(text, "'", "''")
	escapedText = strings.ReplaceAll(escapedText, "\"", "`\"")

	psScript := fmt.Sprintf(`
Add-Type -AssemblyName System.Speech
$synth = New-Object System.Speech.Synthesis.SpeechSynthesizer
$synth.Rate = 1

$voices = $synth.GetInstalledVoices()
$cheerfulVoice = $voices | Where-Object { 
    $_.VoiceInfo.Name -like '*Zira*' -or 
    $_.VoiceInfo.Name -like '*Hazel*' -or 
    $_.VoiceInfo.Name -like '*Sonia*' -or
    $_.VoiceInfo.Name -like '*Susan*' -or
    ($_.VoiceInfo.Gender -eq 'Female')
} | Select-Object -First 1

if ($cheerfulVoice) {
    $synth.SelectVoice($cheerfulVoice.VoiceInfo.Name)
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
		voices = append(voices, fmt.Sprintf("Piper Neural TTS (%s)", filepath.Base(modelFile)))
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

// PlayJarvisSample plays a cheerful female voice greeting sample using Piper TTS.
func PlayJarvisSample() error {
	sampleText := "Hello boss! Nexa cheerful voice output is online and ready for your command."
	tts := NewTTS("piper", 1, true)
	return tts.Speak(sampleText)
}

var (
	codeBlockRegex  = regexp.MustCompile("(?s)```.*?```")
	inlineCodeRegex = regexp.MustCompile("`.*?`")
	urlRegex        = regexp.MustCompile(`https?://\S+`)
	symbolRegex     = regexp.MustCompile(`[\*#_~>|]`)
)

// sanitizeTextForSpeech removes markdown formatting and code blocks for smooth TTS reading.
func sanitizeTextForSpeech(text string) string {
	text = codeBlockRegex.ReplaceAllString(text, " [code block omitted] ")
	text = inlineCodeRegex.ReplaceAllString(text, " ")
	text = urlRegex.ReplaceAllString(text, " ")
	text = symbolRegex.ReplaceAllString(text, " ")
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
