import Hls, { Events, type ErrorData } from 'hls.js';
import type {
  CreatePlaybackMedia,
  PlaybackMedia,
  PlaybackMediaCallbacks,
} from '../ports/PlaybackMedia';

const hlsMimeType = 'application/vnd.apple.mpegurl';

function levelLabel(height: number, name: string | undefined, index: number): string {
  if (height > 0) return `${height}p`;
  if (name !== undefined && name.trim() !== '') return name;
  return `Level ${index + 1}`;
}

function createHlsMedia(video: HTMLVideoElement, callbacks: PlaybackMediaCallbacks): PlaybackMedia {
  const hls = new Hls({ enableWorker: true });
  hls.attachMedia(video);
  hls.on(Events.MANIFEST_PARSED, () => {
    callbacks.onLevels(
      hls.levels.map((level, index) => ({
        id: index,
        label: levelLabel(level.height, level.name, index),
      })),
    );
    callbacks.onReady();
  });
  hls.on(Events.ERROR, (_event, data: ErrorData) => {
    const status = data.response?.code;
    const url = data.url ?? data.frag?.url;
    if (status !== undefined) callbacks.onError({ status, retryable: status === 503, url });
    else if (data.fatal) callbacks.onError({ status: undefined, retryable: false, url });
  });
  return {
    kind: 'hls',
    load(source) {
      hls.loadSource(source);
    },
    stopLoad() {
      hls.stopLoad();
    },
    selectLevel(level) {
      // `nextLevel`, not `currentLevel`: currentLevel flushes the buffer and interrupts playback
      // to re-buffer at the new rendition (a visible stall and audio gap). nextLevel keeps the
      // already-buffered fragment playing and swaps at the next fragment boundary, once the new
      // level's data is loaded ahead - a seamless switch for both video and audio, at the cost of
      // the change taking effect a fragment later rather than instantly.
      hls.nextLevel = level === 'auto' ? -1 : level;
    },
    destroy() {
      hls.destroy();
    },
  };
}

function createNativeMedia(
  video: HTMLVideoElement,
  callbacks: PlaybackMediaCallbacks,
): PlaybackMedia {
  const ready = () => callbacks.onReady();
  // The native branch never learns which request failed - the element reports only that it did.
  const failed = () => callbacks.onError({ status: undefined, retryable: true, url: undefined });
  video.addEventListener('loadedmetadata', ready);
  video.addEventListener('error', failed);
  return {
    kind: 'native',
    load(source) {
      video.src = source;
      video.load();
    },
    stopLoad() {
      video.pause();
      video.removeAttribute('src');
      video.load();
    },
    selectLevel() {},
    destroy() {
      video.removeEventListener('loadedmetadata', ready);
      video.removeEventListener('error', failed);
      video.removeAttribute('src');
    },
  };
}

export const createPlaybackMedia: CreatePlaybackMedia = (video, callbacks) => {
  // This order is load-bearing: Chromium reports native HLS as "maybe" even though hls.js is
  // the branch that provides working playback, quality selection and HTTP error status there.
  if (Hls.isSupported()) return createHlsMedia(video, callbacks);
  if (video.canPlayType(hlsMimeType) !== '') return createNativeMedia(video, callbacks);
  return createNativeMedia(video, callbacks);
};
