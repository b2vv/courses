import { test, expect, type Page } from '@playwright/test';
import axe from 'axe-core';
declare global {
  interface Window {
    axe: typeof axe;
  }
}

const baseURL = process.env.E2E_BASE_URL ?? 'http://localhost:8080';

async function authenticate(
  page: Page,
  credentials: { readonly email: string; readonly password: string },
) {
  const response = await page.request.post('/api/v1/auth/login', {
    data: credentials,
    headers: { Origin: new URL(baseURL).origin },
  });
  expect(response.status()).toBe(200);
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

test('signed-in header controls are keyboard reachable with a visible focus ring', async ({
  page,
}) => {
  await authenticate(page, {
    email: 'student@courseforge.local',
    password: 'courseforge-student-demo',
  });
  await page.goto('/');
  const signOut = page.getByRole('button', { name: 'Вийти' });
  const focusedLabels: string[] = [];
  for (
    let index = 0;
    index < 12 && !(await signOut.evaluate((node) => node === document.activeElement));
    index += 1
  ) {
    await page.keyboard.press('Tab');
    focusedLabels.push(
      await page.evaluate(() => document.activeElement?.textContent?.trim() ?? ''),
    );
  }
  await expect(signOut).toBeFocused();
  expect(focusedLabels).not.toContain('Студентка Курсу');
  const focusRing = await signOut.evaluate((node) => {
    const style = getComputedStyle(node);
    return {
      background: getComputedStyle(document.body).backgroundColor,
      color: style.outlineColor,
      style: style.outlineStyle,
      width: Number.parseFloat(style.outlineWidth),
    };
  });
  expect(focusRing.style).not.toBe('none');
  expect(focusRing.width).toBeGreaterThanOrEqual(3);
  expect(focusRing.color).not.toBe('transparent');
  expect(focusRing.color).not.toBe('rgba(0, 0, 0, 0)');
  expect(focusRing.color).not.toBe(focusRing.background);
  await page.keyboard.press('Enter');
  await expect(page.getByRole('link', { name: 'Увійти' })).toBeVisible();
});

test('admin navigation remains separate and target-sized at 200% in dark mode', async ({
  page,
}) => {
  await page.setViewportSize({ width: 1280, height: 900 });
  await page.emulateMedia({ colorScheme: 'dark', reducedMotion: 'reduce' });
  await authenticate(page, {
    email: 'admin@courseforge.local',
    password: 'courseforge-admin-demo',
  });
  await page.goto('/');
  await page.addStyleTag({ content: 'html { font-size: 200% !important; }' });

  const administration = page.getByRole('link', { name: 'Адміністрування' });
  const signOut = page.getByRole('button', { name: 'Вийти' });
  const administrationBox = await administration.boundingBox();
  const signOutBox = await signOut.boundingBox();
  if (administrationBox === null || signOutBox === null)
    throw new Error('Signed-in navigation controls are not rendered');
  expect(Math.min(administrationBox.width, administrationBox.height)).toBeGreaterThanOrEqual(44);
  expect(Math.min(signOutBox.width, signOutBox.height)).toBeGreaterThanOrEqual(44);

  const overlaps = await page.locator('.header-inner').evaluate((header) => {
    const navigation = header.querySelector('nav');
    if (navigation === null) return ['navigation missing'];
    const findOverlaps = (elements: Element[]) => {
      const items = elements.map((element) => ({
        label: element.textContent,
        rectangle: element.getBoundingClientRect(),
      }));
      return items.flatMap((left, leftIndex) =>
        items.slice(leftIndex + 1).flatMap((right) => {
          const horizontal =
            Math.min(left.rectangle.right, right.rectangle.right) -
            Math.max(left.rectangle.left, right.rectangle.left);
          const vertical =
            Math.min(left.rectangle.bottom, right.rectangle.bottom) -
            Math.max(left.rectangle.top, right.rectangle.top);
          return horizontal > 1 && vertical > 1 ? [`${left.label}/${right.label}`] : [];
        }),
      );
    };
    return [
      ...findOverlaps(Array.from(header.children)),
      ...findOverlaps(Array.from(navigation.children)),
    ];
  });
  expect(overlaps).toEqual([]);
  expect(
    await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth + 1),
  ).toBe(true);
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
