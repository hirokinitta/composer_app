import { contextBridge } from 'electron'
import { electronAPI } from '@electron-toolkit/preload'

// メインプロセスの機能をレンダラープロセス（Svelte）に安全に公開する
if (process.contextIsolated) {
  try {
    // 'window.electron' としてElectronの基本APIを公開
    contextBridge.exposeInMainWorld('electron', electronAPI)
    
    // 'window.api' としてカスタムAPIを公開（必要に応じて追加）
    contextBridge.exposeInMainWorld('api', {
      // 例: バージョン情報の取得など
      // getVersion: () => ipcRenderer.invoke('get-version')
    })
  } catch (error) {
    console.error(error)
  }
} else {
  // contextIsolationが無効な場合のフォールバック
  window.electron = electronAPI
  window.api = {}
}