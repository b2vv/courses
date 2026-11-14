import Hls, { Events, type ErrorData } from 'hls.js';
import type {
  CreatePlaybackMedia,
  PlaybackAudioTrack,
  PlaybackMedia,
  PlaybackMediaCallbacks,
  PlaybackMediaOptions,
} from '../ports/PlaybackMedia';

const hlsMimeType = 'application/vnd.apple.mpegurl';

function levelLabel(height: number, name: string | undefined, index: number): string {
  if (height > 0) return `${height}p`;
  if (name !== undefined && name.trim() !== '') return name;
  return `Level ${index + 1}`;
}

type ManifestAudioTrack = {
  readonly lang?: string | undefined;
  readonly name: string;
  readonly default: boolean;
};

/**
 * The index of the track whose LANGUAGE is exactly `language`, or -1. Deliberately not hls.js
 * matching (`setAudioOption`/`audioPreference`): its `languagesMatch` accepts a prefix, so `und`
 * matches `und-x-original` and `en` matches `en-us`, and both pairs can coexist on one asset
 * (migration 00028). Case is kept: the backend accepts `en-US` as given.
 */
export function exactAudioTrackIndex(
  tracks: readonly ManifestAudioTrack[],
  language: string | null,
): number {
  if (language === null) return -1;
  return tracks.findIndex((track) => track.lang === language);
}

function audioTracksOf(tracks: readonly ManifestAudioTrack[]): PlaybackAudioTrack[] {
  return tracks.flatMap((track) =>
    track.lang === undefined || track.lang === ''
      ? []
      : [{ language: track.lang, label: track.name, isDefault: track.default }],
  );
}

function createHlsMedia(
  video: HTMLVideoElement,
  callbacks: PlaybackMediaCallbacks,
  options: PlaybackMediaOptions | undefined,
): PlaybackMedia {
  const hls = new Hls({ enableWorker: true });
  // The language the viewer asked for; kept across manifests so a retry or a new session URL
  // (both `loadSource` on this instance) starts on it again.
  let wanted = options?.audioLanguage ?? null;
  hls.on(Events.AUDIO_TRACKS_UPDATED, (_event, data) => {
    callbacks.onAudioTracks(audioTracksOf(data.audioTracks));
    // The controller triggers this event and only then selects its default, and only when no
    // track is selected yet (audio-track-controller.ts:195). Selecting synchronously here starts
    // on the wanted track before the default's playlist is requested.
    const index = exactAudioTrackIndex(data.audioTracks, wanted);
    if (index !== -1 && hls.audioTrack !== index) hls.audioTrack = index;
  });
  hls.on(Events.AUDIO_TRACK_SWITCHING, (_event, data) => {
    if (data.lang !== undefined && data.lang !== '') callbacks.onAudioTrack(data.lang);
  });
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
    selectAudioTrack(language) {
      const index = exactAudioTrackIndex(hls.audioTracks, language);
      if (index === -1) return false;
      wanted = language;
      // `nextAudioTrack`, like `nextLevel`: the buffered audio keeps playing instead of an
      // immediate flush. Re-selecting the current track is a no-op without an event.
      if (hls.audioTrack !== index) hls.nextAudioTrack = index;
      return true;
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
    // The native element exposes no track list we report, so there is nothing to select.
    selectAudioTrack: () => false,
    destroy() {
      video.removeEventListener('loadedmetadata', ready);
      video.removeEventListener('error', failed);
      video.removeAttribute('src');
    },
  };
}

export const createPlaybackMedia: CreatePlaybackMedia = (video, callbacks, options) => {
  // This order is load-bearing: Chromium reports native HLS as "maybe" even though hls.js is
  // the branch that provides working playback, quality selection and HTTP error status there.
  if (Hls.isSupported()) return createHlsMedia(video, callbacks, options);
  if (video.canPlayType(hlsMimeType) !== '') return createNativeMedia(video, callbacks);
  return createNativeMedia(video, callbacks);
};
