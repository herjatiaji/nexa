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

// TTS handles text-to-speech synthesis using Piper TTS (rhasspy/piper) with SAPI5 fallback.
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
	// Search candidates for piper.exe
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
		// Check PATH
		if p, err := exec.LookPath("piper.exe"); err == nil {
			piperExe = p
		}
	}

	if piperExe == "" {
		return "", ""
	}

	// Search for ONNX model in same directory or subdirectories
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

// speakPiper uses Rhasspy Piper ONNX neural TTS engine.
func (t *TTS) speakPiper(text string) error {
	piperExe, modelFile := findPiperLocation()
	if piperExe == "" || modelFile == "" {
		return fmt.Errorf("piper engine or model not found")
	}

	// Create temp wav file
	tempWav := filepath.Join(os.TempDir(), fmt.Sprintf("friday_tts_%d.wav", os.Getpid()))
	defer os.Remove(tempWav)

	// Run piper: echo "text" | piper.exe --model model.onnx --output_file temp.wav
	cmd := exec.Command(piperExe, "--model", modelFile, "--output_file", tempWav)
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

	// Play generated WAV audio file
	if runtime.GOOS == "windows" {
		psPlay := fmt.Sprintf("(New-Object System.Media.SoundPlayer '%s').PlaySync()", strings.ReplaceAll(tempWav, "'", "''"))
		playCmd := exec.Command("powershell", "-NoProfile", "-Command", psPlay)
		return playCmd.Run()
	}

	return nil
}

// speakWindows utilizes Windows System.Speech (SAPI5) fallback for speech generation.
func (t *TTS) speakWindows(text string) error {
	escapedText := strings.ReplaceAll(text, "'", "''")
	escapedText = strings.ReplaceAll(escapedText, "\"", "`\"")

	psScript := fmt.Sprintf(`
Add-Type -AssemblyName System.Speech
$synth = New-Object System.Speech.Synthesis.SpeechSynthesizer
$synth.Rate = %d

$voices = $synth.GetInstalledVoices()
# Priority 1: British Female Voice (Hazel, Sonia, en-GB Female)
$fridayVoice = $voices | Where-Object { 
    ($_.VoiceInfo.Culture.Name -eq 'en-GB' -and $_.VoiceInfo.Gender -eq 'Female') -or 
    $_.VoiceInfo.Name -like '*Hazel*' -or 
    $_.VoiceInfo.Name -like '*Sonia*' -or 
    $_.VoiceInfo.Name -like '*Susan*'
} | Select-Object -First 1

# Priority 2: Any Female Voice (e.g. Zira)
if (-not $fridayVoice) {
    $fridayVoice = $voices | Where-Object { $_.VoiceInfo.Gender -eq 'Female' -or $_.VoiceInfo.Name -like '*Zira*' } | Select-Object -First 1
}

if ($fridayVoice) {
    $synth.SelectVoice($fridayVoice.VoiceInfo.Name)
}

$synth.Speak('%s')
`, t.Rate, escapedText)

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

// PlayJarvisSample plays an authentic FRIDAY British female voice greeting sample using Piper TTS.
func PlayJarvisSample() error {
	sampleText := "Hello boss, I'm Friday. Piper neural text-to-speech is online and ready for your command."
	tts := NewTTS("piper", 0, true)
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
	// Remove code blocks
	text = codeBlockRegex.ReplaceAllString(text, " [code block omitted] ")
	// Remove inline code
	text = inlineCodeRegex.ReplaceAllString(text, " ")
	// Remove URLs
	text = urlRegex.ReplaceAllString(text, " ")
	// Remove markdown symbols (*, #, _, ~, >, |)
	text = symbolRegex.ReplaceAllString(text, " ")
	// Trim extra spaces and newlines
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
