import { upsertDownload, setDownloadError } from '../store/downloadStore';

class WebSocketService {
    constructor() {
        this.socket = null;
        this.reconnectInterval = 3000;
        this.shouldReconnect = true;
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

        switch (type) {
            case 'progress':
                upsertDownload(payload);
                break;
            case 'error':
                setDownloadError(payload);
                break;
            default:
                console.warn('Unknown message type:', type);
        }
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
