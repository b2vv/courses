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

function cueTime(ms: number): string {
  const clamped = Math.max(0, Math.floor(ms));
  const hours = Math.floor(clamped / 3_600_000);
  const minutes = Math.floor((clamped % 3_600_000) / 60_000);
  const seconds = Math.floor((clamped % 60_000) / 1000);
  const millis = clamped % 1000;
  return `${String(hours).padStart(2, '0')}:${String(minutes).padStart(2, '0')}:${String(
    seconds,
  ).padStart(2, '0')}.${String(millis).padStart(3, '0')}`;
}

/**
 * Serialise segments as WebVTT.
 *
 * The output is deliberately plain: no STYLE, no REGION, no positioning. Cue text is written as
 * text, never as markup — the same rule the rendered transcript obeys, for the same reason, and
 * here it matters twice because a browser parses this file itself.
 *
 * A zero-length segment is skipped: WebVTT requires the end to be strictly after the start, so
 * emitting one would make the whole file unparseable rather than lose one line. The domain
 * permits such a segment (it is simply never active), so the format is where it stops.
 */
export function toWebVTT(segments: readonly TranscriptSegment[]): string {
  const cues = segments
    .filter((segment) => segment.endMs > segment.startMs)
    .map((segment) => {
      // A blank line inside cue text terminates the cue, so a payload containing one would shift
      // every following cue's meaning. Collapsing it keeps one segment as one cue.
      const text = segment.text.replace(/\r?\n\s*\r?\n/g, '\n').trim();
      const speaker = segment.speaker === undefined ? '' : `<v ${segment.speaker}>`;
      return `${cueTime(segment.startMs)} --> ${cueTime(segment.endMs)}\n${speaker}${text}`;
    });
  return `WEBVTT\n\n${cues.join('\n\n')}${cues.length === 0 ? '' : '\n'}`;
}
