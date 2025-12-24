# build.ps1

# 環境変数設定
$env:GOROOT = "C:\Program Files\Go"
$env:PATH = "$env:GOROOT\bin;C:\msys64\mingw64\bin;$env:PATH"
$env:CGO_ENABLED = "1"
$env:PKG_CONFIG_PATH = "C:\msys64\mingw64\lib\pkgconfig"

Write-Host "`n🔧 Checking build tools..." -ForegroundColor Cyan

# Go確認
try {
    $goVersion = go version 2>&1
    Write-Host "✅ Go: $goVersion" -ForegroundColor Green
} catch {
    Write-Host "❌ Go not found!" -ForegroundColor Red
    exit 1
}

# GCC確認
try {
    $gccVersion = gcc --version 2>&1 | Select-Object -First 1
    Write-Host "✅ GCC: $gccVersion" -ForegroundColor Green
} catch {
    Write-Host "❌ GCC not found!" -ForegroundColor Red
    exit 1
}

# pkg-config確認
try {
    $pkgVersion = pkg-config --version 2>&1
    Write-Host "✅ pkg-config: version $pkgVersion" -ForegroundColor Green
} catch {
    Write-Host "❌ pkg-config not found!" -ForegroundColor Red
    exit 1
}

# PortAudio確認
try {
    $paVersion = pkg-config --modversion portaudio-2.0 2>&1
    Write-Host "✅ PortAudio: version $paVersion" -ForegroundColor Green
} catch {
    Write-Host "⚠️  PortAudio package info not found (may still work)" -ForegroundColor Yellow
}

Write-Host "`n🔨 Building audio engine..." -ForegroundColor Green
go build -o audio_engine.exe main.go

if ($LASTEXITCODE -eq 0) {
    Write-Host "`n✅ Build successful!" -ForegroundColor Green
    $fileSize = (Get-Item audio_engine.exe).Length / 1MB
    Write-Host "📁 Output: audio_engine.exe ($([math]::Round($fileSize, 2)) MB)" -ForegroundColor Cyan
    Write-Host "`n▶️  Run with: .\run.ps1 or .\audio_engine.exe" -ForegroundColor Yellow
} else {
    Write-Host "`n❌ Build failed!" -ForegroundColor Red
    exit 1
}