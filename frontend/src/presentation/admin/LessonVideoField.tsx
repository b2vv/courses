import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { acceptedVideoFileTypes } from '../../domain/media/video';
import type {
  LessonVideo,
  VideoErrorCode,
  VideoFailureReason,
  VideoTransferProgress,
} from '../../domain/media/video';
import { Button, FileUpload, ProgressBar } from '../ui';

const queuedRetryMs = 10 * 60_000;
const processingRetryMs = 40 * 60_000;

export type LessonVideoFieldActionResult = VideoErrorCode | null;

export type LessonVideoFieldActions = {
  upload(
    lessonId: string,
    file: File,
    onProgress: (progress: VideoTransferProgress) => void,
  ): Promise<LessonVideoFieldActionResult>;
  resume(
    video: LessonVideo,
    file: File,
    onProgress: (progress: VideoTransferProgress) => void,
  ): Promise<LessonVideoFieldActionResult>;
  retry(video: LessonVideo): Promise<LessonVideoFieldActionResult>;
  cancel(video: LessonVideo): Promise<LessonVideoFieldActionResult>;
};

export type CourseEditorVideo = {
  readonly rows: readonly LessonVideo[];
  readonly isLoading: boolean;
  readonly error?: VideoErrorCode;
  readonly actions: LessonVideoFieldActions;
};

type LessonVideoFieldProps = CourseEditorVideo & {
  readonly lessonId: string;
  readonly isSaved: boolean;
};

function shownVideo(rows: readonly LessonVideo[]): LessonVideo | undefined {
  const nonTerminal = rows.find((row) =>
    ['created', 'uploading', 'queued', 'processing'].includes(row.state),
  );
  if (nonTerminal !== undefined) return nonTerminal;
  const newest = (state: LessonVideo['state']) =>
    rows
      .filter((row) => row.state === state)
      .reduce<LessonVideo | undefined>(
        (selected, row) =>
          selected === undefined || row.generation > selected.generation ? row : selected,
        undefined,
      );
  return newest('ready') ?? newest('failed');
}

function canRetry(video: LessonVideo, now: number): boolean {
  if (video.state === 'failed') return video.receivedBytes === video.declaredBytes;
  const updatedAt = Date.parse(video.updatedAt);
  if (!Number.isFinite(updatedAt)) return false;
  const age = now - updatedAt;
  return (
    (video.state === 'queued' && age > queuedRetryMs) ||
    (video.state === 'processing' && age > processingRetryMs)
  );
}

function selectedFile(files: FileList | null): File | undefined {
  return files?.item(0) ?? undefined;
}

function errorMessage(code: VideoErrorCode, t: ReturnType<typeof useTranslation>['t']): string {
  switch (code) {
    case 'authentication_required':
      return t('administration.video.error.authenticationRequired');
    case 'origin_forbidden':
      return t('administration.video.error.originForbidden');
    case 'forbidden':
    case 'course_forbidden':
      return t('administration.video.error.forbidden');
    case 'lesson_not_found':
      return t('administration.video.error.lessonNotFound');
    case 'upload_in_progress':
      return t('administration.video.error.uploadInProgress');
    case 'unsupported_media':
      return t('administration.video.error.unsupportedMedia');
    case 'upload_too_large':
      return t('administration.video.error.uploadTooLarge');
    case 'video_not_found':
      return t('administration.video.error.videoNotFound');
    case 'video_not_retryable':
      return t('administration.video.error.videoNotRetryable');
    case 'upload_resume_unverified':
      return t('administration.video.error.resumeUnverified');
    case 'invalid_request':
      return t('administration.video.error.invalidRequest');
    case 'request_cancelled':
      return t('administration.video.error.requestCancelled');
    case 'request_timeout':
      return t('administration.video.error.requestTimeout');
    case 'dependency_unavailable':
    case 'invalid_response':
      return t('administration.video.error.unavailable');
  }
}

function failureMessage(
  reason: VideoFailureReason,
  t: ReturnType<typeof useTranslation>['t'],
): string {
  switch (reason) {
    case 'unsupported_source':
      return t('administration.video.failure.unsupportedSource');
    case 'probe_failed':
      return t('administration.video.failure.probeFailed');
    case 'render_failed':
      return t('administration.video.failure.renderFailed');
    case 'publish_failed':
      return t('administration.video.failure.publishFailed');
    case 'storage_unavailable':
      return t('administration.video.failure.storageUnavailable');
  }
}

export function LessonVideoField({
  lessonId,
  isSaved,
  rows,
  isLoading,
  error,
  actions,
}: LessonVideoFieldProps) {
  const { t } = useTranslation();
  const [progress, setProgress] = useState<VideoTransferProgress>();
  const [isBusy, setIsBusy] = useState(false);
  const [cancelledVideoId, setCancelledVideoId] = useState<string>();
  const [retriedVersion, setRetriedVersion] = useState<string>();
  const lessonRows = rows.filter(
    (row) => row.lessonId === lessonId && row.videoId !== cancelledVideoId,
  );
  const current = shownVideo(lessonRows);
  const currentVersion =
    current === undefined ? undefined : `${current.videoId}:${current.updatedAt}`;
  const [actionFailure, setActionFailure] = useState<{
    readonly code: VideoErrorCode;
    readonly videoVersion: string | undefined;
  }>();

  async function run(action: () => Promise<LessonVideoFieldActionResult>) {
    setIsBusy(true);
    setActionFailure(undefined);
    try {
      const code = await action();
      if (code !== null) setActionFailure({ code, videoVersion: currentVersion });
      return code;
    } catch {
      setActionFailure({ code: 'dependency_unavailable', videoVersion: currentVersion });
      return 'dependency_unavailable' as const;
    } finally {
      setIsBusy(false);
    }
  }

  async function transfer(files: FileList | null, video?: LessonVideo) {
    const file = selectedFile(files);
    if (file === undefined) return;
    setProgress({ receivedBytes: 0, declaredBytes: file.size });
    await run(() =>
      video === undefined
        ? actions.upload(lessonId, file, setProgress)
        : actions.resume(video, file, setProgress),
    );
    setProgress(undefined);
  }

  async function cancel(video: LessonVideo) {
    const code = await run(() => actions.cancel(video));
    if (code === null) setCancelledVideoId(video.videoId);
  }

  const actionError =
    actionFailure !== undefined && actionFailure.videoVersion === currentVersion
      ? actionFailure.code
      : undefined;
  const statusError = actionError ?? error;
  // The query's queued/processing poll rerenders this field; retry eligibility must use that tick.
  // oxlint-disable-next-line react/purity -- This is a server-aligned wall-clock threshold, not derived UI state.
  const now = Date.now();
  const retryable =
    current === undefined ? false : canRetry(current, now) && retriedVersion !== currentVersion;
  const interrupted =
    current !== undefined &&
    (current.state === 'created' || current.state === 'uploading') &&
    progress === undefined;

  return (
    <section className="admin-video-field" aria-busy={isBusy}>
      <h3>{t('administration.video.title')}</h3>
      {!isSaved ? (
        <p className="admin-video-hint">{t('administration.video.unsavedHint')}</p>
      ) : (
        <>
          {statusError !== undefined && (
            <p className="admin-video-error" role="alert">
              {errorMessage(statusError, t)}
            </p>
          )}
          {isLoading && current === undefined && (
            <output>{t('administration.video.loading')}</output>
          )}
          {current !== undefined && (
            <output className="admin-video-state" data-video-state={current.state}>
              {t(`administration.video.state.${current.state}`)}
            </output>
          )}
          {progress !== undefined && (
            <ProgressBar
              label={t('administration.video.uploading')}
              value={
                progress.declaredBytes === 0
                  ? 0
                  : (progress.receivedBytes / progress.declaredBytes) * 100
              }
            />
          )}
          {current !== undefined &&
            (current.state === 'queued' || current.state === 'processing') &&
            !retryable &&
            progress === undefined && (
              <ProgressBar label={t('administration.video.processing')} isIndeterminate />
            )}
          {current?.state === 'failed' && current.failureReason !== undefined && (
            <p>{failureMessage(current.failureReason, t)}</p>
          )}
          {interrupted && (
            <>
              <p className="admin-video-hint">{t('administration.video.interrupted')}</p>
              <div className="admin-video-actions">
                <FileUpload
                  label={t('administration.video.resume')}
                  acceptedFileTypes={acceptedVideoFileTypes}
                  allowsMultiple={false}
                  isDisabled={isBusy}
                  onSelect={(files) => void transfer(files, current)}
                />
                <Button
                  type="button"
                  variant="danger"
                  isDisabled={isBusy}
                  onPress={() => void cancel(current)}
                >
                  {t('administration.video.cancel')}
                </Button>
              </div>
            </>
          )}
          {current !== undefined && retryable && (
            <Button
              type="button"
              variant="secondary"
              isDisabled={isBusy}
              onPress={() =>
                void run(() => actions.retry(current)).then((code) => {
                  if (code === null) setRetriedVersion(currentVersion);
                })
              }
            >
              {t('administration.video.retry')}
            </Button>
          )}
          {progress === undefined &&
            !interrupted &&
            (current === undefined || current.state === 'ready' || current.state === 'failed') &&
            error === undefined &&
            !isLoading && (
              <FileUpload
                label={
                  current?.state === 'ready'
                    ? t('administration.video.replace')
                    : t('administration.video.upload')
                }
                acceptedFileTypes={acceptedVideoFileTypes}
                allowsMultiple={false}
                isDisabled={isBusy}
                onSelect={(files) => void transfer(files)}
              />
            )}
        </>
      )}
    </section>
  );
}
