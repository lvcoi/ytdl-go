import { Show, createMemo, createSignal } from 'solid-js';
import Icon from './Icon';

const firstNonEmpty = (...values) => {
  for (const value of values) {
    if (typeof value === 'string') {
      const trimmed = value.trim();
      if (trimmed !== '') {
        return trimmed;
      }
    }
  }
  return '';
};

const normalizeMediaType = (value) => (String(value || '').toLowerCase() === 'audio' ? 'audio' : 'video');
const isAudioFile = (filename) => /\.(mp3|wav|ogg|flac|m4a|opus|aac|alac|wma)$/i.test(String(filename || ''));
const isVideoFile = (filename) => /\.(mp4|webm|mov|avi|mkv|m4v|wmv|mpeg|mpg|ts|ogv|3gp)$/i.test(String(filename || ''));

export default function Player(props) {
  const media = createMemo(() => (typeof props.media === 'function' ? props.media() : props.media));
  const metadata = createMemo(() => (media()?.metadata && typeof media().metadata === 'object' ? media().metadata : {}));
  const mediaType = createMemo(() => {
    const filename = String(media()?.filename || '').trim();
    if (filename !== '') {
      if (isAudioFile(filename)) {
        return 'audio';
      }
      if (isVideoFile(filename)) {
        return 'video';
      }
      return 'unknown';
    }
    const declaredType = typeof media()?.type === 'string' ? media().type.trim() : '';
    return declaredType !== '' ? normalizeMediaType(declaredType) : 'unknown';
  });
  const isVideo = createMemo(() => mediaType() === 'video');
  const isAudio = createMemo(() => mediaType() === 'audio');
  const titleLabel = createMemo(() => firstNonEmpty(media()?.title, metadata().title, 'Unknown title'));
  const creatorLabel = createMemo(() => firstNonEmpty(
    media()?.artist,
    metadata().artist,
    metadata().author,
    'Unknown artist',
  ));
  const collectionLabel = createMemo(() => firstNonEmpty(
    media()?.album,
    metadata().album,
    creatorLabel(),
    'Unknown collection',
  ));
  const sourcePlaylistLabel = createMemo(() => firstNonEmpty(
    media()?.playlist?.title,
    metadata().playlist?.title,
    'Standalone',
  ));
  const typeLabel = createMemo(() => (
    mediaType() === 'audio'
      ? 'Audio'
      : mediaType() === 'video'
        ? 'Video'
        : 'Unknown'
  ));
  const thumbnailSrc = createMemo(() => firstNonEmpty(
    media()?.thumbnail_url,
    media()?.thumbnailUrl,
    media()?.thumbnailURL,
    metadata().thumbnail_url,
    metadata().thumbnailUrl,
    metadata().thumbnailURL,
  ));
  const [failedThumbnailSrc, setFailedThumbnailSrc] = createSignal('');
  const showThumbnail = createMemo(() => thumbnailSrc() !== '' && failedThumbnailSrc() !== thumbnailSrc());
  const fallbackIconName = createMemo(() => {
    if (mediaType() === 'audio') {
      return 'music';
    }
    if (mediaType() === 'video') {
      return 'film';
    }
    return 'play-circle';
  });
  const fallbackIconClass = createMemo(() => {
    if (mediaType() === 'audio') {
      return 'w-8 h-8 text-blue-300/70';
    }
    if (mediaType() === 'video') {
      return 'w-8 h-8 text-emerald-300/70';
    }
    return 'w-8 h-8 text-gray-400/70';
  });

  return (
    <div class="fixed bottom-6 right-6 w-[32rem] max-w-[calc(100vw-1.5rem)] glass rounded-[2rem] border border-white/10 shadow-2xl p-6 space-y-5 z-50 animate-in slide-in-from-bottom-10 duration-500">
      <div class="flex items-center justify-between">
        <span class="text-[10px] font-black text-blue-500 uppercase tracking-widest">Now Playing</span>
        <button onClick={props.onClose} class="text-gray-500 hover:text-white">
          <Icon name="x" class="w-4 h-4" />
        </button>
      </div>

      <div class="flex items-start gap-4 min-w-0">
        <div class="w-24 h-24 shrink-0 rounded-2xl border border-white/10 bg-[#0b0f1a] overflow-hidden flex items-center justify-center">
          <Show
            when={showThumbnail()}
            fallback={(
              <Icon name={fallbackIconName()} class={fallbackIconClass()} />
            )}
          >
            <img
              src={thumbnailSrc()}
              alt={`Thumbnail for ${titleLabel()}`}
              class="w-full h-full object-cover"
              onError={() => setFailedThumbnailSrc(thumbnailSrc())}
            />
          </Show>
        </div>

        <div class="min-w-0 flex-1 space-y-2">
          <div class="font-bold text-white text-base leading-snug break-words">{titleLabel()}</div>
          <div class="text-sm text-gray-300 truncate">{creatorLabel()}</div>
          <div class="flex flex-wrap gap-2 text-[10px] font-semibold uppercase tracking-wide">
            <span class="px-2 py-1 rounded-full bg-white/5 border border-white/10 text-gray-300">{typeLabel()}</span>
            <span class="px-2 py-1 rounded-full bg-white/5 border border-white/10 text-gray-400">{collectionLabel()}</span>
            <span class="px-2 py-1 rounded-full bg-white/5 border border-white/10 text-gray-400">{sourcePlaylistLabel()}</span>
          </div>
        </div>
      </div>

      <div class="w-full rounded-2xl border border-white/10 bg-[#06080f] overflow-hidden">
        <Show when={isVideo()}>
          <div class="aspect-video">
            <video
              controls
              autoplay
              class="w-full h-full object-contain bg-black"
              src={media()?.url}
              poster={thumbnailSrc() || undefined}
            />
          </div>
        </Show>
        <Show when={isAudio()}>
          <div class="p-3">
            <audio controls autoplay class="w-full" src={media()?.url} />
          </div>
        </Show>
        <Show when={!isVideo() && !isAudio()}>
          <div class="py-8 text-center">
            <Icon name="play-circle" class="w-10 h-10 text-blue-600/50 mx-auto mb-2" />
            <p class="text-xs text-gray-500">Unsupported media type</p>
          </div>
        </Show>
      </div>
    </div>
  );
}
