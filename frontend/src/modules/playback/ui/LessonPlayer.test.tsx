import { afterEach, describe, expect, it, rstest } from '@rstest/core';
import { act, fireEvent, render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import axe from 'axe-core';
import { createI18n, I18nProvider } from '../../../app/providers/i18n';
import type {
  LearningLesson,
  PlaybackHeartbeat,
  PlaybackResult,
  PlaybackSession,
} from '../domain/playback';
import type {
  CaptionSelection,
  CreatePlaybackMedia,
  PlaybackMedia,
  PlaybackMediaCallbacks,
} from '../ports/PlaybackMedia';
import { LessonPlayer, type LessonPlayerActions } from './LessonPlayer';
import { predictedAudioLanguage } from '../application/audio-language';

const mount = Symbol('player-test');
const lessonId = '30000000-0000-4000-8000-000000000001';
const playbackId = '0123456789abcdef0123456789abcdef';
const session: PlaybackSession = {
  playbackId,
  mediaPath: `/media/p/${playbackId}/hls/50000000-0000-4000-8000-000000000001/master.m3u8`,
  clientSeq: 4,
  heartbeatSeconds: 30,
  ownershipTtlSeconds: 180,
};

function lesson(videoState: LearningLesson['videoState']): LearningLesson {
  return {
    id: lessonId,
    title: 'Ownership and recovery',
    videoState,
    position: 12,
    completed: false,
    clientSeq: 4,
    audioTracks: [],
    isPreview: false,
    locked: false,
  };
}

function actions(overrides: Partial<LessonPlayerActions> = {}): LessonPlayerActions {
  return {
    play: async () => ({ ok: true, value: session }),
    heartbeat: async (_heartbeat: PlaybackHeartbeat): Promise<PlaybackResult<void>> => ({
      ok: true,
      value: undefined,
    }),
    refresh: async () => {},
    ...overrides,
  };
}

function renderPlayer(
  state: LearningLesson['videoState'],
  playerActions = actions(),
  createMedia?: CreatePlaybackMedia,
  onCaptionSelection?: (event: { generation: number; selection: CaptionSelection }) => void,
) {
  const mediaFactory: CreatePlaybackMedia =
    createMedia ??
    (() => {
      throw new Error('Ready player test must provide a media factory');
    });
  return render(
    <I18nProvider instance={createI18n('en')}>
      <LessonPlayer
        mount={mount}
        courseTitle="Reliable Go"
        lesson={lesson(state)}
        actions={playerActions}
        createMedia={mediaFactory}
        {...(onCaptionSelection === undefined ? {} : { onCaptionSelection })}
      />
    </I18nProvider>,
  );
}

afterEach(() => {
  localStorage.clear();
  rstest.useRealTimers();
});

describe('lesson player states', () => {
  it.each([
    [null, 'This lesson has no video yet.'],
    ['processing', 'The video is being prepared.'],
    ['failed', 'The video could not be prepared.'],
  ] as const)('B02 B03 B04 renders %s as a state, not a broken video', (state, copy) => {
    const { container } = renderPlayer(state);
    expect(screen.getByText(copy)).toBeTruthy();
    expect(container.querySelector('video')).toBeNull();
  });

  it('B24 keeps play, seek and hls.js quality keyboard reachable and axe-clean', async () => {
    const user = userEvent.setup();
    const loads: string[] = [];
    const selectedLevels: Array<number | 'auto'> = [];
    let callbacks: PlaybackMediaCallbacks | undefined;
    const createMedia: CreatePlaybackMedia = (video, registeredCallbacks) => {
      callbacks = registeredCallbacks;
      Object.defineProperty(video, 'duration', {
        configurable: true,
        value: 100,
      });
      const media: PlaybackMedia = {
        kind: 'hls',
        load(source) {
          loads.push(source);
          registeredCallbacks.onLevels([
            { id: 0, label: '360p' },
            { id: 1, label: '720p' },
          ]);
          registeredCallbacks.onCaptionSelection({ state: 'pending' });
          registeredCallbacks.onReady();
        },
        stopLoad() {},
        selectLevel(level) {
          selectedLevels.push(level);
        },
        selectAudioTrack: () => false,
        destroy() {},
      };
      return media;
    };
    let plays = 0;
    let pauses = 0;
    const playerActions = actions({
      play: async () => {
        plays += 1;
        return { ok: true, value: session };
      },
    });
    const { container } = renderPlayer('ready', playerActions, createMedia);
    const video = container.querySelector('video');
    if (video === null) throw new Error('Ready lesson did not render a video element');
    // The stub has to carry `paused` as the element does. The player refuses to raise its
    // "this pause is ours" flag for media that is already paused - otherwise a pause that fires
    // no event leaves the flag up and swallows the next real one - and a stub whose `pause()`
    // never changes `paused` reports a state no element is ever in.
    let paused = true;
    Object.defineProperty(video, 'paused', { configurable: true, get: () => paused });
    Object.defineProperty(video, 'play', {
      configurable: true,
      value: () => {
        paused = false;
        return Promise.resolve();
      },
    });
    Object.defineProperty(video, 'pause', {
      configurable: true,
      value: () => {
        paused = true;
        pauses += 1;
      },
    });

    expect(plays).toBe(0);
    expect(loads).toEqual([]);
    // The scrubber sits at the top of the control bar, so it is the first tab stop; Play follows.
    await user.tab();
    expect(screen.getByRole('slider', { name: 'Playback position' })).toBe(document.activeElement);
    await user.tab();
    expect(screen.getByRole('button', { name: 'Play' })).toBe(document.activeElement);
    await user.keyboard('{Enter}');
    await waitFor(() => expect(plays).toBe(1));
    expect(loads).toEqual([session.mediaPath]);
    expect(video.currentTime).toBe(12);
    fireEvent.playing(video);
    await waitFor(() => expect(screen.getByRole('button', { name: 'Pause' })).toBeTruthy());
    await user.keyboard('{Enter}');
    expect(pauses).toBe(1);

    const seek = screen.getByRole('slider', { name: 'Playback position' });
    seek.focus();
    await user.keyboard('{ArrowRight}');
    expect(video.currentTime).toBeGreaterThan(12);

    // Speed, quality and captions now live behind one keyboard-reachable gear. Open it and pick the
    // 360p rendition; the quality radio group drives the same hls.js level switch the Select did.
    const settings = screen.getByRole('button', { name: /Settings/ });
    settings.focus();
    expect(settings).toBe(document.activeElement);
    await user.keyboard('{Enter}');
    const quality360 = await screen.findByRole('radio', { name: '360p' });
    await user.click(quality360);
    await waitFor(() => expect(selectedLevels).toContain(0));
    expect(callbacks).toBeDefined();
    expect((await axe.run(container)).violations).toEqual([]);
  });

  it('keeps one media adapter across time updates and restores before duration is known', async () => {
    const user = userEvent.setup();
    let creations = 0;
    const createMedia: CreatePlaybackMedia = (video, callbacks) => {
      creations += 1;
      Object.defineProperty(video, 'play', {
        configurable: true,
        value: () => Promise.resolve(),
      });
      return {
        kind: 'hls',
        load() {
          callbacks.onCaptionSelection({ state: 'pending' });
          callbacks.onReady();
        },
        stopLoad() {},
        selectLevel() {},
        selectAudioTrack: () => false,
        destroy() {},
      };
    };
    const { container } = renderPlayer('ready', actions(), createMedia);
    const video = container.querySelector('video');
    if (video === null) throw new Error('Ready lesson did not render a video element');

    await user.click(screen.getByRole('button', { name: 'Play' }));
    await waitFor(() => expect(video.currentTime).toBe(12));
    video.currentTime = 18;
    fireEvent.timeUpdate(video);
    await waitFor(() => expect(creations).toBe(1));
  });

  it('continues heartbeats after an expired session is silently reminted', async () => {
    rstest.useFakeTimers();
    const expiringSession = {
      ...session,
      heartbeatSeconds: 1,
      ownershipTtlSeconds: 3,
    };
    let plays = 0;
    let heartbeats = 0;
    const playerActions = actions({
      play: async () => {
        plays += 1;
        return { ok: true, value: expiringSession };
      },
      heartbeat: async () => {
        heartbeats += 1;
        return heartbeats === 1
          ? { ok: false, error: { code: 'playback_session_expired' } }
          : { ok: true, value: undefined };
      },
    });
    const createMedia: CreatePlaybackMedia = (video, callbacks) => {
      Object.defineProperty(video, 'play', {
        configurable: true,
        value: () => Promise.resolve(),
      });
      return {
        kind: 'hls',
        load() {
          callbacks.onCaptionSelection({ state: 'pending' });
          callbacks.onReady();
        },
        stopLoad() {},
        selectLevel() {},
        selectAudioTrack: () => false,
        destroy() {},
      };
    };
    const { container } = renderPlayer('ready', playerActions, createMedia);
    const video = container.querySelector('video');
    if (video === null) throw new Error('Ready lesson did not render a video element');
    fireEvent.click(screen.getByRole('button', { name: 'Play' }));
    await act(async () => {});
    fireEvent.playing(video);
    expect(plays).toBe(1);

    await act(async () => rstest.advanceTimersByTimeAsync(1_000));
    expect(plays).toBe(2);
    expect(heartbeats).toBe(1);
    await act(async () => rstest.advanceTimersByTimeAsync(1_000));
    expect(heartbeats).toBe(2);
  });

  it('bounds consecutive media retries even when each manifest parses', async () => {
    rstest.useFakeTimers();
    let callbacks: PlaybackMediaCallbacks | undefined;
    let loads = 0;
    const createMedia: CreatePlaybackMedia = (video, registeredCallbacks) => {
      callbacks = registeredCallbacks;
      Object.defineProperty(video, 'play', {
        configurable: true,
        value: () => Promise.resolve(),
      });
      return {
        kind: 'hls',
        load() {
          loads += 1;
          registeredCallbacks.onCaptionSelection({ state: 'pending' });
          registeredCallbacks.onReady();
        },
        stopLoad() {},
        selectLevel() {},
        selectAudioTrack: () => false,
        destroy() {},
      };
    };
    renderPlayer('ready', actions(), createMedia);
    fireEvent.click(screen.getByRole('button', { name: 'Play' }));
    await act(async () => {});
    if (callbacks === undefined) throw new Error('Media callbacks were not registered');

    for (const delay of [500, 1_000, 2_000]) {
      callbacks.onError({ status: 503, retryable: true, url: undefined });
      await act(async () => rstest.advanceTimersByTimeAsync(delay));
    }
    callbacks.onError({ status: 503, retryable: true, url: undefined });
    await act(async () => {});

    expect(loads).toBe(4);
    expect(
      screen.getByRole('heading', {
        name: 'This video is unavailable right now.',
      }),
    ).toBeTruthy();
  });

  it('keeps one mint cycle across dependency backoff and loads the successful response', async () => {
    rstest.useFakeTimers();
    let plays = 0;
    const loads: string[] = [];
    const playerActions = actions({
      play: async () => {
        plays += 1;
        return plays === 1
          ? { ok: false, error: { code: 'dependency_unavailable' } }
          : { ok: true, value: session };
      },
    });
    const createMedia: CreatePlaybackMedia = (video, callbacks) => {
      Object.defineProperty(video, 'play', {
        configurable: true,
        value: () => Promise.resolve(),
      });
      return {
        kind: 'hls',
        load(source) {
          loads.push(source);
          callbacks.onCaptionSelection({ state: 'pending' });
          callbacks.onReady();
        },
        stopLoad() {},
        selectLevel() {},
        selectAudioTrack: () => false,
        destroy() {},
      };
    };
    const { container } = renderPlayer('ready', playerActions, createMedia);
    const video = container.querySelector('video');
    if (video === null) throw new Error('Ready lesson did not render a video element');
    fireEvent.click(screen.getByRole('button', { name: 'Play' }));
    await act(async () => {});
    fireEvent.playing(video);
    expect(plays).toBe(1);
    expect(loads).toEqual([]);

    await act(async () => rstest.advanceTimersByTimeAsync(500));
    expect(plays).toBe(2);
    expect(loads).toEqual([session.mediaPath]);
  });

  it('remounts media before an explicit retry after terminal 403', async () => {
    let plays = 0;
    const loads: string[] = [];
    const playerActions = actions({
      play: async () => {
        plays += 1;
        return plays === 1
          ? { ok: false, error: { code: 'playback_forbidden' } }
          : { ok: true, value: session };
      },
    });
    const createMedia: CreatePlaybackMedia = (video, callbacks) => {
      Object.defineProperty(video, 'play', {
        configurable: true,
        value: () => Promise.resolve(),
      });
      return {
        kind: 'hls',
        load(source) {
          loads.push(source);
          callbacks.onCaptionSelection({ state: 'pending' });
          callbacks.onReady();
        },
        stopLoad() {},
        selectLevel() {},
        selectAudioTrack: () => false,
        destroy() {},
      };
    };
    renderPlayer('ready', playerActions, createMedia);
    fireEvent.click(screen.getByRole('button', { name: 'Play' }));
    await screen.findByRole('heading', {
      name: 'Playback stopped. This course is no longer available.',
    });
    expect(loads).toEqual([]);

    fireEvent.click(screen.getByRole('button', { name: 'Start playback again' }));
    await waitFor(() => expect(plays).toBe(2));
    expect(loads).toEqual([session.mediaPath]);
  });

  it('B09 sends exactly three tick heartbeats across 90 seconds of continuous playback', async () => {
    rstest.useFakeTimers();
    const heartbeats: PlaybackHeartbeat[] = [];
    const cadenceSession = { ...session, clientSeq: 12 };
    const playerActions = actions({
      play: async () => ({ ok: true, value: cadenceSession }),
      heartbeat: async (heartbeat) => {
        heartbeats.push(heartbeat);
        await new Promise<void>((resolve) => setTimeout(resolve, 1_000));
        return { ok: true, value: undefined };
      },
    });
    const createMedia: CreatePlaybackMedia = (video, callbacks) => {
      Object.defineProperty(video, 'play', {
        configurable: true,
        value: () => Promise.resolve(),
      });
      return {
        kind: 'hls',
        load() {
          callbacks.onCaptionSelection({ state: 'pending' });
          callbacks.onReady();
        },
        stopLoad() {},
        selectLevel() {},
        selectAudioTrack: () => false,
        destroy() {},
      };
    };
    const { container } = renderPlayer('ready', playerActions, createMedia);
    const video = container.querySelector('video');
    if (video === null) throw new Error('Ready lesson did not render a video element');
    fireEvent.click(screen.getByRole('button', { name: 'Play' }));
    await act(async () => {});
    fireEvent.playing(video);

    await act(async () => rstest.advanceTimersByTimeAsync(90_000));

    expect(heartbeats).toHaveLength(3);
    expect(heartbeats.map(({ reason }) => reason)).toEqual(['tick', 'tick', 'tick']);
    expect(heartbeats.map(({ clientSeq }) => clientSeq)).toEqual([13, 14, 15]);
  });

  it('B-7 follows native playing and pause events for ownership cadence', async () => {
    rstest.useFakeTimers();
    const heartbeats: PlaybackHeartbeat[] = [];
    const playerActions = actions({
      heartbeat: async (heartbeat) => {
        heartbeats.push(heartbeat);
        return { ok: true, value: undefined };
      },
    });
    const createMedia: CreatePlaybackMedia = (video, callbacks) => {
      Object.defineProperty(video, 'play', {
        configurable: true,
        value: () => Promise.resolve(),
      });
      return {
        kind: 'hls',
        load() {
          callbacks.onCaptionSelection({ state: 'pending' });
          callbacks.onReady();
        },
        stopLoad() {},
        selectLevel() {},
        selectAudioTrack: () => false,
        destroy() {},
      };
    };
    const { container } = renderPlayer('ready', playerActions, createMedia);
    const video = container.querySelector('video');
    if (video === null) throw new Error('Ready lesson did not render a video element');

    fireEvent.click(screen.getByRole('button', { name: 'Play' }));
    await act(async () => {});
    await act(async () => rstest.advanceTimersByTimeAsync(30_000));
    expect(heartbeats).toHaveLength(0);

    fireEvent.playing(video);
    await act(async () => rstest.advanceTimersByTimeAsync(30_000));
    expect(heartbeats).toHaveLength(1);

    // A pause from outside our controls does two things, and both matter. It stops the cadence,
    // so ownership is no longer renewed by a tab that plays nothing - that is B-7. And it writes
    // the exit itself, because nobody else will: the button's write lives in the button, so
    // without this the student loses everything watched since the last beat - that is R39.
    fireEvent.pause(video);
    await act(async () => rstest.advanceTimersByTimeAsync(60_000));
    expect(heartbeats).toHaveLength(2);
    expect(heartbeats[1]?.reason).toBe('pause');

    fireEvent.playing(video);
    await act(async () => rstest.advanceTimersByTimeAsync(30_000));
    expect(heartbeats).toHaveLength(3);
  });

  it('B10 writes pause, seek, hide, nav and ended exits and hidden time does not tick', async () => {
    rstest.useFakeTimers();
    const heartbeats: PlaybackHeartbeat[] = [];
    const playerActions = actions({
      heartbeat: async (heartbeat) => {
        heartbeats.push(heartbeat);
        return { ok: true, value: undefined };
      },
    });
    const createMedia: CreatePlaybackMedia = (video, callbacks) => {
      // `paused` is modelled for the same reason as in B24: the player will not flag a pause as
      // its own on media that is already paused.
      let paused = true;
      Object.defineProperty(video, 'paused', { configurable: true, get: () => paused });
      Object.defineProperty(video, 'play', {
        configurable: true,
        value: () => {
          paused = false;
          return Promise.resolve();
        },
      });
      Object.defineProperty(video, 'pause', {
        configurable: true,
        value: () => {
          paused = true;
          video.dispatchEvent(new Event('pause'));
        },
      });
      Object.defineProperty(video, 'duration', {
        configurable: true,
        value: 100,
      });
      return {
        kind: 'hls',
        load() {
          callbacks.onCaptionSelection({ state: 'pending' });
          callbacks.onReady();
        },
        stopLoad() {},
        selectLevel() {},
        selectAudioTrack: () => false,
        destroy() {},
      };
    };
    const rendered = renderPlayer('ready', playerActions, createMedia);
    const video = rendered.container.querySelector('video');
    if (video === null) throw new Error('Ready lesson did not render a video element');
    fireEvent.click(screen.getByRole('button', { name: 'Play' }));
    await act(async () => {});
    fireEvent.playing(video);

    fireEvent.change(screen.getByRole('slider', { name: 'Playback position' }), {
      target: { value: '40' },
    });
    await act(async () => {});
    fireEvent.click(screen.getByRole('button', { name: 'Pause' }));
    await act(async () => {});
    fireEvent.click(screen.getByRole('button', { name: 'Play' }));
    await act(async () => {});
    Object.defineProperty(document, 'visibilityState', {
      configurable: true,
      value: 'hidden',
    });
    fireEvent(document, new Event('visibilitychange'));
    await act(async () => rstest.advanceTimersByTimeAsync(90_000));
    fireEvent.ended(video);
    await act(async () => {});
    rendered.unmount();
    await act(async () => {});
    Reflect.deleteProperty(document, 'visibilityState');

    expect(heartbeats.map(({ reason }) => reason).filter((reason) => reason !== 'tick')).toEqual([
      'seek',
      'pause',
      'hide',
      'ended',
      'nav',
    ]);
    expect(heartbeats.filter(({ reason }) => reason === 'tick')).toHaveLength(0);
  });

  it('B31 recreates media on the remounted video before retrying after takeover', async () => {
    rstest.useFakeTimers();
    const takeoverSession = {
      ...session,
      heartbeatSeconds: 1,
      ownershipTtlSeconds: 3,
    };
    const createdVideos: HTMLVideoElement[] = [];
    const playedVideos: HTMLVideoElement[] = [];
    const destroyedVideos: HTMLVideoElement[] = [];
    let plays = 0;
    const playerActions = actions({
      play: async () => {
        plays += 1;
        return { ok: true, value: takeoverSession };
      },
      heartbeat: async () => ({ ok: false, error: { code: 'playback_taken_over' } }),
    });
    const createMedia: CreatePlaybackMedia = (video, callbacks) => {
      createdVideos.push(video);
      Object.defineProperty(video, 'duration', {
        configurable: true,
        value: 100,
      });
      Object.defineProperty(video, 'play', {
        configurable: true,
        value: () => {
          playedVideos.push(video);
          return Promise.resolve();
        },
      });
      return {
        kind: 'hls',
        load() {
          callbacks.onCaptionSelection({ state: 'pending' });
          callbacks.onReady();
        },
        stopLoad() {},
        selectLevel() {},
        selectAudioTrack: () => false,
        destroy() {
          destroyedVideos.push(video);
        },
      };
    };
    const rendered = renderPlayer('ready', playerActions, createMedia);
    const firstVideo = rendered.container.querySelector('video');
    if (firstVideo === null) throw new Error('Ready lesson did not render a video element');

    fireEvent.click(screen.getByRole('button', { name: 'Play' }));
    await act(async () => {});
    // Widened on purpose: as the literal type 47 the guard below is a TS2367 error, and removing
    // it lets a later fixture edit turn the position assertion into 12 === 12 without notice.
    const resumePosition: number = 47;
    if (resumePosition === 0 || resumePosition === lesson('ready').position)
      throw new Error(
        'B31 resume position equals the fixture start point; the position assertion would be vacuous',
      );
    firstVideo.currentTime = resumePosition;
    fireEvent.timeUpdate(firstVideo);
    const positionControl = screen.getByRole('slider', { name: 'Playback position' });
    if (!(positionControl instanceof HTMLInputElement))
      throw new Error('Playback position is not an input');
    expect(positionControl.value).toBe(String(resumePosition));
    fireEvent.playing(firstVideo);
    await act(async () => rstest.advanceTimersByTimeAsync(1_000));
    screen.getByRole('heading', {
      name: 'This video is playing on another device.',
    });

    fireEvent.click(screen.getByRole('button', { name: 'Start playback again' }));
    await act(async () => {});
    const remountedVideo = rendered.container.querySelector('video');
    if (remountedVideo === null || remountedVideo === firstVideo)
      throw new Error('Retry did not remount a new video element');

    expect(createdVideos).toHaveLength(2);
    expect(createdVideos[1]).toBe(remountedVideo);
    expect(destroyedVideos).toEqual([firstVideo]);
    expect(playedVideos).toEqual([firstVideo, remountedVideo]);
    expect(plays).toBe(2);
    expect(remountedVideo.currentTime).toBe(resumePosition);
  });

  it('R28 stops scheduling and sending heartbeats after a 409 takeover', async () => {
    rstest.useFakeTimers();
    let heartbeats = 0;
    const playerActions = actions({
      heartbeat: async () => {
        heartbeats += 1;
        return { ok: false, error: { code: 'playback_taken_over' } };
      },
    });
    const createMedia: CreatePlaybackMedia = (video, callbacks) => {
      Object.defineProperty(video, 'play', {
        configurable: true,
        value: () => Promise.resolve(),
      });
      return {
        kind: 'hls',
        load() {
          callbacks.onCaptionSelection({ state: 'pending' });
          callbacks.onReady();
        },
        stopLoad() {},
        selectLevel() {},
        selectAudioTrack: () => false,
        destroy() {},
      };
    };
    const { container } = renderPlayer('ready', playerActions, createMedia);
    const video = container.querySelector('video');
    if (video === null) throw new Error('Ready lesson did not render a video element');
    fireEvent.click(screen.getByRole('button', { name: 'Play' }));
    await act(async () => {});
    fireEvent.playing(video);

    await act(async () => rstest.advanceTimersByTimeAsync(30_000));
    await act(async () => rstest.advanceTimersByTimeAsync(90_000));

    expect(heartbeats).toBe(1);
    expect(
      screen.getByRole('heading', {
        name: 'This video is playing on another device.',
      }),
    ).toBeTruthy();
  });
});

describe('A2 Up Next autoplay', () => {
  // The countdown the player renders before it advances (AUTOPLAY_SECONDS in LessonPlayer).
  const AUTOPLAY_MS = 5_000;

  function renderWithNext(onPlayNext: () => void) {
    const createMedia: CreatePlaybackMedia = (video, callbacks) => {
      Object.defineProperty(video, 'duration', { configurable: true, value: 100 });
      Object.defineProperty(video, 'play', {
        configurable: true,
        value: () => Promise.resolve(),
      });
      return {
        kind: 'hls',
        load() {
          callbacks.onCaptionSelection({ state: 'pending' });
          callbacks.onReady();
        },
        stopLoad() {},
        selectLevel() {},
        selectAudioTrack: () => false,
        destroy() {},
      };
    };
    return render(
      <I18nProvider instance={createI18n('en')}>
        <LessonPlayer
          mount={mount}
          courseTitle="Reliable Go"
          lesson={lesson('ready')}
          actions={actions()}
          createMedia={createMedia}
          nextLesson={{ id: 'next-id', title: 'The next lesson' }}
          onPlayNext={onPlayNext}
        />
      </I18nProvider>,
    );
  }

  async function playToEnd(container: HTMLElement) {
    const video = container.querySelector('video');
    if (video === null) throw new Error('Ready lesson did not render a video element');
    fireEvent.click(screen.getByRole('button', { name: 'Play' }));
    await act(async () => {});
    fireEvent.playing(video);
    fireEvent.ended(video);
    await act(async () => {});
    return video;
  }

  it('offers an Up Next card at the end and Play now advances, axe-clean', async () => {
    const onPlayNext = rstest.fn();
    const { container } = renderWithNext(onPlayNext);
    await playToEnd(container);

    expect(screen.getByText('The next lesson')).toBeTruthy();
    expect((await axe.run(container)).violations).toEqual([]);

    fireEvent.click(screen.getByRole('button', { name: 'Play now' }));
    expect(onPlayNext).toHaveBeenCalledTimes(1);
  });

  it('advances automatically once the countdown elapses', async () => {
    rstest.useFakeTimers();
    const onPlayNext = rstest.fn();
    const { container } = renderWithNext(onPlayNext);
    await playToEnd(container);

    expect(onPlayNext).not.toHaveBeenCalled();
    await act(async () => rstest.advanceTimersByTimeAsync(AUTOPLAY_MS));
    expect(onPlayNext).toHaveBeenCalledTimes(1);
  });

  it('Cancel stops the pending autoplay without advancing, leaving Replay', async () => {
    rstest.useFakeTimers();
    const onPlayNext = rstest.fn();
    const { container } = renderWithNext(onPlayNext);
    await playToEnd(container);

    fireEvent.click(screen.getByRole('button', { name: 'Cancel' }));
    await act(async () => rstest.advanceTimersByTimeAsync(AUTOPLAY_MS * 3));
    expect(onPlayNext).not.toHaveBeenCalled();
    // The ticker gone, the secondary action is now Replay; the lesson can still be rewatched.
    expect(screen.getByRole('button', { name: 'Replay' })).toBeTruthy();
    expect(screen.getByRole('button', { name: 'Play now' })).toBeTruthy();
  });
});

describe('MS4 caption selection channel', () => {
  it('increments before every load, including media retries and session remints', async () => {
    rstest.useFakeTimers();
    const events: Array<{ generation: number; selection: CaptionSelection }> = [];
    let registered: PlaybackMediaCallbacks | undefined;
    let loads = 0;
    const createMedia: CreatePlaybackMedia = (video, callbacks) => {
      registered = callbacks;
      Object.defineProperty(video, 'play', { configurable: true, value: () => Promise.resolve() });
      return {
        kind: 'hls',
        load() {
          loads += 1;
          callbacks.onCaptionSelection({ state: 'pending' });
          callbacks.onReady();
        },
        stopLoad() {},
        selectLevel() {},
        selectAudioTrack: () => false,
        destroy() {},
      };
    };
    const playerActions = actions({
      play: async () => ({ ok: true, value: { ...session, heartbeatSeconds: 1 } }),
      heartbeat: async () => ({ ok: false, error: { code: 'playback_session_expired' } }),
    });
    const { container } = renderPlayer('ready', playerActions, createMedia, (event) =>
      events.push(event),
    );
    fireEvent.click(screen.getByRole('button', { name: 'Play' }));
    await act(async () => {});
    if (registered === undefined) throw new Error('Media callbacks not registered');
    expect(events).toEqual([{ mount, generation: 1, selection: { state: 'pending' } }]);
    act(() => registered?.onCaptionSelection({ state: 'confirmed', language: 'en' }));
    expect(events.at(-1)).toEqual({
      mount,
      generation: 1,
      selection: { state: 'confirmed', language: 'en' },
    });
    registered.onError({ status: 503, retryable: true, url: undefined });
    await act(async () => rstest.advanceTimersByTimeAsync(500));
    expect(events.at(-1)).toEqual({ mount, generation: 2, selection: { state: 'pending' } });
    const video = container.querySelector('video');
    if (video === null) throw new Error('Missing video');
    fireEvent.playing(video);
    await act(async () => rstest.advanceTimersByTimeAsync(1_000));
    expect(loads).toBe(3);
    expect(events.at(-1)).toEqual({ mount, generation: 3, selection: { state: 'pending' } });
  });
});

describe('multi-audio S7: the audio track menu', () => {
  type AudioFake = {
    readonly selected: string[];
    readonly options: Array<{ audioLanguage: string | null } | undefined>;
    callbacks?: PlaybackMediaCallbacks;
  };

  function audioMedia(
    tracks: ReadonlyArray<{ language: string; label: string; isDefault: boolean }>,
    kind: 'hls' | 'native' = 'hls',
    accept: (language: string) => boolean = () => true,
  ) {
    const fake: AudioFake = { selected: [], options: [] };
    const createMedia: CreatePlaybackMedia = (video, callbacks, options) => {
      fake.callbacks = callbacks;
      fake.options.push(options);
      Object.defineProperty(video, 'duration', { configurable: true, value: 100 });
      Object.defineProperty(video, 'play', {
        configurable: true,
        value: () => Promise.resolve(),
      });
      return {
        kind,
        load() {
          callbacks.onAudioTracks(tracks);
          const start = predictedAudioLanguage(tracks, options?.audioLanguage ?? null);
          if (start !== null) callbacks.onAudioTrack(start);
          callbacks.onCaptionSelection({ state: 'pending' });
          callbacks.onReady();
        },
        stopLoad() {},
        selectLevel() {},
        selectAudioTrack(language) {
          fake.selected.push(language);
          if (!accept(language)) return false;
          callbacks.onAudioTrack(language);
          return true;
        },
        destroy() {},
      };
    };
    return { fake, createMedia };
  }

  const twoTracks = [
    { language: 'und', label: 'Original', isDefault: true },
    { language: 'en', label: 'English dub', isDefault: false },
  ];

  async function openSettings(user: ReturnType<typeof userEvent.setup>) {
    await user.click(screen.getByRole('button', { name: 'Play' }));
    await user.click(await screen.findByRole('button', { name: /Settings/ }));
  }

  function storedAudioLanguage(): unknown {
    const raw = localStorage.getItem('courses.player.prefs.v1');
    return raw === null
      ? undefined
      : (JSON.parse(raw) as { audioLanguage?: unknown }).audioLanguage;
  }

  it('offers two tracks, switches on choice and remembers the language for the next lesson', async () => {
    const user = userEvent.setup();
    const { fake, createMedia } = audioMedia(twoTracks);
    const first = renderPlayer('ready', actions(), createMedia);
    expect(fake.options[0]).toEqual({ audioLanguage: null });
    await openSettings(user);
    const audio = await screen.findByRole('radiogroup', { name: 'Audio' });
    expect(audio).toBeTruthy();
    const original = screen.getByRole('radio', { name: 'Original' });
    expect((original as HTMLInputElement).checked).toBe(true);

    await user.click(screen.getByRole('radio', { name: 'English dub' }));
    expect(fake.selected).toEqual(['en']);
    await waitFor(() =>
      expect((screen.getByRole('radio', { name: 'English dub' }) as HTMLInputElement).checked).toBe(
        true,
      ),
    );
    await waitFor(() => expect(storedAudioLanguage()).toBe('en'));
    first.unmount();

    const next = audioMedia(twoTracks);
    renderPlayer('ready', actions(), next.createMedia);
    expect(next.fake.options[0]).toEqual({ audioLanguage: 'en' });
  });

  it('keeps the selection and the preference when the manifest has no such track', async () => {
    const user = userEvent.setup();
    const { fake, createMedia } = audioMedia(twoTracks, 'hls', () => false);
    renderPlayer('ready', actions(), createMedia);
    await openSettings(user);
    await user.click(await screen.findByRole('radio', { name: 'English dub' }));
    expect(fake.selected).toEqual(['en']);
    expect((screen.getByRole('radio', { name: 'Original' }) as HTMLInputElement).checked).toBe(
      true,
    );
    expect(storedAudioLanguage() ?? null).toBeNull();
  });

  it.each([
    ['no track', [], 'hls'],
    ['one track', twoTracks.slice(0, 1), 'hls'],
    ['native playback', twoTracks, 'native'],
  ] as const)('shows no audio group for %s', async (_label, tracks, kind) => {
    const user = userEvent.setup();
    const { createMedia } = audioMedia(tracks, kind);
    renderPlayer('ready', actions(), createMedia);
    await openSettings(user);
    await screen.findByRole('radiogroup', { name: 'Speed' });
    expect(screen.queryByRole('radiogroup', { name: 'Audio' })).toBeNull();
  });

  it('names every option by its own label when none collide', async () => {
    const user = userEvent.setup();
    const { createMedia } = audioMedia(twoTracks);
    renderPlayer('ready', actions(), createMedia);
    await openSettings(user);
    expect(await screen.findByRole('radio', { name: 'Original' })).toBeTruthy();
    expect(screen.getByRole('radio', { name: 'English dub' })).toBeTruthy();
  });

  it('tells apart tracks that share a name and renders labels as text', async () => {
    const user = userEvent.setup();
    const { createMedia } = audioMedia([
      { language: 'uk', label: 'Dub', isDefault: true },
      { language: 'en', label: 'Dub', isDefault: false },
      { language: 'de', label: '<b>Dub</b>', isDefault: false },
      { language: 'fr', label: ' ', isDefault: false },
      // Already shaped like a disambiguated label: must not collide with "Dub (uk)".
      { language: 'pl', label: 'Dub (uk)', isDefault: false },
    ]);
    renderPlayer('ready', actions(), createMedia);
    await openSettings(user);
    expect(await screen.findByRole('radio', { name: 'Dub (uk)' })).toBeTruthy();
    expect(screen.getByRole('radio', { name: 'Dub (en)' })).toBeTruthy();
    expect(screen.getByRole('radio', { name: '<b>Dub</b> (de)' })).toBeTruthy();
    expect(screen.getByRole('radio', { name: 'fr (fr)' })).toBeTruthy();
    expect(screen.getByRole('radio', { name: 'Dub (uk) (pl)' })).toBeTruthy();
    const names = screen
      .getAllByRole('radio')
      .map((radio) => radio.getAttribute('aria-label') ?? radio.closest('label')?.textContent);
    expect(new Set(names).size).toBe(names.length);
  });
});

describe('MS4 caption track lifecycle', () => {
  it('M07 recreates the language track, applies mode to its current ref and remembers captions through none', async () => {
    const trackModes = new WeakMap<HTMLTrackElement, { mode: TextTrackMode }>();
    const descriptor = Object.getOwnPropertyDescriptor(HTMLTrackElement.prototype, 'track');
    Object.defineProperty(HTMLTrackElement.prototype, 'track', {
      configurable: true,
      get(this: HTMLTrackElement) {
        let track = trackModes.get(this);
        if (track === undefined) {
          track = { mode: 'disabled' };
          trackModes.set(this, track);
        }
        return track;
      },
    });
    const staleTrack = { mode: 'disabled' };
    let callbacks: PlaybackMediaCallbacks | undefined;
    let creations = 0;
    const createMedia: CreatePlaybackMedia = (video, registered) => {
      callbacks = registered;
      creations += 1;
      Object.defineProperty(video, 'textTracks', { configurable: true, value: [staleTrack] });
      Object.defineProperty(video, 'play', { configurable: true, value: () => Promise.resolve() });
      return {
        kind: 'hls',
        load() {
          registered.onAudioTracks([
            { language: 'en', label: 'English', isDefault: true },
            { language: 'de', label: 'Deutsch', isDefault: false },
          ]);
          registered.onAudioTrack('en');
          registered.onCaptionSelection({ state: 'pending' });
        },
        stopLoad() {},
        selectLevel() {},
        selectAudioTrack: () => true,
        destroy() {},
      };
    };
    const playerActions = actions();
    const renderState = (language: string | undefined, none = false) => (
      <I18nProvider instance={createI18n('en')}>
        <LessonPlayer
          mount={mount}
          courseTitle="Go"
          lesson={lesson('ready')}
          actions={playerActions}
          createMedia={createMedia}
          {...(language === undefined ? {} : { captions: 'WEBVTT\n', captionsLanguage: language })}
          captionsNoneForTrack={none}
        />
      </I18nProvider>
    );
    try {
      const { container, rerender, unmount } = render(renderState('en'));
      const first = container.querySelector('track');
      if (first === null) throw new Error('Missing caption track');
      expect(first.getAttribute('srclang')).toBe('en');
      expect(trackModes.get(first)?.mode).toBe('hidden');
      expect(
        container.querySelector('.lesson-player')?.getAttribute('data-caption-selection'),
      ).toBe('pending');
      fireEvent.click(screen.getByRole('button', { name: 'Play' }));
      await act(async () => {});
      act(() => callbacks?.onCaptionSelection({ state: 'confirmed', language: 'EN' }));
      expect(
        container.querySelector('.lesson-player')?.getAttribute('data-caption-selection'),
      ).toBe('confirmed:en');
      await userEvent.click(screen.getByRole('button', { name: 'Settings' }));
      await userEvent.click(await screen.findByRole('switch', { name: 'Subtitles' }));
      expect(trackModes.get(first)?.mode).toBe('showing');
      expect(staleTrack.mode).toBe('disabled');
      rerender(renderState('de'));
      const second = container.querySelector('track');
      if (second === null) throw new Error('Missing replacement track');
      expect(second).not.toBe(first);
      expect(second.getAttribute('srclang')).toBe('de');
      expect(trackModes.get(second)?.mode).toBe('showing');
      expect(creations).toBe(1);
      rerender(renderState(undefined, true));
      expect(container.querySelector('track')).toBeNull();
      expect(await screen.findByText('Subtitles: none for this audio track')).toBeTruthy();
      expect(screen.queryByRole('switch', { name: 'Subtitles' })).toBeNull();
      await userEvent.click(screen.getByRole('radio', { name: 'Deutsch' }));
      expect(
        JSON.parse(localStorage.getItem('courses.player.prefs.v1') ?? '{}').audioLanguage,
      ).toBe('de');
      expect(JSON.parse(localStorage.getItem('courses.player.prefs.v1') ?? '{}').captionsOn).toBe(
        true,
      );
      act(() => callbacks?.onCaptionSelection({ state: 'unavailable' }));
      expect(
        container.querySelector('.lesson-player')?.getAttribute('data-caption-selection'),
      ).toBe('unavailable');
      rerender(renderState('en'));
      const restored = container.querySelector('track');
      if (restored === null) throw new Error('Missing restored track');
      expect(trackModes.get(restored)?.mode).toBe('showing');
      expect(screen.getByRole('switch', { name: 'Subtitles', checked: true })).toBeTruthy();
      expect(creations).toBe(1);
      unmount();
    } finally {
      if (descriptor) Object.defineProperty(HTMLTrackElement.prototype, 'track', descriptor);
      else Reflect.deleteProperty(HTMLTrackElement.prototype, 'track');
    }
  });
});
