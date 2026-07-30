Add-Type -AssemblyName System.Speech

try {
    $culture = New-Object System.Globalization.CultureInfo("en-US")
    $engine = New-Object System.Speech.Recognition.SpeechRecognitionEngine($culture)
    $engine.SetInputToDefaultAudioDevice()

    $choices = New-Object System.Speech.Recognition.Choices
    $choices.Add(@("Nexa", "Hey Nexa", "Hello Nexa", "Hi Nexa", "Yo Nexa", "Next", "Hey Next", "Alexa", "Hey Alexa", "Friday", "Hey Friday"))
    $gb = New-Object System.Speech.Recognition.GrammarBuilder
    $gb.Append($choices)
    $grammar = New-Object System.Speech.Recognition.Grammar($gb)
    $engine.LoadGrammar($grammar)

    Write-Host "Listening continuously for 8 seconds... Say 'Hey Nexa' or 'Alexa' now!"
    
    $startTime = [DateTime]::Now
    while (([DateTime]::Now - $startTime).TotalSeconds -lt 8) {
        $result = $engine.Recognize([TimeSpan]::FromSeconds(2))
        if ($result) {
            Write-Host "MATCH FOUND: $($result.Text) (Confidence: $($result.Confidence))"
        } else {
            Write-Host "No speech detected in this 2-second window."
        }
    }
    
    $engine.Dispose()
} catch {
    Write-Host "ERROR: $($_.Exception.Message)"
}
