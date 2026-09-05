import { describe, expect, it } from '@rstest/core';
import { render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import axe from 'axe-core';
import { useState } from 'react';
import { AuthoringQueryError } from '../../app/composition/admin-queries';
import { createI18n, I18nProvider } from '../../app/providers/i18n';
import { asBrandedId } from '../../domain/course/course';
import type { AdminCourseDraft, CourseTree, CourseTreeEdit } from '../../domain/course/course';
import type { AuthenticatedUser } from '../../domain/identity/identity';
import { CourseCreateDialog, CourseEditorView } from './CourseEditorView';

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

function savedTree(current: CourseTree, edit: CourseTreeEdit): CourseTree {
  return { ...current, ...edit, updatedAt: '2026-09-01T00:00:00.000002Z' };
}

describe('course editor', () => {
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
  });

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
