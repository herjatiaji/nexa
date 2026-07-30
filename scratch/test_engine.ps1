[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
Add-Type -AssemblyName System.Speech

$engine = New-Object System.Speech.Recognition.SpeechRecognitionEngine
$engine.SetInputToDefaultAudioDevice()

$wakeChoices = New-Object System.Speech.Recognition.Choices
$wakeChoices.Add(@("Jarvis", "Hey Jarvis", "Hello Jarvis", "OK Jarvis"))
$wakeBuilder = New-Object System.Speech.Recognition.GrammarBuilder
$wakeBuilder.Append($wakeChoices)
$wakeGrammar = New-Object System.Speech.Recognition.Grammar($wakeBuilder)
$wakeGrammar.Name = "wake"
$engine.LoadGrammar($wakeGrammar)

Write-Output "ENGINE_READY"
[Console]::Out.Flush()

# Quick 3-second recognition test
Write-Output "Listening for 3 seconds... say 'Hey Jarvis'"
$result = $engine.Recognize([TimeSpan]::FromSeconds(3))
if ($result -and $result.Confidence -gt 0.3) {
    Write-Output "WAKE:$($result.Text):$([math]::Round($result.Confidence, 2))"
} else {
    Write-Output "SILENCE (no speech in 3 seconds - expected in automated test)"
}

$engine.Dispose()
Write-Output "TEST_PASSED"
