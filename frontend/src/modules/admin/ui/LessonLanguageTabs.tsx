import { useCallback, useEffect, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import * as Aria from 'react-aria-components';
import { Button, Select } from '../../../presentation/ui';
import { audioTrackLanguages } from '../../media';
import type { AudioTrack, CourseEditorVideo } from '../../media';
import {
  isUndetermined,
  LessonDraftField,
  LessonTranscriptField,
  useDetachedTranscriptRecovery,
  transcriptHistoryKey,
} from '../../transcript';
import type {
  DetachedTranscriptSnapshot,
  CourseEditorDrafts,
  CourseEditorTranscripts,
  TranscriptDraftOffer,
  TranscriptInventoryItem,
  TranscriptEditorActions,
} from '../../transcript';

export type LessonLanguageServices = {
  useInventory(lessonId: string): {
    data: readonly TranscriptInventoryItem[] | undefined;
    isError?: boolean;
  };
  readInventory(lessonId: string): Promise<readonly TranscriptInventoryItem[]>;
  readTracks(videoId: string): Promise<readonly AudioTrack[]>;
};

type LanguageTab = { id: string; language: string; track?: AudioTrack };

export function lessonLanguageTabs(
  tracks: readonly AudioTrack[],
  inventory: readonly TranscriptInventoryItem[],
): LanguageTab[] {
  const ready = tracks
    .filter((track) => track.state === 'ready')
    .sort(
      (a, b) =>
        Number(b.isDefault) - Number(a.isDefault) ||
        Number(b.kind === 'embedded') - Number(a.kind === 'embedded') ||
        a.language.toLowerCase().localeCompare(b.language.toLowerCase()) ||
        a.trackId.localeCompare(b.trackId),
    );
  const languages = new Set(ready.map((track) => track.language.toLowerCase()));
  return [
    ...ready.map((track) => ({
      id: `track:${track.trackId}`,
      language: track.language.toLowerCase(),
      track,
    })),
    ...inventory
      .filter((item) => !languages.has(item.language.toLowerCase()))
      .sort((a, b) => a.language.localeCompare(b.language))
      .map((item) => ({
        id: `lang:${item.language.toLowerCase()}`,
        language: item.language.toLowerCase(),
      })),
  ];
}

type Props = {
  lessonId: string;
  video?: CourseEditorVideo | undefined;
  drafts?: CourseEditorDrafts | undefined;
  transcripts?: CourseEditorTranscripts | undefined;
  languages?: LessonLanguageServices | undefined;
  onTranscriptKnown(lessonId: string, known: boolean): void;
};

export function LessonLanguageTabs(props: Props) {
  const ready = props.video?.rows
    .filter((row) => row.lessonId === props.lessonId && row.state === 'ready')
    .sort((a, b) => b.generation - a.generation)[0];
  if (ready === undefined || props.video?.audio === undefined || props.languages === undefined)
    return (
      <>
        {props.languages !== undefined && (
          <TranscriptKnown
            lessonId={props.lessonId}
            languages={props.languages}
            onTranscriptKnown={props.onTranscriptKnown}
          />
        )}
        <SingleLanguageEditor
          {...props}
          onTranscriptKnown={props.languages === undefined ? props.onTranscriptKnown : ignoreKnown}
        />
      </>
    );
  return (
    <ReadyLanguageTabs
      {...props}
      videoId={ready.videoId}
      audio={props.video.audio}
      languages={props.languages}
    />
  );
}

function SingleLanguageEditor(props: Props) {
  const detachedSnapshots = useDetachedTranscriptRecovery(
    props.lessonId,
    props.transcripts?.history,
    [],
  );
  return <LanguageEditor {...props} detachedSnapshots={detachedSnapshots} />;
}

const ignoreKnown = () => {};

function TranscriptKnown({
  lessonId,
  languages,
  onTranscriptKnown,
}: {
  lessonId: string;
  languages: LessonLanguageServices;
  onTranscriptKnown: Props['onTranscriptKnown'];
}) {
  const { data } = languages.useInventory(lessonId);
  useEffect(() => {
    if (data !== undefined) onTranscriptKnown(lessonId, data.length > 0);
  }, [data, lessonId, onTranscriptKnown]);
  return null;
}

function ReadyLanguageTabs({
  videoId,
  audio,
  languages,
  ...props
}: Props & {
  videoId: string;
  audio: NonNullable<CourseEditorVideo['audio']>;
  languages: LessonLanguageServices;
}) {
  const { t } = useTranslation();
  const tracks = audio.useTracks(videoId);
  const inventory = languages.useInventory(props.lessonId);
  const { lessonId, onTranscriptKnown } = props;
  useEffect(() => {
    if (inventory.data !== undefined) onTranscriptKnown(lessonId, inventory.data.length > 0);
  }, [inventory.data, lessonId, onTranscriptKnown]);
  // The tab set is computed only from BOTH lists. With the tracks still loading, a transcript
  // whose language belongs to a track looks like an orphan: React Aria selects that orphan tab
  // first and keeps it selected after the tracks arrive, and its editor mounts under the wrong
  // recovery key. Nothing while loading, as `LessonTranscriptField`, for the same status-region
  // reason.
  // A failed background refresh keeps the cached lists: unmounting the editors here would drop
  // unsaved lines. The alert replaces the tabs only when there is nothing to show yet.
  const tabs = lessonLanguageTabs(tracks.data ?? [], inventory.data ?? []);
  const detached = useDetachedTranscriptRecovery(
    lessonId,
    props.transcripts?.history,
    tabs.map((tab) => ({
      kind: tab.track === undefined ? 'orphan' : 'track',
      instanceId: tab.id,
      language: tab.language,
      isDefault: tab.track?.isDefault ?? false,
    })),
    tracks.data !== undefined && inventory.data !== undefined,
  );
  const loadFailed = tracks.error !== undefined || inventory.isError === true;
  const loadFailedAlert = loadFailed ? (
    <p className="admin-video-error" role="alert">
      {t('administration.languageTabs.loadFailed')}
    </p>
  ) : null;
  if (tracks.data === undefined || inventory.data === undefined) return loadFailedAlert;
  function label(tab: LanguageTab) {
    const known = audioTrackLanguages.find((code) => code === tab.language);
    const name = isUndetermined(tab.language)
      ? t('administration.languageTabs.unknown')
      : known === undefined
        ? tab.language
        : t(`administration.audio.language.${known}`);
    return `${name}${tab.track?.isDefault ? ` · ${t('administration.audio.default')}` : ''}${tab.track?.multilingual ? ` · ${t('administration.audio.multilingual')}` : ''}`;
  }
  return (
    <>
      {loadFailedAlert}
      <Aria.Tabs className="ui-tabs">
        <Aria.TabList aria-label={t('administration.languageTabs.label')} className="ui-tab-list">
          {tabs.map((tab) => (
            <Aria.Tab key={tab.id} id={tab.id} className="ui-tab">
              {label(tab)}
            </Aria.Tab>
          ))}
        </Aria.TabList>
        {tabs.map((tab) => (
          <Aria.TabPanel
            key={tab.id}
            id={tab.id}
            shouldForceMount
            className="lesson-language-panel ui-tab-panel"
          >
            <LanguageEditor
              {...props}
              languages={languages}
              tab={tab}
              mountedHistoryKeys={tabs.map((mounted) =>
                transcriptHistoryKey({
                  userId: props.transcripts?.history.authorId ?? '',
                  lessonId,
                  instanceId: mounted.id,
                }),
              )}
              detachedSnapshots={detached.filter((offer) => offer.instanceId === tab.id)}
              videoId={videoId}
              audio={audio}
            />
          </Aria.TabPanel>
        ))}
      </Aria.Tabs>
    </>
  );
}

function LanguageEditor({
  lessonId,
  drafts,
  transcripts,
  onTranscriptKnown,
  languages,
  tab,
  videoId,
  audio,
  detachedSnapshots,
  mountedHistoryKeys,
}: Props & {
  detachedSnapshots?: readonly DetachedTranscriptSnapshot[];
  mountedHistoryKeys?: readonly string[];
  audio?: NonNullable<CourseEditorVideo['audio']>;
  tab?: LanguageTab;
  videoId?: string;
}) {
  const { t, i18n } = useTranslation();
  const [offer, setOffer] = useState<TranscriptDraftOffer>();
  const invalidateOffer = useCallback(() => setOffer(undefined), []);
  const [draftToken, setDraftToken] = useState<string>();
  const draft = drafts?.useLessonDraft(lessonId, tab?.track?.trackId).data;
  const [languageChoice, setLanguageChoice] = useState<{
    options: readonly string[];
    selected: string;
    required: boolean;
  }>();
  const resolveLanguage = useRef<((language: string | undefined) => void) | undefined>(undefined);
  useEffect(() => () => resolveLanguage.current?.(undefined), []);
  function finishLanguage(language: string | undefined) {
    resolveLanguage.current?.(language);
    resolveLanguage.current = undefined;
    setLanguageChoice(undefined);
  }
  const [replacement, setReplacement] = useState<TranscriptInventoryItem>();
  const resolveReplacement = useRef<((confirmed: boolean) => void) | undefined>(undefined);
  const saveTrigger = useRef<HTMLElement | undefined>(undefined);
  useEffect(() => {
    if (languageChoice !== undefined || replacement !== undefined) return;
    // The async inventory read separates Save from opening the overlay. Bind return focus
    // to Save itself and restore after overlay teardown; FocusScope otherwise waits for RAF.
    const trigger = saveTrigger.current;
    saveTrigger.current = undefined;
    if (trigger?.isConnected) trigger.focus({ preventScroll: true });
  }, [languageChoice, replacement]);
  useEffect(() => () => resolveReplacement.current?.(false), []);
  function finishReplacement(confirmed: boolean) {
    resolveReplacement.current?.(confirmed);
    resolveReplacement.current = undefined;
    setReplacement(undefined);
  }
  const save: TranscriptEditorActions['save'] = async (
    language,
    baseToken,
    segments,
    baseLanguage,
  ) => {
    saveTrigger.current =
      document.activeElement instanceof HTMLElement ? document.activeElement : undefined;
    if (transcripts === undefined) return 'unavailable';
    if (tab === undefined || languages === undefined)
      return transcripts.save(lessonId, language, baseToken, segments);
    const trackMatches = (tracks: readonly AudioTrack[] | undefined) =>
      tab.track === undefined ||
      (tracks?.some(
        (track) =>
          track.trackId === tab.track?.trackId &&
          track.state === 'ready' &&
          track.language.toLowerCase() === language,
      ) ??
        false);
    let inventory: readonly TranscriptInventoryItem[];
    let freshTracks: readonly AudioTrack[] | undefined;
    try {
      const [freshInventory, tracks] = await Promise.all([
        languages.readInventory(lessonId),
        tab.track === undefined ? Promise.resolve(undefined) : languages.readTracks(videoId ?? ''),
      ]);
      inventory = freshInventory;
      freshTracks = tracks;
      if (!trackMatches(tracks)) return 'track_changed';
    } catch {
      return 'track_verification_failed';
    }
    if (isUndetermined(language)) {
      const track = freshTracks?.find((track) => track.trackId === tab.track?.trackId);
      if (track === undefined || audio === undefined) return 'track_changed';
      const taken = new Set([
        ...(freshTracks ?? []).map((track) => track.language.toLowerCase()),
        ...inventory.map((item) => item.language.toLowerCase()),
      ]);
      const options = audioTrackLanguages.filter((code) => code !== 'und' && !taken.has(code));
      const detected = draft?.language?.toLowerCase();
      const chosen = await new Promise<string | undefined>((resolve) => {
        resolveLanguage.current = resolve;
        setLanguageChoice({
          options,
          selected: options.find((code) => code === detected) ?? '',
          required: false,
        });
      });
      if (chosen === undefined) return 'language_cancelled';
      try {
        const error = await audio.actions.rename(track, chosen, track.label, track.multilingual);
        if (error !== null) return error === 'duplicate_language' ? error : 'track_rename_failed';
      } catch {
        return 'track_rename_failed';
      }
      return transcripts.save(lessonId, chosen, undefined, segments);
    }
    let token = baseToken;
    if (language !== baseLanguage) {
      const target = inventory.find((item) => item.language.toLowerCase() === language);
      token = target?.updatedAt;
      if (target !== undefined) {
        const confirmed = await new Promise<boolean>((resolve) => {
          resolveReplacement.current = resolve;
          setReplacement(target);
        });
        if (!confirmed) return 'replacement_cancelled';
        // The dialog waits for a person, far longer than the one-request race window the plan
        // accepts: the track may have been renamed or deleted meanwhile. Check it again; the token
        // stays the one the author confirmed.
        if (tab.track !== undefined) {
          try {
            if (!trackMatches(await languages.readTracks(videoId ?? ''))) return 'track_changed';
          } catch {
            return 'track_verification_failed';
          }
        }
      }
    }
    return transcripts.save(lessonId, language, token, segments);
  };
  return (
    <>
      {drafts !== undefined && (tab === undefined || tab.track !== undefined) && (
        <LessonDraftField
          {...drafts}
          lessonId={lessonId}
          instanceId={tab?.id ?? ''}
          {...(tab?.track === undefined ? {} : { track: tab.track })}
          onLoadDraft={(segments, language, updatedAt) =>
            setOffer({
              segments: segments.map((segment) => ({ ...segment, documentIds: [] })),
              language,
              updatedAt,
            })
          }
          onDraftTokenChange={setDraftToken}
          onDraftInvalidated={invalidateOffer}
        />
      )}
      {transcripts !== undefined && (
        <LessonTranscriptField
          {...transcripts}
          {...(detachedSnapshots === undefined ? {} : { detachedSnapshots })}
          {...(mountedHistoryKeys === undefined ? {} : { mountedHistoryKeys })}
          lessonId={lessonId}
          instanceId={tab?.id ?? ''}
          {...(tab === undefined
            ? { onTranscriptKnown }
            : { fixedLanguage: tab.language, editorSave: save })}
          {...(offer === undefined ? {} : { draftOffer: offer })}
          {...(draftToken === undefined ? {} : { draftUpdatedAt: draftToken })}
          onDraftOfferHandled={() => setOffer(undefined)}
        />
      )}
      <Aria.ModalOverlay
        isOpen={languageChoice !== undefined}
        onOpenChange={(open) => {
          if (!open) finishLanguage(undefined);
        }}
        isDismissable
        className="ui-modal-overlay"
      >
        <Aria.Modal className="ui-modal">
          <Aria.Dialog className="ui-dialog">
            <Aria.Heading slot="title">
              {t('administration.languageTabs.chooseHeading')}
            </Aria.Heading>
            <p>{t('administration.languageTabs.attachHint')}</p>
            {languageChoice !== undefined && languageChoice.options.length > 0 ? (
              <>
                <Select
                  label={t('administration.languageTabs.chooseLanguage')}
                  placeholder={t('administration.languageTabs.languagePlaceholder')}
                  options={languageChoice.options.map((code) => ({
                    id: code,
                    label: t(`administration.audio.language.${code}`),
                  }))}
                  value={languageChoice.selected || null}
                  onChange={(key) =>
                    setLanguageChoice({
                      ...languageChoice,
                      selected: String(key ?? ''),
                      required: false,
                    })
                  }
                />
                {languageChoice.required && (
                  <p role="alert">{t('administration.languageTabs.languageRequired')}</p>
                )}
                <Button
                  type="button"
                  onPress={() => {
                    if (!languageChoice.options.includes(languageChoice.selected)) {
                      setLanguageChoice({ ...languageChoice, required: true });
                      return;
                    }
                    finishLanguage(languageChoice.selected);
                  }}
                >
                  {t('administration.languageTabs.saveLanguage')}
                </Button>
              </>
            ) : (
              <p>{t('administration.languageTabs.noLanguages')}</p>
            )}
            <Button type="button" variant="secondary" onPress={() => finishLanguage(undefined)}>
              {t('administration.languageTabs.cancel')}
            </Button>
          </Aria.Dialog>
        </Aria.Modal>
      </Aria.ModalOverlay>
      <Aria.ModalOverlay
        isOpen={replacement !== undefined}
        onOpenChange={(open) => {
          if (!open) finishReplacement(false);
        }}
        isDismissable
        className="ui-modal-overlay"
      >
        <Aria.Modal className="ui-modal">
          <Aria.Dialog className="ui-dialog">
            <Aria.Heading slot="title">
              {t('administration.languageTabs.replace', {
                language: audioTrackLanguages.some(
                  (code) => code === replacement?.language.toLowerCase(),
                )
                  ? t(`administration.audio.language.${replacement?.language.toLowerCase()}`)
                  : replacement?.language,
                time:
                  replacement === undefined
                    ? ''
                    : Number.isNaN(Date.parse(replacement.updatedAt))
                      ? replacement.updatedAt
                      : new Intl.DateTimeFormat(i18n.language, {
                          dateStyle: 'medium',
                          timeStyle: 'short',
                        }).format(new Date(replacement.updatedAt)),
              })}
            </Aria.Heading>
            <Button type="button" onPress={() => finishReplacement(true)}>
              {t('administration.languageTabs.confirm')}
            </Button>
            <Button type="button" variant="secondary" onPress={() => finishReplacement(false)}>
              {t('administration.languageTabs.cancel')}
            </Button>
          </Aria.Dialog>
        </Aria.Modal>
      </Aria.ModalOverlay>
    </>
  );
}
