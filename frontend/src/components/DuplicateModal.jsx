import { For, Show } from 'solid-js';
import Icon from './Icon';

const actions = [
  { value: 'overwrite', label: 'Overwrite', hint: 'Replace this file', key: 'o' },
  { value: 'overwrite_all', label: 'Overwrite All', hint: 'Replace all duplicates in this job', key: 'O' },
  { value: 'skip', label: 'Skip', hint: 'Skip this file', key: 's' },
  { value: 'skip_all', label: 'Skip All', hint: 'Skip all duplicates in this job', key: 'S' },
  { value: 'rename', label: 'Rename', hint: 'Save as next available filename', key: 'r' },
  { value: 'rename_all', label: 'Rename All', hint: 'Rename all duplicates in this job', key: 'R' },
  { value: 'cancel', label: 'Cancel', hint: 'Abort this item', key: 'q' },
];

export default function DuplicateModal(props) {
  return (
    <div class="fixed inset-0 z-50 bg-[#05070a]/80 backdrop-blur-sm flex items-center justify-center p-4">
      <div class="w-full max-w-2xl rounded-3xl border border-white/10 bg-[#0a0c14] shadow-2xl p-8 space-y-6 animate-in fade-in zoom-in-95 duration-300">
        <div class="flex items-center gap-3">
          <div class="p-2 rounded-xl bg-amber-500/10 text-amber-400">
            <Icon name="alert-circle" class="w-5 h-5" />
          </div>
          <div>
            <div class="font-bold text-white text-xl">Duplicate File Detected</div>
            <div class="text-xs text-gray-500">Choose how to handle this file.</div>
          </div>
        </div>

        <div class="p-4 rounded-2xl border border-white/10 bg-[#05070a]">
          <div class="text-xs text-gray-500 mb-1">Filename</div>
          <div class="text-sm text-white font-mono break-all">{props.prompt?.filename || 'Unknown file'}</div>
          <Show when={props.prompt?.path}>
            <div class="text-[11px] text-gray-600 mt-2 break-all">{props.prompt.path}</div>
          </Show>
        </div>

        <Show when={props.error}>
          <div class="p-3 rounded-xl bg-red-500/10 border border-red-500/20 text-red-400 text-xs">{props.error}</div>
        </Show>

        <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
          <For each={actions}>
            {(action) => (
              <button
                onClick={() => props.onSelect?.(action.value)}
                class="text-left p-4 rounded-2xl border border-white/10 bg-white/5 hover:border-blue-500/40 hover:bg-blue-500/10 transition-all"
              >
                <div class="flex items-center justify-between">
                  <span class="font-bold text-white text-sm">{action.label}</span>
                  <span class="text-[10px] px-2 py-0.5 rounded-full bg-white/10 text-gray-400 font-mono">{action.key}</span>
                </div>
                <div class="text-xs text-gray-500 mt-1">{action.hint}</div>
              </button>
            )}
          </For>
        </div>
      </div>
    </div>
  );
}
