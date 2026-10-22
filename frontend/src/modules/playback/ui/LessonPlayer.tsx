import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import {
  heartbeatFailureDirective,
  initialHeartbeatBackoffMs,
  mediaFailureDirective,
  mintFailureDirective,
  nextHeartbeatDelayMs,
  type PlaybackReloadMode,
} from '../application/recovery';
import type {
  LearningLesson,
  PlaybackHeartbeat,
  PlaybackHeartbeatReason,
  PlaybackResult,
  PlaybackSession,
} from '../domain/playback';
import type {
  CreatePlaybackMedia,
  PlaybackMedia,
  PlaybackMediaError,
  PlaybackQuality,
} from '../ports/PlaybackMedia';
import { Button, EmptyState, ErrorState, Select } from '../../../presentation/ui';

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
  /**
   * The channel the transcript needs, handed out rather than read from the outside.
   *
   * `seek` is null until the duration is known, because before the first play `duration` is 0 and
   * `seekTo` clamps every request into the first second — a jump that silently lands somewhere
   * else is worse than one that is unavailable. Position travels through a ref-style callback so
   * that a transcript re-rendering on every frame cannot re-render the player with it.
   */
  /**
   * WebVTT for this lesson, or undefined when there is no transcript. It becomes a real <track>,
   * which is what finally removes the jsx-a11y suppression this file carried since Phase 6: the
   * suppression said an empty track would be misleading, and it was right — the fix was never to
   * silence the rule but to have captions.
   */
  readonly captions?: string;
  readonly captionsLanguage?: string;
  readonly onPlaybackChannel?: (channel: {
    readonly positionMs: number;
    readonly seek: ((positionMs: number) => void) | null;
  }) => void;
};

type PlayerFailure = 'denied' | 'taken-over' | 'unavailable' | null;

function finiteTime(value: number): number {
  return Number.isFinite(value) && value >= 0 ? value : 0;
}

// mm:ss, or h:mm:ss past an hour. Floor, never negative: a scrubber one frame past the end must
// not read 60:00 because duration rounded up.
function clockTime(value: number): string {
  const total = Math.max(0, Math.floor(value));
  const seconds = total % 60;
  const minutes = Math.floor(total / 60) % 60;
  const hours = Math.floor(total / 3600);
  const mm = hours > 0 ? String(minutes).padStart(2, '0') : String(minutes);
  const ss = String(seconds).padStart(2, '0');
  return hours > 0 ? `${hours}:${mm}:${ss}` : `${mm}:${ss}`;
}

const PLAYBACK_RATES = [0.75, 1, 1.25, 1.5, 2] as const;

export function LessonPlayer({
  courseTitle,
  lesson,
  actions,
  createMedia,
  captions,
  captionsLanguage,
  onPlaybackChannel,
}: LessonPlayerProps) {
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
  // Chrome state that mirrors the element, so the controls render the element's truth even when it
  // is changed from elsewhere (a media key, the browser's own fullscreen/PiP affordance).
  const frameRef = useRef<HTMLDivElement>(null);
  const stageRef = useRef<HTMLDivElement>(null);
  const [muted, setMuted] = useState(false);
  const [volume, setVolume] = useState(1);
  const [rate, setRate] = useState(1);
  const [isFullscreen, setIsFullscreen] = useState(false);
  const [isPictureInPicture, setIsPictureInPicture] = useState(false);
  const pipSupported = typeof document !== 'undefined' && document.pictureInPictureEnabled === true;

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

  // Published after every render so the transcript sees the current position and learns the
  // moment seeking becomes possible. This sits ABOVE the early returns below: a lesson with no
  // video still renders, and hooks may not be skipped for it.
  // One definition, above the early returns, so the control and the transcript seek identically
  // and hooks are never skipped for a lesson that renders an empty state instead of a player.
  const seekTo = useCallback(
    (next: number) => {
      const bounded = Math.max(0, Math.min(Math.max(1, duration), next));
      if (videoRef.current !== null) videoRef.current.currentTime = bounded;
      positionRef.current = bounded;
      setPosition(bounded);
      void queueHeartbeat('seek');
    },
    [duration, queueHeartbeat],
  );
  const seekMs = useCallback((positionMs: number) => seekTo(positionMs / 1000), [seekTo]);

  // One toggle shared by the labelled control, a click on the frame and the keyboard, so every
  // affordance starts, resumes and pauses through the same path the session logic already guards.
  const togglePlay = useCallback(() => {
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
  }, [playing, pauseSelf, queueHeartbeat, loadSession, position]);

  const applyMuted = useCallback((next: boolean) => {
    const video = videoRef.current;
    if (video !== null) video.muted = next;
    setMuted(next);
  }, []);
  const applyVolume = useCallback((next: number) => {
    const bounded = Math.max(0, Math.min(1, next));
    const video = videoRef.current;
    if (video !== null) {
      video.volume = bounded;
      // Nudging the volume off zero unmutes: a muted bar the student is dragging up should be heard.
      if (bounded > 0 && video.muted) video.muted = false;
    }
    setVolume(bounded);
  }, []);
  const applyRate = useCallback((next: number) => {
    const video = videoRef.current;
    if (video !== null) video.playbackRate = next;
    setRate(next);
  }, []);

  const toggleFullscreen = useCallback(() => {
    const frame = frameRef.current;
    if (frame === null) return;
    if (document.fullscreenElement === frame) void document.exitFullscreen();
    else void frame.requestFullscreen?.();
  }, []);
  const togglePictureInPicture = useCallback(() => {
    const video = videoRef.current;
    if (video === null || !pipSupported) return;
    if (document.pictureInPictureElement === video) void document.exitPictureInPicture();
    else void video.requestPictureInPicture?.();
  }, [pipSupported]);

  useEffect(() => {
    const onFullscreenChange = () =>
      setIsFullscreen(document.fullscreenElement === frameRef.current);
    document.addEventListener('fullscreenchange', onFullscreenChange);
    return () => document.removeEventListener('fullscreenchange', onFullscreenChange);
  }, []);

  useEffect(() => {
    const video = videoElement;
    if (video === null) return;
    const onEnter = () => setIsPictureInPicture(true);
    const onLeave = () => setIsPictureInPicture(false);
    video.addEventListener('enterpictureinpicture', onEnter);
    video.addEventListener('leavepictureinpicture', onLeave);
    return () => {
      video.removeEventListener('enterpictureinpicture', onEnter);
      video.removeEventListener('leavepictureinpicture', onLeave);
    };
  }, [videoElement]);

  useEffect(() => {
    const stage = stageRef.current;
    if (stage === null) return;
    const onKey = (event: KeyboardEvent) => {
      // Only when the stage itself holds focus: a key aimed at the slider or a button is theirs.
      if (event.target !== stage) return;
      switch (event.key) {
        case ' ':
        case 'k':
          event.preventDefault();
          togglePlay();
          break;
        case 'ArrowLeft':
          event.preventDefault();
          seekTo(positionRef.current - 5);
          break;
        case 'ArrowRight':
          event.preventDefault();
          seekTo(positionRef.current + 5);
          break;
        case 'f':
          event.preventDefault();
          toggleFullscreen();
          break;
        case 'm':
          event.preventDefault();
          applyMuted(!videoRef.current?.muted);
          break;
        default:
          break;
      }
    };
    stage.addEventListener('keydown', onKey);
    return () => stage.removeEventListener('keydown', onKey);
  }, [togglePlay, seekTo, toggleFullscreen, applyMuted]);

  // A blob URL rather than an endpoint: the transcript is already in memory, so a second request
  // would fetch what the page holds. Revoked on change and on unmount — an object URL outlives
  // the element that used it and would otherwise leak for the life of the document.
  const captionsUrl = useMemo(
    () =>
      captions === undefined
        ? null
        : URL.createObjectURL(new Blob([captions], { type: 'text/vtt' })),
    [captions],
  );
  useEffect(() => {
    if (captionsUrl === null) return;
    return () => URL.revokeObjectURL(captionsUrl);
  }, [captionsUrl]);

  useEffect(() => {
    onPlaybackChannel?.({
      positionMs: position * 1000,
      seek: duration > 0 ? seekMs : null,
    });
  });

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

  const speedOptions = PLAYBACK_RATES.map((value) => ({
    id: String(value),
    label: `${value}×`,
  }));
  const boundedDuration = Math.max(1, duration);

  return (
    <section className="lesson-player" aria-labelledby="lesson-player-title">
      <header className="lesson-player-head">
        <p className="lesson-player-course">{courseTitle}</p>
        <h1 id="lesson-player-title">{lesson.title}</h1>
      </header>
      {/* The whole player is one surface: the frame and its controls read as a single media object
          rather than a video followed by a form. The stage is focusable only programmatically
          (tabIndex -1) — a click on the video focuses it — so Space/arrows/F/M drive playback
          without adding a tab stop ahead of the controls. The handler (attached in an effect) acts
          only when the stage itself is the focus target, so the slider's arrows and the buttons
          keep their own keys. */}
      <div
        className="lesson-player-stage"
        data-playing={playing || undefined}
        tabIndex={-1}
        aria-label={t('learn.video.controls')}
        ref={stageRef}
      >
        <div className="lesson-player-frame" ref={frameRef}>
          {/*
            oxlint-disable-next-line jsx-a11y/media-has-caption -- A <track> IS attached whenever the
            lesson has a transcript (see captionsUrl above). It cannot be unconditional: a lesson
            without a transcript has nothing to attach, and an empty track is exactly the misleading
            thing the Phase 6 comment warned about. The rule cannot express "captions when they
            exist", so the suppression stays — with an accurate reason this time.
          */}
          <video
            ref={attachVideo}
            aria-label={lesson.title}
            onClick={() => {
              togglePlay();
              // Focus the stage so the keyboard shortcuts take over after a click on the video.
              stageRef.current?.focus();
            }}
            onVolumeChange={(event) => {
              setMuted(event.currentTarget.muted);
              setVolume(event.currentTarget.volume);
            }}
            onRateChange={(event) => setRate(event.currentTarget.playbackRate)}
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
          >
            {captionsUrl === null ? null : (
              <track
                kind="captions"
                default
                src={captionsUrl}
                srcLang={captionsLanguage ?? 'uk'}
                label={t('learn.transcript.heading')}
              />
            )}
          </video>
          {!playing && !starting && (
            <span className="lesson-player-bigplay" aria-hidden="true">
              <svg viewBox="0 0 24 24" width="34" height="34" fill="currentColor">
                <path d="M8 5v14l11-7z" />
              </svg>
            </span>
          )}
        </div>
        <div className="lesson-player-controlbar">
          <div className="lesson-player-scrubrow">
            <span className="lesson-player-time">{clockTime(position)}</span>
            <input
              className="lesson-player-seek"
              type="range"
              aria-label={t('learn.video.position')}
              min={0}
              max={boundedDuration}
              step={1}
              value={Math.min(position, boundedDuration)}
              onChange={(event) => seekTo(Number(event.currentTarget.value))}
              onKeyDown={(event) => {
                if (event.key !== 'ArrowLeft' && event.key !== 'ArrowRight') return;
                event.preventDefault();
                seekTo(position + (event.key === 'ArrowRight' ? 5 : -5));
              }}
            />
            <span className="lesson-player-time lesson-player-time--total">
              {clockTime(duration)}
            </span>
          </div>
          <div className="lesson-player-buttons">
            <div className="lesson-player-primary">
              <Button type="button" isDisabled={starting} onPress={togglePlay}>
                {starting
                  ? t('learn.video.starting')
                  : playing
                    ? t('learn.video.pause')
                    : t('learn.video.play')}
              </Button>
              <div className="lesson-player-volume">
                <button
                  type="button"
                  className="lesson-player-iconbtn"
                  aria-label={muted ? t('learn.video.unmute') : t('learn.video.mute')}
                  aria-pressed={muted}
                  onClick={() => applyMuted(!muted)}
                >
                  <svg viewBox="0 0 24 24" width="20" height="20" aria-hidden="true">
                    <path
                      d="M4 9v6h4l5 4V5L8 9H4z"
                      fill="currentColor"
                      stroke="currentColor"
                      strokeWidth="1.2"
                      strokeLinejoin="round"
                    />
                    {muted ? (
                      <path
                        d="M16 9l5 6m0-6l-5 6"
                        stroke="currentColor"
                        strokeWidth="1.8"
                        strokeLinecap="round"
                        fill="none"
                      />
                    ) : (
                      <path
                        d="M16.5 8.5a5 5 0 0 1 0 7M18.8 6.2a8 8 0 0 1 0 11.6"
                        stroke="currentColor"
                        strokeWidth="1.5"
                        strokeLinecap="round"
                        fill="none"
                      />
                    )}
                  </svg>
                </button>
                <input
                  className="lesson-player-volumeslider"
                  type="range"
                  aria-label={t('learn.video.volume')}
                  min={0}
                  max={1}
                  step={0.05}
                  value={muted ? 0 : volume}
                  onChange={(event) => applyVolume(Number(event.currentTarget.value))}
                />
              </div>
            </div>
            <div className="lesson-player-tools">
              <Select
                label={t('learn.video.speed')}
                options={speedOptions}
                selectedKey={String(rate)}
                onSelectionChange={(key) => {
                  const next = Number(key);
                  if (Number.isFinite(next)) applyRate(next);
                }}
              />
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
              {pipSupported && (
                <button
                  type="button"
                  className="lesson-player-iconbtn"
                  aria-label={isPictureInPicture ? t('learn.video.pipExit') : t('learn.video.pip')}
                  aria-pressed={isPictureInPicture}
                  onClick={togglePictureInPicture}
                >
                  <svg viewBox="0 0 24 24" width="20" height="20" aria-hidden="true" fill="none">
                    <rect
                      x="3"
                      y="4"
                      width="18"
                      height="14"
                      rx="2"
                      stroke="currentColor"
                      strokeWidth="1.6"
                    />
                    <rect x="11" y="10" width="8" height="6" rx="1" fill="currentColor" />
                  </svg>
                </button>
              )}
              <button
                type="button"
                className="lesson-player-iconbtn"
                aria-label={
                  isFullscreen ? t('learn.video.fullscreenExit') : t('learn.video.fullscreen')
                }
                aria-pressed={isFullscreen}
                onClick={toggleFullscreen}
              >
                <svg
                  viewBox="0 0 24 24"
                  width="20"
                  height="20"
                  aria-hidden="true"
                  fill="none"
                  stroke="currentColor"
                  strokeWidth="1.8"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                >
                  {isFullscreen ? (
                    <path d="M9 4v5H4m11-5v5h5M9 20v-5H4m11 5v-5h5" />
                  ) : (
                    <path d="M4 9V4h5M20 9V4h-5M4 15v5h5m11-5v5h-5" />
                  )}
                </svg>
              </button>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
