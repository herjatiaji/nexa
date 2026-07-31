package voice

import (
	"encoding/binary"
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

// TTS handles text-to-speech synthesis using Piper TTS with cute anime companion pitch shift.
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

	// 1. Try Piper TTS first if available (with cute anime pitch shift)
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

// pitchShiftWAV shifts the sample rate of a WAV file to create a cute anime character voice pitch.
func pitchShiftWAV(wavPath string, multiplier float64) error {
	data, err := os.ReadFile(wavPath)
	if err != nil || len(data) < 44 {
		return err
	}

	if string(data[0:4]) != "RIFF" || string(data[8:12]) != "WAVE" {
		return fmt.Errorf("invalid WAV file header")
	}

	origSampleRate := binary.LittleEndian.Uint32(data[24:28])
	origByteRate := binary.LittleEndian.Uint32(data[28:32])

	newSampleRate := uint32(float64(origSampleRate) * multiplier)
	newByteRate := uint32(float64(origByteRate) * multiplier)

	binary.LittleEndian.PutUint32(data[24:28], newSampleRate)
	binary.LittleEndian.PutUint32(data[28:32], newByteRate)

	return os.WriteFile(wavPath, data, 0644)
}

// speakPiper uses Rhasspy Piper ONNX neural TTS engine + cute anime character pitch modulation.
func (t *TTS) speakPiper(text string) error {
	piperExe, modelFile := findPiperLocation()
	if piperExe == "" || modelFile == "" {
		return fmt.Errorf("piper engine or model not found")
	}

	tempWav := filepath.Join(os.TempDir(), fmt.Sprintf("nexa_cute_tts_%d.wav", os.Getpid()))
	defer os.Remove(tempWav)

	// Run piper with length_scale 0.82 for cute lively speech pacing
	cmd := exec.Command(piperExe, "--model", modelFile, "--output_file", tempWav, "--length_scale", "0.82")
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

	// Apply 1.34x pitch shift for cute anime companion character voice
	_ = pitchShiftWAV(tempWav, 1.34)

	// Play audio
	if runtime.GOOS == "windows" {
		psPlay := fmt.Sprintf("(New-Object System.Media.SoundPlayer '%s').PlaySync()", strings.ReplaceAll(tempWav, "'", "''"))
		playCmd := exec.Command("powershell", "-NoProfile", "-Command", psPlay)
		return playCmd.Run()
	}

	return nil
}

// speakWindows utilizes Windows System.Speech (SAPI5) fallback tuned for cute character pitch.
func (t *TTS) speakWindows(text string) error {
	escapedText := strings.ReplaceAll(text, "'", "''")
	escapedText = strings.ReplaceAll(escapedText, "\"", "`\"")

	psScript := fmt.Sprintf(`
Add-Type -AssemblyName System.Speech
$synth = New-Object System.Speech.Synthesis.SpeechSynthesizer
$synth.Rate = 2

$voices = $synth.GetInstalledVoices()
$cuteVoice = $voices | Where-Object { 
    $_.VoiceInfo.Gender -eq 'Female' -or 
    $_.VoiceInfo.Name -like '*Zira*' -or 
    $_.VoiceInfo.Name -like '*Hazel*' -or 
    $_.VoiceInfo.Name -like '*Sonia*'
} | Select-Object -First 1

if ($cuteVoice) {
    $synth.SelectVoice($cuteVoice.VoiceInfo.Name)
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
		voices = append(voices, fmt.Sprintf("Cute Anime Companion TTS (%s)", filepath.Base(modelFile)))
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

// PlayJarvisSample plays a cute anime companion voice greeting sample.
func PlayJarvisSample() error {
	sampleText := "Hello boss! Nexa cute anime companion voice is online and ready!"
	tts := NewTTS("piper", 2, true)
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
