<script>
    // 💡 デバッグ: スクリプト実行確認 (一番最初にログを出す)
    console.log('📜 App.svelte: Script executing...');

    import { onMount, onDestroy } from 'svelte';
    import { startWebSocketConnection, subscribeToMixerStatus, subscribeToWsStatus, getMixerData } from './api/websocket.js';
    import { AudioEngineClient } from './lib/audioEngineClient.js';

    // リアクティブ変数 (初期値)
    let deckAInfo = { playing: false, position: 0, duration: 0 };
    let deckBInfo = { playing: false, position: 0, duration: 0 };
    let mixerControls = { masterVolume: 1, crossfader: 0.5, syncEnabled: false, syncMaster: 'a' };

    let client = null;
    let isAppReady = false;
    let currentWsStatus = 'disconnected';

    // 💡 修正: onMountを待たずに、スクリプト実行時に即座に購読を開始する
    // これにより、コンポーネントが作成された瞬間にデータ監視が始まります
    let unsubMixer = () => {};
    let unsubWs = () => {};

    try {
        // グローバルストアを直接購読
        unsubMixer = subscribeToMixerStatus(data => {
            console.log('⚡ App.svelte: Mixer update:', data ? 'DATA' : 'NULL');
            if (data) {
                isAppReady = true;
                
                // データをローカル変数に反映
                deckAInfo = data.DeckA || deckAInfo;
                deckBInfo = data.DeckB || deckBInfo;
                const srcMixer = data.Mixer || data;
                mixerControls = {
                    masterVolume: srcMixer.masterVolume ?? srcMixer.MasterVolume ?? 1.0,
                    crossfader: srcMixer.crossfader ?? srcMixer.Crossfader ?? 0.0,
                    syncEnabled: srcMixer.syncEnabled ?? srcMixer.SyncEnabled ?? false,
                    syncMaster: srcMixer.syncMaster ?? srcMixer.SyncMaster ?? 'a'
                };

                // データが来たら強制的に接続済みにする(表示上)
                if (currentWsStatus !== 'connected') {
                    currentWsStatus = 'connected';
                }
            }
        });

        unsubWs = subscribeToWsStatus(status => {
            console.log('⚡ App.svelte: WS status update:', status);
            if (isAppReady && status !== 'connected') {
                 currentWsStatus = 'connected';
            } else {
                currentWsStatus = status;
            }
        });
    } catch (e) {
        console.error('❌ Error setting up subscriptions:', e);
    }

    // 💡 修正: onMountを待たずに即座に接続を開始する
    console.log('🚀 App.svelte: calling startWebSocketConnection immediately...');
    startWebSocketConnection();

    // コンポーネント破棄時に購読解除
    onDestroy(() => {
        console.log('💥 App.svelte: onDestroy called');
        if (unsubMixer) unsubMixer();
        if (unsubWs) unsubWs();
    });

    // アプリ起動時の処理
    onMount(() => {
        console.log('🚀 App.svelte: onMount START');
        try {
            try {
                client = new AudioEngineClient();
            } catch (err) {
                console.error('❌ Failed to initialize AudioEngineClient:', err);
            }
        } catch (e) {
            console.error(`❌ Error in App.svelte onMount:`, e);
        }
    });

    // --- API呼び出し用の関数 ---
    async function apiRequest(endpoint, body = {}) {
        try {
            await fetch(`http://localhost:8080/api/${endpoint}`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(body)
            });
        } catch (e) { console.error(e); }
    }
    
    // ProMixer用の関数定義
    const setCrossfader = (v) => apiRequest('mixer/crossfader', { value: v });
    const setMasterVolume = (v) => apiRequest('mixer/master', { volume: v });
    const enableSync = (e, m) => apiRequest('mixer/sync', { enabled: e, master: m.toLowerCase() });
    const setGain = (d, v) => apiRequest(`deck/${d.toLowerCase()}/gain`, { value: v });
    const toggleCue = (d, e) => apiRequest(`deck/${d.toLowerCase()}/cue`, { enabled: e });

    function handleReconnect() {
        console.log('🔄 Force Connect clicked');
        startWebSocketConnection();
        
        // 💡 追加: 購読がうまくいかない場合でも、ボタンを押したら強制的にデータを取得して画面を更新する
        setTimeout(() => {
            const data = getMixerData();
            if (data) {
                console.log('🔄 Force Connect: Manual data sync success', data);
                // 手動で状態を更新
                deckAInfo = data.DeckA || deckAInfo;
                deckBInfo = data.DeckB || deckBInfo;
                const srcMixer = data.Mixer || data;
                mixerControls = {
                    masterVolume: srcMixer.masterVolume ?? srcMixer.MasterVolume ?? 1.0,
                    crossfader: srcMixer.crossfader ?? srcMixer.Crossfader ?? 0.0,
                    syncEnabled: srcMixer.syncEnabled ?? srcMixer.SyncEnabled ?? false,
                    syncMaster: srcMixer.syncMaster ?? srcMixer.SyncMaster ?? 'a'
                };
                
                isAppReady = true;
                currentWsStatus = 'connected';
            }
        }, 500); // 接続処理待ちとして0.5秒確保
    }
</script>

<main class="dj-pro">
    <header class="header">
        <div class="logo">
            <div class="logo-icon">🎧</div>
            <h1>DJ MIXER PRO</h1>
        </div>
        
        <div class="status-bar">
            <div class="connection-status {currentWsStatus}">
                <span class="status-dot"></span>
                <span>{currentWsStatus.toUpperCase()}</span> 
            </div>
            <button class="settings-btn" on:click={handleReconnect}>🔄</button>
        </div>
    </header>

    <!-- メインの表示切り替え: $mixerStatus にデータがあるかどうかだけで判断する -->
    {#if isAppReady}
        <div class="dj-layout">
            <div class="deck-section deck-a">
                <ProDeck 
                    deckName="A"
                    client={client}
                    deckInfo={deckAInfo} 
                />
            </div>

            <div class="mixer-section">
                <ProMixer 
                    mixerStatus={mixerControls}
                    onCrossfaderChange={setCrossfader}
                    onMasterVolumeChange={setMasterVolume}
                    onSyncToggle={enableSync}
                    onGainChange={setGain}
                    onCueToggle={toggleCue}
                />
            </div>

            <div class="deck-section deck-b">
                <ProDeck 
                    deckName="B"
                    client={client}
                    deckInfo={deckBInfo} 
                />
            </div>
        </div>
    {:else}
        <div class="connection-error">
            <div class="error-icon">⏳</div>
            <h2>Connecting...</h2>
            <p>Waiting for Audio Engine...</p>
            <button on:click={handleReconnect} class="retry-btn">
                FORCE CONNECT
            </button>
        </div>
    {/if}
</main>

<style>
/* 提供されたスタイルを維持 */
:global(body) {
    margin: 0;
    padding: 0;
    font-family: 'Inter', -apple-system, BlinkMacSystemFont, sans-serif;
    background: #000;
    color: #fff;
    overflow: hidden;
}

:global(#app) {
    width: 100%;
    height: 100vh;
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
.connection-status.connecting {
    background: rgba(255, 255, 0, 0.1);
    color: #ffff00;
    border: 1px solid #ffff00;
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

.debug-info {
    font-family: monospace;
    background: rgba(0,0,0,0.3);
    padding: 10px;
    border-radius: 8px;
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
    margin-top: 20px;
}

.retry-btn:hover {
    transform: scale(1.05);
    box-shadow: 0 0 30px rgba(255, 0, 136, 0.5);
}

.dj-layout {
    flex: 1;
    display: grid;
    grid-template-columns: 1fr 400px 1fr;
    gap: 20px;
    padding: 20px;
    overflow: hidden;
}

.deck-section, .mixer-section {
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
