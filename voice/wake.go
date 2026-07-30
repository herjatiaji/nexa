package voice

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// WakeWordListener handles continuous background listening for the wake word.
type WakeWordListener struct {
	WakePhrase string
	OnWake     func()
}

// NewWakeWordListener creates a new wake word listener.
func NewWakeWordListener(phrase string, onWake func()) *WakeWordListener {
	if phrase == "" {
		phrase = "Friday"
	}
	return &WakeWordListener{
		WakePhrase: phrase,
		OnWake:     onWake,
	}
}

// CheckMicrophone verifies if Speech Engine / Python openWakeWord can access mic.
func CheckMicrophone() (bool, string) {
	if runtime.GOOS != "windows" {
		return false, "Microphone check only supported on Windows OS"
	}

	pyPath, err := exec.LookPath("python")
	if err == nil && pyPath != "" {
		return true, "openWakeWord Engine Ready (Python Neural ONNX Stream)"
	}

	psScript := `
Add-Type -AssemblyName System.Speech
try {
    $e = New-Object System.Speech.Recognition.SpeechRecognitionEngine
    $e.SetInputToDefaultAudioDevice()
    Write-Output "OK"
    $e.Dispose()
} catch {
    Write-Output "ERROR:$($_.Exception.Message)"
}
`
	cmd := exec.Command("powershell", "-NoProfile", "-Command", psScript)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Sprintf("Failed to check speech engine: %v", err)
	}

	result := strings.TrimSpace(string(out))
	if strings.HasPrefix(result, "OK") {
		return true, "Headless Speech Engine Ready (Background Mode - No UI Popup)"
	}

	return false, fmt.Sprintf("Speech Engine status: %s", result)
}

type persistentEngine struct {
	stdin      io.WriteCloser
	resultChan <-chan string
	mu         sync.Mutex
}

var globalEngine *persistentEngine

// StartPersistentListener starts a single persistent process for Stage 1 Wake Word Detection.
// Prefers openWakeWord (Python neural ONNX stream on 80ms audio frames) for ultra-accurate detection,
// with SAPI5 Constrained Grammar fallback.
func StartPersistentListener() (<-chan string, func(), error) {
	if runtime.GOOS != "windows" {
		return nil, nil, fmt.Errorf("voice listening is only supported on Windows")
	}

	// 1. Try launching openWakeWord Python engine first
	scriptPath := filepath.Join("voice", "openwakeword_listener.py")
	if _, err := os.Stat(scriptPath); err == nil {
		if pyExe, err := exec.LookPath("python"); err == nil && pyExe != "" {
			cmd := exec.Command(pyExe, scriptPath)
			ch, cleanup, err := startEngineProcess(cmd, "OPENWAKEWORD_READY")
			if err == nil {
				return ch, cleanup, nil
			}
		}
	}

	// 2. Fallback: PowerShell SAPI5 Constrained Engine
	psScript := `
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
Add-Type -AssemblyName System.Speech

$culture = New-Object System.Globalization.CultureInfo("en-US")
$engine = New-Object System.Speech.Recognition.SpeechRecognitionEngine($culture)
$engine.SetInputToDefaultAudioDevice()

$choices = New-Object System.Speech.Recognition.Choices
$choices.Add(@("Jarvis", "Hey Jarvis", "Hello Jarvis", "Friday", "Hey Friday", "Hello Friday", "Computer"))
$gb = New-Object System.Speech.Recognition.GrammarBuilder
$gb.Append($choices)
$grammar = New-Object System.Speech.Recognition.Grammar($gb)
$grammar.Name = "wake"
$engine.LoadGrammar($grammar)

Write-Output "ENGINE_READY"
[Console]::Out.Flush()

while ($true) {
    $line = [Console]::In.ReadLine()
    if ($line -eq "QUIT") { break }

    if ($line -eq "PAUSE") {
        try { $engine.SetInputToNull() } catch {}
        Write-Output "PAUSED"
    }
    elseif ($line -eq "RESUME") {
        try { $engine.SetInputToDefaultAudioDevice() } catch {}
        Write-Output "RESUMED"
    }
    elseif ($line -eq "LISTEN_WAKE") {
        try {
            $result = $engine.Recognize([TimeSpan]::FromSeconds(4))
            if ($result -and $result.Confidence -gt 0.25) {
                Write-Output "WAKE:$($result.Text):$([math]::Round($result.Confidence, 2))"
            } else {
                Write-Output "SILENCE"
            }
        } catch {
            Write-Output "SILENCE"
        }
    }
    [Console]::Out.Flush()
}

$engine.Dispose()
Write-Output "ENGINE_STOPPED"
`
	cmd := exec.Command("powershell", "-NoProfile", "-NoLogo", "-Command", psScript)
	return startEngineProcess(cmd, "ENGINE_READY")
}

func startEngineProcess(cmd *exec.Cmd, readySignal string) (<-chan string, func(), error) {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	cmd.Stderr = nil

	if err := cmd.Start(); err != nil {
		return nil, nil, fmt.Errorf("failed to start process: %w", err)
	}

	resultChan := make(chan string, 20)
	scanner := bufio.NewScanner(stdout)

	ready := make(chan bool, 1)
	go func() {
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == readySignal {
				ready <- true
				break
			}
		}
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" {
				resultChan <- line
			}
		}
		close(resultChan)
	}()

	select {
	case <-ready:
	case <-time.After(15 * time.Second):
		cmd.Process.Kill()
		return nil, nil, fmt.Errorf("engine startup timed out")
	}

	cleanup := func() {
		io.WriteString(stdin, "QUIT\n")
		stdin.Close()
		time.AfterFunc(3*time.Second, func() {
			if cmd.Process != nil {
				cmd.Process.Kill()
			}
		})
		cmd.Wait()
	}

	globalEngine = &persistentEngine{
		stdin:      stdin,
		resultChan: resultChan,
	}

	return resultChan, cleanup, nil
}

// SendListenCommand sends a command ("LISTEN_WAKE", "PAUSE", "RESUME", "QUIT") to the Stage 1 engine.
func SendListenCommand(command string) error {
	if globalEngine == nil {
		return fmt.Errorf("speech engine not running")
	}
	globalEngine.mu.Lock()
	defer globalEngine.mu.Unlock()

	_, err := io.WriteString(globalEngine.stdin, command+"\n")
	return err
}

// ListenOnceForWakeWord is a legacy fallback helper.
func (w *WakeWordListener) ListenOnceForWakeWord() (bool, error) {
	if runtime.GOOS != "windows" {
		return false, fmt.Errorf("wake word detection is currently supported on Windows")
	}

	psScript := `
Add-Type -AssemblyName System.Speech
try {
    $engine = New-Object System.Speech.Recognition.SpeechRecognitionEngine
    $engine.SetInputToDefaultAudioDevice()

    $Choices = New-Object System.Speech.Recognition.Choices
    $Choices.Add(@("Jarvis", "Hey Jarvis", "Friday", "Hey Friday", "Computer"))
    $gb = New-Object System.Speech.Recognition.GrammarBuilder
    $gb.Append($Choices)
    $grammar = New-Object System.Speech.Recognition.Grammar($gb)

    $engine.LoadGrammar($grammar)
    $result = $engine.Recognize([TimeSpan]::FromSeconds(4))

    if ($result -and $result.Confidence -gt 0.25) {
        Write-Output "WAKE:$($result.Text)"
    }
    $engine.Dispose()
} catch {
    Write-Output "MIC_ERROR:$($_.Exception.Message)"
}
`

	cmd := exec.Command("powershell", "-NoProfile", "-Command", psScript)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return false, err
	}

	result := strings.TrimSpace(string(out))
	if strings.HasPrefix(result, "MIC_ERROR:") {
		return false, fmt.Errorf("speech engine error: %s", strings.TrimPrefix(result, "MIC_ERROR:"))
	}

	if strings.HasPrefix(result, "WAKE:") {
		return true, nil
	}

	return false, nil
}
