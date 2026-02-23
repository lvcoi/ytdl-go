import { createSignal, onCleanup } from 'solid-js';
import { useAppStore } from '../store/appStore';

const MAX_SAVED_PLAYLIST_NAME_LENGTH = 80;
const SAVED_PLAYLISTS_ENDPOINT = '/api/library/playlists';
const SAVED_PLAYLISTS_MIGRATION_ENDPOINT = '/api/library/playlists/migrate';
const SAVED_PLAYLISTS_MIGRATION_KEY = 'ytdl-go:saved-playlists:backend-migration:v1';

const normalizeSavedPlaylistName = (value) => (
    String(value || '')
        .trim()
        .replace(/\s+/g, ' ')
        .slice(0, MAX_SAVED_PLAYLIST_NAME_LENGTH)
);
const normalizeSavedPlaylistId = (value) => String(value || '').trim();
const normalizeMediaKey = (value) => String(value || '').trim();

const normalizeSavedPlaylistEntry = (value) => {
    const source = value && typeof value === 'object' ? value : {};
    const id = normalizeSavedPlaylistId(source.id);
    const name = normalizeSavedPlaylistName(source.name);
    if (id === '' || name === '') {
        return null;
    }
    return {
        id,
        name,
        createdAt: typeof source.createdAt === 'string' ? source.createdAt.trim() : '',
        updatedAt: typeof source.updatedAt === 'string' ? source.updatedAt.trim() : '',
    };
};

const normalizeSavedPlaylists = (value) => {
    if (!Array.isArray(value)) {
        return [];
    }
    const out = [];
    const seenIds = new Set();
    const seenNames = new Set();
    for (const entry of value) {
        const normalized = normalizeSavedPlaylistEntry(entry);
        if (!normalized) {
            continue;
        }
        if (seenIds.has(normalized.id)) {
            continue;
        }
        const nameKey = normalized.name.toLowerCase();
        if (seenNames.has(nameKey)) {
            continue;
        }
        seenIds.add(normalized.id);
        seenNames.add(nameKey);
        out.push(normalized);
    }
    return out;
};

const normalizePlaylistAssignments = (value, validPlaylistIds) => {
    const source = value && typeof value === 'object' ? value : {};
    const out = {};
    for (const [mediaKey, playlistId] of Object.entries(source)) {
        const normalizedMediaKey = normalizeMediaKey(mediaKey);
        const normalizedPlaylistId = normalizeSavedPlaylistId(playlistId);
        if (normalizedMediaKey === '' || normalizedPlaylistId === '' || !validPlaylistIds.has(normalizedPlaylistId)) {
            continue;
        }
        out[normalizedMediaKey] = normalizedPlaylistId;
    }
    return out;
};

const normalizeSavedPlaylistStatePayload = (value) => {
    const source = value && typeof value === 'object' ? value : {};
    const playlists = normalizeSavedPlaylists(source.playlists);
    const validPlaylistIds = new Set(playlists.map((playlist) => playlist.id));
    return {
        playlists,
        assignments: normalizePlaylistAssignments(source.assignments, validPlaylistIds),
    };
};

const hasSavedPlaylistStateData = (value) => (
    Array.isArray(value?.playlists) && value.playlists.length > 0
) || (
        value?.assignments && typeof value.assignments === 'object' && Object.keys(value.assignments).length > 0
    );

const hasSavedPlaylistNameConflict = (playlists, playlistName, excludedId = '') => (
    playlists.some((playlist) => (
        playlist.id !== excludedId &&
        playlist.name.localeCompare(playlistName, undefined, { sensitivity: 'base' }) === 0
    ))
);

const createSavedPlaylistId = () => {
    if (typeof globalThis !== 'undefined' && globalThis.crypto && typeof globalThis.crypto.randomUUID === 'function') {
        return globalThis.crypto.randomUUID();
    }
    return `saved-${Date.now()}-${Math.random().toString(36).slice(2, 10)}`;
};

const responseErrorMessage = async (response, fallbackMessage) => {
    try {
        const payload = await response.json();
        if (payload && typeof payload.error === 'string' && payload.error.trim() !== '') {
            return payload.error.trim();
        }
    } catch (error) {
        // Ignore JSON parse issues and use fallback.
    }
    return fallbackMessage;
};

const isAbortError = (error) => (
    error && typeof error === 'object' && error.name === 'AbortError'
);

export function useSavedPlaylists() {
    const { state, setState } = useAppStore();
    const [initError, setInitError] = createSignal('');

    // Use refs (local variables in closure) for non-reactive state to avoid unnecessary re-renders
    let initAbortController = null;
    let persistAbortController = null;
    let mutationVersion = 0;
    let mutationQueue = Promise.resolve();
    let isDisposed = false;

    const getSnapshot = () => normalizeSavedPlaylistStatePayload({
        playlists: state.library.savedPlaylists,
        assignments: state.library.playlistAssignments,
    });

    const applyState = (nextValue) => {
        const normalized = normalizeSavedPlaylistStatePayload(nextValue);
        const validIds = new Set(normalized.playlists.map((playlist) => playlist.id));
        setState('library', 'savedPlaylists', normalized.playlists);
        setState('library', 'playlistAssignments', normalized.assignments);

        // Clear filter if the selected playlist is deleted
        setState('library', 'filters', 'savedPlaylistId', (current) => (
            validIds.has(normalizeSavedPlaylistId(current)) ? current : ''
        ));
    };

    const fetchState = async (signal) => {
        const response = await fetch(SAVED_PLAYLISTS_ENDPOINT, { signal });
        if (!response.ok) {
            throw new Error(await responseErrorMessage(response, 'Unable to load saved playlists from backend.'));
        }
        const payload = await response.json();
        return normalizeSavedPlaylistStatePayload(payload);
    };

    const persistState = async (nextState, signal) => {
        const normalized = normalizeSavedPlaylistStatePayload(nextState);
        const response = await fetch(SAVED_PLAYLISTS_ENDPOINT, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(normalized),
            signal,
        });
        if (!response.ok) {
            throw new Error(await responseErrorMessage(response, 'Unable to persist saved playlists.'));
        }
        const payload = await response.json();
        return normalizeSavedPlaylistStatePayload(payload);
    };

    const migrateState = async (legacyState, signal) => {
        const normalizedLegacy = normalizeSavedPlaylistStatePayload(legacyState);
        const response = await fetch(SAVED_PLAYLISTS_MIGRATION_ENDPOINT, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(normalizedLegacy),
            signal,
        });
        if (!response.ok) {
            throw new Error(await responseErrorMessage(response, 'Unable to migrate saved playlists.'));
        }
        const payload = await response.json();
        return {
            state: normalizeSavedPlaylistStatePayload(payload),
            migrated: Boolean(payload && payload.migrated),
        };
    };

    const persistAndApply = async (nextState) => {
        mutationVersion += 1;
        const currentVersion = mutationVersion;

        if (persistAbortController) {
            persistAbortController.abort();
        }
        const controller = new AbortController();
        persistAbortController = controller;

        try {
            const persisted = await persistState(nextState, controller.signal);
            if (!isDisposed && currentVersion === mutationVersion) {
                applyState(persisted);
                setInitError('');
            }
            return persisted;
        } finally {
            if (persistAbortController === controller) {
                persistAbortController = null;
            }
        }
    };

    const enqueueMutation = (mutateFn) => {
        const next = mutationQueue.then(mutateFn, mutateFn);
        mutationQueue = next.catch(() => { });
        return next;
    };

    const hasCompletedMigration = () => {
        if (typeof window === 'undefined') return false;
        try {
            return window.localStorage.getItem(SAVED_PLAYLISTS_MIGRATION_KEY) === '1';
        } catch (error) {
            return false;
        }
    };

    const markMigrationComplete = () => {
        if (typeof window === 'undefined') return;
        try {
            window.localStorage.setItem(SAVED_PLAYLISTS_MIGRATION_KEY, '1');
        } catch (error) {
            // Ignore
        }
    };

    const initialize = async () => {
        if (initAbortController) {
            initAbortController.abort();
        }
        const controller = new AbortController();
        initAbortController = controller;
        const startVersion = mutationVersion;
        const legacySnapshot = getSnapshot();

        try {
            const backendState = await fetchState(controller.signal);
            if (isDisposed || controller.signal.aborted || mutationVersion !== startVersion) {
                return;
            }
            applyState(backendState);
            setInitError('');

            if (hasSavedPlaylistStateData(backendState)) {
                markMigrationComplete();
                return;
            }
            if (!hasSavedPlaylistStateData(legacySnapshot)) {
                if (!hasCompletedMigration()) {
                    markMigrationComplete();
                }
                return;
            }

            const migration = await migrateState(legacySnapshot, controller.signal);
            if (isDisposed || controller.signal.aborted || mutationVersion !== startVersion) {
                return;
            }
            applyState(migration.state);
            if (migration.migrated || hasSavedPlaylistStateData(migration.state)) {
                markMigrationComplete();
            }
        } catch (error) {
            if (isAbortError(error) || (controller.signal && controller.signal.aborted)) {
                return;
            }
            const message = error instanceof Error ? error.message : 'Failed to initialize saved playlists.';
            setInitError(message);
            console.error('Failed to initialize saved playlists:', error);
        } finally {
            if (initAbortController === controller) {
                initAbortController = null;
            }
        }
    };

    const createPlaylist = (rawName) => enqueueMutation(async () => {
        const normalizedName = normalizeSavedPlaylistName(rawName);
        if (normalizedName === '') {
            return { ok: false, error: 'Playlist name is required.' };
        }

        const current = getSnapshot();
        if (hasSavedPlaylistNameConflict(current.playlists, normalizedName)) {
            return { ok: false, error: 'A playlist with that name already exists.' };
        }

        const timestamp = new Date().toISOString();
        const createdPlaylist = {
            id: createSavedPlaylistId(),
            name: normalizedName,
            createdAt: timestamp,
            updatedAt: timestamp,
        };

        try {
            const persisted = await persistAndApply({
                playlists: [...current.playlists, createdPlaylist],
                assignments: current.assignments,
            });
            const persistedPlaylist = persisted.playlists.find((playlist) => playlist.id === createdPlaylist.id);
            if (!persistedPlaylist) {
                return { ok: false, error: 'A playlist with that name already exists.' };
            }
            return { ok: true, playlist: persistedPlaylist };
        } catch (error) {
            if (isAbortError(error)) return { ok: false, error: 'Request canceled.' };
            return { ok: false, error: error instanceof Error ? error.message : 'Unable to create saved playlist.' };
        }
    });

    const renamePlaylist = (playlistId, rawName) => enqueueMutation(async () => {
        const normalizedPlaylistId = normalizeSavedPlaylistId(playlistId);
        const normalizedName = normalizeSavedPlaylistName(rawName);
        if (normalizedPlaylistId === '') return { ok: false, error: 'Playlist not found.' };
        if (normalizedName === '') return { ok: false, error: 'Playlist name is required.' };

        const current = getSnapshot();
        const exists = current.playlists.some((playlist) => playlist.id === normalizedPlaylistId);
        if (!exists) return { ok: false, error: 'Playlist not found.' };
        if (hasSavedPlaylistNameConflict(current.playlists, normalizedName, normalizedPlaylistId)) {
            return { ok: false, error: 'A playlist with that name already exists.' };
        }

        const updatedAt = new Date().toISOString();
        const nextPlaylists = current.playlists.map((playlist) => (
            playlist.id === normalizedPlaylistId
                ? { ...playlist, name: normalizedName, updatedAt }
                : playlist
        ));

        try {
            const persisted = await persistAndApply({
                playlists: nextPlaylists,
                assignments: current.assignments,
            });
            const persistedPlaylist = persisted.playlists.find((playlist) => playlist.id === normalizedPlaylistId);
            if (!persistedPlaylist) return { ok: false, error: 'Playlist not found.' };
            if (persistedPlaylist.name.localeCompare(normalizedName, undefined, { sensitivity: 'base' }) !== 0) {
                return { ok: false, error: 'A playlist with that name already exists.' };
            }
            return { ok: true };
        } catch (error) {
            if (isAbortError(error)) return { ok: false, error: 'Request canceled.' };
            return { ok: false, error: error instanceof Error ? error.message : 'Unable to rename saved playlist.' };
        }
    });

    const deletePlaylist = (playlistId) => enqueueMutation(async () => {
        const normalizedPlaylistId = normalizeSavedPlaylistId(playlistId);
        if (normalizedPlaylistId === '') return { ok: false, error: 'Playlist not found.' };

        const current = getSnapshot();
        const nextPlaylists = current.playlists.filter((playlist) => playlist.id !== normalizedPlaylistId);
        if (nextPlaylists.length === current.playlists.length) {
            return { ok: false, error: 'Playlist not found.' };
        }

        const nextAssignments = {};
        for (const [mediaKey, assignedPlaylistId] of Object.entries(current.assignments)) {
            if (assignedPlaylistId === normalizedPlaylistId) continue;
            nextAssignments[mediaKey] = assignedPlaylistId;
        }

        try {
            await persistAndApply({
                playlists: nextPlaylists,
                assignments: nextAssignments,
            });
            return { ok: true };
        } catch (error) {
            if (isAbortError(error)) return { ok: false, error: 'Request canceled.' };
            return { ok: false, error: error instanceof Error ? error.message : 'Unable to delete saved playlist.' };
        }
    });

    const assignPlaylist = (mediaKey, playlistId) => enqueueMutation(async () => {
        const normalizedMediaKey = normalizeMediaKey(mediaKey);
        const normalizedPlaylistId = normalizeSavedPlaylistId(playlistId);
        if (normalizedMediaKey === '') return;

        const current = getSnapshot();
        const playlistExists = normalizedPlaylistId === '' || current.playlists.some(
            (playlist) => playlist.id === normalizedPlaylistId,
        );
        if (!playlistExists) return;

        const nextAssignments = { ...current.assignments };
        if (normalizedPlaylistId === '') {
            if (!(normalizedMediaKey in nextAssignments)) return;
            delete nextAssignments[normalizedMediaKey];
        } else {
            if (nextAssignments[normalizedMediaKey] === normalizedPlaylistId) return;
            nextAssignments[normalizedMediaKey] = normalizedPlaylistId;
        }

        try {
            await persistAndApply({
                playlists: current.playlists,
                assignments: nextAssignments,
            });
        } catch (error) {
            if (isAbortError(error)) return;
            console.error('Failed to assign saved playlist:', error);
        }
    });

    onCleanup(() => {
        isDisposed = true;
        if (initAbortController) initAbortController.abort();
        if (persistAbortController) persistAbortController.abort();
    });

    return {
        initError,
        initialize,
        createPlaylist,
        renamePlaylist,
        deletePlaylist,
        assignPlaylist,
    };
}
