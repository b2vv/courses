import { useTranslation } from 'react-i18next';
import type {
  LessonTranscript,
  TranscriptDocument,
  TranscriptErrorCode,
  TranscriptSegment,
} from '../../domain/transcript/transcript';
import { Skeleton } from '../ui';
import { TranscriptEditor } from './TranscriptEditor';

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
  readonly save: (
    lessonId: string,
    language: string,
    expectedUpdatedAt: string | undefined,
    segments: readonly TranscriptSegment[],
  ) => Promise<TranscriptErrorCode | null>;
};

type LessonTranscriptFieldProps = CourseEditorTranscripts & {
  readonly lessonId: string;
};

export function LessonTranscriptField({
  lessonId,
  useLessonTranscript,
  documents,
  save,
}: LessonTranscriptFieldProps) {
  const { t } = useTranslation();
  const { data, isLoading } = useLessonTranscript(lessonId);

  if (isLoading) return <Skeleton label={t('administration.transcriptEditor.heading')} />;

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
      actions={{
        save: (nextLanguage, token, segments) => save(lessonId, nextLanguage, token, segments),
      }}
    />
  );
}
