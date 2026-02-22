import { Show } from 'solid-js';
import Icon from './Icon';

export default function Player(props) {
  const media = () => (typeof props.media === 'function' ? props.media() : props.media);
  const isVideo = () => !!media()?.filename?.match(/\.(mp4|webm|mov|avi|mkv)$/i);
  const isAudio = () => !!media()?.filename?.match(/\.(mp3|wav|ogg|flac|m4a)$/i);

  return (
    <div class="fixed bottom-10 right-10 w-[28rem] glass rounded-[2.5rem] border border-white/10 shadow-2xl p-8 space-y-5 z-50 animate-in slide-in-from-bottom-10 duration-500">
      <div class="flex items-center justify-between">
        <span class="text-[10px] font-black text-blue-500 uppercase tracking-widest">Now Playing</span>
        <button onClick={props.onClose} class="text-gray-500 hover:text-white">
          <Icon name="x" class="w-4 h-4" />
        </button>
      </div>

      <div class="w-full aspect-video bg-black rounded-3xl border border-white/5 overflow-hidden flex items-center justify-center">
        <Show when={isVideo()}>
          <video controls autoplay class="w-full h-full object-contain" src={media()?.url}></video>
        </Show>
        <Show when={isAudio()}>
          <div class="w-full h-full flex flex-col items-center justify-center bg-gradient-to-br from-blue-600/20 to-green-600/20">
            <Icon name="music" class="w-12 h-12 text-blue-400/70 mb-3" />
            <audio controls autoplay class="w-[88%]" src={media()?.url}></audio>
          </div>
        </Show>
        <Show when={!isVideo() && !isAudio()}>
          <div class="text-center">
            <Icon name="play-circle" class="w-12 h-12 text-blue-600/50 mx-auto mb-2" />
            <p class="text-xs text-gray-500">Unsupported media type</p>
          </div>
        </Show>
      </div>

      <div class="space-y-1">
        <div class="font-bold text-white truncate">{media()?.title || 'Unknown title'}</div>
        <div class="text-xs text-gray-500 font-medium">{media()?.artist || 'Unknown artist'}</div>
      </div>
    </div>
  );
}
