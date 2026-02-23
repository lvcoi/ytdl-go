import { createMemo, Show } from 'solid-js';
import { useAppStore } from '../../store/appStore';

export default function StorageWidget({ rowSpan, colSpan }) {
  const { state } = useAppStore();

  const downloads = createMemo(() => state.library.downloads || []);
  const videoCount = createMemo(() => downloads().filter(d => d.type !== 'audio').length);
  const audioCount = createMemo(() => downloads().filter(d => d.type === 'audio').length);
  const totalCount = createMemo(() => downloads().length);

  return (
    <div class="flex flex-col gap-3 h-full">
      <Show when={rowSpan >= 2}>
        <div class="grid grid-cols-2 gap-2">
          <div class="bg-white/5 rounded-xl p-3">
            <div class="text-xs text-gray-500 mb-1">Videos</div>
            <div class="text-2xl font-bold text-purple-400">{videoCount()}</div>
          </div>
          <div class="bg-white/5 rounded-xl p-3">
            <div class="text-xs text-gray-500 mb-1">Audio</div>
            <div class="text-2xl font-bold text-pink-400">{audioCount()}</div>
          </div>
        </div>
        <Show when={rowSpan >= 3}>
          <div class="bg-white/5 rounded-xl p-3">
            <div class="text-xs text-gray-500 mb-1">Total Media</div>
            <div class="text-2xl font-bold text-white">{totalCount()}</div>
          </div>
        </Show>
      </Show>
      <Show when={rowSpan < 2}>
        <div class="flex items-center justify-between">
          <span class="text-sm text-gray-400">Videos: <span class="text-purple-400 font-bold">{videoCount()}</span></span>
          <span class="text-sm text-gray-400">Audio: <span class="text-pink-400 font-bold">{audioCount()}</span></span>
        </div>
      </Show>
    </div>
  );
}
