import { createIndexedDbTranscriptHistory } from '../../adapters/storage/transcript-history-indexed-database';
import { createMemoryTranscriptHistory } from '../../adapters/storage/transcript-history-memory';
import type { TranscriptHistoryStore } from '../../ports/TranscriptHistoryStore';

/**
 * One store per composition, never a module-level singleton: SSR renders many requests in one
 * process, and a shared mutable store would leak one author's unsaved lines into another's page.
 *
 * On the server and in a browser that refuses IndexedDB this is the in-memory store, so the
 * editor has no branch to take - recovery simply never has anything to offer.
 */
export function createTranscriptHistory(): TranscriptHistoryStore {
  return createIndexedDbTranscriptHistory() ?? createMemoryTranscriptHistory();
}
