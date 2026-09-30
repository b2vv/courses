import { test, expect, type Locator, type Page } from '@playwright/test';
import axe from 'axe-core';
import { authFixtures, authStatePath, recordRequest, trackPageRequests } from './auth';
declare global {
  interface Window {
    axe: typeof axe;
  }
}

let finishRequestTracking = () => Promise.resolve();

async function axeAudit(page: Page, scope: 'wcag' | 'contrast') {
  return page.evaluate(async (auditScope) => {
    const results =
      auditScope === 'contrast'
        ? await window.axe.run(document, {
            runOnly: { type: 'rule', values: ['color-contrast'] },
          })
        : await window.axe.run(document, {
            runOnly: ['wcag2a', 'wcag2aa', 'wcag21aa'],
          });
    const findings = (items: typeof results.violations) =>
      items.flatMap((item) =>
        item.nodes.map((node) => {
          const rawMessageKey = [...node.any, ...node.all, ...node.none].flatMap((check) => {
            const checkData: unknown = check.data;
            if (typeof checkData !== 'object' || checkData === null) return [];
            const messageKey: unknown = Reflect.get(checkData, 'messageKey');
            return typeof messageKey === 'string' ? [messageKey] : [];
          })[0];
          return {
            id: item.id,
            messageKey: typeof rawMessageKey === 'string' ? rawMessageKey : null,
            target: node.target.join(' '),
          };
        }),
      );
    return { violations: findings(results.violations), incomplete: findings(results.incomplete) };
  }, scope);
}

type AxeFinding = Awaited<ReturnType<typeof axeAudit>>['incomplete'][number];

function expectOnlySelectChevronIncomplete(incomplete: AxeFinding[]) {
  // Axe declines a verdict for Select's decorative non-BMP chevron. Remove this pin when axe can
  // evaluate that glyph or when the UI-kit markup changes and makes the selector suffix obsolete.
  const isSelectChevron = ({ id, messageKey, target }: AxeFinding) =>
    id === 'color-contrast' &&
    messageKey === 'nonBmp' &&
    target.endsWith('span[aria-hidden="true"]');
  expect(incomplete).toHaveLength(1);
  expect(incomplete.filter(isSelectChevron)).toHaveLength(1);
}

async function expectVisibleFocus(control: Locator, controls: Locator) {
  await expect(control).toBeFocused();
  const ring = await control.evaluate(
    (node, controlsElement) => {
      const style = getComputedStyle(node);
      const controlsStyle = getComputedStyle(controlsElement);
      return {
        background: controlsStyle.backgroundColor,
        color: style.outlineColor,
        style: style.outlineStyle,
        width: Number.parseFloat(style.outlineWidth),
      };
    },
    await controls.elementHandle(),
  );
  expect(ring.style).not.toBe('none');
  expect(ring.width).toBeGreaterThanOrEqual(3);
  expect(ring.color).not.toBe('transparent');
  expect(ring.color).not.toBe('rgba(0, 0, 0, 0)');
  expect(ring.color).not.toBe(ring.background);
}

async function expectPlaybackLayout(page: Page, controls: Locator, playbackControls: Locator[]) {
  const controlsBox = await controls.boundingBox();
  if (controlsBox === null) throw new Error('Lesson playback controls are not rendered');
  for (const control of playbackControls) {
    const box = await control.boundingBox();
    if (box === null) throw new Error('Lesson playback control is not rendered');
    expect(box.width).toBeGreaterThanOrEqual(44);
    expect(box.height).toBeGreaterThanOrEqual(44);
    expect(box.x).toBeGreaterThanOrEqual(controlsBox.x - 1);
    expect(box.y).toBeGreaterThanOrEqual(controlsBox.y - 1);
    expect(box.x + box.width).toBeLessThanOrEqual(controlsBox.x + controlsBox.width + 1);
    expect(box.y + box.height).toBeLessThanOrEqual(controlsBox.y + controlsBox.height + 1);
  }
  expect(
    await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth + 1),
  ).toBe(true);
  const layoutFailures = await controls.evaluate((container) => {
    const children = Array.from(container.children).map((element) => ({
      label: element.textContent?.trim() ?? element.tagName,
      rectangle: element.getBoundingClientRect(),
    }));
    const overlaps = children.flatMap((left, index) =>
      children.slice(index + 1).flatMap((right) => {
        const horizontal =
          Math.min(left.rectangle.right, right.rectangle.right) -
          Math.max(left.rectangle.left, right.rectangle.left);
        const vertical =
          Math.min(left.rectangle.bottom, right.rectangle.bottom) -
          Math.max(left.rectangle.top, right.rectangle.top);
        return horizontal > 1 && vertical > 1 ? [`overlap: ${left.label}/${right.label}`] : [];
      }),
    );
    const clippedText = Array.from(container.querySelectorAll('*')).flatMap((element) => {
      const hasDirectText = Array.from(element.childNodes).some(
        (node) => node.nodeType === Node.TEXT_NODE && node.textContent?.trim() !== '',
      );
      if (!hasDirectText) return [];
      const clippedHorizontally = element.scrollWidth > element.clientWidth + 1;
      const clippedVertically = element.scrollHeight > element.clientHeight + 1;
      return clippedHorizontally || clippedVertically
        ? [`clipped: ${element.tagName}.${element.className}`]
        : [];
    });
    return [...overlaps, ...clippedText];
  });
  expect(layoutFailures).toEqual([]);
}

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
  expect(await axeAudit(page, 'wcag')).toEqual({ violations: [], incomplete: [] });
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

    const administration = page.getByRole('link', { name: 'В адмінку' });
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
          price: null,
          paid: false,
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
    const controls = page.locator('.lesson-player-controls');
    const play = page.getByRole('button', { name: 'Відтворити', exact: true });
    const seek = page.getByRole('slider', { name: 'Позиція відтворення', exact: true });
    const quality = page.getByRole('button', { name: /Якість/ });
    const playbackControls = [play, seek, quality];
    for (const control of playbackControls) await expect(control).toBeVisible();
    await expectPlaybackLayout(page, controls, playbackControls);

    await page.addStyleTag({ content: 'html { font-size: 200% !important; }' });
    const lightAppearance = await controls.evaluate((element) => ({
      colorScheme: getComputedStyle(document.documentElement).colorScheme,
      surface: getComputedStyle(element).backgroundColor,
    }));
    expect(lightAppearance.colorScheme).toBe('light');

    for (
      let index = 0;
      index < 12 && !(await play.evaluate((node) => node === document.activeElement));
      index += 1
    )
      await page.keyboard.press('Tab');
    for (const [index, control] of playbackControls.entries()) {
      if (index > 0) await page.keyboard.press('Tab');
      await expectVisibleFocus(control, controls);
    }
    const auditedVideoState = await page.locator('video').evaluate((video) => {
      if (!(video instanceof HTMLVideoElement)) throw new Error('Expected the lesson video');
      return {
        readyState: video.readyState,
        videoWidth: video.videoWidth,
        videoHeight: video.videoHeight,
      };
    });
    console.log(
      `B25 video at contrast audit: readyState=${auditedVideoState.readyState}, videoWidth=${auditedVideoState.videoWidth}, videoHeight=${auditedVideoState.videoHeight}`,
    );
    expect(auditedVideoState).toEqual({ readyState: 0, videoWidth: 0, videoHeight: 0 });
    const lightContrast = await axeAudit(page, 'contrast');
    expect(lightContrast.violations).toEqual([]);
    expectOnlySelectChevronIncomplete(lightContrast.incomplete);
    await expectPlaybackLayout(page, controls, playbackControls);

    await page.emulateMedia({ colorScheme: 'dark', reducedMotion: 'reduce' });
    const darkAppearance = await controls.evaluate((element) => ({
      colorScheme: getComputedStyle(document.documentElement).colorScheme,
      surface: getComputedStyle(element).backgroundColor,
    }));
    expect(darkAppearance.colorScheme).toBe('dark');
    expect(darkAppearance.surface).not.toBe(lightAppearance.surface);
    await page.keyboard.press('Shift+Tab');
    await page.keyboard.press('Shift+Tab');
    for (const [index, control] of playbackControls.entries()) {
      if (index > 0) await page.keyboard.press('Tab');
      await expectVisibleFocus(control, controls);
    }
    const darkContrast = await axeAudit(page, 'contrast');
    expect(darkContrast.violations).toEqual([]);
    expectOnlySelectChevronIncomplete(darkContrast.incomplete);
    await expectPlaybackLayout(page, controls, playbackControls);

    await page.getByRole('button', { name: 'EN' }).click();
    await expect(page.locator('html')).toHaveAttribute('lang', 'en');
    const englishControls = [
      page.getByRole('button', { name: 'Play', exact: true }),
      page.getByRole('slider', { name: 'Playback position', exact: true }),
      page.getByRole('button', { name: /Quality/ }),
    ];
    for (const control of englishControls) await expect(control).toBeVisible();
    await expectPlaybackLayout(page, controls, englishControls);
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
