// electoron_ui/main.js

// 1. すべて import 構文であること
import { app, BrowserWindow } from 'electron';
import path from 'path';
import { spawn } from 'child_process'; 
import { fileURLToPath } from 'url';

// 2. ESM環境での __dirname の再構築
const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

let goProcess = null;

function createWindow() {
    const mainWindow = new BrowserWindow({
        width: 1000,
        height: 700,
        webPreferences: {
            // Note: 以前のエラー回避のため、preloadはコメントアウトを維持
            // nodeIntegration: true,
            // contextIsolation: false
        }
    });

    if (process.env.VITE_DEV_SERVER_URL) {
        mainWindow.loadURL(process.env.VITE_DEV_SERVER_URL);
    } else {
        mainWindow.loadFile(path.join(__dirname, 'dist', 'index.html'));
    }
    
    // 開発中にDevToolsを開いておく
    mainWindow.webContents.openDevTools();
}

// ... (Go Engine起動ロジック、app.on('will-quit') などは省略)

app.whenReady().then(() => {
    createWindow();
    // ...
});

app.on('window-all-closed', () => {
    if (process.platform !== 'darwin') {
        app.quit();
    }
});