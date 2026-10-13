import { afterEach, expect, test } from '@rstest/core';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { cleanup, fireEvent, render, screen, within } from '@testing-library/react';
import { createI18n, I18nProvider } from '../../app/providers/i18n';
import { createAdminDocumentQueries } from '../../app/composition/admin-document-queries';
import type { LessonDocument } from '../../domain/document/document';
import type { DocumentAuthoringRepository } from '../../ports/DocumentAuthoringRepository';
import { LessonDocumentEditor } from './LessonDocumentEditor';

afterEach(cleanup);

const lessonId = '20000000-0000-4000-8000-000000000001';

type Calls = {
  saved: { document: LessonDocument; expectedUpdatedAt: string | null }[];
  published: { expectedUpdatedAt: string }[];
};

function setup(
  sections: LessonDocument['sections'],
  updatedAt: string | null = '2026-10-13T00:00:00Z',
) {
  const calls: Calls = { saved: [], published: [] };
  const repository: DocumentAuthoringRepository = {
    getDraft: async () => ({
      ok: true,
      value: { document: { schemaVersion: 1, locale: 'uk', sections }, updatedAt },
    }),
    saveDraft: async (_lessonId, document, expectedUpdatedAt) => {
      calls.saved.push({ document, expectedUpdatedAt });
      return { ok: true, value: { updatedAt: '2026-10-13T01:00:00Z' } };
    },
    publish: async (_lessonId, expectedUpdatedAt) => {
      calls.published.push({ expectedUpdatedAt });
      return { ok: true, value: { updatedAt: '2026-10-13T02:00:00Z' } };
    },
    uploadImage: async () => ({ ok: false, error: { code: 'dependency_unavailable' } }),
  };
  const queries = createAdminDocumentQueries(repository);
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  render(
    <I18nProvider instance={createI18n('uk')}>
      <QueryClientProvider client={client}>
        <LessonDocumentEditor lessonId={lessonId} documents={queries} />
      </QueryClientProvider>
    </I18nProvider>,
  );
  return calls;
}

function sectionTypes(): string[] {
  return Array.from(document.querySelectorAll('.lesson-editor-section')).map(
    (node) => node.getAttribute('data-type') ?? '',
  );
}

const videoSection = { type: 'video', id: 'v' } as const;
const imageSection = {
  type: 'image',
  id: 'img',
  imageId: '60000000-0000-4000-8000-000000000001',
  contentType: 'image/png',
  alt: 'Схема',
} as const;

test('the loaded sections render in order', async () => {
  setup([videoSection, imageSection]);
  expect(await screen.findByText('Тут грає відео уроку (додається окремо).')).toBeTruthy();
  expect(sectionTypes()).toEqual(['video', 'image']);
});

test('adding a text section appends a rich-text editor', async () => {
  setup([videoSection]);
  await screen.findByText('Тут грає відео уроку (додається окремо).');
  const addGroup = screen.getByRole('group', { name: 'Додати секцію' });
  fireEvent.click(within(addGroup).getByRole('button', { name: 'Текст' }));
  expect(sectionTypes()).toEqual(['video', 'text']);
  expect(screen.getByRole('textbox', { name: 'Текст' })).toBeTruthy();
});

test('move keeps the button enabled at the top (no isDisabled focus trap) and is a no-op there', async () => {
  setup([videoSection, imageSection]);
  await screen.findByText('Тут грає відео уроку (додається окремо).');
  const firstSection = document.querySelectorAll('.lesson-editor-section')[0] as HTMLElement;
  const moveUp = within(firstSection).getByRole('button', { name: 'Пересунути вгору' });
  // The reorder button is never `disabled`: a native disabled button drops focus to <body>.
  expect(moveUp.hasAttribute('disabled')).toBe(false);
  fireEvent.click(moveUp);
  expect(sectionTypes()).toEqual(['video', 'image']); // unchanged at the edge
});

test('moving the second section up reorders the list', async () => {
  setup([videoSection, imageSection]);
  await screen.findByText('Тут грає відео уроку (додається окремо).');
  const second = document.querySelectorAll('.lesson-editor-section')[1] as HTMLElement;
  fireEvent.click(within(second).getByRole('button', { name: 'Пересунути вгору' }));
  expect(sectionTypes()).toEqual(['image', 'video']);
});

test('save sends the built document under the loaded optimistic-lock token', async () => {
  const calls = setup([videoSection, imageSection], '2026-10-13T00:00:00Z');
  await screen.findByText('Тут грає відео уроку (додається окремо).');
  fireEvent.click(screen.getByRole('button', { name: 'Зберегти чернетку' }));
  await screen.findByText('Тут грає відео уроку (додається окремо).');
  expect(calls.saved).toHaveLength(1);
  expect(calls.saved[0]?.expectedUpdatedAt).toBe('2026-10-13T00:00:00Z');
  expect(calls.saved[0]?.document.sections).toHaveLength(2);
});

test('publish saves the current edits first, then publishes the saved token', async () => {
  const calls = setup([videoSection], '2026-10-13T00:00:00Z');
  await screen.findByText('Тут грає відео уроку (додається окремо).');
  fireEvent.click(screen.getByRole('button', { name: 'Опублікувати' }));
  await screen.findByRole('button', { name: 'Опублікувати' });
  // Give the chained mutateAsync calls a tick to resolve.
  await Promise.resolve();
  await new Promise((resolve) => setTimeout(resolve, 0));
  expect(calls.saved).toHaveLength(1);
  expect(calls.published).toHaveLength(1);
  expect(calls.published[0]?.expectedUpdatedAt).toBe('2026-10-13T01:00:00Z');
});
