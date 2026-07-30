# Test Wake Word Engine with pause/resume support
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
Add-Type -AssemblyName System.Speech

$culture = New-Object System.Globalization.CultureInfo("en-US")
$engine = New-Object System.Speech.Recognition.SpeechRecognitionEngine($culture)
$engine.SetInputToDefaultAudioDevice()

# Constrained Grammar for Wake Words ONLY
$choices = New-Object System.Speech.Recognition.Choices
$choices.Add(@("Jarvis", "Hey Jarvis", "Friday", "Hey Friday", "Computer"))
$gb = New-Object System.Speech.Recognition.GrammarBuilder
$gb.Append($choices)
$grammar = New-Object System.Speech.Recognition.Grammar($gb)
$engine.LoadGrammar($grammar)

Write-Host "Engine Ready! Say 'Hey Friday' or 'Jarvis'..." -ForegroundColor Green

for ($i = 0; $i -lt 3; $i++) {
    $result = $engine.Recognize([TimeSpan]::FromSeconds(4))
    if ($result -and $result.Confidence -gt 0.3) {
        Write-Host "WAKE DETECTED: '$($result.Text)' (confidence: $([math]::Round($result.Confidence, 2)))" -ForegroundColor Green
        
        # Test pausing audio device so mic is free for recording
        Write-Host "Pausing engine (SetInputToNull)..." -ForegroundColor Yellow
        $engine.SetInputToNull()
        Start-Sleep -Seconds 2
        Write-Host "Resuming engine (SetInputToDefaultAudioDevice)..." -ForegroundColor Yellow
        $engine.SetInputToDefaultAudioDevice()
    } else {
        Write-Host "Silence or low confidence score" -ForegroundColor Gray
    }
}

$engine.Dispose()
Write-Host "Engine Stopped." -ForegroundColor Cyan
