<script>
    export let mixerStatus = {};
    export let onCrossfaderChange = () => {};
    export let onMasterVolumeChange = () => {};
    export let onSyncToggle = () => {};

    // --- 新規Export: ゲインとCUEを追加 ---
    export let onGainChange = () => {};
    export let onCueToggle = () => {};

    // deckIdに対応する情報を安全に取得するヘルパー
    // (APIからのレスポンス形式が大文字小文字混在する可能性があるため)
    $: getDeckStatus = (deckId) => {
        const key = deckId === 'a' ? 'DeckA' : 'DeckB';
        const keyLower = deckId === 'a' ? 'deckA' : 'deckB';
        return mixerStatus[key] || mixerStatus[keyLower] || {};
    };

    $: deckA = getDeckStatus('a');
    $: deckB = getDeckStatus('b');

    $: crossfaderPercent = (( (mixerStatus.Crossfader ?? mixerStatus.crossfader ?? 0) + 1) * 50).toFixed(0);
    $: crossfaderVal = mixerStatus.Crossfader ?? mixerStatus.crossfader ?? 0;
    $: masterVol = mixerStatus.MasterVolume ?? mixerStatus.masterVolume ?? 1.0;
    $: syncEnabled = mixerStatus.SyncEnabled ?? mixerStatus.syncEnabled ?? false;
    $: syncMaster = mixerStatus.SyncMaster ?? mixerStatus.syncMaster ?? 'a';

    $: crossfaderLabel = crossfaderVal < -0.3 ? 'A' : 
                         crossfaderVal > 0.3 ? 'B' : 'CENTER';

    function handleCrossfader(e) {
        onCrossfaderChange(parseFloat(e.target.value));
    }

    function handleMaster(e) {
        onMasterVolumeChange(parseFloat(e.target.value));
    }

    function toggleSync() {
        const master = syncMaster || 'a';
        onSyncToggle(!syncEnabled, master);
    }

    function switchSyncMaster() {
        const newMaster = syncMaster === 'a' ? 'b' : 'a';
        onSyncToggle(syncEnabled, newMaster);
    }

    function handleGain(deckId, e) {
        onGainChange(deckId, parseFloat(e.target.value));
    }

    function toggleCue(deckId) {
        const status = deckId === 'a' ? deckA : deckB;
        onCueToggle(deckId, !status?.CueEnabled);
    }
</script>

<div class="pro-mixer">
    <div class="mixer-header">
        <h2>MIXER</h2>
    </div>

    <div class="deck-controls">
        <div class="deck-channel deck-a-channel">
            <div class="gain-control">
                <label for="gain-a">GAIN A</label>
                <div class="knob-wrapper">
                    <input 
                        id="gain-a"
                        type="range" 
                        min="0" max="2" step="0.01" 
                        value={deckA.Gain ?? 1.0}
                        on:input={(e) => handleGain('a', e)}
                        class="gain-knob"
                        title="Gain A"
                    />
                </div>
                <span class="gain-value">{(deckA.Gain ?? 1.0).toFixed(2)}</span>
            </div>
            <button 
                class="cue-btn"
                class:active={deckA.CueEnabled}
                on:click={() => toggleCue('a')}
            >
                🎧 CUE A
            </button>
        </div>

        <div class="deck-channel-spacer"></div>

        <div class="deck-channel deck-b-channel">
            <div class="gain-control">
                <label for="gain-b">GAIN B</label>
                <div class="knob-wrapper">
                    <input 
                        id="gain-b"
                        type="range" 
                        min="0" max="2" step="0.01" 
                        value={deckB.Gain ?? 1.0}
                        on:input={(e) => handleGain('b', e)}
                        class="gain-knob"
                        title="Gain B"
                    />
                </div>
                <span class="gain-value">{(deckB.Gain ?? 1.0).toFixed(2)}</span>
            </div>
            <button 
                class="cue-btn"
                class:active={deckB.CueEnabled}
                on:click={() => toggleCue('b')}
            >
                🎧 CUE B
            </button>
        </div>
    </div>

    <div class="sync-section">
        <button 
            class="sync-btn" 
            class:active={syncEnabled}
            on:click={toggleSync}
        >
            <span class="sync-icon">🔗</span>
            <span>SYNC</span>
        </button>
        
        {#if syncEnabled}
            <button class="sync-master-btn" on:click={switchSyncMaster}>
                Master: Deck {syncMaster.toUpperCase()}
            </button>
        {/if}
    </div>

    <div class="bpm-comparison">
        <div class="bpm-deck bpm-a">
            <span class="bpm-label">A</span>
            <span class="bpm-value">{deckA.BPM?.toFixed(1) || '--'}</span>
        </div>
        <div class="bpm-sync-icon">⇄</div>
        <div class="bpm-deck bpm-b">
            <span class="bpm-label">B</span>
            <span class="bpm-value">{deckB.BPM?.toFixed(1) || '--'}</span>
        </div>
    </div>

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
                    value={crossfaderVal}
                    on:input={handleCrossfader}
                />
                <div class="crossfader-fill" style="width: {crossfaderPercent}%"></div>
                <div class="crossfader-thumb" style="left: {crossfaderPercent}%"></div>
            </div>
            
            <span class="deck-indicator deck-b">B</span>
        </div>
        
        <div class="crossfader-position">{crossfaderLabel}</div>
    </div>

    <div class="master-section">
        <div class="master-label">MASTER VOLUME</div>
        
        <div class="master-container">
            <div class="master-fader">
                <div class="master-track">
                    <div class="master-fill" style="height: {masterVol * 100}%"></div>
                </div>
                <input 
                    type="range" 
                    class="master-input"
                    min="0" 
                    max="1" 
                    step="0.01"
                    value={masterVol}
                    on:input={handleMaster}
                    orient="vertical"
                />
            </div>
            
            <div class="master-value">
                {Math.round(masterVol * 100)}%
            </div>
        </div>

        <div class="vu-meter">
            {#each Array(20) as _, i}
                <div 
                    class="vu-bar" 
                    class:active={(20 - i) <= masterVol * 20}
                    class:red={i < 4}
                    class:yellow={i >= 4 && i < 8}
                ></div>
            {/each}
        </div>
    </div>

    <div class="mixer-info">
        <div class="info-row">
            <span class="info-label">Deck A</span>
            <span class="info-status" class:playing={deckA.IsPlaying || deckA.Playing}>
                {deckA.IsPlaying || deckA.Playing ? '▶ PLAYING' : '⏸ PAUSED'}
            </span>
        </div>
        <div class="info-row">
            <span class="info-label">Deck B</span>
            <span class="info-status" class:playing={deckB.IsPlaying || deckB.Playing}>
                {deckB.IsPlaying || deckB.Playing ? '▶ PLAYING' : '⏸ PAUSED'}
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
        color: white;
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

    /* Deck Controls (Gain & Cue) */
    .deck-controls {
        display: flex;
        justify-content: space-between;
        align-items: flex-start;
        padding-bottom: 10px;
        border-bottom: 1px solid rgba(255,255,255,0.1);
    }
    .deck-channel {
        display: flex;
        flex-direction: column;
        align-items: center;
        gap: 10px;
        width: 80px;
    }
    .deck-channel-spacer { flex: 1; }
    
    .gain-control {
        display: flex;
        flex-direction: column;
        align-items: center;
        gap: 5px;
    }
    .gain-control label {
        font-size: 10px;
        font-weight: 700;
        color: #888;
    }
    .gain-knob {
        width: 100%;
        cursor: pointer;
    }
    .gain-value {
        font-family: 'Courier New', monospace;
        font-size: 12px;
        color: #ddd;
    }
    
    .cue-btn {
        padding: 8px 12px;
        background: rgba(255, 255, 255, 0.1);
        border: 1px solid rgba(255, 255, 255, 0.2);
        border-radius: 20px;
        color: #fff;
        font-size: 11px;
        font-weight: 700;
        cursor: pointer;
        transition: all 0.2s;
    }
    .cue-btn:hover { background: rgba(255, 255, 255, 0.2); }
    .cue-btn.active {
        background: #ffcc00;
        color: #000;
        box-shadow: 0 0 10px #ffcc00;
        border-color: #ffcc00;
    }

    /* Sync Section */
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
    .sync-btn:hover { background: rgba(255, 255, 255, 0.1); }
    .sync-btn.active {
        background: rgba(0, 170, 255, 0.2);
        border-color: #00aaff;
        box-shadow: 0 0 20px rgba(0, 170, 255, 0.5);
    }
    .sync-icon { font-size: 20px; }
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
    .sync-master-btn:hover { background: rgba(0, 170, 255, 0.2); }

    /* BPM Comparison */
    .bpm-comparison {
        display: flex;
        align-items: center;
        justify-content: space-between;
        padding: 15px;
        background: rgba(0, 0, 0, 0.5);
        border-radius: 12px;
        border: 1px solid rgba(255, 255, 255, 0.1);
    }
    .bpm-deck { display: flex; flex-direction: column; align-items: center; gap: 5px; }
    .bpm-label { font-size: 12px; color: #888; font-weight: 700; }
    .bpm-value { font-size: 28px; font-weight: 900; font-family: 'Courier New', monospace; }
    .bpm-a .bpm-value { color: #00ff88; }
    .bpm-b .bpm-value { color: #ff0088; }
    .bpm-sync-icon { font-size: 24px; color: #666; }

    /* Crossfader Section */
    .crossfader-section { display: flex; flex-direction: column; gap: 15px; }
    .crossfader-label { font-size: 12px; font-weight: 700; letter-spacing: 1px; color: #888; text-align: center; }
    .crossfader-container { display: flex; align-items: center; gap: 15px; }
    .deck-indicator { font-weight: 900; font-size: 18px; }
    .deck-indicator.deck-a { color: #00ff88; }
    .deck-indicator.deck-b { color: #ff0088; }
    
    .crossfader-track {
        flex: 1;
        height: 40px;
        background: #111;
        border-radius: 20px;
        position: relative;
        border: 1px solid #333;
        box-shadow: inset 0 0 10px rgba(0,0,0,0.8);
    }
    .crossfader-input {
        position: absolute;
        top: 0; left: 0; width: 100%; height: 100%;
        opacity: 0; cursor: pointer; z-index: 10; margin: 0;
    }
    .crossfader-fill {
        position: absolute;
        top: 0; left: 0; height: 100%;
        background: linear-gradient(90deg, #00ff8833, #ff008833);
        border-radius: 20px;
        pointer-events: none;
    }
    .crossfader-thumb {
        position: absolute;
        top: 50%;
        width: 30px; height: 30px;
        background: #fff;
        border-radius: 4px;
        transform: translate(-50%, -50%);
        box-shadow: 0 0 10px rgba(255,255,255,0.5);
        pointer-events: none;
    }
    .crossfader-position { text-align: center; font-family: 'Courier New', monospace; font-weight: bold; font-size: 14px; color: #aaa; }

    /* Master Section */
    .master-section { display: flex; flex-direction: column; align-items: center; gap: 10px; margin-top: 20px; }
    .master-label { font-size: 12px; font-weight: 700; letter-spacing: 1px; color: #888; }
    .master-container { display: flex; flex-direction: column; align-items: center; gap: 10px; }
    
    .master-fader {
        position: relative;
        width: 60px; height: 150px;
        background: #111;
        border-radius: 10px;
        border: 2px solid #333;
        overflow: hidden;
    }
    .master-track {
        position: absolute;
        bottom: 0; left: 0; width: 100%; height: 100%;
        display: flex; align-items: flex-end;
    }
    .master-fill {
        width: 100%;
        background: linear-gradient(to top, #00aaff, #00ff88);
        opacity: 0.6;
    }
    .master-input {
        position: absolute;
        top: 0; left: 0; width: 60px; height: 150px;
        opacity: 0; cursor: pointer;
        writing-mode: vertical-lr; direction: rtl;
        -webkit-appearance: none; appearance: none;
    }
    .master-value { font-family: 'Courier New', monospace; font-weight: bold; }

    /* VU Meter */
    .vu-meter {
        display: flex;
        flex-direction: column;
        gap: 2px;
        margin-top: 10px;
        background: #000;
        padding: 5px;
        border-radius: 4px;
        border: 1px solid #333;
    }
    .vu-bar {
        width: 40px; height: 4px;
        background: #222;
        border-radius: 1px;
    }
    .vu-bar.active { background: #00ff88; box-shadow: 0 0 5px #00ff88; }
    .vu-bar.active.yellow { background: #ffcc00; box-shadow: 0 0 5px #ffcc00; }
    .vu-bar.active.red { background: #ff0000; box-shadow: 0 0 5px #ff0000; }

    /* Info */
    .mixer-info { margin-top: 20px; width: 100%; font-size: 12px; color: #666; border-top: 1px solid rgba(255,255,255,0.1); padding-top: 10px; }
    .info-row { display: flex; justify-content: space-between; margin-bottom: 5px; }
    .info-status { font-weight: bold; }
    .info-status.playing { color: #00ff88; text-shadow: 0 0 5px rgba(0,255,136,0.5); }
</style>