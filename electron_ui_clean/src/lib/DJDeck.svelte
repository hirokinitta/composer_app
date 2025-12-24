<script>
    export let deck = 'a';
    export let deckInfo = {
        file: '',
        playing: false,
        position: 0,
        duration: 0,
        volume: 1.0
    };
    export let onPlay = () => {};
    export let onPause = () => {};
    export let onStop = () => {};
    export let onVolumeChange = () => {};
    export let onSeek = () => {};
    export let onLoadTrack = () => {};

    $: progress = deckInfo.duration > 0 ? (deckInfo.position / deckInfo.duration) * 100 : 0;
    $: fileName = deckInfo.file ? deckInfo.file.split(/[/\\]/).pop() : 'No track loaded';

    function formatTime(seconds) {
        const mins = Math.floor(seconds / 60);
        const secs = Math.floor(seconds % 60);
        return `${mins}:${secs.toString().padStart(2, '0')}`;
    }

    function handleVolumeChange(e) {
        onVolumeChange(parseFloat(e.target.value));
    }

    function handleSeek(e) {
        const newPosition = (e.target.value / 100) * deckInfo.duration;
        onSeek(newPosition);
    }
</script>

<div class="deck deck-{deck}">
    <div class="deck-header">
        <h2>Deck {deck.toUpperCase()}</h2>
        <div class="track-info">
            <div class="file-name" title={deckInfo.file}>{fileName}</div>
            <div class="time-display">
                {formatTime(deckInfo.position)} / {formatTime(deckInfo.duration)}
            </div>
        </div>
    </div>

    <div class="waveform">
        <div class="progress-bar">
            <div class="progress-fill" style="width: {progress}%"></div>
        </div>
        <input 
            type="range" 
            min="0" 
            max="100" 
            value={progress}
            on:change={handleSeek}
            class="seek-bar"
        />
    </div>

    <div class="controls">
        <button on:click={onLoadTrack} class="btn-load">
            📁 Load Track
        </button>
        
        <div class="transport-controls">
            <button on:click={onPlay} disabled={deckInfo.playing} class="btn-play">
                ▶️ Play
            </button>
            <button on:click={onPause} disabled={!deckInfo.playing} class="btn-pause">
                ⏸️ Pause
            </button>
            <button on:click={onStop} class="btn-stop">
                ⏹️ Stop
            </button>
        </div>
    </div>

    <div class="volume-control">
        <label>Volume</label>
        <input 
            type="range" 
            min="0" 
            max="1" 
            step="0.01"
            value={deckInfo.volume}
            on:input={handleVolumeChange}
            class="volume-slider"
        />
        <span class="volume-value">{Math.round(deckInfo.volume * 100)}%</span>
    </div>
</div>

<style>
    .deck {
        background: linear-gradient(145deg, #2a2a2a, #1a1a1a);
        border-radius: 12px;
        padding: 20px;
        box-shadow: 0 4px 20px rgba(0, 0, 0, 0.5);
        min-width: 350px;
    }

    .deck-a {
        border: 2px solid #00ff88;
    }

    .deck-b {
        border: 2px solid #ff0088;
    }

    .deck-header {
        margin-bottom: 15px;
    }

    .deck-header h2 {
        margin: 0 0 10px 0;
        color: #fff;
        font-size: 24px;
        font-weight: bold;
    }

    .track-info {
        display: flex;
        justify-content: space-between;
        align-items: center;
        gap: 10px;
    }

    .file-name {
        flex: 1;
        color: #aaa;
        font-size: 14px;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
    }

    .time-display {
        color: #00ff88;
        font-family: 'Courier New', monospace;
        font-size: 14px;
        font-weight: bold;
    }

    .waveform {
        position: relative;
        margin: 20px 0;
    }

    .progress-bar {
        height: 60px;
        background: #000;
        border-radius: 4px;
        overflow: hidden;
        position: relative;
    }

    .progress-fill {
        height: 100%;
        background: linear-gradient(90deg, #00ff88, #00aa55);
        transition: width 0.1s linear;
    }

    .seek-bar {
        position: absolute;
        top: 0;
        left: 0;
        width: 100%;
        height: 60px;
        opacity: 0;
        cursor: pointer;
    }

    .controls {
        display: flex;
        flex-direction: column;
        gap: 10px;
        margin: 20px 0;
    }

    .transport-controls {
        display: flex;
        gap: 10px;
    }

    button {
        padding: 12px 20px;
        border: none;
        border-radius: 6px;
        font-size: 14px;
        font-weight: bold;
        cursor: pointer;
        transition: all 0.2s;
    }

    .btn-load {
        background: #444;
        color: #fff;
    }

    .btn-load:hover {
        background: #555;
    }

    .btn-play {
        flex: 1;
        background: #00ff88;
        color: #000;
    }

    .btn-play:hover:not(:disabled) {
        background: #00cc6a;
    }

    .btn-pause {
        flex: 1;
        background: #ffaa00;
        color: #000;
    }

    .btn-pause:hover:not(:disabled) {
        background: #cc8800;
    }

    .btn-stop {
        flex: 1;
        background: #ff0088;
        color: #fff;
    }

    .btn-stop:hover {
        background: #cc0066;
    }

    button:disabled {
        opacity: 0.5;
        cursor: not-allowed;
    }

    .volume-control {
        display: flex;
        align-items: center;
        gap: 10px;
    }

    .volume-control label {
        color: #aaa;
        font-size: 14px;
    }

    .volume-slider {
        flex: 1;
        height: 8px;
        border-radius: 4px;
        outline: none;
        background: #333;
    }

    .volume-slider::-webkit-slider-thumb {
        width: 20px;
        height: 20px;
        border-radius: 50%;
        background: #00ff88;
        cursor: pointer;
    }

    .volume-value {
        color: #00ff88;
        font-family: 'Courier New', monospace;
        font-weight: bold;
        min-width: 45px;
        text-align: right;
    }
</style>