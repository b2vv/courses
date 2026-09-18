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

export type TranscriptSnapshot = {
  readonly key: string;
  readonly capturedAt: number;
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
  readonly lines: readonly TranscriptDraftLine[];
};

export type TranscriptBaseline = {
  readonly expectedUpdatedAt: string | undefined;
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
  return stored.expectedUpdatedAt === baseline.expectedUpdatedAt
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
 * The key scopes a snapshot to one author, one lesson and one language.
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
 * All three parts are server-constrained (two UUIDs and a validated BCP-47 tag), so none can
 * contain the delimiter.
 */
export function transcriptHistoryKey(parts: {
  readonly userId: string;
  readonly lessonId: string;
  readonly language: string;
}): string {
  return `${parts.userId}:${parts.lessonId}:${parts.language}`;
}
