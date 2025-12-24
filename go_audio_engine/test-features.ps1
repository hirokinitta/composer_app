# test-features.ps1 - 全機能テスト

`$baseUrl = "http://localhost:8080"

function Test-API {
    param(
        [string]`$Endpoint,
        [string]`$Method = "GET",
        [hashtable]`$JsonBody = `$null,
        [string]`$Description
    )
    
    Write-Host ""
    Write-Host "`$Description" -ForegroundColor Cyan
    Write-Host ("─" * 60) -ForegroundColor DarkGray
    
    try {
        `$params = @{
            Uri = "`$baseUrl`$Endpoint"
            Method = `$Method
            TimeoutSec = 5
        }
        
        if (`$JsonBody) {
            `$params.ContentType = "application/json"
            `$params.Body = (`$JsonBody | ConvertTo-Json)
        }
        
        `$response = Invoke-WebRequest @params
        `$content = `$response.Content | ConvertFrom-Json | ConvertTo-Json -Depth 10 -Compress
        Write-Host "✅ `$content" -ForegroundColor Green
        return `$true
        
    } catch {
        Write-Host "❌ Error: `$_" -ForegroundColor Red
        return `$false
    }
}

Write-Host ""
Write-Host "🎛️  DJ Audio Engine - Feature Test Suite" -ForegroundColor Yellow
Write-Host "═══════════════════════════════════════════════" -ForegroundColor Yellow

# 接続テスト
Write-Host ""
Write-Host "🔌 Testing connection..." -ForegroundColor Magenta
if (-not (Test-API -Endpoint "/api/mixer/status" -Description "Connection test")) {
    Write-Host ""
    Write-Host "❌ Cannot connect to audio engine!" -ForegroundColor Red
    Write-Host "Please start the engine: .\audio_engine.exe" -ForegroundColor Yellow
    exit 1
}

# テストファイルのパス
`$testFileA = "C:/composer-dj-app/go_audio_engine/testdata/tone_440hz.wav"
`$testFileB = "C:/composer-dj-app/go_audio_engine/testdata/tone_523hz.wav"

Write-Host ""
Write-Host "═══════════════════════════════════════════════" -ForegroundColor Yellow
Write-Host "📀 DECK A TESTS" -ForegroundColor Yellow
Write-Host "═══════════════════════════════════════════════" -ForegroundColor Yellow

Test-API -Endpoint "/api/deck/a/load?file=`$testFileA" -Description "1. Load track to Deck A"
Start-Sleep -Milliseconds 500

Test-API -Endpoint "/api/deck/a/play" -Method POST -Description "2. Play Deck A"
Start-Sleep -Seconds 1

Test-API -Endpoint "/api/deck/a/volume" -Method POST -JsonBody @{volume = 0.7} -Description "3. Set volume to 70%"
Start-Sleep -Milliseconds 500

Test-API -Endpoint "/api/deck/a/eq" -Method POST -JsonBody @{low = 0.3; mid = -0.2; high = 0.5} -Description "4. Adjust EQ"
Start-Sleep -Milliseconds 500

Test-API -Endpoint "/api/deck/a/speed" -Method POST -JsonBody @{speed = 1.05} -Description "5. Set speed to +5%"
Start-Sleep -Seconds 1

Test-API -Endpoint "/api/deck/a/pause" -Method POST -Description "6. Pause Deck A"
Start-Sleep -Milliseconds 500

Write-Host ""
Write-Host "═══════════════════════════════════════════════" -ForegroundColor Yellow
Write-Host "📀 DECK B TESTS" -ForegroundColor Yellow
Write-Host "═══════════════════════════════════════════════" -ForegroundColor Yellow

Test-API -Endpoint "/api/deck/b/load?file=`$testFileB" -Description "7. Load track to Deck B"
Start-Sleep -Milliseconds 500

Test-API -Endpoint "/api/deck/b/play" -Method POST -Description "8. Play Deck B"
Start-Sleep -Seconds 1

Write-Host ""
Write-Host "═══════════════════════════════════════════════" -ForegroundColor Yellow
Write-Host "🎚️  MIXER TESTS" -ForegroundColor Yellow
Write-Host "═══════════════════════════════════════════════" -ForegroundColor Yellow

Test-API -Endpoint "/api/mixer/crossfader" -Method POST -JsonBody @{value = -1.0} -Description "9. Crossfader to A"
Start-Sleep -Milliseconds 500

Test-API -Endpoint "/api/mixer/crossfader" -Method POST -JsonBody @{value = 0.0} -Description "10. Crossfader to Center"
Start-Sleep -Milliseconds 500

Test-API -Endpoint "/api/mixer/crossfader" -Method POST -JsonBody @{value = 1.0} -Description "11. Crossfader to B"
Start-Sleep -Milliseconds 500

Test-API -Endpoint "/api/mixer/master" -Method POST -JsonBody @{volume = 0.8} -Description "12. Master volume to 80%"
Start-Sleep -Milliseconds 500

Test-API -Endpoint "/api/mixer/sync" -Method POST -JsonBody @{enabled = `$true; master = "a"} -Description "13. Enable BPM Sync (Master: A)"
Start-Sleep -Seconds 1

Write-Host ""
Write-Host "═══════════════════════════════════════════════" -ForegroundColor Yellow
Write-Host "🔄 ADVANCED FEATURES" -ForegroundColor Yellow
Write-Host "═══════════════════════════════════════════════" -ForegroundColor Yellow

Test-API -Endpoint "/api/deck/a/cuepoint/add" -Method POST -JsonBody @{name = "Test Cue"; color = "#00ff88"} -Description "14. Add cue point"
Start-Sleep -Milliseconds 500

Test-API -Endpoint "/api/deck/a/loop/set" -Method POST -JsonBody @{start = 1.0; end = 3.0} -Description "15. Set loop (1s - 3s)"
Start-Sleep -Milliseconds 500

Test-API -Endpoint "/api/deck/a/filter" -Method POST -JsonBody @{type = "lowpass"; cutoff = 0.5; resonance = 0.3} -Description "16. Apply lowpass filter"
Start-Sleep -Seconds 1

Test-API -Endpoint "/api/deck/a/filter" -Method POST -JsonBody @{type = "none"; cutoff = 0.5; resonance = 0} -Description "17. Remove filter"
Start-Sleep -Milliseconds 500

# クリーンアップ
Test-API -Endpoint "/api/deck/a/stop" -Method POST -Description "18. Stop Deck A"
Test-API -Endpoint "/api/deck/b/stop" -Method POST -Description "19. Stop Deck B"

Write-Host ""
Write-Host "═══════════════════════════════════════════════" -ForegroundColor Yellow
Write-Host "📊 FINAL STATUS" -ForegroundColor Yellow
Write-Host "═══════════════════════════════════════════════" -ForegroundColor Yellow

Test-API -Endpoint "/api/mixer/status" -Description "Final mixer status"

Write-Host ""
Write-Host "✅ All tests completed!" -ForegroundColor Green
Write-Host ""