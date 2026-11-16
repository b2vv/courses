import { useCallback, useEffect, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import * as Aria from 'react-aria-components';
import { Button } from '../../../presentation/ui';
import { audioTrackLanguages } from '../../media';
import type { AudioTrack, CourseEditorVideo } from '../../media';
import { isUndetermined, LessonDraftField, LessonTranscriptField } from '../../transcript';
import type {
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
        <LanguageEditor
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
  const loadFailed = tracks.error !== undefined || inventory.isError === true;
  const loadFailedAlert = loadFailed ? (
    <p className="admin-video-error" role="alert">
      {t('administration.languageTabs.loadFailed')}
    </p>
  ) : null;
  if (tracks.data === undefined || inventory.data === undefined) return loadFailedAlert;
  const tabs = lessonLanguageTabs(tracks.data, inventory.data);
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
            <LanguageEditor {...props} languages={languages} tab={tab} videoId={videoId} />
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
}: Props & {
  tab?: LanguageTab;
  videoId?: string;
}) {
  const { t } = useTranslation();
  const [offer, setOffer] = useState<TranscriptDraftOffer>();
  const invalidateOffer = useCallback(() => setOffer(undefined), []);
  const [draftToken, setDraftToken] = useState<string>();
  const [replacement, setReplacement] = useState<TranscriptInventoryItem>();
  const resolveReplacement = useRef<((confirmed: boolean) => void) | undefined>(undefined);
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
    if (transcripts === undefined) return 'unavailable';
    if (tab === undefined || languages === undefined)
      return transcripts.save(lessonId, language, baseToken, segments);
    if (isUndetermined(language)) return 'track_language_required';
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
    try {
      const [freshInventory, tracks] = await Promise.all([
        languages.readInventory(lessonId),
        tab.track === undefined ? Promise.resolve(undefined) : languages.readTracks(videoId ?? ''),
      ]);
      inventory = freshInventory;
      if (!trackMatches(tracks)) return 'track_changed';
    } catch {
      return 'track_verification_failed';
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
                language: replacement?.language,
                time: replacement?.updatedAt,
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
