# test-dj-mixer.ps1

$baseUrl = "http://localhost:8080"

function Invoke-API {
    param(
        [string]$Endpoint,
        [string]$Method = "GET",
        [hashtable]$JsonBody = $null,
        [string]$Description
    )
    
    Write-Host "`n$Description" -ForegroundColor Cyan
    
    try {
        $params = @{
            Uri = "$baseUrl$Endpoint"
            Method = $Method
        }
        
        if ($JsonBody) {
            $params.ContentType = "application/json"
            $params.Body = ($JsonBody | ConvertTo-Json)
        }
        
        $response = Invoke-WebRequest @params
        $content = $response.Content | ConvertFrom-Json | ConvertTo-Json -Depth 10
        Write-Host $content -ForegroundColor Green
        
    } catch {
        Write-Host "Error: $_" -ForegroundColor Red
    }
}

Write-Host "`n🎛️  Testing DJ Mixer" -ForegroundColor Yellow
Write-Host "═══════════════════════════════════════`n" -ForegroundColor Yellow

# 1. ミキサーステータス確認
Invoke-API -Endpoint "/api/mixer/status" -Description "1. Check mixer status"

# 2. Deck Aにトラックをロード
$trackA = "C:/composer-dj-app/go_audio_engine/testdata/tone_440hz.wav"
Invoke-API -Endpoint "/api/deck/a/load?file=$trackA" -Description "2. Load track to Deck A (440Hz)"

# 3. Deck Bにトラックをロード
$trackB = "C:/composer-dj-app/go_audio_engine/testdata/tone_523hz.wav"
Invoke-API -Endpoint "/api/deck/b/load?file=$trackB" -Description "3. Load track to Deck B (523Hz)"

Start-Sleep -Seconds 1

# 4. Deck A 再生
Invoke-API -Endpoint "/api/deck/a/play" -Method POST -Description "4. Play Deck A (440Hz tone)"

Start-Sleep -Seconds 2

# 5. Deck A 音量を50%に
Invoke-API -Endpoint "/api/deck/a/volume" -Method POST -JsonBody @{volume = 0.5} -Description "5. Set Deck A volume to 50%"

Start-Sleep -Seconds 2

# 6. Deck B も再生開始
Invoke-API -Endpoint "/api/deck/b/play" -Method POST -Description "6. Play Deck B (523Hz tone)"

Start-Sleep -Seconds 1

# 7. クロスフェーダーを動かす（A → 中央 → B）
Write-Host "`n7. Testing crossfader animation..." -ForegroundColor Cyan

Invoke-API -Endpoint "/api/mixer/crossfader" -Method POST -JsonBody @{value = -1.0} -Description "   [====A====|     B    ] Full A"
Start-Sleep -Seconds 1

Invoke-API -Endpoint "/api/mixer/crossfader" -Method POST -JsonBody @{value = -0.5} -Description "   [===A=====|=    B   ] A side"
Start-Sleep -Seconds 1

Invoke-API -Endpoint "/api/mixer/crossfader" -Method POST -JsonBody @{value = 0.0} -Description "   [==A======|======B==] Center (both equal)"
Start-Sleep -Seconds 1

Invoke-API -Endpoint "/api/mixer/crossfader" -Method POST -JsonBody @{value = 0.5} -Description "   [=A=======|=======B=] B side"
Start-Sleep -Seconds 1

Invoke-API -Endpoint "/api/mixer/crossfader" -Method POST -JsonBody @{value = 1.0} -Description "   [A========|========B] Full B"
Start-Sleep -Seconds 2

# 8. マスターボリュームテスト
Write-Host "`n8. Testing master volume..." -ForegroundColor Cyan
Invoke-API -Endpoint "/api/mixer/master" -Method POST -JsonBody @{volume = 0.3} -Description "   Master volume: 30%"
Start-Sleep -Seconds 1

Invoke-API -Endpoint "/api/mixer/master" -Method POST -JsonBody @{volume = 1.0} -Description "   Master volume: 100%"
Start-Sleep -Seconds 1

# 9. シークテスト
Write-Host "`n9. Testing seek functionality..." -ForegroundColor Cyan
Invoke-API -Endpoint "/api/deck/a/seek" -Method POST -JsonBody @{position = 2.0} -Description "   Deck A: Seek to 2.0 seconds"
Start-Sleep -Seconds 1

# 10. Deck A 停止
Invoke-API -Endpoint "/api/deck/a/stop" -Method POST -Description "10. Stop Deck A"

Start-Sleep -Seconds 1

# 11. Deck B 停止
Invoke-API -Endpoint "/api/deck/b/stop" -Method POST -Description "11. Stop Deck B"

# 12. 最終ステータス
Invoke-API -Endpoint "/api/mixer/status" -Description "12. Final mixer status"

Write-Host "`n✅ All DJ mixer tests completed!" -ForegroundColor Green
Write-Host "You should have heard:" -ForegroundColor Yellow
Write-Host "  - 440Hz tone from Deck A" -ForegroundColor White
Write-Host "  - 523Hz tone from Deck B" -ForegroundColor White
Write-Host "  - Crossfader transitioning between both" -ForegroundColor White
