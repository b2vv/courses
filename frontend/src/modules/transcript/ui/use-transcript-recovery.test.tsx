import { afterEach, expect, rstest, test } from '@rstest/core';
import { act, renderHook, waitFor } from '@testing-library/react';
import { createMemoryTranscriptHistory } from '../adapters/transcript-history-memory';
import type {
  TranscriptBaseline,
  TranscriptDraftLine,
  TranscriptSnapshot,
} from '../application/recovery';
import { idleCaptureMs, useTranscriptRecovery } from './use-transcript-recovery';

const line: TranscriptDraftLine = {
  id: 'line',
  startMs: '0',
  endMs: '1000',
  text: 'base',
  documentIds: [],
};
const baseline: TranscriptBaseline = {
  language: 'uk',
  baseLanguage: 'en',
  expectedUpdatedAt: 'T1',
  lines: [line],
};
afterEach(() => {
  rstest.useRealTimers();
  rstest.restoreAllMocks();
});

function fixture() {
  const store = createMemoryTranscriptHistory();
  const write = rstest.spyOn(store, 'write');
  const clear = rstest.spyOn(store, 'clear');
  const hook = renderHook(
    ({ base }) =>
      useTranscriptRecovery({ historyKey: 'user:lesson:track:a', store, baseline: base }),
    { initialProps: { base: baseline } },
  );
  return { store, write, clear, ...hook };
}

test('M3-R12 flush on unmount writes exactly once with latest lines and baseline', async () => {
  const f = fixture();
  act(() => f.result.current.noteChange([{ ...line, text: 'first' }]));
  act(() => f.result.current.noteChange([{ ...line, text: 'latest' }]));
  f.rerender({ base: { ...baseline, language: 'es', expectedUpdatedAt: 'T2' } });
  f.unmount();
  expect(f.write).toHaveBeenCalledTimes(1);
  expect(await f.store.read('user:lesson:track:a')).toMatchObject({
    language: 'es',
    baseLanguage: 'en',
    expectedUpdatedAt: 'T2',
    lines: [{ ...line, text: 'latest' }],
  });
});

test('M3-R12 unmount without a pending change performs zero writes or clears', () => {
  const f = fixture();
  f.unmount();
  expect(f.write).not.toHaveBeenCalled();
  expect(f.clear).not.toHaveBeenCalled();
});

test('M3-R12 clean unmount without a known stored version performs no deletion', async () => {
  const f = fixture();
  await waitFor(() => expect(f.result.current.isLoaded).toBe(true));
  act(() => f.result.current.noteChange([{ ...line, text: 'edited' }]));
  act(() => f.result.current.noteChange([line]));
  f.unmount();
  expect(f.write).not.toHaveBeenCalled();
  expect(f.clear).not.toHaveBeenCalled();
});

test('M3-R12 an idle or immediate capture is not repeated on unmount', () => {
  rstest.useFakeTimers();
  const f = fixture();
  act(() => f.result.current.noteChange([{ ...line, text: 'edited' }]));
  act(() => rstest.advanceTimersByTime(idleCaptureMs));
  expect(f.write).toHaveBeenCalledTimes(1);
  f.unmount();
  expect(f.write).toHaveBeenCalledTimes(1);
});

test('M3-R6 capture in A leaves B untouched', async () => {
  const f = fixture();
  await f.store.write({
    key: 'user:lesson:track:b',
    revision: crypto.randomUUID(),
    capturedAt: 1,
    expectedUpdatedAt: 'B1',
    lines: [{ ...line, text: 'B' }],
  });
  f.write.mockClear();
  act(() => f.result.current.noteChange([{ ...line, text: 'A' }]));
  f.unmount();
  expect(f.write).toHaveBeenCalledTimes(1);
  expect((await f.store.read('user:lesson:track:b'))?.lines[0]?.text).toBe('B');
});

for (const failure of ['rejected', 'thrown'] as const) {
  test(`M3-R12 flush swallows a ${failure} store write`, async () => {
    const f = fixture();
    if (failure === 'rejected') f.write.mockRejectedValue(new Error('quota'));
    else
      f.write.mockImplementation(() => {
        throw new Error('quota');
      });
    act(() => f.result.current.noteChange([{ ...line, text: 'edited' }]));
    f.unmount();
    await Promise.resolve();
    expect(f.write).toHaveBeenCalledTimes(1);
  });
}

test('MS3b fix2 / M3-R6 resolved own offer allows ordinary capture and accepted-save finalization', async () => {
  const store = createMemoryTranscriptHistory();
  const key = 'user:lesson:track:a';
  const snapshot = {
    key,
    revision: crypto.randomUUID(),
    capturedAt: 1,
    expectedUpdatedAt: 'T1',
    lines: [{ ...line, text: 'Recovered' }],
  };
  await store.write(snapshot);
  const clear = rstest.spyOn(store, 'clearIfVersion');
  const write = rstest.spyOn(store, 'write');
  const f = renderHook(() => useTranscriptRecovery({ historyKey: key, store, baseline }));
  await waitFor(() => expect(f.result.current.offer.kind).not.toBe('none'));
  act(() => f.result.current.accept());
  act(() => f.result.current.captureNow([line]));
  act(() => f.result.current.captureNow([{ ...line, text: 'Other edit' }]));
  act(() => f.result.current.finalize());
  await act(async () => {});
  f.unmount();
  expect(clear).toHaveBeenCalledTimes(2);
  expect(write).toHaveBeenCalledTimes(1);
  expect(await store.read(key)).toBeNull();
});

test('MS3b fix2 / M3-R12 no clean capture or finalization clears before the initial read resolves', async () => {
  const store = createMemoryTranscriptHistory();
  let resolveRead: (value: null) => void = () => {};
  rstest.spyOn(store, 'read').mockReturnValueOnce(
    new Promise<null>((resolve) => {
      resolveRead = resolve;
    }),
  );
  const clear = rstest.spyOn(store, 'clear');
  const f = renderHook(() =>
    useTranscriptRecovery({ historyKey: 'fix2-initial-clear', store, baseline }),
  );
  act(() => {
    f.result.current.captureNow([line]);
    f.result.current.finalize();
  });
  expect(f.result.current.isLoaded).toBe(false);
  expect(clear).not.toHaveBeenCalled();
  await act(async () => resolveRead(null));
  expect(f.result.current.isLoaded).toBe(true);
  act(() => f.result.current.captureNow([line]));
  expect(clear).not.toHaveBeenCalled();
});

// Ordinary deletion paths all share the same ownership rule, including timer/unmount captures.
for (const path of ['blur', 'idle', 'unmount', 'own Discard', 'Save'] as const) {
  for (const version of ['same', 'newer'] as const) {
    test(`MS3b fix6 ownership ${path} / ${version}`, async () => {
      const store = createMemoryTranscriptHistory();
      const key = `fix6:${path}:${version}`;
      const a = {
        key,
        revision: crypto.randomUUID(),
        capturedAt: 1,
        expectedUpdatedAt: 'T1',
        lines: [{ ...line, text: 'A' }],
      };
      await store.write(a);
      const f = renderHook(() => useTranscriptRecovery({ historyKey: key, store, baseline }));
      await waitFor(() => expect(f.result.current.offer.kind).toBe('offer'));
      let owned: TranscriptSnapshot = a;
      if (path === 'Save') {
        act(() => f.result.current.accept());
        act(() => f.result.current.captureNow(a.lines));
        await act(async () => {});
        const captured = await store.read(key);
        if (captured === null) throw new Error('Save must have its own capture');
        owned = captured;
      }
      const c = {
        ...owned,
        revision: crypto.randomUUID(),
        capturedAt: owned.capturedAt,
        lines: [{ ...line, text: 'C' }],
      };
      if (version === 'newer') await store.write(c);
      if (path === 'own Discard') act(() => f.result.current.dismiss());
      else if (path === 'Save') act(() => f.result.current.finalize());
      else {
        act(() => f.result.current.accept());
        if (path === 'blur') act(() => f.result.current.captureNow([line]));
        else {
          rstest.useFakeTimers();
          act(() => f.result.current.noteChange([line]));
          if (path === 'unmount') f.unmount();
          else act(() => rstest.advanceTimersByTime(idleCaptureMs));
          rstest.useRealTimers();
        }
      }
      await act(async () => {});
      expect(await store.read(key)).toEqual(version === 'same' ? null : c);
      if (version === 'newer' && path !== 'unmount' && path !== 'Save')
        expect(f.result.current.offer).toMatchObject({ kind: 'offer', snapshot: c });
      f.unmount();
      const revisit = renderHook(() => useTranscriptRecovery({ historyKey: key, store, baseline }));
      await waitFor(() => expect(revisit.result.current.isLoaded).toBe(true));
      expect(revisit.result.current.offer.kind).toBe(version === 'newer' ? 'offer' : 'none');
      revisit.unmount();
    });
  }
}

test('MS3b fix6 unknown version never deletes another window capture', async () => {
  const f = fixture();
  await waitFor(() => expect(f.result.current.isLoaded).toBe(true));
  const c = {
    key: 'user:lesson:track:a',
    revision: crypto.randomUUID(),
    capturedAt: 2,
    expectedUpdatedAt: 'T1',
    lines: [{ ...line, text: 'C' }],
  };
  await f.store.write(c);
  act(() => f.result.current.captureNow([line]));
  act(() => f.result.current.finalize());
  await act(async () => {});
  expect(await f.store.read(c.key)).toEqual(c);
});

test('MS3d hook captures unique revisions in one millisecond and Save preserves the other window', async () => {
  const store = createMemoryTranscriptHistory();
  const key = 'ms3d:lesson:track:one';
  const first = renderHook(() => useTranscriptRecovery({ historyKey: key, store, baseline }));
  const second = renderHook(() => useTranscriptRecovery({ historyKey: key, store, baseline }));
  await waitFor(() =>
    expect(first.result.current.isLoaded && second.result.current.isLoaded).toBe(true),
  );
  rstest.spyOn(Date, 'now').mockReturnValue(42);
  act(() => first.result.current.captureNow([{ ...line, text: 'first window' }]));
  const a = await store.read(key);
  act(() => second.result.current.captureNow([{ ...line, text: 'second window' }]));
  const b = await store.read(key);
  if (a === null || b === null) throw new Error('Both captures must exist');
  expect(a.capturedAt).toBe(b.capturedAt);
  expect(a.revision).not.toBe(b.revision);
  act(() => first.result.current.finalize());
  await act(async () => {});
  expect(await store.read(key)).toEqual(b);
  first.unmount();
  second.unmount();
});
