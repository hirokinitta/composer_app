<script>
  import { onMount } from 'svelte';
  
  export let deckId = 'a'; // 'a' or 'b'
  
  let trackInfo = null;
  let loading = false;
  let playing = false;
  let volume = 100;
  let position = '0:00';
  let duration = '0:00';
  
  let ipcRenderer;
  
  onMount(() => {
    // ElectronのipcRendererを取得
    if (window.require) {
      try {
        const electron = window.require('electron');
        ipcRenderer = electron.ipcRenderer;
        console.log('Electron IPC initialized for deck', deckId);
      } catch (error) {
        console.error('Failed to initialize Electron IPC:', error);
      }
    } else {
      console.warn('Electron require not available - running in browser mode?');
    }
  });
  
  async function loadTrack() {
    if (!ipcRenderer) {
      alert('Electron IPC not available. Make sure you are running in Electron.');
      return;
    }
    
    loading = true;
    try {
      // ファイル選択ダイアログを開く
      const filePath = await ipcRenderer.invoke('select-audio-file');
      
      if (!filePath) {
        console.log('No file selected');
        loading = false;
        return;
      }
      
      console.log('Selected file:', filePath);
      
      // Goバックエンドにファイルパスを送信
      const response = await fetch(
        `http://localhost:8080/api/deck/${deckId}/load?file=${encodeURIComponent(filePath)}`
      );
      
      const data = await response.json();
      
      if (response.ok) {
        trackInfo = data;
        console.log('Track loaded successfully:', data);
      } else {
        console.error('Failed to load track:', data.error);
        alert(`Failed to load track: ${data.error}`);
      }
    } catch (error) {
      console.error('Error loading track:', error);
      alert(`Error: ${error.message}`);
    } finally {
      loading = false;
    }
  }
  
  async function play() {
    if (!trackInfo) {
      alert('Please load a track first');
      return;
    }
    
    try {
      const response = await fetch(`http://localhost:8080/api/deck/${deckId}/play`, {
        method: 'POST'
      });
      const data = await response.json();
      if (response.ok) {
        playing = true;
        console.log('Playing:', data);
      }
    } catch (error) {
      console.error('Play error:', error);
      alert(`Play error: ${error.message}`);
    }
  }
  
  async function pause() {
    try {
      const response = await fetch(`http://localhost:8080/api/deck/${deckId}/pause`, {
        method: 'POST'
      });
      const data = await response.json();
      if (response.ok) {
        playing = false;
        console.log('Paused:', data);
      }
    } catch (error) {
      console.error('Pause error:', error);
    }
  }
  
  async function stop() {
    try {
      const response = await fetch(`http://localhost:8080/api/deck/${deckId}/stop`, {
        method: 'POST'
      });
      const data = await response.json();
      if (response.ok) {
        playing = false;
        console.log('Stopped:', data);
      }
    } catch (error) {
      console.error('Stop error:', error);
    }
  }
  
  async function updateVolume() {
    try {
      const response = await fetch(`http://localhost:8080/api/deck/${deckId}/volume`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ volume: volume / 100 })
      });
      const data = await response.json();
      console.log('Volume updated:', data);
    } catch (error) {
      console.error('Volume error:', error);
    }
  }
</script>

<div class="deck">
  <h2>Deck {deckId.toUpperCase()}</h2>
  
  <div class="track-info">
    {#if trackInfo}
      <p class="track-name">🎵 {trackInfo.file?.split(/[\\/]/).pop() || 'Track loaded'}</p>
    {:else}
      <p class="no-track">No track loaded</p>
    {/if}
    <p class="time">{position} / {duration}</p>
  </div>
  
  <div class="waveform">
    <div class="waveform-placeholder">
      {#if trackInfo}
        <span>🎵 Waveform visualization</span>
      {:else}
        <span>Load a track to see waveform</span>
      {/if}
    </div>
  </div>
  
  <button class="load-btn" on:click={loadTrack} disabled={loading}>
    {loading ? '⏳ Loading...' : '📁 Load Track'}
  </button>
  
  <div class="controls">
    <button class="play-btn" on:click={play} disabled={!trackInfo || playing}>
      ▶ Play
    </button>
    <button class="pause-btn" on:click={pause} disabled={!playing}>
      ⏸ Pause
    </button>
    <button class="stop-btn" on:click={stop} disabled={!trackInfo}>
      ⏹ Stop
    </button>
  </div>
  
  <div class="volume-control">
    <label for="volume-{deckId}">Volume</label>
    <input 
      id="volume-{deckId}"
      type="range" 
      min="0" 
      max="100" 
      bind:value={volume}
      on:input={updateVolume}
    />
    <span>{volume}%</span>
  </div>
</div>

<style>
  .deck {
    background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%);
    border: 2px solid #0f3460;
    border-radius: 12px;
    padding: 20px;
    color: #e94560;
    min-width: 400px;
  }
  
  h2 {
    color: #00d9ff;
    margin: 0 0 15px 0;
    text-align: center;
    font-size: 24px;
  }
  
  .track-info {
    background: rgba(0, 0, 0, 0.3);
    padding: 15px;
    border-radius: 8px;
    margin-bottom: 15px;
    min-height: 70px;
    display: flex;
    flex-direction: column;
    justify-content: center;
  }
  
  .track-info p {
    margin: 5px 0;
    color: #fff;
  }
  
  .track-name {
    font-weight: bold;
    word-break: break-all;
  }
  
  .no-track {
    color: #888 !important;
    font-style: italic;
  }
  
  .time {
    color: #00d9ff !important;
    font-family: 'Courier New', monospace;
    font-size: 18px;
    font-weight: bold;
  }
  
  .waveform {
    background: rgba(0, 0, 0, 0.5);
    height: 100px;
    border-radius: 8px;
    margin-bottom: 15px;
    display: flex;
    align-items: center;
    justify-content: center;
    border: 1px solid rgba(0, 217, 255, 0.3);
  }
  
  .waveform-placeholder {
    color: #00d9ff;
    font-size: 14px;
    opacity: 0.6;
  }
  
  .load-btn {
    width: 100%;
    padding: 12px;
    background: linear-gradient(135deg, #ffd700 0%, #ffed4e 100%);
    border: none;
    border-radius: 8px;
    font-size: 16px;
    font-weight: bold;
    cursor: pointer;
    margin-bottom: 15px;
    transition: all 0.3s ease;
  }
  
  .load-btn:hover:not(:disabled) {
    background: linear-gradient(135deg, #ffed4e 0%, #ffd700 100%);
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(255, 215, 0, 0.4);
  }
  
  .load-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
    transform: none;
  }
  
  .controls {
    display: flex;
    gap: 10px;
    margin-bottom: 15px;
  }
  
  .controls button {
    flex: 1;
    padding: 12px;
    border: none;
    border-radius: 8px;
    font-size: 14px;
    font-weight: bold;
    cursor: pointer;
    transition: all 0.2s ease;
  }
  
  .play-btn {
    background: linear-gradient(135deg, #00ff88 0%, #00cc6a 100%);
    color: #000;
  }
  
  .pause-btn {
    background: linear-gradient(135deg, #ff8c00 0%, #ff6600 100%);
    color: #fff;
  }
  
  .stop-btn {
    background: linear-gradient(135deg, #ff0055 0%, #cc0044 100%);
    color: white;
  }
  
  .controls button:hover:not(:disabled) {
    opacity: 0.9;
    transform: translateY(-2px);
  }
  
  .controls button:disabled {
    opacity: 0.3;
    cursor: not-allowed;
    transform: none;
  }
  
  .volume-control {
    display: flex;
    align-items: center;
    gap: 10px;
  }
  
  .volume-control label {
    color: #fff;
    font-weight: bold;
    min-width: 60px;
  }
  
  .volume-control input[type="range"] {
    flex: 1;
    height: 6px;
    border-radius: 3px;
    background: rgba(255, 255, 255, 0.2);
    outline: none;
    -webkit-appearance: none;
  }
  
  .volume-control input[type="range"]::-webkit-slider-thumb {
    -webkit-appearance: none;
    appearance: none;
    width: 18px;
    height: 18px;
    border-radius: 50%;
    background: #00ff88;
    cursor: pointer;
    box-shadow: 0 0 8px rgba(0, 255, 136, 0.5);
  }
  
  .volume-control input[type="range"]::-moz-range-thumb {
    width: 18px;
    height: 18px;
    border-radius: 50%;
    background: #00ff88;
    cursor: pointer;
    border: none;
    box-shadow: 0 0 8px rgba(0, 255, 136, 0.5);
  }
  
  .volume-control span {
    color: #00d9ff;
    font-weight: bold;
    min-width: 45px;
    text-align: right;
    font-family: 'Courier New', monospace;
  }
</style>