import { describe, it, expect } from 'vitest';
import { normalizeDownloadStatus } from './downloadStatus';

describe('normalizeDownloadStatus', () => {
    it('returns empty string for non-string input', () => {
        expect(normalizeDownloadStatus(null)).toBe('');
        expect(normalizeDownloadStatus(undefined)).toBe('');
        expect(normalizeDownloadStatus(123)).toBe('');
    });

    it('returns valid statuses as is', () => {
        expect(normalizeDownloadStatus('queued')).toBe('queued');
        expect(normalizeDownloadStatus('running')).toBe('running');
        expect(normalizeDownloadStatus('reconnecting')).toBe('reconnecting');
        expect(normalizeDownloadStatus('complete')).toBe('complete');
        expect(normalizeDownloadStatus('error')).toBe('error');
    });

    it('maps "done" to "complete"', () => {
        expect(normalizeDownloadStatus('done')).toBe('complete');
        expect(normalizeDownloadStatus('DONE')).toBe('complete');
    });

    it('returns empty string for invalid statuses', () => {
        expect(normalizeDownloadStatus('invalid')).toBe('');
        expect(normalizeDownloadStatus('finished')).toBe(''); // Unless we add it
    });
});
