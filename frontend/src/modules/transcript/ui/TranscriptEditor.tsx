import { useEffect, useMemo, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import type { TranscriptDraftLine } from '../application/recovery';
import type {
  TranscriptDocument,
  TranscriptErrorCode,
  TranscriptSegment,
} from '../domain/transcript';
import type { TranscriptHistoryStore } from '../ports/TranscriptHistoryStore';
import { Button, TextArea, TextField } from '../../../presentation/ui';
import { useTranscriptRecovery } from './use-transcript-recovery';
import type { TranscriptDraftOffer } from './transcript-draft-offer';

type DraftLine = TranscriptDraftLine;

export type TranscriptSaveErrorCode =
  | TranscriptErrorCode
  | 'track_verification_failed'
  | 'track_changed'
  | 'track_language_required'
  | 'replacement_cancelled';

export type TranscriptEditorActions = {
  save(
    language: string,
    expectedUpdatedAt: string | undefined,
    segments: readonly TranscriptSegment[],
    baseLanguage?: string,
  ): Promise<TranscriptSaveErrorCode | null>;
};

type TranscriptEditorProps = {
  readonly lessonId: string;
  readonly instanceId?: string;
  readonly fixedLanguage?: boolean;
  readonly language: string;
  readonly baseLanguage?: string;
  /** Whether the server returned a saved transcript; absent means a first transcript. */
  readonly hasSavedTranscript?: boolean;
  readonly expectedUpdatedAt?: string;
  readonly segments: readonly TranscriptSegment[];
  readonly documents: readonly TranscriptDocument[];
  readonly actions: TranscriptEditorActions;
  /**
   * Where unsaved lines are kept between visits, and under which key. Required rather than
   * optional so that presentation never reaches for a store itself: the composition root decides
   * whether that is IndexedDB or the in-memory fallback.
   */
  readonly history: {
    readonly key: string;
    readonly store: TranscriptHistoryStore;
  };
  /**
   * An ASR draft the author asked to load, offered through the SAME accept/dismiss affordance
   * as crash recovery rather than applied silently: both are "text that is not yet the saved
   * transcript", and the author must not lose one kind of unsaved text to the other. Shown only
   * once `recovery.offer` is `none` — an unsaved edit already on screen outranks a draft the
   * author can ask for again.
   */
  readonly draftOffer?: TranscriptDraftOffer;
  readonly draftUpdatedAt?: string;
  /** The draft offer was accepted or dismissed: the composition root drops it so a remount (or a
   * fresh regeneration) does not resurface a handled offer. */
  readonly onDraftOfferHandled?: () => void;
};

function toDraft(segment: TranscriptSegment): DraftLine {
  return {
    id: segment.id,
    startMs: String(segment.startMs),
    endMs: String(segment.endMs),
    text: segment.text,
    documentIds: segment.documentIds,
  };
}

export function TranscriptEditor({
  lessonId,
  instanceId = '',
  fixedLanguage = false,
  language,
  baseLanguage = language,
  expectedUpdatedAt,
  segments,
  documents,
  actions,
  history,
  draftOffer,
  draftUpdatedAt: suppliedDraftUpdatedAt,
  onDraftOfferHandled,
}: TranscriptEditorProps) {
  const { t } = useTranslation();
  const [lines, setLines] = useState<readonly DraftLine[]>(() => segments.map(toDraft));
  const [currentLanguage, setCurrentLanguage] = useState(language);
  const [draftUpdatedAt, setDraftUpdatedAt] = useState<string | undefined>(draftOffer?.updatedAt);
  const [status, setStatus] = useState<'idle' | 'saving' | 'saved'>('idle');
  const [failure, setFailure] = useState<TranscriptSaveErrorCode | null>(null);

  const [base, setBase] = useState(() => ({
    language: baseLanguage,
    token: expectedUpdatedAt,
    segments,
  }));
  const dirty =
    status !== 'saved' && JSON.stringify(lines) !== JSON.stringify(base.segments.map(toDraft));
  // Adjust only a clean editor when a new server base arrives. Keeping this conditional
  // avoids an effect render with old lines under a new language, and preserves dirty tokens.
  if (
    !dirty &&
    (base.language !== baseLanguage ||
      base.token !== expectedUpdatedAt ||
      base.segments !== segments)
  ) {
    setBase({ language: baseLanguage, token: expectedUpdatedAt, segments });
    setLines(segments.map(toDraft));
    setCurrentLanguage(language);
  }
  const saveLanguage = fixedLanguage ? language : currentLanguage;
  const identity = `${lessonId}${instanceId ? `-${instanceId}` : ''}`;

  const currentDraftUpdatedAt = draftUpdatedAt ?? suppliedDraftUpdatedAt;
  const baseline = useMemo(() => {
    const value = {
      language: saveLanguage,
      expectedUpdatedAt: base.token,
      lines: base.segments.map(toDraft),
    };
    return currentDraftUpdatedAt === undefined
      ? value
      : { ...value, draftUpdatedAt: currentDraftUpdatedAt };
  }, [currentDraftUpdatedAt, saveLanguage, base]);
  const recovery = useTranscriptRecovery({
    historyKey: history.key,
    store: history.store,
    baseline,
  });

  // The idle capture is armed from an effect rather than from inside the state updater: React
  // may invoke an updater twice, and a store write is not something to do twice for one keystroke.
  const { captureNow, noteChange } = recovery;
  useEffect(() => {
    noteChange(lines);
  }, [lines, noteChange]);

  // Leaving a field is the moment the author is most likely to leave the page next, so it is
  // where the snapshot is taken. The listener is attached to the DOM rather than written as a
  // JSX handler: `focusout` bubbles natively, so one listener covers every field without each of
  // them knowing about recovery, and a section is not an element that should carry a handler.
  const root = useRef<HTMLElement | null>(null);
  // Mirrored in a ref, and written from an effect rather than during render, so that the listener
  // below can stay registered once instead of being torn down and rebuilt on every keystroke.
  const linesRef = useRef(lines);
  useEffect(() => {
    linesRef.current = lines;
  }, [lines]);
  useEffect(() => {
    const element = root.current;
    if (element === null) return undefined;
    const onFocusOut = () => {
      captureNow(linesRef.current);
    };
    element.addEventListener('focusout', onFocusOut);
    return () => {
      element.removeEventListener('focusout', onFocusOut);
    };
  }, [captureNow]);

  const recoveryLabel =
    recovery.offer.kind === 'offer-stale'
      ? 'administration.transcriptEditor.recovery.staleHeading'
      : 'administration.transcriptEditor.recovery.heading';

  // Times are held as TEXT while editing and parsed only on save. Parsing on every keystroke turns
  // a half-typed "1" into a real value and reorders the lines under the author's cursor.
  const update = (index: number, change: Partial<DraftLine>) => {
    setStatus('idle');
    setLines((current) =>
      current.map((line, position) => (position === index ? { ...line, ...change } : line)),
    );
  };

  const parsed = (): readonly TranscriptSegment[] | null => {
    const result: TranscriptSegment[] = [];
    for (const line of lines) {
      const startMs = Number(line.startMs);
      const endMs = Number(line.endMs);
      if (!Number.isInteger(startMs) || !Number.isInteger(endMs)) return null;
      result.push({
        id: crypto.randomUUID(),
        startMs,
        endMs,
        text: line.text,
        documentIds: line.documentIds,
      });
    }
    return result;
  };

  return (
    <section
      ref={root}
      className="transcript-editor"
      aria-labelledby={`transcript-editor-${identity}`}
    >
      <h4 id={`transcript-editor-${identity}`}>{t('administration.transcriptEditor.heading')}</h4>

      {fixedLanguage && dirty && base.language !== language && (
        <p>{t('administration.transcriptEditor.languageChanged', { language })}</p>
      )}

      {recovery.offer.kind === 'none' ? null : (
        // Deliberately NOT an `output`, whose implicit role is `status`: this section already owns
        // one for "transcript saved", and a second would make both ambiguous to a screen reader
        // and to any test that asks the page for its status.
        <div
          className="transcript-editor-recovery"
          aria-labelledby={`transcript-recovery-${identity}`}
        >
          <h5 id={`transcript-recovery-${identity}`}>{t(recoveryLabel)}</h5>
          <p>
            {t(
              recovery.offer.kind === 'offer-stale'
                ? 'administration.transcriptEditor.recovery.staleDescription'
                : 'administration.transcriptEditor.recovery.description',
            )}
          </p>
          <Button
            type="button"
            isDisabled={status === 'saving'}
            onPress={() => {
              if (recovery.offer.kind === 'none') return;
              setStatus('idle');
              setLines(recovery.offer.snapshot.lines);
              if (recovery.offer.snapshot.language !== undefined) {
                if (!fixedLanguage) setCurrentLanguage(recovery.offer.snapshot.language);
              }
              setBase((current) => ({
                ...current,
                token:
                  recovery.offer.kind === 'none'
                    ? current.token
                    : recovery.offer.snapshot.expectedUpdatedAt,
              }));
              recovery.accept();
            }}
          >
            {t('administration.transcriptEditor.recovery.restore')}
          </Button>
          <Button type="button" variant="secondary" onPress={() => recovery.dismiss()}>
            {t('administration.transcriptEditor.recovery.discard')}
          </Button>
        </div>
      )}

      {recovery.offer.kind === 'none' && draftOffer !== undefined && (
        <div
          className="transcript-editor-draft-offer"
          aria-labelledby={`transcript-draft-offer-${identity}`}
        >
          <h5 id={`transcript-draft-offer-${identity}`}>
            {t('administration.transcriptEditor.draftOffer.heading')}
          </h5>
          <p>{t('administration.transcriptEditor.draftOffer.description')}</p>
          <Button
            type="button"
            isDisabled={status === 'saving'}
            onPress={() => {
              setLines(draftOffer.segments.map(toDraft));
              if (!fixedLanguage) setCurrentLanguage(draftOffer.language);
              setDraftUpdatedAt(draftOffer.updatedAt);
              setStatus('idle');
              onDraftOfferHandled?.();
            }}
          >
            {t('administration.transcriptEditor.draftOffer.load')}
          </Button>
          <Button type="button" variant="secondary" onPress={() => onDraftOfferHandled?.()}>
            {t('administration.transcriptEditor.draftOffer.dismiss')}
          </Button>
        </div>
      )}

      {lines.length === 0 ? <p>{t('administration.transcriptEditor.empty')}</p> : null}

      <ol className="transcript-editor-lines">
        {lines.map((line, index) => (
          <li key={line.id}>
            <TextField
              isDisabled={status === 'saving'}
              label={t('administration.transcriptEditor.start', { line: index + 1 })}
              value={line.startMs}
              onChange={(value) => update(index, { startMs: value })}
            />
            <TextField
              isDisabled={status === 'saving'}
              label={t('administration.transcriptEditor.end', { line: index + 1 })}
              value={line.endMs}
              onChange={(value) => update(index, { endMs: value })}
            />
            <TextArea
              isDisabled={status === 'saving'}
              label={t('administration.transcriptEditor.text', { line: index + 1 })}
              value={line.text}
              onChange={(value) => update(index, { text: value })}
            />
            <fieldset disabled={status === 'saving'}>
              <legend>{t('administration.transcriptEditor.documents', { line: index + 1 })}</legend>
              {documents.map((document) => (
                <label key={document.id}>
                  <input
                    type="checkbox"
                    checked={line.documentIds.includes(document.id)}
                    onChange={(event) =>
                      update(index, {
                        documentIds: event.target.checked
                          ? [...line.documentIds, document.id]
                          : line.documentIds.filter((id) => id !== document.id),
                      })
                    }
                  />
                  {document.label}
                </label>
              ))}
            </fieldset>
            <Button
              type="button"
              variant="secondary"
              isDisabled={status === 'saving'}
              onPress={() => {
                setStatus('idle');
                setLines((current) => current.filter((_, at) => at !== index));
              }}
            >
              {t('administration.transcriptEditor.remove', { line: index + 1 })}
            </Button>
          </li>
        ))}
      </ol>

      <Button
        type="button"
        variant="secondary"
        isDisabled={status === 'saving'}
        onPress={() => {
          setStatus('idle');
          setLines((current) => [
            ...current,
            {
              id: crypto.randomUUID(),
              startMs: String(current.at(-1)?.endMs ?? '0'),
              endMs: String(Number(current.at(-1)?.endMs ?? '0') + 1000),
              text: '',
              documentIds: [],
            },
          ]);
        }}
      >
        {t('administration.transcriptEditor.add')}
      </Button>

      <Button
        type="button"
        isDisabled={status === 'saving'}
        onPress={async () => {
          const segmentsToSave = parsed();
          if (segmentsToSave === null) {
            setFailure('invalid_request');
            return;
          }
          setStatus('saving');
          setFailure(null);
          let error: TranscriptSaveErrorCode | null;
          try {
            error = await actions.save(saveLanguage, base.token, segmentsToSave, base.language);
          } catch {
            error = 'unavailable';
          }
          setStatus(error === null ? 'saved' : 'idle');
          setFailure(error);
          // Only a save the server accepted drops the local copy. A reload must never do it:
          // that is the one case the copy exists for.
          if (error === null) {
            recovery.finalize();
          }
        }}
      >
        {t('administration.transcriptEditor.save')}
      </Button>

      {failure === null ? null : (
        <output className="transcript-editor-error">
          {t(`administration.transcriptEditor.errors.${failure}`, {
            defaultValue: t('administration.transcriptEditor.errors.unavailable'),
          })}
        </output>
      )}
      {status === 'saved' && failure === null ? (
        <output>{t('administration.transcriptEditor.saved')}</output>
      ) : null}
    </section>
  );
}
