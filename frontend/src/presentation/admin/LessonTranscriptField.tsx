import { transcriptHistoryKey } from '../../application/transcript/recovery';
import { useEffect } from 'react';
import type {
  LessonTranscript,
  TranscriptDocument,
  TranscriptErrorCode,
  TranscriptSegment,
} from '../../domain/transcript/transcript';
import type { TranscriptHistoryStore } from '../../ports/TranscriptHistoryStore';
import { TranscriptEditor } from './TranscriptEditor';
import type { TranscriptDraftOffer } from './transcript-draft-offer';

export type CourseEditorTranscripts = {
  /** Loaded lazily per lesson: a course may hold 120 lessons of 5000 lines each, and the editor
   * needs one of them. */
  /** A real hook, named as one: it is called during this component's render, and calling it
   * anything else would hide that from the rules that check hook usage. */
  readonly useLessonTranscript: (lessonId: string) => {
    readonly data: LessonTranscript | undefined;
    readonly isLoading: boolean;
  };
  readonly documents: readonly TranscriptDocument[];
  /**
   * Where unsaved lines survive a reload, and whose they are. The author is part of the key
   * because a snapshot outlives the session on disk: on a shared machine, a key without them
   * would offer the next person a course they do not own.
   */
  readonly history: {
    readonly store: TranscriptHistoryStore;
    readonly authorId: string;
  };
  readonly library: {
    add(document: TranscriptDocument): Promise<TranscriptErrorCode | null>;
    remove(documentId: string): Promise<TranscriptErrorCode | null>;
  };
  readonly save: (
    lessonId: string,
    language: string,
    expectedUpdatedAt: string | undefined,
    segments: readonly TranscriptSegment[],
  ) => Promise<TranscriptErrorCode | null>;
};

type LessonTranscriptFieldProps = CourseEditorTranscripts & {
  readonly lessonId: string;
  /** An ASR draft loaded for this lesson, if the author asked for one — see `TranscriptEditor`. */
  readonly draftOffer?: TranscriptDraftOffer;
  readonly draftUpdatedAt?: string;
  readonly onDraftOfferHandled?: () => void;
  readonly onTranscriptKnown?: (lessonId: string, hasTranscript: boolean) => void;
};

export function LessonTranscriptField({
  lessonId,
  useLessonTranscript,
  documents,
  history,
  save,
  draftOffer,
  draftUpdatedAt,
  onDraftOfferHandled,
  onTranscriptKnown,
}: LessonTranscriptFieldProps) {
  const { data, isLoading } = useLessonTranscript(lessonId);

  useEffect(() => {
    if (!isLoading)
      onTranscriptKnown?.(lessonId, data?.language !== undefined && data.language !== null);
  }, [data, isLoading, lessonId, onTranscriptKnown]);

  // Nothing while loading, deliberately NOT a skeleton. `Skeleton` renders an <output>, whose
  // implicit role is `status` — and the course editor already owns a status region for "changes
  // saved". Two status elements on one page make that message ambiguous to a screen reader and to
  // any test that asks the page for its status. A field that appears when it is ready costs the
  // author nothing here: the editor is far below the fold of the lesson they just saved.
  if (isLoading) return null;

  // A lesson with no transcript yet still gets an editor: the author is expected to receive a
  // machine-made draft and correct it, but nothing stops them writing the first lines by hand
  // before that exists. The token is absent, which is exactly what the write path reads as
  // "the author believes nothing is there".
  const language = data?.language ?? 'uk';
  const expectedUpdatedAt = data?.updatedAt ?? undefined;

  return (
    <TranscriptEditor
      key={`${lessonId}-${expectedUpdatedAt ?? 'new'}`}
      lessonId={lessonId}
      language={language}
      {...(expectedUpdatedAt === undefined ? {} : { expectedUpdatedAt })}
      segments={data?.segments ?? []}
      documents={documents}
      history={{
        store: history.store,
        key: transcriptHistoryKey({ userId: history.authorId, lessonId }),
      }}
      hasSavedTranscript={data?.language !== undefined && data.language !== null}
      actions={{
        save: (nextLanguage, token, segments) => save(lessonId, nextLanguage, token, segments),
      }}
      {...(draftOffer === undefined ? {} : { draftOffer })}
      {...(draftUpdatedAt === undefined ? {} : { draftUpdatedAt })}
      {...(onDraftOfferHandled === undefined ? {} : { onDraftOfferHandled })}
    />
  );
}
