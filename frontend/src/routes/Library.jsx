import { createMemo, onMount, onCleanup, createEffect } from 'solid-js';
import { useAppStore } from '../store/appStore';
import { useSavedPlaylists } from '../hooks/useSavedPlaylists';
import { usePlayerController } from '../hooks/usePlayerController';
import LibraryView from '../components/LibraryView';
import { useLibrarySync } from '../hooks/useLibrarySync';

export function Library() {
    const { state, setState } = useAppStore();
    const {
        createPlaylist,
        renamePlaylist,
        deletePlaylist,
        assignPlaylist
    } = useSavedPlaylists();
    const { openPlayer } = usePlayerController();
    const { retryLibraryMetadataScan } = useLibrarySync();

    const handleOpenPlayer = (item, queueItems) => {
        openPlayer(item, queueItems);
    };

    return (
        <LibraryView
            downloads={() => state.library.downloads}
            section={() => state.library.section}
            viewMode={() => state.library.viewMode}
            typeFilter={() => state.library.typeFilter}
            navPath={() => state.library.navPath}
            filters={() => state.library.filters}
            sortKey={() => state.library.sortKey}
            uiState={() => state.library.ui}
            savedPlaylists={() => state.library.savedPlaylists}
            playlistAssignments={() => state.library.playlistAssignments}
            onSectionChange={(nextSection) => setState('library', 'section', nextSection)}
            onViewModeChange={(nextViewMode) => setState('library', 'viewMode', nextViewMode)}
            onTypeFilterChange={(nextTypeFilter) => setState('library', 'typeFilter', nextTypeFilter)}
            onNavPathChange={(nextNavPath) => setState('library', 'navPath', nextNavPath)}
            onFilterChange={(filterKey, value) => setState('library', 'filters', filterKey, value)}
            onClearFilters={() => setState('library', 'filters', {
                query: '',
                savedPlaylistId: '',
            })}
            onSortKeyChange={(nextSortKey) => setState('library', 'sortKey', nextSortKey)}
            onUiStateChange={(key, value) => setState('library', 'ui', key, value)}
            openPlayer={handleOpenPlayer}
            onCreateSavedPlaylist={createPlaylist}
            onRenameSavedPlaylist={renamePlaylist}
            onDeleteSavedPlaylist={deletePlaylist}
            onAssignSavedPlaylist={assignPlaylist}
            onRetryMetadataScan={retryLibraryMetadataScan}
        />
    );
}
