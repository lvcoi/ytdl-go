import { createSignal, Show, splitProps } from 'solid-js';
import Icon from './Icon';
import SongActions from './SongActions';

export default function Thumbnail(props) {
  const [local, rest] = splitProps(props, ['src', 'alt', 'class', 'size', 'item', 'viewMode']);

  const hasItem = () => !!local.item;
  const isList = () => local.viewMode === 'list';
  const [isHovered, setIsHovered] = createSignal(false);

  // Resolve src/alt from item if provided, otherwise use direct props
  const imgSrc = () => local.src || local.item?.thumbnail || local.item?.thumbnailUrl || '';
  const imgAlt = () => local.alt || local.item?.title || 'Video thumbnail';

  const sizeClasses = {
    sm: 'aspect-video w-24',
    md: 'aspect-video w-full',
    lg: 'aspect-video w-full max-w-2xl',
  }[local.size || 'md'];

  // Enhanced list-view layout when item + viewMode="list"
  if (hasItem() && isList()) {
    return (
      <div
        class={`thumbnail-container relative group overflow-hidden rounded-xl bg-bg-surface-soft flex items-center h-16 w-full ${local.class || ''}`}
        onPointerEnter={() => setIsHovered(true)}
        onPointerLeave={() => setIsHovered(false)}
        {...rest}
      >
        <img
          src={imgSrc() || '/placeholder.png'}
          alt={imgAlt()}
          class="w-24 h-full object-cover"
          loading="lazy"
        />
        <div class="flex-1 flex justify-between items-center px-4 h-full min-w-0">
          <div class="min-w-0">
            <div class="text-sm font-semibold truncate">{local.item.title}</div>
            <div class="text-xs opacity-80 truncate">{local.item.creator || local.item.channel || ''}</div>
          </div>
          <Show when={isHovered()}>
            <SongActions media={local.item} />
          </Show>
        </div>
      </div>
    );
  }

  // Enhanced gallery-view layout when item + viewMode="gallery"
  if (hasItem()) {
    return (
      <div
        class={`thumbnail-container relative group/thumb overflow-hidden rounded-xl bg-bg-surface-soft ${sizeClasses} ${local.class || ''}`}
        onPointerEnter={() => setIsHovered(true)}
        onPointerLeave={() => setIsHovered(false)}
        {...rest}
      >
        <Show
          when={imgSrc()}
          fallback={
            <div class="absolute inset-0 flex items-center justify-center text-gray-700">
              <Icon name="image" class="w-8 h-8" />
            </div>
          }
        >
          <img
            src={imgSrc()}
            alt={imgAlt()}
            class="w-full h-full object-cover transition-smooth group-hover/thumb:scale-105"
            loading="lazy"
          />
        </Show>
        <Show when={isHovered()}>
          <div class="absolute inset-0 bg-black/40 transition-opacity">
            <div class="absolute top-2 right-2 z-10">
              <SongActions media={local.item} />
            </div>
            <div class="absolute bottom-0 left-0 right-0 p-2 bg-gradient-to-t from-black to-transparent text-white">
              <div class="text-sm font-semibold truncate">{local.item.title}</div>
              <div class="text-xs opacity-80 truncate">{local.item.creator || local.item.channel || ''}</div>
            </div>
          </div>
        </Show>
      </div>
    );
  }

  // Default simple mode (backward compatible)
  return (
    <div
      class={`relative rounded-xl overflow-hidden bg-bg-surface-soft group/thumb ${sizeClasses} ${local.class || ''}`}
      {...rest}
    >
      <Show
        when={local.src}
        fallback={
          <div class="absolute inset-0 flex items-center justify-center text-gray-700">
            <Icon name="image" class="w-8 h-8" />
          </div>
        }
      >
        <img
          src={local.src}
          alt={local.alt || 'Video thumbnail'}
          class="w-full h-full object-cover transition-smooth group-hover/thumb:scale-105"
          loading="lazy"
        />
        <div class="absolute inset-0 bg-gradient-to-t from-black/60 via-transparent to-transparent opacity-0 group-hover/thumb:opacity-100 transition-smooth" />
      </Show>
    </div>
  );
}
