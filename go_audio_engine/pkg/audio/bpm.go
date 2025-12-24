package audio

import (
	"math"
)

// BPMDetector はBPM（テンポ）を検出
type BPMDetector struct {
	sampleRate int
	bpm        float64
	confidence float64 // 検出の信頼度（0.0 - 1.0）
}

// NewBPMDetector はBPM検出器を作成
// コンストラクタパターン：構造体を初期化して返す
func NewBPMDetector(sampleRate int) *BPMDetector {
	return &BPMDetector{
		sampleRate: sampleRate,
		bpm:        0,
		confidence: 0,
	}
}

// DetectBPM は音声データからBPMを検出
// スライス（[]float32）は動的配列、Goの重要なデータ構造
func (d *BPMDetector) DetectBPM(samples []float32) float64 {
	// サンプル数が少なすぎる場合は検出不可
	if len(samples) < d.sampleRate*2 {
		return 0
	}

	// 1. エネルギー包絡線を計算
	// 音の大きさの変化を追跡する
	envelope := d.calculateEnvelope(samples)

	// 2. ピーク検出（ビートの位置を見つける）
	peaks := d.detectPeaks(envelope)

	// 3. ピーク間隔からBPMを計算
	if len(peaks) < 2 {
		return 0 // ピークが少なすぎる
	}

	// ピーク間の平均間隔を計算
	intervals := make([]float64, 0, len(peaks)-1)
	for i := 1; i < len(peaks); i++ {
		interval := float64(peaks[i] - peaks[i-1])
		intervals = append(intervals, interval)
	}

	// 中央値を使用（外れ値に強い）
	avgInterval := d.median(intervals)

	// サンプル数から秒に変換し、BPMを計算
	// BPM = 60秒 / (ビート間隔[秒])
	intervalInSeconds := avgInterval / float64(d.sampleRate)
	if intervalInSeconds > 0 {
		d.bpm = 60.0 / intervalInSeconds

		// BPMの妥当な範囲にクランプ（60-200 BPM）
		if d.bpm < 60 {
			d.bpm *= 2 // ハーフタイムの可能性
		}
		if d.bpm > 200 {
			d.bpm /= 2 // ダブルタイムの可能性
		}

		d.confidence = d.calculateConfidence(intervals)
	}

	return d.bpm
}

// calculateEnvelope はエネルギー包絡線を計算
// 解説：音の「大きさの変化」を滑らかな曲線として表現
func (d *BPMDetector) calculateEnvelope(samples []float32) []float64 {
	// ウィンドウサイズ：約0.05秒分のサンプル
	windowSize := d.sampleRate / 20
	envelope := make([]float64, len(samples)/windowSize)

	// 各ウィンドウでRMS（Root Mean Square）を計算
	for i := range envelope {
		start := i * windowSize
		end := start + windowSize
		if end > len(samples) {
			end = len(samples)
		}

		// RMS計算：音の実効値
		var sum float64
		for j := start; j < end; j++ {
			val := float64(samples[j])
			sum += val * val // 二乗和
		}
		envelope[i] = math.Sqrt(sum / float64(end-start)) // 平均の平方根
	}

	return envelope
}

// detectPeaks はエネルギー包絡線からピークを検出
// 解説：「ドン」というビートの位置を見つける
func (d *BPMDetector) detectPeaks(envelope []float64) []int {
	if len(envelope) < 3 {
		return nil
	}

	// 動的閾値を計算（平均値の1.5倍）
	var sum float64
	for _, val := range envelope {
		sum += val
	}
	threshold := (sum / float64(len(envelope))) * 1.5

	peaks := make([]int, 0, 100)

	// ピーク検出：前後より大きく、閾値を超える点
	for i := 1; i < len(envelope)-1; i++ {
		if envelope[i] > envelope[i-1] &&
			envelope[i] > envelope[i+1] &&
			envelope[i] > threshold {
			peaks = append(peaks, i)
		}
	}

	return peaks
}

// median は中央値を計算
// 解説：平均値より外れ値に強い統計量
func (d *BPMDetector) median(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	// コピーしてソート
	sorted := make([]float64, len(values))
	copy(sorted, values)

	// バブルソート（小規模データなので十分）
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i] > sorted[j] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	// 中央値を返す
	mid := len(sorted) / 2
	if len(sorted)%2 == 0 {
		return (sorted[mid-1] + sorted[mid]) / 2
	}
	return sorted[mid]
}

// calculateConfidence は検出の信頼度を計算
// 解説：ビート間隔のばらつきが小さいほど信頼度が高い
func (d *BPMDetector) calculateConfidence(intervals []float64) float64 {
	if len(intervals) < 2 {
		return 0
	}

	// 標準偏差を計算
	mean := 0.0
	for _, val := range intervals {
		mean += val
	}
	mean /= float64(len(intervals))

	variance := 0.0
	for _, val := range intervals {
		diff := val - mean
		variance += diff * diff
	}
	variance /= float64(len(intervals))
	stdDev := math.Sqrt(variance)

	// 変動係数（CV）から信頼度を計算
	// CV = 標準偏差 / 平均
	cv := stdDev / mean
	confidence := 1.0 - math.Min(cv, 1.0)

	return confidence
}

// GetBPM は検出されたBPMを返す
func (d *BPMDetector) GetBPM() float64 {
	return d.bpm
}

// GetConfidence は検出の信頼度を返す
func (d *BPMDetector) GetConfidence() float64 {
	return d.confidence
}
