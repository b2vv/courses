import { afterEach, expect, rstest, test } from '@rstest/core';
import { act, renderHook } from '@testing-library/react';
import { createMemoryTranscriptHistory } from '../adapters/transcript-history-memory';
import type { TranscriptRecoveryTab, TranscriptSnapshot } from '../application/recovery';
import { useDetachedTranscriptRecovery } from './use-detached-transcript-recovery';

const snapshot: TranscriptSnapshot = {
  key: 'author:lesson:track:gone',
  revision: 'A',
  capturedAt: 1,
  language: 'en',
  expectedUpdatedAt: 'T1',
  lines: [{ id: 'line', startMs: '0', endMs: '1000', text: 'A', documentIds: [] }],
};
const tabs = [{ kind: 'track', instanceId: 'track:en', language: 'en', isDefault: true }] as const;
afterEach(() => {
  rstest.restoreAllMocks();
});

test('GAP disabled detached inventory performs no storage IO', async () => {
  const store = createMemoryTranscriptHistory();
  const keys = rstest.spyOn(store, 'keys');
  const read = rstest.spyOn(store, 'read');
  const f = renderHook(() =>
    useDetachedTranscriptRecovery('lesson', { authorId: 'author', store }, tabs, false),
  );
  await act(async () => {});
  expect(keys).not.toHaveBeenCalled();
  expect(read).not.toHaveBeenCalled();
  expect(f.result.current).toEqual([]);
});

test('GAP detached inventory ignores a read result belonging to a different listed key', async () => {
  const store = createMemoryTranscriptHistory();
  rstest.spyOn(store, 'keys').mockResolvedValue(['author:lesson:track:listed']);
  const read = rstest.spyOn(store, 'read').mockResolvedValue(snapshot);
  const f = renderHook(() =>
    useDetachedTranscriptRecovery('lesson', { authorId: 'author', store }, tabs),
  );
  await act(async () => {});
  expect(read).toHaveBeenCalledExactlyOnceWith('author:lesson:track:listed');
  expect(f.result.current).toEqual([]);
});

test('GAP obsolete detached inventory cannot erase the completed current inventory', async () => {
  const store = createMemoryTranscriptHistory();
  const old = deferred<readonly string[]>();
  await store.write(snapshot);
  rstest.spyOn(store, 'keys').mockReturnValueOnce(old.promise);
  const f = renderHook(
    ({ lesson }) => useDetachedTranscriptRecovery(lesson, { authorId: 'author', store }, tabs),
    { initialProps: { lesson: 'old' } },
  );
  f.rerender({ lesson: 'lesson' });
  await act(async () => {});
  expect(f.result.current.map((offer) => offer.snapshot)).toEqual([snapshot]);
  await act(async () => old.resolve([]));
  expect(f.result.current.map((offer) => offer.snapshot)).toEqual([snapshot]);
});

for (const change of ['store', 'tabs', 'disabled'] as const) {
  test(`GAP detached inventory hides stale same-lesson offers after ${change} changes`, async () => {
    const store = createMemoryTranscriptHistory();
    await store.write(snapshot);
    const pending = deferred<readonly string[]>();
    const other = { ...createMemoryTranscriptHistory(), keys: () => pending.promise };
    const f = renderHook(
      ({ source, currentTabs, enabled }) =>
        useDetachedTranscriptRecovery(
          'lesson',
          { authorId: 'author', store: source },
          currentTabs,
          enabled,
        ),
      {
        initialProps: {
          source: store,
          currentTabs: [...tabs] as TranscriptRecoveryTab[],
          enabled: true,
        },
      },
    );
    await act(async () => {});
    expect(f.result.current).toHaveLength(1);
    if (change === 'tabs') rstest.spyOn(store, 'keys').mockReturnValue(pending.promise);
    f.rerender({
      source: change === 'store' ? other : store,
      currentTabs: change === 'tabs' ? [{ ...tabs[0], language: 'es' }] : [...tabs],
      enabled: change !== 'disabled',
    });
    expect(f.result.current).toEqual([]);
    await act(async () => pending.resolve([]));
  });
}

test('GAP one refused detached read does not suppress a different valid recovery', async () => {
  const store = createMemoryTranscriptHistory();
  rstest.spyOn(store, 'keys').mockResolvedValue(['author:lesson:track:bad', snapshot.key]);
  const read = rstest
    .spyOn(store, 'read')
    .mockRejectedValueOnce(new Error('read refused'))
    .mockResolvedValueOnce(snapshot);
  const f = renderHook(() =>
    useDetachedTranscriptRecovery('lesson', { authorId: 'author', store }, tabs),
  );
  await act(async () => {});
  expect(read).toHaveBeenCalledTimes(2);
  expect(f.result.current.map((offer) => offer.snapshot)).toEqual([snapshot]);
});

test('GAP inventory refusal remains best effort without an escaped rejection', async () => {
  const store = createMemoryTranscriptHistory();
  const keys = rstest.spyOn(store, 'keys').mockRejectedValue(new Error('keys refused'));
  // Observe escaped promise rejections in this named test, rather than counting a global
  // failed file as a mutation kill. setImmediate is an event-loop barrier, not a race delay.
  const escaped: unknown[] = [];
  const onRejection = (reason: unknown) => {
    escaped.push(reason);
  };
  process.on('unhandledRejection', onRejection);
  try {
    const f = renderHook(() =>
      useDetachedTranscriptRecovery('lesson', { authorId: 'author', store }, tabs),
    );
    await act(async () => {});
    await new Promise<void>((resolve) => setImmediate(resolve));
    expect(keys).toHaveBeenCalledTimes(1);
    expect(f.result.current).toEqual([]);
    expect(escaped).toEqual([]);
  } finally {
    process.off('unhandledRejection', onRejection);
  }
});

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
