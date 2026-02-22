import { createMemo, Show } from 'solid-js';
import { useAppStore } from '../../store/appStore';

export default function SystemStatsWidget({ rowSpan, colSpan }) {
  const { state } = useAppStore();

  const isDownloading = createMemo(() => state.download.isDownloading);
  const activeCount = createMemo(() => isDownloading() ? 1 : 0);
  const completedCount = createMemo(() => (state.library.downloads || []).length);
  const jobStatus = createMemo(() => state.download.jobStatus);

  return (
    <div class="flex flex-col gap-3 h-full">
      <Show when={rowSpan >= 2}>
        <div class="grid grid-cols-2 gap-2">
          <div class="bg-white/5 rounded-xl p-3">
            <div class="text-xs text-gray-500 mb-1">Active</div>
            <div class="text-2xl font-bold text-blue-400">{activeCount()}</div>
          </div>
          <div class="bg-white/5 rounded-xl p-3">
            <div class="text-xs text-gray-500 mb-1">Completed</div>
            <div class="text-2xl font-bold text-green-400">{completedCount()}</div>
          </div>
        </div>
        <Show when={rowSpan >= 3 && jobStatus()}>
          <div class="bg-white/5 rounded-xl p-3">
            <div class="text-xs text-gray-500 mb-1">Last Job</div>
            <div class="text-sm text-white truncate">{jobStatus()?.jobId || 'None'}</div>
            <div class="text-xs text-gray-400">{jobStatus()?.status || ''}</div>
          </div>
        </Show>
      </Show>
      <Show when={rowSpan < 2}>
        <div class="flex items-center justify-between">
          <span class="text-sm text-gray-400">Active: <span class="text-blue-400 font-bold">{activeCount()}</span></span>
          <span class="text-sm text-gray-400">Done: <span class="text-green-400 font-bold">{completedCount()}</span></span>
        </div>
      </Show>
    </div>
  );
}
