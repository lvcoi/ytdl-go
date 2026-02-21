const AUDIO_FILE_PATTERN = /\.(mp3|wav|ogg|flac|m4a|opus|aac|alac|wma|aiff|aif|ape|dsf|dff|mka)$/i;
const VIDEO_FILE_PATTERN = /\.(mp4|webm|mov|avi|mkv|m4v|wmv|mpeg|mpg|ts|ogv|3gp|f4v|flv|vob|m2ts)$/i;


export const isAudioFile = (filename) => AUDIO_FILE_PATTERN.test(String(filename || ''));
export const isVideoFile = (filename) => VIDEO_FILE_PATTERN.test(String(filename || ''));

export const normalizeMediaType = (value) => {
  const lowered = String(value || '').trim().toLowerCase();
  if (lowered === 'audio') {
    return 'audio';
  }
  if (lowered === 'video') {
    return 'video';
  }
  return 'unknown';
};

export const detectMediaType = (item) => {
  const filename = String(item?.filename || '').trim();
  if (filename !== '') {
    if (isAudioFile(filename)) {
      return 'audio';
    }
    if (isVideoFile(filename)) {
      return 'video';
    }
  }
  return normalizeMediaType(item?.type);
};
