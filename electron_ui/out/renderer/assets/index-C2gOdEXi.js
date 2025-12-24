function getAugmentedNamespace(n) {
  if (n.__esModule) return n;
  var f = n.default;
  if (typeof f == "function") {
    var a = function a2() {
      if (this instanceof a2) {
        return Reflect.construct(f, arguments, this.constructor);
      }
      return f.apply(this, arguments);
    };
    a.prototype = f.prototype;
  } else a = {};
  Object.defineProperty(a, "__esModule", { value: true });
  Object.keys(n).forEach(function(k) {
    var d = Object.getOwnPropertyDescriptor(n, k);
    Object.defineProperty(a, k, d.get ? d : {
      enumerable: true,
      get: function() {
        return n[k];
      }
    });
  });
  return a;
}
const __viteBrowserExternal = {};
const __viteBrowserExternal$1 = /* @__PURE__ */ Object.freeze(/* @__PURE__ */ Object.defineProperty({
  __proto__: null,
  default: __viteBrowserExternal
}, Symbol.toStringTag, { value: "Module" }));
const require$$1 = /* @__PURE__ */ getAugmentedNamespace(__viteBrowserExternal$1);
var define_process_env_default = {};
const fs = require$$1;
const path = require$$1;
const pathFile = path.join(__dirname, "path.txt");
function getElectronPath() {
  let executablePath;
  if (fs.existsSync(pathFile)) {
    executablePath = fs.readFileSync(pathFile, "utf-8");
  }
  if (define_process_env_default.ELECTRON_OVERRIDE_DIST_PATH) {
    return path.join(define_process_env_default.ELECTRON_OVERRIDE_DIST_PATH, executablePath || "electron");
  }
  if (executablePath) {
    return path.join(__dirname, "dist", executablePath);
  } else {
    throw new Error("Electron failed to install correctly, please delete node_modules/electron and try installing again");
  }
}
var electron = getElectronPath();
let mainWindow;
function createWindow() {
  mainWindow = new electron.BrowserWindow({
    width: 1400,
    height: 900,
    webPreferences: {
      nodeIntegration: true,
      contextIsolation: false
    }
  });
  mainWindow.loadURL("http://localhost:5173");
  mainWindow.webContents.openDevTools();
}
electron.ipcMain.handle("select-audio-file", async () => {
  const result = await electron.dialog.showOpenDialog(mainWindow, {
    properties: ["openFile"],
    filters: [
      { name: "WAV Files", extensions: ["wav"] },
      { name: "All Files", extensions: ["*"] }
    ],
    defaultPath: "C:\\\\composer-dj-app\\\\go_audio_engine\\\\testdata"
  });
  if (!result.canceled && result.filePaths.length > 0) {
    return result.filePaths[0].replace(/\\\\/g, "/");
  }
  return null;
});
electron.app.whenReady().then(createWindow);
electron.app.on("window-all-closed", () => {
  if (process.platform !== "darwin") electron.app.quit();
});
