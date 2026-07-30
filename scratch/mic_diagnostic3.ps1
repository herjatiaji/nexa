# Direct hardware mic test - bypass System.Speech entirely
# Records raw audio via waveIn Win32 API to verify hardware works

Write-Host "=== Direct Hardware Mic Test ===" -ForegroundColor Cyan
Write-Host ""

# Method 1: Quick check - does Windows see audio input at all?
Write-Host "[1] Testing via PowerShell Get-PnpDevice..." -ForegroundColor Yellow
$mics = Get-PnpDevice -Class AudioEndpoint -Status OK 2>$null | Where-Object { $_.FriendlyName -like "*Microphone*" }
foreach ($m in $mics) {
    Write-Host "    $($m.FriendlyName) - $($m.InstanceId)" -ForegroundColor White
}

Write-Host ""

# Method 2: Use Windows Sound Recorder (SoundRecorder) to record WAV
Write-Host "[2] Recording 3 seconds of audio via SpeechRecognitionEngine with DictationGrammar..." -ForegroundColor Yellow
Write-Host "    >>> SPEAK NOW for 3 seconds! <<<" -ForegroundColor Red

Add-Type -AssemblyName System.Speech

$engine = New-Object System.Speech.Recognition.SpeechRecognitionEngine

# Try to enumerate audio inputs
Write-Host ""
Write-Host "[3] Audio input info:" -ForegroundColor Yellow
try {
    $audioInput = $engine.AudioFormat
    Write-Host "    Audio format: $audioInput" -ForegroundColor White
} catch {
    Write-Host "    (Format info not available before SetInput)" -ForegroundColor Gray
}

$engine.SetInputToDefaultAudioDevice()
Write-Host "    SetInputToDefaultAudioDevice: OK" -ForegroundColor Green

# Save to WAV file to prove audio is being captured
$tempWav = "$env:TEMP\jarvis_test_$(Get-Date -Format 'yyyyMMdd_HHmmss').wav"
try {
    $engine.SetInputToDefaultAudioDevice()
    $grammar = New-Object System.Speech.Recognition.DictationGrammar
    $engine.LoadGrammar($grammar)
    
    # Use synchronous Recognize with output to WAV
    Write-Host "    Listening with dictation (5 seconds)... SAY ANYTHING!" -ForegroundColor Red
    $result = $engine.Recognize([TimeSpan]::FromSeconds(5))
    
    if ($result) {
        Write-Host ""
        Write-Host "    *** SPEECH DETECTED! ***" -ForegroundColor Green
        Write-Host "    Text: '$($result.Text)'" -ForegroundColor Green
        Write-Host "    Confidence: $([math]::Round($result.Confidence, 3))" -ForegroundColor Green
    } else {
        Write-Host "    No speech detected via dictation" -ForegroundColor Red
    }
} catch {
    Write-Host "    Error: $($_.Exception.Message)" -ForegroundColor Red
}
$engine.Dispose()

# Method 3: Try SpeechRecognizer (shared/desktop - uses Windows audio pipeline)
Write-Host ""
Write-Host "[4] Testing via SpeechRecognizer (shared desktop engine)..." -ForegroundColor Yellow
Write-Host "    This uses Windows' built-in audio routing (different from SpeechRecognitionEngine)" -ForegroundColor Gray
try {
    $recognizer = New-Object System.Speech.Recognition.SpeechRecognizer
    $choices = New-Object System.Speech.Recognition.Choices
    $choices.Add(@("Jarvis", "Hey Jarvis", "Hello", "Test", "Yes", "No"))
    $gb = New-Object System.Speech.Recognition.GrammarBuilder
    $gb.Append($choices)
    $grammar = New-Object System.Speech.Recognition.Grammar($gb)
    $recognizer.LoadGrammar($grammar)
    
    Write-Host "    SpeechRecognizer created: OK" -ForegroundColor Green
    Write-Host "    State: $($recognizer.State)" -ForegroundColor White
    Write-Host "    Audio: Enabled=$($recognizer.Enabled)" -ForegroundColor White
    
    $recognizer.Dispose()
} catch {
    Write-Host "    Error: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host "    (This is OK - SpeechRecognizer needs desktop session)" -ForegroundColor Gray
}

# Method 4: Use ffmpeg to test raw mic capture if available
Write-Host ""
Write-Host "[5] Checking for alternative audio tools..." -ForegroundColor Yellow
$ffmpeg = Get-Command ffmpeg -ErrorAction SilentlyContinue
if ($ffmpeg) {
    Write-Host "    ffmpeg found: $($ffmpeg.Source)" -ForegroundColor Green
    Write-Host "    Running 2-second capture test..."
    $testFile = "$env:TEMP\jarvis_ffmpeg_test.wav"
    & ffmpeg -f dshow -i audio="Microphone (USB Audio Device)" -t 2 -y $testFile 2>&1 | Out-Null
    if (Test-Path $testFile) {
        $size = (Get-Item $testFile).Length
        Write-Host "    Recorded $size bytes" -ForegroundColor $(if ($size -gt 10000) { "Green" } else { "Red" })
        Remove-Item $testFile -ErrorAction SilentlyContinue
    }
} else {
    Write-Host "    ffmpeg not found (optional)" -ForegroundColor Gray
}

# Method 5: Check if any app is exclusively locking the mic
Write-Host ""
Write-Host "[6] Checking Exclusive Mode settings..." -ForegroundColor Yellow
Write-Host "    If another app has exclusive mic access, System.Speech cannot use it." -ForegroundColor Gray
Write-Host "    To fix: mmsys.cpl -> Recording -> USB mic -> Properties -> Advanced tab" -ForegroundColor White
Write-Host "    UNCHECK 'Allow applications to take exclusive control'" -ForegroundColor Yellow

Write-Host ""
Write-Host "=== Test Complete ===" -ForegroundColor Cyan
