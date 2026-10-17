import { describe, expect, it, rs } from '@rstest/core';
import {
  createMemoryHistory,
  createRootRoute,
  createRoute,
  createRouter,
  Outlet,
  RouterProvider,
} from '@tanstack/react-router';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import type { ComponentProps } from 'react';
import { createI18n, I18nProvider } from '../../app/providers/i18n';
import type { LearnerRow } from '../../domain/teach/teach';
import { LearnerOverview } from './LearnerOverview';

const courseId = '10000000-0000-4000-8000-000000000001';
const userId = '90000000-0000-4000-8000-000000000001';

const row = {
  userId,
  displayName: 'Марія',
  lastActivityAt: '2026-10-07T10:00:00Z',
  courses: [
    {
      courseId,
      slug: 'intro',
      title: 'Інтро курс',
      state: 'active',
      progressPercent: 42,
      enrolledAt: '2026-10-01T09:00:00Z',
    },
  ],
} satisfies LearnerRow;

async function renderOverview(overrides: Partial<ComponentProps<typeof LearnerOverview>> = {}) {
  const props = {
    entries: [row],
    isLoading: false,
    error: null,
    hasNextPage: false,
    isFetchingNextPage: false,
    filters: {},
    onRetry: rs.fn(),
    onLoadMore: rs.fn(),
    onFiltersChange: rs.fn(),
    ...overrides,
  } satisfies ComponentProps<typeof LearnerOverview>;
  const rootRoute = createRootRoute({ component: Outlet });
  const index = createRoute({
    getParentRoute: () => rootRoute,
    path: '/',
    component: () => <LearnerOverview {...props} />,
  });
  // The course link target must exist so <Link> resolves to a real href at runtime.
  const courseRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/teach/courses/$courseId',
    component: () => <p>roster</p>,
  });
  const router = createRouter({
    routeTree: rootRoute.addChildren([index, courseRoute]),
    history: createMemoryHistory({ initialEntries: ['/'] }),
    Wrap: ({ children }) => <I18nProvider instance={createI18n('uk')}>{children}</I18nProvider>,
  });
  await router.load();
  render(<RouterProvider router={router} />);
  return props;
}

describe('LearnerOverview', () => {
  it('renders a row, its progress, and links each course to its roster (Lo10)', async () => {
    await renderOverview();
    expect(screen.getByRole('heading', { name: 'Марія' })).toBeTruthy();
    const link = screen.getByRole('link', { name: 'Інтро курс' });
    expect(link.getAttribute('href')).toBe(`/teach/courses/${courseId}`);
    const progress = screen.getByRole('progressbar', { name: 'Прогрес' });
    expect(progress.getAttribute('aria-valuenow')).toBe('42');
  });

  it('anonymises a deleted account and marks a learner with no activity', async () => {
    await renderOverview({ entries: [{ ...row, displayName: '', lastActivityAt: null }] });
    expect(screen.getByRole('heading', { name: 'Колишній учасник' })).toBeTruthy();
    expect(screen.getByText('активності ще не було')).toBeTruthy();
  });

  it('shows the empty state', async () => {
    await renderOverview({ entries: [] });
    expect(screen.getByText('Учнів поки немає')).toBeTruthy();
  });

  it('surfaces a named error and retries', async () => {
    const props = await renderOverview({
      entries: [],
      error: new Error('learner_course_forbidden'),
    });
    expect(screen.getByText('Ви не можете переглядати учнів цього курсу.')).toBeTruthy();
    await userEvent.click(screen.getByRole('button', { name: 'Спробувати ще раз' }));
    expect(props.onRetry).toHaveBeenCalled();
  });

  it('loads the next page', async () => {
    const props = await renderOverview({ hasNextPage: true });
    await userEvent.click(screen.getByRole('button', { name: 'Показати ще' }));
    expect(props.onLoadMore).toHaveBeenCalled();
  });

  it('keeps the load-more button enabled while a page is fetching (isPending, not disabled)', async () => {
    await renderOverview({ hasNextPage: true, isFetchingNextPage: true });
    // A native `disabled` on the focused button would drop focus to <body> (B-L7b.1); isPending
    // keeps the element enabled, so the attribute must be absent.
    expect(screen.getByRole('button', { name: 'Показати ще' }).hasAttribute('disabled')).toBe(
      false,
    );
  });

  it('commits a trimmed search to the filters', async () => {
    const props = await renderOverview();
    await userEvent.type(screen.getByRole('searchbox', { name: 'Пошук за ім’ям' }), '  Олена  ');
    await userEvent.click(screen.getByRole('button', { name: 'Шукати' }));
    expect(props.onFiltersChange).toHaveBeenCalledWith({ q: 'Олена' });
  });

  it('clears the course filter', async () => {
    const props = await renderOverview({ filters: { courseId } });
    await userEvent.click(screen.getByRole('button', { name: 'Скинути фільтр за курсом' }));
    expect(props.onFiltersChange).toHaveBeenCalledWith({});
  });
});
