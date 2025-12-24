<script>
    import { onMount } from 'svelte';

    // -------------------------------------------------------------------
    // 1. Props 定義 (APIクライアントを追加)
    // -------------------------------------------------------------------
    export let deckName = 'A';
    export let deckInfo = {};
    export let client = null; // 💡 追加: App.svelteからAPIクライアントを受け取る

    // -------------------------------------------------------------------
    // 2. ローカル状態
    // -------------------------------------------------------------------
    let electronAvailable = false;
    let loading = false;
    let isSeeking = false;
    
    // -------------------------------------------------------------------
    // 3. リアクティブ変数
    // -------------------------------------------------------------------
    $: progress = deckInfo.Duration > 0 ? (deckInfo.Position / deckInfo.Duration) * 100 : 0;
    $: fileName = deckInfo.File ? deckInfo.File.split(/[/\\]/).pop() : 'NO TRACK';
    $: pitchPercent = deckInfo.Speed ? ((deckInfo.Speed - 1.0) * 100).toFixed(1) : "0.0";
    $: deckColor = deckName.toLowerCase() === 'a' ? '#00ff88' : '#ff0088';
    $: title = deckInfo.FilePath ? deckInfo.FilePath.split('/').pop() : "NO TRACK";
    $: bpm = deckInfo.BPM ? deckInfo.BPM.toFixed(1) : "0.0";
    $: isPlaying = deckInfo.Playing || deckInfo.IsPlaying || false;
    $: position = deckInfo.Position || 0;
    $: duration = deckInfo.Duration || 0;

    // -------------------------------------------------------------------
    // 4. イベントハンドラ (シーク操作)
    // -------------------------------------------------------------------
    function startSeek(e) {
        // 💡 修正: タッチイベント対応 (e.buttonはMouseEventのみチェック)
        if ((e.type === 'mousedown' && e.button !== 0) || !deckInfo.FilePath) return; 
        isSeeking = true;
        document.addEventListener('mousemove', updateSeek);
        document.addEventListener('mouseup', endSeek);
        updateSeek(e);
    }

    function updateSeek(e) {
        if (!isSeeking || !deckInfo.FilePath) return;

        let clientX = e.clientX;
        if (e.touches && e.touches.length > 0) {
            clientX = e.touches[0].clientX;
        }

        const container = document.querySelector(`.pro-deck.deck-${deckName.toLowerCase()} .waveform-container`);
        if (!container) return;

        const rect = container.getBoundingClientRect();
        let x = clientX - rect.left;
        
        // 境界チェック
        if (x < 0) x = 0;
        if (x > rect.width) x = rect.width;

        const percent = x / rect.width;
        const newPos = percent * deckInfo.Duration;
        
        client?.seek(deckName, newPos); 
    }

    function endSeek() {
        if (!isSeeking) return;
        isSeeking = false;
        document.removeEventListener('mousemove', updateSeek);
        document.removeEventListener('mouseup', endSeek);
    }
    
    // -------------------------------------------------------------------
    // 5. ライフサイクルとElectron連携
    // -------------------------------------------------------------------
    onMount(() => {
        if (window.electronAPI) {
            electronAvailable = true;
            console.log("✅ Electron API Detected");
        } else {
            electronAvailable = false;
            // console.error("❌ Electron API NOT Found. Are you in a web browser?");
        }
    });

    async function handleLoadTrack() {
        if (!client) {
            alert('API client is not ready.');
            return;
        }

        if (!electronAvailable) {
            alert('Electron API not available. Make sure you are running in Electron.');
            return;
        }
        
        loading = true;
        try {
            const filePath = await window.electronAPI.selectAudioFile();
            
            if (!filePath) {
                console.log('No file selected');
                return;
            }
            
            // 💡 修正: propsで渡されたclientのloadメソッドを呼び出す
            await client.load(deckName, filePath);
            console.log(`✅ Load command sent for Deck ${deckName}`);

        } catch (error) {
            console.error(`❌ Error loading track for Deck ${deckName}:`, error);
            alert(`Error: ${error.message}`);
        } finally {
            loading = false;
        }
    }

    // -------------------------------------------------------------------
    // 6. UIユーティリティ
    // -------------------------------------------------------------------
    function formatTime(seconds) {
        const mins = Math.floor(seconds / 60);
        const secs = Math.floor(seconds % 60);
        return `${mins}:${secs.toString().padStart(2, '0')}`;
    }

    function handleEQChange(band, value) {
        const eq = { ...deckInfo.eq, [band]: parseFloat(value) };
        client?.setEQ(deckName, eq);
    }

    function handleSpeedChange(e) {
        const speed = parseFloat(e.target.value);
        client?.setSpeed(deckName, speed);
    }

    function toggleLoop() {
        // 大文字小文字の揺らぎに対応
        const currentEnabled = deckInfo.loop?.enabled || deckInfo.loop?.Enabled;
        
        // if (currentEnabled) {
        //     // ループ解除
        //     onEnableLoop(false);
        // } else {
            // 新規ループ設定
            const start = deckInfo.Position;
            
            if (!deckInfo.BPM || deckInfo.BPM <= 0) {
                 // BPMがない場合は単純にONにする（サーバー側のデフォルトに任せる）
                 console.warn('Cannot calculate 4-beat loop (No BPM). Enabling default loop.');
                 // onEnableLoop(true); 
                 return;
            }

            const beatsPerSecond = deckInfo.BPM / 60;
            const loopLength = 4 / beatsPerSecond; // 4拍分
            
            // 範囲を指定してループ
            // onSetLoop(start, start + loopLength);
        // }
    }

    $: if (deckInfo.FilePath || deckInfo.FilePath === "") {
        // console.log(`🎚️ Deck ${deckName} Status Update:`, deckInfo);
    }
</script>

<div class="pro-deck deck-{deckName.toLowerCase()}">
    <div class="deck-header">
        <div class="deck-label" style="color: {deckColor}">
            DECK {deckName.toUpperCase()}
        </div>
        <div class="bpm-display">
            {#if bpm > 0}
                <span class="bpm-value">{bpm} </span>
                <span class="bpm-label">BPM</span>
            {:else}
                <span class="bpm-label">-- BPM</span>
            {/if}
        </div>
    </div>

    <div class="track-info">
        <div class="track-name" title={deckInfo.FilePath}>{title}</div>
        <button 
            class="load-btn" 
            on:click={handleLoadTrack}
            disabled={loading || !electronAvailable}
        >
            {#if loading}
                ⏳ LOADING...
            {:else if !electronAvailable}
                ⚠️ NO ELECTRON
            {:else}
                📁 LOAD
            {/if}
        </button>
    </div>

    <div
        class="waveform-container" 
        on:mousedown={startSeek}
        on:touchstart={startSeek}
        on:touchmove={updateSeek}
        on:touchend={endSeek}
    >
        <div class="waveform-progress" style="width: {progress}%; background: {deckColor}"></div>
        <div class="playhead" style="left: {progress}%"></div>
        
        {#if deckInfo.CuePoints}
            {#each deckInfo.CuePoints as cue}
                <div class="cue-point" style="left: {(cue.Position / deckInfo.Duration) * 100}%" title={cue.name}></div>
            {/each}
        {/if}

        {#if (deckInfo.loop?.Enabled || deckInfo.loop?.enabled) && deckInfo.loop?.start !== deckInfo.loop?.end}
            <div class="loop-area" 
                 style="
                    left: {(deckInfo.loop.start / deckInfo.Duration) * 100}%;
                    width: {((deckInfo.loop.end - deckInfo.loop.start) / deckInfo.Duration) * 100}%;
                 "
                 class:active={true}
            ></div>
        {/if}
        
        <div class="timecode">
            <span class="time-current" style="color: {deckColor}">{formatTime(deckInfo.Position || 0)}</span>
            <span class="time-separator">/</span>
            <span class="time-total">{formatTime(deckInfo.Duration || 0)}</span>
        </div>
    </div>

    <div class="transport">
        <button 
            class="transport-btn play" 
            class:active={isPlaying}
            on:click={() => isPlaying ? client?.pause(deckName) : client?.play(deckName)}
            disabled={!deckInfo.FilePath}
        >
            {isPlaying ? '⏸' : '▶'}
        </button>
        <button 
            class="transport-btn stop" 
            on:click={() => client?.stop(deckName)}
            disabled={!deckInfo.FilePath}
        >
            ⏹
        </button>
        <button 
            class="transport-btn loop" 
            class:active={deckInfo.loop?.Enabled || deckInfo.loop?.enabled}
            on:click={() => { /* onEnableLoop(!deckInfo.loop?.enabled) */ }}
            disabled={!deckInfo.FilePath}
        >
            🔁
        </button>
        <button 
            class="transport-btn cue" 
            on:click={() => client?.addCuePoint(deckName, `Cue ${deckInfo.CuePoints?.length + 1 || 1}`, deckColor)}
            disabled={!deckInfo.FilePath}
        >
            📍
        </button>
    </div>

    <div class="pitch-control">
        <div class="pitch-label">PITCH</div>
        <input 
            type="range" 
            class="pitch-slider"
            min="0.8" 
            max="1.2" 
            step="0.001"
            value={deckInfo.Speed || 1.0}
            on:input={handleSpeedChange}
            style="--thumb-color: {deckColor}"
            disabled={!deckInfo.FilePath}
        />
        <div class="pitch-value" style="color: {deckColor}">
            {pitchPercent > 0 ? '+' : ''}{pitchPercent}%
        </div>
    </div>

    <div class="eq-section">
        <div class="eq-label">EQUALIZER</div>
        <div class="eq-controls">
            <div class="eq-knob">
                <input 
                    type="range" 
                    class="eq-slider"
                    min="-1" 
                    max="1" 
                    step="0.01"
                    value={deckInfo.EQ?.high || 0}
                    on:input={(e) => handleEQChange('high', e.target.value)}
                    orient="vertical"
                    disabled={!deckInfo.FilePath}
                />
                <span class="eq-label-text">HIGH</span>
            </div>
            <div class="eq-knob">
                <input 
                    type="range" 
                    class="eq-slider"
                    min="-1" 
                    max="1" 
                    step="0.01"
                    value={deckInfo.EQ?.mid || 0}
                    on:input={(e) => handleEQChange('mid', e.target.value)}
                    orient="vertical"
                    disabled={!deckInfo.FilePath}
                />
                <span class="eq-label-text">MID</span>
            </div>
            <div class="eq-knob">
                <input 
                    type="range" 
                    class="eq-slider"
                    min="-1" 
                    max="1" 
                    step="0.01"
                    value={deckInfo.EQ?.low || 0}
                    on:input={(e) => handleEQChange('low', e.target.value)}
                    orient="vertical"
                    disabled={!deckInfo.FilePath}
                />
                <span class="eq-label-text">LOW</span>
            </div>
        </div>
    </div>

    <div class="volume-fader">
        <div class="volume-label">VOLUME</div>
        <input 
            type="range" 
            class="volume-slider"
            min="0" 
            max="1" 
            step="0.01"
            value={deckInfo.Volume || 1.0}
            on:input={(e) => client?.setVolume(deckName, parseFloat(e.target.value))}
            orient="vertical"
            disabled={!deckInfo.FilePath}
        />
        <div class="volume-meter">
            <div class="volume-fill" style="height: {(deckInfo.Volume || 1.0) * 100}%; background: {deckColor}"></div>
        </div>
    </div>
</div>

<style>
/* スタイルは提供されたものをそのまま使用 */
.pro-deck {
    display: flex;
    flex-direction: column;
    gap: 15px;
    padding: 20px;
    background: rgba(0, 0, 0, 0.6);
    backdrop-filter: blur(20px);
    border-radius: 16px;
    border: 2px solid rgba(255, 255, 255, 0.1);
    box-shadow: 0 8px 32px rgba(0, 0, 0, 0.5);
}

.deck-a {
    border-color: rgba(0, 255, 136, 0.3);
}

.deck-b {
    border-color: rgba(255, 0, 136, 0.3);
}

.deck-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
}

.deck-label {
    font-size: 24px;
    font-weight: 900;
    letter-spacing: 2px;
    text-shadow: 0 0 20px currentColor;
}

.bpm-display {
    display: flex;
    align-items: baseline;
    gap: 5px;
}

.bpm-value {
    font-size: 32px;
    font-weight: 900;
    color: #fff;
    font-family: 'Courier New', monospace;
}

.bpm-label {
    font-size: 12px;
    color: #888;
    letter-spacing: 1px;
}

.track-info {
    display: flex;
    gap: 10px;
    align-items: center;
}

.track-name {
    flex: 1;
    padding: 12px 16px;
    background: rgba(0, 0, 0, 0.5);
    border-radius: 8px;
    font-size: 14px;
    font-weight: 600;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    border: 1px solid rgba(255, 255, 255, 0.1);
}

.load-btn {
    padding: 12px 24px;
    background: linear-gradient(135deg, rgba(255, 255, 255, 0.1), rgba(255, 255, 255, 0.05));
    border: 1px solid rgba(255, 255, 255, 0.2);
    border-radius: 8px;
    color: #fff;
    font-weight: 700;
    cursor: pointer;
    transition: all 0.3s;
    white-space: nowrap;
}

.load-btn:hover:not(:disabled) {
    background: linear-gradient(135deg, rgba(255, 255, 255, 0.2), rgba(255, 255, 255, 0.1));
    transform: translateY(-2px);
}

.load-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
    transform: none;
}

.waveform-container {
    position: relative;
    height: 100px;
    background: rgba(0, 0, 0, 0.8);
    border-radius: 8px;
    overflow: hidden;
    cursor: pointer;
    border: 1px solid rgba(255, 255, 255, 0.1);
    touch-action: none; /* 💡 スクロール阻害の警告を消すためにCSSで制御 */
}

.waveform-progress {
    position: absolute;
    top: 0;
    left: 0;
    height: 100%;
    opacity: 0.3;
    transition: width 0.1s linear;
}

.playhead {
    position: absolute;
    top: 0;
    width: 2px;
    height: 100%;
    background: #fff;
    box-shadow: 0 0 10px #fff;
    transition: left 0.1s linear;
}

/* キューポイントの視覚化 */
.cue-point {
    position: absolute;
    top: 0;
    width: 3px;
    height: 100%;
    background: yellow;
    box-shadow: 0 0 5px yellow;
    cursor: pointer;
    z-index: 5;
}

/* ループエリアの視覚化 */
.loop-area {
    position: absolute;
    top: 0;
    height: 100%;
    background: rgba(0, 170, 255, 0.1);
    border-left: 1px solid #00aaff;
    border-right: 1px solid #00aaff;
    z-index: 2;
}

.timecode {
    position: absolute;
    bottom: 10px;
    right: 10px;
    display: flex;
    gap: 5px;
    font-family: 'Courier New', monospace;
    font-size: 18px;
    font-weight: 700;
    text-shadow: 0 0 5px #000;
    z-index: 10;
}

.time-separator {
    color: #666;
}

.transport {
    display: flex;
    gap: 10px;
}

.transport-btn {
    flex: 1;
    padding: 20px;
    background: rgba(255, 255, 255, 0.05);
    border: 2px solid rgba(255, 255, 255, 0.1);
    border-radius: 12px;
    color: #fff;
    font-size: 24px;
    cursor: pointer;
    transition: all 0.2s;
}

.transport-btn:hover:not(:disabled) {
    background: rgba(255, 255, 255, 0.1);
    transform: translateY(-2px);
}

.transport-btn:disabled {
    opacity: 0.3;
    cursor: not-allowed;
    transform: none;
}

.transport-btn.active {
    background: rgba(0, 255, 136, 0.2);
    border-color: #00ff88;
    box-shadow: 0 0 20px rgba(0, 255, 136, 0.5);
}

.deck-b .transport-btn.active {
    background: rgba(255, 0, 136, 0.2);
    border-color: #ff0088;
    box-shadow: 0 0 20px rgba(255, 0, 136, 0.5);
}

.pitch-control {
    display: flex;
    flex-direction: column;
    gap: 10px;
}

.pitch-label {
    font-size: 12px;
    font-weight: 700;
    letter-spacing: 1px;
    color: #888;
}

.pitch-slider {
    width: 100%;
    height: 40px;
    -webkit-appearance: none;
    background: rgba(255, 255, 255, 0.05);
    border-radius: 20px;
    outline: none;
}

.pitch-slider:disabled {
    opacity: 0.3;
    cursor: not-allowed;
}

.pitch-slider::-webkit-slider-thumb {
    -webkit-appearance: none;
    width: 30px;
    height: 30px;
    background: var(--thumb-color, #00ff88);
    border-radius: 50%;
    cursor: pointer;
    box-shadow: 0 0 15px var(--thumb-color, #00ff88);
}

.pitch-value {
    font-size: 20px;
    font-weight: 900;
    text-align: center;
    font-family: 'Courier New', monospace;
}

.eq-section {
    display: flex;
    flex-direction: column;
    gap: 10px;
}

.eq-label {
    font-size: 12px;
    font-weight: 700;
    letter-spacing: 1px;
    color: #888;
}

.eq-controls {
    display: flex;
    gap: 15px;
    justify-content: space-around;
}

.eq-knob {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 10px;
}

.eq-slider {
    width: 40px;
    height: 120px;
    background: rgba(255, 255, 255, 0.05);
    border-radius: 20px;
    writing-mode: vertical-lr;
    direction: rtl;
    -webkit-appearance: none;
    appearance: none;
}

.eq-slider:disabled {
    opacity: 0.3;
    cursor: not-allowed;
}

.eq-slider::-webkit-slider-thumb {
    -webkit-appearance: none;
    width: 30px;
    height: 30px;
    background: #00aaff;
    border-radius: 50%;
    cursor: pointer;
    box-shadow: 0 0 10px #00aaff;
}

.eq-label-text {
    font-size: 11px;
    font-weight: 700;
    color: #888;
    letter-spacing: 1px;
}

.volume-fader {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 10px;
    position: relative;
}

.volume-label {
    font-size: 12px;
    font-weight: 700;
    letter-spacing: 1px;
    color: #888;
}

.volume-meter {
    width: 60px;
    height: 150px;
    background: rgba(0, 0, 0, 0.8);
    border-radius: 30px;
    position: relative;
    overflow: hidden;
    border: 2px solid rgba(255, 255, 255, 0.1);
}

.volume-fill {
    position: absolute;
    bottom: 0;
    width: 100%;
    transition: height 0.1s;
    box-shadow: 0 0 20px currentColor;
}

.volume-slider {
    position: absolute;
    width: 60px;
    height: 150px;
    opacity: 0;
    cursor: pointer;
    z-index: 10;
    writing-mode: vertical-lr;
    direction: rtl;
    -webkit-appearance: none;
    appearance: none;
}

.volume-slider:disabled {
    cursor: not-allowed;
}
</style>