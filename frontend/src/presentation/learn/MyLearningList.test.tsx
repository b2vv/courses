import { expect, test } from '@rstest/core';
import { render, screen } from '@testing-library/react';
import { MyLearningList } from './MyLearningList';

const item = {
  courseId: '10000000-0000-4000-8000-000000000001',
  slug: 'go-basics',
  title: 'Go на практиці',
  state: 'active' as const,
  progressPercent: 40,
  enrolledAt: '2026-10-07T10:00:00Z',
};

test('shows membership cards with progress and learning links', () => {
  render(<MyLearningList items={[item]} />);
  expect(screen.getByRole('heading', { level: 1, name: 'Моє навчання' })).toBeTruthy();
  expect(screen.getByRole('progressbar').getAttribute('aria-valuenow')).toBe('40');
  expect(screen.getByRole('link', { name: 'Почати навчання' }).getAttribute('href')).toBe(
    '/learn/go-basics',
  );
});

test('shows a blocked membership note and an empty-state catalog action', () => {
  const { rerender } = render(<MyLearningList items={[{ ...item, state: 'blocked' }]} />);
  expect(screen.getByRole('status').textContent).toContain(
    'Доступ до контенту відкрито, участь в обговореннях обмежено',
  );
  rerender(<MyLearningList items={[]} />);
  expect(screen.getByText('Ви ще не записалися на жоден курс.')).toBeTruthy();
  expect(screen.getByRole('link', { name: 'Переглянути курси' }).getAttribute('href')).toBe(
    '/courses',
  );
});
