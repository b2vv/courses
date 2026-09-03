import { test, expect } from '@playwright/test';
test('private areas stay locked until identity exists', async ({ page }) => {
  for (const route of ['/admin', '/admin/courses', '/learn', '/learn/go/intro']) {
    await page.goto(route);
    await expect(page.getByRole('heading', { name: 'Розділ поки недоступний' })).toBeVisible();
    await expect(page.getByRole('link', { name: 'Перейти до входу' })).toHaveAttribute(
      'href',
      '/login',
    );
  }
});
