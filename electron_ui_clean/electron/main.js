const { app, BrowserWindow, dialog, ipcMain } = require('electron');
const path = require('path');
const { spawn, execSync } = require('child_process');

// -------------------------------------------------------------------
// 1. グローバル変数とGoサーバープロセスの管理
// -------------------------------------------------------------------
let mainWindow;
let goServerProcess;
const IS_DEV = !app.isPackaged;
const GO_SERVER_PORT = 8080;

/**
 * Goサーバー（バイナリ）のパスを取得
 * 開発モードでは go run、本番ビルドでは同梱されたバイナリを使用
 */
function getGoServerPath() {
    if (IS_DEV) {
        // 開発モード: Goソースファイルを直接実行
        return path.join(__dirname, '..', '..', 'go_audio_engine', 'main.go');
    }
    // 本番モード: Electronのresourceディレクトリからバイナリを取得
    const serverName = (process.platform === 'win32' ? 'go_audio_engine.exe' : 'go_audio_engine');
    return path.join(process.resourcesPath, 'app.asar.unpacked', 'go_audio_engine', serverName);
}

/**
 * ポート8080を占有しているプロセスがあれば強制終了する (Windows専用)
 * これにより "Address already in use" エラーを防ぎます
 */
function killPort8080() {
    if (process.platform === 'win32') {
        try {
            const pidsToKill = new Set();
            // ポート8080を使用しているプロセスを特定
            const output = execSync('netstat -ano | findstr :8080').toString();
            const lines = output.split('\r\n');
            lines.forEach(line => {
                if (line.includes('LISTENING')) {
                    const parts = line.trim().split(/\s+/);
                    const pid = parts[parts.length - 1];
                    if (pid && /^\d+$/.test(pid)) pidsToKill.add(pid);
                }
            });

            pidsToKill.forEach(pid => {
                console.log(`⚠️ Port 8080 is in use by PID ${pid}. Killing it...`);
                // プロセスが既になくなっている場合のエラーを無視するため、個別にtry-catch
                try { execSync(`taskkill /F /PID ${pid}`); } catch (killError) { /* ignore */ }
            });
        } catch (e) {
            // プロセスが見つからない場合はエラーになるが無視
        }
    }
}

/**
 * Goサーバーを起動する
 * @param {BrowserWindow} win - 状態通知のための BrowserWindow インスタンス
 */
function startGoServer(win) {
    // 💡 修正: サーバーが既に起動している場合は再起動しない
    // これにより、リロード時の接続切断やゾンビプロセスの発生を防ぎます
    if (goServerProcess) {
        console.log('⚠️ Go Server is already running. Skipping start.');
        if (win && win.webContents) {
            win.webContents.send('go-server-status', 'ready');
            
            // 💡 追加: サーバーが既に起動済みで、まだロード画面ならUIをロードする
            if (IS_DEV) {
                win.loadURL('http://localhost:5173');
            } else {
                win.loadFile(path.join(__dirname, '..', 'dist', 'index.html'));
            }
        }
        return;
    }

    // 起動前にゾンビプロセスを一掃
    killPort8080();

    const serverPath = getGoServerPath();
    console.log(`Starting Go Server at: ${serverPath}`);
    if (win && win.webContents) {
        win.webContents.send('go-server-status', 'starting');
    }

    if (IS_DEV) {
        // 開発モード: 'go run' を使用
        goServerProcess = spawn('go', ['run', serverPath], {
            cwd: path.join(__dirname, '..', '..', 'go_audio_engine'),
            shell: true
        });
    } else {
        // 本番モード: バイナリを直接実行
        goServerProcess = spawn(serverPath, {
            shell: true
        });
    }
    
    // サーバーの標準出力をログに出力
    let serverLogBuffer = ''; // 💡 追加: ログのバッファリング用
    goServerProcess.stdout.on('data', (data) => {
        serverLogBuffer += data.toString();
        
        // 行ごとに分割して処理
        let lines = serverLogBuffer.split('\n');
        serverLogBuffer = lines.pop(); // 最後の不完全な行をバッファに戻す

        lines.forEach(line => {
            const output = line.trim();
            if (output) console.log(`[GO_SERVER]: ${output}`);
            
            // サーバーが正常に起動したことを検知し、UIをロードする
            if (win && win.webContents && output.includes('Server running on http://localhost:8080')) {
                console.log('✅ Go Server is ready for connection. Loading UI...');
                if (IS_DEV) {
                    win.loadURL('http://localhost:5173');
                } else {
                    win.loadFile(path.join(__dirname, '..', 'dist', 'index.html'));
                }
            }
        });
    });

    // 💡 追加: 安全策 - 10秒経っても起動ログが検知できない場合は強制的にUIをロードする
    setTimeout(() => {
        if (win && !win.isDestroyed() && win.webContents.getURL().includes('loading.html')) {
            console.log('⚠️ Server start detection timed out. Forcing UI load...');
            const loadUrl = IS_DEV ? 'http://localhost:5173' : path.join(__dirname, '..', 'dist', 'index.html');
            IS_DEV ? win.loadURL(loadUrl) : win.loadFile(loadUrl);
        }
    }, 10000);

    // サーバーのエラー出力をログに出力
    goServerProcess.stderr.on('data', (data) => {
        console.error(`[GO_SERVER_ERR]: ${data.toString().trim()}`);
    });

    // プロセス終了時の処理
    goServerProcess.on('close', (code) => {
        console.log(`❌ Go Server process exited with code ${code}`);
        goServerProcess = null;
        if (win && win.webContents && code !== 0) {
            win.webContents.send('go-server-status', 'error');
        }
    });

    goServerProcess.on('error', (err) => {
        console.error(`❌ Failed to start Go Server process: ${err}`);
        goServerProcess = null;
        if (win && win.webContents) {
            win.webContents.send('go-server-status', 'error');
        }
        dialog.showErrorBox('Server Error', `Failed to start the Go Audio Engine. Check console for details: ${err.message}`);
    });
}

/**
 * Goサーバープロセスを終了する
 */
function stopGoServer() {
    if (goServerProcess) {
        console.log('Killing Go Server process...');
        // プロセスを正常に終了させる (プラットフォーム依存の処理)
        if (process.platform === 'win32') {
            const pid = goServerProcess.pid;
            spawn('taskkill', ['/pid', pid, '/f', '/t']);
        } else {
            goServerProcess.kill();
        }
        goServerProcess = null;
    }
}

// -------------------------------------------------------------------
// 2. Electron アプリケーションのライフサイクル
// -------------------------------------------------------------------

function createWindow() {
  mainWindow = new BrowserWindow({
    width: 1600,
    height: 1000,
    webPreferences: {
      preload: path.join(__dirname, '..', 'electron', 'preload.js'), // 💡 パスを修正
      contextIsolation: true,
      nodeIntegration: false,
    }
  });

    // Goサーバーを起動
    startGoServer(mainWindow);

    // 最初にローディング画面を表示する
    mainWindow.loadFile(path.join(__dirname, 'loading.html'));

    // 開発モードではViteサーバーを、本番ではビルドされたファイルをロード
    if (IS_DEV) {
        // UIのロードはサーバー起動後に行うため、ここではDevToolsを開くだけ
        mainWindow.webContents.openDevTools();
    }
}

// アプリが準備完了したらウィンドウを作成
app.whenReady().then(() => {
    createWindow();

    app.on('activate', function () {
        if (BrowserWindow.getAllWindows().length === 0) createWindow();
    });
});

// 全てのウィンドウが閉じられたらアプリを終了し、Goサーバーも停止
app.on('window-all-closed', function () {
  if (process.platform !== 'darwin') {
    stopGoServer();
    app.quit();
  }
});

// アプリ終了直前のイベントでもサーバーを確実に停止
app.on('before-quit', () => {
    stopGoServer();
});

// -------------------------------------------------------------------
// 3. IPC (プロセス間通信) ハンドラ
// -------------------------------------------------------------------

/**
 * ファイル選択ダイアログを開くIPCハンドラ
 */
ipcMain.handle('select-audio-file', async () => {
    console.log('IPC: Received request for audio file selection.');
    
    // ダイアログを開く
    const { canceled, filePaths } = await dialog.showOpenDialog(mainWindow, {
        title: 'Select Audio Track',
        properties: ['openFile'],
        filters: [
            { name: 'Audio Files', extensions: ['wav', 'mp3'] }
        ]
    });

    if (canceled || filePaths.length === 0) {
        return null;
    }
    
    const filePath = filePaths[0];
    console.log(`IPC: Selected file path: ${filePath}`);
    return filePath;
});