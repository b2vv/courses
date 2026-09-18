import type { TranscriptSnapshot } from '../../application/transcript/recovery';
import type { TranscriptHistoryStore } from '../../ports/TranscriptHistoryStore';

/**
 * The store used on the server, in tests, and in any browser that refuses IndexedDB. It is not a
 * test double bolted on afterwards: having a second implementation is what lets the editor be
 * tested without a real database, and what lets SSR render the same component tree.
 *
 * It keeps nothing across a reload, which is honest - a recovery offer simply never appears.
 */
export function createMemoryTranscriptHistory(): TranscriptHistoryStore {
  const snapshots = new Map<string, TranscriptSnapshot>();
  return {
    read: (key) => Promise.resolve(snapshots.get(key) ?? null),
    write: (snapshot) => {
      snapshots.set(snapshot.key, snapshot);
      return Promise.resolve();
    },
    clear: (key) => {
      snapshots.delete(key);
      return Promise.resolve();
    },
  };
}
