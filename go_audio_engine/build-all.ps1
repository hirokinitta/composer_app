# build-all.ps1 - Professional DJ Audio Engine Build Script

Write-Host ""
Write-Host "🎛️  Professional DJ Audio Engine v2.0" -ForegroundColor Cyan
Write-Host "======================================" -ForegroundColor Cyan
Write-Host ""

# 環境変数設定
Write-Host "📋 Setting up environment..." -ForegroundColor Yellow
$env:GOROOT = "C:\Program Files\Go"
$env:PATH = "$env:GOROOT\bin;C:\msys64\mingw64\bin;$env:PATH"
$env:CGO_ENABLED = "1"
$env:PKG_CONFIG_PATH = "C:\msys64\mingw64\lib\pkgconfig"

# Go確認
Write-Host "✓ Go:" (go version) -ForegroundColor Green

# GCC確認
try {
    $gccVersion = gcc --version 2>&1 | Select-Object -First 1
    Write-Host "✓ GCC: $gccVersion" -ForegroundColor Green
} catch {
    Write-Host "❌ GCC not found!" -ForegroundColor Red
    Write-Host "Make sure MSYS2 is installed with GCC and PortAudio" -ForegroundColor Yellow
    exit 1
}

# 依存関係チェック
Write-Host ""
Write-Host "📦 Checking dependencies..." -ForegroundColor Yellow
go mod tidy

# ビルド
Write-Host ""
Write-Host "🔨 Building audio engine..." -ForegroundColor Yellow
go build -o audio_engine.exe main.go

if ($LASTEXITCODE -eq 0) {
    Write-Host ""
    Write-Host "✅ Build successful!" -ForegroundColor Green
    $fileSize = (Get-Item audio_engine.exe).Length / 1MB
    Write-Host "📁 Output: audio_engine.exe ($([math]::Round($fileSize, 2)) MB)" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "▶️  Run with: .\audio_engine.exe" -ForegroundColor Yellow
} else {
    Write-Host ""
    Write-Host "❌ Build failed!" -ForegroundColor Red
    exit 1
}