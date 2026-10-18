import type { TranscriptSegment } from '../domain/transcript';

export type TranscriptDraftOffer = {
  readonly segments: readonly TranscriptSegment[];
  readonly language: string;
  readonly updatedAt: string | undefined;
};
