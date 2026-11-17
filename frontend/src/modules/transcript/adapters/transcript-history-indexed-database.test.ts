import { afterEach, expect, rstest, test } from '@rstest/core';
import { createIndexedDbTranscriptHistory } from './transcript-history-indexed-database';

// Only the request/event boundary is faked: read() still takes the stored unknown value through
// the production adapter's validation. This does not prove browser persistence or migrations.
function storedValue(value: unknown) {
  const get = () => {
    const request = Object.assign(new EventTarget(), { result: value });
    queueMicrotask(() => request.dispatchEvent(new Event('success')));
    return request;
  };
  const transaction = Object.assign(new EventTarget(), {
    objectStore: () => ({ get }),
  });
  const open = rstest.fn(() => {
    const request = Object.assign(new EventTarget(), {
      result: { transaction: () => transaction },
    });
    queueMicrotask(() => request.dispatchEvent(new Event('success')));
    return request;
  });
  rstest.stubGlobal('indexedDB', { open });
  return open;
}

afterEach(() => {
  rstest.unstubAllGlobals();
});

const snapshot = {
  key: 'author:lesson:track:one',
  revision: crypto.randomUUID(),
  capturedAt: 1,
  language: 'es',
  expectedUpdatedAt: 'T1',
  lines: [{ id: 'line', startMs: '0', endMs: '1000', text: 'Unsaved', documentIds: [] }],
};

for (const base of [{ baseLanguage: 'en' }, {}]) {
  test(`M3-R13 IndexedDB accepts and returns ${'baseLanguage' in base ? 'baseLanguage' : 'legacy snapshot without baseLanguage'} at version 1`, async () => {
    const open = storedValue({ ...snapshot, ...base });
    const store = createIndexedDbTranscriptHistory();
    expect(store).not.toBeNull();
    expect(await store?.read(snapshot.key)).toEqual({ ...snapshot, ...base });
    expect(open).toHaveBeenCalledWith('courses-transcript-history', 1);
  });
}

for (const value of [null, { ...snapshot, baseLanguage: 42 }, { ...snapshot, lines: ['bad'] }]) {
  test(`M3-R13 IndexedDB refuses malformed snapshot ${JSON.stringify(value)}`, async () => {
    storedValue(value);
    expect(await createIndexedDbTranscriptHistory()?.read(snapshot.key)).toBeNull();
  });
}

for (const source of [[{ key: 'author:lesson:lang:es', revision: 'source-revision' }], undefined]) {
  test(`MS3b fix4 IndexedDB roundtrips acceptedSource=${JSON.stringify(source)}`, async () => {
    const value = { ...snapshot, ...(source === undefined ? {} : { acceptedSource: source }) };
    storedValue(value);
    expect(await createIndexedDbTranscriptHistory()?.read(snapshot.key)).toEqual(value);
  });
}
for (const source of [
  null,
  [{ key: 42, revision: 'source' }],
  [{ key: 'k', revision: 1 }],
  [{ key: 'k' }],
  { key: 'k', capturedAt: 1 },
]) {
  test(`MS3b fix4 IndexedDB rejects malformed acceptedSource=${JSON.stringify(source)}`, async () => {
    storedValue({ ...snapshot, acceptedSource: source });
    expect(await createIndexedDbTranscriptHistory()?.read(snapshot.key)).toBeNull();
  });
}

for (const outcome of ['same', 'newer', 'missing', 'abort'] as const) {
  test(`MS3b fix4 atomic IndexedDB cleanup ${outcome} uses one readwrite transaction and waits for commit`, async () => {
    const transaction = new EventTarget();
    const value =
      outcome === 'missing'
        ? undefined
        : { ...snapshot, revision: outcome === 'newer' ? 'newer-revision' : snapshot.revision };
    const request = Object.assign(new EventTarget(), { result: value });
    const remove = rstest.fn();
    const get = rstest.fn(() => request);
    const objectStore = rstest.fn(() => ({ get, delete: remove }));
    const start = rstest.fn(() => Object.assign(transaction, { objectStore }));
    const open = () => {
      const opening = Object.assign(new EventTarget(), { result: { transaction: start } });
      queueMicrotask(() => opening.dispatchEvent(new Event('success')));
      return opening;
    };
    rstest.stubGlobal('indexedDB', { open });
    const store = createIndexedDbTranscriptHistory();
    if (store === null) throw new Error('test requires IndexedDB stub');
    let settled = false;
    const pending = store.clearIfVersion(snapshot.key, snapshot.revision);
    const observed = pending.then((value) => {
      settled = true;
      return value;
    });
    await Promise.resolve();
    await Promise.resolve();
    expect(start).toHaveBeenCalledExactlyOnceWith('snapshots', 'readwrite');
    expect(objectStore).toHaveBeenCalledExactlyOnceWith('snapshots');
    expect(get).toHaveBeenCalledExactlyOnceWith(snapshot.key);
    expect(remove).not.toHaveBeenCalled();
    request.dispatchEvent(new Event('success'));
    if (outcome === 'same' || outcome === 'abort')
      expect(remove).toHaveBeenCalledExactlyOnceWith(snapshot.key);
    else expect(remove).not.toHaveBeenCalled();
    expect(start).toHaveBeenCalledTimes(1);
    await Promise.resolve();
    expect(settled).toBe(false);
    if (outcome === 'abort') {
      transaction.dispatchEvent(new Event('abort'));
      await expect(observed).rejects.toThrow('transaction aborted');
    } else {
      transaction.dispatchEvent(new Event('complete'));
      expect(await observed).toBe(outcome === 'same');
    }
  });
}

for (const count of [0, 1, 8]) {
  test(`MS3b fix5 IndexedDB validates ancestry list length ${count}`, async () => {
    const acceptedSource = Array.from({ length: count }, (_, index) => ({
      key: `ancestor-${index}`,
      revision: crypto.randomUUID(),
    }));
    const value = { ...snapshot, acceptedSource };
    storedValue(value);
    expect(await createIndexedDbTranscriptHistory()?.read(snapshot.key)).toEqual(value);
  });
}
for (const acceptedSource of [
  [null],
  [{ key: 'k', revision: Infinity }],
  [{ key: 'k', revision: 1 }],
  Array.from({ length: 9 }, (_, index) => ({ key: `ancestor-${index}`, revision: String(index) })),
]) {
  test(`MS3b fix5 IndexedDB rejects malformed or unbounded ancestry ${JSON.stringify(acceptedSource)}`, async () => {
    storedValue({ ...snapshot, acceptedSource });
    expect(await createIndexedDbTranscriptHistory()?.read(snapshot.key)).toBeNull();
  });
}

test('MS3d IndexedDB rejects a stored snapshot without revision', async () => {
  const { revision: _revision, ...legacy } = snapshot;
  storedValue(legacy);
  expect(await createIndexedDbTranscriptHistory()?.read(snapshot.key)).toBeNull();
});

test('MS3d IndexedDB cleanup preserves a replacement with equal capturedAt and different revision', async () => {
  // Stateful request/event boundary: production write/read/atomic cleanup all execute.
  let stored: unknown;
  const transaction = () => {
    const tx = new EventTarget();
    const request = (action: () => unknown) => {
      const req = Object.assign(new EventTarget(), { result: stored });
      queueMicrotask(() => {
        req.result = action();
        req.dispatchEvent(new Event('success'));
        queueMicrotask(() => tx.dispatchEvent(new Event('complete')));
      });
      return req;
    };
    return Object.assign(tx, {
      objectStore: () => ({
        put: (value: unknown) =>
          request(() => {
            stored = value;
            return snapshot.key;
          }),
        get: () => request(() => stored),
        delete: () =>
          request(() => {
            stored = undefined;
          }),
      }),
    });
  };
  rstest.stubGlobal('indexedDB', {
    open: () => {
      const req = Object.assign(new EventTarget(), { result: { transaction } });
      queueMicrotask(() => req.dispatchEvent(new Event('success')));
      return req;
    },
  });
  const store = createIndexedDbTranscriptHistory();
  if (store === null) throw new Error('test requires IndexedDB stub');
  const second = { ...snapshot, revision: 'replacement' };
  await store.write(snapshot);
  await store.write(second);
  expect(await store.clearIfVersion(snapshot.key, snapshot.revision)).toBe(false);
  expect(await store.read(snapshot.key)).toEqual(second);
  expect(await store.clearIfVersion(second.key, second.revision)).toBe(true);
  expect(await store.read(snapshot.key)).toBeNull();
});
