import { writable, get } from 'svelte/store';

// モジュールスコープ変数
let pingInterval = null;
let reconnectTimer = null;
let shouldReconnect = true;

// 💡 TypeScriptエラー回避用: windowをany型として扱う
/** @type {any} */
const globalScope = window;

// 💡 ストアのシングルトン化: HMR(ホットリロード)でストアが複数作成されるのを防ぐ
// これにより、UIとWebSocketの接続状態が常に同期されます
if (!globalScope.__dj_stores) {
    globalScope.__dj_stores = {
        id: Math.random().toString(36).substring(7), // ストア識別用ID
        wsStatus: writable('disconnected'),
        mixerStatus: writable(null)
    };
    console.log('✨ Creating singleton stores. ID:', globalScope.__dj_stores.id);
} else {
    console.log('♻️ Reusing existing stores. ID:', globalScope.__dj_stores.id);
}

export const wsStatus = globalScope.__dj_stores.wsStatus;
export const mixerStatus = globalScope.__dj_stores.mixerStatus;

/**
 * 現在の接続状態を確認し、接続済みならストアを更新する
 * @returns {boolean} 接続済みなら true
 */
export function checkCurrentConnection() {
    if (globalScope.__dj_ws && globalScope.__dj_ws.readyState === WebSocket.OPEN) {
        wsStatus.set('connected');
        setupEventHandlers(globalScope.__dj_ws); // ハンドラを再設定してストア更新を確実にする
        return true;
    }
    return false;
}

/**
 * ストアの現在のデータを直接取得する (購読タイミング問題の回避用)
 */
export function getMixerData() {
    // 💡 修正: 常に最新のグローバルストアから取得する
    return get(globalScope.__dj_stores.mixerStatus);
}

/**
 * ストアを安全に購読するヘルパー関数
 */
export function subscribeToMixerStatus(callback) {
    return globalScope.__dj_stores.mixerStatus.subscribe(callback);
}

export function subscribeToWsStatus(callback) {
    return globalScope.__dj_stores.wsStatus.subscribe(callback);
}

/**
 * WebSocket接続を開始する
 */
export function startWebSocketConnection() {
    // 🛡️ シングルトンガード: グローバル変数に接続があるかチェック
    // これにより、HMR（ホットリロード）でファイルが再読み込みされても重複接続を防ぎます
    if (globalScope.__dj_ws) {
        const existingWs = globalScope.__dj_ws;
        if (existingWs.readyState === WebSocket.OPEN) {
            const currentData = get(mixerStatus);
            
            // 💡 修正: 接続があってもデータ(ストア)が空なら、初期データを取り直すために再接続する
            if (!currentData) {
                existingWs.onclose = null; // 💡 追加: 閉じる際の自動再接続トリガーを防止
                try { existingWs.close(); } catch(e){}
                globalScope.__dj_ws = null;
                // このまま下の新規接続ロジックへ進む
            } else {
                wsStatus.set('connected');
                setupEventHandlers(existingWs); // イベントハンドラを現在のストアに付け直す
                return;
            }
        } else if (existingWs.readyState === WebSocket.CONNECTING) {
            return;
        } else {
            // 死んでいる接続ならクリーンアップ
            try { existingWs.close(); } catch(e){}
            globalScope.__dj_ws = null;
        }
    }

    if (reconnectTimer) {
        clearTimeout(reconnectTimer);
        reconnectTimer = null;
    }

    shouldReconnect = true;
    const url = 'ws://127.0.0.1:8080/ws/status'; // 💡 修正: WindowsでのIPv4/IPv6重複接続を防ぐためIP指定
    console.log('🔄 Connecting to:', url);
    wsStatus.set('connecting');

    try {
        // グローバル変数に格納
        globalScope.__dj_ws = new WebSocket(url);
        setupEventHandlers(globalScope.__dj_ws);
    } catch (err) {
        console.error("❌ Failed to create WebSocket:", err);
        retryConnection();
    }
}

// イベントハンドラのセットアップ関数（再利用可能にする）
function setupEventHandlers(socket) {
    // 既存のリスナーを無効化（二重発火防止）はWebSocket仕様上できないが、上書きはされる
    
    socket.onopen = () => {
        console.log('✅ WebSocket Connected');
        wsStatus.set('connected'); // ここでストアが更新されるはず

        // 接続成功時に古いPingタイマーをクリアし、新しいものを設定
        // Ping (Keep-alive) 5秒ごとに短縮（反応を良くする）
        if (pingInterval) clearInterval(pingInterval);
        pingInterval = setInterval(() => {
            if (socket.readyState === WebSocket.OPEN) {
                socket.send(JSON.stringify({ type: 'ping' }));
            }
        }, 5000);
    };

    socket.onmessage = (event) => {
        console.log('📩 WS Message received. Length:', event.data ? event.data.length : 0);
        if (!event.data) return;
        try {
            const rawData = JSON.parse(event.data);
            if (rawData.type === 'pong') return;

            // 🛡️ 安全装置: データが来ているなら、ステータスは絶対に 'connected' であるはず
            // 画面が 'disconnected' になっていたら強制的に直す
            if (get(wsStatus) !== 'connected') {
                wsStatus.set('connected');
            }

            // 💡 デバッグ: データセット直前のログ
            console.log('🚚 websocket.js: Setting mixerStatus store (Global Ref). Data keys:', Object.keys(rawData));
            
            // タイムスタンプを付与して、内容が同じでも確実に更新検知させる
            // 💡 修正: モジュール変数のmixerStatusではなく、グローバル参照を直接使用して更新する
            globalScope.__dj_stores.mixerStatus.set({ ...rawData, _timestamp: Date.now() });

        } catch (error) {
            console.error('❌ JSON Parse Error:', error);
        }
    };

    socket.onerror = (error) => {
        // エラーログは出すが、再接続はoncloseに任せる
        console.warn('⚠️ WebSocket Error');
    };

    socket.onclose = (event) => {
        // 明示的に閉じた場合以外はログを出す
        if (shouldReconnect) {
            console.log(`🔌 WebSocket disconnected (Code: ${event.code})...`);
            wsStatus.set('disconnected');
            
            // windowの参照も消す
            if (globalScope.__dj_ws === socket) {
                globalScope.__dj_ws = null;
            }

            retryConnection();
        }
    };
}

function retryConnection() {
    if (!shouldReconnect) return;
    
    if (!reconnectTimer) {
        console.log('⏳ Retrying in 2s...');
        reconnectTimer = setTimeout(() => {
            reconnectTimer = null;
            startWebSocketConnection();
        }, 2000);
    }
}

export function closeWebSocketConnection() {
    shouldReconnect = false;
    
    if (pingInterval) clearInterval(pingInterval);
    if (reconnectTimer) clearTimeout(reconnectTimer);
    
    if (globalScope.__dj_ws) {
        // 再接続しないようにoncloseを無効化
        globalScope.__dj_ws.onclose = null; 
        globalScope.__dj_ws.close();
        globalScope.__dj_ws = null;
    }
    
    wsStatus.set('disconnected');
    console.log('🛑 WebSocket connection closed explicitly.');
}

// 💡 追加: ページリロード/終了時に確実に接続を閉じる
// これにより、サーバー側で "Total: 2" のようなゾンビ接続が残るのを防ぎます
window.addEventListener('beforeunload', () => {
    shouldReconnect = false;
    if (globalScope.__dj_ws) {
        console.log('🛑 Page unloading, closing WebSocket immediately.');
        globalScope.__dj_ws.close();
    }
});