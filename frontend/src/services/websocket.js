import { upsertDownload, setDownloadError } from '../store/downloadStore';

class WebSocketService {
    constructor() {
        this.socket = null;
        this.reconnectInterval = 3000;
        this.shouldReconnect = true;
        this.listeners = new Set();
    }

    addListener(callback) {
        this.listeners.add(callback);
        return () => this.listeners.delete(callback);
    }

    connect() {
        if (this.socket && (this.socket.readyState === WebSocket.OPEN || this.socket.readyState === WebSocket.CONNECTING)) {
            return;
        }

        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const host = window.location.host;
        const url = `${protocol}//${host}/ws`;

        this.socket = new WebSocket(url);

        this.socket.onopen = () => {
            console.log('WebSocket connected');
        };

        this.socket.onmessage = (event) => {
            try {
                const message = JSON.parse(event.data);
                this.dispatch(message);
            } catch (error) {
                console.error('Failed to parse WebSocket message:', error);
            }
        };

        this.socket.onclose = () => {
            console.log('WebSocket disconnected');
            if (this.shouldReconnect) {
                setTimeout(() => {
                    console.log('Reconnecting...');
                    this.connect();
                }, this.reconnectInterval);
            }
        };

        this.socket.onerror = (error) => {
            console.error('WebSocket error:', error);
            this.socket.close();
        };
    }

    dispatch(message) {
        const { type, payload } = message;

        // 1. Update Global Store (Single Source of Truth)
        switch (type) {
            case 'progress':
                upsertDownload(payload);
                break;
            case 'error':
                setDownloadError(payload);
                break;
            default:
                // Other types handled by listeners
                break;
        }

        // 2. Notify Listeners (e.g., useDownloadManager for logs/toasts)
        this.listeners.forEach((listener) => {
            try {
                listener(message);
            } catch (err) {
                console.error('Error in WebSocket listener:', err);
            }
        });
    }

    close() {
        this.shouldReconnect = false;
        if (this.socket) {
            this.socket.close();
        }
    }
}

// Singleton instance
const wsService = new WebSocketService();
export default wsService;
