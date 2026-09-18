import type { TranscriptSnapshot } from '../application/transcript/recovery';

/**
 * Local, best-effort storage for unsaved editor work. Best-effort is the contract, not a caveat:
 * browsers evict this store under pressure and block it outright in some private modes, so a
 * mechanism that OFFERS recovery degrades correctly where one that promised it would lie.
 *
 * Every method may reject - a quota breach on write is the realistic one, because the domain
 * permits 5000 segments of 1000 runes. Callers treat a rejection as "no local copy" and never as
 * a failure of the thing the author was actually doing.
 */
export interface TranscriptHistoryStore {
  read(key: string): Promise<TranscriptSnapshot | null>;
  write(snapshot: TranscriptSnapshot): Promise<void>;
  clear(key: string): Promise<void>;
}
