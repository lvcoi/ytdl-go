import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import wsService from './websocket';
import { downloadStore, setDownloadStore } from '../store/downloadStore';

// Mock the store functions to prevent side effects during testing
vi.mock('../store/downloadStore', () => ({
    upsertDownload: vi.fn(),
    setDownloadError: vi.fn(),
    downloadStore: { activeDownloads: {} },
    setDownloadStore: vi.fn(),
}));

describe('WebSocketService', () => {
    let mockSocket;

    beforeEach(() => {
        // Mock the global WebSocket object
        mockSocket = {
            send: vi.fn(),
            close: vi.fn(),
            readyState: WebSocket.OPEN,
            addEventListener: vi.fn(),
            removeEventListener: vi.fn(),
        };
        global.WebSocket = vi.fn(() => mockSocket);
        global.WebSocket.OPEN = 1;
        global.WebSocket.CONNECTING = 0;
        global.WebSocket.CLOSED = 3;
        
        // Reset the service state if needed (it's a singleton, so we might need to be careful)
        // Since it's a singleton exported as default, we're testing the same instance.
        // Ideally we'd reset its internal state, but we can't easily access private fields.
        // However, we can clear listeners.
        if (wsService.listeners) {
            wsService.listeners.clear();
        }
    });

    afterEach(() => {
        vi.clearAllMocks();
    });

    it('should have addListener method', () => {
        expect(typeof wsService.addListener).toBe('function');
    });

    it('should register and call listeners on dispatch', () => {
        const listener = vi.fn();
        wsService.addListener(listener);

        const message = { type: 'test-event', payload: { data: 123 } };
        wsService.dispatch(message);

        expect(listener).toHaveBeenCalledWith(message);
    });

    it('should return a cleanup function that removes the listener', () => {
        const listener = vi.fn();
        const cleanup = wsService.addListener(listener);

        expect(typeof cleanup).toBe('function');

        // Verify listener is active
        wsService.dispatch({ type: 'event-1' });
        expect(listener).toHaveBeenCalledTimes(1);

        // Cleanup
        cleanup();

        // Verify listener is removed
        wsService.dispatch({ type: 'event-2' });
        expect(listener).toHaveBeenCalledTimes(1); // Should not increase
    });

    it('should dispatch core events to the store (upsertDownload)', async () => {
        const { upsertDownload } = await import('../store/downloadStore');
        
        const payload = { id: '1', percent: 50 };
        wsService.dispatch({ type: 'progress', payload });

        expect(upsertDownload).toHaveBeenCalledWith(payload);
    });

    it('should dispatch error events to the store (setDownloadError)', async () => {
        const { setDownloadError } = await import('../store/downloadStore');
        
        const payload = { id: '1', message: 'Failed' };
        wsService.dispatch({ type: 'error', payload });

        expect(setDownloadError).toHaveBeenCalledWith(payload);
    });
});
