const { app, BrowserWindow, ipcMain, dialog } = require('electron');
const path = require('path');
const { spawn } = require('child_process');

// -------------------------------------------------------------------
// 1. グローバル変数とGoサーバープロセスの管理
// -------------------------------------------------------------------
let mainWindow;
let goServerProcess;
const IS_DEV = process.env.NODE_ENV === 'development';
const GO_SERVER_PORT = 8080;

/**
 * Goサーバー（バイナリ）のパスを取得
 * 開発モードでは go run、本番ビルドでは同梱されたバイナリを使用
 */
function getGoServerPath() {
    if (IS_DEV) {
        // 開発モード: Goソースファイルを直接実行
        return path.join(__dirname, '..', 'go_audio_engine', 'main.go');
    }
    // 本番モード: Electronのresourceディレクトリからバイナリを取得
    const serverName = (process.platform === 'win32' ? 'go_audio_engine.exe' : 'go_audio_engine');
    return path.join(process.resourcesPath, 'app.asar.unpacked', 'go_audio_engine', serverName);
}

/**
 * Goサーバーを起動する
 * @param {BrowserWindow} win - 状態通知のための BrowserWindow インスタンス
 */
function startGoServer(win) {
    const serverPath = getGoServerPath();
    console.log(`Starting Go Server at: ${serverPath}`);
    win.webContents.send('go-server-status', 'starting');

    if (IS_DEV) {
        // 開発モード: 'go run' を使用
        goServerProcess = spawn('go', ['run', serverPath], {
            cwd: path.join(__dirname, '..', '..', 'go_audio_engine'),
            // 💡 追加: Windows環境で 'go' コマンドがPATHに見つからない問題を解決するため、シェル経由で実行
            shell: true
        });
    } else {
        // 本番モード: バイナリを直接実行
        goServerProcess = spawn(serverPath, {
            shell: true
        });
    }
    
    // サーバーの標準出力をログに出力
    goServerProcess.stdout.on('data', (data) => {
        const output = data.toString().trim();
        console.log(`[GO_SERVER]: ${output}`);
        
        // サーバーが正常に起動したことを検知 (特定のログメッセージをチェック)
        if (output.includes('Server running on http://localhost:8080')) {
            console.log('✅ Go Server is ready for connection.');
            win.webContents.send('go-server-status', 'ready');
        }
    });

    // サーバーのエラー出力をログに出力
    goServerProcess.stderr.on('data', (data) => {
        console.error(`[GO_SERVER_ERR]: ${data.toString().trim()}`);
    });

    // プロセス終了時の処理
    goServerProcess.on('close', (code) => {
        console.log(`❌ Go Server process exited with code ${code}`);
        goServerProcess = null;
        if (code !== 0) {
            win.webContents.send('go-server-status', 'error');
        }
    });

    goServerProcess.on('error', (err) => {
        console.error(`❌ Failed to start Go Server process: ${err}`);
        goServerProcess = null;
        win.webContents.send('go-server-status', 'error');
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
        width: 1200,
        height: 800,
        minWidth: 1000,
        minHeight: 700,
        title: "Professional DJ Audio Mixer",
        webPreferences: {
            preload: path.join(__dirname, 'preload.js'), // プリロードスクリプトを適用
            nodeIntegration: false,
            contextIsolation: true // セキュリティ強化
        },
    });

    // Goサーバーを起動
    // 💡 修正: mainWindowが準備できてからGoサーバーを起動する
    mainWindow.webContents.on('did-finish-load', () => {
        startGoServer(mainWindow);
    });

    // 開発モードではViteサーバーを、本番ではビルドされたファイルをロード
    if (IS_DEV) {
        // Vite開発サーバーのURL
        mainWindow.loadURL('http://localhost:5173');
        mainWindow.webContents.openDevTools();
    } else {
        // 本番ビルドのパス
        // 💡 注意: 'public' ではなく 'dist' ディレクトリを参照するように変更
        // Viteのデフォルトビルド出力先は 'dist' です
        mainWindow.loadFile(path.join(__dirname, '..', 'dist', 'index.html'));
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
ipcMain.handle('dialog:select-audio-file', async () => {
    console.log('IPC: Received request for audio file selection.');
    
    // ダイアログを開く
    const { canceled, filePaths } = await dialog.showOpenDialog(mainWindow, {
        title: 'Select Audio Track (WAV)',
        properties: ['openFile'],
        filters: [
            { name: 'Audio Files (WAV)', extensions: ['wav'] }
        ]
    });

    if (canceled || filePaths.length === 0) {
        return null;
    }
    
    const filePath = filePaths[0];
    console.log(`IPC: Selected file path: ${filePath}`);
    return filePath;
});

// -------------------------------------------------------------------
// 4. その他の設定
// -------------------------------------------------------------------

// macOSでアプリがドックアイコンをクリックされたときにウィンドウを再作成する
app.on('activate', () => {
    if (BrowserWindow.getAllWindows().length === 0) createWindow();
});