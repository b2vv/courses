import { expect, test, type Page } from '@playwright/test';
import axe from 'axe-core';
import { authStatePath, trackPageRequests } from './auth';

declare global {
  interface Window {
    axe: typeof axe;
  }
}

let finishRequestTracking = () => Promise.resolve();

test.beforeEach(({ page }) => {
  finishRequestTracking = trackPageRequests(page);
});

test.afterEach(async () => {
  await finishRequestTracking();
});

async function wcagViolations(page: Page) {
  await page.addScriptTag({ content: axe.source });
  return page.evaluate(async () =>
    (await window.axe.run(document, { runOnly: ['wcag2a', 'wcag2aa', 'wcag21aa'] })).violations.map(
      (item) => item.id,
    ),
  );
}

async function overflowAtDoubleText(page: Page) {
  await page.setViewportSize({ width: 390, height: 844 });
  await page.addStyleTag({ content: 'html { font-size: 200% !important; }' });
  return page.evaluate(() => document.documentElement.scrollWidth > window.innerWidth + 1);
}

test.describe('with a stored admin session', () => {
  test.use({ storageState: authStatePath('admin') });

  test('Rt-analytics: admin reads real numbers, the leaderboard is hidden, and the page is accessible', async ({
    page,
  }) => {
    await page.goto('/admin/analytics');
    await expect(page.getByRole('heading', { name: 'Аналітика', level: 1 })).toBeVisible();

    // Scenario 2: hero tiles carry real numbers, not placeholders.
    const totalTile = page.locator('.ui-stat', { hasText: 'Усього акаунтів' });
    await expect(totalTile.locator('.ui-stat-value')).toHaveText(/\d/);

    // Scenario 9a: single-instructor is the default, so the leaderboard and the
    // instructor dimension are absent.
    await expect(page.getByRole('table', { name: 'Рейтинг інструкторів' })).toHaveCount(0);
    await expect(page.getByRole('heading', { name: 'Інструктори' })).toHaveCount(0);

    // Scenarios 2/3/6: the trend section renders a chart (plot or empty state).
    await expect(page.getByRole('heading', { name: 'Тренд' })).toBeVisible();
    await expect(page.locator('.ui-chart')).toBeVisible();

    // Scenario 13: deferred metrics are an honest placeholder, no fabricated number.
    await expect(page.getByText('Догляд, теги та популярність')).toBeVisible();

    // Scenario 14: stats.active + blocked equals the unfiltered user-list total.
    const statsResponse = await page.request.get('/api/v1/admin/stats');
    expect(statsResponse.ok()).toBeTruthy();
    const snapshot = (await statsResponse.json()) as {
      users: { active: number; blocked: number };
      platform: { singleInstructor: boolean };
    };
    const usersResponse = await page.request.get('/api/v1/admin/users');
    expect(usersResponse.ok()).toBeTruthy();
    const userPage = (await usersResponse.json()) as { pagination: { total: number } };
    expect(snapshot.users.active + snapshot.users.blocked).toBe(userPage.pagination.total);
    // The default platform flag reaches the payload (9a precondition).
    expect(snapshot.platform.singleInstructor).toBe(true);

    expect(await wcagViolations(page)).toEqual([]);
    expect(await overflowAtDoubleText(page)).toBe(false);
  });
});

test.describe('with a stored instructor session', () => {
  test.use({ storageState: authStatePath('instructor') });

  test('Rt-analytics: a non-admin is refused the analytics page and the stats API', async ({
    page,
  }) => {
    // Scenario 10 (UI): an instructor reaches /admin but the analytics child
    // guard answers not-found.
    await page.goto('/admin/analytics');
    await expect(page.getByRole('heading', { name: 'Сторінку не знайдено' })).toBeVisible();

    // Scenario 10 (API): the backend gate refuses a non-admin with 403.
    const statsResponse = await page.request.get('/api/v1/admin/stats');
    expect(statsResponse.status()).toBe(403);
    const body = (await statsResponse.json()) as { error: { code: string } };
    expect(body.error.code).toBe('grant_forbidden');
  });
});
