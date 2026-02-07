import { Show, createEffect, createMemo, createSignal, onCleanup } from 'solid-js';
import Icon from './Icon';
import { detectMediaType } from '../utils/mediaType';

const PLAYER_EDGE_GAP = 12;

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

const clamp = (value, min, max) => Math.min(Math.max(value, min), max);

export default function Player(props) {
  let panelRef;
  let dragSession = null;
  const media = createMemo(() => (typeof props.media === 'function' ? props.media() : props.media));
  const metadata = createMemo(() => (media()?.metadata && typeof media().metadata === 'object' ? media().metadata : {}));
  const isMinimized = createMemo(() => {
    const value = typeof props.minimized === 'function' ? props.minimized() : props.minimized;
    return Boolean(value);
  });
  const queueCount = createMemo(() => {
    const raw = typeof props.queueCount === 'function' ? props.queueCount() : props.queueCount;
    const parsed = Number(raw);
    if (!Number.isFinite(parsed)) {
      return 0;
    }
    return Math.max(0, Math.trunc(parsed));
  });
  const canGoNext = createMemo(() => {
    const raw = typeof props.canGoNext === 'function' ? props.canGoNext() : props.canGoNext;
    return Boolean(raw);
  });
  const mediaType = createMemo(() => {
    return detectMediaType(media());
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
  const [floatingPosition, setFloatingPosition] = createSignal(null);
  const [isDragging, setIsDragging] = createSignal(false);
  const [mediaElement, setMediaElement] = createSignal(null);
  const [isPlaying, setIsPlaying] = createSignal(false);
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
  const floatingStyle = createMemo(() => {
    const current = floatingPosition();
    if (!current) {
      return undefined;
    }
    return {
      left: `${current.left}px`,
      top: `${current.top}px`,
      right: 'auto',
      bottom: 'auto',
    };
  });

  const syncPlaybackState = () => {
    const element = mediaElement();
    setIsPlaying(Boolean(element && !element.paused && !element.ended));
  };

  const clampPosition = (left, top, width, height) => {
    if (typeof window === 'undefined') {
      return { left, top };
    }
    const maxLeft = Math.max(PLAYER_EDGE_GAP, window.innerWidth - width - PLAYER_EDGE_GAP);
    const maxTop = Math.max(PLAYER_EDGE_GAP, window.innerHeight - height - PLAYER_EDGE_GAP);
    return {
      left: clamp(left, PLAYER_EDGE_GAP, maxLeft),
      top: clamp(top, PLAYER_EDGE_GAP, maxTop),
    };
  };

  const stopDragging = () => {
    dragSession = null;
    setIsDragging(false);
  };

  const handleDragPointerMove = (event) => {
    if (!dragSession || event.pointerId !== dragSession.pointerId || isMinimized()) {
      return;
    }
    const nextLeft = event.clientX - dragSession.offsetX;
    const nextTop = event.clientY - dragSession.offsetY;
    setFloatingPosition(clampPosition(nextLeft, nextTop, dragSession.width, dragSession.height));
  };

  const endDrag = (event) => {
    if (!dragSession || event.pointerId !== dragSession.pointerId) {
      return;
    }
    if (event.currentTarget?.hasPointerCapture?.(event.pointerId)) {
      event.currentTarget.releasePointerCapture(event.pointerId);
    }
    stopDragging();
  };

  const beginDrag = (event) => {
    if (isMinimized() || event.button !== 0) {
      return;
    }
    if (event.target instanceof Element && event.target.closest('button')) {
      return;
    }
    const panel = panelRef;
    if (!panel) {
      return;
    }
    const rect = panel.getBoundingClientRect();
    const anchored = clampPosition(rect.left, rect.top, rect.width, rect.height);
    dragSession = {
      pointerId: event.pointerId,
      offsetX: event.clientX - rect.left,
      offsetY: event.clientY - rect.top,
      width: rect.width,
      height: rect.height,
    };
    setFloatingPosition(anchored);
    setIsDragging(true);
    event.currentTarget.setPointerCapture?.(event.pointerId);
    event.preventDefault();
  };

  const handleTogglePlayback = async () => {
    const element = mediaElement();
    if (!element) {
      return;
    }
    if (element.paused || element.ended) {
      try {
        await element.play();
      } catch (error) {
        return;
      }
    } else {
      element.pause();
    }
    syncPlaybackState();
  };

  const handleResize = () => {
    const current = floatingPosition();
    if (!panelRef || !current) {
      return;
    }
    const rect = panelRef.getBoundingClientRect();
    setFloatingPosition(clampPosition(current.left, current.top, rect.width, rect.height));
  };
  const bindMediaElement = (element) => {
    setMediaElement(element ?? null);
    if (!element) {
      setIsPlaying(false);
    }
  };

  createEffect(() => {
    thumbnailSrc();
    setFailedThumbnailSrc('');
  });

  createEffect(() => {
    media()?.url;
    setIsPlaying(false);
  });

  createEffect(() => {
    if (!isVideo() && !isAudio()) {
      setMediaElement(null);
      setIsPlaying(false);
    }
  });

  createEffect(() => {
    if (isMinimized()) {
      stopDragging();
    }
  });

  onCleanup(() => {
    const pointerId = dragSession?.pointerId;
    if (panelRef && typeof pointerId === 'number' && panelRef.hasPointerCapture?.(pointerId)) {
      panelRef.releasePointerCapture(pointerId);
    }
    stopDragging();
    if (typeof window !== 'undefined') {
      window.removeEventListener('resize', handleResize);
    }
  });

  if (typeof window !== 'undefined') {
    window.addEventListener('resize', handleResize);
  }

  return (
    <>
      <div
        ref={panelRef}
        class={`fixed bottom-6 right-6 w-[32rem] max-w-[calc(100vw-1.5rem)] glass rounded-[2rem] border border-white/10 shadow-2xl p-6 space-y-5 z-50 ${isMinimized() ? 'hidden' : 'animate-in slide-in-from-bottom-10 duration-500'}`}
        style={floatingStyle()}
      >
        <div
          class={`flex items-center justify-between ${isDragging() ? 'cursor-grabbing' : 'cursor-move'}`}
          onPointerDown={beginDrag}
          onPointerMove={handleDragPointerMove}
          onPointerUp={endDrag}
          onPointerCancel={endDrag}
        >
          <span class="text-[10px] font-black text-blue-500 uppercase tracking-widest">Now Playing</span>
          <div class="flex items-center gap-1">
            <button
              onClick={() => {
                if (typeof props.onMinimize === 'function') {
                  props.onMinimize();
                }
              }}
              class="p-1 rounded-lg text-gray-500 hover:text-gray-300 hover:bg-white/5 transition-colors"
              title="Minimize player"
            >
              <Icon name="chevron-down" class="w-4 h-4" />
            </button>
            <button
              onClick={() => {
                if (typeof props.onClose === 'function') {
                  props.onClose();
                }
              }}
              class="p-1 rounded-lg text-gray-500 hover:text-red-300 hover:bg-red-500/10 transition-colors"
              title="Close player"
            >
              <Icon name="x" class="w-4 h-4" />
            </button>
          </div>
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
          <Show when={isVideo() ? media() : null} keyed>
            {(mediaValue) => (
              <div class="aspect-video">
                <video
                  ref={bindMediaElement}
                  controls
                  autoplay
                  class="w-full h-full object-contain bg-black"
                  src={mediaValue?.url}
                  poster={thumbnailSrc() || undefined}
                  onPlay={syncPlaybackState}
                  onPause={syncPlaybackState}
                  onLoadedMetadata={syncPlaybackState}
                  onEnded={syncPlaybackState}
                />
              </div>
            )}
          </Show>
          <Show when={isAudio() ? media() : null} keyed>
            {(mediaValue) => (
              <div class="p-3">
                <audio
                  ref={bindMediaElement}
                  controls
                  autoplay
                  class="w-full"
                  src={mediaValue?.url}
                  onPlay={syncPlaybackState}
                  onPause={syncPlaybackState}
                  onLoadedMetadata={syncPlaybackState}
                  onEnded={syncPlaybackState}
                />
              </div>
            )}
          </Show>
          <Show when={!isVideo() && !isAudio()}>
            <div class="py-8 text-center">
              <Icon name="play-circle" class="w-10 h-10 text-blue-600/50 mx-auto mb-2" />
              <p class="text-xs text-gray-500">Unsupported media type</p>
            </div>
          </Show>
        </div>
      </div>

      <div class={`fixed bottom-4 right-4 w-[34rem] max-w-[calc(100vw-1.5rem)] glass rounded-2xl border border-white/10 shadow-2xl p-3 z-50 ${isMinimized() ? 'animate-in slide-in-from-bottom-10 duration-300' : 'hidden'}`}>
        <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
          <div class="flex items-center gap-3 min-w-0">
            <div class="w-12 h-12 shrink-0 rounded-xl border border-white/10 bg-[#0b0f1a] overflow-hidden flex items-center justify-center">
              <Show
                when={showThumbnail()}
                fallback={(
                  <Icon name={fallbackIconName()} class="w-5 h-5 text-gray-400/70" />
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
            <div class="min-w-0">
              <div class="text-xs font-bold text-white truncate">{titleLabel()}</div>
              <div class="text-[11px] text-gray-400 truncate">{creatorLabel()}</div>
            </div>
          </div>

          <div class="flex items-center justify-end gap-2">
            <button
              onClick={handleTogglePlayback}
              disabled={!mediaElement()}
              class={`px-3 py-1.5 rounded-lg text-xs font-semibold border transition-all ${mediaElement() ? 'border-white/10 bg-white/5 text-gray-200 hover:text-white hover:border-white/20' : 'border-white/5 bg-white/5 text-gray-500 cursor-not-allowed'}`}
              title="Toggle playback"
            >
              <span class="inline-flex items-center gap-1.5">
                <Icon name={isPlaying() ? 'pause' : 'play'} class="w-3.5 h-3.5" />
                {isPlaying() ? 'Pause' : 'Play'}
              </span>
            </button>
            <button
              onClick={() => {
                if (typeof props.onNext === 'function') {
                  props.onNext();
                }
              }}
              disabled={!canGoNext()}
              class={`px-3 py-1.5 rounded-lg text-xs font-semibold border transition-all ${canGoNext() ? 'border-white/10 bg-white/5 text-gray-200 hover:text-white hover:border-white/20' : 'border-white/5 bg-white/5 text-gray-500 cursor-not-allowed'}`}
              title="Next in queue"
            >
              <span class="inline-flex items-center gap-1.5">
                <Icon name="skip-forward" class="w-3.5 h-3.5" />
                Next
              </span>
            </button>
            <button
              onClick={() => {
                if (typeof props.onQueue === 'function') {
                  props.onQueue();
                }
              }}
              class="px-3 py-1.5 rounded-lg text-xs font-semibold border border-white/10 bg-white/5 text-gray-200 hover:text-white hover:border-white/20 transition-all"
              title="Open queue"
            >
              <span class="inline-flex items-center gap-1.5">
                <Icon name="layers" class="w-3.5 h-3.5" />
                Queue{queueCount() > 0 ? ` (${queueCount()})` : ''}
              </span>
            </button>
            <button
              onClick={() => {
                if (typeof props.onRestore === 'function') {
                  props.onRestore();
                }
              }}
              class="p-1.5 rounded-lg text-gray-400 hover:text-gray-200 hover:bg-white/5 transition-colors"
              title="Restore player"
            >
              <Icon name="chevron-down" class="w-4 h-4 rotate-180" />
            </button>
            <button
              onClick={() => {
                if (typeof props.onClose === 'function') {
                  props.onClose();
                }
              }}
              class="p-1.5 rounded-lg text-gray-400 hover:text-red-300 hover:bg-red-500/10 transition-colors"
              title="Close player"
            >
              <Icon name="x" class="w-4 h-4" />
            </button>
          </div>
        </div>
      </div>
    </>
  );
}
