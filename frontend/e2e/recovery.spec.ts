import { test, expect } from '@playwright/test';
import { execFile } from 'node:child_process';
import { promisify } from 'node:util';

const execute = promisify(execFile);
async function compose(...args: string[]) {
  return execute('docker', ['compose', ...args], {
    cwd: new URL('../../', import.meta.url).pathname,
    timeout: 90_000,
    maxBuffer: 1024 * 1024,
  });
}
test.describe.configure({ mode: 'serial', timeout: 120_000 });

test('SSR API failure is visible and retry preserves search after recovery', async ({ page }) => {
  const url = '/courses?q=Go&sort=title&limit=1&offset=0';
  try {
    await compose('stop', 'api');
    const response = await page.goto(url);
    expect(response?.status()).toBeGreaterThanOrEqual(500);
    await expect(page.getByText('Не вдалося завантажити курси')).toBeVisible();
    await expect(page.getByRole('heading', { name: 'Каталог курсів' })).toHaveCount(0);
    const retry = page.getByRole('link', { name: 'Спробувати ще раз' });
    await expect(retry).toHaveAttribute('href', url);
    await page.screenshot({
      path: '../docs/foundation/screenshots/catalog-error.png',
      fullPage: true,
    });
    await compose('up', '-d', '--wait', 'api');
    await retry.click();
    await expect(
      page.getByRole('heading', { name: 'Go: фундамент надійного бекенду' }),
    ).toBeVisible();
    expect(new URL(page.url()).searchParams.get('q')).toBe('Go');
    expect(new URL(page.url()).searchParams.get('sort')).toBe('title');
    expect(new URL(page.url()).searchParams.get('limit')).toBe('1');
    expect(new URL(page.url()).searchParams.get('offset')).toBe('0');
  } finally {
    await compose('up', '-d', '--wait', 'api');
  }
});

test('PostgreSQL outage produces a safe API error and recovers', async ({ request }) => {
  try {
    await compose('stop', 'postgres');
    const response = await request.get('/api/v1/courses');
    expect(response.status()).toBe(503);
    const body: unknown = await response.json();
    expect(body).toEqual({
      error: {
        code: 'dependency_unavailable',
        message: 'Course catalog is temporarily unavailable',
        requestId: expect.any(String),
      },
    });
  } finally {
    await compose('up', '-d', '--wait', 'postgres');
  }
  await expect.poll(async () => (await request.get('/api/v1/courses')).status()).toBe(200);
});

test('restart retains seeded courses and Nginx never stores query or Referer', async ({
  request,
}) => {
  await compose('restart', 'api', 'media-worker', 'frontend', 'nginx');
  await compose('up', '-d', '--wait');
  const response = await request.get('/api/v1/courses');
  expect(response.status()).toBe(200);
  const page: unknown = await response.json();
  expect(page).toMatchObject({ pagination: { total: 3 } });

  const marker = 'courses-log-probe-' + Date.now();
  const since = new Date().toISOString();
  const burst = await Promise.all(
    Array.from({ length: 60 }, () =>
      request.get(`/api/v1/missing?probe=${marker}`, {
        headers: { Referer: `https://example.test/?probe=${marker}` },
      }),
    ),
  );
  expect(burst.some((item) => item.status() === 429)).toBe(true);
  try {
    await compose('stop', 'api');
    expect(
      (
        await request.get(`/api/v1/courses?probe=${marker}`, {
          headers: { Referer: `https://example.test/?probe=${marker}` },
        })
      ).status(),
    ).toBeGreaterThanOrEqual(500);
  } finally {
    await compose('up', '-d', '--wait', 'api');
  }
  const logs = await compose('logs', '--since', since, 'nginx');
  expect(logs.stdout + logs.stderr).not.toContain(marker);
});
