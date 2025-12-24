import './app.css';
// @ts-ignore
import App from './App.svelte';

let app;
try {
    console.log('🚀 main.js: Mounting App...');
    app = new App({
        target: document.getElementById('app'),
    });
    console.log('✅ main.js: App mounted successfully.');
} catch (e) {
    console.error('❌ main.js: Failed to mount App:', e);
}

export default app;