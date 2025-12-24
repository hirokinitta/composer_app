/**
 * AudioEngineClient - Goバックエンドとの通信クライアント
 * 
 * 設計パターン:
 * - クラスベース: 状態（baseURL）と機能をカプセル化
 * - async/await: 非同期処理を同期的に書ける
 * - Fetch API: HTTPリクエストを送信
 */
export class AudioEngineClient {
    constructor(baseUrl = 'http://localhost:8080', wsUrl = 'ws://localhost:8080/ws/status') {
        this.baseUrl = baseUrl;
        this.wsUrl = wsUrl;
        this.ws = null;
        this.onStatusUpdate = (status) => {}; // ステータス更新時のコールバック
    }

    /**
     * WebSocket接続を確立
     * @param {(status: object) => void} onStatusUpdateCallback - ステータス更新時のコールバック関数
     */
    connectWebSocket(onStatusUpdateCallback) {
        this.onStatusUpdate = onStatusUpdateCallback;
        this.ws = new WebSocket(this.wsUrl);

        this.ws.onopen = () => {
            console.log('WebSocket Connected');
            // 接続確立時に初期ステータスを取得（任意）
            this.getMixerStatus().then(status => {
                this.onStatusUpdate(status);
            }).catch(error => {
                console.error('Initial status fetch via HTTP failed:', error);
            });
        };

        this.ws.onmessage = (event) => {
            const status = JSON.parse(event.data);
            this.onStatusUpdate(status);
        };

        this.ws.onclose = () => {
            console.log('WebSocket Disconnected');
            // 再接続ロジックをここに追加することも可能
        };

        this.ws.onerror = (error) => {
            console.error('WebSocket Error:', error);
        };
    }

    /**
     * WebSocket接続を閉じる
     */
    disconnectWebSocket() {
        if (this.ws) {
            this.ws.close();
            this.ws = null;
        }
    }

    // ========================================
    // 基本操作（既存）
    // ========================================

    /**
     * トラックをロード
     * @param {string} deck - 'a' or 'b'
     * @param {string} filePath - ファイルパス
     */
    async loadTrack(deck, filePath) {
        const response = await fetch(
            `${this.baseUrl}/api/deck/${deck}/load?file=${encodeURIComponent(filePath)}`,
            { method: 'GET' }
        );
        return response.json();
    }

    async play(deck) {
        const response = await fetch(`${this.baseUrl}/api/deck/${deck}/play`, {
            method: 'POST'
        });
        return response.json();
    }

    async pause(deck) {
        const response = await fetch(`${this.baseUrl}/api/deck/${deck}/pause`, {
            method: 'POST'
        });
        return response.json();
    }

    async stop(deck) {
        const response = await fetch(`${this.baseUrl}/api/deck/${deck}/stop`, {
            method: 'POST'
        });
        return response.json();
    }

    async setVolume(deck, volume) {
        const response = await fetch(`${this.baseUrl}/api/deck/${deck}/volume`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ volume })
        });
        return response.json();
    }

    async seek(deck, position) {
        const response = await fetch(`${this.baseUrl}/api/deck/${deck}/seek`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ position })
        });
        return response.json();
    }

    // ========================================
    // 新機能: EQ（イコライザー）
    // ========================================

    /**
     * EQ設定
     * 
     * Go側の対応する構造体:
     * type ThreeBandEQ struct {
     *     Low  float64
     *     Mid  float64
     *     High float64
     * }
     * 
     * @param {string} deck - 'a' or 'b'
     * @param {object} eq - { low: -1~1, mid: -1~1, high: -1~1 }
     */
    async setEQ(deck, eq) {
        // JSON.stringify: JavaScriptオブジェクト → JSON文字列
        // Go側では json.Decoder でパース
        const response = await fetch(`${this.baseUrl}/api/deck/${deck}/eq`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                low: eq.low || 0,
                mid: eq.mid || 0,
                high: eq.high || 0
            })
        });
        return response.json();
    }

    // ========================================
    // 新機能: Filter（フィルター）
    // ========================================

    /**
     * フィルター設定
     * 
     * Go側の対応:
     * type Filter struct {
     *     Type      string  // "lowpass", "highpass", "none"
     *     Cutoff    float64 // 0.0 - 1.0
     *     Resonance float64 // 0.0 - 1.0
     * }
     * 
     * @param {string} deck 
     * @param {object} filter - { type, cutoff, resonance }
     */
    async setFilter(deck, filter) {
        const response = await fetch(`${this.baseUrl}/api/deck/${deck}/filter`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                type: filter.type || 'none',
                cutoff: filter.cutoff || 0.5,
                resonance: filter.resonance || 0
            })
        });
        return response.json();
    }

    // ========================================
    // 新機能: Speed/Pitch Control
    // ========================================

    /**
     * スピード（ピッチ）設定
     * 
     * 0.5 = 半分の速度（-50%）
     * 1.0 = 通常速度（±0%）
     * 2.0 = 2倍速（+100%）
     */
    async setSpeed(deck, speed) {
        const response = await fetch(`${this.baseUrl}/api/deck/${deck}/speed`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ speed })
        });
        return response.json();
    }

    // ========================================
    // 新機能: Cue Points（キューポイント）
    // ========================================

    /**
     * キューポイントを追加
     * 
     * Go側では現在の再生位置にキューポイントを設定
     * 
     * Go側の処理:
     * func (t *Track) AddCuePoint(name, color string) {
     *     pos := t.GetPosition()  // 現在位置を取得
     *     t.CueManager.AddCuePoint(name, pos, color)
     * }
     */
    async addCuePoint(deck, name, color) {
        const response = await fetch(`${this.baseUrl}/api/deck/${deck}/cuepoint/add`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ name, color })
        });
        return response.json();
    }

    /**
     * キューポイントにジャンプ
     * 
     * @param {string} deck 
     * @param {number} index - キューポイントのインデックス（0から始まる）
     */
    async jumpToCuePoint(deck, index) {
        const response = await fetch(`${this.baseUrl}/api/deck/${deck}/cuepoint/jump`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ index })
        });
        return response.json();
    }

    // ========================================
    // 新機能: Loop（ループ）
    // ========================================

    /**
     * ループ区間を設定
     * 
     * @param {string} deck 
     * @param {number} start - 開始位置（秒）
     * @param {number} end - 終了位置（秒）
     */
    async setLoop(deck, start, end) {
        const response = await fetch(`${this.baseUrl}/api/deck/${deck}/loop/set`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ start, end })
        });
        return response.json();
    }

    /**
     * ループの有効/無効を切り替え
     */
    async enableLoop(deck, enabled) {
        const response = await fetch(`${this.baseUrl}/api/deck/${deck}/loop/enable`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ enabled })
        });
        return response.json();
    }

    // ========================================
    // Mixer操作
    // ========================================

    async setCrossfader(value) {
        const response = await fetch(`${this.baseUrl}/api/mixer/crossfader`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ value })
        });
        return response.json();
    }

    async setMasterVolume(volume) {
        const response = await fetch(`${this.baseUrl}/api/mixer/master`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ volume })
        });
        return response.json();
    }

    /**
     * BPM同期の設定
     * 
     * Go側の処理:
     * - 2つのトラックのBPMを検出
     * - マスター（基準）デッキのBPMに合わせて、もう一方のスピードを調整
     * 
     * 例: Deck A = 120 BPM（マスター）, Deck B = 130 BPM
     * → Deck Bのスピード = 120/130 ≈ 0.92
     * 
     * @param {boolean} enabled - 同期を有効にするか
     * @param {string} master - 'a' or 'b' - どちらをマスターにするか
     */
    async enableSync(enabled, master) {
        const response = await fetch(`${this.baseUrl}/api/mixer/sync`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ enabled, master })
        });
        return response.json();
    }

    /**
     * ミキサーの状態を取得
     * 
     * Go側の返却値（JSON）:
     * {
     *   "deckA": {
     *     "file": "path/to/file.wav",
     *     "playing": true,
     *     "position": 45.2,
     *     "duration": 180.0,
     *     "volume": 0.8,
     *     "speed": 1.0,
     *     "bpm": 128.0,
     *     "eq": { "low": 0, "mid": 0, "high": 0 },
     *     "filter": { "type": "none", ... },
     *     "cuePoints": [...],
     *     "loop": { ... }
     *   },
     *   "deckB": { ... },
     *   "crossfader": 0.0,
     *   "masterVolume": 1.0,
     *   "syncEnabled": false,
     *   "syncMaster": "a"
     * }
     */
    async getMixerStatus() {
        const response = await fetch(`${this.baseUrl}/api/mixer/status`);
        return response.json();
    }

    /**
     * 接続確認
     * 
     * try-catchパターン:
     * - 成功: true を返す
     * - 失敗: false を返す（エラーを投げない）
     */
    async checkConnection() {
        try {
            await this.getMixerStatus();
            return true;
        } catch {
            return false;
        }
    }
}

/**
 * ========================================
 * Go側の対応する型定義（参考）
 * ========================================
 * 
 * // DeckのJSON表現
 * type DeckStatus struct {
 *     File     string              `json:"file"`
 *     Playing  bool                `json:"playing"`
 *     Position float64             `json:"position"`
 *     Duration float64             `json:"duration"`
 *     Volume   float64             `json:"volume"`
 *     Speed    float64             `json:"speed"`
 *     BPM      float64             `json:"bpm"`
 *     EQ       EQStatus            `json:"eq"`
 *     Filter   FilterStatus        `json:"filter"`
 *     CuePoints []CuePointStatus   `json:"cuePoints"`
 *     Loop     LoopStatus          `json:"loop"`
 * }
 * 
 * // PHPとの比較:
 * 
 * // Go
 * type User struct {
 *     Name string `json:"name"`
 *     Age  int    `json:"age"`
 * }
 * 
 * // PHP（配列またはクラス）
 * $user = [
 *     'name' => 'Alice',
 *     'age' => 25
 * ];
 * 
 * または
 * 
 * class User {
 *     public string $name;
 *     public int $age;
 * }
 */