import Icon from './Icon';
import Tooltip from './Tooltip';
import { useQueueManager } from '../hooks/useQueueManager';

export default function PlaylistActions(props) {
    const { playPlaylist, addPlaylistToQueue } = useQueueManager();

    const handlePlay = (e) => {
        e.stopPropagation();
        playPlaylist(props.tracks);
    };

    const handleAddToQueue = (e) => {
        e.stopPropagation();
        addPlaylistToQueue(props.tracks);
    };

    return (
        <div class="flex items-center gap-1.5">
            {/* Play Playlist */}
            <Tooltip text="Play Playlist">
                <button
                    onClick={handlePlay}
                    aria-label={`Play ${props.playlistName || 'playlist'}`}
                    class="p-2 rounded-xl text-gray-400 hover:text-emerald-400
                            hover:bg-emerald-500/10 transition-all duration-150
                            focus:outline-none focus:ring-1 focus:ring-emerald-500/30
                           active:scale-95"
                >
                    <Icon name="play-circle" class="w-5 h-5" />
                </button>
            </Tooltip>

            {/* Add All to Queue */}
            <Tooltip text="Add to Queue">
                <button
                    onClick={handleAddToQueue}
                    aria-label={`Add ${props.playlistName || 'playlist'} to queue`}
                    class="p-2 rounded-xl text-gray-400 hover:text-blue-400
                            hover:bg-blue-500/10 transition-all duration-150
                            focus:outline-none focus:ring-1 focus:ring-blue-500/30
                           active:scale-95"
                >
                    <Icon name="circle-plus" class="w-5 h-5" />
                </button>
            </Tooltip>
        </div>
    );
}
