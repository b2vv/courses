import Hls, { Events, type ErrorData } from 'hls.js';
import {
  exactAudioTrackIndex as languageIndex,
  predictedAudioLanguage,
} from '../application/audio-language';
import { createAudioCoverage } from './audio-coverage';
import type {
  CaptionSelection,
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

/** Exact whole-tag matching, shared with pre-Play prediction and the menu. */
export function exactAudioTrackIndex(
  tracks: readonly ManifestAudioTrack[],
  language: string | null,
): number {
  return languageIndex(
    tracks.map((track) => ({ language: track.lang ?? '' })),
    language,
  );
}

function selectionReporter(callbacks: PlaybackMediaCallbacks) {
  let previous: CaptionSelection | undefined;
  return (selection: CaptionSelection, force = false) => {
    if (
      !force &&
      previous?.state === selection.state &&
      (previous.state !== 'confirmed' ||
        (selection.state === 'confirmed' && previous.language === selection.language)) &&
      (previous.state !== 'pending' ||
        (selection.state === 'pending' && previous.requested === selection.requested))
    )
      return;
    previous = selection;
    callbacks.onCaptionSelection(selection);
  };
}

/**
 * The caption language of a manifest rendition. A rendition without LANGUAGE (the master ffmpeg
 * writes for a fresh upload, before any track rename regenerates it) is a track of unknown language:
 * `und`, which has no captions (D1). Without this the selection never confirms and the panel stays
 * loading (MS4 E6, 2026-11-18).
 */
export function captionLanguageOf(lang: string | undefined): string {
  return lang === undefined || lang === '' ? 'und' : lang;
}

function audioTracksOf(tracks: readonly ManifestAudioTrack[]): PlaybackAudioTrack[] {
  return tracks.flatMap((track) =>
    track.lang === undefined || track.lang === ''
      ? []
      : [{ language: track.lang, label: track.name, isDefault: track.default }],
  );
}

export function createHlsMedia(
  video: HTMLVideoElement,
  callbacks: PlaybackMediaCallbacks,
  options: PlaybackMediaOptions | undefined,
): PlaybackMedia {
  const hls = new Hls({ enableWorker: true });
  const coverage = createAudioCoverage();
  const report = selectionReporter(callbacks);
  let selected: string | null = null;
  let audioBuffer: { readonly buffered: TimeRanges } | undefined;
  let destroyed = false;
  let scheduled = false;
  const selection = () => {
    const buffer = audioBuffer;
    const current = coverage.selection(
      selected,
      video.currentTime,
      buffer === undefined ? undefined : () => buffer.buffered,
    );
    return current.state === 'pending' && selected !== null
      ? { state: 'pending' as const, requested: selected }
      : current;
  };
  const recompute = () => {
    if (destroyed) return;
    // Suppress synchronously, but confirm after nested hls event dispatch has finished.
    const current = selection();
    if (current.state === 'pending') report(current);
    if (scheduled) return;
    scheduled = true;
    queueMicrotask(() => {
      scheduled = false;
      coverage.endDispatch();
      if (!destroyed) report(selection());
    });
  };
  const reset = () => {
    coverage.reset();
    audioBuffer = undefined;
    recompute();
  };
  const mediaEvents = ['timeupdate', 'seeked', 'playing', 'pause', 'progress'] as const;
  for (const event of mediaEvents) video.addEventListener(event, recompute);
  hls.on(Events.BUFFER_CREATED, (_event, data) => {
    audioBuffer = data.tracks.audio?.buffer;
    recompute();
  });
  hls.on(Events.BUFFER_APPENDING, (_event, data) => {
    if (data.type !== 'audio') return;
    const track = hls.audioTracks[data.frag.level];
    coverage.appending(data.frag, track === undefined ? undefined : captionLanguageOf(track.lang));
    recompute();
  });
  hls.on(Events.BUFFER_APPENDED, (_event, data) => {
    if (data.type === 'audio') coverage.appended(data.frag);
    recompute();
  });
  hls.on(Events.FRAG_BUFFERED, (_event, data) => {
    if (data.id !== 'audio') return;
    const track = hls.audioTracks[data.frag.level];
    coverage.buffered(
      data.frag,
      track === undefined ? undefined : captionLanguageOf(track.lang),
      video.currentTime,
    );
    recompute();
  });
  hls.on(Events.BUFFER_FLUSHED, recompute);
  hls.on(Events.BUFFER_EOS, recompute);
  hls.on(Events.BUFFER_RESET, reset);
  hls.on(Events.MEDIA_DETACHING, reset);
  // The language the viewer asked for; kept across manifests so a retry or a new session URL
  // (both `loadSource` on this instance) starts on it again.
  let wanted = options?.audioLanguage ?? null;
  hls.on(Events.AUDIO_TRACKS_UPDATED, (_event, data) => {
    callbacks.onAudioTracks(audioTracksOf(data.audioTracks));
    // The controller triggers this event and only then selects its default, and only when no
    // track is selected yet (audio-track-controller.ts:195). Selecting synchronously here starts
    // on the wanted track before the default's playlist is requested.
    const start = predictedAudioLanguage(audioTracksOf(data.audioTracks), wanted);
    const index = exactAudioTrackIndex(data.audioTracks, start);
    if (index !== -1 && hls.audioTrack !== index) hls.audioTrack = index;
    const current = hls.audioTracks[hls.audioTrack];
    selected = current === undefined ? null : captionLanguageOf(current.lang);
    recompute();
  });
  hls.on(Events.AUDIO_TRACK_SWITCHING, (_event, data) => {
    selected = captionLanguageOf(data.lang);
    // The menu value (S7) still names only tracks the manifest gives a language.
    if (data.lang !== undefined && data.lang !== '') callbacks.onAudioTrack(data.lang);
    recompute();
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
    if (data.frag) coverage.error(data.frag);
    recompute();
    const status = data.response?.code;
    const url = data.url ?? data.frag?.url;
    if (status !== undefined) callbacks.onError({ status, retryable: status === 503, url });
    else if (data.fatal) callbacks.onError({ status: undefined, retryable: false, url });
  });
  return {
    kind: 'hls',
    load(source) {
      selected = null;
      report({ state: 'pending' }, true);
      reset();
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
      wanted = hls.audioTracks[index]?.lang ?? language;
      selected = wanted;
      recompute();
      // `nextAudioTrack`, like `nextLevel`: the buffered audio keeps playing instead of an
      // immediate flush. Re-selecting the current track is a no-op without an event.
      if (hls.audioTrack !== index) hls.nextAudioTrack = index;
      return true;
    },
    destroy() {
      destroyed = true;
      for (const event of mediaEvents) video.removeEventListener(event, recompute);
      hls.destroy();
    },
  };
}

export function createNativeMedia(
  video: HTMLVideoElement,
  callbacks: PlaybackMediaCallbacks,
): PlaybackMedia {
  const report = selectionReporter(callbacks);
  type NativeTracks = EventTarget & {
    readonly length: number;
    readonly [index: number]: { readonly enabled: boolean; readonly language: string };
  };
  function isNativeTracks(value: unknown): value is NativeTracks {
    if (
      !(value instanceof EventTarget) ||
      !('length' in value) ||
      typeof value.length !== 'number' ||
      !Number.isSafeInteger(value.length) ||
      value.length < 0
    )
      return false;
    for (let i = 0; i < value.length; i += 1) {
      const track: unknown = Reflect.get(value, i);
      if (
        typeof track !== 'object' ||
        track === null ||
        !('enabled' in track) ||
        typeof track.enabled !== 'boolean' ||
        !('language' in track) ||
        typeof track.language !== 'string'
      )
        return false;
    }
    return true;
  }
  const trackList = (): NativeTracks | undefined => {
    if (!('audioTracks' in video)) return undefined;
    // audioTracks is not in TypeScript's lib.dom; validate the optional browser API.
    const tracks = video.audioTracks;
    return isNativeTracks(tracks) ? tracks : undefined;
  };
  const selectedLanguage = (list: NativeTracks | undefined) => {
    const enabled = list === undefined ? [] : Array.from(list).filter((track) => track.enabled);
    const language = enabled.length === 1 ? enabled[0]?.language : undefined;
    return language && language.trim() !== '' ? language : undefined;
  };
  let tracks: NativeTracks | undefined;
  const trackEvents = ['change', 'addtrack', 'removetrack'] as const;
  const update = () => {
    const current = trackList();
    if (current !== tracks) {
      for (const event of trackEvents) tracks?.removeEventListener(event, update);
      tracks = current;
      for (const event of trackEvents) tracks?.addEventListener(event, update);
    }
    const language = selectedLanguage(tracks);
    report(language !== undefined ? { state: 'confirmed', language } : { state: 'unavailable' });
  };
  const ready = () => {
    update();
    callbacks.onReady();
  };
  // The native branch never learns which request failed - the element reports only that it did.
  const failed = () => callbacks.onError({ status: undefined, retryable: true, url: undefined });
  video.addEventListener('loadedmetadata', ready);
  video.addEventListener('error', failed);
  tracks = trackList();
  for (const event of trackEvents) tracks?.addEventListener(event, update);
  return {
    kind: 'native',
    load(source) {
      const language = selectedLanguage(trackList());
      report(
        { state: 'pending', ...(language === undefined ? {} : { requested: language }) },
        true,
      );
      video.src = source;
      video.load();
    },
    stopLoad() {
      video.pause();
      video.removeAttribute('src');
      video.load();
    },
    selectLevel() {},
    // Native caption detection does not add a custom native audio menu.
    selectAudioTrack: () => false,
    destroy() {
      for (const event of trackEvents) tracks?.removeEventListener(event, update);
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
