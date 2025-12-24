const { app, BrowserWindow, dialog, ipcMain } = require('electron');

let mainWindow;

function createWindow() {
  mainWindow = new BrowserWindow({
    width: 1400,
    height: 900,
    webPreferences: {
      nodeIntegration: true,
      contextIsolation: false
    }
  });

  // Viteの開発サーバーに接続
  mainWindow.loadURL('http://localhost:5173');
  mainWindow.webContents.openDevTools();
}

// ファイル選択ダイアログのIPC
ipcMain.handle('select-audio-file', async () => {
  if (!mainWindow) return null;
  
  const result = await dialog.showOpenDialog(mainWindow, {
    properties: ['openFile'],
    filters: [
      { name: 'WAV Files', extensions: ['wav'] },
      { name: 'MP3 Files', extensions: ['mp3'] },
      { name: 'All Audio', extensions: ['wav', 'mp3', 'flac'] }
    ],
    defaultPath: 'C:\\\\composer-dj-app\\\\go_audio_engine\\\\testdata'
  });
  
  if (!result.canceled && result.filePaths.length > 0) {
    // Windowsパスをスラッシュに変換
    return result.filePaths[0].replace(/\\\\/g, '/');
  }
  return null;
});

app.whenReady().then(createWindow);

app.on('window-all-closed', () => {
  if (process.platform !== 'darwin') {
    app.quit();
  }
});

app.on('activate', () => {
  if (BrowserWindow.getAllWindows().length === 0) {
    createWindow();
  }
});