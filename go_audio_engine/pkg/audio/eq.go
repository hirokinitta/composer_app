package audio

import (
	"math"
)

// ThreeBandEQ は3バンドイコライザー
type ThreeBandEQ struct {
	Low  float64 // -1.0 ～ 1.0 (0がフラット)
	Mid  float64
	High float64

	// フィルター状態（バイカッド）
	lowState  [2]float64
	midState  [2]float64
	highState [2]float64

	sampleRate float64
}

func NewThreeBandEQ(sampleRate float64) *ThreeBandEQ {
	return &ThreeBandEQ{
		Low:        0,
		Mid:        0,
		High:       0,
		sampleRate: sampleRate,
	}
}

// Process はサンプルにEQを適用
func (eq *ThreeBandEQ) Process(samples []float32) {
	for i := 0; i < len(samples); i += 2 {
		// 左チャンネル
		left := float64(samples[i])
		left = eq.processLowShelf(left, 0)
		left = eq.processPeaking(left, 1)
		left = eq.processHighShelf(left, 2)
		samples[i] = float32(left)

		// 右チャンネル
		if i+1 < len(samples) {
			right := float64(samples[i+1])
			right = eq.processLowShelf(right, 0)
			right = eq.processPeaking(right, 1)
			right = eq.processHighShelf(right, 2)
			samples[i+1] = float32(right)
		}
	}
}

// ローシェルフフィルター（低音域）
func (eq *ThreeBandEQ) processLowShelf(sample float64, channel int) float64 {
	if eq.Low == 0 {
		return sample
	}

	// 簡易的なローシェルフ（100Hz付近）
	freq := 100.0
	gain := math.Pow(10, eq.Low*12/20) // ±12dB

	omega := 2 * math.Pi * freq / eq.sampleRate
	sinOmega := math.Sin(omega)
	cosOmega := math.Cos(omega)
	alpha := sinOmega / 2

	a := math.Sqrt(gain)

	b0 := a * ((a + 1) - (a-1)*cosOmega + 2*math.Sqrt(a)*alpha)
	b1 := 2 * a * ((a - 1) - (a+1)*cosOmega)
	b2 := a * ((a + 1) - (a-1)*cosOmega - 2*math.Sqrt(a)*alpha)
	a0 := (a + 1) + (a-1)*cosOmega + 2*math.Sqrt(a)*alpha
	a1 := -2 * ((a - 1) + (a+1)*cosOmega)
	a2 := (a + 1) + (a-1)*cosOmega - 2*math.Sqrt(a)*alpha

	// バイカッドフィルター
	output := (b0/a0)*sample + (b1/a0)*eq.lowState[0] + (b2/a0)*eq.lowState[1] -
		(a1/a0)*eq.lowState[0] - (a2/a0)*eq.lowState[1]

	eq.lowState[1] = eq.lowState[0]
	eq.lowState[0] = sample

	return output
}

// ピーキングフィルター（中音域）
func (eq *ThreeBandEQ) processPeaking(sample float64, channel int) float64 {
	if eq.Mid == 0 {
		return sample
	}

	// 1kHz付近
	freq := 1000.0
	gain := math.Pow(10, eq.Mid*12/20)
	Q := 1.0

	omega := 2 * math.Pi * freq / eq.sampleRate
	sinOmega := math.Sin(omega)
	cosOmega := math.Cos(omega)
	alpha := sinOmega / (2 * Q)

	b0 := 1 + alpha*gain
	b1 := -2 * cosOmega
	b2 := 1 - alpha*gain
	a0 := 1 + alpha/gain
	a1 := -2 * cosOmega
	a2 := 1 - alpha/gain

	output := (b0/a0)*sample + (b1/a0)*eq.midState[0] + (b2/a0)*eq.midState[1] -
		(a1/a0)*eq.midState[0] - (a2/a0)*eq.midState[1]

	eq.midState[1] = eq.midState[0]
	eq.midState[0] = sample

	return output
}

// ハイシェルフフィルター（高音域）
func (eq *ThreeBandEQ) processHighShelf(sample float64, channel int) float64 {
	if eq.High == 0 {
		return sample
	}

	// 10kHz付近
	freq := 10000.0
	gain := math.Pow(10, eq.High*12/20)

	omega := 2 * math.Pi * freq / eq.sampleRate
	sinOmega := math.Sin(omega)
	cosOmega := math.Cos(omega)
	alpha := sinOmega / 2

	a := math.Sqrt(gain)

	b0 := a * ((a + 1) + (a-1)*cosOmega + 2*math.Sqrt(a)*alpha)
	b1 := -2 * a * ((a - 1) + (a+1)*cosOmega)
	b2 := a * ((a + 1) + (a-1)*cosOmega - 2*math.Sqrt(a)*alpha)
	a0 := (a + 1) - (a-1)*cosOmega + 2*math.Sqrt(a)*alpha
	a1 := 2 * ((a - 1) - (a+1)*cosOmega)
	a2 := (a + 1) - (a-1)*cosOmega - 2*math.Sqrt(a)*alpha

	output := (b0/a0)*sample + (b1/a0)*eq.highState[0] + (b2/a0)*eq.highState[1] -
		(a1/a0)*eq.highState[0] - (a2/a0)*eq.highState[1]

	eq.highState[1] = eq.highState[0]
	eq.highState[0] = sample

	return output
}

// SetLow は低音域のゲインを設定（-1.0 ～ 1.0）
func (eq *ThreeBandEQ) SetLow(gain float64) {
	if gain < -1.0 {
		gain = -1.0
	}
	if gain > 1.0 {
		gain = 1.0
	}
	eq.Low = gain
}

// SetMid は中音域のゲインを設定
func (eq *ThreeBandEQ) SetMid(gain float64) {
	if gain < -1.0 {
		gain = -1.0
	}
	if gain > 1.0 {
		gain = 1.0
	}
	eq.Mid = gain
}

// SetHigh は高音域のゲインを設定
func (eq *ThreeBandEQ) SetHigh(gain float64) {
	if gain < -1.0 {
		gain = -1.0
	}
	if gain > 1.0 {
		gain = 1.0
	}
	eq.High = gain
}
