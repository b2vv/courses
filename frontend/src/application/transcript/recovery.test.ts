import { expect, test } from '@rstest/core';
import {
  captureDirective,
  recoveryOffer,
  transcriptHistoryKey,
  type TranscriptDraftLine,
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
  key: 'user:lesson:uk',
  capturedAt: 1_700_000_000_000,
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

test('the key separates authors, lessons and languages', () => {
  const parts = { userId: 'user-1', lessonId: 'lesson-1', language: 'uk' };
  expect(transcriptHistoryKey(parts)).toBe('user-1:lesson-1:uk');
  // A snapshot outlives the session on disk; without the author in the key a shared machine
  // offers the next person a course they do not own.
  expect(transcriptHistoryKey({ ...parts, userId: 'user-2' })).not.toBe(
    transcriptHistoryKey(parts),
  );
  expect(transcriptHistoryKey({ ...parts, language: 'en' })).not.toBe(transcriptHistoryKey(parts));
});
