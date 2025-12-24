interface Window {
  electronAPI?: {
    selectAudioFile: () => Promise<string | null>;
  }
}