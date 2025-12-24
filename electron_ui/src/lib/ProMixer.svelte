<script>
    export let mixerStatus = {};
    export let onCrossfaderChange = () => {};
    export let onMasterVolumeChange = () => {};
    export let onSyncToggle = () => {};

    $: crossfaderPercent = ((mixerStatus.crossfader + 1) * 50).toFixed(0);
    $: crossfaderLabel = mixerStatus.crossfader < -0.3 ? 'A' : 
                         mixerStatus.crossfader > 0.3 ? 'B' : 'CENTER';

    function handleCrossfader(e) {
        onCrossfaderChange(parseFloat(e.target.value));
    }

    function handleMaster(e) {
        onMasterVolumeChange(parseFloat(e.target.value));
    }

    function toggleSync() {
        const newEnabled = !mixerStatus.syncEnabled;
        const master = mixerStatus.syncMaster || 'a';
        onSyncToggle(newEnabled, master);
    }

    function switchSyncMaster() {
        const newMaster = mixerStatus.syncMaster === 'a' ? 'b' : 'a';
        onSyncToggle(mixerStatus.syncEnabled, newMaster);
    }
</script>

<div class="pro-mixer">
    <!-- ミキサーヘッダー -->
    <div class="mixer-header">
        <h2>MIXER</h2>
    </div>

    <!-- BPM同期セクション -->
    <div class="sync-section">
        <button 
            class="sync-btn" 
            class:active={mixerStatus.syncEnabled}
            on:click={toggleSync}
        >
            <span class="sync-icon">🔗</span>
            <span>SYNC</span>
        </button>
        
        {#if mixerStatus.syncEnabled}
            <button class="sync-master-btn" on:click={switchSyncMaster}>
                Master: Deck {mixerStatus.syncMaster?.toUpperCase()}
            </button>
        {/if}
    </div>

    <!-- BPM表示 -->
    <div class="bpm-comparison">
        <div class="bpm-deck bpm-a">
            <span class="bpm-label">A</span>
            <span class="bpm-value">{mixerStatus.deckA?.bpm?.toFixed(1) || '--'}</span>
        </div>
        <div class="bpm-sync-icon">⇄</div>
        <div class="bpm-deck bpm-b">
            <span class="bpm-label">B</span>
            <span class="bpm-value">{mixerStatus.deckB?.bpm?.toFixed(1) || '--'}</span>
        </div>
    </div>

    <!-- クロスフェーダー -->
    <div class="crossfader-section">
        <div class="crossfader-label">CROSSFADER</div>
        
        <div class="crossfader-container">
            <span class="deck-indicator deck-a">A</span>
            
            <div class="crossfader-track">
                <input 
                    type="range" 
                    class="crossfader-input"
                    min="-1" 
                    max="1" 
                    step="0.001"
                    value={mixerStatus.crossfader}
                    on:input={handleCrossfader}
                />
                <div class="crossfader-fill" style="width: {crossfaderPercent}%"></div>
                <div class="crossfader-thumb" style="left: {crossfaderPercent}%"></div>
            </div>
            
            <span class="deck-indicator deck-b">B</span>
        </div>
        
        <div class="crossfader-position">{crossfaderLabel}</div>
    </div>

    <!-- マスターボリューム -->
    <div class="master-section">
        <div class="master-label">MASTER VOLUME</div>
        
        <div class="master-container">
            <div class="master-fader">
                <div class="master-track">
                    <div class="master-fill" style="height: {mixerStatus.masterVolume * 100}%"></div>
                </div>
                <input 
                    type="range" 
                    class="master-input"
                    min="0" 
                    max="1" 
                    step="0.01"
                    value={mixerStatus.masterVolume}
                    on:input={handleMaster}
                    orient="vertical"
                />
            </div>
            
            <div class="master-value">
                {Math.round(mixerStatus.masterVolume * 100)}%
            </div>
        </div>

        <!-- VUメーター風表示 -->
        <div class="vu-meter">
            {#each Array(20) as _, i}
                <div 
                    class="vu-bar" 
                    class:active={i < mixerStatus.masterVolume * 20}
                    class:red={i >= 16}
                    class:yellow={i >= 12 && i < 16}
                ></div>
            {/each}
        </div>
    </div>

    <!-- ミキサー情報 -->
    <div class="mixer-info">
        <div class="info-row">
            <span class="info-label">Deck A</span>
            <span class="info-status" class:playing={mixerStatus.deckA?.playing}>
                {mixerStatus.deckA?.playing ? '▶ PLAYING' : '⏸ PAUSED'}
            </span>
        </div>
        <div class="info-row">
            <span class="info-label">Deck B</span>
            <span class="info-status" class:playing={mixerStatus.deckB?.playing}>
                {mixerStatus.deckB?.playing ? '▶ PLAYING' : '⏸ PAUSED'}
            </span>
        </div>
    </div>
</div>

<style>
    .pro-mixer {
        display: flex;
        flex-direction: column;
        gap: 20px;
        padding: 30px;
        background: rgba(0, 0, 0, 0.8);
        backdrop-filter: blur(20px);
        border-radius: 16px;
        border: 2px solid rgba(255, 255, 255, 0.1);
        box-shadow: 0 8px 32px rgba(0, 0, 0, 0.5);
    }

    .mixer-header h2 {
        margin: 0;
        text-align: center;
        font-size: 20px;
        font-weight: 900;
        letter-spacing: 3px;
        background: linear-gradient(90deg, #00ff88, #00aaff, #ff0088);
        -webkit-background-clip: text;
        -webkit-text-fill-color: transparent;
        background-clip: text;
    }

    .sync-section {
        display: flex;
        flex-direction: column;
        gap: 10px;
    }

    .sync-btn {
        display: flex;
        align-items: center;
        justify-content: center;
        gap: 10px;
        padding: 15px;
        background: rgba(255, 255, 255, 0.05);
        border: 2px solid rgba(255, 255, 255, 0.1);
        border-radius: 12px;
        color: #fff;
        font-size: 16px;
        font-weight: 700;
        cursor: pointer;
        transition: all 0.3s;
    }

    .sync-btn:hover {
        background: rgba(255, 255, 255, 0.1);
    }

    .sync-btn.active {
        background: rgba(0, 170, 255, 0.2);
        border-color: #00aaff;
        box-shadow: 0 0 20px rgba(0, 170, 255, 0.5);
    }

    .sync-icon {
        font-size: 20px;
    }

    .sync-master-btn {
        padding: 10px;
        background: rgba(0, 170, 255, 0.1);
        border: 1px solid #00aaff;
        border-radius: 8px;
        color: #00aaff;
        font-size: 12px;
        font-weight: 700;
        cursor: pointer;
        transition: all 0.2s;
    }

    .sync-master-btn:hover {
        background: rgba(0, 170, 255, 0.2);
    }

    .bpm-comparison {
        display: flex;
        align-items: center;
        justify-content: space-between;
        padding: 15px;
        background: rgba(0, 0, 0, 0.5);
        border-radius: 12px;
        border: 1px solid rgba(255, 255, 255, 0.1);
    }

    .bpm-deck {
        display: flex;
        flex-direction: column;
        align-items: center;
        gap: 5px;
    }

    .bpm-label {
        font-size: 12px;
        color: #888;
        font-weight: 700;
    }

    .bpm-value {
        font-size: 28px;
        font-weight: 900;
        font-family: 'Courier New', monospace;
    }

    .bpm-a .bpm-value {
        color: #00ff88;
    }

    .bpm-b .bpm-value {
        color: #ff0088;
    }

    .bpm-sync-icon {
        font-size: 24px;
        color: #666;
    }

    .crossfader-section {
        display: flex;
        flex-direction: column;
        gap: 15px;
    }

    .crossfader-label {
        font-size: 12px;
        font-weight: 700;
        letter-spacing: 1px;
        color: #888;
        text-align: center;
    }

    .crossfader-container {
        display: flex;
        align-items: center;
        gap: 15px;
    }

    .deck-indicator {
        font-size: 18px;
        font-weight: 900;
        width: 30px;
        text-align: center;
    }

    .deck-indicator.deck-a {
        color: #00ff88;
    }

    .deck-indicator.deck-b {
        color: #ff0088;
    }

    .crossfader-track {
        flex: 1;
        position: relative;
        height: 60px;
        background: rgba(0, 0, 0, 0.8);
        border-radius: 30px;
        border: 2px solid rgba(255, 255, 255, 0.1);
        overflow: hidden;
    }

    .crossfader-fill {
        position: absolute;
        left: 0;
        top: 0;
        height: 100%;
        background: linear-gradient(90deg, #00ff88, #ff0088);
        opacity: 0.3;
        transition: width 0.05s;
    }

    .crossfader-thumb {
        position: absolute;
        top: 50%;
        transform: translate(-50%, -50%);
        width: 50px;
        height: 50px;
        background: linear-gradient(135deg, #fff, #ddd);
        border-radius: 50%;
        box-shadow: 0 4px 20px rgba(0, 0, 0, 0.5), 0 0 0 3px rgba(255, 255, 255, 0.2);
        transition: left 0.05s;
        pointer-events: none;
    }

    .crossfader-input {
        position: absolute;
        width: 100%;
        height: 100%;
        opacity: 0;
        cursor: pointer;
        z-index: 10;
    }

    .crossfader-position {
        text-align: center;
        font-size: 16px;
        font-weight: 900;
        color: #fff;
        font-family: 'Courier New', monospace;
    }

    .master-section {
        display: flex;
        flex-direction: column;
        align-items: center;
        gap: 15px;
    }

    .master-label {
        font-size: 12px;
        font-weight: 700;
        letter-spacing: 1px;
        color: #888;
    }

    .master-container {
        display: flex;
        flex-direction: column;
        align-items: center;
        gap: 10px;
    }

    .master-fader {
        position: relative;
        width: 80px;
        height: 200px;
    }

    .master-track {
        width: 100%;
        height: 100%;
        background: rgba(0, 0, 0, 0.8);
        border-radius: 40px;
        border: 2px solid rgba(255, 255, 255, 0.1);
        position: relative;
        overflow: hidden;
    }

    .master-fill {
        position: absolute;
        bottom: 0;
        width: 100%;
        background: linear-gradient(to top, #ff0088, #ff6666);
        box-shadow: 0 0 20px currentColor;
        transition: height 0.1s;
    }

    .master-input {
        position: absolute;
        writing-mode: bt-lr;
        -webkit-appearance: slider-vertical;
        width: 80px;
        height: 200px;
        opacity: 0;
        cursor: pointer;
        z-index: 10;
    }

    .master-value {
        font-size: 24px;
        font-weight: 900;
        color: #fff;
        font-family: 'Courier New', monospace;
    }

    .vu-meter {
        display: flex;
        gap: 3px;
        height: 30px;
        align-items: flex-end;
    }

    .vu-bar {
        flex: 1;
        background: rgba(255, 255, 255, 0.1);
        border-radius: 2px;
        transition: all 0.1s;
    }

    .vu-bar.active {
        background: #00ff88;
        box-shadow: 0 0 5px #00ff88;
    }

    .vu-bar.active.yellow {
        background: #ffaa00;
        box-shadow: 0 0 5px #ffaa00;
    }

    .vu-bar.active.red {
        background: #ff0088;
        box-shadow: 0 0 5px #ff0088;
    }

    .mixer-info {
        display: flex;
        flex-direction: column;
        gap: 10px;
    }

    .info-row {
        display: flex;
        justify-content: space-between;
        align-items: center;
        padding: 10px 15px;
        background: rgba(0, 0, 0, 0.5);
        border-radius: 8px;
        border: 1px solid rgba(255, 255, 255, 0.1);
    }

    .info-label {
        font-size: 12px;
        font-weight: 700;
        color: #888;
    }

    .info-status {
        font-size: 12px;
        font-weight: 700;
        color: #666;
    }

    .info-status.playing {
        color: #00ff88;
    }
</style>