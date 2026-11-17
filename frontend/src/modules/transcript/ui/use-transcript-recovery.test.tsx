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
      if (version === 'newer' && path === 'Save')
        expect(f.result.current.offer).toEqual({ kind: 'none' });
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

function recoverySnapshot(key: string, text = 'Recovered'): TranscriptSnapshot {
  return {
    key,
    revision: `revision:${key}:${text}`,
    capturedAt: 1,
    language: 'uk',
    baseLanguage: 'en',
    expectedUpdatedAt: 'T1',
    lines: [{ ...line, text }],
  };
}

test('GAP switching a loaded visit hides its offer until the new store read resolves', async () => {
  const store = createMemoryTranscriptHistory();
  const a = recoverySnapshot('gap-switch:a');
  await store.write(a);
  const next = deferred<TranscriptSnapshot | null>();
  const other = { ...createMemoryTranscriptHistory(), read: () => next.promise };
  const f = renderHook(
    ({ key, source }) => useTranscriptRecovery({ historyKey: key, store: source, baseline }),
    { initialProps: { key: a.key, source: store } },
  );
  await act(async () => {});
  expect(f.result.current.offer).toMatchObject({ snapshot: a });
  f.rerender({ key: 'gap-switch:b', source: other });
  expect(f.result.current.isLoaded).toBe(false);
  expect(f.result.current.offer).toEqual({ kind: 'none' });
  await act(async () => next.resolve(null));
  expect(f.result.current.isLoaded).toBe(true);
});

for (const outcome of ['resolve', 'reject'] as const) {
  test(`GAP obsolete initial read ${outcome} cannot replace the current visit offer`, async () => {
    const store = createMemoryTranscriptHistory();
    const old = deferred<TranscriptSnapshot | null>();
    const b = recoverySnapshot(`gap-obsolete-${outcome}:b`);
    await store.write(b);
    const read = rstest.spyOn(store, 'read').mockReturnValueOnce(old.promise);
    const f = renderHook(({ key }) => useTranscriptRecovery({ historyKey: key, store, baseline }), {
      initialProps: { key: `gap-obsolete-${outcome}:a` },
    });
    f.rerender({ key: b.key });
    await act(async () => {});
    expect(f.result.current.offer).toMatchObject({ snapshot: b });
    await act(async () => {
      if (outcome === 'resolve') old.resolve(recoverySnapshot('obsolete-a'));
      else old.reject(new Error('obsolete read refused'));
    });
    expect(read).toHaveBeenCalledTimes(2);
    expect(f.result.current.offer).toMatchObject({ snapshot: b });
    expect(f.result.current.isLoaded).toBe(true);
  });
}

test('GAP obsolete read completion cannot authorize retirement in an unread visit', async () => {
  const store = createMemoryTranscriptHistory();
  const a = deferred<TranscriptSnapshot | null>();
  const b = deferred<TranscriptSnapshot | null>();
  rstest.spyOn(store, 'read').mockReturnValueOnce(a.promise).mockReturnValueOnce(b.promise);
  const clear = rstest.spyOn(store, 'clearIfVersion');
  const f = renderHook(({ key }) => useTranscriptRecovery({ historyKey: key, store, baseline }), {
    initialProps: { key: 'gap-unread:a' },
  });
  f.rerender({ key: 'gap-unread:b' });
  await act(async () => a.resolve(null));
  act(() => f.result.current.captureNow([{ ...line, text: 'own B' }]));
  await act(async () => {});
  act(() => f.result.current.finalize());
  await act(async () => {});
  expect(clear).not.toHaveBeenCalled();
  await act(async () => b.resolve(null));
});

for (const action of ['finalize', 'clean capture'] as const) {
  test(`GAP ${action} cannot retire an own write before the initial read finishes`, async () => {
    const store = createMemoryTranscriptHistory();
    const pending = deferred<TranscriptSnapshot | null>();
    rstest.spyOn(store, 'read').mockReturnValueOnce(pending.promise);
    const clear = rstest.spyOn(store, 'clearIfVersion');
    const key = `gap-initial:${action}`;
    sessionStorage.setItem(`transcript-recovery-dismissed:${key}`, '1');
    const f = renderHook(() => useTranscriptRecovery({ historyKey: key, store, baseline }));
    act(() => f.result.current.captureNow([{ ...line, text: 'own dirty' }]));
    await act(async () => {});
    // Successful capture clears a previous dismissal; reinstate it to observe finalize's guard.
    sessionStorage.setItem(`transcript-recovery-dismissed:${key}`, '1');
    act(() => {
      if (action === 'finalize') f.result.current.finalize();
      else f.result.current.captureNow([line]);
    });
    await act(async () => {});
    expect(clear).not.toHaveBeenCalled();
    expect(sessionStorage.getItem(`transcript-recovery-dismissed:${key}`)).toBe('1');
    await act(async () => pending.resolve(null));
    sessionStorage.removeItem(`transcript-recovery-dismissed:${key}`);
  });
}

test('GAP finalization waits for its write commit before conditional retirement', async () => {
  const memory = createMemoryTranscriptHistory();
  const commit = deferred<void>();
  const write = rstest.fn(async (snapshot: TranscriptSnapshot) => {
    await commit.promise;
    await memory.write(snapshot);
  });
  const clear = rstest.spyOn(memory, 'clearIfVersion');
  const store = { ...memory, write };
  const f = renderHook(() => useTranscriptRecovery({ historyKey: 'gap-commit', store, baseline }));
  await act(async () => {});
  act(() => f.result.current.captureNow([{ ...line, text: 'Own' }]));
  act(() => f.result.current.finalize());
  await act(async () => {});
  expect(write).toHaveBeenCalledTimes(1);
  expect(clear).not.toHaveBeenCalled();
  await act(async () => commit.resolve());
  expect(clear).toHaveBeenCalledExactlyOnceWith('gap-commit', write.mock.calls[0]?.[0].revision);
  expect(await memory.read('gap-commit')).toBeNull();
});

test('GAP finalization without an own write retains the accepted recovery source', async () => {
  const store = createMemoryTranscriptHistory();
  const a = recoverySnapshot('gap-accepted');
  await store.write(a);
  const clear = rstest.spyOn(store, 'clearIfVersion');
  const f = renderHook(() => useTranscriptRecovery({ historyKey: a.key, store, baseline }));
  await act(async () => {});
  act(() => f.result.current.accept());
  act(() => f.result.current.finalize());
  await act(async () => {});
  expect(clear).not.toHaveBeenCalled();
  expect(await store.read(a.key)).toEqual(a);
});

test('GAP mismatch refresh from an obsolete visit cannot replace the current offer', async () => {
  const store = createMemoryTranscriptHistory();
  const a = recoverySnapshot('gap-refresh:a');
  const b = recoverySnapshot('gap-refresh:b');
  await store.write(a);
  await store.write(b);
  const delayed = deferred<TranscriptSnapshot | null>();
  const read = rstest.spyOn(store, 'read');
  const clear = rstest.spyOn(store, 'clearIfVersion').mockResolvedValue(false);
  const f = renderHook(({ key }) => useTranscriptRecovery({ historyKey: key, store, baseline }), {
    initialProps: { key: a.key },
  });
  await act(async () => {});
  read.mockReturnValueOnce(delayed.promise);
  act(() => f.result.current.dismiss());
  await act(async () => {});
  expect(clear).toHaveBeenCalledExactlyOnceWith(a.key, a.revision);
  expect(read).toHaveBeenCalledTimes(2);
  f.rerender({ key: b.key });
  await act(async () => {});
  expect(f.result.current.offer).toMatchObject({ snapshot: b });
  await act(async () =>
    delayed.resolve({ ...a, revision: 'foreign', lines: [{ ...line, text: 'foreign' }] }),
  );
  expect(f.result.current.offer).toMatchObject({ snapshot: b });
});

test('GAP repeated Discard retires the refreshed revision rather than its predecessor', async () => {
  const store = createMemoryTranscriptHistory();
  const a = recoverySnapshot('gap-discard');
  const c = { ...a, revision: 'C', lines: [{ ...line, text: 'C' }] };
  await store.write(a);
  const clear = rstest.spyOn(store, 'clearIfVersion');
  const f = renderHook(() => useTranscriptRecovery({ historyKey: a.key, store, baseline }));
  await act(async () => {});
  await store.write(c);
  act(() => f.result.current.dismiss());
  await act(async () => {});
  expect(f.result.current.offer).toMatchObject({ snapshot: c });
  clear.mockClear();
  act(() => f.result.current.dismiss());
  await act(async () => {});
  expect(clear).toHaveBeenCalledExactlyOnceWith(c.key, 'C');
  expect(await store.read(c.key)).toBeNull();
});

test('GAP a vanished mismatch closes the stale own recovery prompt', async () => {
  const store = createMemoryTranscriptHistory();
  const a = recoverySnapshot('gap-vanished');
  await store.write(a);
  const f = renderHook(() => useTranscriptRecovery({ historyKey: a.key, store, baseline }));
  await act(async () => {});
  expect(f.result.current.offer.kind).toBe('offer');
  await store.clear(a.key);
  act(() => f.result.current.captureNow([line]));
  await act(async () => {});
  expect(f.result.current.offer).toEqual({ kind: 'none' });
});

for (const action of ['captureNow', 'noteChange'] as const) {
  test(`GAP suspended ${action} does not arm capture when resumed or unmounted`, async () => {
    const f = fixture();
    await act(async () => {});
    rstest.useFakeTimers();
    act(() => f.result.current.setCapturing(false));
    act(() => f.result.current[action]([{ ...line, text: 'suspended edit' }]));
    expect(f.write).not.toHaveBeenCalled();
    act(() => f.result.current.setCapturing(true));
    act(() => rstest.advanceTimersByTime(idleCaptureMs));
    f.unmount();
    expect(f.write).not.toHaveBeenCalled();
  });
}

function deferred<T>() {
  let resolve: (value: T) => void = () => {
    throw new Error('Deferred promise not initialized');
  };
  let reject: (reason: unknown) => void = () => {
    throw new Error('Deferred promise not initialized');
  };
  const promise = new Promise<T>((resolvePromise, rejectPromise) => {
    resolve = resolvePromise;
    reject = rejectPromise;
  });
  return { promise, resolve, reject };
}
