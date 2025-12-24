// src/lib/wants.js

/**
 * Goバックエンドに送信するクライアント側の要求 (WANTS) メッセージを定義します。
 * @param {string} command - 実行したいコマンド名 (例: 'loadTrack', 'play', 'crossfade')
 * @param {object} payload - コマンドに付随するデータ
 * @returns {string} JSON文字列
 */
export function buildWantMessage(command, payload = {}) {
    // ログのためにコンソールに出力
    console.log('WANT:', command, payload); 

    const message = {
        type: 'WANT', // メッセージタイプはWANT (要求)
        command: command,
        payload: payload,
    };
    
    return JSON.stringify(message);
}