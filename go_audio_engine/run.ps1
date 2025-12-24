# run.ps1

# DLLのパスを追加（実行時に必要）
$env:PATH = "C:\msys64\mingw64\bin;$env:PATH"

Write-Host "`n🎵 DJ Audio Engine" -ForegroundColor Cyan
Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor DarkGray

if (Test-Path ".\audio_engine.exe") {
    Write-Host "Starting server on http://localhost:8080" -ForegroundColor Green
    Write-Host "Press Ctrl+C to stop`n" -ForegroundColor Yellow
    .\audio_engine.exe
} else {
    Write-Host "❌ audio_engine.exe not found!" -ForegroundColor Red
    Write-Host "Run .\build.ps1 first to build the application." -ForegroundColor Yellow
    exit 1
}