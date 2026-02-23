import { onMount, onCleanup } from 'solid-js';
import { useAppStore } from '../store/appStore';

export function useGlobalShortcuts() {
    const { state, setState } = useAppStore();

    const submitDuplicateChoice = async (choice) => {
        const prompt = state.download.duplicateQueue[0];
        if (!prompt) return;
        setState('download', 'duplicateError', '');
        try {
            const res = await fetch('/api/download/duplicate-response', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    jobId: prompt.jobId,
                    promptId: prompt.promptId,
                    choice,
                }),
            });
            const data = await res.json().catch(() => ({}));
            if (!res.ok || data.error) {
                if (res.status === 404 || res.status === 409) {
                    setState('download', 'duplicateQueue', (prev) => prev.filter((item) => item.promptId !== prompt.promptId));
                    setState('download', 'duplicateError', '');
                    return;
                }
                setState('download', 'duplicateError', data.error || 'Failed to submit duplicate choice');
                return;
            }
            setState('download', 'duplicateQueue', (prev) => prev.filter((item) => item.promptId !== prompt.promptId));
            setState('download', 'duplicateError', '');
        } catch (e) {
            setState('download', 'duplicateError', e.message || 'Failed to submit duplicate choice');
        }
    };

    const handleDuplicateShortcut = (e) => {
        if (state.download.duplicateQueue.length === 0) return;

        // Check if user is typing in an input or textarea
        const activeElement = document.activeElement;
        const isTyping = activeElement && (
            activeElement.tagName === 'INPUT' ||
            activeElement.tagName === 'TEXTAREA' ||
            activeElement.isContentEditable
        );

        if (isTyping) return;
        if (e.metaKey || e.ctrlKey || e.altKey) return;

        switch (e.key) {
            case 'o': e.preventDefault(); submitDuplicateChoice('overwrite'); break;
            case 'O': e.preventDefault(); submitDuplicateChoice('overwrite_all'); break;
            case 's': e.preventDefault(); submitDuplicateChoice('skip'); break;
            case 'S': e.preventDefault(); submitDuplicateChoice('skip_all'); break;
            case 'r': e.preventDefault(); submitDuplicateChoice('rename'); break;
            case 'R': e.preventDefault(); submitDuplicateChoice('rename_all'); break;
            case 'q': e.preventDefault(); submitDuplicateChoice('cancel'); break;
            default: break;
        }
    };

    onMount(() => {
        if (typeof window !== 'undefined') {
            window.addEventListener('keydown', handleDuplicateShortcut);
        }
    });

    onCleanup(() => {
        if (typeof window !== 'undefined') {
            window.removeEventListener('keydown', handleDuplicateShortcut);
        }
    });

    return { submitDuplicateChoice };
}
