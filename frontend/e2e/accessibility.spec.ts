import { test, expect } from '@playwright/test';
import axe from 'axe-core';
import { authFixtures, authStatePath, recordRequest, trackPageRequests } from './auth';
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
  baseURL,
}) => {
  if (baseURL === undefined) throw new Error('Playwright baseURL is required');
  const disposableSession = await page.request.post('/api/v1/auth/login', {
    data: {
      email: authFixtures.student.email,
      password: authFixtures.student.password,
    },
    headers: { Origin: new URL(baseURL).origin },
  });
  await recordRequest('login');
  expect(disposableSession.status()).toBe(200);

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

test.describe('with a stored admin session', () => {
  test.use({ storageState: authStatePath('admin') });

  test('admin navigation remains separate and target-sized at 200% in dark mode', async ({
    page,
  }) => {
    await page.setViewportSize({ width: 1280, height: 900 });
    await page.emulateMedia({ colorScheme: 'dark', reducedMotion: 'reduce' });
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
});

test.describe('with a stored student session', () => {
  test.use({ storageState: authStatePath('student') });

  test('B24a B25 lesson controls remain keyboard-visible and usable at 200%', async ({ page }) => {
    const lessonId = '30000000-0000-4000-8000-000000000001';
    await page.route('**/api/v1/learn/courses/layout-course', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          id: '10000000-0000-4000-8000-000000000001',
          slug: 'layout-course',
          title: 'Курс для перевірки програвача',
          summary: 'Перевірка доступного компонування.',
          description: 'Макет без завантаження медіа.',
          status: 'published',
          createdAt: '2026-09-08T08:00:00Z',
          updatedAt: '2026-09-08T08:00:00Z',
          publishedAt: '2026-09-08T08:00:00Z',
          modules: [
            {
              id: '20000000-0000-4000-8000-000000000001',
              title: 'Довга назва модуля для перевірки компонування',
              lessons: [
                {
                  id: lessonId,
                  title: 'Доступне відтворення на вузькому екрані',
                  videoState: 'ready',
                  position: 0,
                  completed: false,
                  clientSeq: 0,
                },
              ],
            },
          ],
          continueLessonId: lessonId,
        }),
      });
    });
    await page.setViewportSize({ width: 390, height: 844 });
    await page.emulateMedia({ colorScheme: 'light', reducedMotion: 'reduce' });
    await page.goto(`/learn/layout-course/${lessonId}`);
    await page.addScriptTag({ content: axe.source });
    const contrastViolations = () =>
      page.evaluate(async () => {
        const result = await window.axe.run(document, {
          runOnly: { type: 'rule', values: ['color-contrast'] },
        });
        return result.violations.map((violation) => violation.id);
      });
    expect(await contrastViolations()).toEqual([]);
    await page.emulateMedia({ colorScheme: 'dark', reducedMotion: 'reduce' });
    expect(await contrastViolations()).toEqual([]);
    await page.addStyleTag({ content: 'html { font-size: 200% !important; }' });

    const play = page.getByRole('button', { name: 'Відтворити' });
    const seek = page.getByRole('slider', { name: 'Позиція відтворення' });
    const quality = page.getByRole('button', { name: /Якість/ });
    await expect(play).toBeVisible();
    await expect(seek).toBeVisible();
    await expect(quality).toBeVisible();
    for (
      let index = 0;
      index < 12 && !(await play.evaluate((node) => node === document.activeElement));
      index += 1
    )
      await page.keyboard.press('Tab');
    await expect(play).toBeFocused();
    await page.keyboard.press('Tab');
    await expect(seek).toBeFocused();
    await page.keyboard.press('Tab');
    await expect(quality).toBeFocused();

    const focusRing = await quality.evaluate((node) => {
      const style = getComputedStyle(node);
      return { style: style.outlineStyle, width: Number.parseFloat(style.outlineWidth) };
    });
    expect(focusRing.style).not.toBe('none');
    expect(focusRing.width).toBeGreaterThanOrEqual(3);

    for (const control of [play, seek, quality]) {
      const box = await control.boundingBox();
      if (box === null) throw new Error('Lesson playback control is not rendered');
      expect(Math.min(box.width, box.height)).toBeGreaterThanOrEqual(44);
    }
    expect(
      await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth + 1),
    ).toBe(true);
    const overlap = await page.locator('.lesson-player-controls').evaluate((controls) => {
      const items = Array.from(controls.children).map((element) => element.getBoundingClientRect());
      return items.some((left, index) =>
        items.slice(index + 1).some((right) => {
          const horizontal = Math.min(left.right, right.right) - Math.max(left.left, right.left);
          const vertical = Math.min(left.bottom, right.bottom) - Math.max(left.top, right.top);
          return horizontal > 1 && vertical > 1;
        }),
      );
    });
    expect(overlap).toBe(false);
  });
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
