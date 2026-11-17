import { expect, rstest, test } from '@rstest/core';
import { createMemoryTranscriptHistory } from '../adapters/transcript-history-memory';
import {
  captureDirective,
  clearAcceptedTranscriptSnapshot,
  detachedTranscriptSnapshots,
  forgetOtherAuthors,
  foreignSnapshotKeys,
  recoveryOffer,
  transcriptHistoryKey,
  type TranscriptDraftLine,
  type TranscriptRecoveryTab,
  type TranscriptSnapshot,
} from './recovery';

const line = (overrides: Partial<TranscriptDraftLine> = {}): TranscriptDraftLine => ({
  id: '40000000-0000-4000-8000-000000000001',
  startMs: '0',
  endMs: '1000',
  text: 'перший рядок',
  documentIds: [],
  ...overrides,
});

const snapshot = (overrides: Partial<TranscriptSnapshot> = {}): TranscriptSnapshot => ({
  key: 'user:lesson',
  revision: 'snapshot-revision',
  capturedAt: 1_700_000_000_000,
  language: 'uk',
  expectedUpdatedAt: '2026-09-18T10:00:00Z',
  lines: [line()],
  ...overrides,
});

test('nothing stored, and a dismissed offer, both stay silent', () => {
  const baseline = { expectedUpdatedAt: '2026-09-18T10:00:00Z', lines: [line()] };
  expect(recoveryOffer(null, baseline, false)).toEqual({ kind: 'none' });
  expect(recoveryOffer(snapshot({ lines: [line({ text: 'інше' })] }), baseline, true)).toEqual({
    kind: 'none',
  });
});

test('a snapshot equal to the server is not a recovery', () => {
  // The offer keys on a DIFFERENCE, not on a snapshot existing. Prompting when there is nothing
  // to restore teaches the author to dismiss without reading, which disarms the real prompt.
  const baseline = { expectedUpdatedAt: '2026-09-18T10:00:00Z', lines: [line()] };
  expect(recoveryOffer(snapshot(), baseline, false)).toEqual({ kind: 'none' });
});

test('a half-typed time is a difference, and survives as the text it was', () => {
  // This is why the snapshot holds the editor's own lines and not parsed segments: "15" mid-typing
  // is not a number yet, and a snapshot of parsed segments could not carry it back.
  const baseline = { expectedUpdatedAt: '2026-09-18T10:00:00Z', lines: [line()] };
  const halfTyped = snapshot({ lines: [line({ startMs: '1' })] });
  const offered = recoveryOffer(halfTyped, baseline, false);
  expect(offered.kind).toBe('offer');
  expect(offered.kind === 'offer' ? offered.snapshot.lines[0]?.startMs : null).toBe('1');
});

test('every field of a line counts as a difference, including an emptied document list', () => {
  const baseline = {
    expectedUpdatedAt: '2026-09-18T10:00:00Z',
    lines: [line({ documentIds: ['doc-1'] })],
  };
  for (const changed of [
    line({ documentIds: [] }),
    line({ documentIds: ['doc-2'] }),
    line({ text: 'виправлено', documentIds: ['doc-1'] }),
    line({ startMs: '5', documentIds: ['doc-1'] }),
    line({ endMs: '5', documentIds: ['doc-1'] }),
    line({ id: '40000000-0000-4000-8000-000000000009', documentIds: ['doc-1'] }),
  ]) {
    expect(recoveryOffer(snapshot({ lines: [changed] }), baseline, false).kind).toBe('offer');
  }
  // A line removed and a line added are differences too, and neither is a length coincidence.
  expect(recoveryOffer(snapshot({ lines: [] }), baseline, false).kind).toBe('offer');
});

test('a differing snapshot on the same server version is a plain offer', () => {
  const baseline = { expectedUpdatedAt: '2026-09-18T10:00:00Z', lines: [] };
  expect(recoveryOffer(snapshot(), baseline, false)).toEqual({
    kind: 'offer',
    snapshot: snapshot(),
  });
});

test('a server that moved under the snapshot is a different offer, not the same one', () => {
  // This is the case Phase 8 creates on purpose: recognition writes a draft while the author is
  // away, so the base a snapshot was taken against is no longer the base it would be restored
  // onto. Restoring stays the author's choice, but it overwrites newer text and may not be
  // dressed as an ordinary restore.
  const baseline = { expectedUpdatedAt: '2026-09-18T12:00:00Z', lines: [] };
  expect(recoveryOffer(snapshot(), baseline, false).kind).toBe('offer-stale');
});

test('a draft snapshot stays ordinary when the current draft token is unchanged', () => {
  const stored = snapshot({ draftUpdatedAt: '2026-09-24T10:00:00Z' });
  expect(
    recoveryOffer(
      stored,
      {
        expectedUpdatedAt: stored.expectedUpdatedAt,
        draftUpdatedAt: '2026-09-24T10:00:00Z',
        lines: [],
      },
      false,
    ).kind,
  ).toBe('offer');
});

test('a regenerated draft makes its local snapshot stale', () => {
  const stored = snapshot({ draftUpdatedAt: '2026-09-24T10:00:00Z' });
  expect(
    recoveryOffer(
      stored,
      {
        expectedUpdatedAt: stored.expectedUpdatedAt,
        draftUpdatedAt: '2026-09-24T11:00:00Z',
        lines: [],
      },
      false,
    ).kind,
  ).toBe('offer-stale');
});

test('a manual snapshot without a draft token keeps the old comparison', () => {
  const stored = snapshot();
  expect(
    recoveryOffer(
      stored,
      {
        expectedUpdatedAt: stored.expectedUpdatedAt,
        draftUpdatedAt: '2026-09-24T11:00:00Z',
        lines: [],
      },
      false,
    ).kind,
  ).toBe('offer');
});

test('an absent token is a value, not a missing one', () => {
  // `undefined` means "the author began typing into a lesson with no transcript". A transcript
  // appearing since then is exactly a moved base, so it must be stale rather than plain.
  const none = snapshot({ expectedUpdatedAt: undefined });
  expect(recoveryOffer(none, { expectedUpdatedAt: undefined, lines: [] }, false).kind).toBe(
    'offer',
  );
  expect(
    recoveryOffer(none, { expectedUpdatedAt: '2026-09-18T12:00:00Z', lines: [] }, false).kind,
  ).toBe('offer-stale');
});

test('editing back to the server text clears the snapshot rather than skipping the write', () => {
  // Skipping would leave the previous snapshot in place, and it would be offered on the next
  // visit as work the author had already undone by hand.
  const baseline = { expectedUpdatedAt: '2026-09-18T10:00:00Z', lines: [line()] };
  expect(captureDirective([line()], baseline)).toEqual({ kind: 'clear' });
  expect(captureDirective([line({ text: 'змінено' })], baseline)).toEqual({ kind: 'write' });
  expect(captureDirective([], baseline)).toEqual({ kind: 'write' });
});

test('the key separates authors and lessons, not editor languages', () => {
  const parts = { userId: 'user-1', lessonId: 'lesson-1' };
  expect(transcriptHistoryKey(parts)).toBe('user-1:lesson-1');
  // A snapshot outlives the session on disk; without the author in the key a shared machine
  // offers the next person a course they do not own.
  expect(transcriptHistoryKey({ ...parts, userId: 'user-2' })).not.toBe(
    transcriptHistoryKey(parts),
  );
  // Changing the editor language must keep the same recovery bucket; the snapshot carries it.
  expect(transcriptHistoryKey(parts)).toBe(transcriptHistoryKey(parts));
});

test('a change of viewer sweeps what is not theirs, and signing out sweeps everything', async () => {
  // Keying by author stops one person being OFFERED another's text. It does not remove the text,
  // and on a shared machine "not offered" is not the same as "not there".
  const keys = ['author-1:lesson-1', 'author-1:lesson-2', 'author-2:lesson-1'];
  expect(foreignSnapshotKeys(keys, 'author-1')).toEqual(['author-2:lesson-1']);
  expect(foreignSnapshotKeys(keys, 'author-2')).toEqual(['author-1:lesson-1', 'author-1:lesson-2']);
  // Nobody is viewing: everything goes. Signing out is a change of viewer like any other.
  expect(foreignSnapshotKeys(keys, null)).toEqual(keys);
  // A prefix that merely starts the same is a different author, not the same one.
  expect(foreignSnapshotKeys(['author-10:lesson-1'], 'author-1')).toEqual(['author-10:lesson-1']);
});

test('the sweep removes exactly the foreign keys and survives a store that refuses', async () => {
  const store = createMemoryTranscriptHistory();
  const snapshotFor = (key: string) => ({
    key,
    revision: crypto.randomUUID(),
    capturedAt: 1,
    expectedUpdatedAt: undefined,
    lines: [line()],
  });
  await store.write(snapshotFor('author-1:lesson-1'));
  await store.write(snapshotFor('author-2:lesson-1'));

  await forgetOtherAuthors(store, 'author-1');
  expect(await store.keys()).toEqual(['author-1:lesson-1']);

  await forgetOtherAuthors(store, null);
  expect(await store.keys()).toEqual([]);

  // Best-effort: a refusal costs a stale snapshot on disk and must never surface as a failure of
  // whatever page the person is actually on.
  const refusing = {
    read: () => Promise.reject(new Error('blocked')),
    write: () => Promise.reject(new Error('blocked')),
    clear: () => Promise.reject(new Error('blocked')),
    clearIfVersion: () => Promise.reject(new Error('blocked')),
    keys: () => Promise.reject(new Error('blocked')),
  };
  await expect(forgetOtherAuthors(refusing, 'author-1')).resolves.toBeUndefined();
});

const tabs: TranscriptRecoveryTab[] = [
  { kind: 'track', instanceId: 'track:en', language: 'en', isDefault: true },
  { kind: 'track', instanceId: 'track:uk', language: 'uk', isDefault: false },
];
function place(snapshots: readonly TranscriptSnapshot[], currentTabs = tabs) {
  return detachedTranscriptSnapshots(
    'user:lesson',
    snapshots.map((s) => s.key),
    snapshots,
    currentTabs,
  );
}

test('M3-R6 legacy language matches case-insensitively; unknown and unmatched go to default', () => {
  expect(place([snapshot({ language: 'UK' })])[0]).toMatchObject({
    instanceId: 'track:uk',
    legacy: true,
  });
  const { language: _language, ...unknown } = snapshot();
  expect(place([unknown])[0]?.instanceId).toBe('track:en');
  expect(place([snapshot({ language: 'fr' })])[0]?.instanceId).toBe('track:en');
});

test('M3-R6 gone tracks go to their language/default; own keys and another lesson prefix are excluded', () => {
  const snapshots = [
    snapshot({ key: 'user:lesson:track:gone', language: 'uk' }),
    snapshot({ key: 'user:lesson:track:gone2', language: 'fr' }),
    snapshot({ key: 'user:lesson:track:uk' }),
    snapshot({ key: 'user:lesson-other:track:gone' }),
    snapshot({ key: 'other:lesson' }),
  ];
  expect(place(snapshots).map((s) => [s.snapshot.key, s.instanceId])).toEqual([
    ['user:lesson:track:gone', 'track:uk'],
    ['user:lesson:track:gone2', 'track:en'],
  ]);
});

test('M3-R6 multiple detached snapshots are newest first and only keys actually listed are used', () => {
  const old = snapshot({ revision: crypto.randomUUID(), capturedAt: 1 });
  const newest = snapshot({
    key: 'user:lesson:track:gone',
    revision: crypto.randomUUID(),
    capturedAt: 3,
  });
  expect(place([old, newest]).map((s) => s.snapshot.capturedAt)).toEqual([3, 1]);
  expect(detachedTranscriptSnapshots('user:lesson', [], [old], tabs)).toEqual([]);
});

test('M3-R6 single editor owns the legacy key and receives detached track/lang snapshots', () => {
  const detached = snapshot({ key: 'user:lesson:track:gone' });
  const orphan = snapshot({ key: 'user:lesson:lang:uk' });
  expect(place([snapshot(), detached, orphan], []).map((s) => [s.instanceId, s.legacy])).toEqual([
    ['', false],
    ['', false],
  ]);
});

test('M3-R12 c orphan attached by track rename is detached and placed in that track', () => {
  const orphan = snapshot({ key: 'user:lesson:lang:uk' });
  expect(
    place(
      [orphan],
      [...tabs, { kind: 'orphan', instanceId: 'lang:uk', language: 'uk', isDefault: false }],
    ),
  ).toEqual([]);
  expect(place([orphan])[0]?.instanceId).toBe('track:uk');
});

test('M3-R6 viewer sweep keeps all own tab key forms and removes foreign ones', async () => {
  const store = createMemoryTranscriptHistory();
  const keys = [
    'user:lesson',
    'user:lesson:track:one',
    'user:lesson:lang:uk',
    'other:lesson:track:one',
  ];
  for (const key of keys) await store.write(snapshot({ key }));
  await forgetOtherAuthors(store, 'user');
  expect(await store.keys()).toEqual(keys.slice(0, 3));
});

for (const version of ['same', 'newer', 'same-time-content', 'read-fails', 'owned'] as const) {
  test(`MS3b fix3 compare-delete ${version}`, async () => {
    const store = createMemoryTranscriptHistory();
    const accepted = snapshot();
    const stored =
      version === 'newer'
        ? { ...accepted, revision: crypto.randomUUID(), capturedAt: accepted.capturedAt }
        : version === 'same-time-content'
          ? { ...accepted, lines: [line({ text: 'C' })] }
          : accepted;
    await store.write(stored);
    const clears: string[] = [];
    await clearAcceptedTranscriptSnapshot(
      {
        ...store,
        read:
          version === 'read-fails'
            ? async () => {
                throw new Error('blocked');
              }
            : store.read,
        clearIfVersion: async (key, revision) => {
          clears.push(key);
          return store.clearIfVersion(key, revision);
        },
      },
      accepted,
      () => version === 'owned',
    );
    expect(clears).toEqual(version === 'same' ? [accepted.key] : []);
    expect(await store.read(accepted.key)).toEqual(version === 'same' ? null : stored);
  });
}

test('MS3b fix3 compare-delete rechecks mounted ownership after deferred read', async () => {
  const store = createMemoryTranscriptHistory();
  const accepted = snapshot();
  await store.write(accepted);
  let owned = false;
  let resolveRead: (value: TranscriptSnapshot | null) => void = () => {};
  const read = new Promise<TranscriptSnapshot | null>((resolve) => {
    resolveRead = resolve;
  });
  const pending = clearAcceptedTranscriptSnapshot(
    { ...store, read: () => read },
    accepted,
    () => owned,
  );
  owned = true;
  resolveRead(accepted);
  await pending;
  expect(await store.read(accepted.key)).toEqual(accepted);
});

test('MS3b fix4 atomic cleanup preserves capture C between read A and conditional delete', async () => {
  const store = createMemoryTranscriptHistory();
  const accepted = snapshot();
  const newer = {
    ...accepted,
    revision: crypto.randomUUID(),
    capturedAt: accepted.capturedAt,
    lines: [line({ text: 'C' })],
  };
  await store.write(accepted);
  await clearAcceptedTranscriptSnapshot(
    {
      ...store,
      read: async (key) => {
        const read = await store.read(key);
        await store.write(newer);
        return read;
      },
    },
    accepted,
    () => false,
  );
  expect(await store.read(accepted.key)).toEqual(newer);
  expect(await store.clearIfVersion(accepted.key, accepted.revision)).toBe(false);
  expect(await store.clearIfVersion(newer.key, newer.revision)).toBe(true);
  expect(await store.read(newer.key)).toBeNull();
});

// The only unconditional cleanup is an identity-boundary sweep: the version is irrelevant.
for (const version of ['same', 'newer'] as const) {
  test(`MS3b fix6 viewer sweep exception / ${version}`, async () => {
    const store = createMemoryTranscriptHistory();
    const own = snapshot({ key: 'viewer:lesson', revision: crypto.randomUUID(), capturedAt: 1 });
    const foreign = snapshot({
      key: 'other:lesson',
      revision: crypto.randomUUID(),
      capturedAt: version === 'same' ? 1 : 2,
    });
    await store.write(own);
    await store.write(foreign);
    await forgetOtherAuthors(store, 'viewer');
    expect(await store.read(own.key)).toEqual(own);
    expect(await store.read(foreign.key)).toBeNull();
  });
}

test('MS3d memory cleanup preserves a replacement with equal capturedAt and different revision', async () => {
  const store = createMemoryTranscriptHistory();
  const first = snapshot({ revision: 'first' });
  const second = { ...first, revision: 'second' };
  await store.write(first);
  await store.write(second);
  expect(await store.clearIfVersion(first.key, first.revision)).toBe(false);
  expect(await store.read(first.key)).toEqual(second);
  expect(await store.clearIfVersion(second.key, second.revision)).toBe(true);
  expect(await store.read(second.key)).toBeNull();
});

test('MS3d detached uk snapshot skips a matching orphan: matching track, else default track', () => {
  const detached = snapshot({ key: 'user:lesson:track:gone', language: 'UK' });
  const orphan: TranscriptRecoveryTab = {
    kind: 'orphan',
    instanceId: 'lang:uk',
    language: 'uk',
    isDefault: false,
  };
  const other: TranscriptRecoveryTab = {
    kind: 'track',
    instanceId: 'track:fr',
    language: 'fr',
    isDefault: false,
  };
  const defaultTrack: TranscriptRecoveryTab = {
    kind: 'track',
    instanceId: 'track:en',
    language: 'en',
    isDefault: true,
  };
  expect(place([detached], [defaultTrack, orphan])[0]?.instanceId).toBe('track:en');
  expect(place([detached], [orphan, other, defaultTrack])[0]?.instanceId).toBe('track:en');
  expect(
    place(
      [detached],
      [other, defaultTrack, { ...other, instanceId: 'track:uk', language: 'uk' }],
    )[0]?.instanceId,
  ).toBe('track:uk');
  // No default track (tracks still processing, or only orphans): never an orphan or a
  // non-default track; the snapshot waits in storage.
  expect(place([detached], [{ ...orphan, isDefault: true }, other])).toEqual([]);
  expect(place([detached], [orphan])).toEqual([]);
  // No tabs at all: the single editor.
  expect(place([detached], [])[0]?.instanceId).toBe('');
});

test.each([
  { key: 'another-key' },
  { language: 'es' },
  { baseLanguage: 'fr' },
  { expectedUpdatedAt: 'T2' },
  { draftUpdatedAt: 'D2' },
])('GAP accepted cleanup preserves a same-revision snapshot differing in %j', async (change) => {
  const accepted = snapshot({ baseLanguage: 'en', draftUpdatedAt: 'D1' });
  const replacement = { ...accepted, ...change };
  const memory = createMemoryTranscriptHistory();
  await memory.write(replacement);
  const clearIfVersion = rstest.fn(async () => true);
  await clearAcceptedTranscriptSnapshot(
    { ...memory, read: async () => replacement, clearIfVersion },
    accepted,
    () => false,
  );
  expect(clearIfVersion).not.toHaveBeenCalled();
  expect(await memory.read(replacement.key)).toEqual(replacement);
});
