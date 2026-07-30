Add-Type -AssemblyName System.Speech
$synth = New-Object System.Speech.Synthesis.SpeechSynthesizer
$voices = $synth.GetInstalledVoices()
foreach ($v in $voices) {
    Write-Host "Name: $($v.VoiceInfo.Name), Culture: $($v.VoiceInfo.Culture), Gender: $($v.VoiceInfo.Gender)"
}
