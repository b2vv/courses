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
  CreatePlaybackMedia,
  PlaybackMedia,
  PlaybackMediaCallbacks,
} from '../ports/PlaybackMedia';
import { LessonPlayer, type LessonPlayerActions } from './LessonPlayer';

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
) {
  const mediaFactory: CreatePlaybackMedia =
    createMedia ??
    (() => {
      throw new Error('Ready player test must provide a media factory');
    });
  return render(
    <I18nProvider instance={createI18n('en')}>
      <LessonPlayer
        courseTitle="Reliable Go"
        lesson={lesson(state)}
        actions={playerActions}
        createMedia={mediaFactory}
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
          registeredCallbacks.onReady();
        },
        stopLoad() {},
        selectLevel(level) {
          selectedLevels.push(level);
        },
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

    const quality = screen.getByRole('button', { name: /Quality/ });
    quality.focus();
    await user.keyboard('{Enter}');
    await screen.findByRole('listbox');
    screen.getByRole('option', { name: 'Auto' }).focus();
    await user.keyboard('{ArrowDown}{Enter}');
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
          callbacks.onReady();
        },
        stopLoad() {},
        selectLevel() {},
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
          callbacks.onReady();
        },
        stopLoad() {},
        selectLevel() {},
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
          registeredCallbacks.onReady();
        },
        stopLoad() {},
        selectLevel() {},
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
          callbacks.onReady();
        },
        stopLoad() {},
        selectLevel() {},
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
          callbacks.onReady();
        },
        stopLoad() {},
        selectLevel() {},
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
          callbacks.onReady();
        },
        stopLoad() {},
        selectLevel() {},
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
          callbacks.onReady();
        },
        stopLoad() {},
        selectLevel() {},
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
          callbacks.onReady();
        },
        stopLoad() {},
        selectLevel() {},
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
          callbacks.onReady();
        },
        stopLoad() {},
        selectLevel() {},
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
          callbacks.onReady();
        },
        stopLoad() {},
        selectLevel() {},
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
