import type { LessonTranscript } from '../domain/transcript';

export type TranscriptPanelState =
  | { readonly kind: 'loading' }
  | { readonly kind: 'none-for-track' }
  | { readonly kind: 'none' }
  | { readonly kind: 'error'; readonly onRetry: () => void }
  | { readonly kind: 'transcript'; readonly transcript: LessonTranscript };

export function transcriptPanelState({
  language,
  data,
  isPending,
  isError,
  onRetry,
}: {
  /** Undefined is the primary read for a lesson without Ready media. */
  readonly language: string | undefined;
  readonly data: LessonTranscript | undefined;
  readonly isPending: boolean;
  readonly isError: boolean;
  readonly onRetry: () => void;
}): TranscriptPanelState {
  if (isPending) return { kind: 'loading' };
  if (isError) return { kind: 'error', onRetry };
  if (
    data === undefined ||
    data.language === null ||
    data.segments.length === 0 ||
    (language !== undefined && data.language.toLowerCase() !== language.toLowerCase())
  )
    return { kind: language === undefined ? 'none' : 'none-for-track' };
  return { kind: 'transcript', transcript: data };
}
