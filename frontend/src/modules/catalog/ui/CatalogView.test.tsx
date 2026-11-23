import { describe, expect, it } from '@rstest/core';
import {
  createMemoryHistory,
  createRootRoute,
  createRoute,
  createRouter,
  Outlet,
  RouterProvider,
} from '@tanstack/react-router';
import { render, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { RouterProvider as AriaRouterProvider } from 'react-aria-components';
import { createI18n, I18nProvider } from '../../../app/providers/i18n';
import type { CatalogQuery, CourseDetail, CoursePage } from '../../../contracts/course';
import {
  CatalogFailure,
  CatalogListFailure,
  CatalogListLoading,
  CourseFailure,
  CourseLoading,
  CourseNotFound,
} from './CatalogStates';
import { CatalogView } from './CatalogView';
import { CourseView } from './CourseView';

const architectureCourse: CourseDetail = {
  id: '7b823b02-65f6-4875-8e66-65e44fd08329',
  slug: 'go-architecture',
  title: 'Архітектура застосунків Go',
  summary: 'Побудуйте API з чіткими межами модулів.',
  description: 'Працюємо з портами та адаптерами.\nЗавершуємо наскрізним тестом API.',
  publishedAt: '2026-09-01T09:00:00Z',
  price: null,
  paid: false,
  program: [
    {
      title: 'Межі модулів',
      lessons: [{ title: 'Порти' }, { title: 'Адаптери' }],
    },
    { title: 'Перевірка', lessons: [{ title: 'Контрактний тест' }] },
  ],
};

const initialQuery: CatalogQuery = { q: '', sort: 'recent', limit: 20, offset: 0 };
const firstPage: CoursePage = {
  items: [architectureCourse],
  pagination: { limit: 20, offset: 0, total: 1 },
};

describe('catalog presentation', () => {
  it('offers published course links with meaningful names and its summary', () => {
    render(<CatalogView page={firstPage} query={initialQuery} />);

    expect(screen.getByRole('heading', { level: 1, name: 'Каталог курсів' })).toBeTruthy();
    expect(
      screen.getByRole('link', { name: 'Архітектура застосунків Go' }).getAttribute('href'),
    ).toBe('/courses/go-architecture');
    expect(
      screen
        .getByRole('link', { name: 'Відкрити: Архітектура застосунків Go' })
        .getAttribute('href'),
    ).toBe('/courses/go-architecture');
    expect(screen.getByText('Побудуйте API з чіткими межами модулів.')).toBeTruthy();
    // One page of results needs no pagination; the one counter sits under the heading.
    expect(screen.queryByRole('navigation', { name: 'Сторінки каталогу' })).toBeNull();
    expect(screen.getByText('1 курс')).toBeTruthy();
    // No lessons yet: the card says so instead of "0 lessons" and a "new course" placeholder.
    expect(screen.getByText('Скоро буде')).toBeTruthy();
    expect(screen.queryByText('0 уроків')).toBeNull();
    expect(screen.queryByText('Новий курс')).toBeNull();
    expect(screen.queryByText('Практичний курс')).toBeNull();
  });

  it('shows rating and lesson count once a course has lessons', () => {
    render(
      <CatalogView
        page={{
          items: [{ ...architectureCourse, lessonCount: 3, rating: null }],
          pagination: { limit: 20, offset: 0, total: 1 },
        }}
        query={initialQuery}
      />,
    );
    expect(screen.getByText('3 уроки')).toBeTruthy();
    expect(screen.getByText('Новий курс')).toBeTruthy();
    expect(screen.queryByText('Скоро буде')).toBeNull();
  });

  it('applies a new sort at once with the text typed so far', async () => {
    const user = userEvent.setup();
    const changes: unknown[] = [];
    render(
      <CatalogView
        page={firstPage}
        query={{ ...initialQuery, q: 'Go' }}
        onSortChange={(next) => changes.push(next)}
      />,
    );
    await user.clear(screen.getByRole('textbox', { name: 'Знайти курс' }));
    await user.type(screen.getByRole('textbox', { name: 'Знайти курс' }), ' React ');
    await user.selectOptions(screen.getByRole('combobox', { name: 'Сортування' }), 'title');
    expect(changes).toEqual([{ q: 'React', sort: 'title' }]);
  });

  it('formats catalog prices and keeps the visual purchase action on course detail', () => {
    const paidCourse: CourseDetail = {
      ...architectureCourse,
      id: '7b823b02-65f6-4875-8e66-65e44fd08330',
      slug: 'paid-architecture',
      title: 'Платна архітектура',
      price: { minor: 149900, currency: 'UAH' },
      paid: true,
    };
    render(
      <CatalogView
        page={{ items: [paidCourse], pagination: { limit: 20, offset: 0, total: 1 } }}
        query={initialQuery}
      />,
    );

    // Hryvnia sign, as in the mockup — not the "грн" abbreviation.
    expect(screen.getByText(/1\s?499\s?₴/u)).toBeTruthy();
    // The catalog carries the same truthful value strip as the home page.
    expect(screen.getByRole('heading', { name: 'Що дає навчання' })).toBeTruthy();
    expect(screen.getByRole('heading', { name: 'Транскрипти' })).toBeTruthy();
    const action = screen.getByText('Купити').closest('a');
    expect(action?.getAttribute('href')).toBe('/courses/paid-architecture');
    expect(action?.getAttribute('aria-label')).toBe('Купити: Платна архітектура');
  });

  it('shows a complete detail description and a course learning entry', () => {
    render(<CourseView course={architectureCourse} action={{ kind: 'enroll' }} />);

    expect(
      screen.getByRole('heading', { level: 1, name: 'Архітектура застосунків Go' }),
    ).toBeTruthy();
    expect(
      screen.getByText('Працюємо з портами та адаптерами. Завершуємо наскрізним тестом API.'),
    ).toBeTruthy();
    expect(screen.getByRole('heading', { name: 'Готові почати?' })).toBeTruthy();
    expect(screen.getByRole('button', { name: 'Записатися' })).toBeTruthy();
    expect(screen.getByRole('link', { name: '← Усі курси' }).getAttribute('href')).toBe('/courses');
    // No constant label above the title (C3 #7); the price is in the action box (C3 #4).
    expect(screen.queryByText('Практичний курс')).toBeNull();
    expect(within(screen.getByRole('complementary')).getByText('Безкоштовно')).toBeTruthy();

    const programHeading = screen.getByRole('heading', { name: 'Програма курсу' });
    const program = programHeading.closest('section');
    if (!program) throw new Error('The program heading must label a section');
    expect(program.textContent).toContain('Межі модулівПортиАдаптериПеревіркаКонтрактний тест');
    expect(within(program).getAllByRole('list')).toHaveLength(3);
    expect(within(program).queryByRole('link')).toBeNull();
  });

  it('puts the action box before the course text in reading order (C3 #5)', () => {
    render(<CourseView course={architectureCourse} action={{ kind: 'enroll' }} />);
    const box = screen.getByRole('complementary');
    const about = screen.getByRole('heading', { name: 'Про курс' });
    expect(box.compareDocumentPosition(about) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
  });

  it('renders the enrol state and the member states', async () => {
    const user = userEvent.setup();
    let enrolled = false;
    const { rerender } = render(
      <CourseView
        course={architectureCourse}
        action={{ kind: 'enroll' }}
        onEnroll={() => {
          enrolled = true;
        }}
      />,
    );
    await user.click(screen.getByRole('button', { name: 'Записатися' }));
    expect(enrolled).toBe(true);

    rerender(
      <CourseView
        course={architectureCourse}
        action={{ kind: 'member', blocked: false, comingSoon: false }}
      />,
    );
    expect(screen.getByRole('heading', { name: 'Ви учасник' })).toBeTruthy();
    // Still a link (it navigates), drawn as the main button; the complaint is a quiet text
    // button below it (C3 #3).
    const start = screen.getByRole('link', { name: 'Почати навчання' });
    expect(start.getAttribute('href')).toBe('/learn/go-architecture');
    expect(start.className).toContain('ui-button--primary');
    expect(screen.getByRole('button', { name: 'Поскаржитися на викладача' }).className).toContain(
      'ui-button--quiet',
    );

    rerender(
      <CourseView
        course={architectureCourse}
        action={{ kind: 'member', blocked: true, comingSoon: false }}
      />,
    );
    expect(screen.getByRole('status').textContent).toContain(
      'Доступ до контенту відкрито, участь в обговореннях обмежено',
    );

    rerender(
      <CourseView
        course={architectureCourse}
        action={{ kind: 'enroll' }}
        enrollError="membership_conflict"
      />,
    );
    expect(screen.getByRole('button', { name: 'Записатися' })).toBeTruthy();
    expect(screen.getByRole('alert').textContent).toContain('Стан участі змінився');
  });

  it('lets current members report the instructor', async () => {
    const report = async (reason: string) => {
      expect(reason).toBe('Неприйнятна поведінка');
    };
    const { rerender } = render(
      <CourseView
        course={architectureCourse}
        action={{ kind: 'member', blocked: false, comingSoon: false }}
        onReportLecturer={report}
      />,
    );
    await userEvent.click(screen.getByRole('button', { name: 'Поскаржитися на викладача' }));
    await userEvent.type(
      screen.getByRole('textbox', { name: 'Причина скарги' }),
      'Неприйнятна поведінка',
    );
    await userEvent.click(screen.getByRole('button', { name: 'Надіслати скаргу' }));
    expect((await screen.findByRole('status')).textContent).toContain('Скаргу надіслано');

    rerender(<CourseView course={architectureCourse} action={{ kind: 'enroll' }} />);
    expect(screen.queryByRole('button', { name: 'Поскаржитися на викладача' })).toBeNull();
  });

  it('sends a guest to sign in and back to this course, never to a write (C3 #1)', () => {
    const { rerender } = render(
      <CourseView course={architectureCourse} action={{ kind: 'sign-in', paid: false }} />,
    );
    const signIn = screen.getByRole('link', { name: 'Увійти, щоб записатися' });
    expect(signIn.getAttribute('href')).toBe('/login?returnTo=%2Fcourses%2Fgo-architecture');
    expect(screen.queryByRole('button', { name: 'Записатися' })).toBeNull();

    rerender(
      <CourseView
        course={{ ...architectureCourse, price: { minor: 189000, currency: 'UAH' } }}
        action={{ kind: 'sign-in', paid: true }}
      />,
    );
    expect(screen.getByRole('link', { name: 'Увійти, щоб купити' })).toBeTruthy();
    expect(within(screen.getByRole('complementary')).getByText(/1\s?890\s?₴/u)).toBeTruthy();
  });

  it('offers to buy a paid course and names a refusal (C3 #4)', async () => {
    let bought = 0;
    const paid = {
      ...architectureCourse,
      price: { minor: 189000, currency: 'UAH' as const },
    };
    const { rerender } = render(
      <CourseView course={paid} action={{ kind: 'buy' }} onBuy={() => (bought += 1)} />,
    );
    await userEvent.click(screen.getByRole('button', { name: 'Купити курс' }));
    expect(bought).toBe(1);

    rerender(<CourseView course={paid} action={{ kind: 'buy' }} buyError="course_already_owned" />);
    expect(screen.getByRole('alert').textContent).toBe('Цей курс зараз не можна купити.');

    // A lost session is handled by re-reading it, not by a message that asks to retry.
    rerender(
      <CourseView course={paid} action={{ kind: 'buy' }} buyError="authentication_required" />,
    );
    expect(screen.queryByRole('alert')).toBeNull();
  });

  it('shows only «Скоро буде» for a course without lessons (C3 #2)', () => {
    const empty = { ...architectureCourse, program: [], lessonCount: 0 };
    const { rerender } = render(<CourseView course={empty} action={{ kind: 'coming-soon' }} />);
    const box = screen.getByRole('complementary');
    expect(within(box).getByRole('heading', { name: 'Скоро буде' })).toBeTruthy();
    expect(within(box).queryByRole('button')).toBeNull();
    expect(within(box).queryByRole('link')).toBeNull();
    // The header line matches the catalog card.
    expect(screen.getAllByText('Скоро буде')).toHaveLength(2);
    expect(screen.queryByText('Новий курс')).toBeNull();

    rerender(
      <CourseView course={empty} action={{ kind: 'member', blocked: false, comingSoon: true }} />,
    );
    expect(screen.getByRole('heading', { name: 'Ви учасник' })).toBeTruthy();
    expect(screen.queryByRole('link', { name: 'Почати навчання' })).toBeNull();
    expect(screen.getByRole('button', { name: 'Поскаржитися на викладача' })).toBeTruthy();
  });

  it('shows the author their course with a way to edit it (owner decision)', () => {
    render(
      <CourseView course={architectureCourse} action={{ kind: 'owner', comingSoon: false }} />,
    );
    expect(screen.getByRole('heading', { name: 'Це ваш курс' })).toBeTruthy();
    expect(screen.getByRole('link', { name: 'Редагувати курс' }).getAttribute('href')).toBe(
      `/teach/courses/${architectureCourse.id}`,
    );
    expect(screen.queryByRole('button', { name: 'Записатися' })).toBeNull();
    expect(screen.queryByRole('button', { name: 'Поскаржитися на викладача' })).toBeNull();
  });

  it('keeps the administrator exemption (G8)', () => {
    render(
      <CourseView course={architectureCourse} action={{ kind: 'admin', comingSoon: false }} />,
    );
    expect(screen.getByRole('link', { name: 'Почати навчання' }).getAttribute('href')).toBe(
      '/learn/go-architecture',
    );
    expect(screen.queryByRole('button', { name: 'Записатися' })).toBeNull();
  });

  it('holds every action while access is checked, and offers a retry when it fails', async () => {
    let retried = 0;
    const { rerender } = render(
      <CourseView course={architectureCourse} action={{ kind: 'checking' }} />,
    );
    const box = screen.getByRole('complementary');
    expect(within(box).getByRole('status').textContent).toBe('Перевіряємо доступ…');
    expect(within(box).queryByRole('button')).toBeNull();

    rerender(
      <CourseView
        course={architectureCourse}
        action={{ kind: 'unavailable' }}
        onRetryAccess={() => (retried += 1)}
      />,
    );
    expect(screen.getByRole('alert').textContent).toContain('Не вдалося перевірити доступ');
    await userEvent.click(screen.getByRole('button', { name: 'Спробувати ще раз' }));
    expect(retried).toBe(1);
    // The course itself stays readable.
    expect(screen.getByRole('heading', { name: 'Про курс' })).toBeTruthy();
  });

  it('omits an empty program section and localizes a present program heading', () => {
    const { rerender } = render(<CourseView course={{ ...architectureCourse, program: [] }} />);
    expect(screen.queryByRole('heading', { name: 'Програма курсу' })).toBeNull();

    rerender(
      <I18nProvider instance={createI18n('en')}>
        <CourseView course={architectureCourse} />
      </I18nProvider>,
    );
    expect(screen.getByRole('heading', { name: 'Course program' })).toBeTruthy();
  });

  it('submits edited search and sort with the chosen limit and resets the old offset', async () => {
    const user = userEvent.setup();
    const query: CatalogQuery = { q: 'Go', sort: 'title', limit: 10, offset: 40 };
    const page: CoursePage = {
      items: [architectureCourse],
      pagination: { limit: 10, offset: 40, total: 60 },
    };
    render(<CatalogView page={page} query={query} />);
    const form = screen.getByRole('textbox', { name: 'Знайти курс' }).closest('form');
    if (!(form instanceof HTMLFormElement)) throw new Error('Search must be a semantic HTML form');

    expect(form.getAttribute('method')).toBe('get');
    expect(form.getAttribute('action')).toBe('/courses');
    expect(new FormData(form).get('q')).toBe('Go');
    expect(new FormData(form).get('sort')).toBe('title');
    expect(new FormData(form).get('limit')).toBe('10');

    let submitted: FormData | undefined;
    form.addEventListener('submit', (event) => {
      event.preventDefault();
      submitted = new FormData(form);
    });
    await user.clear(screen.getByRole('textbox', { name: 'Знайти курс' }));
    await user.type(screen.getByRole('textbox', { name: 'Знайти курс' }), 'TypeScript');
    await user.selectOptions(screen.getByRole('combobox', { name: 'Сортування' }), 'recent');
    await user.click(screen.getByRole('button', { name: 'Знайти' }));

    expect(submitted?.get('q')).toBe('TypeScript');
    expect(submitted?.get('sort')).toBe('recent');
    expect(submitted?.get('limit')).toBe('10');
    expect(submitted?.has('offset')).toBe(false);
  });

  it('keeps search, sort and limit in next-page navigation', () => {
    const query: CatalogQuery = { q: 'Go & API', sort: 'title', limit: 10, offset: 20 };
    render(
      <CatalogView
        query={query}
        page={{ items: [architectureCourse], pagination: { limit: 10, offset: 20, total: 45 } }}
      />,
    );
    const href = screen.getByRole('link', { name: 'Наступна →' }).getAttribute('href');
    if (!href) throw new Error('Next page must have a URL');
    const url = new URL(href, 'https://courses.test');

    expect(url.pathname).toBe('/courses');
    expect(Object.fromEntries(url.searchParams)).toEqual({
      q: 'Go & API',
      sort: 'title',
      limit: '10',
      offset: '30',
    });
  });

  it('shows a real empty catalog with one action that clears the search and no pages', () => {
    render(
      <CatalogView
        query={initialQuery}
        page={{ items: [], pagination: { limit: 20, offset: 0, total: 0 } }}
      />,
    );

    expect(screen.getByRole('heading', { name: 'Курсів не знайдено' })).toBeTruthy();
    expect(screen.queryByRole('list')).toBeNull();
    const reset = screen.getByRole('link', { name: 'Скинути пошук' });
    expect(reset.getAttribute('href')).toBe('/courses');
    expect(reset.className).toContain('ui-button--primary');
    expect(screen.queryByRole('navigation', { name: 'Сторінки каталогу' })).toBeNull();
  });

  it('leads from beyond the end to the last page with courses, with no page links', () => {
    const query: CatalogQuery = { q: 'Go', sort: 'title', limit: 10, offset: 40 };
    render(
      <CatalogView
        query={query}
        page={{ items: [], pagination: { limit: 10, offset: 40, total: 3 } }}
      />,
    );

    expect(screen.getByRole('heading', { name: 'На цій сторінці немає курсів' })).toBeTruthy();
    expect(screen.queryByRole('navigation', { name: 'Сторінки каталогу' })).toBeNull();
    const href = screen.getByRole('link', { name: 'До останньої сторінки' }).getAttribute('href');
    const url = new URL(href ?? '', 'https://courses.test');
    expect(Object.fromEntries(url.searchParams)).toEqual({
      q: 'Go',
      sort: 'title',
      limit: '10',
      offset: '0',
    });
    expect(screen.queryByRole('link', { name: '← Попередня' })).toBeNull();
  });

  it('computes the last page with courses, not the first, for several pages', () => {
    // 10 per page: 25 courses fill pages at 0, 10 and 20, so the last is 20; exactly 20
    // courses fill pages at 0 and 10, so the last is 10.
    for (const [total, offset] of [
      [25, '20'],
      [20, '10'],
    ] as const) {
      const query: CatalogQuery = { q: '', sort: 'recent', limit: 10, offset: 40 };
      const { unmount } = render(
        <CatalogView
          query={query}
          page={{ items: [], pagination: { limit: 10, offset: 40, total } }}
        />,
      );
      const href = screen.getByRole('link', { name: 'До останньої сторінки' }).getAttribute('href');
      expect(new URL(href ?? '', 'https://courses.test').searchParams.get('offset')).toBe(offset);
      unmount();
    }
  });

  it('announces loading and makes a missing course recoverable', () => {
    render(
      <>
        <CourseLoading />
        <CourseNotFound />
      </>,
    );

    // The course page's own words, not the catalog's (C3b R5).
    expect(screen.getByRole('status').textContent).toBe('Завантажуємо курс');
    // The same page as the site 404: a level-one heading and a button back (C3 #8).
    expect(screen.getByRole('heading', { level: 1, name: 'Курс не знайдено' })).toBeTruthy();
    const back = screen.getByRole('link', { name: 'Перейти до каталогу' });
    expect(back.getAttribute('href')).toBe('/courses');
    expect(back.className).toContain('ui-button--primary');
  });

  it('brings the search controls back in line when the address changes under them', async () => {
    const user = userEvent.setup();
    const { rerender } = render(
      <CatalogView page={firstPage} query={{ ...initialQuery, q: 'React', sort: 'title' }} />,
    );
    await user.clear(screen.getByRole('textbox', { name: 'Знайти курс' }));
    await user.type(screen.getByRole('textbox', { name: 'Знайти курс' }), 'React');
    // Back: the same mounted view now renders an older address.
    rerender(<CatalogView page={firstPage} query={{ ...initialQuery, q: 'Go', sort: 'recent' }} />);
    expect((screen.getByRole('textbox', { name: 'Знайти курс' }) as HTMLInputElement).value).toBe(
      'Go',
    );
    expect((screen.getByRole('combobox', { name: 'Сортування' }) as HTMLSelectElement).value).toBe(
      'recent',
    );
    // The field's own state follows too: hovering and focusing it re-renders the input, which
    // must not bring the earlier text back (Codex review C2 R2).
    await user.hover(screen.getByRole('textbox', { name: 'Знайти курс' }));
    await user.click(screen.getByRole('textbox', { name: 'Знайти курс' }));
    rerender(<CatalogView page={firstPage} query={{ ...initialQuery, q: 'Go', sort: 'recent' }} />);
    expect((screen.getByRole('textbox', { name: 'Знайти курс' }) as HTMLInputElement).value).toBe(
      'Go',
    );
  });

  it('keeps the heading and the search while the list loads; only the form is inert', () => {
    render(<CatalogListLoading query={{ ...initialQuery, q: 'Go' }} />);
    expect(screen.getByRole('heading', { level: 1, name: 'Каталог курсів' })).toBeTruthy();
    const field = screen.getByRole('textbox', { name: 'Знайти курс' }) as HTMLInputElement;
    expect(field.value).toBe('Go');
    // The pending frame is replaced when the list arrives, so its form takes no focus or text
    // (C3b R4a); the heading and the loading status are outside the inert form.
    const form = field.closest('form');
    expect(form?.hasAttribute('inert')).toBe(true);
    const status = screen.getByRole('status');
    expect(status.textContent).toBe('Завантажуємо курси');
    expect(status.closest('[inert]')).toBeNull();
    expect(screen.getByRole('heading', { level: 1 }).closest('[inert]')).toBeNull();
  });

  it('keeps the loaded catalog form interactive', () => {
    render(<CatalogView page={firstPage} query={initialQuery} />);
    expect(screen.getByRole('textbox', { name: 'Знайти курс' }).closest('[inert]')).toBeNull();
  });

  it('asks for a full document load from a failed course page retry', async () => {
    const root = createRootRoute({ component: Outlet });
    const course = createRoute({
      getParentRoute: () => root,
      path: '/courses/$slug',
      component: CourseFailure,
    });
    const router = createRouter({
      routeTree: root.addChildren([course]),
      history: createMemoryHistory({ initialEntries: ['/courses/go-basics'] }),
    });
    const navigated: unknown[] = [];
    await router.load();
    render(
      <AriaRouterProvider navigate={(href, options) => navigated.push([href, options])}>
        <RouterProvider router={router} />
      </AriaRouterProvider>,
    );
    const error = await screen.findByRole('alert');
    expect(
      within(error).getByRole('heading', { name: 'Не вдалося завантажити курс' }),
    ).toBeTruthy();
    await userEvent.click(within(error).getByRole('link', { name: 'Спробувати ще раз' }));
    expect(navigated).toEqual([['/courses/go-basics', { reloadDocument: true }]]);
  });

  it('shows a failed list inside the catalog frame', async () => {
    const root = createRootRoute({ component: Outlet });
    const courses = createRoute({
      getParentRoute: () => root,
      path: '/courses',
      component: () => <CatalogListFailure query={{ ...initialQuery, q: 'Go' }} />,
    });
    const router = createRouter({
      routeTree: root.addChildren([courses]),
      history: createMemoryHistory({ initialEntries: ['/courses?q=Go'] }),
    });
    await router.load();
    render(<RouterProvider router={router} />);
    const error = await screen.findByRole('alert');
    expect(within(error).getByRole('link', { name: 'Спробувати ще раз' })).toBeTruthy();
    expect(screen.getByRole('heading', { level: 1, name: 'Каталог курсів' })).toBeTruthy();
    expect(screen.getByRole('textbox', { name: 'Знайти курс' })).toBeTruthy();
  });

  it('keeps the entire router URL in the catalog error recovery link', async () => {
    const root = createRootRoute({ component: Outlet });
    const courses = createRoute({
      getParentRoute: () => root,
      path: '/courses',
      component: CatalogFailure,
    });
    const retryHref = '/courses?q=Go&sort=title&limit=10&offset=20';
    const router = createRouter({
      routeTree: root.addChildren([courses]),
      history: createMemoryHistory({ initialEntries: [retryHref] }),
    });
    await router.load();
    render(<RouterProvider router={router} />);

    const error = await screen.findByRole('alert');
    expect(
      within(error).getByRole('heading', { name: 'Не вдалося завантажити курси' }),
    ).toBeTruthy();
    expect(
      within(error).getByRole('link', { name: 'Спробувати ще раз' }).getAttribute('href'),
    ).toBe(retryHref);
  });
});
