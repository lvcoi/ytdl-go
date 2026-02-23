import { describe, it, expect, beforeEach } from 'vitest';
import { downloadStore, setDownloadStore, upsertDownload, setDownloadError } from './downloadStore';

describe('downloadStore', () => {
    beforeEach(() => {
        // Reset store state
        setDownloadStore('activeDownloads', {});
        setDownloadStore('error', null);
    });

    it('upsertDownload updates progress and status', () => {
        const payload = {
            id: 'video_1',
            filename: 'Test Video.mp4',
            percent: 50.5,
            status: 'downloading',
            eta: '30s'
        };

        upsertDownload(payload);

        const task = downloadStore.activeDownloads['video_1'];
        expect(task).toBeDefined();
        expect(task.percent).toBe(50.5);
        expect(task.status).toBe('downloading');
        expect(task.done).toBe(false);
    });

    it('upsertDownload marks done when percent is 100', () => {
        const payload = {
            id: 'video_2',
            percent: 100,
            status: 'downloading'
        };

        upsertDownload(payload);

        const task = downloadStore.activeDownloads['video_2'];
        expect(task.done).toBe(true);
    });

    it('setDownloadError sets error state', () => {
        // First create a task
        upsertDownload({ id: 'video_3', percent: 10 });

        const errorPayload = {
            id: 'video_3',
            message: 'Network error',
            code: 500
        };

        setDownloadError(errorPayload);

        const task = downloadStore.activeDownloads['video_3'];
        expect(task.status).toBe('error');
        expect(task.error).toBe('Network error');
        
        // Also check global error if your implementation sets it
        // expect(downloadStore.error).toEqual(errorPayload);
    });
});
