export class AudioEngineClient {
    constructor(baseUrl = 'http://localhost:8080') {
        this.baseUrl = baseUrl;
    }

    async load(deck, filePath) {
        const response = await fetch(`${this.baseUrl}/api/deck/${deck.toLowerCase()}/load`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ file: filePath })
        });
        if (!response.ok) throw new Error(response.statusText);
        return response.json();
    }

    async play(deck) {
        const response = await fetch(`${this.baseUrl}/api/deck/${deck.toLowerCase()}/play`, {
            method: 'POST'
        });
        return response.json();
    }

    async pause(deck) {
        const response = await fetch(`${this.baseUrl}/api/deck/${deck.toLowerCase()}/pause`, {
            method: 'POST'
        });
        return response.json();
    }

    async stop(deck) {
        const response = await fetch(`${this.baseUrl}/api/deck/${deck.toLowerCase()}/stop`, {
            method: 'POST'
        });
        return response.json();
    }

    async setVolume(deck, volume) {
        const response = await fetch(`${this.baseUrl}/api/deck/${deck.toLowerCase()}/volume`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ volume })
        });
        return response.json();
    }

    async seek(deck, position) {
        const response = await fetch(`${this.baseUrl}/api/deck/${deck.toLowerCase()}/seek`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ position })
        });
        return response.json();
    }

    async setEQ(deck, eq) {
        const response = await fetch(`${this.baseUrl}/api/deck/${deck.toLowerCase()}/eq`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                low: eq.low || 0,
                mid: eq.mid || 0,
                high: eq.high || 0
            })
        });
        return response.json();
    }

    async setFilter(deck, filter) {
        const response = await fetch(`${this.baseUrl}/api/deck/${deck.toLowerCase()}/filter`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                type: filter.type || 'none',
                cutoff: filter.cutoff || 0.5,
                resonance: filter.resonance || 0
            })
        });
        return response.json();
    }

    async setSpeed(deck, speed) {
        const response = await fetch(`${this.baseUrl}/api/deck/${deck.toLowerCase()}/speed`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ speed })
        });
        return response.json();
    }

    async addCuePoint(deck, name, color) {
        const response = await fetch(`${this.baseUrl}/api/deck/${deck.toLowerCase()}/cuepoint/add`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ name, color })
        });
        return response.json();
    }

    async setLoop(deck, start, end) {
        const response = await fetch(`${this.baseUrl}/api/deck/${deck.toLowerCase()}/loop/set`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ start, end })
        });
        return response.json();
    }

    async enableLoop(deck, enabled) {
        const response = await fetch(`${this.baseUrl}/api/deck/${deck.toLowerCase()}/loop/enable`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ enabled })
        });
        return response.json();
    }

    async setCrossfader(value) {
        const response = await fetch(`${this.baseUrl}/api/mixer/crossfader`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ value })
        });
        return response.json();
    }

    async setMasterVolume(volume) {
        const response = await fetch(`${this.baseUrl}/api/mixer/master`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ volume })
        });
        return response.json();
    }

    async enableSync(enabled, master) {
        const response = await fetch(`${this.baseUrl}/api/mixer/sync`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ enabled, master: master.toLowerCase() })
        });
        return response.json();
    }

    async getMixerStatus() {
        const response = await fetch(`${this.baseUrl}/api/mixer/status`);
        if (!response.ok) {
            throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }
        return response.json();
    }

    async checkConnection() {
        try {
            await this.getMixerStatus();
            return true;
        } catch {
            return false;
        }
    }
}