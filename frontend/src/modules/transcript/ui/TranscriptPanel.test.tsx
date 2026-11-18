import { expect, test, rstest } from '@rstest/core';
import { fireEvent, render, screen, within } from '@testing-library/react';
import type { LessonTranscript } from '../domain/transcript';
import { TranscriptPanel } from './TranscriptPanel';

const documentId = '60000000-0000-4000-8000-000000000001';

const transcript: LessonTranscript = {
  lessonId: '30000000-0000-4000-8000-000000000001',
  language: 'uk',
  updatedAt: '2026-09-18T00:00:00Z',
  stale: false,
  segments: [
    {
      id: '70000000-0000-4000-8000-000000000001',
      startMs: 0,
      endMs: 5_000,
      text: 'Перша репліка про канали',
      speaker: 'Лектор',
      documentIds: [documentId],
    },
    {
      id: '70000000-0000-4000-8000-000000000002',
      startMs: 65_000,
      endMs: 70_000,
      text: 'Друга репліка про горутини',
      documentIds: [],
    },
  ],
  documents: [{ id: documentId, label: 'Довідка', url: 'https://example.test/a' }],
};

test('a lesson without a transcript says so instead of showing an empty list', () => {
  render(<TranscriptPanel state={{ kind: 'none' }} positionMs={0} />);
  expect(screen.getByText('У цього уроку ще немає транскрипту.')).toBeTruthy();
  expect(screen.queryByRole('list')).toBeNull();
});

test('exactly one line is current, and it follows the position', () => {
  const { rerender } = render(
    <TranscriptPanel
      state={{ kind: 'transcript', transcript }}
      positionMs={1_000}
      onSeek={() => undefined}
    />,
  );
  expect(
    screen.getAllByRole('listitem').filter((item) => item.getAttribute('aria-current')),
  ).toHaveLength(1);
  expect(
    screen
      .getByRole('button', { name: 'Перейти до 00:00' })
      .closest('li')
      ?.getAttribute('aria-current'),
  ).toBe('true');

  rerender(
    <TranscriptPanel
      state={{ kind: 'transcript', transcript }}
      positionMs={66_000}
      onSeek={() => undefined}
    />,
  );
  expect(
    screen
      .getByRole('button', { name: 'Перейти до 01:05' })
      .closest('li')
      ?.getAttribute('aria-current'),
  ).toBe('true');
});

test('a position between lines makes none current rather than keeping the last one', () => {
  render(<TranscriptPanel state={{ kind: 'transcript', transcript }} positionMs={30_000} />);
  expect(
    screen.getAllByRole('listitem').filter((item) => item.getAttribute('aria-current')),
  ).toHaveLength(0);
});

test('clicking a line seeks to its start', () => {
  const seeks: number[] = [];
  render(
    <TranscriptPanel
      state={{ kind: 'transcript', transcript }}
      positionMs={0}
      onSeek={(ms) => seeks.push(ms)}
    />,
  );
  fireEvent.click(screen.getByRole('button', { name: 'Перейти до 01:05' }));
  expect(seeks).toEqual([65_000]);
});

test('without a seek a line is text, not a dead control', () => {
  // Before the first play the duration is unknown and a seek would land in the first second. A
  // DISABLED button would be worse than no button: it drops out of keyboard traversal, so the
  // text itself would stop being reachable for exactly the readers a transcript is for.
  render(<TranscriptPanel state={{ kind: 'transcript', transcript }} positionMs={0} />);
  expect(screen.queryByRole('button', { name: 'Перейти до 00:00' })).toBeNull();
  expect(screen.getByText('Перша репліка про канали')).toBeTruthy();
  expect(screen.getByText('00:00')).toBeTruthy();
});

test('search narrows to matching lines and names the empty result', () => {
  render(<TranscriptPanel state={{ kind: 'transcript', transcript }} positionMs={0} />);
  const field = screen.getByRole('searchbox');

  fireEvent.change(field, { target: { value: 'горутини' } });
  expect(screen.getAllByRole('listitem')).toHaveLength(1);
  expect(screen.getByText('Знайдено реплік: 1')).toBeTruthy();

  fireEvent.change(field, { target: { value: 'відсутнє' } });
  expect(screen.getByText('Нічого не знайдено.')).toBeTruthy();

  // Clearing is not a search with no results: the whole transcript comes back.
  fireEvent.change(field, { target: { value: '' } });
  expect(screen.getAllByRole('listitem').length).toBeGreaterThan(1);
});

test('scrolling by hand stops the follow, and it can be turned back on', () => {
  render(<TranscriptPanel state={{ kind: 'transcript', transcript }} positionMs={0} />);
  expect(screen.queryByRole('button', { name: 'Стежити за відтворенням' })).toBeNull();

  // Wheel, not scroll: autoscroll fires a scroll event too, so only user intent may pause it.
  fireEvent.wheel(screen.getByRole('list', { name: 'Транскрипт' }));
  const resume = screen.getByRole('button', { name: 'Стежити за відтворенням' });

  fireEvent.click(resume);
  expect(screen.queryByRole('button', { name: 'Стежити за відтворенням' })).toBeNull();
});

test('a stale transcript says the video was replaced', () => {
  render(
    <TranscriptPanel
      state={{ kind: 'transcript', transcript: { ...transcript, stale: true } }}
      positionMs={0}
    />,
  );
  expect(screen.getByRole('status').textContent).toContain('Відео замінили');
});

test('a line carries its documents as real links', () => {
  render(<TranscriptPanel state={{ kind: 'transcript', transcript }} positionMs={0} />);
  const list = screen.getByRole('list', { name: 'Матеріали репліки' });
  const link = within(list).getByRole('link', { name: 'Довідка' });
  expect(link.getAttribute('href')).toBe('https://example.test/a');
  expect(link.getAttribute('rel')).toContain('noopener');
});

test.each([
  [{ kind: 'loading' }, 'Завантаження'],
  [{ kind: 'none-for-track' }, 'Немає субтитрів для цієї доріжки'],
  [{ kind: 'none' }, 'У цього уроку ще немає транскрипту.'],
] as const)('M09 renders $kind state', (state, copy) => {
  render(<TranscriptPanel state={state} positionMs={0} />);
  const panel = within(screen.getByRole('region', { name: 'Транскрипт' }));
  expect(panel.getAllByRole('status')).toHaveLength(1);
  expect(panel.getByRole('status').textContent).toBe(copy);
  expect(panel.queryByRole('alert')).toBeNull();
  expect(screen.queryByRole('searchbox')).toBeNull();
});

test('M09 error offers retry', () => {
  const onRetry = rstest.fn();
  render(<TranscriptPanel state={{ kind: 'error', onRetry }} positionMs={0} />);
  const panel = within(screen.getByRole('region', { name: 'Транскрипт' }));
  expect(panel.getAllByRole('alert')).toHaveLength(1);
  expect(panel.getByRole('alert').textContent).toBe('Не вдалося завантажити транскрипт');
  expect(panel.queryByRole('status')).toBeNull();
  fireEvent.click(screen.getByRole('button', { name: 'Спробувати ще' }));
  expect(onRetry).toHaveBeenCalledTimes(1);
});

test('M09 language key resets BOTH search and follow', () => {
  const { rerender } = render(
    <TranscriptPanel key="uk" state={{ kind: 'transcript', transcript }} positionMs={0} />,
  );
  fireEvent.wheel(screen.getByRole('list', { name: 'Транскрипт' }));
  expect(screen.getByRole('button', { name: 'Стежити за відтворенням' })).toBeTruthy();
  fireEvent.change(screen.getByRole('searchbox'), { target: { value: 'горутини' } });
  rerender(
    <TranscriptPanel
      key="en"
      state={{ kind: 'transcript', transcript: { ...transcript, language: 'en' } }}
      positionMs={0}
    />,
  );
  expect(screen.getByRole('searchbox').getAttribute('value')).toBe('');
  expect(screen.getByText('Перша репліка про канали')).toBeTruthy();
  expect(screen.queryByRole('button', { name: 'Стежити за відтворенням' })).toBeNull();
});
