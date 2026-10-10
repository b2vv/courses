import { afterEach, describe, expect, it } from '@rstest/core';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { cleanup, render, screen, within } from '@testing-library/react';
import axe from 'axe-core';
import { createAdminStatsQueries } from '../../app/composition/admin-stats-queries';
import { createI18n, I18nProvider } from '../../app/providers/i18n';
import type { Snapshot } from '../../domain/admin/stats';
import type { AdminStatsRepository } from '../../ports/AdminStatsRepository';
import { AnalyticsView } from './AnalyticsView';

afterEach(cleanup);

const stubRepository: AdminStatsRepository = {
  snapshot: async () => ({ ok: false, error: { code: 'invalid_response' } }),
  timeseries: async () => ({ ok: true, value: [] }),
};
const stats = createAdminStatsQueries(stubRepository);

function makeSnapshot(overrides: Partial<Snapshot> = {}): Snapshot {
  return {
    users: {
      active: 5,
      blocked: 1,
      deleted: 2,
      byRole: { student: 3, instructor: 2, admin: 1 },
      verified: 4,
      twoFactor: 1,
    },
    courses: { draft: 1, published: 2, archived: 1 },
    grants: { active: 3, blocked: 1, revoked: 1 },
    orders: { pending: 1, paid: 2, failed: 1, revenueMinorUAH: 1200000 },
    reports: { open: 1, resolved: 2 },
    media: {
      inProgress: 1,
      ready: 2,
      failed: 1,
      totalDurationSeconds: 247,
      avgDurationSeconds: null,
      withDurationCount: 0,
    },
    engagement: { activeLearners30d: 4, avgProgress: null, completions: 2 },
    instructors: [
      {
        instructorId: '10000000-0000-4000-8000-0000000000a1',
        displayName: 'Ada',
        publishedCourses: 2,
        activeEnrollments: 3,
        revenueMinorUAH: 1000000,
      },
      {
        instructorId: '10000000-0000-4000-8000-0000000000a2',
        displayName: 'Bob',
        publishedCourses: 1,
        activeEnrollments: 1,
        revenueMinorUAH: 300000,
      },
    ],
    platform: { singleInstructor: false },
    ...overrides,
  };
}

function renderView(snapshot: Snapshot) {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <I18nProvider instance={createI18n('uk')}>
      <QueryClientProvider client={client}>
        <AnalyticsView snapshot={snapshot} stats={stats} />
      </QueryClientProvider>
    </I18nProvider>,
  );
}

describe('AnalyticsView', () => {
  it('shows the leaderboard in the given order when the platform has many instructors', async () => {
    const { container } = renderView(makeSnapshot());
    const table = screen.getByRole('table', { name: 'Рейтинг інструкторів' });
    const rows = within(table).getAllByRole('row');
    // Header row + two instructor rows, in the order supplied (revenue DESC).
    expect(rows).toHaveLength(3);
    expect(within(rows[1] as HTMLElement).getByText('Ada')).toBeTruthy();
    expect(within(rows[2] as HTMLElement).getByText('Bob')).toBeTruthy();
    // The instructor dimension is present in single-instructor OFF mode: the
    // count tile label and the leaderboard section heading.
    expect(screen.getAllByText('Інструктори')).toHaveLength(2);
    expect((await axe.run(container)).violations).toEqual([]);
  });

  it('hides the leaderboard and the instructor tile for a single-instructor platform', () => {
    renderView(makeSnapshot({ platform: { singleInstructor: true } }));
    expect(screen.queryByRole('table', { name: 'Рейтинг інструкторів' })).toBeNull();
    expect(screen.queryAllByText('Інструктори')).toHaveLength(0);
  });

  it('renders an em dash for a null average, never a fabricated zero', () => {
    renderView(makeSnapshot());
    const progressTile = screen.getByText('Середній прогрес').closest('.ui-stat');
    expect(progressTile).not.toBeNull();
    expect(within(progressTile as HTMLElement).getByText('—')).toBeTruthy();
    const durationTile = screen.getByText('Середнє відео (секунди)').closest('.ui-stat');
    expect(within(durationTile as HTMLElement).getByText('—')).toBeTruthy();
  });

  it('shows an honest placeholder for the deferred metrics', () => {
    renderView(makeSnapshot());
    expect(screen.getByText('Догляд, теги та популярність')).toBeTruthy();
  });
});
