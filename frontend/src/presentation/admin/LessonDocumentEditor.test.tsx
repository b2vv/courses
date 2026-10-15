import axe from 'axe-core';
import { afterEach, expect, test } from '@rstest/core';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, cleanup, fireEvent, render, screen, waitFor, within } from '@testing-library/react';
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
  overrides: Partial<DocumentAuthoringRepository> = {},
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
    ...overrides,
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

test('palette tiles keep exact type names despite their descriptions', async () => {
  setup([]);
  const palette = await screen.findByRole('group', { name: 'Додати секцію' });
  for (const name of ['Текст', 'Зображення', 'Завдання', 'Відео', 'Ряд']) {
    expect(within(palette).getByRole('button', { name })).toBeTruthy();
  }
  expect(within(palette).getByText('Пояснення та списки').getAttribute('aria-hidden')).toBe('true');
});

test('section identity has ordered indices and decorative icons without axe violations', async () => {
  setup([videoSection, imageSection]);
  await screen.findByText('Тут грає відео уроку (додається окремо).');
  const sections = Array.from(document.querySelectorAll('.lesson-editor-section'));
  for (const [index, section] of sections.entries()) {
    expect(section.querySelector('.lesson-editor-section-index')?.textContent).toBe(
      String(index + 1),
    );
    expect(section.querySelector('svg')?.getAttribute('aria-hidden')).toBe('true');
    expect(section.querySelector('svg')?.hasAttribute('role')).toBe(false);
  }
  const editor = screen.getByRole('region', { name: 'Редактор уроку' });
  expect((await axe.run(editor)).violations).toEqual([]);
  const move = screen.getAllByRole('button', { name: 'Пересунути вгору' }).at(1);
  if (move === undefined) throw new Error('Missing second section move control');
  fireEvent.click(move);
  expect(sectionTypes()).toEqual(['image', 'video']);
  expect(
    Array.from(document.querySelectorAll('.lesson-editor-section-index')).map(
      (node) => node.textContent,
    ),
  ).toEqual(['1', '2']);
});

test('toolbar retains aria-pressed when toggled', async () => {
  setup([{ type: 'text', id: 'text', content: { blocks: [] } }]);
  const toolbar = await screen.findByRole('toolbar', { name: 'Форматування тексту' });
  const link = within(toolbar).getByRole('button', { name: 'Посилання', pressed: false });
  fireEvent.click(link);
  expect(within(toolbar).getByRole('button', { name: 'Посилання', pressed: true })).toBe(link);
  fireEvent.click(link);
  expect(link.getAttribute('aria-pressed')).toBe('false');
});

function deferred<T>() {
  let resolve: (value: T) => void = () => {
    throw new Error('Deferred promise not initialized');
  };
  const promise = new Promise<T>((resolvePromise) => {
    resolve = resolvePromise;
  });
  return { promise, resolve };
}

type UploadResult = Awaited<ReturnType<DocumentAuthoringRepository['uploadImage']>>;
const uploadedImage: UploadResult = {
  ok: true,
  value: {
    imageId: '60000000-0000-4000-8000-000000000002',
    contentType: 'image/png',
    width: 1,
    height: 1,
  },
};
const secondImageSection = { ...imageSection, id: 'img-b', alt: 'Друга схема' };

function sectionForAlt(alt: string): HTMLElement {
  const section = screen.getByDisplayValue(alt).closest('.lesson-editor-section');
  if (!(section instanceof HTMLElement)) throw new Error(`Missing image section: ${alt}`);
  return section;
}

function selectImage(section: HTMLElement, name: string) {
  const input = section.querySelector('input[type="file"]');
  if (!(input instanceof HTMLInputElement)) throw new Error('Missing image upload input');
  const file = new File(['image'], name, { type: 'image/png' });
  fireEvent.change(input, { target: { files: [file] } });
  return file;
}

test('uploading an image shows pending feedback only in its own section', async () => {
  const pending = deferred<UploadResult>();
  const uploads: File[] = [];
  setup([imageSection, secondImageSection], undefined, {
    uploadImage: async (_lessonId, file) => {
      uploads.push(file);
      return pending.promise;
    },
  });
  await screen.findByDisplayValue('Схема');
  const first = sectionForAlt('Схема');
  const second = sectionForAlt('Друга схема');
  const file = selectImage(first, 'first.png');
  await waitFor(() => expect(uploads).toEqual([file]));
  expect(await within(first).findByText('Завантажуємо…')).toBeTruthy();
  expect(within(second).queryByText('Завантажуємо…')).toBeNull();
  await act(async () => {
    pending.resolve(uploadedImage);
  });
  await waitFor(() => expect(within(first).queryByText('Завантажуємо…')).toBeNull());
});

test('successful save announces draft saved as a status without an alert', async () => {
  setup([imageSection]);
  await screen.findByDisplayValue('Схема');
  fireEvent.click(screen.getByRole('button', { name: 'Зберегти чернетку' }));
  const message = await screen.findByText('Чернетку збережено');
  expect(
    screen.getAllByRole('status').filter((node) => node.textContent === 'Чернетку збережено'),
  ).toEqual([message]);
  expect(screen.queryAllByRole('alert')).toEqual([]);
});

test('an alt edit made while save is pending survives save completion', async () => {
  const pending = deferred<Awaited<ReturnType<DocumentAuthoringRepository['saveDraft']>>>();
  const savedDocuments: LessonDocument[] = [];
  setup([imageSection], undefined, {
    saveDraft: async (_lessonId, document) => {
      savedDocuments.push(document);
      return pending.promise;
    },
  });
  await screen.findByDisplayValue('Схема');
  fireEvent.click(screen.getByRole('button', { name: 'Зберегти чернетку' }));
  await waitFor(() => expect(savedDocuments).toHaveLength(1));
  expect(savedDocuments[0]?.sections).toEqual([imageSection]);
  expect(screen.queryByText('Чернетку збережено')).toBeNull();
  fireEvent.change(screen.getByRole('textbox', { name: 'Опис зображення' }), {
    target: { value: 'Правка під час збереження' },
  });
  expect(screen.getByDisplayValue('Правка під час збереження')).toBeTruthy();
  await act(async () => {
    pending.resolve({ ok: true, value: { updatedAt: '2026-10-15T01:00:00Z' } });
  });
  await screen.findByText('Чернетку збережено');
  expect(screen.getByDisplayValue('Правка під час збереження')).toBeTruthy();
  expect(screen.queryByDisplayValue('Схема')).toBeNull();
});

test('settling the first parallel upload does not clear the second upload indicator', async () => {
  const firstPending = deferred<UploadResult>();
  const secondPending = deferred<UploadResult>();
  const uploads: File[] = [];
  const completed: string[] = [];
  setup([imageSection, secondImageSection], undefined, {
    uploadImage: async (_lessonId, file) => {
      uploads.push(file);
      const result = await (file.name === 'first.png'
        ? firstPending.promise
        : secondPending.promise);
      completed.push(file.name);
      return result;
    },
  });
  await screen.findByDisplayValue('Схема');
  const first = sectionForAlt('Схема');
  const second = sectionForAlt('Друга схема');
  const firstFile = selectImage(first, 'first.png');
  expect(await within(first).findByText('Завантажуємо…')).toBeTruthy();
  const secondFile = selectImage(second, 'second.png');
  await waitFor(() => expect(uploads).toEqual([firstFile, secondFile]));
  expect(await within(second).findByText('Завантажуємо…')).toBeTruthy();
  await act(async () => {
    firstPending.resolve(uploadedImage);
  });
  expect(completed).toEqual(['first.png']);
  expect(within(second).getByText('Завантажуємо…')).toBeTruthy();
  await act(async () => {
    secondPending.resolve(uploadedImage);
  });
  await waitFor(() => expect(within(second).queryByText('Завантажуємо…')).toBeNull());
  expect(completed).toEqual(['first.png', 'second.png']);
});
