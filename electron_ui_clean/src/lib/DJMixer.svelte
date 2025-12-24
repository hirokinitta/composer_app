<script>
    export let crossfaderValue = 0;
    export let masterVolume = 1.0;
    export let onCrossfaderChange = () => {};
    export let onMasterVolumeChange = () => {};

    function handleCrossfaderChange(e) {
        onCrossfaderChange(parseFloat(e.target.value));
    }

    function handleMasterVolumeChange(e) {
        onMasterVolumeChange(parseFloat(e.target.value));
    }

    $: crossfaderLabel = crossfaderValue < -0.3 ? 'A' : 
                         crossfaderValue > 0.3 ? 'B' : 
                         'CENTER';
</script>

<div class="mixer">
    <h2>🎚️ Mixer</h2>
    
    <div class="crossfader-section">
        <!-- svelte-ignore a11y_label_has_associated_control -->
        <label>Crossfader</label>
        <div class="crossfader-container">
            <span class="deck-label">A</span>
            <div class="crossfader-track">
                <input 
                    type="range" 
                    min="-1" 
                    max="1" 
                    step="0.01"
                    value={crossfaderValue}
                    on:input={handleCrossfaderChange}
                    class="crossfader-slider"
                />
                <div class="crossfader-indicator" style="left: {(crossfaderValue + 1) * 50}%"></div>
            </div>
            <span class="deck-label">B</span>
        </div>
        <div class="crossfader-value">{crossfaderLabel}</div>
    </div>

    <div class="master-section">
        <!-- svelte-ignore a11y_label_has_associated_control -->
        <label>Master Volume</label>
        <div class="master-volume-container">
            <input 
                type="range" 
                min="0" 
                max="1" 
                step="0.01"
                value={masterVolume}
                on:input={handleMasterVolumeChange}
                class="master-slider"
                orient="vertical"
            />
            <div class="volume-bar">
                <div class="volume-fill" style="height: {masterVolume * 100}%"></div>
            </div>
        </div>
        <div class="master-value">{Math.round(masterVolume * 100)}%</div>
    </div>
</div>

<style>
    .mixer {
        background: linear-gradient(145deg, #1a1a1a, #0a0a0a);
        border: 2px solid #444;
        border-radius: 12px;
        padding: 25px;
        box-shadow: 0 6px 30px rgba(0, 0, 0, 0.7);
        min-width: 400px;
    }

    .mixer h2 {
        margin: 0 0 20px 0;
        color: #fff;
        font-size: 24px;
        text-align: center;
    }

    .crossfader-section {
        margin-bottom: 30px;
    }

    .crossfader-section label {
        display: block;
        color: #aaa;
        font-size: 14px;
        margin-bottom: 10px;
        text-align: center;
    }

    .crossfader-container {
        display: flex;
        align-items: center;
        gap: 15px;
    }

    .deck-label {
        color: #fff;
        font-weight: bold;
        font-size: 18px;
        min-width: 20px;
        text-align: center;
    }

    .crossfader-track {
        flex: 1;
        position: relative;
        height: 50px;
        background: #222;
        border-radius: 25px;
        padding: 5px;
    }

    .crossfader-slider {
        position: absolute;
        width: 100%;
        height: 100%;
        top: 0;
        left: 0;
        opacity: 0;
        cursor: pointer;
        z-index: 2;
    }

    .crossfader-indicator {
        position: absolute;
        top: 50%;
        transform: translate(-50%, -50%);
        width: 40px;
        height: 40px;
        background: linear-gradient(145deg, #00ff88, #00aa55);
        border-radius: 50%;
        box-shadow: 0 2px 10px rgba(0, 255, 136, 0.5);
        transition: left 0.1s ease-out;
        pointer-events: none;
    }

    .crossfader-value {
        text-align: center;
        color: #00ff88;
        font-weight: bold;
        font-size: 16px;
        margin-top: 10px;
        font-family: 'Courier New', monospace;
    }

    .master-section {
        display: flex;
        flex-direction: column;
        align-items: center;
    }

    .master-section label {
        color: #aaa;
        font-size: 14px;
        margin-bottom: 15px;
    }

    .master-volume-container {
        position: relative;
        height: 200px;
        display: flex;
        justify-content: center;
        align-items: center;
    }

    .volume-bar {
        width: 50px;
        height: 200px;
        background: #222;
        border-radius: 25px;
        position: relative;
        overflow: hidden;
    }

    .volume-fill {
        position: absolute;
        bottom: 0;
        left: 0;
        width: 100%;
        background: linear-gradient(to top, #ff0088, #ff6666);
        transition: height 0.1s ease-out;
    }

    .master-slider {
        position: absolute;
        width: 200px;
        height: 50px;
        transform: rotate(-90deg);
        opacity: 0;
        cursor: pointer;
        z-index: 2;
    }

    .master-value {
        margin-top: 15px;
        color: #ff0088;
        font-weight: bold;
        font-size: 20px;
        font-family: 'Courier New', monospace;
    }
</style>