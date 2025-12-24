// scripts/generate-test-audio.js
import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';
// 修正: デフォルトエクスポートとしてインポート
import pkg from 'wavefile'; 
const { WaveFile } = pkg;

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const outputDir = path.join(__dirname, '../../go_audio_engine/testdata');

if (!fs.existsSync(outputDir)) {
    fs.mkdirSync(outputDir, { recursive: true });
}

function generateTone(filename, frequency, duration, sampleRate = 44100) {
    const numSamples = Math.floor(sampleRate * duration);
    const samples = new Int16Array(numSamples * 2);
    
    for (let i = 0; i < numSamples; i++) {
        const t = i / sampleRate;
        const value = Math.floor(32767 * 0.3 * Math.sin(2 * Math.PI * frequency * t));
        samples[i * 2] = value;
        samples[i * 2 + 1] = value;
    }
    
    const wav = new WaveFile();
    wav.fromScratch(2, sampleRate, '16', samples);
    
    const filepath = path.join(outputDir, filename);
    fs.writeFileSync(filepath, wav.toBuffer());
    console.log(`✅ ${filename}`);
}

console.log('🎵 Generating test audio files...\n');

generateTone('tone_440hz.wav', 440, 5);
generateTone('tone_523hz.wav', 523, 5);
generateTone('tone_261hz.wav', 261, 5);

console.log('\n✅ Test files created in go_audio_engine/testdata/');