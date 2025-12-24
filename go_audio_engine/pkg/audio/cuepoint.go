package audio

// CuePoint はキューポイント（頭出し位置）
type CuePoint struct {
	Name     string  // キューポイントの名前（例："Intro", "Drop"）
	Position float64 // 位置（秒）
	Color    string  // UI表示用の色
}

// Loop はループ区間
type Loop struct {
	Enabled  bool    // ループが有効か
	Start    float64 // 開始位置（秒）
	End      float64 // 終了位置（秒）
	Length   float64 // ループの長さ（秒）
	IsActive bool    // 現在ループ中か
}

// CuePointManager はキューポイントとループを管理
type CuePointManager struct {
	CuePoints []CuePoint // スライス：可変長の配列
	Loop      Loop
}

// NewCuePointManager はマネージャーを作成
func NewCuePointManager() *CuePointManager {
	return &CuePointManager{
		CuePoints: make([]CuePoint, 0, 8), // 初期容量8
		Loop: Loop{
			Enabled: false,
		},
	}
}

// AddCuePoint はキューポイントを追加
func (m *CuePointManager) AddCuePoint(name string, position float64, color string) {
	// append: スライスに要素を追加（Goの重要な組み込み関数）
	m.CuePoints = append(m.CuePoints, CuePoint{
		Name:     name,
		Position: position,
		Color:    color,
	})
}

// RemoveCuePoint は指定インデックスのキューポイントを削除
func (m *CuePointManager) RemoveCuePoint(index int) bool {
	if index < 0 || index >= len(m.CuePoints) {
		return false // 範囲外
	}

	// スライスから要素を削除する慣用句
	// 解説：削除したい要素の前と後をつなげる
	m.CuePoints = append(
		m.CuePoints[:index],      // 前半部分
		m.CuePoints[index+1:]..., // 後半部分（...は展開演算子）
	)
	return true
}

// GetCuePoint は指定インデックスのキューポイントを取得
func (m *CuePointManager) GetCuePoint(index int) *CuePoint {
	if index < 0 || index >= len(m.CuePoints) {
		return nil // nilはポインタのゼロ値（存在しないことを表す）
	}
	return &m.CuePoints[index]
}

// SetLoop はループ区間を設定
func (m *CuePointManager) SetLoop(start, end float64) {
	if start >= end {
		return // 無効な区間
	}

	m.Loop.Start = start
	m.Loop.End = end
	m.Loop.Length = end - start
	m.Loop.Enabled = true
}

// EnableLoop はループを有効化
func (m *CuePointManager) EnableLoop(enabled bool) {
	m.Loop.Enabled = enabled
	if !enabled {
		m.Loop.IsActive = false
	}
}

// CheckLoop は現在位置がループ終点を超えたかチェック
// 超えていたらループ開始位置を返す
func (m *CuePointManager) CheckLoop(currentPosition float64) (shouldLoop bool, newPosition float64) {
	if !m.Loop.Enabled || !m.Loop.IsActive {
		return false, currentPosition
	}

	// ループ終点を超えた場合
	if currentPosition >= m.Loop.End {
		return true, m.Loop.Start
	}

	return false, currentPosition
}

// ActivateLoop はループを開始
func (m *CuePointManager) ActivateLoop() {
	if m.Loop.Enabled {
		m.Loop.IsActive = true
	}
}

// DeactivateLoop はループを停止（次の通過時にループしない）
func (m *CuePointManager) DeactivateLoop() {
	m.Loop.IsActive = false
}

// FindNearestCuePoint は指定位置に最も近いキューポイントを探す
func (m *CuePointManager) FindNearestCuePoint(position float64) *CuePoint {
	if len(m.CuePoints) == 0 {
		return nil
	}

	// 最も近いキューポイントを見つける
	var nearest *CuePoint
	minDistance := -1.0

	for i := range m.CuePoints {
		// 距離を計算（絶対値）
		distance := position - m.CuePoints[i].Position
		if distance < 0 {
			distance = -distance
		}

		// 最小距離を更新
		if minDistance < 0 || distance < minDistance {
			minDistance = distance
			nearest = &m.CuePoints[i]
		}
	}

	return nearest
}

// ClearAllCuePoints は全てのキューポイントを削除
func (m *CuePointManager) ClearAllCuePoints() {
	// 新しい空のスライスを作成
	m.CuePoints = make([]CuePoint, 0, 8)
}

// GetCuePointCount はキューポイントの数を返す
func (m *CuePointManager) GetCuePointCount() int {
	return len(m.CuePoints)
}
