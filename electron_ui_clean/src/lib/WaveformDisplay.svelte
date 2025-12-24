<script>
    import { onMount } from 'svelte';

    export let audioData = []; // Normalized data (-1.0 to 1.0)
    export let position = 0;   // Current time in seconds
    export let duration = 0;   // Total duration in seconds
    export let color = '#00ff88';

    let canvas;
    let ctx;
    let width, height;

    // 💡 Reactive statement that explicitly watches all dependencies
    // This ensures the canvas redraws when position, data, or dimensions change.
    $: if (ctx && width && height && (audioData || position || duration)) {
        drawFrame();
    }

    onMount(() => {
        if (canvas) {
            ctx = canvas.getContext('2d');
            // Initial draw
            handleResize(); 
        }

        // Handle window resizing to keep waveform sharp
        const resizeObserver = new ResizeObserver(() => handleResize());
        resizeObserver.observe(canvas);
        
        return () => resizeObserver.disconnect();
    });

    function handleResize() {
        if (!canvas) return;
        
        // Get the CSS size
        const rect = canvas.getBoundingClientRect();
        
        // Adjust for High-DPI (Retina) displays
        const dpr = window.devicePixelRatio || 1;
        
        canvas.width = rect.width * dpr;
        canvas.height = rect.height * dpr;
        
        // Scale context to match
        ctx.scale(dpr, dpr);
        
        // Store logical width/height for drawing calculations
        width = rect.width;
        height = rect.height;
        
        drawFrame();
    }

    function drawFrame() {
        if (!ctx || !width || !height) return;

        // Clear Canvas
        ctx.clearRect(0, 0, width, height);
        ctx.fillStyle = '#111'; // Dark background
        ctx.fillRect(0, 0, width, height);

        // Draw Center Line
        ctx.beginPath();
        ctx.strokeStyle = '#333';
        ctx.lineWidth = 1;
        ctx.moveTo(0, height / 2);
        ctx.lineTo(width, height / 2);
        ctx.stroke();

        if (!audioData || audioData.length === 0) return;

        // --- 🌊 Waveform Drawing (Peak Detection Algorithm) ---
        ctx.beginPath();
        ctx.strokeStyle = color;
        ctx.lineWidth = 1.5;

        // Number of audio samples per canvas pixel
        const step = Math.ceil(audioData.length / width);
        const amp = height / 2;

        for (let x = 0; x < width; x++) {
            // Calculate the chunk of audio data for this pixel
            const startIdx = x * step;
            let min = 1.0;
            let max = -1.0;

            // Find min/max in this chunk (Peak detection)
            for (let i = 0; i < step; i++) {
                const datum = audioData[startIdx + i];
                if (datum < min) min = datum;
                if (datum > max) max = datum;
            }

            // Fallback if data is missing
            if (max === -1.0) max = 0;
            if (min === 1.0) min = 0;

            // Draw vertical line for this pixel (Mirror style)
            // (Converts -1..1 range to canvas Y coordinates)
            ctx.moveTo(x, (1 + min) * amp);
            ctx.lineTo(x, (1 + max) * amp);
        }
        ctx.stroke();

        // --- 📍 Playhead Drawing ---
        if (duration > 0) {
            const playheadX = (position / duration) * width;
            
            // Draw Playhead Line
            ctx.beginPath();
            ctx.strokeStyle = '#ffffff';
            ctx.lineWidth = 2;
            ctx.moveTo(playheadX, 0);
            ctx.lineTo(playheadX, height);
            ctx.stroke();

            // Optional: Draw Playhead "Cap" (Triangle or Circle)
            ctx.fillStyle = '#ffffff';
            ctx.beginPath();
            ctx.moveTo(playheadX - 4, 0);
            ctx.lineTo(playheadX + 4, 0);
            ctx.lineTo(playheadX, 6);
            ctx.fill();
        }
    }
</script>

<canvas bind:this={canvas} class="waveform-canvas"></canvas>

<style>
    .waveform-canvas {
        width: 100%;
        height: 100%;
        display: block; /* Removes bottom spacing in some browsers */
        background: #000;
        border-radius: 4px;
    }
</style>