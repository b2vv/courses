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
import type { CatalogQuery, CoursePage, PublishedCourse } from '../../domain/course/course';
import { CatalogFailure, CatalogLoading, CourseNotFound } from './CatalogStates';
import { CatalogView } from './CatalogView';
import { CourseView } from './CourseView';

const architectureCourse: PublishedCourse = {
  id: '7b823b02-65f6-4875-8e66-65e44fd08329',
  slug: 'go-architecture',
  title: 'Архітектура застосунків Go',
  summary: 'Побудуйте API з чіткими межами модулів.',
  description: 'Працюємо з портами та адаптерами.\nЗавершуємо наскрізним тестом API.',
  publishedAt: '2026-09-01T09:00:00Z',
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

  it('shows a complete detail description and an honest availability message', () => {
    render(<CourseView course={architectureCourse} />);

    expect(
      screen.getByRole('heading', { level: 1, name: 'Архітектура застосунків Go' }),
    ).toBeTruthy();
    expect(
      screen.getByText('Працюємо з портами та адаптерами. Завершуємо наскрізним тестом API.'),
    ).toBeTruthy();
    expect(screen.getByRole('heading', { name: 'Навчання незабаром' })).toBeTruthy();
    expect(screen.getByRole('link', { name: '← Усі курси' }).getAttribute('href')).toBe('/courses');
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
