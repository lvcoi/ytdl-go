import { Show, splitProps } from 'solid-js';
import Icon from './Icon';

export default function Thumbnail(props) {
  const [local, rest] = splitProps(props, ['src', 'alt', 'class', 'size']);
  
  const sizeClasses = {
    sm: 'aspect-video w-24',
    md: 'aspect-video w-full',
    lg: 'aspect-video w-full max-w-2xl',
  }[local.size || 'md'];

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
