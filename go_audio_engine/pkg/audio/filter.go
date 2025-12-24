package audio

import "math"

// Filter はオーディオフィルター
type Filter struct {
	Type      string  // "lowpass", "highpass", "none"
	Cutoff    float64 // 0.0 ～ 1.0
	Resonance float64 // 0.0 ～ 1.0

	sampleRate float64

	// フィルター状態
	buf0 [2]float64
	buf1 [2]float64
}

func NewFilter(sampleRate float64) *Filter {
	return &Filter{
		Type:       "none",
		Cutoff:     0.5,
		Resonance:  0.0,
		sampleRate: sampleRate,
	}
}

// Process はフィルターを適用
func (f *Filter) Process(samples []float32) {
	if f.Type == "none" {
		return
	}

	for i := 0; i < len(samples); i += 2 {
		// 左チャンネル
		samples[i] = float32(f.processChannel(float64(samples[i]), 0))

		// 右チャンネル
		if i+1 < len(samples) {
			samples[i+1] = float32(f.processChannel(float64(samples[i+1]), 1))
		}
	}
}

func (f *Filter) processChannel(sample float64, channel int) float64 {
	// カットオフ周波数を計算（20Hz - 20kHz）
	minFreq := 20.0
	maxFreq := 20000.0
	freq := minFreq + (maxFreq-minFreq)*f.Cutoff

	// 正規化された周波数
	omega := 2 * math.Pi * freq / f.sampleRate
	sinOmega := math.Sin(omega)
	cosOmega := math.Cos(omega)

	// レゾナンス（Q値）
	Q := 1.0 + f.Resonance*9.0 // 1.0 - 10.0
	alpha := sinOmega / (2 * Q)

	var b0, b1, b2, a0, a1, a2 float64

	switch f.Type {
	case "lowpass":
		b0 = (1 - cosOmega) / 2
		b1 = 1 - cosOmega
		b2 = (1 - cosOmega) / 2
		a0 = 1 + alpha
		a1 = -2 * cosOmega
		a2 = 1 - alpha

	case "highpass":
		b0 = (1 + cosOmega) / 2
		b1 = -(1 + cosOmega)
		b2 = (1 + cosOmega) / 2
		a0 = 1 + alpha
		a1 = -2 * cosOmega
		a2 = 1 - alpha

	default:
		return sample
	}

	// バイカッドフィルター適用
	output := (b0/a0)*sample + (b1/a0)*f.buf0[channel] + (b2/a0)*f.buf1[channel] -
		(a1/a0)*f.buf0[channel] - (a2/a0)*f.buf1[channel]

	f.buf1[channel] = f.buf0[channel]
	f.buf0[channel] = sample

	// ソフトクリッピング
	if output > 1.0 {
		output = 1.0
	}
	if output < -1.0 {
		output = -1.0
	}

	return output
}

// SetLowpass はローパスフィルターを設定
func (f *Filter) SetLowpass(cutoff, resonance float64) {
	f.Type = "lowpass"
	f.Cutoff = clamp(cutoff, 0, 1)
	f.Resonance = clamp(resonance, 0, 1)
}

// SetHighpass はハイパスフィルターを設定
func (f *Filter) SetHighpass(cutoff, resonance float64) {
	f.Type = "highpass"
	f.Cutoff = clamp(cutoff, 0, 1)
	f.Resonance = clamp(resonance, 0, 1)
}

// Reset はフィルターをリセット
func (f *Filter) Reset() {
	f.Type = "none"
	f.buf0 = [2]float64{0, 0}
	f.buf1 = [2]float64{0, 0}
}

func clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
