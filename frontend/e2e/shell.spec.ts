import { test, expect } from '@playwright/test';

test('SSR shell renders without JavaScript', async ({ browser, baseURL }) => {
  const context = await browser.newContext({
    javaScriptEnabled: false,
    baseURL: baseURL ?? 'http://localhost:8080',
  });
  const page = await context.newPage();
  await page.goto('/');
  await expect(page.getByRole('heading', { level: 1 })).toContainText('Навички, які');
  await page.getByRole('link', { name: 'Переглянути курси' }).click();
  await expect(page.getByRole('heading', { name: 'Каталог курсів' })).toBeVisible();
  await context.close();
});

test('shell hydrates and has keyboard skip navigation', async ({ page }) => {
  const errors: string[] = [];
  page.on('pageerror', (error) => errors.push(error.message));
  await page.goto('/');
  await page.keyboard.press('Tab');
  await expect(page.getByRole('link', { name: 'До основного вмісту' })).toBeFocused();
  await page.keyboard.press('Enter');
  await expect(page.locator('#main')).toBeFocused();
  expect(errors).toEqual([]);
});
