import type { TranscriptHistoryStore } from '../ports/TranscriptHistoryStore';

/**
 * Local recovery for the transcript editor. The editor holds its lines in component state, so a
 * reload loses them - and Phase 8 makes that worse rather than better, because recognition takes
 * hours and the author is expected to leave the page while it runs.
 *
 * The shape is taken from `cassiopeia-admin-ui/src/shared/lib/form-local-history` (read
 * 2026-09-18), which ships the same mechanism: port, IndexedDB adapter, in-memory adapter for
 * tests, pure decision, presentation hook. Three things it does not do are done here, and one
 * thing it does is deliberately not copied - see `docs/asr/spec.md` §7.6.
 */

/**
 * What the editor actually holds: times as TEXT, because parsing them per keystroke turns a
 * half-typed "1" into a real value and reorders the lines under the cursor.
 *
 * The snapshot is taken in this shape rather than in parsed segments so that a restore returns
 * the screen the author left, including a time they were halfway through typing. A snapshot of
 * parsed segments could not represent that state at all, and would have to drop the line.
 */
export type TranscriptDraftLine = {
  readonly id: string;
  readonly startMs: string;
  readonly endMs: string;
  readonly text: string;
  readonly documentIds: readonly string[];
};

export type AcceptedTranscriptSource = {
  readonly key: string;
  readonly capturedAt: number;
};

export const maximumAcceptedTranscriptSources = 8;

/** A captured editor retains cleanup ownership of the detached version it restored. */
export type TranscriptSnapshot = {
  readonly acceptedSource?: readonly AcceptedTranscriptSource[];
  readonly key: string;
  readonly capturedAt: number;
  /** The editor language at capture time. Older snapshots may omit this field. */
  readonly language?: string;
  /** Language owning expectedUpdatedAt; absent on legacy snapshots. */
  readonly baseLanguage?: string;
  /**
   * The server token the edits were made against. `undefined` is a real value: it means the
   * lesson had no transcript when the author started typing, which is exactly what the write
   * path reads as "the author believes nothing is there".
   *
   * Carrying it is the first of the three additions. Without it a restore silently resurrects
   * work over a base that moved - and in Phase 8 the base moves on its own, because recognition
   * writes while the author is away.
   */
  readonly expectedUpdatedAt: string | undefined;
  /** The ASR draft generation the lines came from, when they were loaded from a draft. */
  readonly draftUpdatedAt?: string;
  readonly lines: readonly TranscriptDraftLine[];
};

export type TranscriptBaseline = {
  readonly acceptedSource?: readonly AcceptedTranscriptSource[] | undefined;
  /** The editor language used for newly captured snapshots. */
  readonly language?: string;
  /** Language owning expectedUpdatedAt; absent on legacy snapshots. */
  readonly baseLanguage?: string;
  readonly expectedUpdatedAt: string | undefined;
  /** The current ASR draft generation, when the editor's lines came from a draft. */
  readonly draftUpdatedAt?: string;
  readonly lines: readonly TranscriptDraftLine[];
};

export type RecoveryOffer =
  | { readonly kind: 'none' }
  /** Stored work differs from the server, and was made against the same server version. */
  | { readonly kind: 'offer'; readonly snapshot: TranscriptSnapshot }
  /**
   * Stored work differs AND the server has moved since it was captured. Restoring is still the
   * author's to choose, but it is an overwrite of someone else's newer text, so it may never be
   * presented as the same action.
   */
  | { readonly kind: 'offer-stale'; readonly snapshot: TranscriptSnapshot };

function sameLines(
  left: readonly TranscriptDraftLine[],
  right: readonly TranscriptDraftLine[],
): boolean {
  if (left.length !== right.length) return false;
  return left.every((line, index) => {
    const other = right[index];
    if (other === undefined) return false;
    return (
      line.id === other.id &&
      line.startMs === other.startMs &&
      line.endMs === other.endMs &&
      line.text === other.text &&
      line.documentIds.length === other.documentIds.length &&
      line.documentIds.every((id, position) => id === other.documentIds[position])
    );
  });
}

/**
 * Whether anything is worth offering back.
 *
 * The comparison is against the server baseline, not against the mere existence of a snapshot:
 * a stored copy identical to what the server already returns is not a recovery, and offering it
 * would train the author to dismiss the prompt without reading it.
 */
export function recoveryOffer(
  stored: TranscriptSnapshot | null,
  baseline: TranscriptBaseline,
  dismissed: boolean,
): RecoveryOffer {
  if (stored === null || dismissed) return { kind: 'none' };
  if (sameLines(stored.lines, baseline.lines)) return { kind: 'none' };
  const baseMatches = stored.expectedUpdatedAt === baseline.expectedUpdatedAt;
  const draftMatches =
    stored.draftUpdatedAt === undefined ? true : stored.draftUpdatedAt === baseline.draftUpdatedAt;
  return baseMatches && draftMatches
    ? { kind: 'offer', snapshot: stored }
    : { kind: 'offer-stale', snapshot: stored };
}

export type CaptureDirective = { readonly kind: 'write' } | { readonly kind: 'clear' };

/**
 * Editing back to the server's text is not a draft worth keeping, and leaving the old snapshot
 * behind would offer it again on the next visit. So "no difference" clears rather than skips.
 */
export function captureDirective(
  current: readonly TranscriptDraftLine[],
  baseline: TranscriptBaseline,
): CaptureDirective {
  return sameLines(current, baseline.lines) ? { kind: 'clear' } : { kind: 'write' };
}

/**
 * The key scopes a snapshot to one author and one lesson. The language belongs in the snapshot
 * because the editor has no language switcher and must restore the language together with its
 * lines.
 *
 * The user id is the second of the three additions: a snapshot outlives the session on disk, and
 * without it a shared machine offers the next person a course they do not own.
 *
 * The tab scoping that `form-local-history` performs is deliberately NOT copied. It gives each
 * tab its own bucket, which is right for a per-tab edit session and wrong here: a browser that
 * was killed reopens with a fresh tab identity, and a bucket keyed by it would be invisible to
 * the very crash this mechanism exists to survive. The cost is accepted and named - two tabs on
 * one lesson share a snapshot, and the last capture wins.
 *
 * Both parts are server-constrained UUIDs, so neither can contain the delimiter.
 */
/**
 * Which stored snapshots must go, given who the viewer now is.
 *
 * A snapshot outlives the session on disk. Keying it by author stops one person being OFFERED
 * another's text, but it does not remove the text, and on a shared machine "not offered" is not
 * the same as "not there". So the arrival of a different viewer - including no viewer at all,
 * after a sign-out - is what sweeps the rest away.
 *
 * `null` means the viewer is known to be nobody, not that the session is still loading: a
 * pending or failed identity query must never reach this, or a transient network error would
 * delete work the author has not saved.
 */
export function foreignSnapshotKeys(
  keys: readonly string[],
  viewer: string | null,
): readonly string[] {
  if (viewer === null) return keys;
  return keys.filter((key) => key.slice(0, key.indexOf(':')) !== viewer);
}

/**
 * Sweep what the current viewer must not keep. Best-effort, like everything else that touches
 * this store: a refusal costs a stale snapshot on disk, and must never interrupt the page the
 * person is actually on.
 */
export async function forgetOtherAuthors(
  store: TranscriptHistoryStore,
  viewer: string | null,
): Promise<void> {
  try {
    const foreign = foreignSnapshotKeys(await store.keys(), viewer);
    await Promise.all(foreign.map((key) => store.clear(key)));
  } catch {
    // Nothing to report: the viewer did not ask for this, and it is not their task that failed.
  }
}

export function transcriptHistoryKey(parts: {
  readonly userId: string;
  readonly lessonId: string;
  readonly instanceId?: string;
}): string {
  return `${parts.userId}:${parts.lessonId}${parts.instanceId ? `:${parts.instanceId}` : ''}`;
}

/** Current product tab identities, independent of media-module types. Empty means one editor. */
export type TranscriptRecoveryTab = {
  readonly instanceId: string;
  readonly language: string;
  readonly isDefault: boolean;
};

export type DetachedTranscriptSnapshot = {
  readonly instanceId: string;
  readonly legacy: boolean;
  readonly snapshot: TranscriptSnapshot;
};

export function isLessonHistoryKey(key: string, lessonKey: string): boolean {
  return key === lessonKey || key.startsWith(`${lessonKey}:`);
}

/** Read-only placement: never migrate or write a snapshot under another tab's key. */
export function detachedTranscriptSnapshots(
  lessonKey: string,
  keys: readonly string[],
  snapshots: readonly TranscriptSnapshot[],
  tabs: readonly TranscriptRecoveryTab[],
): readonly DetachedTranscriptSnapshot[] {
  const storedKeys = new Set(keys.filter((key) => isLessonHistoryKey(key, lessonKey)));
  const ownKeys = new Set(
    tabs.length === 0 ? [lessonKey] : tabs.map((tab) => `${lessonKey}:${tab.instanceId}`),
  );
  const fallback = tabs.find((tab) => tab.isDefault) ?? tabs[0];
  return snapshots
    .filter((snapshot) => storedKeys.has(snapshot.key) && !ownKeys.has(snapshot.key))
    .sort((a, b) => b.capturedAt - a.capturedAt || a.key.localeCompare(b.key))
    .map((snapshot) => {
      const matches = tabs.filter(
        (tab) => tab.language.toLowerCase() === snapshot.language?.toLowerCase(),
      );
      // Construction excludes a track and orphan with the same language; prefer the track
      // defensively if a caller supplies both.
      const target = matches.find((tab) => tab.instanceId.startsWith('track:')) ?? matches[0];
      return {
        instanceId: (target ?? fallback)?.instanceId ?? '',
        legacy: snapshot.key === lessonKey,
        snapshot,
      };
    });
}

/** A Save may retire only the exact detached copy it accepted, while no mounted tab owns it. */
export async function clearAcceptedTranscriptSnapshot(
  store: TranscriptHistoryStore,
  accepted: TranscriptSnapshot | AcceptedTranscriptSource,
  isOwned: () => boolean,
): Promise<void> {
  try {
    const stored = await store.read(accepted.key);
    if (
      !isOwned() &&
      stored !== null &&
      stored.key === accepted.key &&
      stored.capturedAt === accepted.capturedAt &&
      (!('lines' in accepted) ||
        (stored.language === accepted.language &&
          stored.baseLanguage === accepted.baseLanguage &&
          stored.expectedUpdatedAt === accepted.expectedUpdatedAt &&
          stored.draftUpdatedAt === accepted.draftUpdatedAt &&
          sameLines(stored.lines, accepted.lines)))
    )
      await store.clearIfVersion(accepted.key, accepted.capturedAt);
  } catch {
    // A failed read is not permission to delete; atomic cleanup preserves intervening captures.
  }
}
