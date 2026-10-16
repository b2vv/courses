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
import { asBrandedId } from '../../domain/course/course';
import type { AdminCoursePage, CourseStatus } from '../../domain/course/course';
import { CourseListFailure, CourseListView } from './CourseListView';

const firstPage: AdminCoursePage = {
  items: [
    {
      id: asBrandedId('10000000-0000-4000-8000-000000000001', 'course'),
      slug: 'go-architecture',
      title: 'Архітектура Go',
      status: 'draft',
      instructorId: '20000000-0000-4000-8000-000000000002',
      updatedAt: '2026-09-05T10:15:30.000000Z',
      lessonCount: 12,
      studentCount: 0,
      price: null,
    },
  ],
  pagination: { limit: 20, offset: 0, total: 21 },
  statusCounts: { total: 21, draft: 8, published: 10, archived: 3 },
};

const server = setupServer();
beforeAll(() => server.listen({ onUnhandledRequest: 'error' }));
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('admin course list', () => {
  it('renders title, slug, localized status, owner and updated time with an editor link', () => {
    render(
      <CourseListView
        page={firstPage}
        state={{ page: 1, search: '' }}
        onStatusChange={() => {}}
        onPageChange={() => {}}
      />,
    );

    const table = screen.getByRole('table', { name: 'Курси' });
    expect(within(table).getByRole('columnheader', { name: 'Власник' })).toBeTruthy();
    expect(within(table).getByText('go-architecture')).toBeTruthy();
    expect(within(table).getByText('Чернетка')).toBeTruthy();
    expect(within(table).getByText('20000000-0000-4000-8000-000000000002')).toBeTruthy();
    expect(within(table).getByRole('link', { name: 'Архітектура Go' }).getAttribute('href')).toBe(
      '/admin/courses/10000000-0000-4000-8000-000000000001',
    );
    expect(within(table).getByText(/2026/).closest('time')?.getAttribute('datetime')).toBe(
      '2026-09-05T10:15:30.000000Z',
    );
  });

  it('changes the status filter and moves to the next page through accessible controls', async () => {
    const user = userEvent.setup();
    let status: CourseStatus | undefined;
    let page = 1;
    render(
      <CourseListView
        page={firstPage}
        state={{ page: 1, search: '' }}
        onStatusChange={(nextStatus) => {
          status = nextStatus;
        }}
        onPageChange={(nextPage) => {
          page = nextPage;
        }}
      />,
    );

    await user.click(screen.getByRole('tab', { name: 'Опубліковано' }));
    expect(status).toBe('published');
    await user.click(screen.getByRole('button', { name: 'Далі' }));
    expect(page).toBe(2);
  });

  it('renders the localized empty state and both pagination buttons in English', () => {
    render(
      <I18nProvider instance={createI18n('en')}>
        <CourseListView
          page={{
            items: [],
            pagination: { limit: 20, offset: 0, total: 0 },
            statusCounts: { total: 0, draft: 0, published: 0, archived: 0 },
          }}
          state={{ page: 1, search: '' }}
          onStatusChange={() => {}}
          onPageChange={() => {}}
        />
      </I18nProvider>,
    );

    expect(screen.getByRole('heading', { name: 'No courses yet' })).toBeTruthy();
    expect(screen.queryByRole('table')).toBeNull();
    const pages = screen.getByRole('navigation', { name: 'Course pages' });
    expect(within(pages).getByRole('button', { name: 'Back' })).toBeTruthy();
    expect(within(pages).getByRole('button', { name: 'Next' })).toBeTruthy();
    expect(pages.textContent).toContain('No pages');
  });

  it('R37 retries an unavailable primary list read and renders the recovered route', async () => {
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const admin = createAdminQueries(createAuthoring(createApiAdminCourseRepository()));
    const query = adminCourseQuery({ page: 1, search: '' });
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
