import { createSignal } from 'solid-js';

export default function QuickDownloadWidget({ rowSpan, colSpan }) {
  const [url, setUrl] = createSignal('');
  const [format, setFormat] = createSignal('best');

  const handleDownload = () => {
    if (!url().trim()) return;
    window.dispatchEvent(new CustomEvent('ytdl-quick-download', { detail: { url: url(), format: format() } }));
    setUrl('');
  };

  return (
    <div class="flex flex-col gap-3 h-full">
      <input
        type="text"
        value={url()}
        onInput={(e) => setUrl(e.target.value)}
        onKeyDown={(e) => e.key === 'Enter' && handleDownload()}
        placeholder="Paste URL here..."
        class="w-full bg-white/5 border border-white/10 rounded-xl px-3 py-2 text-sm text-white placeholder-gray-500 outline-none focus:border-blue-500/50 focus:bg-white/8 transition-all"
      />
      <div class="flex gap-2">
        <select
          value={format()}
          onChange={(e) => setFormat(e.target.value)}
          class="flex-1 bg-white/5 border border-white/10 rounded-xl px-3 py-2 text-sm text-gray-300 outline-none focus:border-blue-500/50 transition-all"
        >
          <option value="best">Best Quality</option>
          <option value="1080p">1080p</option>
          <option value="720p">720p</option>
          <option value="480p">480p</option>
          <option value="audio">Audio Only</option>
        </select>
        <button
          onClick={handleDownload}
          class="px-4 py-2 bg-blue-600 hover:bg-blue-500 text-white text-sm font-semibold rounded-xl transition-all"
        >
          Download
        </button>
      </div>
    </div>
  );
}
