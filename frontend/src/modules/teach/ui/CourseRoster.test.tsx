import { describe, expect, it, rs } from '@rstest/core';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import type { ComponentProps } from 'react';
import type { RosterEntry } from '../../enrollment';
import { CourseRoster } from './CourseRoster';

const active = {
  userId: '10000000-0000-4000-8000-000000000001',
  displayName: 'Марія',
  state: 'active',
  enrolledAt: '2026-10-07T10:00:00Z',
  progressPercent: 0,
} satisfies RosterEntry;

function renderRoster(overrides: Partial<ComponentProps<typeof CourseRoster>> = {}) {
  const props = {
    entries: [active],
    isLoading: false,
    error: null,
    hasNextPage: false,
    isFetchingNextPage: false,
    isPending: false,
    onRetry: rs.fn(),
    onLoadMore: rs.fn(),
    onBlock: rs.fn(),
    onUnblock: rs.fn(),
    onReportStudent: rs.fn(async () => {}),
    ...overrides,
  };
  render(<CourseRoster {...props} />);
  return props;
}

describe('CourseRoster', () => {
  it('renders an accessible 0% progress row and blocks an active participant', async () => {
    const roster = renderRoster();
    expect(screen.getByRole('progressbar', { name: '0% пройдено' })).toBeTruthy();
    await userEvent.click(screen.getByRole('button', { name: 'Заблокувати' }));
    expect(roster.onBlock).toHaveBeenCalledWith(active.userId);
  });

  it('anonymizes a missing display name and offers unblock for a blocked participant', async () => {
    const roster = renderRoster({
      entries: [{ ...active, displayName: '', state: 'blocked', progressPercent: 50 }],
    });
    expect(screen.getByText('Колишній учасник')).toBeTruthy();
    expect(screen.getByText('Заблокований')).toBeTruthy();
    await userEvent.click(screen.getByRole('button', { name: 'Розблокувати' }));
    expect(roster.onUnblock).toHaveBeenCalledWith(active.userId);
  });

  it('shows the required empty state', () => {
    renderRoster({ entries: [] });
    expect(screen.getByRole('heading', { name: 'Поки немає учасників' })).toBeTruthy();
  });

  it('retries an error and loads the next keyset page when available', async () => {
    const retry = renderRoster({ entries: [], error: new Error('dependency_unavailable') });
    await userEvent.click(screen.getByRole('button', { name: 'Спробувати ще раз' }));
    expect(retry.onRetry).toHaveBeenCalledTimes(1);

    const more = renderRoster({ hasNextPage: true });
    await userEvent.click(screen.getByRole('button', { name: 'Показати ще' }));
    expect(more.onLoadMore).toHaveBeenCalledTimes(1);
  });

  it('sends a report for the selected participant and announces success', async () => {
    const roster = renderRoster();
    await userEvent.click(screen.getByRole('button', { name: 'Поскаржитися' }));
    await userEvent.type(
      screen.getByRole('textbox', { name: 'Причина скарги' }),
      'Образливе повідомлення',
    );
    await userEvent.click(screen.getByRole('button', { name: 'Надіслати скаргу' }));
    expect(roster.onReportStudent).toHaveBeenCalledWith(active.userId, 'Образливе повідомлення');
    expect((await screen.findByRole('status')).textContent).toContain('Скаргу надіслано');
  });
});
