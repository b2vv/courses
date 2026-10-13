import { afterEach, describe, expect, it } from '@rstest/core';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import axe from 'axe-core';
import { createI18n, I18nProvider } from '../../app/providers/i18n';
import { asBrandedId } from '../../domain/course/course';
import type { CourseTree } from '../../domain/course/course';
import type { AuthenticatedUser } from '../../domain/identity/identity';
import type { LessonVideo } from '../../domain/media/video';
import { CourseEditorView } from './CourseEditorView';
import {
  LessonVideoField,
  type LessonVideoFieldActions,
  type LessonVideoFieldActionResult,
} from './LessonVideoField';

const lessonId = '20000000-0000-4000-8000-000000000001';
const videoId = '50000000-0000-4000-8000-000000000001';
const admin: AuthenticatedUser = {
  id: '40000000-0000-4000-8000-000000000003',
  email: 'admin@courseforge.local',
  displayName: 'Admin',
  emailVerified: true,
  twoFactorEnabled: false,
  role: 'admin',
};
const tree: CourseTree = {
  id: asBrandedId('10000000-0000-4000-8000-000000000001', 'course'),
  slug: 'go-basics',
  title: 'Go basics',
  summary: 'Practical Go',
  description: 'Learn Go through practice.',
  status: 'draft',
  instructorId: admin.id,
  createdAt: '2026-09-01T00:00:00.000000Z',
  updatedAt: '2026-09-01T00:00:00.000001Z',
  publishedAt: null,
  price: null,
  modules: [
    {
      id: asBrandedId('10000000-0000-4000-8000-000000000101', 'module'),
      title: 'Foundations',
      lessons: [
        {
          id: asBrandedId(lessonId, 'lesson'),
          title: 'First lesson',
          isPreview: false,
        },
      ],
    },
  ],
};

function actionResult(value: LessonVideoFieldActionResult = null) {
  return Promise.resolve(value);
}

function actions(overrides: Partial<LessonVideoFieldActions> = {}): LessonVideoFieldActions {
  return {
    upload: async () => actionResult(),
    resume: async () => actionResult(),
    retry: async () => actionResult(),
    cancel: async () => actionResult(),
    ...overrides,
  };
}

function video(state: LessonVideo['state'], overrides: Partial<LessonVideo> = {}): LessonVideo {
  return {
    lessonId,
    videoId,
    state,
    generation: 1,
    originalName: 'lesson.mp4',
    receivedBytes: 4,
    declaredBytes: 12,
    updatedAt: new Date().toISOString(),
    ...overrides,
  };
}

function renderField(
  rows: readonly LessonVideo[],
  fieldActions: LessonVideoFieldActions = actions(),
  options: { isSaved?: boolean; isLoading?: boolean; error?: 'dependency_unavailable' } = {},
) {
  return render(
    <I18nProvider instance={createI18n('en')}>
      <main>
        <LessonVideoField
          lessonId={lessonId}
          isSaved={options.isSaved ?? true}
          rows={rows}
          isLoading={options.isLoading ?? false}
          actions={fieldActions}
          {...(options.error === undefined ? {} : { error: options.error })}
        />
      </main>
    </I18nProvider>,
  );
}

afterEach(() => localStorage.clear());

describe('lesson video field', () => {
  it('P09 B20 accepts a file from the keyboard path, exposes progress, and passes axe', async () => {
    const user = userEvent.setup();
    let release: (() => void) | undefined;
    const pending = new Promise<void>((resolve) => {
      release = resolve;
    });
    const { container } = renderField(
      [],
      actions({
        upload: async (_lessonId, _file, onProgress) => {
          onProgress({ receivedBytes: 6, declaredBytes: 12 });
          await pending;
          return null;
        },
      }),
    );

    await user.tab();
    const trigger = screen.getByRole('button', { name: 'Upload video' });
    expect(document.activeElement).toBe(trigger);
    await user.keyboard('{Enter}');
    const input = container.querySelector<HTMLInputElement>('input[type="file"]');
    if (input === null) throw new Error('The file control was not rendered');
    await user.upload(input, new File(['video bytes'], 'lesson.mp4', { type: 'video/mp4' }));

    const progress = await screen.findByRole('progressbar', { name: 'Uploading video' });
    expect(progress.getAttribute('aria-valuenow')).toBe('50');
    expect((await axe.run(container)).violations).toEqual([]);
    if (release === undefined) throw new Error('The upload was not initialized');
    release();
    await waitFor(() => expect(screen.queryByRole('progressbar')).toBeNull());
  });

  it('R24 mounts no upload control for a newly added lesson until the tree is saved', async () => {
    const user = userEvent.setup();
    render(
      <I18nProvider instance={createI18n('en')}>
        <CourseEditorView
          course={tree}
          actor={admin}
          ownerOptions={null}
          video={{ rows: [], isLoading: false, actions: actions() }}
          onSave={async (current) => current}
          onTransition={async (current) => current}
          onDelete={async () => {}}
          onAuthenticationRequired={() => {}}
        />
      </I18nProvider>,
    );

    expect(screen.getAllByRole('button', { name: 'Upload video' })).toHaveLength(1);
    await user.click(screen.getByRole('button', { name: 'Add lesson to module 1' }));
    expect(screen.getAllByRole('button', { name: 'Upload video' })).toHaveLength(1);
    expect(screen.getByText('Save the course before adding video to this lesson.')).toBeTruthy();
  });

  it('R25 offers resume and cancel after interruption, then frees the field immediately', async () => {
    const user = userEvent.setup();
    let cancelled = '';
    const { container } = renderField(
      [video('uploading')],
      actions({
        cancel: async (row) => {
          cancelled = row.videoId;
          return null;
        },
      }),
    );

    expect(screen.getByText(/upload was interrupted/i)).toBeTruthy();
    expect(screen.getByRole('button', { name: 'Choose file to resume' })).toBeTruthy();
    await user.click(screen.getByRole('button', { name: 'Cancel upload' }));
    expect(cancelled).toBe(videoId);
    expect(await screen.findByRole('button', { name: 'Upload video' })).toBeTruthy();
    expect(container.querySelector('[data-video-state="uploading"]')).toBeNull();
  });

  it('refuses an unattested resume and keeps cancel-and-restart available', async () => {
    const user = userEvent.setup();
    const { container } = renderField(
      [video('uploading')],
      actions({ resume: async () => actionResult('upload_resume_unverified') }),
    );
    const input = container.querySelector<HTMLInputElement>('input[type="file"]');
    if (input === null) throw new Error('The resume file control was not rendered');
    await user.upload(input, new File(['different bytes'], 'other.mp4', { type: 'video/mp4' }));

    expect((await screen.findByRole('alert')).textContent).toContain(
      'This file cannot be verified for resume. Cancel the upload and start again.',
    );
    expect(screen.getByRole('button', { name: 'Cancel upload' })).toBeTruthy();
    expect(screen.getByRole('button', { name: 'Choose file to resume' })).toBeTruthy();
  });

  it.each([
    ['queued', 11, true],
    ['queued', 1, false],
    ['processing', 41, true],
    ['processing', 1, false],
  ] as const)(
    'R41 exposes retry for a %s row aged %i minutes only at the server threshold',
    (state, ageMinutes, retryable) => {
      renderField([
        video(state, { updatedAt: new Date(Date.now() - ageMinutes * 60_000).toISOString() }),
      ]);

      expect(screen.queryByRole('button', { name: 'Retry video' }) !== null).toBe(retryable);
      expect(screen.getByText(state === 'queued' ? 'Queued' : 'Processing')).toBeTruthy();
      if (!retryable)
        expect(screen.getByRole('progressbar', { name: 'Video processing' })).toBeTruthy();
    },
  );

  it('uses the D-12 badge order and surfaces a closed failure reason with retry', () => {
    const failed = video('failed', {
      videoId: '50000000-0000-4000-8000-000000000002',
      generation: 3,
      receivedBytes: 12,
      failureReason: 'render_failed',
    });
    const ready = video('ready', { generation: 2 });
    const queued = video('queued', {
      videoId: '50000000-0000-4000-8000-000000000003',
      generation: 1,
    });
    const { rerender } = renderField([failed, ready, queued]);

    expect(screen.getByText('Queued')).toBeTruthy();
    expect(screen.queryByText('Ready')).toBeNull();
    rerender(
      <I18nProvider instance={createI18n('en')}>
        <main>
          <LessonVideoField
            lessonId={lessonId}
            isSaved
            rows={[failed, ready]}
            isLoading={false}
            actions={actions()}
          />
        </main>
      </I18nProvider>,
    );
    expect(screen.getByText('Ready')).toBeTruthy();
    expect(screen.queryByText('Rendering failed')).toBeNull();
    rerender(
      <I18nProvider instance={createI18n('en')}>
        <main>
          <LessonVideoField
            lessonId={lessonId}
            isSaved
            rows={[failed]}
            isLoading={false}
            actions={actions()}
          />
        </main>
      </I18nProvider>,
    );
    expect(screen.getByText('Rendering failed')).toBeTruthy();
    expect(screen.getByRole('button', { name: 'Retry video' })).toBeTruthy();
  });

  it('shows a localized degraded state and disables minting while status is unavailable', () => {
    const { container } = renderField([], actions(), { error: 'dependency_unavailable' });

    expect(screen.getByRole('alert').textContent).toContain(
      'Video service is temporarily unavailable. Try again later.',
    );
    expect(screen.queryByRole('button', { name: 'Upload video' })).toBeNull();
    expect(container.textContent).not.toMatch(/[\u0400-\u04ff]/u);
    expect(container.innerHTML).not.toMatch(/[\u0400-\u04ff]/u);
  });

  it('surfaces a degraded upload without replacing the retryable file control', async () => {
    const user = userEvent.setup();
    const { container } = renderField(
      [],
      actions({ upload: async () => actionResult('dependency_unavailable') }),
    );
    const input = container.querySelector<HTMLInputElement>('input[type="file"]');
    if (input === null) throw new Error('The file control was not rendered');
    await user.upload(input, new File(['video'], 'lesson.mp4', { type: 'video/mp4' }));

    await waitFor(() =>
      expect(screen.getByRole('alert').textContent).toContain(
        'Video service is temporarily unavailable. Try again later.',
      ),
    );
    expect(screen.getByRole('button', { name: 'Upload video' })).toBeTruthy();
  });
});
