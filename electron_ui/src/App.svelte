<script>
    import { onMount, onDestroy } from 'svelte';
    import ProDeck from './lib/ProDeck.svelte';
    import ProMixer from './lib/ProMixer.svelte';
    import WaveformDisplay from './lib/WaveformDisplay.svelte';
    import { AudioEngineClient } from './lib/audioEngineClient.js';

    const audioEngine = new AudioEngineClient();
    
    let connected = false;
    let mixerStatus = {
        deckA: {
            file: '',
            playing: false,
            position: 0,
            duration: 0,
            volume: 1.0,
            speed: 1.0,
            bpm: 0,
            eq: { low: 0, mid: 0, high: 0 },
            filter: { type: 'none', cutoff: 0.5, resonance: 0 },
            cuePoints: [],
            loop: { enabled: false, start: 0, end: 0, isActive: false }
        },
        deckB: {
            file: '',
            playing: false,
            position: 0,
            duration: 0,
            volume: 1.0,
            speed: 1.0,
            bpm: 0,
            eq: { low: 0, mid: 0, high: 0 },
            filter: { type: 'none', cutoff: 0.5, resonance: 0 },
            cuePoints: [],
            loop: { enabled: false, start: 0, end: 0, isActive: false }
        },
        crossfader: 0,
        masterVolume: 1.0,
        syncEnabled: false,
        syncMaster: 'a'
    };

    let statusInterval;
    let showSettings = false;

    onMount(async () => {
        connected = await audioEngine.checkConnection();
        
        if (connected) {
            updateStatus();
            statusInterval = setInterval(updateStatus, 50); // 高頻度更新
        }
    });

    onDestroy(() => {
        if (statusInterval) clearInterval(statusInterval);
    });

    async function updateStatus() {
        try {
            const status = await audioEngine.getMixerStatus();
            mixerStatus = status;
            if (!connected) connected = true;
        } catch {
            connected = false;
        }
    }

    async function loadTrack(deck) {
        let filePath;
        
        if (window.require) {
            try {
            const { ipcRenderer } = window.require('electron');
                filePath = await ipcRenderer.invoke('select-audio-file');
            } catch (error) {
                console.error('Electron IPC error:', error);
            }
        } else {
            filePath = prompt(
                `Enter WAV file path for Deck ${deck.toUpperCase()}:`,
                `C:/composer-dj-app/go_audio_engine/testdata/tone_${deck === 'a' ? '440' : '523'}hz.wav`
            );
        }
        
        if (filePath) {
            filePath = filePath.replace(/^["']|["']$/g, '').trim().replace(/\\/g, '/');
            console.log('Loading track:', filePath);
            
            try {
                await audioEngine.loadTrack(deck, filePath);
                await updateStatus();
            } catch (error) {
                console.error('Load error:', error);
            }
        }
    }

    // 基本操作
    async function play(deck) {
        await audioEngine.play(deck);
    }

    async function pause(deck) {
        await audioEngine.pause(deck);
    }

    async function stop(deck) {
        await audioEngine.stop(deck);
    }

    async function setVolume(deck, volume) {
        await audioEngine.setVolume(deck, volume);
    }

    async function seek(deck, position) {
        await audioEngine.seek(deck, position);
    }

    // 新機能
    async function setEQ(deck, eq) {
        await audioEngine.setEQ(deck, eq);
    }

    async function setFilter(deck, filter) {
        await audioEngine.setFilter(deck, filter);
    }

    async function setSpeed(deck, speed) {
        await audioEngine.setSpeed(deck, speed);
    }

    async function setCrossfader(value) {
        await audioEngine.setCrossfader(value);
    }

    async function setMasterVolume(volume) {
        await audioEngine.setMasterVolume(volume);
    }

    async function addCuePoint(deck, name, color) {
        await audioEngine.addCuePoint(deck, name, color);
        await updateStatus();
    }

    async function setLoop(deck, start, end) {
        await audioEngine.setLoop(deck, start, end);
        await updateStatus();
    }

    async function enableSync(enabled, master) {
        await audioEngine.enableSync(enabled, master);
        await updateStatus();
    }
</script>

<main class="dj-pro">
    <!-- ヘッダー -->
    <header class="header">
        <div class="logo">
            <div class="logo-icon">🎧</div>
            <h1>DJ MIXER PRO</h1>
        </div>
        
        <div class="status-bar">
            <div class="connection-status {connected ? 'connected' : 'disconnected'}">
                <span class="status-dot"></span>
                <span>{connected ? 'CONNECTED' : 'DISCONNECTED'}</span>
            </div>
            
            {#if mixerStatus.syncEnabled}
                <div class="sync-indicator">
                    <span class="sync-icon">🔗</span>
                    SYNC: Deck {mixerStatus.syncMaster.toUpperCase()}
                </div>
            {/if}
            
            <button class="settings-btn" on:click={() => showSettings = !showSettings}>
                ⚙️
            </button>
        </div>
    </header>

    {#if !connected}
        <div class="connection-error">
            <div class="error-icon">⚠️</div>
            <h2>Audio Engine Offline</h2>
            <p>Start the Go audio engine to begin</p>
            <code>cd go_audio_engine && .\audio_engine.exe</code>
            <button on:click={() => window.location.reload()} class="retry-btn">
                🔄 RECONNECT
            </button>
        </div>
    {:else}
        <!-- メインDJレイアウト -->
        <div class="dj-layout">
            <!-- Deck A -->
            <div class="deck-section deck-a">
                <ProDeck 
                    deck="a"
                    deckInfo={mixerStatus.deckA}
                    onLoadTrack={() => loadTrack('a')}
                    onPlay={() => play('a')}
                    onPause={() => pause('a')}
                    onStop={() => stop('a')}
                    onVolumeChange={(vol) => setVolume('a', vol)}
                    onSeek={(pos) => seek('a', pos)}
                    onEQChange={(eq) => setEQ('a', eq)}
                    onFilterChange={(filter) => setFilter('a', filter)}
                    onSpeedChange={(speed) => setSpeed('a', speed)}
                    onAddCuePoint={(name, color) => addCuePoint('a', name, color)}
                    onSetLoop={(start, end) => setLoop('a', start, end)}
                />
            </div>

            <!-- Center Mixer -->
            <div class="mixer-section">
                <ProMixer 
                    mixerStatus={mixerStatus}
                    onCrossfaderChange={setCrossfader}
                    onMasterVolumeChange={setMasterVolume}
                    onSyncToggle={(enabled, master) => enableSync(enabled, master)}
                />
            </div>

            <!-- Deck B -->
            <div class="deck-section deck-b">
                <ProDeck 
                    deck="b"
                    deckInfo={mixerStatus.deckB}
                    onLoadTrack={() => loadTrack('b')}
                    onPlay={() => play('b')}
                    onPause={() => pause('b')}
                    onStop={() => stop('b')}
                    onVolumeChange={(vol) => setVolume('b', vol)}
                    onSeek={(pos) => seek('b', pos)}
                    onEQChange={(eq) => setEQ('b', eq)}
                    onFilterChange={(filter) => setFilter('b', filter)}
                    onSpeedChange={(speed) => setSpeed('b', speed)}
                    onAddCuePoint={(name, color) => addCuePoint('b', name, color)}
                    onSetLoop={(start, end) => setLoop('b', start, end)}
                />
            </div>
        </div>
    {/if}
</main>

<style>
    :global(body) {
        margin: 0;
        padding: 0;
        font-family: 'Inter', -apple-system, BlinkMacSystemFont, sans-serif;
        background: #000;
        color: #fff;
        overflow: hidden;
    }

    .dj-pro {
        width: 100vw;
        height: 100vh;
        display: flex;
        flex-direction: column;
        background: 
            radial-gradient(ellipse at top, #1a1a2e 0%, #000 50%),
            radial-gradient(ellipse at bottom, #0f0f23 0%, #000 50%);
        position: relative;
    }

    /* アニメーション背景 */
    .dj-pro::before {
        content: '';
        position: absolute;
        top: 0;
        left: 0;
        right: 0;
        bottom: 0;
        background: 
            linear-gradient(45deg, transparent 30%, rgba(0, 255, 136, 0.03) 50%, transparent 70%),
            linear-gradient(-45deg, transparent 30%, rgba(255, 0, 136, 0.03) 50%, transparent 70%);
        animation: gradient 15s ease infinite;
        pointer-events: none;
    }

    @keyframes gradient {
        0%, 100% { opacity: 0.3; }
        50% { opacity: 0.6; }
    }

    /* ヘッダー */
    .header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        padding: 15px 30px;
        background: rgba(0, 0, 0, 0.8);
        backdrop-filter: blur(20px);
        border-bottom: 1px solid rgba(0, 255, 136, 0.2);
        z-index: 100;
    }

    .logo {
        display: flex;
        align-items: center;
        gap: 15px;
    }

    .logo-icon {
        font-size: 32px;
        animation: pulse 3s ease-in-out infinite;
    }

    @keyframes pulse {
        0%, 100% { transform: scale(1); }
        50% { transform: scale(1.1); }
    }

    .logo h1 {
        margin: 0;
        font-size: 24px;
        font-weight: 900;
        letter-spacing: 3px;
        background: linear-gradient(90deg, #00ff88, #00aaff);
        -webkit-background-clip: text;
        -webkit-text-fill-color: transparent;
        background-clip: text;
    }

    .status-bar {
        display: flex;
        align-items: center;
        gap: 20px;
    }

    .connection-status {
        display: flex;
        align-items: center;
        gap: 8px;
        padding: 8px 16px;
        border-radius: 20px;
        font-size: 12px;
        font-weight: 700;
        letter-spacing: 1px;
    }

    .connection-status.connected {
        background: rgba(0, 255, 136, 0.1);
        color: #00ff88;
        border: 1px solid #00ff88;
    }

    .connection-status.disconnected {
        background: rgba(255, 0, 136, 0.1);
        color: #ff0088;
        border: 1px solid #ff0088;
    }

    .status-dot {
        width: 8px;
        height: 8px;
        border-radius: 50%;
        background: currentColor;
        animation: blink 2s ease-in-out infinite;
    }

    @keyframes blink {
        0%, 100% { opacity: 1; }
        50% { opacity: 0.3; }
    }

    .sync-indicator {
        display: flex;
        align-items: center;
        gap: 8px;
        padding: 8px 16px;
        background: rgba(0, 170, 255, 0.1);
        border: 1px solid #00aaff;
        border-radius: 20px;
        font-size: 12px;
        font-weight: 700;
        color: #00aaff;
    }

    .settings-btn {
        background: rgba(255, 255, 255, 0.05);
        border: 1px solid rgba(255, 255, 255, 0.1);
        color: #fff;
        padding: 10px 16px;
        border-radius: 8px;
        cursor: pointer;
        font-size: 18px;
        transition: all 0.3s;
    }

    .settings-btn:hover {
        background: rgba(255, 255, 255, 0.1);
        transform: rotate(90deg);
    }

    /* エラー画面 */
    .connection-error {
        flex: 1;
        display: flex;
        flex-direction: column;
        justify-content: center;
        align-items: center;
        gap: 20px;
        text-align: center;
    }

    .error-icon {
        font-size: 64px;
        animation: shake 0.5s ease-in-out infinite;
    }

    @keyframes shake {
        0%, 100% { transform: translateX(0); }
        25% { transform: translateX(-10px); }
        75% { transform: translateX(10px); }
    }

    .connection-error h2 {
        color: #ff0088;
        font-size: 32px;
        margin: 0;
    }

    .connection-error code {
        background: rgba(0, 255, 136, 0.1);
        padding: 15px 30px;
        border-radius: 8px;
        border: 1px solid #00ff88;
        color: #00ff88;
        font-family: 'Courier New', monospace;
    }

    .retry-btn {
        padding: 15px 40px;
        background: linear-gradient(135deg, #ff0088, #ff6666);
        border: none;
        border-radius: 30px;
        color: #fff;
        font-size: 16px;
        font-weight: 700;
        cursor: pointer;
        transition: all 0.3s;
    }

    .retry-btn:hover {
        transform: scale(1.05);
        box-shadow: 0 0 30px rgba(255, 0, 136, 0.5);
    }

    /* メインレイアウト */
    .dj-layout {
        flex: 1;
        display: grid;
        grid-template-columns: 1fr 400px 1fr;
        gap: 20px;
        padding: 20px;
        overflow: hidden;
    }

    .deck-section {
        display: flex;
        flex-direction: column;
        min-height: 0;
    }

    .mixer-section {
        display: flex;
        flex-direction: column;
        min-height: 0;
    }

    @media (max-width: 1400px) {
        .dj-layout {
            grid-template-columns: 1fr;
            grid-template-rows: auto auto auto;
        }
    }
</style>