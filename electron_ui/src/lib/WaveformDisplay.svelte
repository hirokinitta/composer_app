<script>
    import { onMount } from 'svelte';
    
    export let audioData = [];
    export let position = 0;
    export let duration = 0;
    export let color = '#00ff88';
    
    let canvas;
    let ctx;
    
    $: if (canvas && audioData.length > 0) {
        drawWaveform();
    }
    
    onMount(() => {
        if (canvas) {
            ctx = canvas.getContext('2d');
            drawWaveform();
        }
    });
    
    function drawWaveform() {
        if (!ctx || !canvas) return;
        
        const width = canvas.width;
        const height = canvas.height;
        
        // キャンバスをクリア
        ctx.fillStyle = '#000';
        ctx.fillRect(0, 0, width, height);
        
        if (audioData.length === 0) {
            // データがない場合は中央線のみ
            ctx.strokeStyle = '#333';
            ctx.beginPath();
            ctx.moveTo(0, height / 2);
            ctx.lineTo(width, height / 2);
            ctx.stroke();
            return;
        }
        
        // 波形を描画
        ctx.strokeStyle = color;
        ctx.lineWidth = 2;
        ctx.beginPath();
        
        const step = audioData.length / width;
        
        for (let x = 0; x < width; x++) {
            const index = Math.floor(x * step);
            const value = audioData[index] || 0;
            const y = (height / 2) - (value * height / 2);
            
            if (x === 0) {
                ctx.moveTo(x, y);
            } else {
                ctx.lineTo(x, y);
            }
        }
        
        ctx.stroke();
        
        // 再生位置を描画
        if (duration > 0) {
            const playheadX = (position / duration) * width;
            ctx.strokeStyle = '#fff';
            ctx.lineWidth = 2;
            ctx.beginPath();
            ctx.moveTo(playheadX, 0);
            ctx.lineTo(playheadX, height);
            ctx.stroke();
        }
    }
</script>

<canvas 
    bind:this={canvas}
    width={800}
    height={120}
    class="waveform-canvas"
></canvas>

<style>
    .waveform-canvas {
        width: 100%;
        height: 100%;
        border-radius: 8px;
    }
</style>