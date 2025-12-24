package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"go_audio_engine/pkg/mixer"

	"github.com/gordonklaus/portaudio"
	"github.com/gorilla/websocket"
)

const (
	sampleRate      = 44100
	channels        = 2
	framesPerBuffer = 512
)

type AudioEngine struct {
	stream *portaudio.Stream
	mixer  *mixer.DJMixer
}

type LoadRequest struct {
	File string `json:"file"`
}

// ---------------------------------------------------------
// WebSocket 設定
// ---------------------------------------------------------

// WebSocketアップグレーダーの設定
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// すべてのオリジンからの接続を許可 (CORS回避)
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// 接続中のクライアントを管理する構造体
type Client struct {
	conn  *websocket.Conn
	mutex sync.Mutex // 💡 接続ごとの書き込みロックを追加
}

type ClientManager struct {
	clients map[*websocket.Conn]*Client // 💡 Client構造体を保持するように変更
	sync.RWMutex
}

var manager = ClientManager{
	clients: make(map[*websocket.Conn]*Client),
}

// 💡 共通の書き込み関数を作成 (安全な送信)
func (c *Client) safeWrite(messageType int, data []byte) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.conn.WriteMessage(messageType, data)
}

// ---------------------------------------------------------

func NewAudioEngine() (*AudioEngine, error) {
	if err := portaudio.Initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize PortAudio: %v", err)
	}

	djMixer := mixer.NewDJMixer(sampleRate)

	engine := &AudioEngine{
		mixer: djMixer,
	}

	stream, err := portaudio.OpenDefaultStream(
		0,
		channels,
		float64(sampleRate),
		framesPerBuffer,
		func(out []float32) {
			engine.mixer.Mix(out)
		},
	)
	if err != nil {
		portaudio.Terminate()
		return nil, fmt.Errorf("failed to open stream: %v", err)
	}

	engine.stream = stream

	if err := stream.Start(); err != nil {
		return nil, fmt.Errorf("failed to start stream: %v", err)
	}

	return engine, nil
}

func (ae *AudioEngine) Close() {
	ae.stream.Stop()
	ae.stream.Close()
	portaudio.Terminate()
}

// enableCORS: 標準的な http.Handler ラッパーとして実装
func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// CORSヘッダーの設定
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "*")

		// プリフライトリクエスト (OPTIONS) の処理
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// 次のハンドラを実行
		next.ServeHTTP(w, r)
	})
}

// 💡 リファクタリング: Deck A と B の共通ハンドラを作成

// deckActionHandler は、引数なしでデッキを操作するアクション（play, pause, stop）のためのハンドラを生成します。
func deckActionHandler(action func()) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		action()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}
}

// deckLoadHandler は、ファイルのロード要求を処理するハンドラを生成します。
func deckLoadHandler(mixer *mixer.DJMixer, deckID mixer.DeckID) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req LoadRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if req.File == "" {
			http.Error(w, "file parameter required", http.StatusBadRequest)
			return
		}

		if _, err := os.Stat(req.File); os.IsNotExist(err) {
			http.Error(w, "file not found", http.StatusNotFound)
			return
		}

		log.Printf("⏳ Deck %d: Starting ASYNC WAV loading for: %s", deckID, req.File)
		go mixer.LoadTrackAsync(deckID, req.File)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(map[string]string{"status": "loading started", "file": req.File})
	}
}

// deckRequestHandler は、JSONボディを持つリクエストを処理する汎用ハンドラを生成します。
// seek, volume, eq, speed など、様々なアクションに使用できます。
func deckRequestHandler[T any](action func(T)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req T
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		action(req)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "ok", "request": req})
	}
}

func main() {
	engine, err := NewAudioEngine()
	if err != nil {
		log.Fatal("Failed to create audio engine:", err)
	}
	defer engine.Close()

	// 新しいマルチプレクサ(ルーター)を作成
	mux := http.NewServeMux()

	// =======================================================
	// 📡 WebSocket Endpoint (ステータス配信)
	// =======================================================
	mux.HandleFunc("/ws/status", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("❌ WebSocket upgrade failed: %v", err)
			return
		}

		// 💡 Clientオブジェクトを作成して登録
		newClient := &Client{conn: conn}
		// 💡 修正 1: クライアント識別子を取得
		clientAddr := conn.RemoteAddr().String()
		// クライアント登録
		manager.Lock()
		manager.clients[conn] = newClient
		currentTotal := len(manager.clients) // 登録後の数を取得
		manager.Unlock()
		log.Printf("✅ New WebSocket client connected. Addr: %s, Total: %d", clientAddr, currentTotal)

		// 💡 修正: defer を使って、このハンドラ関数を抜ける際に必ず切断処理を実行する
		defer func() {
			manager.Lock()
			delete(manager.clients, conn)
			remainingTotal := len(manager.clients) // 削除後の数を取得
			manager.Unlock()

			conn.Close()
			// 💡 修正: ログ出力の引数を修正
			log.Printf("🔌 WebSocket client disconnected. Addr: %s, Total remaining: %d", clientAddr, remainingTotal)
		}()

		// 💡 修正 A: 接続直後に現在のステータスを即座に送信
		go func() {
			status := engine.mixer.GetStatus()
			data, err := json.Marshal(status) // 💡 エラーチェックの変数名修正

			// err == nil のチェックは不要。Marshalが失敗した場合は data が nil になるが、ログがあればOK
			if err != nil {
				log.Printf("❌ Initial status JSON Marshal Error: %v", err)
				return
			}

			// 💡 修正 H: グローバルロックを削除し、Clientの safeWrite を使用
			if err = newClient.safeWrite(websocket.TextMessage, data); err == nil {
				log.Printf("📧 Initial status sent to new client: %s", clientAddr)
			} else {
				log.Printf("❌ Initial status send failed to %s: %v", clientAddr, err)
			}
		}()

		// 💡 修正: Ping/Pong の Read Deadline を設定
		// サーバーからの Ping (10秒間隔) に対する Pong 応答を待つ最大時間 (30秒)
		const pongWait = 30 * time.Second

		conn.SetReadDeadline(time.Now().Add(pongWait))

		// 💡 修正: クライアントからの Pong 応答で Read Deadline をリセット
		conn.SetPongHandler(func(string) error {
			conn.SetReadDeadline(time.Now().Add(pongWait))
			return nil
		})

		// 切断検知ループ
		for {
			// データを受け取らない場合は、タイプを無視してエラーのみチェック
			// クライアントが接続を閉じると、このReadMessageがエラーを返す
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
	})

	// =======================================================
	// 🎛️ HTTP API Endpoints (アクション)
	// =======================================================

	// ... (Deck A, Deck B, Mixer API ハンドラは変更なし。省略) ...

	// ========== Deck A API ==========

	mux.HandleFunc("/api/deck/a/load", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req LoadRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("❌ Deck A load: JSON decode error: %v", err)
			http.Error(w, "Invalid request body (JSON decode failed)", http.StatusBadRequest)
			return
		}

		file := req.File
		log.Printf("📥 Deck A load request received: %s", file)

		if file == "" {
			log.Printf("❌ Deck A load: file parameter missing in JSON")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "file parameter required"})
			return
		}

		if _, err := os.Stat(file); os.IsNotExist(err) {
			log.Printf("❌ Deck A load: file not found: %s", file)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "file not found"})
			return
		}

		log.Printf("⏳ Deck A: Starting ASYNC WAV loading for: %s", file)

		// 💡 修正: デッドロックを避けるため、チャンネル経由で安全にロード処理を依頼する
		// LoadTrackAsync は mixer パッケージ側での実装が必要になります。
		// このメソッドは内部でゴルーチンを起動し、ファイルのデコードを行い、
		// デコード結果をチャンネル経由でオーディオ処理ループに安全に渡します。
		go engine.mixer.LoadTrackAsync(mixer.DeckA, file) // 💡 修正: DeckAを識別する定数を渡す

		// 💡 修正: HTTP 応答は即座に返す (この修正を維持)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted) // 202 Accepted
		json.NewEncoder(w).Encode(map[string]string{"status": "loading started asynchronously", "file": file})
	})

	mux.HandleFunc("/api/deck/a/play", func(w http.ResponseWriter, r *http.Request) {
		engine.mixer.DeckA.Play()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "playing"})
	})

	mux.HandleFunc("/api/deck/a/pause", func(w http.ResponseWriter, r *http.Request) {
		engine.mixer.DeckA.Pause()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "paused"})
	})

	mux.HandleFunc("/api/deck/a/stop", func(w http.ResponseWriter, r *http.Request) {
		engine.mixer.DeckA.Stop()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "stopped"})
	})

	mux.HandleFunc("/api/deck/a/seek", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Position float64 `json:"position"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		engine.mixer.DeckA.Seek(req.Position)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":   "seeked",
			"position": req.Position,
		})
	})

	mux.HandleFunc("/api/deck/a/volume", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Volume float64 `json:"volume"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		engine.mixer.DeckA.SetVolume(req.Volume)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
			"volume": req.Volume,
		})
	})

	mux.HandleFunc("/api/deck/a/eq", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Low  float64 `json:"low"`
			Mid  float64 `json:"mid"`
			High float64 `json:"high"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		engine.mixer.DeckA.EQ.SetLow(req.Low)
		engine.mixer.DeckA.EQ.SetMid(req.Mid)
		engine.mixer.DeckA.EQ.SetHigh(req.High)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
			"eq":     req,
		})
	})

	mux.HandleFunc("/api/deck/a/filter", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Type      string  `json:"type"`
			Cutoff    float64 `json:"cutoff"`
			Resonance float64 `json:"resonance"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		switch req.Type {
		case "lowpass":
			engine.mixer.DeckA.Filter.SetLowpass(req.Cutoff, req.Resonance)
		case "highpass":
			engine.mixer.DeckA.Filter.SetHighpass(req.Cutoff, req.Resonance)
		case "none":
			engine.mixer.DeckA.Filter.Reset()
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	mux.HandleFunc("/api/deck/a/speed", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Speed float64 `json:"speed"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		engine.mixer.DeckA.SetSpeed(req.Speed)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
			"speed":  req.Speed,
		})
	})

	mux.HandleFunc("/api/deck/a/cuepoint/add", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Name  string `json:"name"`
			Color string `json:"color"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		engine.mixer.DeckA.AddCuePoint(req.Name, req.Color)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "added"})
	})

	mux.HandleFunc("/api/deck/a/loop/set", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Start float64 `json:"start"`
			End   float64 `json:"end"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		engine.mixer.DeckA.CueManager.SetLoop(req.Start, req.End)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "loop set"})
	})

	mux.HandleFunc("/api/deck/a/loop/enable", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Enabled bool `json:"enabled"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		engine.mixer.DeckA.CueManager.EnableLoop(req.Enabled)

		if req.Enabled {
			engine.mixer.DeckA.CueManager.ActivateLoop()
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "ok",
			"enabled": req.Enabled,
		})
	})

	// ========== Deck B API ==========

	mux.HandleFunc("/api/deck/b/load", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req LoadRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("❌ Deck B load: JSON decode error: %v", err)
			http.Error(w, "Invalid request body (JSON decode failed)", http.StatusBadRequest)
			return
		}

		file := req.File
		log.Printf("📥 Deck B load request received: %s", file)

		if file == "" {
			log.Printf("❌ Deck B load: file parameter missing in JSON")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "file parameter required"})
			return
		}

		if _, err := os.Stat(file); os.IsNotExist(err) {
			log.Printf("❌ Deck B load: file not found: %s", file)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "file not found"})
			return
		}

		log.Printf("⏳ Deck B: Starting ASYNC WAV loading for: %s", file)

		// 💡 修正: デッドロックを避けるため、チャンネル経由で安全にロード処理を依頼する
		// LoadTrackAsync は mixer パッケージ側での実装が必要になります。
		go engine.mixer.LoadTrackAsync(mixer.DeckB, file) // 💡 修正: DeckBを識別する定数を渡す

		// 💡 修正: HTTP 応答は即座に返す (この修正を維持)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted) // 202 Accepted
		json.NewEncoder(w).Encode(map[string]string{"status": "loading started asynchronously", "file": file})
	})

	mux.HandleFunc("/api/deck/b/play", func(w http.ResponseWriter, r *http.Request) {
		engine.mixer.DeckB.Play()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "playing"})
	})

	mux.HandleFunc("/api/deck/b/pause", func(w http.ResponseWriter, r *http.Request) {
		engine.mixer.DeckB.Pause()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "paused"})
	})

	mux.HandleFunc("/api/deck/b/stop", func(w http.ResponseWriter, r *http.Request) {
		engine.mixer.DeckB.Stop()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "stopped"})
	})

	mux.HandleFunc("/api/deck/b/seek", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Position float64 `json:"position"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		engine.mixer.DeckB.Seek(req.Position)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":   "seeked",
			"position": req.Position,
		})
	})

	mux.HandleFunc("/api/deck/b/volume", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Volume float64 `json:"volume"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		engine.mixer.DeckB.SetVolume(req.Volume)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
			"volume": req.Volume,
		})
	})

	mux.HandleFunc("/api/deck/b/eq", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Low  float64 `json:"low"`
			Mid  float64 `json:"mid"`
			High float64 `json:"high"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		engine.mixer.DeckB.EQ.SetLow(req.Low)
		engine.mixer.DeckB.EQ.SetMid(req.Mid)
		engine.mixer.DeckB.EQ.SetHigh(req.High)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
			"eq":     req,
		})
	})

	mux.HandleFunc("/api/deck/b/filter", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Type      string  `json:"type"`
			Cutoff    float64 `json:"cutoff"`
			Resonance float64 `json:"resonance"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		switch req.Type {
		case "lowpass":
			engine.mixer.DeckB.Filter.SetLowpass(req.Cutoff, req.Resonance)
		case "highpass":
			engine.mixer.DeckB.Filter.SetHighpass(req.Cutoff, req.Resonance)
		case "none":
			engine.mixer.DeckB.Filter.Reset()
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	mux.HandleFunc("/api/deck/b/speed", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Speed float64 `json:"speed"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		engine.mixer.DeckB.SetSpeed(req.Speed)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
			"speed":  req.Speed,
		})
	})

	mux.HandleFunc("/api/deck/b/cuepoint/add", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Name  string `json:"name"`
			Color string `json:"color"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		engine.mixer.DeckB.AddCuePoint(req.Name, req.Color)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "added"})
	})

	mux.HandleFunc("/api/deck/b/loop/set", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Start float64 `json:"start"`
			End   float64 `json:"end"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		engine.mixer.DeckB.CueManager.SetLoop(req.Start, req.End)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "loop set"})
	})

	mux.HandleFunc("/api/deck/b/loop/enable", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Enabled bool `json:"enabled"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		engine.mixer.DeckB.CueManager.EnableLoop(req.Enabled)

		if req.Enabled {
			engine.mixer.DeckB.CueManager.ActivateLoop()
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "ok",
			"enabled": req.Enabled,
		})
	})

	// ========== Mixer API ==========

	mux.HandleFunc("/api/mixer/crossfader", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Value float64 `json:"value"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		engine.mixer.SetCrossfader(req.Value)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
			"value":  req.Value,
		})
	})

	mux.HandleFunc("/api/mixer/master", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Volume float64 `json:"volume"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		engine.mixer.SetMasterVolume(req.Volume)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
			"volume": req.Volume,
		})
	})

	mux.HandleFunc("/api/mixer/sync", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Enabled bool   `json:"enabled"`
			Master  string `json:"master"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		engine.mixer.EnableSync(req.Enabled, req.Master)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "ok",
			"enabled": req.Enabled,
			"master":  req.Master,
		})
	})

	// ⚠️ HTTPポーリング用のStatus API（WebSocketへの移行により、バックアップとして残すか削除可能）
	mux.HandleFunc("/api/mixer/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(engine.mixer.GetStatus())
	})

	// ========== Utility API ==========

	mux.HandleFunc("/api/devices", func(w http.ResponseWriter, r *http.Request) {
		devices, err := portaudio.Devices()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		deviceList := make([]map[string]interface{}, len(devices))
		for i, device := range devices {
			deviceList[i] = map[string]interface{}{
				"id":                i,
				"name":              device.Name,
				"maxInputChannels":  device.MaxInputChannels,
				"maxOutputChannels": device.MaxOutputChannels,
				"defaultSampleRate": device.DefaultSampleRate,
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(deviceList)
	})

	fmt.Println("🎛️ Professional DJ Audio Engine v2.0")
	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Println("Server running on http://localhost:8080")
	fmt.Println("\nFeatures enabled:")
	fmt.Println(" ✅ 2-Deck System")
	fmt.Println(" ✅ 3-Band EQ")
	fmt.Println(" ✅ Hi/Low Pass Filters")
	fmt.Println(" ✅ BPM Detection & Sync")
	fmt.Println(" ✅ Cue Points & Loops")
	fmt.Println(" ✅ Pitch Control")
	fmt.Println(" ✅ WebSocket Status Stream")
	fmt.Println("\nPress Ctrl+C to stop")

	// =======================================================
	// 🚀 ステータス配信ゴルーチン (10ms間隔)
	// =======================================================
	// 🚀 ステータス配信ゴルーチン (20 FPS)
	go func() {
		log.Println("📡 Status broadcasting goroutine started.") // 💡 起動ログ
		const pingInterval = 10 * time.Second
		lastPing := time.Now()

		// 💡 50msだと負荷が高い可能性があるため、一旦 100ms (10FPS) で安定させます
		ticker := time.NewTicker(200 * time.Millisecond)
		defer ticker.Stop()

		// 💡 ログの間隔を管理
		logTicker := time.NewTicker(5 * time.Second)
		defer logTicker.Stop()

		for {
			select {
			case <-ticker.C:
				status := engine.mixer.GetStatus() // 💡 ここで止まっていないか確認が必要
				data, err := json.Marshal(status)
				if err != nil {
					continue
				}

				manager.RLock()
				clientsCount := len(manager.clients)
				if clientsCount == 0 {
					manager.RUnlock()
					continue
				}

				// クライアントリストのコピー作成
				var targetClients []*Client
				for _, c := range manager.clients {
					targetClients = append(targetClients, c)
				}
				manager.RUnlock()

				// 全クライアントに配信
				for _, c := range targetClients {
					c.safeWrite(websocket.TextMessage, data)
				}

				// 10秒おきに Ping
				if time.Since(lastPing) > pingInterval {
					for _, c := range targetClients {
						c.safeWrite(websocket.PingMessage, nil)
					}
					lastPing = time.Now()
				}

			case <-logTicker.C:
				// 💡 5秒おきに「生きている」ことをログに出す
				manager.RLock()
				log.Printf("📡 Status Heartbeat: sending to %d clients", len(manager.clients))
				manager.RUnlock()
			}
		}
	}()

	// サーバー起動時に、作成した mux を enableCORS でラップします
	log.Fatal(http.ListenAndServe(":8080", enableCORS(mux)))
}
