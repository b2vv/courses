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
import { createI18n, I18nProvider } from '../../../app/providers/i18n';
import type { CatalogQuery, CourseDetail, CoursePage } from '../../../contracts/course';
import {
  CatalogFailure,
  CatalogListFailure,
  CatalogListLoading,
  CatalogLoading,
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
    render(<CourseView course={architectureCourse} />);

    expect(
      screen.getByRole('heading', { level: 1, name: 'Архітектура застосунків Go' }),
    ).toBeTruthy();
    expect(
      screen.getByText('Працюємо з портами та адаптерами. Завершуємо наскрізним тестом API.'),
    ).toBeTruthy();
    expect(screen.getByRole('heading', { name: 'Готові почати?' })).toBeTruthy();
    expect(screen.getByRole('button', { name: 'Записатися' })).toBeTruthy();
    expect(screen.getByRole('link', { name: '← Усі курси' }).getAttribute('href')).toBe('/courses');

    const programHeading = screen.getByRole('heading', { name: 'Програма курсу' });
    const program = programHeading.closest('section');
    if (!program) throw new Error('The program heading must label a section');
    expect(program.textContent).toContain('Межі модулівПортиАдаптериПеревіркаКонтрактний тест');
    expect(within(program).getAllByRole('list')).toHaveLength(3);
    expect(within(program).queryByRole('link')).toBeNull();
  });

  it('renders enrollment CTA states and switches after a successful enrollment', async () => {
    const user = userEvent.setup();
    let enrolled = false;
    const { rerender } = render(
      <CourseView
        course={architectureCourse}
        onEnroll={() => {
          enrolled = true;
        }}
      />,
    );
    await user.click(screen.getByRole('button', { name: 'Записатися' }));
    expect(enrolled).toBe(true);

    rerender(<CourseView course={architectureCourse} membership={{ state: 'active' }} />);
    expect(screen.getByRole('heading', { name: 'Ви учасник' })).toBeTruthy();
    expect(screen.getByRole('link', { name: 'Почати навчання' }).getAttribute('href')).toBe(
      '/learn/go-architecture',
    );

    rerender(<CourseView course={architectureCourse} membership={{ state: 'blocked' }} />);
    expect(screen.getByRole('status').textContent).toContain(
      'Доступ до контенту відкрито, участь в обговореннях обмежено',
    );

    rerender(
      <CourseView
        course={architectureCourse}
        membership={{ state: 'revoked' }}
        enrollError="membership_conflict"
      />,
    );
    expect(screen.getByRole('button', { name: 'Записатися' })).toBeTruthy();
    expect(screen.getByRole('alert').textContent).toContain('Стан участі змінився');
  });

  it('lets current members report the instructor and keeps the action from revoked viewers', async () => {
    const report = async (reason: string) => {
      expect(reason).toBe('Неприйнятна поведінка');
    };
    const { rerender } = render(
      <CourseView
        course={architectureCourse}
        membership={{ state: 'active' }}
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

    rerender(<CourseView course={architectureCourse} membership={{ state: 'revoked' }} />);
    expect(screen.queryByRole('button', { name: 'Поскаржитися на викладача' })).toBeNull();
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
        <CatalogLoading />
        <CourseNotFound />
      </>,
    );

    expect(screen.getByRole('status').textContent).toBe('Завантажуємо курси');
    expect(screen.getByRole('heading', { name: 'Курс не знайдено' })).toBeTruthy();
    expect(screen.getByRole('link', { name: 'Перейти до каталогу' }).getAttribute('href')).toBe(
      '/courses',
    );
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

  it('keeps the heading and the editable search while the list loads', () => {
    render(<CatalogListLoading query={{ ...initialQuery, q: 'Go' }} />);
    expect(screen.getByRole('heading', { level: 1, name: 'Каталог курсів' })).toBeTruthy();
    expect((screen.getByRole('textbox', { name: 'Знайти курс' }) as HTMLInputElement).value).toBe(
      'Go',
    );
    expect(screen.getByRole('status').textContent).toBe('Завантажуємо курси');
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
