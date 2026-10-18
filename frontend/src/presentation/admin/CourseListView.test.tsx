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
import { render, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { http, HttpResponse } from 'msw';
import { setupServer } from 'msw/node';
import { apiOrigin } from '../../adapters/api/api-origin';
import { createApiAdminCourseRepository } from '../../adapters/api/admin-course-repository';
import { createAuthoring } from '../../application/admin/authoring';
import { createAdminQueries } from '../../app/composition/admin-queries';
import { ensureRouteQuery } from '../../app/composition/catalog-queries';
import { createI18n, I18nProvider } from '../../app/providers/i18n';
import { adminCourseQuery } from '../../app/router/admin-search';
import { asBrandedId } from '../../contracts/course';
import type { AdminCourse, AdminCoursePage, CourseStatus } from '../../contracts/course';
import { CourseListFailure, CourseListView } from './CourseListView';

const draftCourse: AdminCourse = {
  id: asBrandedId('10000000-0000-4000-8000-000000000001', 'course'),
  slug: 'go-architecture',
  title: 'Архітектура Go',
  status: 'draft',
  instructorId: '20000000-0000-4000-8000-000000000002',
  updatedAt: '2026-09-05T10:15:30.000000Z',
  lessonCount: 12,
  studentCount: 0,
  price: null,
};

const firstPage: AdminCoursePage = {
  items: [draftCourse],
  pagination: { limit: 20, offset: 0, total: 21 },
  statusCounts: { total: 21, draft: 8, published: 10, archived: 3 },
};

const server = setupServer();
beforeAll(() => server.listen({ onUnhandledRequest: 'error' }));
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

const noop = {
  onStatusChange: () => {},
  onSearch: () => {},
  onSortChange: () => {},
  onPageChange: () => {},
  onClearFilters: () => {},
  onDelete: async () => {},
  onAuthenticationRequired: () => {},
};

describe('admin course list', () => {
  it('renders the real columns, the editor link, the global summary and a free price', () => {
    render(
      <CourseListView page={firstPage} state={{ page: 1, search: '', sort: 'recent' }} {...noop} />,
    );

    const table = screen.getByRole('table', { name: 'Курси' });
    for (const header of ['Назва', 'Статус', 'Студенти', 'Ціна', 'Оновлено']) {
      expect(within(table).getByRole('columnheader', { name: new RegExp(header) })).toBeTruthy();
    }
    expect(
      within(table)
        .getByRole('link', { name: /^Архітектура Go/ })
        .getAttribute('href'),
    ).toBe('/admin/courses/10000000-0000-4000-8000-000000000001');
    expect(within(table).getByText('уроків: 12')).toBeTruthy();
    expect(within(table).getByText('Чернетка')).toBeTruthy();
    expect(within(table).getByText('Безкоштовно')).toBeTruthy();
    expect(within(table).getByText(/2026/).closest('time')?.getAttribute('datetime')).toBe(
      '2026-09-05T10:15:30.000000Z',
    );
    // The summary strip reads the global snapshot, not the (filtered) page.
    const summary = screen.getByRole('group', { name: 'Підсумок курсів' });
    expect(within(summary).getByText('21')).toBeTruthy();
    expect(within(summary).getByText('10')).toBeTruthy();
    expect(within(summary).getByText('8')).toBeTruthy();
  });

  it('filters, searches, sorts and pages through accessible controls', async () => {
    const user = userEvent.setup();
    let status: CourseStatus | undefined = undefined;
    let query = '';
    let sort = '';
    let page = 1;
    render(
      <CourseListView
        page={firstPage}
        state={{ page: 1, search: '', sort: 'recent' }}
        onStatusChange={(next) => {
          status = next;
        }}
        onSearch={(next) => {
          query = next;
        }}
        onSortChange={(next) => {
          sort = next;
        }}
        onPageChange={(next) => {
          page = next;
        }}
        onClearFilters={() => {}}
        onDelete={async () => {}}
        onAuthenticationRequired={() => {}}
      />,
    );

    await user.click(screen.getByRole('radio', { name: 'Опубліковано' }));
    expect(status).toBe('published');

    await user.type(screen.getByLabelText('Пошук курсів'), 'postgres');
    await user.click(screen.getByRole('button', { name: 'Знайти' }));
    expect(query).toBe('postgres');

    // A sortable header sets its key; the status column is not sortable (no button).
    await user.click(screen.getByRole('button', { name: /Студенти/ }));
    expect(sort).toBe('students');
    expect(screen.queryByRole('button', { name: /^Статус/ })).toBeNull();

    await user.click(screen.getByRole('button', { name: 'Далі' }));
    expect(page).toBe(2);
  });

  it('shows the localized empty state inside the card, without a table or pager, in English', () => {
    render(
      <I18nProvider instance={createI18n('en')}>
        <CourseListView
          page={{
            items: [],
            pagination: { limit: 20, offset: 0, total: 0 },
            statusCounts: { total: 0, draft: 0, published: 0, archived: 0 },
          }}
          state={{ page: 1, search: '', sort: 'recent' }}
          {...noop}
        />
      </I18nProvider>,
    );

    expect(screen.getByRole('heading', { name: 'No courses yet' })).toBeTruthy();
    expect(screen.queryByRole('table')).toBeNull();
    // The empty slot replaces the table and its pager (the proto's editorial empty state);
    // the filter bar above still lets the admin clear a search or change the status.
    expect(screen.queryByRole('navigation', { name: 'Course pages' })).toBeNull();
    expect(screen.getByRole('textbox', { name: 'Search courses' })).toBeTruthy();
    expect(screen.queryByRole('button', { name: 'Clear filters' })).toBeNull();
  });

  it('offers a distinct filtered-empty state that clears the active filters', async () => {
    const user = userEvent.setup();
    let cleared = 0;
    render(
      <I18nProvider instance={createI18n('en')}>
        <CourseListView
          page={{
            items: [],
            pagination: { limit: 20, offset: 0, total: 0 },
            statusCounts: { total: 21, draft: 8, published: 10, archived: 3 },
          }}
          state={{ page: 1, search: 'nothing-matches', sort: 'recent' }}
          {...noop}
          onClearFilters={() => {
            cleared += 1;
          }}
        />
      </I18nProvider>,
    );

    expect(screen.getByRole('heading', { name: 'No courses found' })).toBeTruthy();
    expect(screen.queryByRole('heading', { name: 'No courses yet' })).toBeNull();
    await user.click(screen.getByRole('button', { name: 'Clear filters' }));
    expect(cleared).toBe(1);
  });

  it('scopes row actions to status and confirms a draft delete', async () => {
    const user = userEvent.setup();
    const publishedCourse: AdminCourse = {
      id: asBrandedId('10000000-0000-4000-8000-000000000002', 'course'),
      slug: 'systems-design',
      title: 'Проєктування систем',
      status: 'published',
      instructorId: '20000000-0000-4000-8000-000000000002',
      updatedAt: '2026-10-01T09:00:00.000000Z',
      lessonCount: 24,
      studentCount: 96,
      price: { minor: 149900, currency: 'UAH' },
    };
    const page: AdminCoursePage = {
      items: [draftCourse, publishedCourse],
      pagination: { limit: 20, offset: 0, total: 2 },
      statusCounts: { total: 2, draft: 1, published: 1, archived: 0 },
    };
    let deleted: AdminCourse | undefined;
    render(
      <CourseListView
        page={page}
        state={{ page: 1, search: '', sort: 'recent' }}
        {...noop}
        onDelete={async (course) => {
          deleted = course;
        }}
      />,
    );

    const draftRow = screen.getByRole('row', { name: /Архітектура Go/ });
    expect(within(draftRow).getByRole('link', { name: /Редагувати курс/ })).toBeTruthy();
    expect(within(draftRow).queryByRole('link', { name: /Переглянути курс/ })).toBeNull();
    const publishedRow = screen.getByRole('row', { name: /Проєктування систем/ });
    // Preview links to the public page; delete is absent on a published course.
    expect(
      within(publishedRow)
        .getByRole('link', { name: /Переглянути курс/ })
        .getAttribute('href'),
    ).toBe('/courses/systems-design');
    expect(within(publishedRow).queryByRole('button', { name: /Видалити/ })).toBeNull();

    // Deleting a draft asks for confirmation, then forwards the row and reports success.
    await user.click(within(draftRow).getByRole('button', { name: /Видалити чернетку/ }));
    await user.click(screen.getByRole('button', { name: 'Підтвердити видалення курсу' }));
    expect(deleted?.slug).toBe('go-architecture');
    expect(await screen.findByText('Курс видалено.')).toBeTruthy();
  });

  it('R37 retries an unavailable primary list read and renders the recovered route', async () => {
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const admin = createAdminQueries(createAuthoring(createApiAdminCourseRepository()));
    const query = adminCourseQuery({ page: 1, search: '', sort: 'recent' });
    let reads = 0;
    server.use(
      http.get(`${apiOrigin()}/api/v1/admin/courses`, () => {
        reads += 1;
        return reads === 1
          ? HttpResponse.json(
              {
                error: {
                  code: 'dependency_unavailable',
                  message: 'x',
                  requestId: 'a'.repeat(32),
                },
              },
              { status: 503 },
            )
          : HttpResponse.json(firstPage);
      }),
    );
    const root = createRootRoute({ component: Outlet });
    const list = createRoute({
      getParentRoute: () => root,
      path: '/admin',
      loader: ({ abortController }) =>
        ensureRouteQuery(queryClient, admin.list(query), abortController.signal),
      component: () => <p>Courses recovered</p>,
      errorComponent: CourseListFailure,
    });
    const router = createRouter({
      routeTree: root.addChildren([list]),
      history: createMemoryHistory({ initialEntries: ['/admin'] }),
      defaultPendingMs: 0,
      defaultPendingMinMs: 0,
      Wrap: ({ children }) => <I18nProvider instance={createI18n('en')}>{children}</I18nProvider>,
    });

    await router.load();
    render(<RouterProvider router={router} />);

    const error = await screen.findByRole('alert');
    expect(within(error).getByRole('heading', { name: 'Could not load courses' })).toBeTruthy();
    await userEvent.click(within(error).getByRole('button', { name: 'Try again' }));

    expect(await screen.findByText('Courses recovered')).toBeTruthy();
    expect(reads).toBe(2);
  });
});
