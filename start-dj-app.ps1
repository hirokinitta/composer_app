# start-dj-app.ps1 - DJ アプリ完全起動スクリプト

Write-Host ""
Write-Host "🎧 Starting Professional DJ Application" -ForegroundColor Cyan
Write-Host "════════════════════════════════════════" -ForegroundColor Cyan
Write-Host ""

# 1. Goエンジンを起動
Write-Host "🚀 Starting Go Audio Engine..." -ForegroundColor Yellow
Start-Process powershell -ArgumentList @(
    '-NoExit',
    '-Command',
    'cd C:\composer-dj-app\go_audio_engine; `$env:PATH = "C:\msys64\mingw64\bin;`$env:PATH"; .\audio_engine.exe'
)

Write-Host "   Waiting for audio engine to start..." -ForegroundColor Gray
Start-Sleep -Seconds 3

# 2. 接続確認
Write-Host "🔌 Checking connection..." -ForegroundColor Yellow
try {
    `$response = Invoke-WebRequest -Uri "http://localhost:8080/api/mixer/status" -TimeoutSec 5
    Write-Host "   ✅ Audio engine is running!" -ForegroundColor Green
} catch {
    Write-Host "   ❌ Audio engine failed to start!" -ForegroundColor Red
    Write-Host "   Please check the Go terminal window for errors." -ForegroundColor Yellow
    exit 1
}

# 3. Electron UIを起動
Write-Host ""
Write-Host "🖥️  Starting Electron UI..." -ForegroundColor Yellow
Start-Process powershell -ArgumentList @(
    '-NoExit',
    '-Command',
    'cd C:\composer-dj-app\electron_ui_clean; npm run electron:dev'
)

Write-Host ""
Write-Host "✅ All services started!" -ForegroundColor Green
Write-Host ""
Write-Host "📋 Services:" -ForegroundColor Cyan
Write-Host "   • Go Audio Engine: http://localhost:8080" -ForegroundColor White
Write-Host "   • Electron UI: Starting..." -ForegroundColor White
Write-Host ""
Write-Host "💡 Tip: Close this window to keep services running" -ForegroundColor Yellow
Write-Host "     Close individual PowerShell windows to stop services" -ForegroundColor Yellow
Write-Host ""