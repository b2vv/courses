export type TranscriptDocument = {
  readonly id: string;
  readonly label: string;
  readonly url: string;
};

export type TranscriptSegment = {
  readonly id: string;
  readonly startMs: number;
  readonly endMs: number;
  readonly text: string;
  readonly speaker?: string;
  readonly confidence?: number;
  readonly documentIds: readonly string[];
};

export type LessonTranscript = {
  readonly lessonId: string;
  /** Null when the lesson has no transcript: an absent transcript is a state, not a failure. */
  readonly language: string | null;
  readonly stale: boolean;
  readonly segments: readonly TranscriptSegment[];
  /** Referenced library entries, carried once each rather than inlined into every segment. */
  readonly documents: readonly TranscriptDocument[];
};

export type TranscriptErrorCode =
  | 'course_forbidden'
  | 'lesson_not_found'
  | 'authentication_required'
  | 'invalid_request'
  | 'dependency_unavailable'
  | 'unavailable';

export type TranscriptResult<T> =
  | { readonly ok: true; readonly value: T }
  | { readonly ok: false; readonly error: { readonly code: TranscriptErrorCode } };

/**
 * The index of the segment covering a position, or -1.
 *
 * This is the third copy of one rule: the backend domain owns it (transcript/domain,
 * Segment.Active and Transcript.ActiveSegment), and it is repeated here because the player
 * highlights on every frame and a round trip per frame is not an option. The interval is
 * HALF-OPEN so a boundary belongs to exactly one segment, and when cues overlap — which WebVTT
 * and SRT both permit for two speakers at once — the earliest wins, so the answer is always one
 * segment or none. Changing this without changing the Go copy makes the highlight disagree with
 * anything the server derives from the same times.
 */
export function activeSegmentIndex(
  segments: readonly TranscriptSegment[],
  positionMs: number,
): number {
  if (!Number.isFinite(positionMs)) return -1;
  for (let index = 0; index < segments.length; index += 1) {
    const segment = segments[index];
    if (segment !== undefined && positionMs >= segment.startMs && positionMs < segment.endMs) {
      return index;
    }
  }
  return -1;
}

/** Segments whose text contains the query, case-insensitively. Search runs over the WHOLE lesson
 * because the read is never paginated; a partial list would make a miss indistinguishable from an
 * absent word. */
export function searchSegments(
  segments: readonly TranscriptSegment[],
  query: string,
): readonly TranscriptSegment[] {
  const needle = query.trim().toLocaleLowerCase();
  if (needle === '') return [];
  return segments.filter((segment) => segment.text.toLocaleLowerCase().includes(needle));
}

export function documentsOf(
  transcript: LessonTranscript,
  segment: TranscriptSegment,
): readonly TranscriptDocument[] {
  return segment.documentIds
    .map((id) => transcript.documents.find((document) => document.id === id))
    .filter((document): document is TranscriptDocument => document !== undefined);
}
