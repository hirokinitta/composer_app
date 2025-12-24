import { app, BrowserWindow, dialog, ipcMain } from 'electron';
import { join } from 'path'; 
// @electron-vite/plugin がインストールできないため、インポートは削除したまま

// 開発サーバーのURLを環境変数から取得
const VITE_DEV_SERVER_URL = process.env.VITE_DEV_SERVER_URL; 
// プリロードスクリプトのパスを環境変数から取得
const VITE_PRELOAD_PATH = process.env.VITE_PRELOAD_PATH; 

let mainWindow;

process.on('uncaughtException', function(err) {
  log.error('electron:event:uncaughtException');
  log.error(err);
  log.error(err.stack);
  app.quit();
});

function createWindow() {
  mainWindow = new BrowserWindow({
    width: 1400,
    height: 900,
    show: false, // 準備ができるまで隠す
    webPreferences: {
      // 修正: 開発モードでは VITE_PRELOAD_PATH を優先し、
      // 本番モードでのみ join を使用する。
      preload: VITE_PRELOAD_PATH 
        ? VITE_PRELOAD_PATH 
        // VITE_PRELOAD_PATHがない場合（本番ビルド後）、out/main/index.jsから見て
        // out/preload/index.js のパスを計算。
        : join(__dirname, '../preload/index.js'), 
      contextIsolation: true,
      nodeIntegration: false,
      sandbox: false
    }
  });

  mainWindow.on('ready-to-show', () => {
    mainWindow.show();
  });

  // 開発モードと本番モードのロード切り替え
  if (VITE_DEV_SERVER_URL) {
    mainWindow.loadURL(VITE_DEV_SERVER_URL);
    mainWindow.webContents.openDevTools();
  } else {
    // 本番ビルド時のパスはjoinを使用
    mainWindow.loadFile(join(__dirname, '../renderer/index.html'));
  }
}

// ファイル選択ダイアログ
ipcMain.handle('select-audio-file', async () => {
  if (!mainWindow) return null;
  
  const result = await dialog.showOpenDialog(mainWindow, {
    properties: ['openFile'],
    filters: [
      { name: 'Audio Files', extensions: ['wav', 'mp3', 'flac'] },
      { name: 'WAV Files', extensions: ['wav'] },
      { name: 'All Files', extensions: ['*'] }
    ],
    defaultPath: 'C:/composer-dj-app/go_audio_engine/testdata'
  });
  
  if (!result.canceled && result.filePaths.length > 0) {
    // Windowsのバックスラッシュをスラッシュに変換
    return result.filePaths[0].replace(/\\/g, '/');
  }
  return null;
});

app.whenReady().then(() => {
  createWindow();

  app.on('activate', () => {
    if (BrowserWindow.getAllWindows().length === 0) {
      createWindow();
    }
  });
});

app.on('window-all-closed', () => {
  if (process.platform !== 'darwin') {
    app.quit();
  }
});