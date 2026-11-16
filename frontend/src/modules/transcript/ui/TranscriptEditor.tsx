import { useEffect, useMemo, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import {
  clearAcceptedTranscriptSnapshot,
  maximumAcceptedTranscriptSources,
  type AcceptedTranscriptSource,
  type DetachedTranscriptSnapshot,
  type TranscriptDraftLine,
  type TranscriptSnapshot,
} from '../application/recovery';
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
  | 'track_rename_failed'
  | 'duplicate_language'
  | 'replacement_cancelled'
  | 'language_cancelled';

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
  readonly detachedSnapshots?: readonly DetachedTranscriptSnapshot[];
  readonly mountedHistoryKeys?: readonly string[];
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
  detachedSnapshots = [],
  mountedHistoryKeys = [],
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
  // Track identity/language is independent of a restored snapshot's token owner.
  const [editLanguage, setEditLanguage] = useState(language);
  const [draftUpdatedAt, setDraftUpdatedAt] = useState<string | undefined>(draftOffer?.updatedAt);
  const [status, setStatus] = useState<'idle' | 'saving' | 'saved'>('idle');
  const [handledDetached, setHandledDetached] = useState<readonly string[]>([]);
  const [refreshedDetached, setRefreshedDetached] = useState<
    ReadonlyMap<string, TranscriptSnapshot>
  >(() => new Map());
  const acceptedDetached = useRef<readonly (TranscriptSnapshot | AcceptedTranscriptSource)[]>([]);
  const ownedKeys = useRef(mountedHistoryKeys);
  useEffect(() => {
    ownedKeys.current = mountedHistoryKeys;
    acceptedDetached.current = acceptedDetached.current.filter(
      (source) => !mountedHistoryKeys.includes(source.key),
    );
  }, [mountedHistoryKeys]);
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
    setEditLanguage(language);
  }
  if (!dirty && editLanguage !== language) setEditLanguage(language);
  const saveLanguage = fixedLanguage ? language : currentLanguage;
  const identity = `${lessonId}${instanceId ? `-${instanceId}` : ''}`;

  const currentDraftUpdatedAt = draftUpdatedAt ?? suppliedDraftUpdatedAt;
  const baseline = useMemo(() => {
    const value = {
      language: saveLanguage,
      baseLanguage: base.language,
      expectedUpdatedAt: base.token,
      lines: base.segments.map(toDraft),
    };
    return currentDraftUpdatedAt === undefined
      ? value
      : { ...value, draftUpdatedAt: currentDraftUpdatedAt };
  }, [currentDraftUpdatedAt, saveLanguage, base]);
  // Read the ref at capture time, including unmounts before another render.
  const captureBaseline = {
    ...baseline,
    get acceptedSource() {
      return acceptedDetached.current.length === 0
        ? undefined
        : acceptedDetached.current.map(({ key, capturedAt }) => ({ key, capturedAt }));
    },
  };
  const recovery = useTranscriptRecovery({
    historyKey: history.key,
    store: history.store,
    baseline: captureBaseline,
  });

  const recoveryLocked = !recovery.isLoaded || recovery.offer.kind !== 'none';
  const readOnly = recoveryLocked || status === 'saving';
  const { captureNow, noteChange } = recovery;
  // Only author actions arm capture. Server/base updates must never flush a clean editor
  // over a still-unresolved recovery snapshot.
  const changeLines = (next: readonly DraftLine[]) => {
    setLines(next);
    noteChange(next);
  };

  // Leaving a field is the moment the author is most likely to leave the page next, so it is
  // where the snapshot is taken. The listener is attached to the DOM rather than written as a
  // JSX handler: `focusout` bubbles natively, so one listener covers every field without each of
  // them knowing about recovery, and a section is not an element that should carry a handler.
  const root = useRef<HTMLElement | null>(null);
  const focusAfterRecovery = useRef<Element | null>(null);
  const firstLineText = useRef<HTMLTextAreaElement | null>(null);
  useEffect(() => {
    const initiatingControl = focusAfterRecovery.current;
    if (initiatingControl === null) return;
    focusAfterRecovery.current = null;
    if (document.activeElement !== initiatingControl && document.activeElement !== document.body)
      return;
    const nextOffer = root.current?.querySelector<HTMLButtonElement>(
      '.transcript-editor-recovery button, .transcript-editor-draft-offer button',
    );
    const target =
      nextOffer ??
      firstLineText.current ??
      root.current?.querySelector<HTMLButtonElement>('[data-add-line]');
    target?.focus();
  });
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
      if (!recoveryLocked) captureNow(linesRef.current);
    };
    element.addEventListener('focusout', onFocusOut);
    return () => {
      element.removeEventListener('focusout', onFocusOut);
    };
  }, [captureNow, recoveryLocked]);

  const selectedDetached =
    recovery.isLoaded && recovery.offer.kind === 'none'
      ? detachedSnapshots.find((offer) => !handledDetached.includes(offer.snapshot.key))
      : undefined;
  const detached =
    selectedDetached === undefined
      ? undefined
      : {
          ...selectedDetached,
          snapshot:
            refreshedDetached.get(selectedDetached.snapshot.key) ?? selectedDetached.snapshot,
        };
  const offeredSnapshot =
    recovery.offer.kind === 'none' ? detached?.snapshot : recovery.offer.snapshot;
  const releaseAcceptedDetached = () => {
    const previous = acceptedDetached.current;
    acceptedDetached.current = [];
    setHandledDetached((current) =>
      current.filter((key) => !previous.some((source) => source.key === key)),
    );
  };
  const handleDetached = async (accepted: boolean, initiatingControl: Element) => {
    if (detached === undefined) return;
    const key = detached.snapshot.key;
    if (accepted) {
      releaseAcceptedDetached();
      const sources = [detached.snapshot, ...(detached.snapshot.acceptedSource ?? [])];
      // Keep newest restoration ownership first; overflow stays recoverable in its own key.
      acceptedDetached.current = sources
        .filter(
          (source, index) =>
            sources.findIndex(
              (other) => other.key === source.key && other.capturedAt === source.capturedAt,
            ) === index,
        )
        .slice(0, maximumAcceptedTranscriptSources);
    } else {
      try {
        if (!(await history.store.clearIfVersion(key, detached.snapshot.capturedAt))) {
          const latest = await history.store.read(key).catch(() => undefined);
          if (latest !== null) {
            focusAfterRecovery.current = initiatingControl;
            if (latest !== undefined)
              setRefreshedDetached((current) => new Map(current).set(key, latest));
            return; // A newer copy was not discarded or handled: offer it again.
          }
        }
      } catch {
        // Refused storage must not prevent continuing to the next offer.
      }
      focusAfterRecovery.current = initiatingControl;
    }
    setHandledDetached((current) => [
      ...current,
      ...(accepted ? acceptedDetached.current.map((source) => source.key) : [key]),
    ]);
  };

  const recoveryLabel =
    recovery.offer.kind === 'offer-stale'
      ? 'administration.transcriptEditor.recovery.staleHeading'
      : 'administration.transcriptEditor.recovery.heading';

  // Times are held as TEXT while editing and parsed only on save. Parsing on every keystroke turns
  // a half-typed "1" into a real value and reorders the lines under the author's cursor.
  const update = (index: number, change: Partial<DraftLine>) => {
    if (readOnly) return;
    setStatus('idle');
    changeLines(
      lines.map((line, position) => (position === index ? { ...line, ...change } : line)),
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

      {instanceId.startsWith('track:') && fixedLanguage && dirty && editLanguage !== language && (
        <p>{t('administration.transcriptEditor.languageChanged', { language })}</p>
      )}

      {offeredSnapshot === undefined ? null : (
        // Deliberately NOT an `output`, whose implicit role is `status`: this section already owns
        // one for "transcript saved", and a second would make both ambiguous to a screen reader
        // and to any test that asks the page for its status.
        <div
          className="transcript-editor-recovery"
          aria-labelledby={`transcript-recovery-${identity}`}
        >
          <h5 id={`transcript-recovery-${identity}`}>
            {detached === undefined
              ? t(recoveryLabel)
              : t(
                  `administration.transcriptEditor.recovery.${detached.legacy ? 'legacy' : 'detached'}${offeredSnapshot.language ? 'Language' : 'Unknown'}`,
                  { language: offeredSnapshot.language },
                )}
          </h5>
          <p>
            {t(
              recovery.offer.kind === 'offer-stale'
                ? 'administration.transcriptEditor.recovery.staleDescription'
                : 'administration.transcriptEditor.recovery.description',
            )}
          </p>
          <Button
            type="button"
            aria-disabled={status === 'saving'}
            onPress={(event) => {
              if (status === 'saving' || offeredSnapshot === undefined) return;
              focusAfterRecovery.current = event.target;
              setStatus('idle');
              changeLines(offeredSnapshot.lines);
              if (offeredSnapshot.language !== undefined) {
                if (!fixedLanguage) setCurrentLanguage(offeredSnapshot.language);
              }
              const snapshot = offeredSnapshot;
              setDraftUpdatedAt(snapshot.draftUpdatedAt);
              setBase((current) => ({
                ...current,
                // Empty string is not a real language tag: legacy tokens have no known owner.
                language: snapshot.baseLanguage ?? '',
                token: snapshot.expectedUpdatedAt,
              }));
              if (detached === undefined) {
                releaseAcceptedDetached();
                acceptedDetached.current = snapshot.acceptedSource ?? [];
                setHandledDetached((current) => [
                  ...current,
                  ...acceptedDetached.current.map((source) => source.key),
                ]);
                recovery.accept();
              } else void handleDetached(true, event.target);
            }}
          >
            {t('administration.transcriptEditor.recovery.restore')}
          </Button>
          <Button
            type="button"
            variant="secondary"
            aria-disabled={status === 'saving'}
            onPress={(event) => {
              if (status === 'saving') return;
              focusAfterRecovery.current = event.target;
              if (detached === undefined) {
                releaseAcceptedDetached();
                recovery.dismiss();
              } else void handleDetached(false, event.target);
            }}
          >
            {t('administration.transcriptEditor.recovery.discard')}
          </Button>
        </div>
      )}

      {offeredSnapshot === undefined && draftOffer !== undefined && (
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
            aria-disabled={readOnly}
            onPress={() => {
              if (readOnly) return;
              releaseAcceptedDetached();
              changeLines(draftOffer.segments.map(toDraft));
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

      {recoveryLocked && (
        <p id={`transcript-recovery-lock-${identity}`}>
          {t('administration.transcriptEditor.recovery.resolveFirst')}
        </p>
      )}

      {lines.length === 0 ? <p>{t('administration.transcriptEditor.empty')}</p> : null}

      <ol className="transcript-editor-lines">
        {lines.map((line, index) => (
          <li key={line.id}>
            <TextField
              isReadOnly={readOnly}
              {...(recoveryLocked
                ? { 'aria-describedby': `transcript-recovery-lock-${identity}` }
                : {})}
              label={t('administration.transcriptEditor.start', { line: index + 1 })}
              value={line.startMs}
              onChange={(value) => update(index, { startMs: value })}
            />
            <TextField
              isReadOnly={readOnly}
              {...(recoveryLocked
                ? { 'aria-describedby': `transcript-recovery-lock-${identity}` }
                : {})}
              label={t('administration.transcriptEditor.end', { line: index + 1 })}
              value={line.endMs}
              onChange={(value) => update(index, { endMs: value })}
            />
            <TextArea
              {...(index === 0 ? { ref: firstLineText } : {})}
              isReadOnly={readOnly}
              {...(recoveryLocked
                ? { 'aria-describedby': `transcript-recovery-lock-${identity}` }
                : {})}
              label={t('administration.transcriptEditor.text', { line: index + 1 })}
              value={line.text}
              onChange={(value) => update(index, { text: value })}
            />
            <fieldset>
              <legend>{t('administration.transcriptEditor.documents', { line: index + 1 })}</legend>
              {documents.map((document) => (
                <label key={document.id}>
                  <input
                    type="checkbox"
                    aria-disabled={readOnly}
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
              aria-disabled={readOnly}
              onPress={() => {
                if (readOnly) return;
                setStatus('idle');
                changeLines(lines.filter((_, at) => at !== index));
              }}
            >
              {t('administration.transcriptEditor.remove', { line: index + 1 })}
            </Button>
          </li>
        ))}
      </ol>

      <Button
        type="button"
        data-add-line
        variant="secondary"
        aria-disabled={readOnly}
        onPress={() => {
          if (readOnly) return;
          setStatus('idle');
          changeLines([
            ...lines,
            {
              id: crypto.randomUUID(),
              startMs: String(lines.at(-1)?.endMs ?? '0'),
              endMs: String(Number(lines.at(-1)?.endMs ?? '0') + 1000),
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
        aria-disabled={readOnly}
        onPress={async () => {
          if (readOnly) return;
          const segmentsToSave = parsed();
          if (segmentsToSave === null) {
            setFailure('invalid_request');
            return;
          }
          recovery.setCapturing(false);
          setStatus('saving');
          setFailure(null);
          let error: TranscriptSaveErrorCode | null;
          try {
            error = await actions.save(saveLanguage, base.token, segmentsToSave, base.language);
          } catch {
            error = 'unavailable';
          }
          recovery.setCapturing(true);
          setStatus(error === null ? 'saved' : 'idle');
          setFailure(error);
          // Only a save the server accepted drops the local copy. A reload must never do it:
          // that is the one case the copy exists for.
          if (error === null) {
            setEditLanguage(saveLanguage);
            recovery.finalize();
            for (const accepted of acceptedDetached.current)
              void clearAcceptedTranscriptSnapshot(history.store, accepted, () =>
                ownedKeys.current.includes(accepted.key),
              );
            acceptedDetached.current = [];
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
