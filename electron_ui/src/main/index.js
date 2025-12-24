import { app, BrowserWindow, dialog, ipcMain } from 'electron';
import path from 'path';
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
let mainWindow;

function createWindow() {
  mainWindow = new BrowserWindow({
    width: 1400,
    height: 900,
    webPreferences: {
      preload: path.join(__dirname, '../preload/index.mjs'),
      contextIsolation: true,
      nodeIntegration: false
    }
  });

  mainWindow.loadURL('http://localhost:5173');
  mainWindow.webContents.openDevTools();
}

ipcMain.handle('select-audio-file', async () => {
  if (!mainWindow) return null;
  
  const result = await dialog.showOpenDialog(mainWindow, {
    properties: ['openFile'],
    filters: [
      { name: 'WAV Files', extensions: ['wav'] },
      { name: 'All Files', extensions: ['*'] }
    ],
    defaultPath: 'C:/composer-dj-app/go_audio_engine/testdata'
  });
  
  if (!result.canceled && result.filePaths.length > 0) {
    return result.filePaths[0].replace(/\\\\/g, '/');
  }
  return null;
});

app.whenReady().then(createWindow);
app.on('window-all-closed', () => {
  if (process.platform !== 'darwin') app.quit();
});
app.on('activate', () => {
  if (BrowserWindow.getAllWindows().length === 0) createWindow();
});