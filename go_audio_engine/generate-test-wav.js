// generate-test-wav.js
const fs = require('fs');
const path = require('path');
const { WaveFile } = require('wavefile');

// testdataディレクトリを作成
const testdataDir = path.join(__dirname, 'testdata');
if (!fs.existsSync(testdataDir)) {
    fs.mkdirSync(testdataDir);
}

function generateTone(filename, frequency, duration, sampleRate = 44100) {
    const numSamples = Math.floor(sampleRate * duration);
    const samples = new Int16Array(numSamples * 2); // ステレオ
    
    for (let i = 0; i < numSamples; i++) {
        const t = i / sampleRate;
        const value = Math.floor(32767 * 0.3 * Math.sin(2 * Math.PI * frequency * t));
        samples[i * 2] = value;     // 左チャンネル
        samples[i * 2 + 1] = value; // 右チャンネル
    }
    
    const wav = new WaveFile();
    wav.fromScratch(2, sampleRate, '16', samples);
    
    const filepath = path.join(testdataDir, filename);
    fs.writeFileSync(filepath, wav.toBuffer());
    console.log(`✅ Generated: ${filepath}`);
}

console.log('🎵 Generating test WAV files...\n');

// テスト用の音を生成
generateTone('tone_440hz.wav', 440, 5);  // A4音、5秒
generateTone('tone_523hz.wav', 523, 5);  // C5音、5秒
generateTone('tone_261hz.wav', 261, 5);  // C4音、5秒

console.log('\n✅ All test files created in testdata/');