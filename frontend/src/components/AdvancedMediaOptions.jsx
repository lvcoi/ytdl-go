import { For } from 'solid-js';
import { useAppStore } from '../store/appStore';
import Icon from './Icon';

const AUDIO_FORMATS = [
  { value: '', label: 'Best (source)' },
  { value: 'mp3', label: 'MP3' },
  { value: 'opus', label: 'Opus' },
  { value: 'm4a', label: 'M4A / AAC' },
  { value: 'flac', label: 'FLAC' },
  { value: 'wav', label: 'WAV' },
];

function Toggle({ label, description, value, onChange }) {
  return (
    <div class="flex items-center justify-between p-3 bg-white/5 rounded-xl border border-white/5 hover:border-white/10 transition-colors">
      <div class="flex flex-col pr-3">
        <span class="text-sm font-bold text-gray-200">{label}</span>
        {description && <span class="text-[11px] text-gray-500 mt-0.5">{description}</span>}
      </div>
      <label class="relative inline-flex items-center cursor-pointer shrink-0">
        <input type="checkbox" class="sr-only peer" checked={value} onChange={(e) => onChange(e.target.checked)} />
        <div class="w-11 h-6 bg-white/10 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-accent-primary"></div>
      </label>
    </div>
  );
}

function TextField({ label, placeholder, value, onInput, hint }) {
  return (
    <div class="space-y-2">
      <label class="text-xs font-bold text-gray-500 uppercase tracking-wider">{label}</label>
      <input
        type="text"
        value={value}
        placeholder={placeholder}
        onInput={(e) => onInput(e.target.value)}
        class="w-full bg-white/5 border border-white/10 rounded-xl px-4 py-2 text-sm font-mono text-gray-300 placeholder-gray-600 focus:outline-none focus:border-blue-500/50"
      />
      {hint && <p class="text-[10px] text-gray-500">{hint}</p>}
    </div>
  );
}

export default function AdvancedMediaOptions() {
  const { state, setState } = useAppStore();
  const set = (key, value) => setState('settings', key, value);

  return (
    <div class="space-y-6 pt-4 border-t border-white/5 animate-in slide-in-from-top-2">
      <div class="flex items-center gap-2">
        <Icon name="sliders-horizontal" class="w-4 h-4 text-blue-400" />
        <h3 class="text-xs font-black uppercase tracking-widest text-gray-400">Media Options</h3>
      </div>

      {/* Audio conversion */}
      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div class="space-y-2">
          <label class="text-xs font-bold text-gray-500 uppercase tracking-wider">Audio Format (Audio Only)</label>
          <select
            value={state.settings.audioFormat}
            onChange={(e) => set('audioFormat', e.target.value)}
            class="w-full bg-white/5 border border-white/10 rounded-xl px-4 py-2 text-sm text-gray-300 focus:outline-none focus:border-blue-500/50 appearance-none"
          >
            <For each={AUDIO_FORMATS}>
              {(format) => <option value={format.value} class="bg-[#0f172a]">{format.label}</option>}
            </For>
          </select>
        </div>
        <TextField
          label="Audio Quality"
          placeholder="160k or 0-9 (VBR)"
          value={state.settings.audioQuality}
          onInput={(v) => set('audioQuality', v)}
          hint="Bitrate (e.g. 160k) or VBR scale 0-9. Empty = format default."
        />
      </div>

      {/* Thumbnails */}
      <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
        <Toggle
          label="Save Thumbnail"
          description="Write the thumbnail image next to the media file."
          value={state.settings.writeThumbnail}
          onChange={(v) => set('writeThumbnail', v)}
        />
        <Toggle
          label="Embed Thumbnail"
          description="Attach cover art into mp3 / m4a / mp4 / mkv."
          value={state.settings.embedThumbnail}
          onChange={(v) => set('embedThumbnail', v)}
        />
      </div>

      {/* Subtitles */}
      <div class="space-y-3">
        <div class="grid grid-cols-1 md:grid-cols-3 gap-3">
          <Toggle
            label="Subtitles"
            description="Write .vtt subtitle files."
            value={state.settings.writeSubs}
            onChange={(v) => set('writeSubs', v)}
          />
          <Toggle
            label="Auto-Generated"
            description="Include auto-generated captions."
            value={state.settings.writeAutoSubs}
            onChange={(v) => set('writeAutoSubs', v)}
          />
          <Toggle
            label="Embed Subtitles"
            description="Embed the best track into the container."
            value={state.settings.embedSubs}
            onChange={(v) => set('embedSubs', v)}
          />
        </div>
        <TextField
          label="Subtitle Languages"
          placeholder="en,de,ja (or all)"
          value={state.settings.subLangs}
          onInput={(v) => set('subLangs', v)}
        />
      </div>

      {/* SponsorBlock */}
      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <TextField
          label="SponsorBlock: Remove"
          placeholder="sponsor,selfpromo,interaction"
          value={state.settings.sponsorblockRemove}
          onInput={(v) => set('sponsorblockRemove', v)}
          hint="Cut these segments out of the file."
        />
        <TextField
          label="SponsorBlock: Mark as Chapters"
          placeholder="sponsor,selfpromo"
          value={state.settings.sponsorblockMark}
          onInput={(v) => set('sponsorblockMark', v)}
          hint="Add chapter markers for these segments."
        />
      </div>

      {/* Playlist selection & filters */}
      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <TextField
          label="Playlist Items"
          placeholder="1:5,8,10:12"
          value={state.settings.playlistItems}
          onInput={(v) => set('playlistItems', v)}
          hint="1-based entries; ~ negates a range."
        />
        <TextField
          label="Match Filter"
          placeholder="duration<600&view_count>1000"
          value={state.settings.matchFilter}
          onInput={(v) => set('matchFilter', v)}
          hint="Only download entries matching all conditions."
        />
      </div>

      {/* Archive & live */}
      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <TextField
          label="Download Archive"
          placeholder="archive.txt"
          value={state.settings.downloadArchive}
          onInput={(v) => set('downloadArchive', v)}
          hint="Skip videos already listed in this file (server-side path)."
        />
        <div class="space-y-3">
          <Toggle
            label="Break on Existing"
            description="Stop a playlist at the first archived entry."
            value={state.settings.breakOnExisting}
            onChange={(v) => set('breakOnExisting', v)}
          />
          <Toggle
            label="Record Live Stream"
            description="Record ongoing live streams via HLS."
            value={state.settings.live}
            onChange={(v) => set('live', v)}
          />
        </div>
      </div>
    </div>
  );
}
