import { useCallback, useEffect, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import {
  heartbeatFailureDirective,
  initialHeartbeatBackoffMs,
  mediaFailureDirective,
  mintFailureDirective,
  nextHeartbeatDelayMs,
  type PlaybackReloadMode,
} from '../../application/playback/recovery';
import type {
  LearningLesson,
  PlaybackHeartbeat,
  PlaybackHeartbeatReason,
  PlaybackResult,
  PlaybackSession,
} from '../../domain/playback/playback';
import type {
  CreatePlaybackMedia,
  PlaybackMedia,
  PlaybackMediaError,
  PlaybackQuality,
} from '../../ports/PlaybackMedia';
import { Button, EmptyState, ErrorState, Select } from '../ui';

export type LessonPlayerActions = {
  play(): Promise<PlaybackResult<PlaybackSession>>;
  heartbeat(heartbeat: PlaybackHeartbeat): Promise<PlaybackResult<void>>;
  refresh(): Promise<void>;
};

type LessonPlayerProps = {
  readonly courseTitle: string;
  readonly lesson: LearningLesson;
  readonly actions: LessonPlayerActions;
  readonly createMedia: CreatePlaybackMedia;
};

type PlayerFailure = 'denied' | 'taken-over' | 'unavailable' | null;

function finiteTime(value: number): number {
  return Number.isFinite(value) && value >= 0 ? value : 0;
}

export function LessonPlayer({ courseTitle, lesson, actions, createMedia }: LessonPlayerProps) {
  const { t } = useTranslation();
  const videoRef = useRef<HTMLVideoElement>(null);
  const [videoElement, setVideoElement] = useState<HTMLVideoElement | null>(null);
  const attachVideo = useCallback((video: HTMLVideoElement | null) => {
    videoRef.current = video;
    setVideoElement(video);
  }, []);
  const mediaRef = useRef<PlaybackMedia | undefined>(undefined);
  const sessionRef = useRef<PlaybackSession | undefined>(undefined);
  const clientSeqRef = useRef(lesson.clientSeq + 1);
  // A finished lesson reopens from the beginning (A25). The saved position stays in
  // `lesson_progress` - it is what marked the lesson complete - but resuming two seconds from
  // the end would hand the student a video that ends before it starts.
  const resumeFrom = lesson.completed ? 0 : lesson.position;
  const pendingTimeRef = useRef(resumeFrom);
  const reloadModeRef = useRef<PlaybackReloadMode>('initial');
  const mintRef = useRef<Promise<void> | undefined>(undefined);
  const heartbeatTimerRef = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);
  const mediaRetryTimerRef = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);
  const mintRetryTimerRef = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);
  const mintRetryResolveRef = useRef<(() => void) | undefined>(undefined);
  const heartbeatQueueRef = useRef<Promise<'continue' | 'stop'>>(Promise.resolve('continue'));
  const heartbeatStoppedRef = useRef(false);
  const heartbeatBackoffRef = useRef(initialHeartbeatBackoffMs);
  const mintRetriesRef = useRef(0);
  const mediaRetriesRef = useRef(0);
  const positionRef = useRef(resumeFrom);
  const [mediaKind, setMediaKind] = useState<'hls' | 'native'>('native');
  const [qualities, setQualities] = useState<readonly PlaybackQuality[]>([]);
  const [failure, setFailure] = useState<PlayerFailure>(null);
  const [starting, setStarting] = useState(false);
  const [playing, setPlaying] = useState(false);
  // Raised immediately before a pause we cause ourselves, and read by the pause listener. The
  // listener must run for every pause - that is the point: the state follows the element, so a
  // pause from the browser's media hub also stops renewing ownership. But the exit write belongs
  // to whoever initiated it, with its own reason; without this flag the visibility handler would
  // record `hide` and `pause` about one event, and teardown would record a pause while shutting
  // down.
  //
  // Only raised when the element is actually playing. `pause()` on already-paused media fires
  // nothing, the flag would stay up, and the next genuine external pause would be swallowed.
  const selfPausedRef = useRef(false);
  const pauseSelf = useCallback(() => {
    const video = videoRef.current;
    if (video === null || video.paused) return;
    selfPausedRef.current = true;
    video.pause();
  }, []);
  const [duration, setDuration] = useState(0);
  const [position, setPosition] = useState(resumeFrom);
  const [restartAt, setRestartAt] = useState<number | null>(null);

  const stop = useCallback(() => {
    pauseSelf();
    mediaRef.current?.stopLoad();
    setPlaying(false);
    if (heartbeatTimerRef.current !== undefined) clearTimeout(heartbeatTimerRef.current);
    if (mediaRetryTimerRef.current !== undefined) clearTimeout(mediaRetryTimerRef.current);
    if (mintRetryTimerRef.current !== undefined) clearTimeout(mintRetryTimerRef.current);
    mintRetryResolveRef.current?.();
    mintRetryResolveRef.current = undefined;
  }, [pauseSelf]);

  const terminalFailure = useCallback(
    (next: Exclude<PlayerFailure, null>, refresh: boolean) => {
      heartbeatStoppedRef.current = true;
      stop();
      setFailure(next);
      if (refresh) void actions.refresh();
    },
    [actions, stop],
  );

  const loadSession = useCallback(
    (resumeAt: number, mode: PlaybackReloadMode): Promise<void> => {
      if (mintRef.current !== undefined) return mintRef.current;
      reloadModeRef.current = mode;
      pendingTimeRef.current = finiteTime(resumeAt);
      setStarting(true);
      const mint = (async () => {
        while (mediaRef.current !== undefined) {
          const result = await actions.play();
          if (!result.ok) {
            const directive = mintFailureDirective(result.error.code, mintRetriesRef.current);
            if (directive.kind === 'retry') {
              mintRetriesRef.current = directive.nextRetryCount;
              await new Promise<void>((resolve) => {
                mintRetryResolveRef.current = resolve;
                mintRetryTimerRef.current = setTimeout(() => {
                  mintRetryTimerRef.current = undefined;
                  mintRetryResolveRef.current = undefined;
                  resolve();
                }, directive.delayMs);
              });
              continue;
            } else {
              terminalFailure(
                directive.kind === 'denied' ? 'denied' : 'unavailable',
                directive.refresh,
              );
            }
            return;
          }
          mintRetriesRef.current = 0;
          sessionRef.current = result.value;
          clientSeqRef.current = result.value.clientSeq + 1;
          heartbeatStoppedRef.current = false;
          mediaRef.current?.load(result.value.mediaPath);
          return;
        }
      })().finally(() => {
        setStarting(false);
        mintRef.current = undefined;
      });
      mintRef.current = mint;
      return mint;
    },
    [actions, terminalFailure],
  );

  const queueHeartbeatRef = useRef<
    (reason: PlaybackHeartbeatReason) => Promise<'continue' | 'stop'>
  >(() => Promise.resolve('stop'));

  const handleMediaError = useCallback(
    (error: PlaybackMediaError) => {
      const directive = mediaFailureDirective(
        error.status,
        error.retryable,
        mediaRetriesRef.current,
        reloadModeRef.current,
        mintRef.current !== undefined,
        error.url,
        sessionRef.current?.mediaPath,
      );
      if (directive.kind === 'ignore') return;
      if (directive.kind === 'await-verdict') {
        // A refused segment says only "not now" - the edge gives the same 403 to an expired
        // session, an evicted one and a revoked course (A7). Stop loading and let the heartbeat
        // answer, since it is the only place those three are distinguishable.
        mediaRef.current?.stopLoad();
        const ownedBefore = sessionRef.current?.playbackId;
        // Through the ref, because `queueHeartbeat` is declared below this handler and the
        // queue also keeps this out of the way of a cadence beat already in flight.
        void queueHeartbeatRef.current('tick').then((outcome) => {
          // A remint answers 'continue' too, and it has already reloaded the source - the slot
          // was merely expired and playback goes on. Only an unchanged session means the
          // heartbeat found nothing wrong, and then the refusal really was about access.
          if (outcome === 'continue' && sessionRef.current?.playbackId === ownedBefore)
            terminalFailure('denied', true);
        });
        return;
      }
      const video = videoRef.current;
      const resumeAt = finiteTime(video?.currentTime ?? positionRef.current);
      if (directive.kind === 'reload-course') {
        mediaRetriesRef.current = directive.nextRetryCount;
        mediaRef.current?.stopLoad();
        void actions.refresh().then(() => loadSession(resumeAt, 'missing'));
        return;
      }
      if (directive.kind === 'retry-media') {
        const session = sessionRef.current;
        if (session === undefined) {
          terminalFailure('unavailable', false);
          return;
        }
        mediaRetriesRef.current = directive.nextRetryCount;
        mediaRef.current?.stopLoad();
        mediaRetryTimerRef.current = setTimeout(() => {
          pendingTimeRef.current = resumeAt;
          mediaRef.current?.load(session.mediaPath);
        }, directive.delayMs);
        return;
      }
      terminalFailure('unavailable', directive.refresh);
    },
    [actions, loadSession, terminalFailure],
  );

  useEffect(() => {
    const video = videoElement;
    if (lesson.videoState !== 'ready' || video === null) return;
    const media = createMedia(video, {
      onReady() {
        const knownDuration = finiteTime(video.duration);
        setDuration(knownDuration);
        const resumeAt =
          knownDuration > 0
            ? Math.min(pendingTimeRef.current, knownDuration)
            : pendingTimeRef.current;
        video.currentTime = resumeAt;
        positionRef.current = finiteTime(video.currentTime);
        setPosition(positionRef.current);
        reloadModeRef.current = 'initial';
        void video.play();
      },
      onLevels: setQualities,
      onError: handleMediaError,
    });
    mediaRef.current = media;
    setMediaKind(media.kind);
    return () => {
      if (heartbeatTimerRef.current !== undefined) clearTimeout(heartbeatTimerRef.current);
      if (mediaRetryTimerRef.current !== undefined) clearTimeout(mediaRetryTimerRef.current);
      if (mintRetryTimerRef.current !== undefined) clearTimeout(mintRetryTimerRef.current);
      mintRetryResolveRef.current?.();
      mintRetryResolveRef.current = undefined;
      media.destroy();
      mediaRef.current = undefined;
    };
  }, [createMedia, handleMediaError, lesson.videoState, videoElement]);

  const heartbeatNow = useCallback(
    async (reason: PlaybackHeartbeatReason): Promise<'continue' | 'stop'> => {
      if (heartbeatStoppedRef.current) return 'stop';
      const session = sessionRef.current;
      if (session === undefined) return 'stop';
      const video = videoRef.current;
      const result = await actions.heartbeat({
        playbackId: session.playbackId,
        lessonId: lesson.id,
        positionSeconds: finiteTime(video?.currentTime ?? positionRef.current),
        clientSeq: clientSeqRef.current,
        reason,
      });
      if (result.ok) {
        clientSeqRef.current += 1;
        heartbeatBackoffRef.current = initialHeartbeatBackoffMs;
        return 'continue';
      }
      const directive = heartbeatFailureDirective(result.error.code, heartbeatBackoffRef.current);
      switch (directive.kind) {
        case 'remint': {
          const resumeAt = finiteTime(video?.currentTime ?? positionRef.current);
          mediaRef.current?.stopLoad();
          await loadSession(resumeAt, 'expired');
          return 'continue';
        }
        case 'taken-over':
          terminalFailure('taken-over', false);
          return 'stop';
        case 'denied':
          terminalFailure('denied', directive.refresh);
          return 'stop';
        case 'unavailable':
          terminalFailure('unavailable', directive.refresh);
          return 'stop';
        case 'retry':
          heartbeatBackoffRef.current = directive.delayMs;
          return 'continue';
      }
    },
    [actions, lesson.id, loadSession, terminalFailure],
  );

  const queueHeartbeat = useCallback(
    (reason: PlaybackHeartbeatReason): Promise<'continue' | 'stop'> => {
      const run = () => heartbeatNow(reason);
      const queued = heartbeatQueueRef.current.then(run, run);
      heartbeatQueueRef.current = queued;
      return queued;
    },
    [heartbeatNow],
  );

  useEffect(() => {
    queueHeartbeatRef.current = queueHeartbeat;
  }, [queueHeartbeat]);

  useEffect(() => {
    if (!playing || sessionRef.current === undefined) return;
    let cancelled = false;
    const cadenceMs = sessionRef.current.heartbeatSeconds * 1_000;
    let nextHeartbeatAt = Date.now() + cadenceMs;
    const schedule = () => {
      const delay = Math.max(0, nextHeartbeatAt - Date.now());
      heartbeatTimerRef.current = setTimeout(() => {
        const scheduledAt = nextHeartbeatAt;
        void queueHeartbeat('tick').then((outcome) => {
          if (cancelled || outcome === 'stop') return;
          const session = sessionRef.current;
          if (session === undefined) return;
          const nextDelay = nextHeartbeatDelayMs(
            session.heartbeatSeconds,
            heartbeatBackoffRef.current,
          );
          nextHeartbeatAt =
            heartbeatBackoffRef.current === initialHeartbeatBackoffMs
              ? scheduledAt + nextDelay
              : Date.now() + nextDelay;
          schedule();
        });
      }, delay);
    };
    schedule();
    return () => {
      cancelled = true;
      if (heartbeatTimerRef.current !== undefined) clearTimeout(heartbeatTimerRef.current);
    };
  }, [pauseSelf, playing, queueHeartbeat]);

  useEffect(() => {
    const handleVisibilityChange = () => {
      if (document.visibilityState !== 'hidden') return;
      pauseSelf();
      setPlaying(false);
      void queueHeartbeat('hide');
    };
    document.addEventListener('visibilitychange', handleVisibilityChange);
    return () => document.removeEventListener('visibilitychange', handleVisibilityChange);
  }, [pauseSelf, queueHeartbeat]);

  useEffect(
    () => () => {
      void queueHeartbeatRef.current('nav');
    },
    [lesson.id],
  );

  useEffect(() => {
    if (restartAt === null || mediaRef.current === undefined) return;
    setRestartAt(null);
    void loadSession(restartAt, 'initial');
  }, [loadSession, restartAt, videoElement]);

  if (lesson.videoState === null) return <EmptyState title={t('learn.video.noVideo')} />;
  if (lesson.videoState === 'processing') return <EmptyState title={t('learn.video.processing')} />;
  if (lesson.videoState === 'failed') return <ErrorState title={t('learn.video.failed')} />;

  if (failure !== null) {
    const title =
      failure === 'taken-over'
        ? t('learn.video.takenOver')
        : failure === 'denied'
          ? t('learn.video.denied')
          : t('learn.video.unavailable');
    return (
      <ErrorState
        title={title}
        onRetry={() => {
          mintRetriesRef.current = 0;
          mediaRetriesRef.current = 0;
          setFailure(null);
          setRestartAt(position);
        }}
        retryLabel={t('learn.video.playAgain')}
      />
    );
  }

  const qualityOptions = [
    { id: 'auto', label: t('learn.video.qualityAuto') },
    ...qualities.map((quality) => ({
      id: String(quality.id),
      label: quality.label,
    })),
  ];
  const seekTo = (next: number) => {
    const bounded = Math.max(0, Math.min(Math.max(1, duration), next));
    if (videoRef.current !== null) videoRef.current.currentTime = bounded;
    positionRef.current = bounded;
    setPosition(bounded);
    void queueHeartbeat('seek');
  };

  return (
    <section className="lesson-player" aria-labelledby="lesson-player-title">
      <div className="lesson-player-status">
        <span>{t('learn.video.status')}</span>
        <strong>{playing ? t('learn.video.playing') : t('learn.video.ready')}</strong>
      </div>
      <h1 id="lesson-player-title">{lesson.title}</h1>
      <p className="lesson-player-course">{courseTitle}</p>
      <div className="lesson-player-frame">
        {/* oxlint-disable-next-line jsx-a11y/media-has-caption -- T7 has no caption source contract; an empty track would be misleading. */}
        <video
          ref={attachVideo}
          aria-label={lesson.title}
          onDurationChange={(event) => setDuration(finiteTime(event.currentTarget.duration))}
          onTimeUpdate={(event) => {
            positionRef.current = finiteTime(event.currentTarget.currentTime);
            mediaRetriesRef.current = 0;
            setPosition(positionRef.current);
          }}
          onPause={() => {
            setPlaying(false);
            // A pause we caused has already recorded its own reason. One from anywhere else -
            // the media hub, picture-in-picture, a media key - is an exit like any other, and
            // without a write the student loses everything watched since the last beat.
            if (selfPausedRef.current) {
              selfPausedRef.current = false;
              return;
            }
            void queueHeartbeatRef.current('pause');
          }}
          onPlaying={() => setPlaying(true)}
          onEnded={(event) => {
            positionRef.current = finiteTime(event.currentTarget.currentTime);
            setPosition(positionRef.current);
            setPlaying(false);
            void queueHeartbeat('ended');
          }}
          playsInline
        />
      </div>
      <div className="lesson-player-controls" aria-label={t('learn.video.controls')}>
        <Button
          type="button"
          isDisabled={starting}
          onPress={() => {
            if (playing) {
              pauseSelf();
              void queueHeartbeat('pause');
              return;
            }
            if (sessionRef.current === undefined) {
              void loadSession(position, 'initial');
              return;
            }
            void videoRef.current?.play();
          }}
        >
          {starting
            ? t('learn.video.starting')
            : playing
              ? t('learn.video.pause')
              : t('learn.video.play')}
        </Button>
        <label className="lesson-player-seek">
          <span>{t('learn.video.position')}</span>
          <input
            type="range"
            min={0}
            max={Math.max(1, duration)}
            step={1}
            value={Math.min(position, Math.max(1, duration))}
            onChange={(event) => seekTo(Number(event.currentTarget.value))}
            onKeyDown={(event) => {
              if (event.key !== 'ArrowLeft' && event.key !== 'ArrowRight') return;
              event.preventDefault();
              seekTo(position + (event.key === 'ArrowRight' ? 5 : -5));
            }}
          />
        </label>
        {mediaKind === 'hls' && (
          <Select
            label={t('learn.video.quality')}
            options={qualityOptions}
            defaultSelectedKey="auto"
            onSelectionChange={(key) => {
              if (key === 'auto') mediaRef.current?.selectLevel('auto');
              else {
                const level = Number(key);
                if (Number.isInteger(level)) mediaRef.current?.selectLevel(level);
              }
            }}
          />
        )}
      </div>
    </section>
  );
}
