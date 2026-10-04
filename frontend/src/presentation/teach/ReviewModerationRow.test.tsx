import { describe, expect, it, rs } from '@rstest/core';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import type { ModerationQueueItem } from '../../domain/teach/teach';
import { ReviewModerationRow } from './ReviewModerationRow';

const item = {
  kind: 'review' as const,
  comment: null,
  openReports: 2,
  review: {
    id: '75000000-0000-4000-8000-000000000001',
    authorDisplayName: 'Студентка',
    stars: 2,
    body: 'Лекції зачитані з папірця, практики немає.',
    completedLessons: 1,
    totalLessons: 4,
    createdAt: '2026-10-01T09:00:00Z',
    editedAt: null,
  },
} satisfies ModerationQueueItem;

function renderRow(admin: boolean, onDismiss = rs.fn(), onRemove = rs.fn()) {
  render(
    <ul>
      <ReviewModerationRow
        item={item}
        busy={false}
        admin={admin}
        onDismiss={onDismiss}
        onRemove={onRemove}
      />
    </ul>,
  );
  return { onDismiss, onRemove };
}

describe('teach queue review row', () => {
  it('shows the author, stars for assistive tech, the text and the report count', () => {
    renderRow(false);
    expect(screen.getByText(/Студентка/)).toBeTruthy();
    expect(screen.getByText('★★').getAttribute('aria-hidden')).toBe('true');
    expect(screen.getByText('2 з 5')).toBeTruthy();
    expect(screen.getByText('Лекції зачитані з папірця, практики немає.')).toBeTruthy();
  });

  it('offers removal only to an admin (G9); the owner can only dismiss (Rt21)', async () => {
    const owner = renderRow(false);
    expect(screen.queryByRole('button', { name: 'Зняти' })).toBeNull();
    await userEvent.click(screen.getByRole('button', { name: /Відхилити/ }));
    expect(owner.onDismiss).toHaveBeenCalledTimes(1);
  });

  it('lets an admin remove the review', async () => {
    const admin = renderRow(true);
    await userEvent.click(screen.getByRole('button', { name: 'Зняти' }));
    expect(admin.onRemove).toHaveBeenCalledTimes(1);
  });
});
