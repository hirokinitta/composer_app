package mixer

import (
	"log"
	"math"
	"sync"

	"go_audio_engine/pkg/audio"
)

// 💡 追加: Deckを識別するための型と定数
type DeckID int

const (
	DeckA DeckID = iota
	DeckB
)

// 💡 追加: ファイルロードリクエストを表す構造体
type loadRequest struct {
	deckID   DeckID
	filePath string
}

// 💡 追加: デコード済みのオーディオデータを表す構造体
type loadedTrack struct {
	deckID    DeckID
	trackData *audio.Track // 新しいTrackオブジェクトをそのまま渡す
}

// DJMixer はプロフェッショナルDJミキサー
type DJMixer struct {
	DeckA        *audio.Track
	DeckB        *audio.Track
	Crossfader   float64 // -1.0 (A) ～ 0.0 (Center) ～ 1.0 (B)
	MasterVolume float64

	// 新機能
	SyncEnabled bool   // BPM同期が有効か
	SyncMaster  string // "a" or "b" - どちらがマスターか

	mu sync.RWMutex

	// 💡 追加: 非同期ロードのためのチャンネル
	loadRequestChan chan loadRequest // UIスレッドからMixスレッドへ
	loadedTrackChan chan loadedTrack // Mixスレッド内で安全に適用するため
	sampleRate      int              // Track生成時に必要なので保持
}

// NewDJMixer は新しいDJミキサーを作成
func NewDJMixer(sampleRate int) *DJMixer {
	m := &DJMixer{
		DeckA:        audio.NewTrack(sampleRate),
		DeckB:        audio.NewTrack(sampleRate),
		Crossfader:   0.0,
		MasterVolume: 1.0,
		SyncEnabled:  false,
		SyncMaster:   "a",
		// 💡 追加: チャンネルの初期化
		loadRequestChan: make(chan loadRequest, 10), // バッファを持たせる
		loadedTrackChan: make(chan loadedTrack, 10),
		sampleRate:      sampleRate,
	}

	// 💡 修正: DJミキサー自身のゴルーチンをコンストラクタで起動する
	go m.processLoadRequests()

	return m
}

// 💡 追加: 非同期でトラックをロードするメソッド
func (m *DJMixer) LoadTrackAsync(deckID DeckID, filePath string) {
	// リクエストをチャンネルに送信するだけ。重い処理は行わない。
	m.loadRequestChan <- loadRequest{deckID: deckID, filePath: filePath}
}

// 💡 追加: 実際にファイルをデコードする内部メソッド
func (m *DJMixer) processLoadRequests() {
	// このゴルーチンは、ロードリクエストを待ち受け、デコード処理を行う
	for req := range m.loadRequestChan {
		log.Printf("🎵 [Decoder] Start decoding: %s for Deck %d", req.filePath, req.deckID)

		// 新しいTrackオブジェクトを作成し、ファイルをロードする
		newTrack := audio.NewTrack(m.sampleRate)
		err := newTrack.LoadWAV(req.filePath) // ここが重い処理
		if err != nil {
			log.Printf("❌ [Decoder] Failed to load WAV for Deck %d: %v", req.deckID, err)
			continue // エラーが発生したら次のリクエストへ
		}

		// 💡 追加: デコード成功後、同じゴルーチン内でBPM検出を実行
		go newTrack.DetectBPMAsync()

		log.Printf("✅ [Decoder] Finished decoding: %s. Sending to mixer.", req.filePath)
		// デコード成功後、結果をloadedTrackChanに送信
		m.loadedTrackChan <- loadedTrack{deckID: req.deckID, trackData: newTrack}
	}
}

// Mix は2つのデッキをミックス
// 解説：DJミキサーの心臓部
func (m *DJMixer) Mix(out []float32) {
	// 💡 追加: デッドロックを避けるため、Mixループ内で安全にトラックを入れ替える
	select {
	case loaded := <-m.loadedTrackChan:
		m.swapTrack(loaded)
	default:
		// 新しいトラックがなければ何もしない (ノンブロッキング)
	}

	m.mu.RLock()
	crossfader := m.Crossfader
	masterVolume := m.MasterVolume
	// syncEnabled := m.SyncEnabled
	// syncMaster := m.SyncMaster
	m.mu.RUnlock()

	// BPM同期処理
	// if syncEnabled {
	// 	m.applySyncSpeed(syncMaster)
	// }

	// デッキA/Bの音声を取得
	bufferA := make([]float32, len(out))
	bufferB := make([]float32, len(out))

	m.DeckA.ReadSamples(bufferA)
	m.DeckB.ReadSamples(bufferB)

	// クロスフェーダーカーブの計算
	// 解説：等パワークロスフェード（聴感上の音量が一定）
	//
	// -1.0 → A側フル: gainA=1.0, gainB=0.0
	//  0.0 → 中央:    gainA=0.707, gainB=0.707 (√2/2)
	//  1.0 → B側フル: gainA=0.0, gainB=1.0

	// 正規化: -1～1 を 0～1 に変換
	normalized := (crossfader + 1.0) / 2.0

	// 等パワーカーブ（三角関数を使用）
	// angleA := (1.0 - normalized) * math.Pi / 2.0 // π/2 ～ 0
	angleB := normalized * math.Pi / 2.0 // 0 ～ π/2

	gainA := math.Cos(angleB) // B側の角度でA側のゲイン
	gainB := math.Sin(angleB) // B側の角度でB側のゲイン

	// ミックス実行
	for i := range out {
		mixed := bufferA[i]*float32(gainA) + bufferB[i]*float32(gainB)
		out[i] = mixed * float32(masterVolume)

		// ハードクリッピング防止
		// 解説：音割れを防ぐため±1.0に制限
		if out[i] > 1.0 {
			out[i] = 1.0
		} else if out[i] < -1.0 {
			out[i] = -1.0
		}
	}
}

// 💡 追加: デコード済みのトラックを安全に入れ替えるメソッド
func (m *DJMixer) swapTrack(loaded loadedTrack) {
	m.mu.Lock()
	defer m.mu.Unlock()

	log.Printf("🔄 [Mixer] Swapping track for Deck %d", loaded.deckID)
	// 古いトラックの再生を停止し、リソースを解放する
	if loaded.deckID == DeckA && m.DeckA != nil {
		m.DeckA.Stop()
	} else if loaded.deckID == DeckB && m.DeckB != nil {
		m.DeckB.Stop()
	}

	// 新しいトラックに差し替える
	if loaded.deckID == DeckA {
		m.DeckA = loaded.trackData
	} else if loaded.deckID == DeckB {
		m.DeckB = loaded.trackData
	}
}

// applySyncSpeed はBPM同期のスピード調整
// 解説：2つのトラックのBPMを合わせる
func (m *DJMixer) applySyncSpeed(master string) {
	var masterBPM, slaveBPM float64
	var slaveDeck *audio.Track

	if master == "a" {
		masterBPM = m.DeckA.BPM.GetBPM()
		slaveBPM = m.DeckB.BPM.GetBPM()
		slaveDeck = m.DeckB
	} else {
		masterBPM = m.DeckB.BPM.GetBPM()
		slaveBPM = m.DeckA.BPM.GetBPM()
		slaveDeck = m.DeckA
	}

	// BPMが検出されていない場合はスキップ
	if masterBPM == 0 || slaveBPM == 0 {
		return
	}

	// スピード比を計算
	// 例: Master=120BPM, Slave=130BPM → Speed=120/130≒0.92
	speedRatio := masterBPM / slaveBPM

	// スレーブのスピードを調整
	slaveDeck.SetSpeed(speedRatio)
}

// SetCrossfader はクロスフェーダー値を設定
func (m *DJMixer) SetCrossfader(value float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if value < -1.0 {
		value = -1.0
	}
	if value > 1.0 {
		value = 1.0
	}
	m.Crossfader = value
}

// SetMasterVolume はマスターボリュームを設定
func (m *DJMixer) SetMasterVolume(volume float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if volume < 0 {
		volume = 0
	}
	if volume > 1.0 {
		volume = 1.0
	}
	m.MasterVolume = volume
}

// EnableSync はBPM同期を有効化
func (m *DJMixer) EnableSync(enabled bool, master string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.SyncEnabled = enabled
	if master == "a" || master == "b" {
		m.SyncMaster = master
	}
}

// GetStatus はミキサーの状態を取得
// 解説：interfaceを使った柔軟なデータ構造
func (m *DJMixer) GetStatus() map[string]interface{} {
	// 💡 修正: ロック時間を最小化するため、必要な値を即座にコピーする
	m.mu.RLock()
	deckA := m.DeckA
	deckB := m.DeckB
	crossfader := m.Crossfader
	masterVolume := m.MasterVolume
	syncEnabled := m.SyncEnabled
	syncMaster := m.SyncMaster
	m.mu.RUnlock()

	// map[string]interface{}: キーが文字列、値が任意の型
	// JSON変換に便利
	// 💡 修正: コピーした値を使ってmapを構築する
	return map[string]interface{}{
		"DeckA":        m.getDeckStatus(deckA),
		"DeckB":        m.getDeckStatus(deckB),
		"Crossfader":   crossfader,
		"MasterVolume": masterVolume,
		"SyncEnabled":  syncEnabled,
		"SyncMaster":   syncMaster,
	}
}

// getDeckStatus は個別デッキの状態を取得（内部ヘルパー）
func (m *DJMixer) getDeckStatus(deck *audio.Track) map[string]interface{} {
	return map[string]interface{}{
		"FilePath":      deck.FilePath, // ✅ "file" -> "FilePath"
		"IsPlaying":     deck.IsPlaying,
		"Position":      deck.GetPosition(), // ✅ ...以下同様に大文字開始へ
		"Duration":      deck.GetDuration(),
		"Volume":        deck.Volume,
		"Speed":         deck.Speed,
		"BPM":           deck.BPM.GetBPM(),
		"BPMConfidence": deck.BPM.GetConfidence(), // 💡 修正: 統一のため大文字開始に
		"EQ": map[string]float64{
			"Low":  deck.EQ.Low,
			"Mid":  deck.EQ.Mid,
			"High": deck.EQ.High,
		},
		"Filter": map[string]interface{}{
			"Type":      deck.Filter.Type,
			"Cutoff":    deck.Filter.Cutoff,
			"Resonance": deck.Filter.Resonance,
		},
		"CuePoints": m.getCuePointsStatus(deck),
		"Loop": map[string]interface{}{
			"Enabled":  deck.CueManager.Loop.Enabled,
			"Start":    deck.CueManager.Loop.Start,
			"End":      deck.CueManager.Loop.End,
			"IsActive": deck.CueManager.Loop.IsActive,
		},
	}
}

// getCuePointsStatus はキューポイント情報を取得
func (m *DJMixer) getCuePointsStatus(deck *audio.Track) []map[string]interface{} {
	cuePoints := make([]map[string]interface{}, 0)

	for i := 0; i < deck.CueManager.GetCuePointCount(); i++ {
		cue := deck.CueManager.GetCuePoint(i)
		if cue != nil {
			// 💡 修正: JSONキーをPascalCaseに統一
			cuePoints = append(cuePoints, map[string]interface{}{
				"Name":     cue.Name,
				"Position": cue.Position,
				"Color":    cue.Color,
			})
		}
	}

	return cuePoints
}
