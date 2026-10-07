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
import { createI18n, I18nProvider } from '../../app/providers/i18n';
import type { CatalogQuery, CourseDetail, CoursePage } from '../../domain/course/course';
import { CatalogFailure, CatalogLoading, CourseNotFound } from './CatalogStates';
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
        .getByRole('link', { name: 'Детальніше: Архітектура застосунків Go' })
        .getAttribute('href'),
    ).toBe('/courses/go-architecture');
    expect(screen.getByText('Побудуйте API з чіткими межами модулів.')).toBeTruthy();
    expect(screen.getByRole('navigation', { name: 'Сторінки каталогу' }).textContent).toContain(
      '1–1 із 1',
    );
    expect(screen.queryByRole('link', { name: 'Наступна →' })).toBeNull();
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
    expect(action?.getAttribute('aria-label')).toBe('Детальніше: Платна архітектура');
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

    rerender(
      <CourseView course={architectureCourse} membership={{ state: 'active' }} />,
    );
    expect(screen.getByRole('heading', { name: 'Ви учасник' })).toBeTruthy();
    expect(screen.getByRole('link', { name: 'Почати навчання' }).getAttribute('href')).toBe(
      '/learn/go-architecture',
    );

    rerender(
      <CourseView course={architectureCourse} membership={{ state: 'blocked' }} />,
    );
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
    await userEvent.type(screen.getByRole('textbox', { name: 'Причина скарги' }), 'Неприйнятна поведінка');
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

  it('shows a real empty catalog and an action to return to its first page', () => {
    render(
      <CatalogView
        query={initialQuery}
        page={{ items: [], pagination: { limit: 20, offset: 0, total: 0 } }}
      />,
    );

    expect(screen.getByRole('heading', { name: 'Курсів не знайдено' })).toBeTruthy();
    expect(screen.queryByRole('list')).toBeNull();
    expect(screen.getByRole('link', { name: 'На першу сторінку' }).getAttribute('href')).toBe(
      '/courses',
    );
    expect(screen.getByRole('navigation', { name: 'Сторінки каталогу' }).textContent).toContain(
      '0 із 0',
    );
  });

  it('retains the result total beyond the end and offers first-page recovery', () => {
    const query: CatalogQuery = { q: 'Go', sort: 'title', limit: 10, offset: 40 };
    render(
      <CatalogView
        query={query}
        page={{ items: [], pagination: { limit: 10, offset: 40, total: 3 } }}
      />,
    );

    expect(screen.getByRole('heading', { name: 'На цій сторінці немає курсів' })).toBeTruthy();
    expect(screen.getByRole('navigation', { name: 'Сторінки каталогу' }).textContent).toContain(
      '0 із 3',
    );
    expect(screen.getByRole('link', { name: 'На першу сторінку' }).getAttribute('href')).toBe(
      '/courses',
    );
    expect(screen.queryByRole('link', { name: 'Наступна →' })).toBeNull();
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
