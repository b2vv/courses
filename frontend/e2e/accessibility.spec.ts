import { test, expect } from '@playwright/test';
import axe from 'axe-core';
declare global {
  interface Window {
    axe: typeof axe;
  }
}

test('catalog keyboard navigation and accessible labels', async ({ page }) => {
  await page.goto('/courses');
  await page.getByLabel('Сортування').focus();
  await page.keyboard.press('ArrowDown');
  await page.keyboard.press('Tab');
  await expect(page.getByRole('button', { name: 'Знайти', exact: true })).toBeFocused();
  await page.keyboard.press('Enter');
  await expect(page.getByRole('heading', { name: 'Каталог курсів' })).toBeVisible();
  await page.addScriptTag({ content: axe.source });
  const violations = await page.evaluate(async () => {
    const results = await window.axe.run(document, { runOnly: ['wcag2a', 'wcag2aa', 'wcag21aa'] });
    return results.violations.map((item) => ({ id: item.id, impact: item.impact }));
  });
  expect(violations).toEqual([]);
});

test('locale switch updates copy, html lang and persists across SSR navigation', async ({
  page,
}) => {
  await page.goto('/courses');
  await page.getByRole('button', { name: 'EN' }).click();
  await expect(page.getByRole('heading', { name: 'Course catalog' })).toBeVisible();
  await expect(page.locator('html')).toHaveAttribute('lang', 'en');
  await page.reload();
  await expect(page.getByRole('heading', { name: 'Course catalog' })).toBeVisible();
  await page.getByRole('button', { name: 'UA' }).click();
  await expect(page.getByRole('heading', { name: 'Каталог курсів' })).toBeVisible();
});

test('catalog remains usable on mobile, dark theme and 200% text zoom', async ({ page }) => {
  await page.setViewportSize({ width: 390, height: 844 });
  await page.emulateMedia({ colorScheme: 'dark', reducedMotion: 'reduce' });
  await page.goto('/courses');
  await expect(page.getByRole('heading', { name: 'Каталог курсів' })).toBeVisible();
  await page.screenshot({
    path: '../docs/foundation/screenshots/catalog-mobile-dark.png',
    fullPage: true,
  });
  await page.addStyleTag({ content: 'html { font-size: 200% !important; }' });
  await expect(page.getByRole('button', { name: 'Знайти', exact: true })).toBeVisible();
  const overflowingElements = await page.evaluate(() =>
    Array.from(document.querySelectorAll('body *'))
      .filter((element) => element.getBoundingClientRect().right > window.innerWidth + 1)
      .map((element) => element.tagName + '.' + element.className),
  );
  expect(overflowingElements).toEqual([]);
});

test('capture desktop catalog and course', async ({ page }) => {
  await page.goto('/courses');
  await page.screenshot({
    path: '../docs/foundation/screenshots/catalog-desktop.png',
    fullPage: true,
  });
  await page.goto('/courses/go-backend-foundations');
  await page.screenshot({
    path: '../docs/foundation/screenshots/course-desktop.png',
    fullPage: true,
  });
});
