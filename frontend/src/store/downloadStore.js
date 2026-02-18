import { createStore } from 'solid-js/store';

const [downloadStore, setDownloadStore] = createStore({
    activeDownloads: {}, // Keyed by task ID { id, filename, percent, status, eta, done }
    error: null,         // { id, message, code }
});

const upsertDownload = (payload) => {
    setDownloadStore('activeDownloads', payload.id, (prev) => ({
        ...prev,
        ...payload,
        // Calculate done state based on status or percent
        done: payload.status === 'complete' || payload.percent >= 100
    }));
};

const setDownloadError = (payload) => {
    // Set global error or per-item error? 
    // The prompt says "Visuals: Error State: If status === 'error', turn the bar Red and display payload.message."
    // This implies per-item error state is useful, but prompt also says "Actions: setDownloadError(payload): Sets status to 'error' and stores the error message."
    
    // We update the item status to error AND store the error details
    setDownloadStore('activeDownloads', payload.id, (prev) => ({
        ...prev,
        status: 'error',
        error: payload.message
    }));
    
    // Also set global error if needed, but per-item seems more robust for the UI requirement
    setDownloadStore('error', payload);
};

export { downloadStore, setDownloadStore, upsertDownload, setDownloadError };
