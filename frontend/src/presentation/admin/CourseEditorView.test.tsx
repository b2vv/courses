import { afterAll, afterEach, beforeAll, describe, expect, it } from '@rstest/core';
import { QueryClient } from '@tanstack/react-query';
import {
  createMemoryHistory,
  createRootRoute,
  createRoute,
  createRouter,
  Outlet,
  RouterProvider,
} from '@tanstack/react-router';
import { render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import axe from 'axe-core';
import { http, HttpResponse } from 'msw';
import { setupServer } from 'msw/node';
import { useState } from 'react';
import { apiOrigin } from '../../adapters/api/api-origin';
import { createApiAdminCourseRepository } from '../../adapters/api/admin-course-repository';
import { createAuthoring } from '../../application/admin/authoring';
import { AuthoringQueryError, createAdminQueries } from '../../app/composition/admin-queries';
import { ensureRouteQuery } from '../../app/composition/catalog-queries';
import { createI18n, I18nProvider } from '../../app/providers/i18n';
import { createMemoryTranscriptHistory } from '../../adapters/storage/transcript-history-memory';
import { asBrandedId } from '../../domain/course/course';
import type { AdminCourseDraft, CourseTree, CourseTreeEdit } from '../../domain/course/course';
import type { AuthenticatedUser } from '../../domain/identity/identity';
import type { LessonDraft } from '../../domain/transcript/draft';
import type { LessonTranscript } from '../../domain/transcript/transcript';
import { CourseCreateDialog, CourseEditorView } from './CourseEditorView';
import { CourseListFailure } from './CourseListView';

const admin: AuthenticatedUser = {
  id: '40000000-0000-4000-8000-000000000003',
  email: 'admin@courseforge.local',
  displayName: 'Admin',
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
  modules: [],
};

const server = setupServer();
beforeAll(() => server.listen({ onUnhandledRequest: 'error' }));
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

function savedTree(current: CourseTree, edit: CourseTreeEdit): CourseTree {
  return { ...current, ...edit, updatedAt: '2026-09-01T00:00:00.000002Z' };
}

describe('course editor', () => {
  it('routes a loaded draft to only its lesson transcript editor', async () => {
    const browser = userEvent.setup();
    const firstLessonId = asBrandedId('20000000-0000-4000-8000-000000000101', 'lesson');
    const secondLessonId = asBrandedId('20000000-0000-4000-8000-000000000102', 'lesson');
    const course = {
      ...tree,
      modules: [
        {
          id: asBrandedId('10000000-0000-4000-8000-000000000101', 'module'),
          title: 'Module',
          lessons: [
            { id: firstLessonId, title: 'First', description: '' },
            { id: secondLessonId, title: 'Second', description: '' },
          ],
        },
      ],
    } satisfies CourseTree;
    const readyDraft: LessonDraft = {
      lessonId: firstLessonId,
      state: 'ready',
      language: 'en',
      failure: null,
      stale: false,
      updatedAt: '2026-09-24T12:00:00Z',
      assetId: null,
      generation: 1,
      segments: [{ id: 'draft-line', startMs: 0, endMs: 900, text: 'Loaded' }],
    };
    const transcripts = new Map<string, LessonTranscript>([
      [
        firstLessonId,
        {
          lessonId: firstLessonId,
          language: 'uk',
          updatedAt: '2026-09-24T11:00:00Z',
          stale: false,
          segments: [
            { id: 'first-line', startMs: 0, endMs: 900, text: 'First old', documentIds: [] },
          ],
          documents: [],
        },
      ],
      [
        secondLessonId,
        {
          lessonId: secondLessonId,
          language: 'uk',
          updatedAt: '2026-09-24T11:00:00Z',
          stale: false,
          segments: [
            { id: 'second-line', startMs: 0, endMs: 900, text: 'Second old', documentIds: [] },
          ],
          documents: [],
        },
      ],
    ]);
    render(
      <I18nProvider instance={createI18n('en')}>
        <CourseEditorView
          course={course}
          actor={admin}
          ownerOptions={[{ id: admin.id, displayName: admin.displayName, role: 'admin' }]}
          onSave={async (current) => current}
          onTransition={async (current) => current}
          onDelete={async () => {}}
          onAuthenticationRequired={() => {}}
          drafts={{
            useLessonDraft: (lessonId) => ({
              data:
                lessonId === firstLessonId ? readyDraft : { ...readyDraft, lessonId, segments: [] },
              isLoading: false,
            }),
            actions: { regenerate: async () => null },
          }}
          transcripts={{
            useLessonTranscript: (lessonId) => ({
              data: transcripts.get(lessonId),
              isLoading: false,
            }),
            documents: [],
            history: { store: createMemoryTranscriptHistory(), authorId: admin.id },
            library: { add: async () => null, remove: async () => null },
            save: async () => null,
          }}
        />
      </I18nProvider>,
    );

    await browser.click(screen.getByRole('button', { name: 'Load draft into editor' }));
    const editors = screen.getAllByRole('heading', { name: 'Lesson transcript' }).map((heading) => {
      const section = heading.closest('section');
      if (section === null) throw new Error('transcript editor heading has no section');
      return within(section);
    });
    const firstEditor = editors[0];
    const secondEditor = editors[1];
    if (firstEditor === undefined || secondEditor === undefined)
      throw new Error('both lesson transcript editors must render');
    expect(firstEditor.getByText('A recognized draft is ready')).toBeTruthy();
    expect(secondEditor.queryByText('A recognized draft is ready')).toBeNull();

    await browser.click(firstEditor.getByRole('button', { name: 'Load the draft' }));
    expect(firstEditor.getByDisplayValue('Loaded')).toBeTruthy();
    expect(secondEditor.getByDisplayValue('Second old')).toBeTruthy();
    expect(firstEditor.queryByDisplayValue('First old')).toBeNull();
  });

  it('P19 saves newly added modules and lessons with the loaded concurrency token', async () => {
    const browser = userEvent.setup();
    let received: { current: CourseTree; edit: CourseTreeEdit } | undefined;
    render(
      <I18nProvider instance={createI18n('en')}>
        <CourseEditorView
          course={tree}
          actor={admin}
          ownerOptions={[{ id: admin.id, displayName: admin.displayName, role: 'admin' }]}
          onSave={async (current, edit) => {
            received = { current, edit };
            return savedTree(current, edit);
          }}
          onTransition={async (current) => current}
          onDelete={async () => {}}
          onAuthenticationRequired={() => {}}
        />
      </I18nProvider>,
    );

    await browser.click(screen.getByRole('button', { name: 'Add module' }));
    await browser.type(screen.getByRole('textbox', { name: 'Module 1 title' }), 'Foundations');
    await browser.click(screen.getByRole('button', { name: 'Add lesson to module 1' }));
    await browser.type(
      screen.getByRole('textbox', { name: 'Lesson 1 title in module 1' }),
      'First program',
    );
    await browser.click(screen.getByRole('button', { name: 'Save changes' }));

    expect(received?.current.updatedAt).toBe('2026-09-01T00:00:00.000001Z');
    expect(received?.edit.modules).toHaveLength(1);
    expect(received?.edit.modules[0]?.id).toMatch(
      /^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/,
    );
    expect(received?.edit.modules[0]?.lessons[0]?.id).toMatch(
      /^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/,
    );
    expect(screen.getByDisplayValue('First program')).toBeTruthy();
  });

  it('reorders and deletes tree nodes through named keyboard buttons', async () => {
    const browser = userEvent.setup();
    const structured: CourseTree = {
      ...tree,
      modules: [
        {
          id: asBrandedId('10000000-0000-4000-8000-000000000101', 'module'),
          title: 'First module',
          lessons: [
            {
              id: asBrandedId('20000000-0000-4000-8000-000000000101', 'lesson'),
              title: 'First lesson',
              description: '',
            },
            {
              id: asBrandedId('20000000-0000-4000-8000-000000000102', 'lesson'),
              title: 'Second lesson',
              description: '',
            },
          ],
        },
        {
          id: asBrandedId('10000000-0000-4000-8000-000000000102', 'module'),
          title: 'Second module',
          lessons: [],
        },
      ],
    };
    let saved: CourseTreeEdit | undefined;
    render(
      <I18nProvider instance={createI18n('en')}>
        <CourseEditorView
          course={structured}
          actor={admin}
          ownerOptions={[{ id: admin.id, displayName: admin.displayName, role: 'admin' }]}
          onSave={async (current, edit) => {
            saved = edit;
            return savedTree(current, edit);
          }}
          onTransition={async (current) => current}
          onDelete={async () => {}}
          onAuthenticationRequired={() => {}}
        />
      </I18nProvider>,
    );

    await browser.click(screen.getByRole('button', { name: 'Move module 2 up' }));
    await browser.click(screen.getByRole('button', { name: 'Move lesson 2 in module 2 up' }));
    await browser.click(screen.getByRole('button', { name: 'Delete lesson 2 in module 2' }));
    await browser.click(screen.getByRole('button', { name: 'Save changes' }));

    expect(saved?.modules.map((module) => module.title)).toEqual(['Second module', 'First module']);
    expect(saved?.modules[1]?.lessons.map((lesson) => lesson.title)).toEqual(['Second lesson']);
  });

  it.each([
    {
      status: 'draft',
      publishedAt: null,
      action: 'Publish',
      deletable: true,
      slugDisabled: false,
    },
    {
      status: 'published',
      publishedAt: '2026-09-01T00:00:00.000001Z',
      action: 'Archive',
      deletable: false,
      slugDisabled: true,
    },
    {
      status: 'archived',
      publishedAt: '2026-09-01T00:00:00.000001Z',
      action: 'Restore',
      deletable: false,
      slugDisabled: true,
    },
  ] as const)(
    'renders the $status lifecycle controls and permanent-slug rule',
    ({ status, publishedAt, action, deletable, slugDisabled }) => {
      render(
        <I18nProvider instance={createI18n('en')}>
          <CourseEditorView
            course={{ ...tree, status, publishedAt }}
            actor={admin}
            ownerOptions={[{ id: admin.id, displayName: admin.displayName, role: 'admin' }]}
            onSave={async (current) => current}
            onTransition={async (current) => current}
            onDelete={async () => {}}
            onAuthenticationRequired={() => {}}
          />
        </I18nProvider>,
      );

      expect(screen.getByRole('button', { name: action })).toBeTruthy();
      expect(screen.queryByRole('button', { name: 'Delete draft' }) !== null).toBe(deletable);
      expect(screen.getByRole('textbox', { name: 'Slug' }).hasAttribute('disabled')).toBe(
        slugDisabled,
      );
    },
  );

  // 40 modules of 3 lessons rendered, walked by keyboard and then audited by axe: measured at
  // roughly five seconds of work against a five-second default that nobody chose for it. Run
  // alone on an idle machine it failed once in three. The budget is stated here rather than
  // inherited, and nothing about what the test asserts changes - raising it does not weaken a
  // single assertion, while leaving it makes a green run a matter of scheduling luck.
  it('B09 keeps a maximal module list keyboard reachable and free of axe violations', async () => {
    const deepTree: CourseTree = {
      ...tree,
      modules: Array.from({ length: 40 }, (_, moduleIndex) => ({
        id: asBrandedId(
          `10000000-0000-4000-8000-${String(moduleIndex + 100).padStart(12, '0')}`,
          'module',
        ),
        title: `Module ${moduleIndex + 1}`,
        lessons: Array.from({ length: 3 }, (_, lessonIndex) => ({
          id: asBrandedId(
            `20000000-0000-4000-8000-${String(moduleIndex * 3 + lessonIndex + 100).padStart(
              12,
              '0',
            )}`,
            'lesson',
          ),
          title: `Lesson ${moduleIndex + 1}.${lessonIndex + 1}`,
          description: '',
        })),
      })),
    };
    const { container } = render(
      <I18nProvider instance={createI18n('en')}>
        <main>
          <CourseEditorView
            course={deepTree}
            actor={admin}
            ownerOptions={[{ id: admin.id, displayName: admin.displayName, role: 'admin' }]}
            onSave={async (current) => current}
            onTransition={async (current) => current}
            onDelete={async () => {}}
            onAuthenticationRequired={() => {}}
          />
        </main>
      </I18nProvider>,
    );

    const moduleTitles = screen.getAllByRole('textbox', { name: /^Module \d+ title$/ });
    const lessonTitles = screen.getAllByRole('textbox', {
      name: /^Lesson \d+ title in module \d+$/,
    });
    expect(moduleTitles).toHaveLength(40);
    expect(lessonTitles).toHaveLength(120);
    for (const field of [...moduleTitles, ...lessonTitles]) expect(field.tabIndex).toBe(0);
    expect((await axe.run(container)).violations).toEqual([]);
  }, 20_000);

  it('R28 replaces a conflicted edit with the refetched stored tree and neutral replay copy', async () => {
    const browser = userEvent.setup();
    const stored = {
      ...tree,
      title: 'Stored title',
      updatedAt: '2026-09-01T00:00:00.000009Z',
    };

    function ConflictHarness() {
      const [course, setCourse] = useState(tree);
      return (
        <CourseEditorView
          course={course}
          actor={admin}
          ownerOptions={[{ id: admin.id, displayName: admin.displayName, role: 'admin' }]}
          onSave={async () => {
            setCourse(stored);
            throw new AuthoringQueryError({ code: 'course_modified' });
          }}
          onTransition={async (current) => current}
          onDelete={async () => {}}
          onAuthenticationRequired={() => {}}
        />
      );
    }

    render(
      <I18nProvider instance={createI18n('en')}>
        <ConflictHarness />
      </I18nProvider>,
    );
    const title = screen.getByRole('textbox', { name: 'Title' });
    await browser.clear(title);
    await browser.type(title, 'My local title');
    await browser.click(screen.getByRole('button', { name: 'Save changes' }));

    expect((await screen.findByRole('alert')).textContent).toContain(
      'The course changed since it was loaded. This may have been your own save',
    );
    await waitFor(() => expect(screen.getByDisplayValue('Stored title')).toBeTruthy());
    expect(screen.getByRole('alert').textContent).not.toContain('another person');
  });

  it('R36 reports field errors and refuses an invalid form before save', async () => {
    const browser = userEvent.setup();
    let saves = 0;
    render(
      <I18nProvider instance={createI18n('en')}>
        <CourseEditorView
          course={tree}
          actor={admin}
          ownerOptions={[{ id: admin.id, displayName: admin.displayName, role: 'admin' }]}
          onSave={async (current) => {
            saves += 1;
            return current;
          }}
          onTransition={async (current) => current}
          onDelete={async () => {}}
          onAuthenticationRequired={() => {}}
        />
      </I18nProvider>,
    );

    await browser.clear(screen.getByRole('textbox', { name: 'Title' }));
    const slug = screen.getByRole('textbox', { name: 'Slug' });
    await browser.clear(slug);
    await browser.type(slug, 'x'.repeat(300));
    await browser.click(screen.getByRole('button', { name: 'Save changes' }));

    expect(screen.getByText('Enter a value.')).toBeTruthy();
    expect(screen.getByText('Use at most 120 characters.')).toBeTruthy();
    expect(saves).toBe(0);
  });

  it('R38 retries an unavailable primary editor read without showing an empty form', async () => {
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const adminQueries = createAdminQueries(createAuthoring(createApiAdminCourseRepository()));
    let reads = 0;
    server.use(
      http.get(`${apiOrigin()}/api/v1/admin/courses/${tree.id}`, () => {
        reads += 1;
        return reads === 1
          ? HttpResponse.json(
              {
                error: {
                  code: 'dependency_unavailable',
                  message: 'x',
                  requestId: 'b'.repeat(32),
                },
              },
              { status: 503 },
            )
          : HttpResponse.json(tree);
      }),
    );
    const root = createRootRoute({ component: Outlet });
    const editor = createRoute({
      getParentRoute: () => root,
      path: '/admin/courses/$courseId',
      loader: ({ abortController }) =>
        ensureRouteQuery(queryClient, adminQueries.detail(tree.id), abortController.signal),
      component: () => (
        <CourseEditorView
          course={tree}
          actor={admin}
          ownerOptions={null}
          onSave={async (current) => current}
          onTransition={async (current) => current}
          onDelete={async () => {}}
          onAuthenticationRequired={() => {}}
        />
      ),
      errorComponent: CourseListFailure,
    });
    const router = createRouter({
      routeTree: root.addChildren([editor]),
      history: createMemoryHistory({ initialEntries: [`/admin/courses/${tree.id}`] }),
      defaultPendingMs: 0,
      defaultPendingMinMs: 0,
      Wrap: ({ children }) => <I18nProvider instance={createI18n('en')}>{children}</I18nProvider>,
    });

    await router.load();
    render(<RouterProvider router={router} />);

    const error = await screen.findByRole('alert');
    expect(within(error).getByRole('heading', { name: 'Could not load courses' })).toBeTruthy();
    expect(screen.queryByRole('textbox', { name: 'Title' })).toBeNull();
    await userEvent.click(within(error).getByRole('button', { name: 'Try again' }));

    expect(await screen.findByRole('textbox', { name: 'Title' })).toBeTruthy();
    expect(reads).toBe(2);
  });

  it('R40 disables Save while one PUT is in flight and ignores a second press', async () => {
    const browser = userEvent.setup();
    const authoring = createAuthoring(createApiAdminCourseRepository());
    let puts = 0;
    let releasePut: (() => void) | undefined;
    const pendingPut = new Promise<void>((resolve) => {
      releasePut = resolve;
    });
    server.use(
      http.put(`${apiOrigin()}/api/v1/admin/courses/${tree.id}`, async () => {
        puts += 1;
        await pendingPut;
        return HttpResponse.json({
          ...tree,
          updatedAt: '2026-09-01T00:00:00.000002Z',
        });
      }),
    );
    render(
      <I18nProvider instance={createI18n('en')}>
        <CourseEditorView
          course={tree}
          actor={admin}
          ownerOptions={[{ id: admin.id, displayName: admin.displayName, role: 'admin' }]}
          onSave={async (current, edit) => {
            const result = await authoring.save(current, edit);
            if (!result.ok) throw new AuthoringQueryError(result.error);
            return result.value;
          }}
          onTransition={async (current) => current}
          onDelete={async () => {}}
          onAuthenticationRequired={() => {}}
        />
      </I18nProvider>,
    );

    const save = screen.getByRole('button', { name: 'Save changes' });
    await browser.click(save);
    await waitFor(() => expect(puts).toBe(1));
    expect(save.hasAttribute('disabled')).toBe(true);
    await browser.click(save);
    expect(puts).toBe(1);

    if (releasePut === undefined) throw new Error('The save request was not initialized');
    releasePut();
    await waitFor(() => expect(save.hasAttribute('disabled')).toBe(false));
  });
});

describe('course create dialog', () => {
  it('R30 disables Create while the mutation is pending and sends one request', async () => {
    const browser = userEvent.setup();
    let requests = 0;
    let release: ((course: CourseTree) => void) | undefined;
    const pending = new Promise<CourseTree>((resolve) => {
      release = resolve;
    });
    render(
      <I18nProvider instance={createI18n('en')}>
        <CourseCreateDialog
          actor={admin}
          ownerOptions={[{ id: admin.id, displayName: admin.displayName, role: 'admin' }]}
          onCreate={async (_draft: AdminCourseDraft) => {
            requests += 1;
            return pending;
          }}
          onSlugTaken={async () => undefined}
          onAuthenticationRequired={() => {}}
        />
      </I18nProvider>,
    );

    await browser.click(screen.getByRole('button', { name: 'Create course' }));
    const dialog = screen.getByRole('dialog');
    await browser.type(within(dialog).getByRole('textbox', { name: 'Title' }), 'New course');
    await browser.type(within(dialog).getByRole('textbox', { name: 'Slug' }), 'new-course');
    await browser.type(within(dialog).getByRole('textbox', { name: 'Summary' }), 'Summary');
    await browser.type(within(dialog).getByRole('textbox', { name: 'Description' }), 'Description');
    const create = within(dialog).getByRole('button', { name: 'Create' });
    await browser.click(create);

    expect(create.hasAttribute('disabled')).toBe(true);
    await browser.click(create);
    expect(requests).toBe(1);
    if (release === undefined) throw new Error('The create request was not initialized');
    release(tree);
    await waitFor(() => expect(create.hasAttribute('disabled')).toBe(false));
  });

  it('refreshes after slug_taken and offers the existing course', async () => {
    const browser = userEvent.setup();
    let refreshedSlug: string | undefined;
    render(
      <I18nProvider instance={createI18n('en')}>
        <CourseCreateDialog
          actor={admin}
          ownerOptions={[{ id: admin.id, displayName: admin.displayName, role: 'admin' }]}
          onCreate={async () => {
            throw new AuthoringQueryError({ code: 'slug_taken' });
          }}
          onSlugTaken={async (slug) => {
            refreshedSlug = slug;
            return tree.id;
          }}
          onAuthenticationRequired={() => {}}
        />
      </I18nProvider>,
    );

    await browser.click(screen.getByRole('button', { name: 'Create course' }));
    const dialog = screen.getByRole('dialog');
    await browser.type(within(dialog).getByRole('textbox', { name: 'Title' }), 'New course');
    await browser.type(within(dialog).getByRole('textbox', { name: 'Slug' }), 'go-basics');
    await browser.type(within(dialog).getByRole('textbox', { name: 'Summary' }), 'Summary');
    await browser.type(within(dialog).getByRole('textbox', { name: 'Description' }), 'Description');
    await browser.click(within(dialog).getByRole('button', { name: 'Create' }));

    expect(await within(dialog).findByText(/already exists/)).toBeTruthy();
    expect(refreshedSlug).toBe('go-basics');
    expect(
      within(dialog).getByRole('link', { name: 'Open the existing course' }).getAttribute('href'),
    ).toBe(`/admin/courses/${tree.id}`);
  });
});
