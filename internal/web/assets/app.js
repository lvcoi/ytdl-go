const form = document.getElementById("download-form");
const responseEl = document.getElementById("response");

const field = (id) => document.getElementById(id);

const parseMeta = () => {
  const raw = field("meta").value.trim();
  if (!raw) return {};
  return raw.split("\n").reduce((acc, line) => {
    const [key, ...rest] = line.split("=");
    const trimmedKey = key ? key.trim() : "";
    if (!trimmedKey || rest.length === 0) return acc;
    acc[trimmedKey.toLowerCase()] = rest.join("=").trim();
    return acc;
  }, {});
};

const parseNumber = (id) => {
  const value = field(id).value.trim();
  if (!value) return 0;
  const num = Number(value);
  if (Number.isNaN(num)) return 0;
  return num;
};

const payload = () => ({
  urls: field("urls").value
    .split("\n")
    .map((line) => line.trim())
    .filter(Boolean),
  options: {
    output: field("output").value.trim(),
    audio: field("audio").checked,
    info: field("info").checked,
    "list-formats": field("list-formats").checked,
    quality: field("quality").value.trim(),
    format: field("format").value.trim(),
    itag: parseNumber("itag"),
    meta: parseMeta(),
    "progress-layout": field("progress-layout").value.trim(),
    "segment-concurrency": parseNumber("segment-concurrency"),
    "playlist-concurrency": parseNumber("playlist-concurrency"),
    jobs: parseNumber("jobs") || 1,
    json: field("json").checked,
    timeout: parseNumber("timeout"),
    quiet: field("quiet").checked,
    "log-level": field("log-level").value,
  },
});

const setResponse = (data) => {
  responseEl.textContent = JSON.stringify(data, null, 2);
};

form.addEventListener("submit", async (event) => {
  event.preventDefault();
  const body = payload();
  if (!body.urls.length) {
    setResponse({ error: "Please enter at least one URL." });
    return;
  }

  responseEl.textContent = "Starting download...";

  try {
    const res = await fetch("/api/download", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
    });
    if (!res.ok) {
      let errorDetail;
      try {
        const errData = await res.json();
        errorDetail = errData && (errData.error || errData.message) ? (errData.error || errData.message) : JSON.stringify(errData);
      } catch (_) {
        errorDetail = await res.text();
      }
      setResponse({
        error: `Request failed with status ${res.status}${errorDetail ? `: ${errorDetail}` : ""}`,
      });
      return;
    }
    const data = await res.json();
    setResponse(data);
  } catch (error) {
    setResponse({ error: error.message });
  }
});
