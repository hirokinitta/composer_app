const { contextBridge, ipcRenderer } = require('electron');

/**
 * contextBridgeを使用して、レンダラープロセス（Svelte）に
 * 必要な機能のみを安全に公開します。
 */
contextBridge.exposeInMainWorld('electronAPI', {
    /**
     * メインプロセスにオーディオファイル選択を要求し、ファイルパスを取得する。
     * @returns {Promise<string|null>} 選択されたファイルパス、またはキャンセルされた場合は null
     */
    selectAudioFile: () => {
        // メインプロセスに 'dialog:select-audio-file' チャンネルでメッセージを送信し、
        // 結果を非同期で待ち受ける。
        return ipcRenderer.invoke('dialog:select-audio-file');
    },

    /**
     * メインプロセスがGoサーバーの状態（例：起動、エラー）を通知するためのイベントリスナー。
     * @param {function} callback - (status: 'starting'|'ready'|'error') を受け取るコールバック関数
     */
    onGoServerStatus: (callback) => {
        // メインプロセスから 'go-server-status' チャンネルでメッセージを受信
        ipcRenderer.on('go-server-status', (event, status) => {
            callback(status);
        });
    },

    // 💡 追加：Deck A/B のファイルダイアログを開くためのAPI
    openFile: (deckId) => ipcRenderer.invoke('open-file', deckId),
		});

console.log('✅ Electron Preload script executed. window.electronAPI is available.');