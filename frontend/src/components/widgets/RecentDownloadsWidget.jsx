import { createMemo, For, Show } from 'solid-js';
import { useAppStore } from '../../store/appStore';

export default function RecentDownloadsWidget({ rowSpan, colSpan }) {
  const { state } = useAppStore();

  const downloads = createMemo(() => {
    const all = state.library.downloads || [];
    return all.slice(0, 10);
  });

  const itemsToShow = createMemo(() => {
    if (rowSpan >= 4) return 3;
    if (rowSpan >= 3) return 2;
    if (rowSpan >= 2) return 1;
    return 0;
  });

  const showThumbnail = createMemo(() => rowSpan >= 4);
  const showCreator = createMemo(() => rowSpan >= 3);

  return (
    <div class="flex flex-col gap-2 h-full overflow-hidden">
      <Show when={rowSpan <= 1}>
        <div class="flex items-center justify-between">
          <span class="text-2xl font-bold text-white">{downloads().length}</span>
          <span class="text-xs text-gray-500">recent files</span>
        </div>
      </Show>
      <Show when={rowSpan > 1}>
        <For each={downloads().slice(0, itemsToShow())}>
          {(item) => (
            <div class="flex items-center gap-3 p-2 bg-white/5 rounded-xl overflow-hidden">
              <Show when={showThumbnail()}>
                <div class="w-12 h-9 bg-white/10 rounded-lg flex-shrink-0 overflow-hidden">
                  <Show when={item.thumbnail}>
                    <img src={item.thumbnail} alt="" class="w-full h-full object-cover" />
                  </Show>
                </div>
              </Show>
              <div class="flex-1 min-w-0">
                <div class="text-sm text-white truncate font-medium">{item.title || item.filename}</div>
                <Show when={showCreator()}>
                  <div class="text-xs text-gray-500 truncate">{item.creator || item.channel || ''}</div>
                </Show>
              </div>
            </div>
          )}
        </For>
        <Show when={downloads().length === 0}>
          <div class="flex-1 flex items-center justify-center text-gray-600 text-sm">No downloads yet</div>
        </Show>
      </Show>
    </div>
  );
}
