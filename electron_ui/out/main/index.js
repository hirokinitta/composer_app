import { ipcMain, app, BrowserWindow, dialog } from "electron";
import path from "path";
import { fileURLToPath } from "url";
import __cjs_url__ from "node:url";
import __cjs_path__ from "node:path";
import __cjs_mod__ from "node:module";
const __filename = __cjs_url__.fileURLToPath(import.meta.url);
const __dirname = __cjs_path__.dirname(__filename);
const require2 = __cjs_mod__.createRequire(import.meta.url);
const __dirname$1 = path.dirname(fileURLToPath(import.meta.url));
let mainWindow;
function createWindow() {
  mainWindow = new BrowserWindow({
    width: 1400,
    height: 900,
    webPreferences: {
      preload: path.join(__dirname$1, "../preload/index.mjs"),
      contextIsolation: true,
      nodeIntegration: false
    }
  });
  mainWindow.loadURL("http://localhost:5173");
  mainWindow.webContents.openDevTools();
}
ipcMain.handle("select-audio-file", async () => {
  if (!mainWindow) return null;
  const result = await dialog.showOpenDialog(mainWindow, {
    properties: ["openFile"],
    filters: [
      { name: "WAV Files", extensions: ["wav"] },
      { name: "All Files", extensions: ["*"] }
    ],
    defaultPath: "C:/composer-dj-app/go_audio_engine/testdata"
  });
  if (!result.canceled && result.filePaths.length > 0) {
    return result.filePaths[0].replace(/\\\\/g, "/");
  }
  return null;
});
app.whenReady().then(createWindow);
app.on("window-all-closed", () => {
  if (process.platform !== "darwin") app.quit();
});
app.on("activate", () => {
  if (BrowserWindow.getAllWindows().length === 0) createWindow();
});
