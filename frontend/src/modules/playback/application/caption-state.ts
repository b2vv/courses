import type { CaptionSelection } from '../ports/PlaybackMedia';

export type CaptionSelectionEvent = {
  readonly mount: symbol;
  readonly generation: number;
  readonly selection: CaptionSelection;
};

export function reduceCaptionSelection(
  mount: symbol,
  record: CaptionSelectionEvent | null,
  event: CaptionSelectionEvent,
): CaptionSelectionEvent | null {
  if (event.mount !== mount) return record;
  if (record?.mount === mount && event.generation < record.generation) return record;
  return event;
}

export type CaptionQuery =
  | { readonly mode: 'off'; readonly panel: 'hidden' | 'loading' | 'none-for-track' }
  | { readonly mode: 'primary' }
  | { readonly mode: 'exact'; readonly language: string; readonly pending?: true };

/** A record for this mount means load happened: prediction can never become authoritative again. */
export function captionQuery({
  locked,
  emailGated,
  ready,
  mount,
  record,
  predictedLanguage,
}: {
  readonly locked: boolean;
  readonly emailGated: boolean;
  readonly ready: boolean;
  readonly mount: symbol;
  readonly record: CaptionSelectionEvent | null;
  readonly predictedLanguage: string | null;
}): CaptionQuery {
  if (locked || emailGated) return { mode: 'off', panel: 'hidden' };
  if (!ready) return { mode: 'primary' };
  const selection = record?.mount === mount ? record.selection : null;
  if (selection?.state === 'unavailable') return { mode: 'off', panel: 'none-for-track' };
  const language = (
    selection?.state === 'pending'
      ? selection.requested
      : selection?.state === 'confirmed'
        ? selection.language
        : predictedLanguage
  )?.toLowerCase();
  if (!language) return { mode: 'off', panel: 'loading' };
  if (language === 'und' || language.startsWith('und-'))
    return { mode: 'off', panel: 'none-for-track' };
  return {
    mode: 'exact',
    language,
    ...(selection?.state === 'pending' ? { pending: true } : {}),
  };
}
