import { createMemo } from 'solid-js';
import { useAppStore } from '../store/appStore';
import { buildLibraryModel } from '../utils/libraryModel';

export function useLibraryModel() {
    const { state } = useAppStore();

    return createMemo(() => buildLibraryModel({
        downloads: state.library.downloads,
        savedPlaylists: state.library.savedPlaylists,
        playlistAssignments: state.library.playlistAssignments,
        typeFilter: state.library.typeFilter,
        sortKey: state.library.sortKey,
        filters: state.library.filters,
    }));
}
