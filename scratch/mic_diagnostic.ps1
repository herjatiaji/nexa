# JARVIS Microphone Deep Diagnostic
# This script checks everything about your audio input setup

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  JARVIS Microphone Deep Diagnostic" -ForegroundColor Cyan  
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# 1. List ALL audio recording devices
Write-Host "[1] Recording Devices (via WMI):" -ForegroundColor Yellow
Get-WmiObject Win32_SoundDevice | ForEach-Object {
    Write-Host "    - $($_.Name) | Status: $($_.Status)" -ForegroundColor White
}
Write-Host ""

# 2. Check default audio endpoint  
Write-Host "[2] Default Audio Input Endpoint:" -ForegroundColor Yellow
try {
    Add-Type -TypeDefinition @"
using System;
using System.Runtime.InteropServices;

[Guid("D666063F-1587-4E43-81F1-B948E807363F"), InterfaceType(ComInterfaceType.InterfaceIsIUnknown)]
interface IMMDevice {
    int Activate(ref Guid id, int clsCtx, IntPtr activationParams, [MarshalAs(UnmanagedType.IUnknown)] out object ppInterface);
    int OpenPropertyStore(int stgmAccess, [MarshalAs(UnmanagedType.IUnknown)] out object ppProperties);
    int GetId([MarshalAs(UnmanagedType.LPWStr)] out string ppstrId);
    int GetState(out int pdwState);
}
"@ -ErrorAction SilentlyContinue
    Write-Host "    (COM endpoint check skipped - using PowerShell method)" -ForegroundColor Gray
} catch {}

# 3. Test System.Speech can access mic
Write-Host "[3] System.Speech Engine Test:" -ForegroundColor Yellow
Add-Type -AssemblyName System.Speech
try {
    $engine = New-Object System.Speech.Recognition.SpeechRecognitionEngine
    $engine.SetInputToDefaultAudioDevice()
    
    # Load a simple grammar
    $choices = New-Object System.Speech.Recognition.Choices
    $choices.Add(@("Jarvis", "Hey Jarvis", "Hello", "Test", "One", "Two", "Three", "Yes", "No"))
    $gb = New-Object System.Speech.Recognition.GrammarBuilder
    $gb.Append($choices)
    $grammar = New-Object System.Speech.Recognition.Grammar($gb)
    $engine.LoadGrammar($grammar)
    
    Write-Host "    Engine created: OK" -ForegroundColor Green
    Write-Host "    Audio input set: OK" -ForegroundColor Green
    Write-Host "    Grammar loaded: OK" -ForegroundColor Green
    
    # 4. Actual listen test with detailed result
    Write-Host "" 
    Write-Host "[4] LIVE LISTEN TEST (5 seconds):" -ForegroundColor Yellow
    Write-Host "    >>> SAY SOMETHING NOW! ('Jarvis', 'Hello', 'Test', 'One', 'Two', 'Three') <<<" -ForegroundColor Red
    Write-Host "    Listening..." -ForegroundColor White
    
    $result = $engine.Recognize([TimeSpan]::FromSeconds(5))
    
    if ($result) {
        Write-Host "    DETECTED: '$($result.Text)'" -ForegroundColor Green
        Write-Host "    Confidence: $([math]::Round($result.Confidence, 3))" -ForegroundColor Green
        Write-Host "    Grammar: $($result.Grammar.Name)" -ForegroundColor Green
        
        if ($result.Alternates) {
            Write-Host "    Alternates:" -ForegroundColor Gray
            foreach ($alt in $result.Alternates) {
                Write-Host "      - '$($alt.Text)' (confidence: $([math]::Round($alt.Confidence, 3)))" -ForegroundColor Gray
            }
        }
    } else {
        Write-Host "    RESULT: No speech detected in 5 seconds" -ForegroundColor Red
        Write-Host ""
        Write-Host "    Possible causes:" -ForegroundColor Yellow
        Write-Host "      1. Mic is muted or volume too low" -ForegroundColor White
        Write-Host "      2. Wrong mic set as default device" -ForegroundColor White
        Write-Host "      3. Privacy settings blocking mic access" -ForegroundColor White
    }
    
    $engine.Dispose()
    
} catch {
    Write-Host "    ERROR: $($_.Exception.Message)" -ForegroundColor Red
}

# 5. Check Windows mic privacy settings
Write-Host ""
Write-Host "[5] Microphone Privacy Settings:" -ForegroundColor Yellow
try {
    $micAccess = Get-ItemProperty -Path "HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\CapabilityAccessManager\ConsentStore\microphone" -Name "Value" -ErrorAction SilentlyContinue
    if ($micAccess.Value -eq "Allow") {
        Write-Host "    System mic access: ALLOWED" -ForegroundColor Green
    } else {
        Write-Host "    System mic access: $($micAccess.Value) (might be blocked!)" -ForegroundColor Red
    }
} catch {
    Write-Host "    Could not check privacy settings" -ForegroundColor Gray
}

# 6. Show active audio endpoints via PowerShell
Write-Host ""
Write-Host "[6] Audio Endpoints (Get-PnpDevice):" -ForegroundColor Yellow
try {
    Get-PnpDevice -Class AudioEndpoint -Status OK 2>$null | Select-Object -First 10 | ForEach-Object {
        Write-Host "    - $($_.FriendlyName) [$($_.Status)]" -ForegroundColor White
    }
} catch {
    Write-Host "    (Could not enumerate - not critical)" -ForegroundColor Gray
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Diagnostic Complete" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
