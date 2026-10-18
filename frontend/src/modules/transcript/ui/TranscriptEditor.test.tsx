import { expect, test } from '@rstest/core';
import { fireEvent, render, screen, within } from '@testing-library/react';
import type { TranscriptErrorCode, TranscriptSegment } from '../domain/transcript';
import { createMemoryTranscriptHistory } from '../adapters/transcript-history-memory';
import { TranscriptEditor } from './TranscriptEditor';

const lessonId = '30000000-0000-4000-8000-000000000001';
const documentId = '60000000-0000-4000-8000-000000000001';

const segments: TranscriptSegment[] = [
  {
    id: '70000000-0000-4000-8000-000000000001',
    startMs: 0,
    endMs: 1000,
    text: 'Перша',
    documentIds: [],
  },
];

function editor(
  save: TranscriptEditorActionsSave,
  overrides: Partial<Parameters<typeof TranscriptEditor>[0]> = {},
) {
  return render(
    <TranscriptEditor
      lessonId={lessonId}
      language="uk"
      expectedUpdatedAt="2026-09-18T00:00:00Z"
      segments={segments}
      documents={[{ id: documentId, label: 'Довідка', url: 'https://example.test/a' }]}
      actions={{ save }}
      history={{ key: 'author:lesson', store: createMemoryTranscriptHistory() }}
      {...overrides}
    />,
  );
}

type TranscriptEditorActionsSave = (
  language: string,
  expectedUpdatedAt: string | undefined,
  segments: readonly TranscriptSegment[],
) => Promise<TranscriptErrorCode | null>;

test('the optimistic-lock token travels with the save', async () => {
  let seen: string | undefined = 'unset';
  editor(async (_language, token) => {
    seen = token;
    return null;
  });
  fireEvent.click(screen.getByRole('button', { name: 'Зберегти транскрипт' }));
  await screen.findByText('Транскрипт збережено.');
  expect(seen).toBe('2026-09-18T00:00:00Z');
});

test('a half-typed time does not reorder the lines under the cursor', () => {
  // Times are text while editing and parsed only on save. Parsing per keystroke turns "1" into a
  // real value the moment it is typed, and the line jumps somewhere else mid-edit.
  editor(async () => null);
  const start = screen.getByLabelText('Початок репліки 1, мс');
  fireEvent.change(start, { target: { value: '1' } });
  expect((start as HTMLInputElement).value).toBe('1');
  fireEvent.change(start, { target: { value: '15' } });
  expect((start as HTMLInputElement).value).toBe('15');
});

test('a time that is not a whole number is refused before anything is sent', async () => {
  let called = false;
  editor(async () => {
    called = true;
    return null;
  });
  fireEvent.change(screen.getByLabelText('Кінець репліки 1, мс'), { target: { value: 'пізніше' } });
  fireEvent.click(screen.getByRole('button', { name: 'Зберегти транскрипт' }));
  await screen.findByText('Перевірте часи й текст реплік.');
  expect(called, 'a malformed time must not reach the server').toBe(false);
});

test('a conflict is named, not shown as a generic failure', async () => {
  editor(async () => 'transcript_modified');
  fireEvent.click(screen.getByRole('button', { name: 'Зберегти транскрипт' }));
  expect(
    await screen.findByText('Транскрипт змінили, поки ви редагували. Перезавантажте сторінку.'),
  ).toBeTruthy();
});

test('an unknown code degrades to a general message rather than showing the code', async () => {
  // A code outside the named set is contract drift. Rendering it verbatim would make a drift look
  // like a product state the author is supposed to understand.
  editor(async () => 'lesson_not_found');
  fireEvent.click(screen.getByRole('button', { name: 'Зберегти транскрипт' }));
  expect(await screen.findByText('Сервіс недоступний. Спробуйте ще раз.')).toBeTruthy();
});

test('a line can be added and removed, and a document attached', async () => {
  let saved: readonly TranscriptSegment[] = [];
  editor(async (_language, _token, next) => {
    saved = next;
    return null;
  });
  fireEvent.click(screen.getByRole('button', { name: 'Додати репліку' }));
  expect(screen.getAllByRole('listitem')).toHaveLength(2);

  // Each line has its own checkbox for the same document, so the lookup must be scoped to a line
  // — an unscoped one would attach the document to whichever line happened to render first.
  const firstLine = screen.getAllByRole('listitem')[0];
  if (firstLine === undefined) throw new Error('the editor rendered no lines');
  fireEvent.click(within(firstLine).getByRole('checkbox', { name: 'Довідка' }));
  fireEvent.click(screen.getByRole('button', { name: 'Видалити репліку 2' }));
  expect(screen.getAllByRole('listitem')).toHaveLength(1);

  fireEvent.click(screen.getByRole('button', { name: 'Зберегти транскрипт' }));
  await screen.findByText('Транскрипт збережено.');
  expect(saved).toHaveLength(1);
  expect(saved[0]?.documentIds).toEqual([documentId]);
});

// --- local recovery ------------------------------------------------------------------------
// The editor keeps its lines in component state, so a reload used to lose them without a word.
// These pin the offer, not the storage: the store is the in-memory adapter, which is exactly why
// there is a second implementation of the port.

const storedLine = {
  id: '70000000-0000-4000-8000-000000000001',
  startMs: '0',
  endMs: '1000',
  text: 'Недописана правка',
  documentIds: [] as readonly string[],
};

async function editorWithSnapshot(
  snapshot: Parameters<TranscriptHistoryStoreForTest['write']>[0] | null,
  overrides: Partial<Parameters<typeof TranscriptEditor>[0]> = {},
) {
  const store = createMemoryTranscriptHistory();
  if (snapshot !== null) await store.write(snapshot);
  const rendered = editor(async () => null, {
    history: { key: 'author:lesson', store },
    ...overrides,
  });
  return { store, rendered };
}

type TranscriptHistoryStoreForTest = ReturnType<typeof createMemoryTranscriptHistory>;

test('unsaved lines left on this device are offered back, never restored on their own', async () => {
  await editorWithSnapshot({
    key: 'author:lesson',
    capturedAt: 1,
    expectedUpdatedAt: '2026-09-18T00:00:00Z',
    lines: [storedLine],
  });

  // The offer is visible...
  const restore = await screen.findByRole('button', { name: 'Відновити зміни' });
  // ...and until it is taken, the editor still shows the server's text. Restoring silently would
  // overwrite text the author may have cleared on purpose.
  expect((screen.getByLabelText('Текст репліки 1') as HTMLTextAreaElement).value).toBe('Перша');

  fireEvent.click(restore);
  expect((screen.getByLabelText('Текст репліки 1') as HTMLTextAreaElement).value).toBe(
    'Недописана правка',
  );
  expect(screen.queryByRole('button', { name: 'Відновити зміни' })).toBeNull();
});

test('a snapshot identical to the server is not offered', async () => {
  // Offering a no-op recovery teaches the author to dismiss the prompt without reading it, which
  // disarms it for the one time it matters.
  await editorWithSnapshot({
    key: 'author:lesson',
    capturedAt: 1,
    expectedUpdatedAt: '2026-09-18T00:00:00Z',
    lines: [{ ...storedLine, text: 'Перша' }],
  });
  await screen.findByLabelText('Текст репліки 1');
  expect(screen.queryByRole('button', { name: 'Відновити зміни' })).toBeNull();
});

test('a snapshot taken against an older server version is offered as an overwrite', async () => {
  // Phase 8 creates this case deliberately: recognition writes while the author is away, so the
  // base the snapshot was taken against is not the base it would land on.
  await editorWithSnapshot({
    key: 'author:lesson',
    capturedAt: 1,
    expectedUpdatedAt: '2026-09-17T00:00:00Z',
    lines: [storedLine],
  });
  expect(
    await screen.findByText(
      'На цьому пристрої лишилися незбережені зміни, але транскрипт на сервері відтоді змінився. Якщо відновити й зберегти, новіший текст буде перезаписано.',
    ),
  ).toBeTruthy();
});

test('a save the server accepted drops the local copy; nothing else does', async () => {
  const { store } = await editorWithSnapshot({
    key: 'author:lesson',
    capturedAt: 1,
    expectedUpdatedAt: '2026-09-18T00:00:00Z',
    lines: [storedLine],
  });
  await screen.findByRole('button', { name: 'Відновити зміни' });
  expect(await store.read('author:lesson')).not.toBeNull();

  fireEvent.click(screen.getByRole('button', { name: 'Зберегти транскрипт' }));
  await screen.findByText('Транскрипт збережено.');
  expect(await store.read('author:lesson')).toBeNull();
});

test('a store that refuses everything costs an offer, never the save', async () => {
  // Best-effort is the contract: browsers evict this store and block it outright in private
  // windows, and the domain's own ceiling (5000 lines of 1000 runes) makes a quota breach real.
  const refusing = {
    read: () => Promise.reject(new Error('blocked')),
    write: () => Promise.reject(new Error('quota')),
    clear: () => Promise.reject(new Error('blocked')),
    keys: () => Promise.reject(new Error('blocked')),
  };
  let saved = false;
  editor(
    async () => {
      saved = true;
      return null;
    },
    { history: { key: 'author:lesson', store: refusing } },
  );
  fireEvent.change(screen.getByLabelText('Текст репліки 1'), { target: { value: 'Змінено' } });
  fireEvent.blur(screen.getByLabelText('Текст репліки 1'));
  fireEvent.click(screen.getByRole('button', { name: 'Зберегти транскрипт' }));
  await screen.findByText('Транскрипт збережено.');
  expect(saved).toBe(true);
  expect(screen.queryByRole('button', { name: 'Відновити зміни' })).toBeNull();
});

test('leaving a field captures what is on screen, including a half-typed time', async () => {
  const { store } = await editorWithSnapshot(null);
  const start = screen.getByLabelText('Початок репліки 1, мс');
  fireEvent.change(start, { target: { value: '15' } });
  fireEvent.blur(start);
  const snapshot = await store.read('author:lesson');
  // "15" is text, not a number: a snapshot of parsed segments could not carry this state back.
  expect(snapshot?.lines[0]?.startMs).toBe('15');
  expect(snapshot?.expectedUpdatedAt).toBe('2026-09-18T00:00:00Z');
  expect(snapshot?.language).toBe('uk');
});

// --- draft offer (C17, C18) -----------------------------------------------------------------
// An ASR draft the author asked to load goes through the SAME accept/dismiss affordance as
// crash recovery: it never overwrites what is on screen by itself.

const draftSegments: TranscriptSegment[] = [
  {
    id: '80000000-0000-4000-8000-000000000001',
    startMs: 0,
    endMs: 900,
    text: 'Розпізнаний текст',
    documentIds: [],
  },
];

test('a loaded draft is offered, not applied, and the text on screen is unchanged until accepted', () => {
  editor(async () => null, {
    draftOffer: { segments: draftSegments, language: 'uk', updatedAt: '2026-09-24T00:00:00Z' },
  });
  expect(screen.getByRole('button', { name: 'Завантажити чернетку' })).toBeTruthy();
  expect((screen.getByLabelText('Текст репліки 1') as HTMLTextAreaElement).value).toBe('Перша');
});

test('a recovery snapshot restores its language after a reload', async () => {
  let savedLanguage = '';
  await editorWithSnapshot(
    {
      key: 'author:lesson',
      capturedAt: 1,
      language: 'en',
      expectedUpdatedAt: '2026-09-18T00:00:00Z',
      lines: [storedLine],
    },
    {
      actions: {
        save: async (nextLanguage) => {
          savedLanguage = nextLanguage;
          return null;
        },
      },
    },
  );

  fireEvent.click(await screen.findByRole('button', { name: 'Відновити зміни' }));
  fireEvent.click(screen.getByRole('button', { name: 'Зберегти транскрипт' }));
  await screen.findByText('Транскрипт збережено.');
  expect(savedLanguage).toBe('en');
});

test('a draft in another language is explained and cannot replace a saved transcript', () => {
  editor(async () => null, {
    hasSavedTranscript: true,
    draftOffer: { segments: draftSegments, language: 'en', updatedAt: 'T1' },
  });

  expect(
    screen.getByText(
      'Чернетка мовою en, а збережений транскрипт — мовою uk. Перегенеруйте чернетку мовою uk.',
    ),
  ).toBeTruthy();
  expect(screen.queryByRole('button', { name: 'Завантажити чернетку' })).toBeNull();
  expect((screen.getByLabelText('Текст репліки 1') as HTMLTextAreaElement).value).toBe('Перша');
});

test('a draft language is accepted when there is no saved transcript', async () => {
  let savedLanguage = '';
  editor(
    async (nextLanguage) => {
      savedLanguage = nextLanguage;
      return null;
    },
    {
      hasSavedTranscript: false,
      segments: [],
      draftOffer: { segments: draftSegments, language: 'en', updatedAt: 'T1' },
    },
  );

  fireEvent.click(screen.getByRole('button', { name: 'Завантажити чернетку' }));
  fireEvent.click(screen.getByRole('button', { name: 'Зберегти транскрипт' }));
  await screen.findByText('Транскрипт збережено.');
  expect(savedLanguage).toBe('en');
});

test('accepting the draft offer replaces the lines and reports the offer as handled', () => {
  let handled = false;
  editor(async () => null, {
    draftOffer: { segments: draftSegments, language: 'uk', updatedAt: '2026-09-24T00:00:00Z' },
    onDraftOfferHandled: () => {
      handled = true;
    },
  });
  fireEvent.click(screen.getByRole('button', { name: 'Завантажити чернетку' }));
  expect((screen.getByLabelText('Текст репліки 1') as HTMLTextAreaElement).value).toBe(
    'Розпізнаний текст',
  );
  expect(handled).toBe(true);
});

test('a loaded draft snapshot keeps the loaded generation after regeneration', () => {
  const store = createMemoryTranscriptHistory();
  const draftOffer = { segments: draftSegments, language: 'uk', updatedAt: 'T1' };
  const rendered = render(
    <TranscriptEditor
      lessonId={lessonId}
      language="uk"
      expectedUpdatedAt="2026-09-18T00:00:00Z"
      segments={segments}
      documents={[]}
      actions={{ save: async () => null }}
      history={{ key: 'author:lesson', store }}
      draftOffer={draftOffer}
      draftUpdatedAt="T1"
    />,
  );

  fireEvent.click(screen.getByRole('button', { name: 'Завантажити чернетку' }));
  fireEvent.change(screen.getByLabelText('Текст репліки 1'), {
    target: { value: 'Ручна правка після завантаження' },
  });
  rendered.rerender(
    <TranscriptEditor
      lessonId={lessonId}
      language="uk"
      expectedUpdatedAt="2026-09-18T00:00:00Z"
      segments={segments}
      documents={[]}
      actions={{ save: async () => null }}
      history={{ key: 'author:lesson', store }}
      draftOffer={draftOffer}
      draftUpdatedAt="T2"
    />,
  );
  fireEvent.blur(screen.getByLabelText('Текст репліки 1'));

  return store.read('author:lesson').then((snapshot) => {
    expect(snapshot?.draftUpdatedAt).toBe('T1');
  });
});

test('dismissing the draft offer leaves the lines untouched but still reports it handled', () => {
  let handled = false;
  editor(async () => null, {
    draftOffer: { segments: draftSegments, language: 'uk', updatedAt: '2026-09-24T00:00:00Z' },
    onDraftOfferHandled: () => {
      handled = true;
    },
  });
  fireEvent.click(screen.getByRole('button', { name: 'Не завантажувати' }));
  expect((screen.getByLabelText('Текст репліки 1') as HTMLTextAreaElement).value).toBe('Перша');
  expect(handled).toBe(true);
});

test('an unsaved local edit outranks a draft offer: only one prompt shows at a time', async () => {
  await editorWithSnapshot(
    {
      key: 'author:lesson',
      capturedAt: 1,
      expectedUpdatedAt: '2026-09-18T00:00:00Z',
      lines: [storedLine],
    },
    {
      draftOffer: {
        segments: draftSegments,
        language: 'uk',
        updatedAt: '2026-09-24T00:00:00Z',
      },
    },
  );
  await screen.findByRole('button', { name: 'Відновити зміни' });
  expect(screen.queryByRole('button', { name: 'Завантажити чернетку' })).toBeNull();
});
