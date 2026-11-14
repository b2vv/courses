import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import {
  acceptedAudioFileTypes,
  audioTrackLanguages,
  availableAudioLanguages,
  canDeleteAudioTrack,
  canSetDefaultAudioTrack,
  isInterruptedAudioTrack,
  validAudioLabel,
} from '../domain/audio-track';
import type {
  AudioTrack,
  AudioTrackErrorCode,
  AudioTrackState,
  AudioTransferProgress,
} from '../domain/audio-track';
import {
  Button,
  Dialog,
  FileUpload,
  ProgressBar,
  Select,
  TextField,
} from '../../../presentation/ui';

export type LessonAudioTracksActionResult = AudioTrackErrorCode | null;

export type LessonAudioTracksActions = {
  upload(
    videoId: string,
    metadata: { readonly language: string; readonly label: string },
    file: File,
    onProgress: (progress: AudioTransferProgress) => void,
  ): Promise<LessonAudioTracksActionResult>;
  resume(
    track: AudioTrack,
    file: File,
    onProgress: (progress: AudioTransferProgress) => void,
  ): Promise<LessonAudioTracksActionResult>;
  cancel(track: AudioTrack): Promise<LessonAudioTracksActionResult>;
  rename(
    track: AudioTrack,
    language: string,
    label: string,
  ): Promise<LessonAudioTracksActionResult>;
  setDefault(track: AudioTrack): Promise<LessonAudioTracksActionResult>;
  remove(track: AudioTrack): Promise<LessonAudioTracksActionResult>;
};

export type LessonAudioTracksState = {
  readonly data: readonly AudioTrack[] | undefined;
  readonly isLoading: boolean;
  readonly error?: AudioTrackErrorCode;
};

/** The route supplies the per-video query as a hook, as it does for lesson drafts. */
export type CourseEditorAudio = {
  useTracks(videoId: string): LessonAudioTracksState;
  readonly actions: LessonAudioTracksActions;
};

type Translate = ReturnType<typeof useTranslation>['t'];
type OfferedLanguage = (typeof audioTrackLanguages)[number];

function isOfferedLanguage(value: string): value is OfferedLanguage {
  return (audioTrackLanguages as readonly string[]).includes(value);
}

function languageName(code: string, t: Translate): string {
  return isOfferedLanguage(code) ? t(`administration.audio.language.${code}`) : code;
}

function languageOptions(codes: readonly string[], t: Translate) {
  return codes.map((code) => ({ id: code, label: languageName(code, t) }));
}

function errorMessage(code: AudioTrackErrorCode, t: Translate): string {
  switch (code) {
    case 'authentication_required':
      return t('administration.video.error.authenticationRequired');
    case 'origin_forbidden':
      return t('administration.video.error.originForbidden');
    case 'forbidden':
    case 'course_forbidden':
      return t('administration.audio.error.forbidden');
    case 'duplicate_language':
      return t('administration.audio.error.duplicateLanguage');
    case 'upload_in_progress':
      return t('administration.audio.error.uploadInProgress');
    case 'unsupported_media':
      return t('administration.audio.error.unsupportedMedia');
    case 'upload_too_large':
      return t('administration.audio.error.uploadTooLarge');
    case 'video_not_found':
      return t('administration.audio.error.notFound');
    case 'upload_resume_unverified':
      return t('administration.audio.error.resumeUnverified');
    case 'invalid_request':
      return t('administration.audio.error.invalidRequest');
    case 'request_cancelled':
      return t('administration.audio.error.requestCancelled');
    case 'request_timeout':
      return t('administration.audio.error.requestTimeout');
    case 'dependency_unavailable':
    case 'invalid_response':
      return t('administration.audio.error.unavailable');
  }
}

function selectedFile(files: FileList | null): File | undefined {
  return files?.item(0) ?? undefined;
}

type FieldProps = {
  readonly videoId: string;
  readonly assetState: AudioTrackState;
  readonly audio: CourseEditorAudio;
};

export function LessonAudioTracksField({ videoId, assetState, audio }: FieldProps) {
  const state = audio.useTracks(videoId);
  return (
    <LessonAudioTracksView
      videoId={videoId}
      assetState={assetState}
      state={state}
      actions={audio.actions}
    />
  );
}

export function LessonAudioTracksView({
  videoId,
  assetState,
  state,
  actions,
}: {
  readonly videoId: string;
  readonly assetState: AudioTrackState;
  readonly state: LessonAudioTracksState;
  readonly actions: LessonAudioTracksActions;
}) {
  const { t } = useTranslation();
  const [progress, setProgress] = useState<AudioTransferProgress>();
  const [isBusy, setIsBusy] = useState(false);
  const [actionError, setActionError] = useState<AudioTrackErrorCode>();
  const [language, setLanguage] = useState<string>();
  const [label, setLabel] = useState('');
  const tracks = state.data ?? [];
  const offered = availableAudioLanguages(tracks);
  const chosenLanguage =
    language !== undefined && offered.includes(language) ? language : undefined;
  const transferring = progress !== undefined;

  async function run(action: () => Promise<LessonAudioTracksActionResult>) {
    setIsBusy(true);
    setActionError(undefined);
    try {
      const code = await action();
      if (code !== null) setActionError(code);
      return code;
    } catch {
      setActionError('dependency_unavailable');
      return 'dependency_unavailable' as const;
    } finally {
      setIsBusy(false);
    }
  }

  async function upload(files: FileList | null) {
    const file = selectedFile(files);
    if (file === undefined || chosenLanguage === undefined || !validAudioLabel(label)) return;
    setProgress({ receivedBytes: 0, declaredBytes: file.size });
    const code = await run(() =>
      actions.upload(videoId, { language: chosenLanguage, label }, file, setProgress),
    );
    setProgress(undefined);
    if (code === null) {
      setLanguage(undefined);
      setLabel('');
    }
  }

  async function resume(track: AudioTrack, files: FileList | null) {
    const file = selectedFile(files);
    if (file === undefined) return;
    setProgress({ receivedBytes: 0, declaredBytes: file.size });
    await run(() => actions.resume(track, file, setProgress));
    setProgress(undefined);
  }

  const shownError = actionError ?? state.error;
  const labelInvalid = label !== '' && !validAudioLabel(label);

  return (
    <section
      className="admin-audio-field"
      aria-busy={isBusy}
      aria-labelledby={`audio-tracks-${videoId}`}
    >
      <h4 id={`audio-tracks-${videoId}`}>{t('administration.audio.title')}</h4>
      {assetState === 'awaiting_audio' && (
        <p className="admin-video-hint">{t('administration.audio.awaitingHint')}</p>
      )}
      {shownError !== undefined && (
        <p className="admin-video-error" role="alert">
          {errorMessage(shownError, t)}
        </p>
      )}
      {state.isLoading && state.data === undefined && (
        <output>{t('administration.audio.loading')}</output>
      )}
      {tracks.length > 0 && (
        <ul className="admin-audio-tracks">
          {tracks.map((track) => (
            <AudioTrackRow
              key={track.trackId}
              track={track}
              tracks={tracks}
              assetState={assetState}
              isBusy={isBusy || transferring}
              onResume={(files) => void resume(track, files)}
              run={run}
              actions={actions}
            />
          ))}
        </ul>
      )}
      {progress !== undefined && (
        <ProgressBar
          label={t('administration.audio.uploading')}
          value={
            progress.declaredBytes === 0
              ? 0
              : (progress.receivedBytes / progress.declaredBytes) * 100
          }
        />
      )}
      {state.data !== undefined && state.error === undefined && !transferring && (
        <fieldset className="admin-audio-new">
          <legend>{t('administration.audio.add')}</legend>
          {offered.length === 0 ? (
            <p className="admin-video-hint">{t('administration.audio.noLanguageLeft')}</p>
          ) : (
            <>
              <Select
                label={t('administration.audio.languageLabel')}
                placeholder={t('administration.audio.languagePlaceholder')}
                options={languageOptions(offered, t)}
                selectedKey={chosenLanguage ?? null}
                onSelectionChange={(key) => setLanguage(key === null ? undefined : String(key))}
                isDisabled={isBusy}
              />
              <TextField
                label={t('administration.audio.nameLabel')}
                value={label}
                onChange={setLabel}
                isInvalid={labelInvalid}
                errorMessage={t('administration.audio.nameInvalid')}
                isDisabled={isBusy}
              />
              {(chosenLanguage === undefined || !validAudioLabel(label)) && (
                <p className="admin-video-hint">{t('administration.audio.chooseFirst')}</p>
              )}
              <FileUpload
                label={t('administration.audio.upload')}
                acceptedFileTypes={acceptedAudioFileTypes}
                allowsMultiple={false}
                isDisabled={isBusy || chosenLanguage === undefined || !validAudioLabel(label)}
                onSelect={(files) => void upload(files)}
              />
            </>
          )}
        </fieldset>
      )}
    </section>
  );
}

function AudioTrackRow({
  track,
  tracks,
  assetState,
  isBusy,
  onResume,
  run,
  actions,
}: {
  readonly track: AudioTrack;
  readonly tracks: readonly AudioTrack[];
  readonly assetState: AudioTrackState;
  readonly isBusy: boolean;
  readonly onResume: (files: FileList | null) => void;
  readonly run: (action: () => Promise<LessonAudioTracksActionResult>) => Promise<unknown>;
  readonly actions: LessonAudioTracksActions;
}) {
  const { t } = useTranslation();
  const interrupted = isInterruptedAudioTrack(track) && !isBusy;
  const deletable = canDeleteAudioTrack(tracks, track.trackId, assetState);
  const name = `${track.label} (${languageName(track.language, t)})`;
  return (
    <li className="admin-audio-track" data-audio-track-state={track.state}>
      <div className="admin-audio-track-head">
        <strong>{track.label}</strong>
        <span>{languageName(track.language, t)}</span>
        <output className="admin-video-state" data-audio-state={track.state}>
          {t(`administration.video.state.${track.state}`)}
        </output>
        {track.isDefault && (
          <span className="admin-audio-default">{t('administration.audio.default')}</span>
        )}
      </div>
      {(track.state === 'queued' || track.state === 'processing') && (
        <ProgressBar label={t('administration.audio.processing', { name })} isIndeterminate />
      )}
      {interrupted && <p className="admin-video-hint">{t('administration.audio.interrupted')}</p>}
      <div className="admin-video-actions">
        {interrupted && (
          <>
            <FileUpload
              label={t('administration.audio.resume')}
              acceptedFileTypes={acceptedAudioFileTypes}
              allowsMultiple={false}
              isDisabled={isBusy}
              onSelect={onResume}
            />
            <Button
              type="button"
              variant="danger"
              isDisabled={isBusy}
              onPress={() => void run(() => actions.cancel(track))}
            >
              {t('administration.audio.cancel')}
            </Button>
          </>
        )}
        {canSetDefaultAudioTrack(track) && (
          <Button
            type="button"
            variant="secondary"
            isDisabled={isBusy}
            aria-label={t('administration.audio.makeDefaultNamed', { name })}
            onPress={() => void run(() => actions.setDefault(track))}
          >
            {t('administration.audio.makeDefault')}
          </Button>
        )}
        {!isInterruptedAudioTrack(track) && (
          <RenameDialog track={track} tracks={tracks} name={name} run={run} actions={actions} />
        )}
        {deletable ? (
          <Dialog
            triggerLabel={t('administration.audio.delete')}
            triggerAriaLabel={t('administration.audio.deleteNamed', { name })}
            title={t('administration.audio.deleteTitle')}
            closeLabel={t('administration.audio.keep')}
          >
            {(close) => (
              <>
                <p>{t('administration.audio.deleteConfirm', { name })}</p>
                <Button
                  type="button"
                  variant="danger"
                  isDisabled={isBusy}
                  onPress={() => {
                    close();
                    void run(() => actions.remove(track));
                  }}
                >
                  {t('administration.audio.deleteConfirmAction')}
                </Button>
              </>
            )}
          </Dialog>
        ) : (
          <p className="admin-video-hint">{t('administration.audio.lastReady')}</p>
        )}
      </div>
    </li>
  );
}

function RenameDialog({
  track,
  tracks,
  name,
  run,
  actions,
}: {
  readonly track: AudioTrack;
  readonly tracks: readonly AudioTrack[];
  readonly name: string;
  readonly run: (action: () => Promise<LessonAudioTracksActionResult>) => Promise<unknown>;
  readonly actions: LessonAudioTracksActions;
}) {
  const { t } = useTranslation();
  const [language, setLanguage] = useState(track.language);
  const [label, setLabel] = useState(track.label);
  const valid = validAudioLabel(label);
  return (
    <Dialog
      triggerLabel={t('administration.audio.edit')}
      triggerAriaLabel={t('administration.audio.editNamed', { name })}
      title={t('administration.audio.editTitle')}
      closeLabel={t('administration.audio.keep')}
    >
      {(close) => (
        <form
          className="admin-audio-edit"
          onSubmit={(event) => {
            event.preventDefault();
            if (!valid) return;
            close();
            void run(() => actions.rename(track, language, label));
          }}
        >
          <Select
            label={t('administration.audio.languageLabel')}
            options={languageOptions(availableAudioLanguages(tracks, track.trackId), t)}
            selectedKey={language}
            onSelectionChange={(key) => {
              if (key !== null) setLanguage(String(key));
            }}
          />
          <TextField
            label={t('administration.audio.nameLabel')}
            value={label}
            onChange={setLabel}
            isInvalid={!valid}
            errorMessage={t('administration.audio.nameInvalid')}
          />
          <Button type="submit" isDisabled={!valid}>
            {t('administration.audio.save')}
          </Button>
        </form>
      )}
    </Dialog>
  );
}
