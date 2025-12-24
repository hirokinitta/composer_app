// src/lib/store.js
import { writable } from 'svelte/store';

/**
 * WebSocket接続の状態を保持するストア
 * @type {import('svelte/store').Writable<'disconnected' | 'connecting' | 'connected' | 'error'>}
 */
export const wsStatus = writable('disconnected');
// 💡 デバッグ用: ストアの変更をログに出力
wsStatus.subscribe(val => console.log(`📡 [Store] wsStatus changed to: ${val}`));

/**
 * Goバックエンドから受信したアプリケーションの全体ステータスを保持するストア
 * Deck A/B、Mixerなどの情報を含む
 * @type {import('svelte/store').Writable<object>}
 */
export const appStatus = writable({
    deckA: { FilePath: '', IsPlaying: false, Position: 0, Duration: 0, Speed: 1.0, Volume: 1.0, BPM: 0, EQ: {}, CuePoints: [], loop: {} },
    deckB: { FilePath: '', IsPlaying: false, Position: 0, Duration: 0, Speed: 1.0, Volume: 1.0, BPM: 0, EQ: {}, CuePoints: [], loop: {} },
    mixer: { crossfader: 0.5, masterVolume: 1.0, syncEnabled: false, syncMaster: 'a' }
});