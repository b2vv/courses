import { execFile } from 'node:child_process';
import { randomUUID } from 'node:crypto';
import { fileURLToPath } from 'node:url';
import { promisify } from 'node:util';
import { test, expect } from '@playwright/test';
import { deleteE2ECourse } from './course-cleanup';
import { authStatePath } from './auth';

const execute = promisify(execFile);
const root = fileURLToPath(new URL('../../', import.meta.url));

// Fixture-only SQL through Compose psql, as in the other specs: a publication date 31 days in
// the past cannot be made through the product. What is checked - the catalog list, its count
// and the course page - goes through the production HTTP and UI path.
async function sql(statement: string) {
  await execute(
    'docker',
    [
      'compose',
      'exec',
      '-T',
      'postgres',
      'psql',
      '-v',
      'ON_ERROR_STOP=1',
      '-U',
      'courses',
      '-d',
      'courses',
      '-c',
      statement,
    ],
    { cwd: root, timeout: 60_000 },
  );
}

test('catalog and detail include PostgreSQL course data in initial HTML', async ({
  browser,
  baseURL,
}) => {
  const context = await browser.newContext({
    javaScriptEnabled: false,
    baseURL: baseURL ?? 'http://localhost:8080',
  });
  const page = await context.newPage();
  const response = await page.goto('/courses');
  expect(response?.status()).toBe(200);
  await expect(
    page.getByRole('heading', { name: 'Go: фундамент надійного бекенду' }),
  ).toBeVisible();
  await expect(page.getByText('Чернетка курсу')).toHaveCount(0);
  await page
    .getByRole('heading', { name: 'Go: фундамент надійного бекенду' })
    .getByRole('link')
    .click();
  await expect(page.getByRole('heading', { level: 1 })).toHaveText(
    'Go: фундамент надійного бекенду',
  );
  await expect(page.getByRole('heading', { name: 'Про курс' })).toBeVisible();

  // Without JavaScript the search form still works as a native GET form: the typed text and
  // the chosen order go into the address and the server renders the result.
  await page.goto('/courses');
  await page.getByLabel('Знайти курс').fill('React');
  await page.getByLabel('Сортування').selectOption('title');
  await page.getByRole('button', { name: 'Знайти', exact: true }).click();
  await expect(page).toHaveURL(/[?&]q=React&sort=title/);
  await expect(page.getByRole('heading', { name: 'React: доступні інтерфейси' })).toBeVisible();
  await expect(page.locator('.course-list').getByRole('listitem')).toHaveCount(1);
  await context.close();
});

// B-13: this test used to read the second card of the *global* catalog - `?limit=1&offset=1`
// - and name the course it expected there. Any sibling file that publishes a course takes
// position zero with `published_at = now` and shifts that window by one, which is not a product
// defect: `docs/foundation/assumptions.md:21` says offset pagination may shift between requests
// during publication. So the assertion demanded stability the product never promised, and
// `workers: 1` was the only thing holding it up.
//
// The window is now scoped with `q`, and the scope is what makes it deterministic: no
// run-scoped fixture matches `React` or `Go` in its title or summary (the catalog's full-text
// search covers both, `courses.sql.go:37-38`), so the claim holds whatever else the catalog
// contains. Row H04 asks
// that reopening a q/sort/limit/offset URL show the same filtered order - which is what is
// asserted here - not that a particular course sit at a particular global offset.
//
// `workers: 1` stays, and not for this reason: five workers under load produced browser timeouts
// and 504s (`playwright.config.ts:9`). That measurement is untouched by this change, so the
// number keeps its own independent cause.
test('search and pagination live in the URL and recover from an empty page', async ({ page }) => {
  await page.goto('/courses?q=React&limit=1&offset=0&sort=recent');
  await expect(page.getByRole('heading', { name: 'React: доступні інтерфейси' })).toBeVisible();
  // The limit is honoured, not merely accepted: the scope holds exactly one course, and one card
  // is what the page renders.
  // Count the cards in the course list, not every h2 on the page: the catalog
  // also carries the value strip's own (visually hidden) heading.
  await expect(page.locator('.course-list').getByRole('listitem')).toHaveCount(1);

  // Offset one past the end of a scoped result is the empty page, and the recovery link works
  // from there - the same claim the ceiling case makes, on a boundary a fixture can reach.
  // The one way out leads to the last page that has courses, with the same query; no page
  // links are offered from an empty page (design-pass C2 #1, #7).
  await page.goto('/courses?q=React&limit=1&offset=1&sort=recent');
  await expect(page.getByText('На цій сторінці немає курсів')).toBeVisible();
  await expect(page.getByRole('navigation', { name: 'Сторінки каталогу' })).toHaveCount(0);
  await page.getByRole('link', { name: 'До останньої сторінки' }).click();
  await expect(page.getByRole('heading', { name: 'React: доступні інтерфейси' })).toBeVisible();
  expect(Number(new URL(page.url()).searchParams.get('offset'))).toBe(0);

  await page.goto('/courses?q=React&limit=1&offset=1&sort=recent');
  await page.getByLabel('Знайти курс').fill('Go');
  await page.getByRole('button', { name: 'Знайти', exact: true }).click();
  await expect(
    page.getByRole('heading', { name: 'Go: фундамент надійного бекенду' }),
  ).toBeVisible();
  expect(Number(new URL(page.url()).searchParams.get('offset') ?? 0)).toBe(0);
  expect(new URL(page.url()).searchParams.get('q')).toBe('Go');

  await page.goto('/courses?limit=1&offset=10000');
  await expect(page.getByText('На цій сторінці немає курсів')).toBeVisible();
  await page.getByRole('link', { name: 'До останньої сторінки' }).click();
  await expect(page.locator('.course-list').getByRole('listitem')).toHaveCount(1);

  // A search with no result offers to clear it, and no pages.
  await page.goto('/courses?q=zzz-no-such-course');
  await expect(page.getByText('Курсів не знайдено')).toBeVisible();
  await expect(page.getByRole('navigation', { name: 'Сторінки каталогу' })).toHaveCount(0);
  await page.getByRole('link', { name: 'Скинути пошук' }).click();
  await expect(page.getByLabel('Знайти курс')).toHaveValue('');
});

test('hidden and absent courses return 404', async ({ page }) => {
  for (const slug of ['draft-course', 'archived-course', 'does-not-exist']) {
    const response = await page.goto(`/courses/${slug}`);
    expect(response?.status()).toBe(404);
    // The same page as the site 404 (C3 #8).
    await expect(page.getByRole('heading', { level: 1, name: 'Курс не знайдено' })).toBeVisible();
  }
});

test('a course without lessons offers only «Скоро буде» (C3 #2)', async ({ browser, baseURL }) => {
  if (baseURL === undefined) throw new Error('Playwright baseURL is required');
  // The seed's published courses have no lessons; the student fixture is a member of each.
  const guest = await browser.newContext({ baseURL, viewport: { width: 390, height: 844 } });
  const student = await browser.newContext({ baseURL, storageState: authStatePath('student') });
  try {
    const page = await guest.newPage();
    await page.goto('/courses/go-backend-foundations');
    const box = page.getByRole('complementary');
    await expect(box.getByRole('heading', { name: 'Скоро буде' })).toBeVisible();
    await expect(box.getByRole('button')).toHaveCount(0);
    await expect(box.getByRole('link')).toHaveCount(0);
    const bounds = await box.boundingBox();
    expect((bounds?.y ?? Infinity) + (bounds?.height ?? 0)).toBeLessThanOrEqual(844);
    const member = await student.newPage();
    await member.goto('/courses/go-backend-foundations');
    const memberBox = member.getByRole('complementary');
    await expect(memberBox.getByRole('heading', { name: 'Ви учасник' })).toBeVisible();
    await expect(memberBox.getByText('Скоро буде')).toBeVisible();
    await expect(memberBox.getByRole('link', { name: 'Почати навчання' })).toHaveCount(0);
  } finally {
    await student.close();
    await guest.close();
  }
});

test('catalog hydrates without duplicate data requests or errors', async ({ page }) => {
  const errors: string[] = [];
  const requests: string[] = [];
  page.on('pageerror', (error) => errors.push(error.message));
  page.on('console', (message) => {
    if (message.type() === 'error') errors.push(message.text());
  });
  page.on('request', (request) => {
    if (request.url().includes('/api/v1/courses')) requests.push(request.url());
  });
  await page.goto('/courses');
  await page.getByLabel('Знайти курс').fill('test hydration');
  await expect(page.getByLabel('Знайти курс')).toHaveValue('test hydration');
  expect(errors).toEqual([]);
  expect(requests).toEqual([]);
});

test('changing the sort applies at once, keeps focus and follows Back', async ({ page }) => {
  await page.goto('/courses?q=React&sort=recent&limit=1&offset=0');
  const field = page.getByLabel('Знайти курс');
  const sort = page.getByLabel('Сортування');
  // The typed text goes with the new order, from the first page.
  await field.fill('Go');
  await sort.focus();
  await sort.selectOption('title');
  await expect(page).toHaveURL(/[?&]q=Go&sort=title&limit=1&offset=0$/);
  await expect(
    page.getByRole('heading', { name: 'Go: фундамент надійного бекенду' }),
  ).toBeVisible();
  await expect(sort).toBeFocused();
  // Back restores the earlier address, and the controls follow it.
  await page.goBack();
  await expect(page).toHaveURL(/[?&]q=React&sort=recent/);
  await expect(page.getByRole('heading', { name: 'React: доступні інтерфейси' })).toBeVisible();
  await expect(field).toHaveValue('React');
  await expect(sort).toHaveValue('recent');
  // Interacting with the field re-renders it; the earlier text must not come back.
  await field.hover();
  await field.click();
  await expect(field).toHaveValue('React');
  await page.goForward();
  await expect(page).toHaveURL(/[?&]q=Go&sort=title/);
  await expect(field).toHaveValue('Go');
  await expect(sort).toHaveValue('title');
  await field.hover();
  await field.click();
  await expect(field).toHaveValue('Go');
});

test('a course with no lessons leaves the catalog 30 days after publication (C2b)', async ({
  page,
  request,
}) => {
  const id = randomUUID();
  // One alphanumeric token, so the catalog search scopes the check to this fixture alone.
  const token = `c2bwindow${id.replaceAll('-', '').slice(0, 12)}`;
  const slug = `c2b-window-${id.slice(0, 8)}`;
  await sql(
    `INSERT INTO courses (id,instructor_id,slug,title,summary,description,status,created_at,updated_at,published_at)
     VALUES ('${id}','40000000-0000-4000-8000-000000000002','${slug}','${token} course','${token} summary','Fixture description','published',now() - interval '40 days',now(),now() - interval '31 days')`,
  );
  try {
    const hidden = await request.get(`/api/v1/courses?q=${token}`);
    expect(hidden.status()).toBe(200);
    expect(await hidden.json()).toMatchObject({ items: [], pagination: { total: 0 } });
    await page.goto(`/courses?q=${token}`);
    await expect(page.getByText('Курсів не знайдено')).toBeVisible();
    // Leaving the list does not close the course's own page.
    const detail = await page.goto(`/courses/${slug}`);
    expect(detail?.status()).toBe(200);
    await expect(page.getByRole('heading', { level: 1 })).toHaveText(`${token} course`);

    await sql(`UPDATE courses SET published_at = now() - interval '29 days' WHERE id = '${id}'`);
    await page.goto(`/courses?q=${token}`);
    const card = page.locator('.course-list').getByRole('listitem');
    await expect(card).toHaveCount(1);
    await expect(card.getByText('Скоро буде')).toBeVisible();
    await expect(page.getByText('1 курс', { exact: true })).toBeVisible();
  } finally {
    await deleteE2ECourse(id);
  }
});

// ---------------------------------------------------------------------------------------------
// C3b (plan §17): React Aria links move through the router. A guest is enough for every case
// here, so no form login is spent (the e2e login budget is 7 of 7).

type Page = import('@playwright/test').Page;

/** Counts main-frame document loads from now on: a move inside the app makes none. */
function countDocuments(page: Page) {
  let documents = 0;
  page.on('request', (request) => {
    if (request.resourceType() === 'document' && request.frame() === page.mainFrame()) {
      documents += 1;
    }
  });
  return () => documents;
}

/** Three fixture courses one search token finds, so `limit=1` gives exactly three pages. */
async function threeCourses() {
  const run = randomUUID().replaceAll('-', '').slice(0, 12);
  const token = `cthreeb${run}`;
  const course = (n: number) => ({
    id: randomUUID(),
    slug: `c3b-${run}-${n}`,
    title: `${token} course ${n}`,
  });
  const courses = [course(1), course(2), course(3)] as const;
  // One statement, so a refused row leaves no partial fixture behind. Distinct publication
  // times give `sort=recent` a fixed order; creation precedes publication (table check).
  await sql(
    `INSERT INTO courses (id,instructor_id,slug,title,summary,description,status,created_at,updated_at,published_at) VALUES ${courses
      .map(
        (course, index) =>
          `('${course.id}','40000000-0000-4000-8000-000000000002','${course.slug}','${course.title}','${token} summary','Fixture description','published',now() - interval '1 hour',now(),now() - interval '${index} minutes')`,
      )
      .join(',')}`,
  );
  return { token, courses };
}

const isCourseRead = (slug: string) => (url: URL) => url.pathname === `/api/v1/courses/${slug}`;

test('C3b a guest moves between the catalog and a course without a page load', async ({
  page,
  context,
}) => {
  const { token, courses } = await threeCourses();
  try {
    await page.goto(`/courses?q=${token}&sort=recent&limit=1&offset=0`);
    await expect(page.getByRole('heading', { name: courses[0].title })).toBeVisible();
    const documents = countDocuments(page);

    // Pages of the catalog.
    await page.getByRole('link', { name: 'Наступна →' }).click();
    await expect(page.getByRole('heading', { name: courses[1].title })).toBeVisible();
    await page.getByRole('link', { name: '← Попередня' }).click();
    await expect(page.getByRole('heading', { name: courses[0].title })).toBeVisible();

    // The card title, then «← Усі курси», then the card's own button.
    await page.getByRole('heading', { name: courses[0].title }).getByRole('link').click();
    await expect(page.getByRole('heading', { level: 1 })).toHaveText(courses[0].title);
    await expect(page).toHaveTitle(`${courses[0].title} — CourseForge`);
    await page.getByRole('link', { name: '← Усі курси' }).click();
    await expect(page.getByRole('heading', { level: 1, name: 'Каталог курсів' })).toBeVisible();
    await expect(page).toHaveTitle('Каталог курсів — CourseForge');
    await page.goBack();
    await expect(page.getByRole('heading', { level: 1 })).toHaveText(courses[0].title);
    await page.goBack();
    await page.locator('.course-list').getByRole('link', { name: 'Відкрити' }).click();
    await expect(page.getByRole('heading', { level: 1 })).toHaveText(courses[0].title);
    expect(documents()).toBe(0);

    // A modifier press stays the browser's: a new tab, this page unchanged.
    await page.getByRole('link', { name: '← Усі курси' }).focus();
    const opened = context.waitForEvent('page');
    await page
      .getByRole('link', { name: '← Усі курси' })
      .click({ modifiers: [process.platform === 'darwin' ? 'Meta' : 'Control'] });
    await (await opened).close();
    await expect(page.getByRole('heading', { level: 1 })).toHaveText(courses[0].title);
    expect(documents()).toBe(0);

    // A course that is gone by the time the card is pressed: the course 404, in the page and
    // the tab, and its way back is a move inside the app too (C3b-5).
    await page.goto(`/courses?q=${token}&sort=recent&limit=1&offset=0`);
    const afterReload = countDocuments(page);
    await page.route(
      (url) => isCourseRead(courses[0].slug)(url),
      (route) =>
        route.fulfill({
          status: 404,
          contentType: 'application/json',
          body: JSON.stringify({
            error: { code: 'course_not_found', message: 'not found', requestId: 'e2e' },
          }),
        }),
    );
    await page.getByRole('heading', { name: courses[0].title }).getByRole('link').click();
    await expect(page.getByRole('heading', { level: 1, name: 'Курс не знайдено' })).toBeVisible();
    await expect(page).toHaveTitle('Сторінку не знайдено — CourseForge');
    await page.getByRole('link', { name: 'Перейти до каталогу' }).click();
    await expect(page.getByRole('heading', { level: 1, name: 'Каталог курсів' })).toBeVisible();
    expect(afterReload()).toBe(0);
  } finally {
    await Promise.all(courses.map(({ id }) => deleteE2ECourse(id)));
  }
});

test('C3b the home page hash link scrolls without a page load', async ({ page }) => {
  await page.goto('/');
  const documents = countDocuments(page);
  await page.getByRole('link', { name: 'Як влаштоване навчання' }).click();
  await expect(page).toHaveURL(/\/#values$/);
  await expect(page.locator('#values')).toBeInViewport();
  expect(documents()).toBe(0);
});

test('C3b the course page loading and failure states are reachable by moving', async ({ page }) => {
  const { token, courses } = await threeCourses();
  try {
    const slug = courses[0].slug;
    await page.goto(`/courses?q=${token}&sort=recent&limit=1&offset=0`);
    // A slow course: the course page's own loading state, then the course.
    await page.route(
      (url) => isCourseRead(slug)(url),
      async (route) => {
        await new Promise((resolve) => setTimeout(resolve, 2_000));
        await route.continue();
      },
    );
    await page.getByRole('heading', { name: courses[0].title }).getByRole('link').click();
    await expect(page.getByRole('status').filter({ hasText: /^Завантажуємо курс$/ })).toBeVisible();
    await expect(page.getByRole('heading', { level: 1 })).toHaveText(courses[0].title);
    await page.unroute((url) => isCourseRead(slug)(url));

    // A failing course: the course page's failure, and its retry is one full load that
    // recovers once the API answers again (the server render is not intercepted).
    await page.goto(`/courses?q=${token}&sort=recent&limit=1&offset=0`);
    await page.route(
      (url) => isCourseRead(slug)(url),
      (route) =>
        route.fulfill({
          status: 503,
          contentType: 'application/json',
          body: JSON.stringify({
            error: { code: 'dependency_unavailable', message: 'down', requestId: 'e2e' },
          }),
        }),
    );
    await page.getByRole('heading', { name: courses[0].title }).getByRole('link').click();
    const alert = page.getByRole('alert');
    await expect(alert.getByRole('heading', { name: 'Не вдалося завантажити курс' })).toBeVisible();
    await expect(page.getByText('Не вдалося завантажити курси')).toHaveCount(0);
    const documents = countDocuments(page);
    await page.unrouteAll();
    await alert.getByRole('link', { name: 'Спробувати ще раз' }).click();
    await expect(page.getByRole('heading', { level: 1 })).toHaveText(courses[0].title);
    expect(documents()).toBe(1);
  } finally {
    await Promise.all(courses.map(({ id }) => deleteE2ECourse(id)));
  }
});

test('C3b focus after a move: the main region, or the control that is still there', async ({
  page,
}) => {
  const { token, courses } = await threeCourses();
  const address = (offset: number) => `/courses?q=${token}&sort=recent&limit=1&offset=${offset}`;
  const main = page.locator('#main');
  const next = page.getByRole('link', { name: 'Наступна →' });
  const previous = page.getByRole('link', { name: '← Попередня' });
  try {
    // A new path: the main region, and the next Tab stays in the page's content.
    await page.goto(`/courses/${courses[0].slug}`);
    await page.getByRole('link', { name: '← Усі курси' }).focus();
    await page.keyboard.press('Enter');
    await expect(page.getByRole('heading', { level: 1, name: 'Каталог курсів' })).toBeVisible();
    await expect(main).toBeFocused();
    await page.keyboard.press('Tab');
    expect(
      await page.evaluate(() => document.getElementById('main')?.contains(document.activeElement)),
    ).toBe(true);

    // «Наступна» from a middle page survives the move and keeps focus.
    await page.goto(address(0));
    await next.focus();
    await page.keyboard.press('Enter');
    await expect(page.getByRole('heading', { name: courses[1].title })).toBeVisible();
    await expect(next).toBeFocused();
    // Onto the last page «Наступна» is gone: the main region, not the body.
    await page.keyboard.press('Enter');
    await expect(page.getByRole('heading', { name: courses[2].title })).toBeVisible();
    await expect(main).toBeFocused();
    // «Попередня» onto the first page is gone too.
    await page.goto(address(1));
    await previous.focus();
    await page.keyboard.press('Enter');
    await expect(page.getByRole('heading', { name: courses[0].title })).toBeVisible();
    await expect(main).toBeFocused();

    // A slow list: the pending frame's form is inert, its loading status is announced, and
    // focus lands on the main region when the list arrives.
    const slowList = (url: URL) =>
      url.pathname === '/api/v1/courses' && url.searchParams.get('q') === token;
    const delay = async (route: import('@playwright/test').Route) => {
      await new Promise((resolve) => setTimeout(resolve, 2_000));
      await route.continue();
    };
    await page.goto(address(1));
    await page.route(slowList, delay);
    await next.focus();
    await page.keyboard.press('Enter');
    const status = page.getByRole('status').filter({ hasText: /^Завантажуємо курси$/ });
    await expect(status).toBeVisible();
    expect(await page.locator('body').ariaSnapshot()).toContain('Завантажуємо курси');
    // React keeps the previous page mounted but hidden while the pending frame shows, so the
    // pending frame is the form that carries `inert`.
    const pendingForm = page.locator('form.catalog-search[inert]');
    await expect(pendingForm).toBeVisible();
    const field = pendingForm.getByLabel('Знайти курс');
    await field.focus();
    await expect(field).not.toBeFocused();
    await expect(page.getByRole('heading', { name: courses[2].title })).toBeVisible();
    await expect(main).toBeFocused();

    // A slow list while the user moves to a header link: focus stays where they put it.
    await page.goto(address(1));
    await next.focus();
    await page.keyboard.press('Enter');
    await expect(status).toBeVisible();
    const headerCatalog = page
      .getByRole('navigation', { name: 'Основна навігація' })
      .getByRole('link', { name: 'Каталог', exact: true });
    await headerCatalog.focus();
    await expect(page.getByRole('heading', { name: courses[2].title })).toBeVisible();
    await expect(headerCatalog).toBeFocused();
    await page.unroute(slowList, delay);
  } finally {
    await Promise.all(courses.map(({ id }) => deleteE2ECourse(id)));
  }
});
