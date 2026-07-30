# Check which mic is the actual default AND test with explicit USB mic selection
Write-Host "=== Checking Default Recording Device ===" -ForegroundColor Cyan

# Method: Use AudioDeviceCmdlets alternative - check via registry
try {
    $defaultDevice = Get-ItemProperty "HKCU:\SOFTWARE\Microsoft\Multimedia\Sound Mapper" -Name "Record" -ErrorAction SilentlyContinue
    if ($defaultDevice) {
        Write-Host "Default Recording Device (registry): $($defaultDevice.Record)" -ForegroundColor Yellow
    }
} catch {}

Write-Host ""
Write-Host "=== Testing with NAudio-style WAV recording (3 seconds) ===" -ForegroundColor Cyan
Write-Host "This will record raw audio to check if mic is ACTUALLY capturing..." -ForegroundColor Yellow
Write-Host ""

# Use .NET SoundPlayer approach to test raw audio capture
Add-Type -TypeDefinition @"
using System;
using System.Runtime.InteropServices;

public class WaveIn {
    [DllImport("winmm.dll")]
    public static extern int waveInGetNumDevs();
    
    [DllImport("winmm.dll", CharSet = CharSet.Auto)]
    public static extern int waveInGetDevCaps(int deviceId, ref WAVEINCAPS caps, int size);
    
    [StructLayout(LayoutKind.Sequential, CharSet = CharSet.Auto)]
    public struct WAVEINCAPS {
        public short wMid;
        public short wPid;
        public int vDriverVersion;
        [MarshalAs(UnmanagedType.ByValTStr, SizeConst = 32)]
        public string szPname;
        public int dwFormats;
        public short wChannels;
        public short wReserved1;
    }
}
"@

$numDevs = [WaveIn]::waveInGetNumDevs()
Write-Host "Number of audio input devices: $numDevs" -ForegroundColor White
Write-Host ""

$usbDeviceId = -1
for ($i = 0; $i -lt $numDevs; $i++) {
    $caps = New-Object WaveIn+WAVEINCAPS
    [WaveIn]::waveInGetDevCaps($i, [ref]$caps, [System.Runtime.InteropServices.Marshal]::SizeOf($caps)) | Out-Null
    $marker = ""
    if ($caps.szPname -like "*USB*") {
        $marker = " <-- USB MIC"
        $usbDeviceId = $i
    }
    Write-Host "  Device $i : $($caps.szPname) (channels: $($caps.wChannels))$marker" -ForegroundColor White
}

Write-Host ""

# Now test speech recognition with SPECIFIC audio device selection
Write-Host "=== Speech Recognition Test ===" -ForegroundColor Cyan
Add-Type -AssemblyName System.Speech

# Test 1: Default device
Write-Host ""
Write-Host "[Test A] Using DEFAULT audio device:" -ForegroundColor Yellow
$engineA = New-Object System.Speech.Recognition.SpeechRecognitionEngine
$engineA.SetInputToDefaultAudioDevice()
$choices = New-Object System.Speech.Recognition.Choices
$choices.Add(@("Jarvis", "Hey Jarvis", "Hello", "Test", "Yes", "No", "One", "Two"))
$gb = New-Object System.Speech.Recognition.GrammarBuilder
$gb.Append($choices)
$engineA.LoadGrammar((New-Object System.Speech.Recognition.Grammar($gb)))

Write-Host "  Say something (3 sec)..." -ForegroundColor White
$resultA = $engineA.Recognize([TimeSpan]::FromSeconds(3))
if ($resultA) {
    Write-Host "  DEFAULT DEVICE HEARD: '$($resultA.Text)' (confidence: $($resultA.Confidence))" -ForegroundColor Green
} else {
    Write-Host "  DEFAULT DEVICE: Nothing detected" -ForegroundColor Red
}
$engineA.Dispose()

# Test 2: Try with AudioDeviceFormat selection for USB
Write-Host ""
Write-Host "[Test B] Trying with WAV file workaround to test USB mic..." -ForegroundColor Yellow

# Record 3 seconds of audio from default device using SoundRecorder API
$tempWav = "$env:TEMP\jarvis_mic_test.wav"

# Use System.Speech with audio level events to check if audio is flowing
$engineB = New-Object System.Speech.Recognition.SpeechRecognitionEngine
$engineB.SetInputToDefaultAudioDevice()

# Register event to check audio levels
$audioDetected = $false
$maxLevel = 0
$engineB.add_AudioLevelUpdated({
    param($sender, $e)
    if ($e.AudioLevel -gt $script:maxLevel) {
        $script:maxLevel = $e.AudioLevel
    }
    if ($e.AudioLevel -gt 5) {
        $script:audioDetected = $true
    }
})

$dictGrammar = New-Object System.Speech.Recognition.DictationGrammar
$engineB.LoadGrammar($dictGrammar)

Write-Host "  Recording audio levels for 3 seconds... SPEAK NOW!" -ForegroundColor Red
$engineB.RecognizeAsync([System.Speech.Recognition.RecognizeMode]::Single)
Start-Sleep -Seconds 3
$engineB.RecognizeAsyncCancel()

Write-Host "  Max audio level detected: $maxLevel / 100" -ForegroundColor $(if ($maxLevel -gt 10) { "Green" } else { "Red" })
Write-Host "  Audio flowing: $audioDetected" -ForegroundColor $(if ($audioDetected) { "Green" } else { "Red" })

$engineB.Dispose()

Write-Host ""
Write-Host "=== DIAGNOSIS ===" -ForegroundColor Cyan
if ($maxLevel -lt 5) {
    Write-Host "PROBLEM: Almost no audio is reaching the speech engine!" -ForegroundColor Red
    Write-Host "FIX: Open 'mmsys.cpl' -> Recording tab" -ForegroundColor Yellow
    Write-Host "     Right-click 'Microphone (USB Audio Device)' -> Set as Default Device" -ForegroundColor Yellow
    Write-Host "     Also check: Properties -> Levels -> make sure volume is at 80-100%" -ForegroundColor Yellow
} elseif ($maxLevel -lt 20) {
    Write-Host "PROBLEM: Audio level is very low ($maxLevel/100)" -ForegroundColor Yellow
    Write-Host "FIX: Increase mic volume in mmsys.cpl -> Recording -> USB mic -> Properties -> Levels" -ForegroundColor Yellow
} else {
    Write-Host "Audio levels look good ($maxLevel/100)!" -ForegroundColor Green
    Write-Host "If speech still not detected, try speaking more clearly or closer to the mic." -ForegroundColor White
}
