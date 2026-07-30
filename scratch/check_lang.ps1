Add-Type -AssemblyName System.Speech
Write-Host "=== Installed Speech Recognizers ===" -ForegroundColor Cyan
$recognizers = [System.Speech.Recognition.SpeechRecognitionEngine]::InstalledRecognizers()
foreach ($r in $recognizers) {
    Write-Host "  ID: $($r.Id)" -ForegroundColor White
    Write-Host "  Culture: $($r.Culture)" -ForegroundColor Yellow
    Write-Host "  Description: $($r.Description)" -ForegroundColor White
    Write-Host ""
}
Write-Host "Total: $($recognizers.Count) recognizer(s)" -ForegroundColor Cyan

Write-Host ""
Write-Host "=== Current Windows Display Language ===" -ForegroundColor Cyan
$lang = (Get-WinSystemLocale).Name
Write-Host "  System Locale: $lang" -ForegroundColor Yellow
$uilang = (Get-WinUserLanguageList) | ForEach-Object { $_.LanguageTag }
Write-Host "  User Languages: $($uilang -join ', ')" -ForegroundColor Yellow
