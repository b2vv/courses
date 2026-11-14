import { describe, expect, it } from '@rstest/core';
import { render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import axe from 'axe-core';
import { createI18n, I18nProvider } from '../../../app/providers/i18n';
import type { AudioTrack, AudioTrackState } from '../domain/audio-track';
import type { LessonVideo } from '../domain/video';
import {
  LessonAudioTracksView,
  type CourseEditorAudio,
  type LessonAudioTracksActions,
  type LessonAudioTracksState,
} from './LessonAudioTracksField';
import { LessonVideoField, type LessonVideoFieldActions } from './LessonVideoField';

const lessonId = '20000000-0000-4000-8000-000000000001';
const videoId = '50000000-0000-4000-8000-000000000001';
const embeddedId = '60000000-0000-4000-8000-000000000001';
const dubId = '60000000-0000-4000-8000-000000000002';

function track(trackId: string, state: AudioTrack['state'], overrides: Partial<AudioTrack> = {}) {
  return {
    trackId,
    videoId,
    kind: 'uploaded',
    language: 'en',
    label: 'English',
    isDefault: false,
    state,
    updatedAt: '2026-11-13T12:00:00.000000Z',
    ...overrides,
  } satisfies AudioTrack;
}

const embedded = track(embeddedId, 'ready', {
  kind: 'embedded',
  language: 'uk',
  label: 'Українська',
  isDefault: true,
});

type Calls = { readonly [K in keyof LessonAudioTracksActions]: unknown[][] };

function actions(overrides: Partial<LessonAudioTracksActions> = {}) {
  const calls: Calls = {
    upload: [],
    resume: [],
    cancel: [],
    rename: [],
    setDefault: [],
    remove: [],
  };
  const record =
    <K extends keyof LessonAudioTracksActions>(name: K) =>
    async (...args: unknown[]) => {
      calls[name].push(args);
      return null;
    };
  const value: LessonAudioTracksActions = {
    upload: record('upload'),
    resume: record('resume'),
    cancel: record('cancel'),
    rename: record('rename'),
    setDefault: record('setDefault'),
    remove: record('remove'),
    ...overrides,
  };
  return { value, calls };
}

function renderView(
  tracks: readonly AudioTrack[] | undefined,
  fieldActions: LessonAudioTracksActions = actions().value,
  assetState: AudioTrackState = 'ready',
  state: Partial<LessonAudioTracksState> = {},
) {
  return render(
    <I18nProvider instance={createI18n('en')}>
      <main>
        <LessonAudioTracksView
          videoId={videoId}
          assetState={assetState}
          state={{ data: tracks, isLoading: false, ...state }}
          actions={fieldActions}
        />
      </main>
    </I18nProvider>,
  );
}

function fileInput(container: HTMLElement, index = 0): HTMLInputElement {
  const input = container.querySelectorAll<HTMLInputElement>('input[type="file"]').item(index);
  if (input === null) throw new Error('The file control was not rendered');
  return input;
}

const videoActions: LessonVideoFieldActions = {
  upload: async () => null,
  resume: async () => null,
  retry: async () => null,
  cancel: async () => null,
};

function video(state: LessonVideo['state'], overrides: Partial<LessonVideo> = {}): LessonVideo {
  return {
    lessonId,
    videoId,
    state,
    generation: 1,
    originalName: 'lesson.mp4',
    receivedBytes: 12,
    declaredBytes: 12,
    updatedAt: new Date().toISOString(),
    ...overrides,
  };
}

function renderVideoField(rows: readonly LessonVideo[]) {
  const asked: string[] = [];
  const audio: CourseEditorAudio = {
    useTracks: (id) => {
      asked.push(id);
      return { data: [], isLoading: false };
    },
    actions: actions().value,
  };
  render(
    <I18nProvider instance={createI18n('en')}>
      <main>
        <LessonVideoField
          lessonId={lessonId}
          isSaved
          rows={rows}
          isLoading={false}
          actions={videoActions}
          audio={audio}
        />
      </main>
    </I18nProvider>,
  );
  return asked;
}

describe('lesson audio tracks field', () => {
  it.each([
    { name: 'no video', rows: [], shown: false },
    { name: 'a processing video', rows: [video('processing')], shown: false },
    { name: 'a ready video', rows: [video('ready')], shown: true },
    { name: 'a video waiting for audio', rows: [video('awaiting_audio')], shown: true },
  ])('D-L2-1 shows the audio section for $name: $shown', ({ rows, shown }) => {
    const asked = renderVideoField(rows);
    expect(screen.queryByRole('region', { name: 'Audio tracks' }) !== null).toBe(shown);
    expect(asked.length > 0).toBe(shown);
  });

  it('D-L2-8 a newer video-only asset is the shown video and explains the missing sound', () => {
    const replacement = '50000000-0000-4000-8000-000000000002';
    const asked = renderVideoField([
      video('ready', { generation: 1 }),
      video('awaiting_audio', { videoId: replacement, generation: 2 }),
    ]);
    expect(screen.getByText('Waiting for an audio track')).toBeTruthy();
    expect(screen.getByText(/has no sound/)).toBeTruthy();
    expect(screen.getByRole('button', { name: 'Replace video' })).toBeTruthy();
    expect(asked).toContain(replacement);
  });

  it('D-L2-8 a stale older video-only asset never hides the Ready video', () => {
    const stale = '50000000-0000-4000-8000-000000000002';
    const asked = renderVideoField([
      video('ready', { generation: 2 }),
      video('awaiting_audio', { videoId: stale, generation: 1 }),
    ]);
    expect(screen.getByText('Ready')).toBeTruthy();
    expect(asked).not.toContain(stale);
  });

  it('uploads after a language and a name are chosen, exposes progress, and passes axe', async () => {
    const user = userEvent.setup();
    let release: (() => void) | undefined;
    const pending = new Promise<void>((resolve) => {
      release = resolve;
    });
    const captured: unknown[] = [];
    const { container } = renderView(
      [embedded],
      actions({
        upload: async (_videoId, metadata, _file, onProgress) => {
          captured.push(metadata);
          onProgress({ receivedBytes: 3, declaredBytes: 12 });
          await pending;
          return null;
        },
      }).value,
    );

    expect(screen.getByRole('button', { name: 'Upload audio' }).getAttribute('data-disabled')).toBe(
      'true',
    );
    await user.click(screen.getByRole('button', { name: /Track language/ }));
    // The embedded track holds uk: one track per language, so it is not offered again.
    expect(screen.queryByRole('option', { name: 'Ukrainian' })).toBeNull();
    expect(screen.getByRole('option', { name: 'Undetermined' })).toBeTruthy();
    await user.click(screen.getByRole('option', { name: 'English' }));
    await user.type(screen.getByRole('textbox', { name: 'Track name' }), 'English dub');
    await user.upload(fileInput(container), new File(['audio bytes!'], 'dub.m4a'));

    const progress = await screen.findByRole('progressbar', { name: 'Uploading audio' });
    expect(progress.getAttribute('aria-valuenow')).toBe('25');
    expect(captured).toEqual([{ language: 'en', label: 'English dub' }]);
    expect((await axe.run(container)).violations).toEqual([]);
    if (release === undefined) throw new Error('The upload was not initialized');
    release();
    await waitFor(() => expect(screen.queryByRole('progressbar')).toBeNull());
  });

  it('a failed track frees its language for a new upload', async () => {
    const user = userEvent.setup();
    renderView([embedded, track(dubId, 'failed', { language: 'pl', label: 'Polski' })]);
    await user.click(screen.getByRole('button', { name: /Track language/ }));
    expect(screen.getByRole('option', { name: 'Polish' })).toBeTruthy();
  });

  it('refuses a name the HLS manifest cannot carry before any file is picked', async () => {
    const user = userEvent.setup();
    renderView([embedded]);
    await user.click(screen.getByRole('button', { name: /Track language/ }));
    await user.click(screen.getByRole('option', { name: 'English' }));
    await user.type(screen.getByRole('textbox', { name: 'Track name' }), 'Say "hi"');
    expect(screen.getByRole('button', { name: 'Upload audio' }).getAttribute('data-disabled')).toBe(
      'true',
    );
    expect(screen.getByText(/no quotes or control characters/)).toBeTruthy();
  });

  it('D-L2-2 an interrupted track offers resume and cancel from the list alone', async () => {
    const user = userEvent.setup();
    const recorded = actions();
    const interrupted = track(dubId, 'uploading');
    const { container } = renderView([embedded, interrupted], recorded.value);

    expect(screen.getByText(/upload was interrupted/i)).toBeTruthy();
    const file = new File(['abc'], 'dub.m4a');
    // The new-track form has its own file control after the row's resume control.
    await user.upload(fileInput(container, 0), file);
    await waitFor(() => expect(recorded.calls.resume).toHaveLength(1));
    expect(recorded.calls.resume[0]?.[0]).toEqual(interrupted);
    await user.click(screen.getByRole('button', { name: 'Cancel upload' }));
    await waitFor(() => expect(recorded.calls.cancel).toHaveLength(1));
    expect(recorded.calls.cancel[0]?.[0]).toEqual(interrupted);
  });

  it('D-L2-4 offers default only on a Ready track that is not already the default', async () => {
    const user = userEvent.setup();
    const recorded = actions();
    const dub = track(dubId, 'ready');
    renderView(
      [
        embedded,
        dub,
        track('60000000-0000-4000-8000-000000000003', 'processing', {
          language: 'pl',
          label: 'Polski',
        }),
      ],
      recorded.value,
    );

    const buttons = screen.getAllByRole('button', { name: /the default track/ });
    expect(buttons).toHaveLength(1);
    await user.click(
      screen.getByRole('button', { name: 'Make English (English) the default track' }),
    );
    await waitFor(() => expect(recorded.calls.setDefault).toEqual([[dub]]));
    expect(
      screen.getByRole('progressbar', { name: 'Processing track Polski (Polish)' }),
    ).toBeTruthy();
  });

  it('D-L2-4 keeps the last Ready track of a Ready asset, and deletes only after confirmation', async () => {
    const user = userEvent.setup();
    const recorded = actions();
    const { unmount } = renderView([embedded], recorded.value);
    expect(screen.queryByRole('button', { name: /Delete track/ })).toBeNull();
    expect(screen.getByText(/last ready track/)).toBeTruthy();
    unmount();

    const failed = track(dubId, 'failed');
    renderView([embedded, failed], recorded.value);
    await user.click(screen.getByRole('button', { name: 'Delete track English (English)' }));
    expect(recorded.calls.remove).toHaveLength(0);
    await user.click(
      within(screen.getByRole('dialog')).getByRole('button', { name: 'Delete track' }),
    );
    await waitFor(() => expect(recorded.calls.remove).toEqual([[failed]]));
  });

  it('D-L2-4 lets a video-only asset delete its only Ready track', () => {
    renderView([track(dubId, 'ready')], actions().value, 'awaiting_audio');
    expect(screen.getByRole('button', { name: 'Delete track English (English)' })).toBeTruthy();
    expect(screen.getByText(/first ready audio track makes the video watchable/)).toBeTruthy();
  });

  it('edits the language and name of a track in a dialog', async () => {
    const user = userEvent.setup();
    const recorded = actions();
    renderView([embedded], recorded.value);
    await user.click(screen.getByRole('button', { name: 'Edit track Українська (Ukrainian)' }));
    const dialog = screen.getByRole('dialog');
    const name = within(dialog).getByRole('textbox', { name: 'Track name' });
    await user.clear(name);
    await user.type(name, 'Оригінал');
    await user.click(within(dialog).getByRole('button', { name: 'Save' }));
    await waitFor(() => expect(recorded.calls.rename).toEqual([[embedded, 'uk', 'Оригінал']]));
  });

  it('shows a 409 language conflict from a race as a named error', async () => {
    const user = userEvent.setup();
    const { container } = renderView(
      [embedded],
      actions({ upload: async () => 'duplicate_language' }).value,
    );
    await user.click(screen.getByRole('button', { name: /Track language/ }));
    await user.click(screen.getByRole('option', { name: 'English' }));
    await user.type(screen.getByRole('textbox', { name: 'Track name' }), 'English');
    await user.upload(fileInput(container), new File(['abc'], 'dub.m4a'));
    expect((await screen.findByRole('alert')).textContent).toContain(
      'This language already has a track.',
    );
  });

  it('shows a localized degraded state and no upload form while the list is unavailable', () => {
    const { container } = renderView(undefined, actions().value, 'ready', {
      error: 'dependency_unavailable',
    });
    expect(screen.getByRole('alert').textContent).toContain(
      'Audio service is temporarily unavailable.',
    );
    expect(screen.queryByRole('button', { name: 'Upload audio' })).toBeNull();
    expect(container.innerHTML).not.toMatch(/[Ѐ-ӿ]/u);
  });
});
