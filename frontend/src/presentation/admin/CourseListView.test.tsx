import { describe, expect, it } from '@rstest/core';
import { render, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { createI18n, I18nProvider } from '../../app/providers/i18n';
import { asBrandedId } from '../../domain/course/course';
import type { AdminCoursePage, CourseStatus } from '../../domain/course/course';
import { CourseListView } from './CourseListView';

const firstPage: AdminCoursePage = {
  items: [
    {
      id: asBrandedId('10000000-0000-4000-8000-000000000001', 'course'),
      slug: 'go-architecture',
      title: 'Архітектура Go',
      status: 'draft',
      instructorId: '20000000-0000-4000-8000-000000000002',
      updatedAt: '2026-09-05T10:15:30.000000Z',
    },
  ],
  pagination: { limit: 20, offset: 0, total: 21 },
};

describe('admin course list', () => {
  it('renders title, slug, localized status, owner and updated time with an editor link', () => {
    render(
      <CourseListView
        page={firstPage}
        state={{ page: 1 }}
        onStatusChange={() => {}}
        onPageChange={() => {}}
      />,
    );

    const table = screen.getByRole('table', { name: 'Курси' });
    expect(within(table).getByRole('columnheader', { name: 'Власник' })).toBeTruthy();
    expect(within(table).getByText('go-architecture')).toBeTruthy();
    expect(within(table).getByText('Чернетка')).toBeTruthy();
    expect(within(table).getByText('20000000-0000-4000-8000-000000000002')).toBeTruthy();
    expect(within(table).getByRole('link', { name: 'Архітектура Go' }).getAttribute('href')).toBe(
      '/admin/courses/10000000-0000-4000-8000-000000000001',
    );
    expect(within(table).getByText(/2026/).closest('time')?.getAttribute('datetime')).toBe(
      '2026-09-05T10:15:30.000000Z',
    );
  });

  it('changes the status filter and moves to the next page through accessible controls', async () => {
    const user = userEvent.setup();
    let status: CourseStatus | undefined;
    let page = 1;
    render(
      <CourseListView
        page={firstPage}
        state={{ page: 1 }}
        onStatusChange={(nextStatus) => {
          status = nextStatus;
        }}
        onPageChange={(nextPage) => {
          page = nextPage;
        }}
      />,
    );

    await user.click(screen.getByRole('tab', { name: 'Опубліковано' }));
    expect(status).toBe('published');
    await user.click(screen.getByRole('button', { name: 'Далі' }));
    expect(page).toBe(2);
  });

  it('renders the localized empty state and both pagination buttons in English', () => {
    render(
      <I18nProvider instance={createI18n('en')}>
        <CourseListView
          page={{ items: [], pagination: { limit: 20, offset: 0, total: 0 } }}
          state={{ page: 1 }}
          onStatusChange={() => {}}
          onPageChange={() => {}}
        />
      </I18nProvider>,
    );

    expect(screen.getByRole('heading', { name: 'No courses yet' })).toBeTruthy();
    expect(screen.queryByRole('table')).toBeNull();
    const pages = screen.getByRole('navigation', { name: 'Course pages' });
    expect(within(pages).getByRole('button', { name: 'Back' })).toBeTruthy();
    expect(within(pages).getByRole('button', { name: 'Next' })).toBeTruthy();
    expect(pages.textContent).toContain('No pages');
  });
});
