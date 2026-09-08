import Hls, { Events, type ErrorData } from 'hls.js';
import type {
  CreatePlaybackMedia,
  PlaybackMedia,
  PlaybackMediaCallbacks,
} from '../../ports/PlaybackMedia';

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
    if (status !== undefined) callbacks.onError({ status, retryable: status === 503 });
    else if (data.fatal) callbacks.onError({ status: undefined, retryable: false });
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
      hls.currentLevel = level === 'auto' ? -1 : level;
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
  const failed = () => callbacks.onError({ status: undefined, retryable: true });
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
  if (Hls.isSupported()) return createHlsMedia(video, callbacks);
  if (video.canPlayType(hlsMimeType) !== '') return createNativeMedia(video, callbacks);
  return createNativeMedia(video, callbacks);
};
