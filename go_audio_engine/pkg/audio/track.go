package audio

import (
	"fmt"
	"os"
	"sync"

	"github.com/go-audio/wav"
)

// Track は拡張されたオーディオトラック
// 全ての機能を統合
type Track struct {
	// 基本情報
	FilePath      string
	SampleRate    int
	Channels      int
	Data          []float32
	Position      int     // 廃止予定だが、互換性のために残す
	floatPosition float64 // 💡 追加: 正確な再生位置
	IsPlaying     bool
	Volume        float64

	// DJ機能
	Speed float64 // ピッチコントロール（0.5 ～ 2.0）

	// エフェクト
	EQ         *ThreeBandEQ     // イコライザー
	Filter     *Filter          // フィルター
	BPM        *BPMDetector     // BPM検出器
	CueManager *CuePointManager // キューポイント管理

	// 同期制御（並行処理の安全性）
	mu sync.RWMutex // RWMutex: 読み書きロック
}

// NewTrack は新しいトラックを作成
func NewTrack(sampleRate int) *Track {
	return &Track{
		Volume:     1.0,
		Speed:      1.0,
		SampleRate: sampleRate,
		EQ:         NewThreeBandEQ(float64(sampleRate)),
		Filter:     NewFilter(float64(sampleRate)),
		BPM:        NewBPMDetector(sampleRate),
		CueManager: NewCuePointManager(),
	}
}

// LoadWAV はWAVファイルをロード
func (t *Track) LoadWAV(filePath string) error {
	// 1. ファイルをオープン（ロックの外）
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	decoder := wav.NewDecoder(file)
	if !decoder.IsValidFile() {
		return fmt.Errorf("invalid WAV file")
	}

	// 基本情報の取得
	sampleRate := int(decoder.SampleRate)
	channels := int(decoder.NumChans)

	fmt.Printf("⏳ Loading WAV: %s (SR:%d, Ch:%d)\n", filePath, sampleRate, channels)

	// 2. デコード実行（重い処理・ロックの外）
	buf, err := decoder.FullPCMBuffer()
	if err != nil {
		return fmt.Errorf("failed to decode audio: %v", err)
	}

	// 3. float32への変換と正規化（重い処理・ロックの外）
	// 💡 正規化 (/ 32768.0) を忘れると、爆音でノイズが発生したり計算負荷が上がります
	convertedData := make([]float32, len(buf.Data))
	for i, sample := range buf.Data {
		convertedData[i] = float32(sample) / 32768.0
	}
	fmt.Printf("⏳ Loading WAV: Conversion completed, samples: %d\n", len(convertedData))

	// 4. データの差し替え（最小限のロック）
	t.mu.Lock()
	// 💡 deferを使わず、必要な代入が終わったらすぐUnlockするのが最も安全です
	t.Data = convertedData
	t.SampleRate = sampleRate
	t.Channels = channels
	t.FilePath = filePath
	t.Position = 0
	t.floatPosition = 0.0 // 💡 追加
	t.IsPlaying = false
	t.mu.Unlock()

	fmt.Printf("✅ Loaded: %s (%.2f seconds)\n", filePath, float64(len(convertedData))/float64(channels)/float64(sampleRate))

	return nil
}

// DetectBPMAsync はBPMを非同期で検出
// goroutineの例：並行処理
func (t *Track) DetectBPMAsync() {
	// 読み取りロック（他の読み取りと並行可能）
	t.mu.RLock()
	data := t.Data
	t.mu.RUnlock()

	// BPM検出（時間がかかる処理）
	bpm := t.BPM.DetectBPM(data)

	fmt.Printf("🎵 BPM detected: %.1f (confidence: %.2f)\n",
		bpm, t.BPM.GetConfidence())
}

// ReadSamples はサンプルを読み取り、エフェクトを適用
func (t *Track) ReadSamples(out []float32) {
	// 💡 最終修正: ループの各イテレーションでロックを取得し、データ競合を完全に防ぐ
	for i := range out {
		t.mu.Lock() // ループ内でロック

		if !t.IsPlaying || len(t.Data) == 0 {
			out[i] = 0
			t.mu.Unlock() // アンロックして次のイテレーションへ
			continue
		}

		currentSampleIndex := int(t.floatPosition)
		if currentSampleIndex >= len(t.Data) {
			// トラック終了
			out[i] = 0
			t.IsPlaying = false
			t.floatPosition = 0.0
		} else {
			// 正常な再生
			out[i] = t.Data[currentSampleIndex] * float32(t.Volume)
			t.floatPosition += t.Speed
		}

		t.mu.Unlock() // ループ内でアンロック
	}

	// エフェクト適用（順番が重要）
	t.Filter.Process(out) // 1. フィルター
	t.EQ.Process(out)     // 2. EQ

	// --- ループチェック ---
	// この処理はロックの外で行う
	t.mu.RLock()
	// ループチェックに必要な値をコピー
	pos := t.floatPosition
	channels := t.Channels
	sampleRate := t.SampleRate
	t.mu.RUnlock()

	// ゼロ除算を防止
	if channels == 0 || sampleRate == 0 {
		return
	}

	// 現在位置を秒に変換
	currentPosInSeconds := pos / float64(channels) / float64(sampleRate)

	// ループをチェック
	shouldLoop, newPos := t.CueManager.CheckLoop(currentPosInSeconds)
	if shouldLoop {
		t.Seek(newPos)
	}
}

// Seek は指定位置にジャンプ
func (t *Track) Seek(seconds float64) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// 💡 修正: floatPosition を更新
	newPosition := seconds * float64(t.SampleRate) * float64(t.Channels)
	if newPosition < 0 {
		newPosition = 0
	}
	if int(newPosition) >= len(t.Data) {
		newPosition = float64(len(t.Data) - 1)
	}
	t.floatPosition = newPosition
}

// GetPosition は現在位置（秒）を返す
func (t *Track) GetPosition() float64 {
	// 💡 修正: ロック時間を最小化するため、必要な値を即座にコピーする
	t.mu.RLock()
	position := t.floatPosition // 💡 修正: floatPosition を使用
	dataLen := len(t.Data)
	t.mu.RUnlock()

	if dataLen == 0 {
		return 0
	}
	return position / float64(t.Channels) / float64(t.SampleRate)
}

// GetDuration はトラックの長さ（秒）を返す
func (t *Track) GetDuration() float64 {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// 💡 修正: ゼロ除算を確実に防ぐ
	if len(t.Data) == 0 || t.Channels == 0 || t.SampleRate == 0 {
		return 0.0
	}
	return float64(len(t.Data)) / float64(t.Channels) / float64(t.SampleRate)
}

// SetVolume は音量を設定
func (t *Track) SetVolume(volume float64) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if volume < 0 {
		volume = 0
	}
	if volume > 1.0 {
		volume = 1.0
	}
	t.Volume = volume
}

// SetSpeed はピッチ/スピードを設定
func (t *Track) SetSpeed(speed float64) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// 0.5倍速 ～ 2.0倍速
	if speed < 0.5 {
		speed = 0.5
	}
	if speed > 2.0 {
		speed = 2.0
	}
	t.Speed = speed
}

// Play は再生開始
func (t *Track) Play() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.IsPlaying = true
}

// Pause は一時停止
func (t *Track) Pause() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.IsPlaying = false
}

// Stop は停止して先頭に戻る
func (t *Track) Stop() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.IsPlaying = false
	t.floatPosition = 0 // 💡 修正
}

// AddCuePoint はキューポイントを追加
func (t *Track) AddCuePoint(name, color string) {
	pos := t.GetPosition()
	t.CueManager.AddCuePoint(name, pos, color)
	fmt.Printf("📍 Cue point added: %s at %.2fs\n", name, pos)
}

// JumpToCuePoint は指定キューポイントにジャンプ
func (t *Track) JumpToCuePoint(index int) bool {
	cue := t.CueManager.GetCuePoint(index)
	if cue == nil {
		return false
	}
	t.Seek(cue.Position)
	fmt.Printf("⏩ Jumped to: %s (%.2fs)\n", cue.Name, cue.Position)
	return true
}
