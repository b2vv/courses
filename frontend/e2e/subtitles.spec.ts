import { execFile } from 'node:child_process';
import { randomUUID } from 'node:crypto';
import { writeFile } from 'node:fs/promises';
import path from 'node:path';
import { promisify } from 'node:util';
import { expect, test, type APIRequestContext, type Page, type Route } from '@playwright/test';
import { z } from 'zod';
import { authStatePath } from './auth';

// MS4 r9 M9/M10. Run the entire file: beforeAll prepares the course, E8 patches E6's lesson.
// A failed serial row stops that run; run the remaining rows with --no-deps -g after setup,
// whose beforeAll creates a fresh course. Never reuse UUIDs from another worker/run.
const namespace = `e2e-subtitles-${Date.now().toString(36)}-${randomUUID().slice(0, 8)}`;
const fixtureDirectory = path.join(import.meta.dirname, '..', 'playwright', '.fixtures');
const execute = promisify(execFile);
const root = path.resolve(import.meta.dirname, '../..');
const ids = { a: randomUUID(), b: randomUUID(), und: randomUUID() };
const tokenSchema = z.object({ id: z.string().uuid(), updatedAt: z.string() });
const videoSchema = z.object({
  items: z.array(
    z.object({ videoId: z.string().uuid(), lessonId: z.string().uuid(), state: z.string() }),
  ),
});
const tracksSchema = z.object({
  items: z.array(
    z.object({
      trackId: z.string().uuid(),
      kind: z.string(),
      language: z.string(),
      state: z.string(),
    }),
  ),
});
let courseID = '';
let deID = '';
let admin: APIRequestContext;
let headers: { Origin: string };
const english = 'English caption 440';
const german = 'Deutscher Untertitel 880';
const player = (page: Page) => page.locator('section.lesson-player');
const panel = (page: Page) => page.getByRole('region', { name: 'Транскрипт', exact: true });
const lessonPath = (id = ids.a) => `/learn/${namespace}/${id}`;
const requestLogs = new WeakMap<Page, object[]>();

async function putTranscript(lesson: string, language: string) {
  const response = await admin.put(`/api/v1/admin/lessons/${lesson}/transcript`, {
    headers,
    data: {
      language,
      segments: [
        {
          id: randomUUID(),
          startMs: 0,
          endMs: 30_000,
          text: language === 'en' ? english : german,
          documentIds: [],
        },
      ],
    },
  });
  expect(response.status(), await response.text()).toBe(200);
}
async function renameEmbedded(lesson: string, language: string) {
  const videos = await admin.get(`/api/v1/admin/courses/${courseID}/videos`);
  expect(videos.status()).toBe(200);
  const video = videoSchema
    .parse(await videos.json())
    .items.find((item) => item.lessonId === lesson && item.state === 'ready');
  if (!video) throw new Error(`Missing Ready video for ${lesson}`);
  const response = await admin.get(`/api/v1/admin/videos/${video.videoId}/audio-tracks`);
  expect(response.status()).toBe(200);
  const embedded = tracksSchema
    .parse(await response.json())
    .items.find((track) => track.kind === 'embedded');
  if (!embedded) throw new Error('Missing embedded track');
  const renamed = await admin.patch(
    `/api/v1/admin/videos/${video.videoId}/audio-tracks/${embedded.trackId}`,
    {
      headers,
      data: { language, label: 'English' },
    },
  );
  expect(renamed.status(), await renamed.text()).toBe(200);
  return video.videoId;
}
async function resetProgress() {
  // Only this run's lessons. No seed reset and no global Docker context change.
  await execute(
    'docker',
    [
      'compose',
      'exec',
      '-T',
      'postgres',
      'sh',
      '-eu',
      '-c',
      `exec psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c "DELETE FROM lesson_progress WHERE lesson_id IN ('${ids.a}', '${ids.b}', '${ids.und}')"`,
    ],
    { cwd: root, timeout: 60_000 },
  );
}
async function open(page: Page, lesson = ids.a, remembered: string | null = null) {
  // New browser context per row; once per context so SPA navigation retains a real menu choice.
  await page.addInitScript((audioLanguage) => {
    if (sessionStorage.getItem('ms4-prepared')) return;
    localStorage.setItem(
      'courses.player.prefs.v1',
      JSON.stringify({
        volume: 1,
        muted: false,
        rate: 1,
        captionsOn: false,
        autoplayNext: false,
        audioLanguage,
      }),
    );
    sessionStorage.setItem('ms4-prepared', 'yes');
  }, remembered);
  await page.goto(lessonPath(lesson));
}
async function captionsOn(page: Page) {
  await page.getByRole('button', { name: 'Налаштування' }).click();
  const toggle = page.getByRole('switch', { name: 'Субтитри', exact: true });
  await expect(toggle).toBeVisible({ timeout: 60_000 });
  if (!(await toggle.isChecked())) await page.locator('label').filter({ has: toggle }).click();
  await expect(toggle).toBeChecked();
  await page.keyboard.press('Escape');
}
async function choose(page: Page, label: string) {
  await page.getByRole('button', { name: 'Налаштування' }).click();
  await page
    .getByRole('radiogroup', { name: 'Аудіо', exact: true })
    .locator('label', { hasText: label })
    .click();
  await page.keyboard.press('Escape');
}
async function currentCaption(page: Page) {
  return page.evaluate(() => {
    const element = document.querySelector('video track');
    if (!(element instanceof HTMLTrackElement))
      return { language: null, mode: null, cues: [], active: [] };
    const text = (cue: TextTrackCue) => (cue instanceof VTTCue ? cue.text : '');
    return {
      language: element.srclang,
      mode: element.track.mode,
      cues: Array.from(element.track.cues ?? []).map(text),
      active:
        element.track.mode === 'showing'
          ? Array.from(element.track.activeCues ?? []).map(text)
          : [],
    };
  });
}
async function showing(page: Page, language: 'en' | 'de') {
  await expect
    .poll(() => currentCaption(page), { timeout: 60_000 })
    .toMatchObject({
      language,
      mode: 'showing',
      cues: [language === 'en' ? english : german],
      active: [language === 'en' ? english : german],
    });
}
async function noShowing(page: Page) {
  expect((await currentCaption(page)).active).toEqual([]);
}
async function start(page: Page, language: 'en' | 'de' = 'en') {
  await page.getByRole('button', { name: 'Відтворити', exact: true }).click();
  await expect(player(page)).toHaveAttribute('data-caption-selection', `confirmed:${language}`, {
    timeout: 60_000,
  });
  await tapAudio(page);
  await waitForTone(page, language === 'en' ? 440 : 880, 3, 60_000);
  await captionsOn(page);
  await showing(page, language);
}
function transcriptURL(url: URL, language: string, lesson = ids.a) {
  return (
    url.pathname === `/api/v1/learn/lessons/${lesson}/transcript` &&
    url.searchParams.get('language') === language
  );
}
function gate() {
  let release = () => {};
  const promise = new Promise<void>((resolve) => {
    release = resolve;
  });
  return { promise, release };
}
type Peak = { hz: number; db: number; state: string; tracks: number; time: number };

/**
 * Taps the video's decoded audio for an analyser. Never connected to the destination, so the
 * player's own output path is unchanged.
 */
async function tapAudio(page: Page) {
  await page.locator('video').evaluate((node) => {
    if (!(node instanceof HTMLVideoElement)) throw new Error('Video element is unavailable');
    const stream = (node as HTMLVideoElement & { captureStream(): MediaStream }).captureStream();
    const context = new AudioContext();
    const analyser = context.createAnalyser();
    analyser.fftSize = 8192;
    context.createMediaStreamSource(stream).connect(analyser);
    Object.assign(window, { s7Tap: { context, analyser, stream } });
  });
}

async function peak(page: Page): Promise<Peak> {
  return page.evaluate(() => {
    const tap = (
      window as unknown as {
        s7Tap: { context: AudioContext; analyser: AnalyserNode; stream: MediaStream };
      }
    ).s7Tap;
    const bins = new Float32Array(tap.analyser.frequencyBinCount);
    tap.analyser.getFloatFrequencyData(bins);
    let top = 0;
    for (let index = 1; index < bins.length; index++) {
      if ((bins[index] ?? -Infinity) > (bins[top] ?? -Infinity)) top = index;
    }
    const video = document.querySelector('video');
    return {
      hz: (top * tap.context.sampleRate) / tap.analyser.fftSize,
      db: bins[top] ?? -Infinity,
      state: tap.context.state,
      tracks: tap.stream.getAudioTracks().length,
      time: video?.currentTime ?? -1,
    };
  });
}

function isTone(sample: Peak, hz: number): boolean {
  return Math.abs(sample.hz - hz) <= 12 && sample.db >= -60;
}

/** Waits for `count` consecutive samples (300 ms apart) of the tone; returns ms to the first. */
async function waitForTone(page: Page, hz: number, count: number, withinMs: number) {
  const started = Date.now();
  let run = 0;
  let firstAt = -1;
  const seen: Peak[] = [];
  while (Date.now() - started <= withinMs) {
    const sample = await peak(page);
    seen.push(sample);
    if (isTone(sample, hz)) {
      if (run === 0) firstAt = Date.now() - started;
      run += 1;
      if (run >= count) return { firstAt, seen };
    } else {
      run = 0;
    }
    await page.waitForTimeout(300);
  }
  throw new Error(
    `No ${count} consecutive ${hz} Hz samples in ${withinMs} ms: ${JSON.stringify(seen)}`,
  );
}

// ISO-BMFF reader used both at the network barrier and synchronously inside appendBuffer.
// The fingerprint covers concatenated mdat payloads, including length; init bytes never qualify.
function bmff(bytes: Uint8Array, inheritedDuration = 0) {
  const view = new DataView(bytes.buffer, bytes.byteOffset, bytes.byteLength);
  type Box = { type: string; p: number; end: number };
  function boxes(start: number, end: number): Box[] {
    const result: Box[] = [];
    for (let offset = start; offset + 8 <= end;) {
      let size = view.getUint32(offset);
      const type = String.fromCharCode(...bytes.subarray(offset + 4, offset + 8));
      let header = 8;
      if (size === 1) {
        size = Number(view.getBigUint64(offset + 8));
        header = 16;
      }
      if (size === 0) size = end - offset;
      if (size < header || offset + size > end) throw new Error(`Invalid ${type} box`);
      result.push({ type, p: offset + header, end: offset + size });
      offset += size;
    }
    return result;
  }
  const top = boxes(0, bytes.length);
  const all: Box[] = [];
  function descend(items: Box[]) {
    for (const box of items) {
      all.push(box);
      if (['moov', 'trak', 'mdia', 'mvex', 'moof', 'traf'].includes(box.type))
        descend(boxes(box.p, box.end));
    }
  }
  descend(top);
  const first = top[0]?.type;
  const kind =
    first === 'ftyp' || first === 'moov'
      ? 'init'
      : first === 'moof' || first === 'styp'
        ? 'media'
        : 'other';
  let timescale = 0;
  let defaultDuration = inheritedDuration;
  let tfdt = 0;
  let ticks = 0;
  for (const box of all) {
    const p = box.p;
    if (box.type === 'mdhd') timescale = view.getUint32(p + (bytes[p] === 1 ? 20 : 12));
    if (box.type === 'trex') defaultDuration = view.getUint32(p + 12);
    if (box.type === 'tfdt')
      tfdt = bytes[p] === 1 ? Number(view.getBigUint64(p + 4)) : view.getUint32(p + 4);
    if (box.type === 'tfhd') {
      const flags = view.getUint32(p) & 0xffffff;
      let offset = p + 8;
      if (flags & 1) offset += 8;
      if (flags & 2) offset += 4;
      if (flags & 8) defaultDuration = view.getUint32(offset);
    }
    if (box.type === 'trun') {
      const flags = view.getUint32(p) & 0xffffff;
      const count = view.getUint32(p + 4);
      let offset = p + 8;
      if (flags & 1) offset += 4;
      if (flags & 4) offset += 4;
      for (let sample = 0; sample < count; sample++) {
        ticks += flags & 0x100 ? view.getUint32(offset) : defaultDuration;
        if (flags & 0x100) offset += 4;
        if (flags & 0x200) offset += 4;
        if (flags & 0x400) offset += 4;
        if (flags & 0x800) offset += 4;
      }
    }
  }
  let hash = 0xcbf29ce484222325n;
  let length = 0;
  for (const box of top.filter((box) => box.type === 'mdat')) {
    for (const byte of bytes.subarray(box.p, box.end)) {
      hash = BigInt.asUintN(64, (hash ^ BigInt(byte)) * 0x100000001b3n);
      length++;
    }
  }
  return {
    kind,
    timescale,
    defaultDuration,
    tfdt,
    ticks,
    mdatFnv: length ? `${length}:${hash.toString(16).padStart(16, '0')}` : null,
  };
}
type AppendEntry = {
  seq: number;
  kind: string;
  tfdt: number;
  timescale: number;
  timestampOffset: number;
  duration: number;
  mdatFnv: string | null;
  outcome: string;
  name?: string;
};
type Probe = {
  entries: AppendEntry[];
  de: Set<string>;
  inject: boolean;
  failed: string | null;
  held: number | null;
  release: (() => void) | null;
  states: { selection: string | null; track: string | null; text: string }[];
  forbiddenEnglish: boolean;
};
type E4bTimings = {
  releasedAt: number | null;
  releaseMediaTime: number | null;
  firstEnCueAt: number | null;
  firstEnCueMediaTime: number | null;
  pending: { start: number; end: number | null; startAt: number; endAt: number | null }[];
};
declare global {
  interface Window {
    ms4Parse: typeof bmff;
    ms4: Probe;
    ms4Times: { clicked: number; confirmed: number; showing: number };
    ms4E4bTimes: E4bTimings;
  }
}
function installProbe() {
  const probe: Probe = {
    entries: [],
    de: new Set(),
    inject: false,
    failed: null,
    held: null,
    release: null,
    states: [],
    forbiddenEnglish: false,
  };
  window.ms4 = probe;
  const audio = new WeakMap<SourceBuffer, { timescale: number; defaultDuration: number }>();
  const add = MediaSource.prototype.addSourceBuffer;
  MediaSource.prototype.addSourceBuffer = function (mime) {
    const buffer = add.call(this, mime);
    if (mime.startsWith('audio/')) audio.set(buffer, { timescale: 0, defaultDuration: 0 });
    return buffer;
  };
  const append = SourceBuffer.prototype.appendBuffer;
  SourceBuffer.prototype.appendBuffer = function (data) {
    const state = audio.get(this);
    if (!state) return append.call(this, data);
    const bytes = ArrayBuffer.isView(data)
      ? new Uint8Array(data.buffer, data.byteOffset, data.byteLength)
      : new Uint8Array(data);
    const parsed = window.ms4Parse(bytes, state.defaultDuration);
    if (parsed.kind === 'init') {
      state.timescale = parsed.timescale;
      state.defaultDuration = parsed.defaultDuration;
    }
    const entry: AppendEntry = {
      seq: probe.entries.length + 1,
      kind: parsed.kind,
      tfdt: parsed.tfdt,
      timescale: state.timescale,
      timestampOffset: this.timestampOffset,
      duration: parsed.ticks / state.timescale,
      mdatFnv: parsed.mdatFnv,
      outcome: 'pending',
    };
    probe.entries.push(entry);
    if (
      probe.inject &&
      parsed.kind === 'media' &&
      parsed.mdatFnv &&
      probe.de.has(parsed.mdatFnv) &&
      !probe.failed
    ) {
      probe.failed = parsed.mdatFnv;
      const injected = new DOMException('injected media append failure', 'InvalidStateError');
      entry.outcome = 'threw';
      entry.name = injected.name;
      throw injected;
    }
    const perform = () => {
      let errored = false;
      const error = () => {
        errored = true;
        entry.outcome = 'error';
      };
      this.addEventListener('error', error, { once: true });
      this.addEventListener(
        'updateend',
        () => {
          this.removeEventListener('error', error);
          if (!errored) entry.outcome = 'ok';
        },
        { once: true },
      );
      try {
        append.call(this, data);
      } catch (error) {
        entry.outcome = 'threw';
        entry.name = error instanceof DOMException ? error.name : String(error);
        throw error;
      }
    };
    if (
      probe.failed &&
      parsed.kind === 'media' &&
      parsed.mdatFnv === probe.failed &&
      probe.held === null
    ) {
      probe.held = entry.seq;
      probe.release = perform;
      return;
    }
    perform();
  };
  const observe = () => {
    const section = document.querySelector('section.lesson-player');
    const track = document.querySelector('video track');
    const text = document.querySelector('.transcript')?.textContent ?? '';
    const snapshot = {
      selection: section?.getAttribute('data-caption-selection') ?? null,
      track: track instanceof HTMLTrackElement ? track.srclang : null,
      text,
    };
    if (section) probe.states.push(snapshot);
    if (
      text.includes('English caption 440') ||
      (track instanceof HTMLTrackElement &&
        Array.from(track.track.cues ?? []).some(
          (cue) => cue instanceof VTTCue && cue.text.includes('English caption 440'),
        ))
    )
      probe.forbiddenEnglish = true;
  };
  new MutationObserver(observe).observe(document, {
    subtree: true,
    childList: true,
    attributes: true,
    characterData: true,
  });
  setInterval(observe, 50);
}
async function probe(page: Page) {
  await page.addInitScript({
    content: `window.ms4Parse = (${bmff.toString()}); (${installProbe.toString()})();`,
  });
}
async function deBarrier(page: Page) {
  const log: { pathname: string; fingerprint: string | null }[] = [];
  await page.route(new RegExp(`/audio-${deID}/[^/]+\\.m4s(?:\\?.*)?$`), async (route) => {
    const response = await route.fetch();
    const body = await response.body();
    const parsed = bmff(body);
    if (parsed.kind === 'media') {
      expect(parsed.mdatFnv).not.toBeNull();
      const acknowledged = await page.evaluate((fingerprint) => {
        if (!fingerprint) return false;
        window.ms4.de.add(fingerprint);
        return window.ms4.de.has(fingerprint);
      }, parsed.mdatFnv);
      expect(acknowledged).toBe(true);
      log.push({ pathname: new URL(route.request().url()).pathname, fingerprint: parsed.mdatFnv });
    }
    await route.fulfill({ response, body });
  });
  return log;
}
async function samples(page: Page, milliseconds: number) {
  const result = [];
  const until = Date.now() + milliseconds;
  while (Date.now() < until) {
    result.push({
      at: Date.now(),
      ...(await peak(page)),
      caption: await currentCaption(page),
      selection: await player(page).getAttribute('data-caption-selection'),
    });
    await page.waitForTimeout(300);
  }
  return result;
}

test.describe('MS4 captions follow decoded audio', () => {
  test.describe.configure({ mode: 'serial', timeout: 300_000 });
  test.use({ storageState: authStatePath('student') });
  test.beforeAll(async ({ browser, baseURL }) => {
    test.setTimeout(900_000);
    if (!baseURL) throw new Error('baseURL required');
    headers = { Origin: new URL(baseURL).origin };
    const context = await browser.newContext({
      baseURL,
      storageState: authStatePath('admin'),
      locale: 'uk-UA',
    });
    admin = context.request;
    const page = await context.newPage();
    try {
      const draft = {
        slug: namespace,
        title: namespace,
        summary: 'MS4 real captions fixture',
        description: 'Run scoped MS4 captions',
        instructorId: '40000000-0000-4000-8000-000000000003',
      };
      const created = await admin.post('/api/v1/admin/courses', { headers, data: draft });
      expect(created.status(), await created.text()).toBe(201);
      const first = tokenSchema.parse(await created.json());
      courseID = first.id;
      const saved = await admin.put(`/api/v1/admin/courses/${courseID}`, {
        headers,
        data: {
          expectedUpdatedAt: first.updatedAt,
          course: draft,
          modules: [
            {
              id: randomUUID(),
              title: 'MS4',
              lessons: [
                { id: ids.a, title: 'MS4 A' },
                { id: ids.b, title: 'MS4 B' },
                { id: ids.und, title: 'MS4 und' },
              ],
            },
          ],
        },
      });
      expect(saved.status(), await saved.text()).toBe(200);
      await page.goto(`/admin/courses/${courseID}`);
      await page.getByRole('tab', { name: /Lessons|Уроки/ }).click();
      for (const [index, lesson] of Object.values(ids).entries()) {
        const group = page
          .getByRole('group', { name: 'Модуль 1' })
          .getByRole('group', { name: `Урок ${index + 1}`, exact: true });
        const minted = page.waitForResponse(
          (r) =>
            r.url().endsWith(`/admin/lessons/${lesson}/video`) && r.request().method() === 'POST',
        );
        await group
          .locator('input[type="file"]:not(.admin-audio-field input)')
          .setInputFiles(path.join(fixtureDirectory, 'e2e-640x360-30s.mp4'));
        expect((await minted).status()).toBe(201);
        await expect(group.locator('[data-video-state="ready"]')).toBeVisible({ timeout: 240_000 });
        if (lesson === ids.und) continue;
        const videoID = await renameEmbedded(lesson, 'en');
        for (const [language, label, file, count] of [
          ['Німецька', 'Deutsch', 'e2e-tone-30s.m4a', 2],
          ['Французька', 'Français', 'e2e-tone-1320-30s.m4a', 3],
        ] as const) {
          const section = group.getByRole('region', { name: 'Аудіодоріжки' });
          await section.getByRole('button', { name: /Мова доріжки/ }).click();
          await page.getByRole('option', { name: language, exact: true }).click();
          await section.getByRole('textbox', { name: 'Назва доріжки' }).fill(label);
          const mintedAudio = page.waitForResponse(
            (r) => r.url().endsWith('/audio-tracks') && r.request().method() === 'POST',
          );
          await section
            .locator('input[type="file"]')
            .setInputFiles(path.join(fixtureDirectory, file));
          expect((await mintedAudio).status()).toBe(201);
          await expect(section.locator('li[data-audio-track-state="ready"]')).toHaveCount(count, {
            timeout: 240_000,
          });
        }
        const tracks = await admin.get(`/api/v1/admin/videos/${videoID}/audio-tracks`);
        expect(tracks.status()).toBe(200);
        const items = tracksSchema.parse(await tracks.json()).items;
        expect(items.map((t) => t.language).sort()).toEqual(['de', 'en', 'fr']);
        if (lesson === ids.a) deID = items.find((t) => t.language === 'de')?.trackId ?? '';
        await putTranscript(lesson, 'en');
        await putTranscript(lesson, 'de');
      }
      const current = await admin.get(`/api/v1/admin/courses/${courseID}`);
      const { updatedAt } = tokenSchema.parse(await current.json());
      const published = await admin.post(`/api/v1/admin/courses/${courseID}/status`, {
        headers,
        data: { expectedUpdatedAt: updatedAt, status: 'published' },
      });
      expect(published.status(), await published.text()).toBe(200);
      const grant = await admin.post(`/api/v1/admin/courses/${courseID}/grants`, {
        headers,
        data: { userId: '40000000-0000-4000-8000-000000000001', reason: 'MS4 e2e student' },
      });
      expect(grant.status(), await grant.text()).toBe(201);
    } finally {
      await page.close();
    }
  });
  test.afterAll(async () => {
    if (!courseID || !admin) return;
    const current = await admin.get(`/api/v1/admin/courses/${courseID}`);
    const { updatedAt } = tokenSchema.parse(await current.json());
    const archived = await admin.post(`/api/v1/admin/courses/${courseID}/status`, {
      headers,
      data: { expectedUpdatedAt: updatedAt, status: 'archived' },
    });
    expect(archived.status(), await archived.text()).toBe(200);
    await admin.dispose();
  });
  test.beforeEach(async ({ page }) => {
    await resetProgress();
    const requests: object[] = [];
    page.on('response', (response) => {
      const url = new URL(response.url());
      if (url.pathname.startsWith('/media/') || url.pathname.includes('/transcript'))
        requests.push({
          pathname: url.pathname,
          language: url.searchParams.get('language'),
          status: response.status(),
        });
    });
    requestLogs.set(page, requests);
  });
  test.afterEach(async ({ page }, info) => {
    const requests = requestLogs.get(page) ?? [];
    await writeFile(info.outputPath('requests.json'), JSON.stringify(requests, null, 2));
    if (info.status !== info.expectedStatus) {
      await page.screenshot({ path: info.outputPath('failure.png'), fullPage: true });
      const evidence = await page.evaluate(() => ({
        entries: window.ms4?.entries,
        states: window.ms4?.states.slice(-100),
      }));
      await writeFile(info.outputPath('probe.json'), JSON.stringify(evidence, null, 2));
    }
  });
  test('E1 prediction before Play becomes confirmed en captions with 440 Hz', async ({ page }) => {
    await open(page);
    await expect(panel(page).getByText(english, { exact: true })).toBeVisible();
    await expect(page.locator('video track')).toHaveCount(0);
    await start(page);
    expect(isTone(await peak(page), 440)).toBe(true);
  });
  test('E2 held de transcript clears showing cues immediately and resets search and follow', async ({
    page,
  }) => {
    await open(page);
    await start(page);
    await panel(page).locator('ol').dispatchEvent('wheel');
    await expect(
      panel(page).getByRole('button', { name: 'Стежити за відтворенням' }),
    ).toBeVisible();
    await panel(page).getByRole('searchbox').fill('English');
    const held = gate();
    let requested = false;
    await page.route(
      (url) => transcriptURL(url, 'de'),
      async (route) => {
        requested = true;
        await held.promise;
        await route.continue();
      },
    );
    try {
      await choose(page, 'Deutsch');
      await expect.poll(() => requested).toBe(true);
      await expect(panel(page).getByRole('status')).toHaveText('Завантаження');
      for (const sample of await samples(page, 1200)) expect(sample.caption.active).toEqual([]);
    } finally {
      held.release();
    }
    await waitForTone(page, 880, 3, 60_000);
    await showing(page, 'de');
    await expect(panel(page).getByText(german, { exact: true })).toBeVisible();
    await expect(panel(page).getByRole('searchbox')).toHaveValue('');
    await expect(panel(page).getByRole('button', { name: 'Стежити за відтворенням' })).toHaveCount(
      0,
    );
  });
  test('E3 a fresh session remembering de never requests or renders en', async ({ page }) => {
    const languages: string[] = [];
    page.on('request', (request) => {
      const url = new URL(request.url());
      if (url.pathname.endsWith('/transcript'))
        languages.push(url.searchParams.get('language') ?? 'primary');
    });
    await probe(page);
    await open(page, ids.a, 'de');
    await expect(panel(page).getByText(german, { exact: true })).toBeVisible();
    await expect(page.locator('video track')).toHaveCount(0);
    await start(page, 'de');
    expect(languages.length).toBeGreaterThan(0);
    expect(languages).not.toContain('en');
    expect(await page.evaluate(() => window.ms4.forbiddenEnglish)).toBe(false);
    expect(isTone(await peak(page), 880)).toBe(true);
  });
  test('E4a cancellation before buffering never admits de sound or cues', async ({ page }) => {
    await open(page);
    await start(page);
    const held = gate();
    let requested = 0;
    await page.route(new RegExp(`/audio-${deID}/[^/]+\\.m4s(?:\\?.*)?$`), async (route) => {
      requested++;
      await held.promise;
      await route.abort();
    });
    try {
      await choose(page, 'Deutsch');
      await expect.poll(() => requested, { timeout: 60_000 }).toBeGreaterThan(0);
      await choose(page, 'English');
      const observed = await samples(page, 3600);
      expect(observed.length).toBeGreaterThanOrEqual(3);
      for (const sample of observed) {
        expect(isTone(sample, 440), JSON.stringify(sample)).toBe(true);
        expect(sample.caption.language).not.toBe('de');
        if (sample.caption.active.length) expect(sample.caption.language).toBe('en');
      }
      await showing(page, 'en');
    } finally {
      held.release();
    }
  });
  test('E4b cancellation after de buffers ahead keeps cues aligned with heard intervals', async ({
    page,
  }) => {
    await probe(page);
    await deBarrier(page);
    await open(page);
    await start(page);
    await choose(page, 'Deutsch');
    // Require a successful de media append ahead of the playhead, not just an HTTP response.
    await expect
      .poll(
        () =>
          page.evaluate(() => {
            const time = document.querySelector('video')?.currentTime ?? Infinity;
            return window.ms4.entries.some(
              (entry) =>
                entry.kind === 'media' &&
                entry.outcome === 'ok' &&
                entry.mdatFnv &&
                window.ms4.de.has(entry.mdatFnv) &&
                entry.tfdt / entry.timescale + entry.timestampOffset > time + 1,
            );
          }),
        { timeout: 60_000 },
      )
      .toBe(true);
    const bufferedDe = await page.evaluate(() =>
      window.ms4.entries
        .filter(
          (entry) =>
            entry.kind === 'media' &&
            entry.outcome === 'ok' &&
            entry.mdatFnv &&
            window.ms4.de.has(entry.mdatFnv),
        )
        .map((entry) => ({
          start: entry.tfdt / entry.timescale + entry.timestampOffset,
          end: entry.tfdt / entry.timescale + entry.timestampOffset + entry.duration,
        })),
    );
    // Owner decision, 2026-11-18 (plan-ms4.md E4b): prevent the new en fetches from
    // overwriting buffered de before it is actually heard under the en selection.
    const heldEn = gate();
    let heldEnRequests = 0;
    let releasedAt: { time: number; hz: number } | null = null;
    await page.route(/\/audio\/[^/]+\.m4s(?:\?.*)?$/, async (route) => {
      heldEnRequests++;
      await heldEn.promise;
      await route.continue();
    });
    const observed = [];
    try {
      await choose(page, 'English');
      // Passive timing only: the existing 300 ms decoded-audio assertion samples are unchanged.
      await page.evaluate(() => {
        const times: E4bTimings = {
          releasedAt: null,
          releaseMediaTime: null,
          firstEnCueAt: null,
          firstEnCueMediaTime: null,
          pending: [],
        };
        window.ms4E4bTimes = times;
        const sample = () => {
          const time = document.querySelector('video')?.currentTime ?? -1;
          const at = performance.now();
          const pending =
            document
              .querySelector('section.lesson-player')
              ?.getAttribute('data-caption-selection') === 'pending';
          const last = times.pending.at(-1);
          if (pending && (!last || last.end !== null))
            times.pending.push({ start: time, end: null, startAt: at, endAt: null });
          else if (!pending && last?.end === null) {
            last.end = time;
            last.endAt = at;
          }
          const track = document.querySelector('video track');
          if (
            times.releasedAt !== null &&
            times.firstEnCueAt === null &&
            track instanceof HTMLTrackElement &&
            track.srclang === 'en' &&
            track.track.mode === 'showing' &&
            track.track.activeCues?.length
          ) {
            times.firstEnCueAt = at;
            times.firstEnCueMediaTime = time;
          }
        };
        new MutationObserver(sample).observe(document, {
          subtree: true,
          childList: true,
          attributes: true,
        });
        setInterval(sample, 20);
        sample();
      });
      const until = Date.now() + 15_000;
      while (Date.now() < until) {
        const sample = {
          at: Date.now(),
          ...(await peak(page)),
          caption: await currentCaption(page),
          selection: await player(page).getAttribute('data-caption-selection'),
        };
        observed.push(sample);
        if (
          releasedAt === null &&
          isTone(sample, 880) &&
          bufferedDe.some((interval) => interval.start <= sample.time && sample.time < interval.end)
        ) {
          releasedAt = { time: sample.time, hz: sample.hz };
          await page.evaluate(() => {
            window.ms4E4bTimes.releasedAt = performance.now();
            window.ms4E4bTimes.releaseMediaTime =
              document.querySelector('video')?.currentTime ?? -1;
          });
          heldEn.release();
        }
        await page.waitForTimeout(300);
      }
    } finally {
      heldEn.release();
    }
    expect(heldEnRequests, 'new en fragments were actually held').toBeGreaterThan(0);
    expect(releasedAt, 'release requires an audible in-interval de sample').not.toBeNull();
    const intervals: { hz: number; start: number; end: number; samples: number }[] = [];
    let mismatches = 0;
    let deSamples = 0;
    for (const sample of observed) {
      const hz = isTone(sample, 440) ? 440 : isTone(sample, 880) ? 880 : 0;
      const last = intervals.at(-1);
      if (last?.hz === hz) {
        last.end = sample.time;
        last.samples++;
      } else intervals.push({ hz, start: sample.time, end: sample.time, samples: 1 });
      if (hz === 880) {
        deSamples++;
      }
      if (sample.caption.active.length && (hz !== 440 || sample.caption.language !== 'en'))
        mismatches++;
    }
    console.log(
      `E4b intervals: ${JSON.stringify(intervals)}; mismatches=${mismatches}; deSamples=${deSamples}; heldEnRequests=${heldEnRequests}; releasedAt=${JSON.stringify(releasedAt)}; bufferedDe=${JSON.stringify(bufferedDe)}`,
    );
    expect(deSamples, 'buffered de must actually be heard after cancellation').toBeGreaterThan(0);
    const timings = await page.evaluate(() => window.ms4E4bTimes);
    const firstEnCueDelayMs =
      timings.firstEnCueAt === null || timings.releasedAt === null
        ? null
        : timings.firstEnCueAt - timings.releasedAt;
    await writeFile(
      test.info().outputPath('e4b-samples.json'),
      JSON.stringify({ observed, timings }, null, 2),
    );
    console.log(
      `E4b timing: pending=${JSON.stringify(timings.pending)}; releaseMediaTime=${timings.releaseMediaTime}; firstEnCueMediaTime=${timings.firstEnCueMediaTime}; first en cue after release=${firstEnCueDelayMs} ms; passive observer 20 ms`,
    );
    expect(mismatches, 'at most one transition sample').toBeLessThanOrEqual(1);
    expect(firstEnCueDelayMs, 'en showing cue returns after release').not.toBeNull();
  });
  test('E5 fr has no captions and returning en retains captionsOn', async ({ page }) => {
    await open(page);
    await start(page);
    await choose(page, 'Français');
    await waitForTone(page, 1320, 3, 60_000);
    await expect(
      panel(page).getByText('Немає субтитрів для цієї доріжки', { exact: true }),
    ).toBeVisible();
    await noShowing(page);
    await page.getByRole('button', { name: 'Налаштування' }).click();
    await expect(page.getByText('Субтитри: немає для цієї доріжки', { exact: true })).toBeVisible();
    await expect(page.getByRole('switch', { name: 'Субтитри', exact: true })).toHaveCount(0);
    await page.keyboard.press('Escape');
    await choose(page, 'English');
    await waitForTone(page, 440, 3, 60_000);
    await showing(page, 'en');
  });
  test('E6 untouched embedded und has no transcript request or captions', async ({ page }) => {
    const requested: string[] = [];
    page.on('request', (r) => {
      const url = new URL(r.url());
      if (url.pathname.endsWith('/transcript'))
        requested.push(url.searchParams.get('language') ?? 'primary');
    });
    await open(page, ids.und);
    await expect(
      panel(page).getByText('Немає субтитрів для цієї доріжки', { exact: true }),
    ).toBeVisible();
    await page.getByRole('button', { name: 'Відтворити', exact: true }).click();
    // Owner's missing-LANGUAGE fix confirms und once audio is buffered. Wait for that
    // actual playback state before captureStream, which has no audio track before metadata.
    await expect(player(page)).toHaveAttribute('data-caption-selection', 'confirmed:und', {
      timeout: 60_000,
    });
    await tapAudio(page);
    await waitForTone(page, 440, 3, 60_000);
    await noShowing(page);
    await expect(
      panel(page).getByText('Немає субтитрів для цієї доріжки', { exact: true }),
    ).toBeVisible();
    expect(requested).toEqual([]);
  });
  test('E7 measures click to confirmed de and first showing de cue', async ({ page }) => {
    await open(page);
    await start(page);
    await page.evaluate(() => {
      const times = { clicked: 0, confirmed: 0, showing: 0 };
      Object.assign(window, { ms4Times: times });
      document.addEventListener(
        'click',
        (event) => {
          if (
            event.target instanceof Element &&
            event.target.closest('label')?.textContent?.includes('Deutsch')
          )
            times.clicked = performance.now();
        },
        true,
      );
      const sample = () => {
        if (!times.clicked) return;
        if (
          !times.confirmed &&
          document
            .querySelector('section.lesson-player')
            ?.getAttribute('data-caption-selection') === 'confirmed:de'
        )
          times.confirmed = performance.now();
        const track = document.querySelector('video track');
        if (
          !times.showing &&
          track instanceof HTMLTrackElement &&
          track.srclang === 'de' &&
          track.track.mode === 'showing' &&
          track.track.activeCues?.length
        )
          times.showing = performance.now();
      };
      new MutationObserver(sample).observe(document, {
        subtree: true,
        childList: true,
        attributes: true,
      });
      setInterval(sample, 20);
    });
    await choose(page, 'Deutsch');
    await waitForTone(page, 880, 3, 60_000);
    await showing(page, 'de');
    await expect
      .poll(() => page.evaluate(() => window.ms4Times.showing), { timeout: 60_000 })
      .toBeGreaterThan(0);
    const times = await page.evaluate(() => window.ms4Times);
    expect(times.clicked).toBeGreaterThan(0);
    expect(times.confirmed).toBeGreaterThanOrEqual(times.clicked);
    expect(times.showing).toBeGreaterThanOrEqual(times.clicked);
    console.log(
      `E7 measurement: click -> confirmed:de ${(times.confirmed - times.clicked).toFixed(1)} ms; click -> first showing de cue ${(times.showing - times.clicked).toFixed(1)} ms; Chromium/Kiri, 300 ms audio samples`,
    );
  });
  test('E8 one embedded en track confirms on initial Play without switching', async ({ page }) => {
    await renameEmbedded(ids.und, 'en');
    await putTranscript(ids.und, 'en');
    await open(page, ids.und);
    await start(page);
    await page.getByRole('button', { name: 'Налаштування' }).click();
    await expect(page.getByRole('radiogroup', { name: 'Аудіо', exact: true })).toHaveCount(0);
    await page.keyboard.press('Escape');
    expect(isTone(await peak(page), 440)).toBe(true);
  });
  test('E9 transcript 503 is retryable while de audio keeps playing', async ({ page }) => {
    await open(page);
    await start(page);
    const failing = async (route: Route) =>
      route.fulfill({
        status: 503,
        contentType: 'application/json',
        body: JSON.stringify({
          error: { code: 'unavailable', message: 'Injected transcript error' },
        }),
      });
    await page.route((url) => transcriptURL(url, 'de'), failing);
    await choose(page, 'Deutsch');
    await waitForTone(page, 880, 3, 60_000);
    await expect(
      panel(page).getByText('Не вдалося завантажити транскрипт', { exact: true }),
    ).toBeVisible({ timeout: 60_000 });
    await noShowing(page);
    const before = (await peak(page)).time;
    await expect
      .poll(async () => (await peak(page)).time, { timeout: 15_000 })
      .toBeGreaterThan(before);
    await page.unrouteAll({ behavior: 'wait' });
    await panel(page).getByRole('button', { name: 'Спробувати ще', exact: true }).click();
    await showing(page, 'de');
    expect(isTone(await peak(page), 880)).toBe(true);
  });
  test('E10 SPA navigation from confirmed A to B remembers de with no en request or text', async ({
    page,
  }) => {
    await probe(page);
    await open(page);
    await start(page);
    await choose(page, 'Deutsch');
    await waitForTone(page, 880, 3, 60_000);
    await showing(page, 'de');
    const languages: string[] = [];
    page.on('request', (r) => {
      const url = new URL(r.url());
      if (url.pathname === `/api/v1/learn/lessons/${ids.b}/transcript`)
        languages.push(url.searchParams.get('language') ?? 'primary');
    });
    await page.evaluate(() => {
      window.ms4.forbiddenEnglish = false;
    });
    await page.getByRole('link', { name: /MS4 B/ }).click();
    await expect(panel(page).getByText(german, { exact: true })).toBeVisible();
    await start(page, 'de');
    expect(languages.length).toBeGreaterThan(0);
    expect(languages).not.toContain('en');
    expect(await page.evaluate(() => window.ms4.forbiddenEnglish)).toBe(false);
  });
  test('E11 A to unplayed B to A discards the old confirmed mount', async ({ page }) => {
    await open(page);
    await start(page);
    await page.getByRole('link', { name: /MS4 B/ }).click();
    await expect(panel(page).getByText(english, { exact: true })).toBeVisible();
    await expect(page.locator('video track')).toHaveCount(0);
    await page.getByRole('link', { name: /MS4 A/ }).click();
    await expect(panel(page).getByText(english, { exact: true })).toBeVisible();
    await expect(page.locator('video track')).toHaveCount(0);
    await expect(page.getByRole('button', { name: 'Відтворити', exact: true })).toBeVisible();
    expect(
      await page
        .locator('video')
        .evaluate((node) => (node instanceof HTMLVideoElement ? node.currentTime : -1)),
    ).toBe(0);
  });
  test('E12 fourth-main-fragment 503 reloads the same instance without prediction', async ({
    page,
  }) => {
    await retryMain(page);
  });
  test('E13 a fingerprinted de media append failure stays unknown until the same payload recovers', async ({
    page,
  }) => {
    await probe(page);
    const network = await deBarrier(page);
    await open(page);
    await start(page);
    await page.evaluate(() => {
      window.ms4.inject = true;
    });
    await choose(page, 'Deutsch');
    await expect
      .poll(() => page.evaluate(() => window.ms4.held), { timeout: 60_000 })
      .not.toBeNull();
    const failure = await page.evaluate(() =>
      window.ms4.entries.find((entry) => entry.outcome === 'threw'),
    );
    if (!failure) throw new Error('No injected media append failure');
    expect(failure.kind).toBe('media');
    expect(failure.name).toBe('InvalidStateError');
    expect(network.some((entry) => entry.fingerprint === failure.mdatFnv)).toBe(true);
    expect(failure.timescale).toBeGreaterThan(0);
    expect(failure.duration).toBeGreaterThan(0);
    const intervalStart = failure.tfdt / failure.timescale + failure.timestampOffset;
    const end = intervalStart + failure.duration;
    const observed = [];
    const deadline = Date.now() + 30_000;
    while (Date.now() < deadline && observed.length < 3) {
      const sample = {
        ...(await peak(page)),
        caption: await currentCaption(page),
        selection: await player(page).getAttribute('data-caption-selection'),
      };
      if (intervalStart - 0.25 <= sample.time && sample.time < end) {
        observed.push(sample);
        expect(sample.caption.active, JSON.stringify(sample)).toEqual([]);
        expect(sample.selection).not.toBe('confirmed:de');
      }
      if (sample.time >= end) break;
      await page.waitForTimeout(300);
    }
    expect(
      observed.length,
      `failure interval [${intervalStart},${end}) ${JSON.stringify(observed)}`,
    ).toBeGreaterThanOrEqual(3);
    await expect(player(page)).toBeVisible();
    await expect(
      page.getByRole('button', { name: 'Почати перегляд знову', exact: true }),
    ).toHaveCount(0);
    await page.evaluate(() => {
      const release = window.ms4.release;
      if (!release) throw new Error('No deferred repeat');
      window.ms4.release = null;
      release();
    });
    await expect
      .poll(
        () =>
          page.evaluate(
            () => window.ms4.entries.find((entry) => entry.seq === window.ms4.held)?.outcome,
          ),
        { timeout: 60_000 },
      )
      .toBe('ok');
    const recovery = await page.evaluate(() =>
      window.ms4.entries.find((entry) => entry.seq === window.ms4.held),
    );
    if (!recovery) throw new Error('Missing recovery entry');
    expect(recovery.mdatFnv).toBe(failure.mdatFnv);
    const recoveryStart = recovery.tfdt / recovery.timescale + recovery.timestampOffset;
    expect(recoveryStart).toBeLessThan(end);
    expect(recoveryStart + recovery.duration).toBeGreaterThan(intervalStart);
    const recoveredSamples = [];
    const recoveredDeadline = Date.now() + 30_000;
    while (Date.now() < recoveredDeadline) {
      const sample = await peak(page);
      recoveredSamples.push(sample);
      if (sample.time >= intervalStart && sample.time < end && isTone(sample, 880)) break;
      if (sample.time >= end) break;
      await page.waitForTimeout(300);
    }
    expect(
      recoveredSamples.some(
        (sample) => sample.time >= intervalStart && sample.time < end && isTone(sample, 880),
      ),
      JSON.stringify(recoveredSamples),
    ).toBe(true);
    await showing(page, 'de');
    const entries = await page.evaluate(() => window.ms4.entries);
    expect(entries.filter((entry) => entry.outcome === 'threw')).toHaveLength(1);
    expect(entries.filter((entry) => entry.outcome === 'error')).toHaveLength(0);
    console.log(
      `E13 recovery: ${JSON.stringify({ failure, recovery, observed, recoveredSamples, network })}`,
    );
  });
  test('E14 generation-two A to first-load B accepts B confirmation', async ({ page }) => {
    await retryMain(page);
    let firstLoadMasters = 0;
    page.on('request', (request) => {
      if (new URL(request.url()).pathname.endsWith('/master.m3u8')) firstLoadMasters++;
    });
    await page.getByRole('link', { name: /MS4 B/ }).click();
    await expect(panel(page).getByText(english, { exact: true })).toBeVisible();
    await expect(page.locator('video track')).toHaveCount(0);
    await start(page);
    expect(isTone(await peak(page), 440)).toBe(true);
    expect(firstLoadMasters, 'B has exactly one first load').toBe(1);
  });
});

async function retryMain(page: Page) {
  await probe(page);
  let mainRequests = 0;
  let masters = 0;
  let failedOnce = false;
  const fourth = gate();
  const fail = gate();
  const requests: string[] = [];
  page.on('request', (request) => {
    if (new URL(request.url()).pathname.endsWith('/master.m3u8')) masters++;
  });
  await page.route(/\/media\/p\/.*\.m4s(?:\?.*)?$/, async (route) => {
    const pathname = new URL(route.request().url()).pathname;
    if (!/\/audio(?:-[0-9a-f-]{36})?\//.test(pathname)) {
      requests.push(pathname);
      mainRequests++;
      if (mainRequests === 4 && !failedOnce) {
        failedOnce = true;
        fourth.release();
        await fail.promise;
        await route.fulfill({ status: 503, body: 'injected fourth main segment failure' });
        return;
      }
    }
    await route.continue();
  });
  await open(page);
  try {
    await start(page);
    await expect.poll(() => failedOnce, { timeout: 60_000 }).toBe(true);
    await fourth.promise;
    await page.evaluate(() => {
      window.ms4.states.length = 0;
    });
    fail.release();
    await expect.poll(() => masters, { timeout: 60_000 }).toBe(2);
    await expect(player(page)).toHaveAttribute('data-caption-selection', 'confirmed:en', {
      timeout: 60_000,
    });
    await waitForTone(page, 440, 3, 60_000);
    await showing(page, 'en');
    const states = await page.evaluate(() => window.ms4.states);
    expect(states.some((state) => state.selection === 'pending')).toBe(true);
    for (const state of states.filter((state) => state.selection === 'pending')) {
      expect(state.track).toBeNull();
      expect(state.text).not.toContain(english);
    }
    expect(masters).toBe(2);
    console.log(
      `E12 retry: ${JSON.stringify({ masters, requests, pending: states.filter((state) => state.selection === 'pending').length })}`,
    );
  } finally {
    fail.release();
  }
}
