import { afterEach, beforeEach, expect, rstest, test } from '@rstest/core';
import { Window } from 'happy-dom';
import { Events } from 'hls.js';
import { createHlsMedia, createNativeMedia, exactAudioTrackIndex } from './hls-media';
import { predictedAudioLanguage } from '../application/audio-language';
import type { CaptionSelection, PlaybackMediaCallbacks } from '../ports/PlaybackMedia';

const fake = rstest.hoisted(() => {
  class FakeHls {
    static instance: FakeHls | undefined;
    listeners = new Map<string, Array<(data: unknown) => void>>();
    before = new Map<string, (data: unknown) => void>();
    audioTracks: Array<{ lang: string; name: string; default: boolean }> = [];
    levels: Array<{ height: number; name: string }> = [];
    current = -1;
    nextLevel = -1;
    loadSources: string[] = [];
    onLoad: (() => void) | undefined;
    constructor() {
      FakeHls.instance = this;
    }
    on(event: string, listener: unknown) {
      if (typeof listener !== 'function') throw new Error('Expected an event listener');
      const callbacks = this.listeners.get(event) ?? [];
      callbacks.push((data) => listener(event, data));
      this.listeners.set(event, callbacks);
    }
    emit(event: string, data: unknown = {}) {
      this.before.get(event)?.(data);
      for (const listener of this.listeners.get(event) ?? []) listener(data);
    }
    get audioTrack() {
      return this.current;
    }
    set audioTrack(index: number) {
      this.current = index;
      this.emit('hlsAudioTrackSwitching', { lang: this.audioTracks[index]?.lang });
    }
    set nextAudioTrack(index: number) {
      this.audioTrack = index;
    }
    attachMedia() {}
    loadSource(source: string) {
      this.onLoad?.();
      this.loadSources.push(source);
    }
    stopLoad() {}
    destroy() {}
  }
  return { FakeHls };
});

// Preserve real public event names; replace only the engine, so tests exercise production wiring.
import * as realHls from 'hls.js' with { rstest: 'importActual' };
rstest.mock('hls.js', () => ({ ...realHls, default: fake.FakeHls }));

beforeEach(() => {
  const window = new Window();
  rstest.stubGlobal('document', window.document);
  rstest.stubGlobal('EventTarget', window.EventTarget);
  rstest.stubGlobal('Event', window.Event);
});
afterEach(() => {
  rstest.unstubAllGlobals();
});

const track = (lang: string | undefined, name = 'Track', isDefault = false) => ({
  lang,
  name,
  default: isDefault,
});
const cases: ReadonlyArray<{
  languages: readonly string[];
  remembered: string | null;
  defaultIndex: number;
  expected: number;
}> = [
  {
    languages: ['und', 'und-x-original'],
    remembered: 'und-x-original',
    defaultIndex: 0,
    expected: 1,
  },
  { languages: ['und-x-original', 'und'], remembered: 'und', defaultIndex: 0, expected: 1 },
  { languages: ['en', 'en-us'], remembered: 'en-US', defaultIndex: 0, expected: 1 },
  { languages: ['en-US', 'uk'], remembered: 'EN-us', defaultIndex: 1, expected: 0 },
  { languages: ['en-US', 'uk'], remembered: 'en', defaultIndex: 1, expected: 1 },
  { languages: ['en', 'uk'], remembered: 'en-US', defaultIndex: 1, expected: 1 },
  { languages: ['und-x-original', 'uk'], remembered: 'und', defaultIndex: 1, expected: 1 },
  { languages: ['uk', 'en'], remembered: 'de', defaultIndex: 1, expected: 1 },
  { languages: ['uk', 'en'], remembered: null, defaultIndex: -1, expected: 0 },
  { languages: [], remembered: 'en', defaultIndex: -1, expected: -1 },
];

function callbacks() {
  const selections: CaptionSelection[] = [];
  const menu: string[] = [];
  const cb: PlaybackMediaCallbacks = {
    onReady() {},
    onLevels() {},
    onAudioTracks() {},
    onAudioTrack(language) {
      menu.push(language);
    },
    onCaptionSelection(selection) {
      selections.push(selection);
    },
    onError() {},
  };
  return { cb, selections, menu };
}
function ranges(start = 0, end = 10): TimeRanges {
  return { length: 1, start: () => start, end: () => end };
}
function hlsFixture(remembered: string | null = 'en') {
  const video = document.createElement('video');
  video.currentTime = 5;
  const { cb, selections, menu } = callbacks();
  const media = createHlsMedia(video, cb, { audioLanguage: remembered });
  const hls = fake.FakeHls.instance;
  if (!hls) throw new Error('No fake engine created');
  hls.audioTracks = [track('en'), track('de')].map((t) => ({ ...t, lang: t.lang ?? '' }));
  const buffer = { buffered: ranges() };
  function manifest() {
    hls?.emit(Events.AUDIO_TRACKS_UPDATED, { audioTracks: hls.audioTracks });
  }
  function created() {
    hls?.emit(Events.BUFFER_CREATED, { tracks: { audio: { buffer } } });
  }
  const frag = { start: 0, end: 10, level: 0 };
  function append(f = frag) {
    hls?.emit(Events.BUFFER_APPENDING, { type: 'audio', frag: f });
    hls?.emit(Events.BUFFER_APPENDED, { type: 'audio', frag: f });
    hls?.emit(Events.FRAG_BUFFERED, { id: 'audio', frag: f });
  }
  return { video, media, hls, selections, menu, buffer, manifest, created, append, frag };
}

// M04: this single table exercises prediction and the actual adapter start selection.
test.each(cases)(
  'prediction and adapter agree for %j',
  async ({ languages, remembered, defaultIndex, expected }) => {
    const tracks = languages.map((language, i) => ({ language, isDefault: i === defaultIndex }));
    const predicted = predictedAudioLanguage(tracks, remembered);
    expect(predicted).toBe(languages[expected] ?? null);
    const { hls, media, manifest } = hlsFixture(remembered);
    hls.audioTracks = languages.map((lang, i) => ({
      lang,
      name: 'Shared name',
      default: i === defaultIndex,
    }));
    media.load('/master.m3u8');
    manifest();
    expect(hls.audioTrack).toBe(expected);
    expect(exactAudioTrackIndex(hls.audioTracks, predicted)).toBe(expected);
    await Promise.resolve();
    media.destroy();
  },
);

test('M7 ignores case on both sides, with no prefix matches or name matches', () => {
  expect(exactAudioTrackIndex([track('en-us')], 'en-US')).toBe(0);
  expect(exactAudioTrackIndex([track('EN-us')], 'en-US')).toBe(0);
  expect(exactAudioTrackIndex([track('en-us')], 'en')).toBe(-1);
  expect(exactAudioTrackIndex([track('en')], 'en-us')).toBe(-1);
  expect(exactAudioTrackIndex([track('und-x-original')], 'und')).toBe(-1);
  expect(exactAudioTrackIndex([track(undefined)], 'und')).toBe(-1);
  expect(exactAudioTrackIndex([track('uk', 'Dub'), track('en', 'Dub')], 'EN')).toBe(1);
});

test('load reports pending synchronously before loadSource, including a second still-pending load', async () => {
  const { hls, media, selections } = hlsFixture();
  hls.onLoad = () => expect(selections.at(-1)).toEqual({ state: 'pending' });
  media.load('/first');
  expect(selections).toEqual([{ state: 'pending' }]);
  media.load('/second');
  expect(selections).toEqual([{ state: 'pending' }, { state: 'pending' }]);
  await Promise.resolve();
  expect(selections).toHaveLength(2);
  media.destroy();
});

test('paused map changes confirm only once; selection cancellation needs no SWITCHED; future de gives pending', async () => {
  const { media, hls, video, selections, menu, buffer, manifest, created, append, frag } =
    hlsFixture();
  media.load('/master');
  manifest();
  created();
  append();
  await Promise.resolve();
  expect(selections).toEqual([
    { state: 'pending' },
    { state: 'pending', requested: 'en' },
    { state: 'confirmed', language: 'en' },
  ]);
  video.dispatchEvent(new Event('pause'));
  hls.emit(Events.BUFFER_APPENDED, { type: 'audio', frag });
  await Promise.resolve();
  expect(selections).toHaveLength(3);
  expect(media.selectAudioTrack('DE')).toBe(true);
  expect(selections.at(-1)).toEqual({ state: 'pending', requested: 'de' });
  expect(media.selectAudioTrack('EN')).toBe(true);
  await Promise.resolve();
  expect(selections.at(-1)).toEqual({ state: 'confirmed', language: 'en' });
  expect(menu).toEqual(['en', 'de', 'en']);
  append({ start: 10, end: 20, level: 1 });
  buffer.buffered = ranges(0, 20);
  video.currentTime = 15;
  video.dispatchEvent(new Event('timeupdate'));
  await Promise.resolve();
  expect(selections.at(-1)).toEqual({ state: 'pending', requested: 'en' });
  media.destroy();
});

test.each(['before', 'after'] as const)(
  'nested ERROR %s our APPENDED listener never emits confirmed',
  async (order) => {
    const { media, hls, selections, manifest, created, frag } = hlsFixture();
    const nested = () => {
      hls.emit(Events.FRAG_BUFFERED, { id: 'audio', frag });
      hls.emit(Events.ERROR, { frag, fatal: false, details: 'bufferAppendNoProgress' });
    };
    if (order === 'before') hls.before.set(Events.BUFFER_APPENDED, nested);
    else hls.on(Events.BUFFER_APPENDED, nested);
    media.load('/master');
    manifest();
    created();
    hls.emit(Events.BUFFER_APPENDING, { type: 'audio', frag });
    hls.emit(Events.BUFFER_APPENDED, { type: 'audio', frag });
    await Promise.resolve();
    expect(selections).toEqual([{ state: 'pending' }, { state: 'pending', requested: 'en' }]);
    media.destroy();
  },
);

test('nested FRAG_BUFFERED before our APPENDED listener confirms only after dispatch', async () => {
  const { media, hls, selections, manifest, created, frag } = hlsFixture();
  hls.before.set(Events.BUFFER_APPENDED, () =>
    hls.emit(Events.FRAG_BUFFERED, { id: 'audio', frag }),
  );
  media.load('/master');
  manifest();
  created();
  hls.emit(Events.BUFFER_APPENDING, { type: 'audio', frag });
  hls.emit(Events.BUFFER_APPENDED, { type: 'audio', frag });
  expect(selections).toEqual([{ state: 'pending' }, { state: 'pending', requested: 'en' }]);
  await Promise.resolve();
  expect(selections.at(-1)).toEqual({ state: 'confirmed', language: 'en' });
  media.destroy();
});

test.each([
  'timeupdate',
  'seeked',
  'playing',
  'pause',
  'progress',
  Events.BUFFER_FLUSHED,
  Events.BUFFER_EOS,
])('%s rechecks actual audio buffered coverage', async (event) => {
  const { media, hls, video, selections, buffer, manifest, created, append } = hlsFixture();
  media.load('/master');
  manifest();
  created();
  append();
  await Promise.resolve();
  expect(selections.at(-1)?.state).toBe('confirmed');
  buffer.buffered = ranges(0, 4); // video.buffered may still include 5, especially at EOS
  if (event.startsWith('hls')) hls.emit(event);
  else video.dispatchEvent(new Event(event));
  expect(selections.at(-1)).toEqual({ state: 'pending', requested: 'en' });
  await Promise.resolve();
  media.destroy();
});

test.each([Events.BUFFER_RESET, Events.MEDIA_DETACHING, 'load'])(
  '%s forgets both labels and SourceBuffer reference',
  async (event) => {
    const { media, hls, selections, manifest, created, append } = hlsFixture();
    media.load('/master');
    manifest();
    created();
    append();
    await Promise.resolve();
    if (event === 'load') {
      media.load('/again');
      manifest();
    } else hls.emit(event);
    expect(selections.at(-1)).toEqual({ state: 'pending', requested: 'en' });
    append(); // the old SourceBuffer reference must not be enough
    await Promise.resolve();
    expect(selections.at(-1)).toEqual({ state: 'pending', requested: 'en' });
    created();
    await Promise.resolve();
    expect(selections.at(-1)).toEqual({ state: 'confirmed', language: 'en' });
    media.destroy();
  },
);

function audioTrackList() {
  class AudioTrackList extends EventTarget {
    [index: number]: { enabled: boolean; language: string };
    length = 0;
    replace(tracks: Array<{ enabled: boolean; language: string }>) {
      for (let i = 0; i < this.length; i += 1) delete this[i];
      this.length = tracks.length;
      tracks.forEach((track, i) => {
        this[i] = track;
      });
    }
  }
  return new AudioTrackList();
}

function nativeFixture() {
  const video = document.createElement('video');
  const tracks = audioTrackList();
  tracks.replace([{ enabled: true, language: 'en' }]);
  Object.defineProperty(video, 'audioTracks', { configurable: true, value: tracks });
  const { cb, selections } = callbacks();
  const media = createNativeMedia(video, cb);
  return { video, tracks, media, selections };
}

test('native pending precedes load; loadedmetadata confirms exactly one enabled nonempty language', () => {
  const { video, media, selections } = nativeFixture();
  Object.defineProperty(video, 'load', {
    value: () => expect(selections.at(-1)).toEqual({ state: 'pending', requested: 'en' }),
  });
  expect(selections).toEqual([]);
  media.load('/native');
  video.dispatchEvent(new Event('loadedmetadata'));
  expect(selections).toEqual([
    { state: 'pending', requested: 'en' },
    { state: 'confirmed', language: 'en' },
  ]);
  media.load('/again');
  expect(selections.at(-1)).toEqual({ state: 'pending', requested: 'en' });
  media.destroy();
});

test.each(['change', 'addtrack', 'removetrack'])(
  'native %s invalidates valid → empty/missing and deduplicates',
  (event) => {
    const { video, tracks, media, selections } = nativeFixture();
    video.dispatchEvent(new Event('loadedmetadata'));
    tracks.replace([{ enabled: true, language: '' }]);
    tracks.dispatchEvent(new Event(event));
    expect(selections.at(-1)).toEqual({ state: 'unavailable' });
    tracks.dispatchEvent(new Event(event));
    expect(selections).toHaveLength(2);
    tracks.replace([{ enabled: true, language: 'de' }]);
    tracks.dispatchEvent(new Event(event));
    expect(selections.at(-1)).toEqual({ state: 'confirmed', language: 'de' });
    tracks.replace([]);
    tracks.dispatchEvent(new Event(event));
    expect(selections.at(-1)).toEqual({ state: 'unavailable' });
    tracks.replace([{ enabled: true, language: 'en' }]);
    tracks.dispatchEvent(new Event(event));
    Object.defineProperty(video, 'audioTracks', { value: undefined });
    tracks.dispatchEvent(new Event(event));
    expect(selections.at(-1)).toEqual({ state: 'unavailable' });
    media.destroy();
  },
);

test.each(
  [
    [],
    [{ enabled: false, language: 'en' }],
    [
      { enabled: true, language: 'en' },
      { enabled: true, language: 'de' },
    ],
    [{ enabled: true, language: ' ' }],
  ].map((list) => ({ list })),
)('native initial ambiguous/empty selection %j is unavailable', ({ list }) => {
  const { video, tracks, media, selections } = nativeFixture();
  tracks.replace(list);
  video.dispatchEvent(new Event('loadedmetadata'));
  expect(selections).toEqual([{ state: 'unavailable' }]);
  media.destroy();
});

test('native without audioTracks API is unavailable on metadata and destroys its listeners', () => {
  const video = document.createElement('video');
  const { cb, selections } = callbacks();
  const media = createNativeMedia(video, cb);
  media.load('/native');
  video.dispatchEvent(new Event('loadedmetadata'));
  expect(selections).toEqual([{ state: 'pending' }, { state: 'unavailable' }]);
  media.destroy();
  video.dispatchEvent(new Event('loadedmetadata'));
  expect(selections).toHaveLength(2);
});

test('destroy removes media listeners and never reports from queued or late events', async () => {
  const { media, hls, video, selections, manifest, created, append } = hlsFixture();
  media.load('/master');
  manifest();
  created();
  append();
  media.destroy();
  hls.emit(Events.MEDIA_DETACHING);
  video.dispatchEvent(new Event('progress'));
  await Promise.resolve();
  expect(selections).toEqual([{ state: 'pending' }, { state: 'pending', requested: 'en' }]);
});

// MS4 E6 (2026-11-18): a fresh upload's master names its embedded rendition without LANGUAGE. That
// track is of unknown language, so the selection confirms `und` (no captions, D1) instead of staying
// pending forever; the S7 menu still gets no value for it.
test('a rendition without LANGUAGE confirms und and adds no menu value', async () => {
  const { media, hls, selections, menu, manifest, created, append } = hlsFixture(null);
  hls.audioTracks = [{ lang: '', name: 'audio_0', default: true }];
  media.load('/master');
  manifest();
  created();
  // The real controller selects the default after AUDIO_TRACKS_UPDATED and emits SWITCHING.
  hls.audioTrack = 0;
  append();
  await Promise.resolve();
  expect(selections.at(-1)).toEqual({ state: 'confirmed', language: 'und' });
  expect(menu).toEqual([]);
  media.destroy();
});

test('an immediate failed replacement retry stays pending through the audio event tap', async () => {
  const { media, hls, video, selections, buffer, manifest, created, append } = hlsFixture('de');
  video.currentTime = 0;
  buffer.buffered = ranges(0, 6);
  media.load('/master');
  manifest();
  created();
  append({ start: 0, end: 6, level: 1 });
  await Promise.resolve();
  expect(selections.at(-1)).toEqual({ state: 'confirmed', language: 'de' });
  video.currentTime = 2.5;
  hls.audioTrack = 0;
  const retry = { start: 0, end: 6, level: 0 };
  hls.emit(Events.BUFFER_APPENDING, { type: 'audio', frag: retry });
  hls.emit(Events.ERROR, { frag: retry, fatal: false, details: 'bufferAppendError' });
  await Promise.resolve();
  video.currentTime = 2.6;
  append(retry);
  await Promise.resolve();
  expect(selections.at(-1)?.state).toBe('pending');
  video.currentTime = 4;
  video.dispatchEvent(new Event('timeupdate'));
  await Promise.resolve();
  expect(selections.at(-1)?.state).toBe('pending');
  video.currentTime = 4.1;
  video.dispatchEvent(new Event('timeupdate'));
  await Promise.resolve();
  expect(selections.at(-1)).toEqual({ state: 'confirmed', language: 'en' });
  media.destroy();
});

test('failed partially appended en cannot immediately reconfirm de on return', async () => {
  const { media, hls, video, selections, buffer, manifest, created, append } = hlsFixture('de');
  video.currentTime = 0;
  buffer.buffered = ranges(0, 6);
  media.load('/master');
  manifest();
  created();
  append({ start: 0, end: 6, level: 1 });
  await Promise.resolve();
  expect(selections.at(-1)).toEqual({ state: 'confirmed', language: 'de' });
  video.currentTime = 2.5;
  hls.audioTrack = 0;
  const failed = { start: 0, end: 6, level: 0 };
  hls.emit(Events.BUFFER_APPENDING, { type: 'audio', frag: failed });
  hls.emit(Events.BUFFER_APPENDED, { type: 'audio', frag: failed });
  hls.emit(Events.ERROR, { frag: failed, fatal: false, details: 'bufferAppendError' });
  await Promise.resolve();
  video.currentTime = 2.6;
  hls.audioTrack = 1;
  append({ start: 0, end: 6, level: 1 });
  await Promise.resolve();
  expect(selections.at(-1)).toEqual({ state: 'pending', requested: 'de' });
  video.currentTime = 4.09;
  video.dispatchEvent(new Event('timeupdate'));
  await Promise.resolve();
  expect(selections.at(-1)?.state).toBe('pending');
  video.currentTime = 4.1;
  video.dispatchEvent(new Event('timeupdate'));
  await Promise.resolve();
  expect(selections.at(-1)).toEqual({ state: 'confirmed', language: 'de' });
  media.destroy();
});

test('initial selection and successive SWITCHING languages report requested before any confirmation', async () => {
  const { media, hls, selections, manifest } = hlsFixture();
  media.load('/master');
  manifest();
  expect(selections.at(-1)).toEqual({ state: 'pending', requested: 'en' });
  hls.audioTrack = 1;
  expect(selections.at(-1)).toEqual({ state: 'pending', requested: 'de' });
  const count = selections.length;
  hls.audioTrack = 1;
  await Promise.resolve();
  expect(selections).toHaveLength(count);
  hls.audioTrack = 0;
  expect(selections.at(-1)).toEqual({ state: 'pending', requested: 'en' });
  expect(selections.every((selection) => selection.state === 'pending')).toBe(true);
  media.destroy();
});

test('native load reports a known requested language before metadata confirms it', () => {
  const { media, video, selections } = nativeFixture();
  media.load('/native');
  expect(selections).toEqual([{ state: 'pending', requested: 'en' }]);
  video.dispatchEvent(new Event('loadedmetadata'));
  expect(selections.at(-1)).toEqual({ state: 'confirmed', language: 'en' });
  media.destroy();
});
