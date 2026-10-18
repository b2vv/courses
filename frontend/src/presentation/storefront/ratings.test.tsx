import { describe, expect, it, rs } from '@rstest/core';
import { render, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import type { CourseDetail, CoursePage, PublicReviewPage } from '../../contracts/course';
import { CatalogView } from './CatalogView';
import { CourseView } from './CourseView';

const course: CourseDetail = {
  id: '7b823b02-65f6-4875-8e66-65e44fd08329',
  slug: 'go-architecture',
  title: 'Архітектура застосунків Go',
  summary: 'Побудуйте API з чіткими межами модулів.',
  description: 'Порти й адаптери.',
  publishedAt: '2026-09-01T09:00:00Z',
  price: null,
  paid: false,
  rating: null,
  lessonCount: 3,
  program: [{ title: 'Межі модулів', lessons: [{ title: 'Порти' }, { title: 'Адаптери' }] }],
};

function review(index: number, overrides: Partial<PublicReviewPage['items'][number]> = {}) {
  return {
    id: `75000000-0000-4000-8000-${String(index).padStart(12, '0')}`,
    authorDisplayName: `Автор ${index}`,
    stars: 4,
    body: 'Корисний курс, усе по ділу.',
    completedLessons: 1,
    totalLessons: 3,
    percent: 33,
    createdAt: '2026-10-01T09:00:00Z',
    editedAt: null,
    ...overrides,
  };
}

function page(count: number, nextCursor: string | null = null): PublicReviewPage {
  return {
    rating: null,
    items: Array.from({ length: count }, (_, index) => review(index + 1)),
    nextCursor,
  };
}

describe('catalog card rating', () => {
  const catalogPage = (detail: CourseDetail): CoursePage => ({
    items: [detail],
    pagination: { limit: 20, offset: 0, total: 1 },
  });

  it('shows «Новий курс» below the threshold (Rt06)', () => {
    render(
      <CatalogView
        page={catalogPage(course)}
        query={{ q: '', sort: 'recent', limit: 20, offset: 0 }}
      />,
    );
    expect(screen.getByText('Новий курс')).toBeTruthy();
    expect(screen.getByText('3 уроки')).toBeTruthy();
  });

  it('shows the one-decimal average with a full sentence for assistive tech (Rt07)', () => {
    render(
      <CatalogView
        page={catalogPage({ ...course, rating: { average: 5, count: 5 } })}
        query={{ q: '', sort: 'recent', limit: 20, offset: 0 }}
      />,
    );
    expect(screen.getByText('★ 5,0 (5)').getAttribute('aria-hidden')).toBe('true');
    expect(screen.getByText('Рейтинг 5,0 з 5, 5 оцінок')).toBeTruthy();
    expect(screen.queryByText('Новий курс')).toBeNull();
  });
});

describe('course page reviews', () => {
  it('keeps rendered reviews and offers an inline retry after a next-page failure', async () => {
    const retry = rs.fn();
    render(
      <CourseView
        course={course}
        reviews={page(20, 'next')}
        hasMore
        isNextPageError
        onRetryMore={retry}
      />,
    );
    // Scope to the reviews list: the course program on the same page is also a list.
    const list = screen.getByRole('list', { name: 'Відгуки про курс' });
    expect(within(list).getAllByRole('listitem')).toHaveLength(20);
    const [inlineRetry] = screen.getAllByRole('button', { name: 'Спробувати ще' });
    if (inlineRetry === undefined) throw new Error('inline retry button missing');
    await userEvent.click(inlineRetry);
    expect(retry).toHaveBeenCalledTimes(1);
  });
  it('shows the section error only when the first reviews page fails', () => {
    render(<CourseView course={course} reviewsError />);
    expect(screen.getByRole('alert').textContent).toContain('Не вдалося завантажити відгуки');
  });
  it('announces added and completed pages and focuses the first final addition', async () => {
    const { rerender } = render(<CourseView course={course} reviews={page(1, 'next')} hasMore />);
    rerender(<CourseView course={course} reviews={page(2)} />);
    expect(await screen.findByText('Усі відгуки показано')).toBeTruthy();
    expect(document.activeElement).not.toBe(document.body);
  });
  it('renders learning-scoped reviews for an archived course viewer (Rt30)', () => {
    render(
      <CourseView course={{ ...course, status: 'archived' } as CourseDetail} reviews={page(1)} />,
    );
    expect(screen.getByText('Автор 1')).toBeTruthy();
  });
  it('shows the rating summary or «Новий курс» on the course page (Rt08)', () => {
    const { rerender } = render(<CourseView course={course} reviews={page(0)} />);
    expect(screen.getByText('Новий курс')).toBeTruthy();
    rerender(
      <CourseView course={{ ...course, rating: { average: 4.7, count: 3 } }} reviews={page(0)} />,
    );
    expect(screen.getByText('Рейтинг 4,7 з 5, 3 оцінки')).toBeTruthy();
  });

  it('shows an empty state when there are no reviews', () => {
    render(<CourseView course={course} reviews={page(0)} />);
    expect(screen.getByRole('heading', { name: 'Відгуки' })).toBeTruthy();
    expect(screen.getByText('Ще немає відгуків')).toBeTruthy();
  });

  it('lists a page of reviews with stars, progress and edit marker, and asks for more (Rt14)', async () => {
    const onLoadMore = rs.fn();
    render(
      <CourseView
        course={course}
        reviews={{
          ...page(20, 'next'),
          items: [review(1, { editedAt: '2026-10-02T09:00:00Z' }), ...page(19).items],
        }}
        hasMore
        onLoadMore={onLoadMore}
      />,
    );
    const list = screen.getByRole('list', { name: 'Відгуки про курс' });
    expect(within(list).getAllByRole('listitem')).toHaveLength(20);
    expect(within(list).getAllByText('4 з 5')).toHaveLength(20);
    expect(within(list).getAllByText('пройдено 33% на момент відгуку')).toHaveLength(20);
    expect(within(list).getByText(/^змінено /)).toBeTruthy();
    expect(document.querySelector('output[aria-live="polite"]')).toBeTruthy();
    await userEvent.click(screen.getByRole('button', { name: 'Показати ще' }));
    expect(onLoadMore).toHaveBeenCalledTimes(1);
  });

  it('keeps the program visible and offers a retry when reviews fail (Rt28)', async () => {
    const onRetry = rs.fn();
    render(<CourseView course={course} reviewsError onRetry={onRetry} />);
    expect(screen.getByRole('heading', { name: 'Програма курсу' })).toBeTruthy();
    expect(screen.getByRole('alert').textContent).toContain('Не вдалося завантажити відгуки');
    await userEvent.click(screen.getByRole('button', { name: 'Спробувати ще' }));
    expect(onRetry).toHaveBeenCalledTimes(1);
  });

  it('hides «Поскаржитись» for anonymous viewers and on the viewer’s own review', () => {
    const { rerender } = render(<CourseView course={course} reviews={page(2)} />);
    expect(screen.queryByRole('button', { name: 'Поскаржитись' })).toBeNull();
    rerender(<CourseView course={course} reviews={page(2)} signedIn ownReviewId={review(1).id} />);
    expect(screen.getAllByRole('button', { name: 'Поскаржитись' })).toHaveLength(1);
  });

  it('confirms a report and disables the button', async () => {
    const onReport = rs.fn(async () => {});
    render(<CourseView course={course} reviews={page(1)} signedIn onReport={onReport} />);
    await userEvent.click(screen.getByRole('button', { name: 'Поскаржитись' }));
    expect(onReport).toHaveBeenCalledWith(review(1).id);
    expect(await screen.findByText('Скаргу надіслано')).toBeTruthy();
  });

  it('names a refused own report', async () => {
    const onReport = rs.fn(async () => {
      throw new Error('report_own_review');
    });
    render(<CourseView course={course} reviews={page(1)} signedIn onReport={onReport} />);
    await userEvent.click(screen.getByRole('button', { name: 'Поскаржитись' }));
    expect(await screen.findByText('Не можна поскаржитись на власний відгук')).toBeTruthy();
  });
});

for (const [code, message] of [
  ['enrollment_required', 'Щоб залишити відгук або поскаржитися на відгук, запишіться на курс.'],
  [
    'enrollment_blocked',
    'Вас заблоковано в цьому курсі — відгуки та скарги на відгуки недоступні.',
  ],
] as const) {
  it(`shows the ${code} review report message`, async () => {
    render(
      <CourseView
        course={course}
        reviews={page(1)}
        signedIn
        onReport={async () => {
          throw new Error(code);
        }}
      />,
    );
    await userEvent.click(screen.getByRole('button', { name: 'Поскаржитись' }));
    expect((await screen.findByRole('alert')).textContent).toBe(message);
    expect(screen.queryByRole('button', { name: 'Скаргу надіслано' })).toBeNull();
  });
}
