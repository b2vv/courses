import type { TranscriptSnapshot } from '../../application/transcript/recovery';
import type { TranscriptHistoryStore } from '../../ports/TranscriptHistoryStore';

const databaseName = 'courses-transcript-history';
const databaseVersion = 1;
const storeName = 'snapshots';

/**
 * IndexedDB rather than `localStorage`, for a reason that can be counted: the domain permits
 * `MaximumSegments` 5000 x `MaximumTextRunes` 1000, which is up to five million runes of text
 * alone - about 10 MB in UTF-16, against a `localStorage` quota of roughly 5 MB per origin. At a
 * realistic lesson size both would fit; at the ceiling the domain itself accepts, only this one
 * does, and a store that needs a second story for its own legal maximum is the wrong store.
 *
 * Requests report failure through an `error` event, never by throwing, so every one is turned
 * into a rejection here. `abort` is listened for as well: a quota breach aborts the transaction,
 * and a listener only on `error` would leave that promise pending forever.
 */
export function createIndexedDbTranscriptHistory(): TranscriptHistoryStore | null {
  if (typeof indexedDB === 'undefined') return null;

  let database: Promise<IDBDatabase> | null = null;

  const open = (): Promise<IDBDatabase> => {
    database ??= new Promise<IDBDatabase>((resolve, reject) => {
      let request: IDBOpenDBRequest;
      try {
        request = indexedDB.open(databaseName, databaseVersion);
      } catch (error) {
        reject(error instanceof Error ? error : new Error('transcript history: open refused'));
        return;
      }
      request.addEventListener('upgradeneeded', () => {
        if (!request.result.objectStoreNames.contains(storeName)) {
          request.result.createObjectStore(storeName, { keyPath: 'key' });
        }
      });
      request.addEventListener('success', () => {
        resolve(request.result);
      });
      request.addEventListener('error', () => {
        // A failed open must not be remembered, or every later call inherits one bad moment.
        database = null;
        reject(request.error ?? new Error('transcript history: open failed'));
      });
      request.addEventListener('blocked', () => {
        database = null;
        reject(new Error('transcript history: open blocked by another tab'));
      });
    });
    return database;
  };

  const run = <T>(
    mode: IDBTransactionMode,
    action: (store: IDBObjectStore) => IDBRequest<T>,
  ): Promise<T> =>
    open().then(
      (connection) =>
        new Promise<T>((resolve, reject) => {
          const transaction = connection.transaction(storeName, mode);
          const request = action(transaction.objectStore(storeName));
          request.addEventListener('success', () => {
            resolve(request.result);
          });
          request.addEventListener('error', () => {
            reject(request.error ?? new Error('transcript history: request failed'));
          });
          transaction.addEventListener('abort', () => {
            reject(transaction.error ?? new Error('transcript history: transaction aborted'));
          });
        }),
    );

  return {
    read: (key) =>
      run<TranscriptSnapshot | undefined>('readonly', (store) => store.get(key)).then(
        (value) => value ?? null,
      ),
    write: (snapshot) =>
      run<IDBValidKey>('readwrite', (store) => store.put(snapshot)).then(() => undefined),
    clear: (key) => run<undefined>('readwrite', (store) => store.delete(key)),
  };
}
