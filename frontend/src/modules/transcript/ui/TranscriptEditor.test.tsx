import { expect, test, rstest } from '@rstest/core';
import { act, fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { createI18n, I18nProvider } from '../../../app/providers/i18n';
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

async function editor(
  save: TranscriptEditorActionsSave,
  overrides: Partial<Parameters<typeof TranscriptEditor>[0]> = {},
) {
  const view = render(
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
  await act(async () => {});
  return view;
}

type TranscriptEditorActionsSave = (
  language: string,
  expectedUpdatedAt: string | undefined,
  segments: readonly TranscriptSegment[],
) => Promise<TranscriptErrorCode | null>;

test('the optimistic-lock token travels with the save', async () => {
  let seen: string | undefined = 'unset';
  await editor(async (_language, token) => {
    seen = token;
    return null;
  });
  fireEvent.click(screen.getByRole('button', { name: 'Зберегти транскрипт' }));
  await screen.findByText('Транскрипт збережено.');
  expect(seen).toBe('2026-09-18T00:00:00Z');
});

test('a half-typed time does not reorder the lines under the cursor', async () => {
  // Times are text while editing and parsed only on save. Parsing per keystroke turns "1" into a
  // real value the moment it is typed, and the line jumps somewhere else mid-edit.
  await editor(async () => null);
  const start = screen.getByLabelText('Початок репліки 1, мс');
  fireEvent.change(start, { target: { value: '1' } });
  expect((start as HTMLInputElement).value).toBe('1');
  fireEvent.change(start, { target: { value: '15' } });
  expect((start as HTMLInputElement).value).toBe('15');
});

test('a time that is not a whole number is refused before anything is sent', async () => {
  let called = false;
  await editor(async () => {
    called = true;
    return null;
  });
  fireEvent.change(screen.getByLabelText('Кінець репліки 1, мс'), { target: { value: 'пізніше' } });
  fireEvent.click(screen.getByRole('button', { name: 'Зберегти транскрипт' }));
  await screen.findByText('Перевірте часи й текст реплік.');
  expect(called, 'a malformed time must not reach the server').toBe(false);
});

test('a conflict is named, not shown as a generic failure', async () => {
  await editor(async () => 'transcript_modified');
  fireEvent.click(screen.getByRole('button', { name: 'Зберегти транскрипт' }));
  expect(
    await screen.findByText('Транскрипт змінили, поки ви редагували. Перезавантажте сторінку.'),
  ).toBeTruthy();
});

test('an unknown code degrades to a general message rather than showing the code', async () => {
  // A code outside the named set is contract drift. Rendering it verbatim would make a drift look
  // like a product state the author is supposed to understand.
  await editor(async () => 'lesson_not_found');
  fireEvent.click(screen.getByRole('button', { name: 'Зберегти транскрипт' }));
  expect(await screen.findByText('Сервіс недоступний. Спробуйте ще раз.')).toBeTruthy();
});

test('a line can be added and removed, and a document attached', async () => {
  let saved: readonly TranscriptSegment[] = [];
  await editor(async (_language, _token, next) => {
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
  const rendered = await editor(async () => null, {
    history: { key: 'author:lesson', store },
    ...overrides,
  });
  return { store, rendered };
}

type TranscriptHistoryStoreForTest = ReturnType<typeof createMemoryTranscriptHistory>;

test('unsaved lines left on this device are offered back, never restored on their own', async () => {
  await editorWithSnapshot({
    key: 'author:lesson',
    revision: crypto.randomUUID(),
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
    revision: crypto.randomUUID(),
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
    revision: crypto.randomUUID(),
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
    revision: crypto.randomUUID(),
    capturedAt: 1,
    expectedUpdatedAt: '2026-09-18T00:00:00Z',
    lines: [storedLine],
  });
  await screen.findByRole('button', { name: 'Відновити зміни' });
  expect(await store.read('author:lesson')).not.toBeNull();
  fireEvent.click(screen.getByRole('button', { name: 'Відновити зміни' }));

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
    clearIfVersion: () => Promise.reject(new Error('blocked')),
    keys: () => Promise.reject(new Error('blocked')),
  };
  let saved = false;
  await editor(
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

test('M3-R13 capture writes baseLanguage with the token and half-typed time', async () => {
  const { store } = await editorWithSnapshot(null);
  const start = screen.getByLabelText('Початок репліки 1, мс');
  fireEvent.change(start, { target: { value: '15' } });
  fireEvent.blur(start);
  const snapshot = await store.read('author:lesson');
  // "15" is text, not a number: a snapshot of parsed segments could not carry this state back.
  expect(snapshot?.lines[0]?.startMs).toBe('15');
  expect(snapshot?.expectedUpdatedAt).toBe('2026-09-18T00:00:00Z');
  expect(snapshot?.language).toBe('uk');
  expect(snapshot?.baseLanguage).toBe('uk');
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

test('a loaded draft is offered, not applied, and the text on screen is unchanged until accepted', async () => {
  await editor(async () => null, {
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
      revision: crypto.randomUUID(),
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

test('draft offer loading no longer applies the obsolete language-mismatch branch', async () => {
  await editor(async () => null, {
    hasSavedTranscript: true,
    draftOffer: { segments: draftSegments, language: 'en', updatedAt: 'T1' },
  });
  expect(screen.getByRole('button', { name: 'Завантажити чернетку' })).toBeTruthy();
});

test('a draft language is accepted when there is no saved transcript', async () => {
  let savedLanguage = '';
  await editor(
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

test('accepting the draft offer replaces the lines and reports the offer as handled', async () => {
  let handled = false;
  await editor(async () => null, {
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

test('a loaded draft snapshot keeps the loaded generation after regeneration', async () => {
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

  await act(async () => {});
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

test('dismissing the draft offer leaves the lines untouched but still reports it handled', async () => {
  let handled = false;
  await editor(async () => null, {
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
      revision: crypto.randomUUID(),
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

test('M3-R12 editor mount without editing does not arm an unmount write/clear', async () => {
  const store = createMemoryTranscriptHistory();
  const write = rstest.spyOn(store, 'write');
  const clear = rstest.spyOn(store, 'clear');
  const view = await editor(async () => null, { history: { key: 'author:lesson', store } });
  view.unmount();
  expect(write).not.toHaveBeenCalled();
  expect(clear).not.toHaveBeenCalled();
});

for (const legacy of [true, false]) {
  test(`M3-R6 Ukrainian detached unknown-language copy (legacy=${legacy}); Discard continues after refused clear`, async () => {
    const store = createMemoryTranscriptHistory();
    const key = legacy ? 'author:lesson' : 'author:lesson:track:gone';
    const snapshot = {
      key,
      revision: crypto.randomUUID(),
      capturedAt: 2,
      expectedUpdatedAt: undefined,
      lines: [storedLine],
    };
    const next = {
      ...snapshot,
      key: 'author:lesson:lang:fr',
      revision: crypto.randomUUID(),
      capturedAt: 1,
      language: 'fr',
    };
    const clear = rstest.spyOn(store, 'clearIfVersion').mockRejectedValue(new Error('blocked'));
    await editor(async () => null, {
      instanceId: 'track:uk',
      fixedLanguage: true,
      history: { key: 'author:lesson:track:uk', store },
      detachedSnapshots: [
        { instanceId: 'track:uk', legacy, snapshot },
        { instanceId: 'track:uk', legacy: false, snapshot: next },
      ],
    });
    await screen.findByRole('heading', {
      name: legacy
        ? 'Незбережені рядки до появи вкладок (мова невідома)'
        : 'Незбережені рядки з іншої вкладки (мова невідома)',
    });
    fireEvent.click(screen.getByRole('button', { name: 'Не відновлювати' }));
    await screen.findByRole('heading', { name: 'Незбережені рядки з іншої вкладки (мова: fr)' });
    expect(clear).toHaveBeenCalledExactlyOnceWith(key, snapshot.revision);
  });
}

test('MS3b fix 1 / M3-R6 Restore A then B saves only B and offers A again', async () => {
  const store = createMemoryTranscriptHistory();
  const a = {
    key: 'author:lesson:track:gone-a',
    revision: crypto.randomUUID(),
    capturedAt: 2,
    language: 'uk',
    expectedUpdatedAt: undefined,
    lines: [{ ...storedLine, text: 'A' }],
  };
  const b = {
    ...a,
    key: 'author:lesson:track:gone-b',
    revision: crypto.randomUUID(),
    capturedAt: 1,
    lines: [{ ...storedLine, text: 'B' }],
  };
  await store.write(a);
  await store.write(b);
  const save = rstest.fn<TranscriptEditorActionsSave>(async () => null);
  await editor(save, {
    instanceId: 'track:uk',
    fixedLanguage: true,
    history: { key: 'author:lesson:track:uk', store },
    detachedSnapshots: [a, b].map((snapshot) => ({
      instanceId: 'track:uk',
      legacy: false,
      snapshot,
    })),
  });
  fireEvent.click(await screen.findByRole('button', { name: 'Відновити зміни' }));
  fireEvent.click(screen.getByRole('button', { name: 'Відновити зміни' }));
  expect(screen.getByLabelText('Текст репліки 1')).toHaveProperty('value', 'B');
  fireEvent.click(screen.getByRole('button', { name: 'Зберегти транскрипт' }));
  await screen.findByText('Транскрипт збережено.');
  expect(save).toHaveBeenCalledTimes(1);
  expect(save.mock.calls[0]?.[2]).toMatchObject([{ text: 'B' }]);
  expect(await store.read(b.key)).toBeNull();
  expect(await store.read(a.key)).toEqual(a);
  fireEvent.click(screen.getByRole('button', { name: 'Відновити зміни' }));
  expect(screen.getByLabelText('Текст репліки 1')).toHaveProperty('value', 'A');
});

test('MS3b fix2 / M3-R6 locked own offer survives clean rename, blur, unmount and remount', async () => {
  const store = createMemoryTranscriptHistory();
  const key = 'author:lesson:track:uk';
  const snapshot = {
    key,
    revision: crypto.randomUUID(),
    capturedAt: 1,
    language: 'uk',
    expectedUpdatedAt: undefined,
    lines: [storedLine],
  };
  await store.write(snapshot);
  const clear = rstest.spyOn(store, 'clear');
  const props = {
    lessonId,
    instanceId: 'track:uk',
    fixedLanguage: true,
    language: 'uk',
    segments,
    documents: [],
    actions: { save: async () => null },
    history: { key, store },
  };
  const view = render(<TranscriptEditor {...props} />);
  await screen.findByRole('button', { name: 'Відновити зміни' });
  view.rerender(
    <TranscriptEditor
      {...props}
      language="en"
      segments={[
        { id: 'server', startMs: 0, endMs: 1000, documentIds: [], text: 'Server English' },
      ]}
    />,
  );
  expect(screen.getByLabelText('Текст репліки 1')).toHaveProperty('readOnly', true);
  fireEvent.focusOut(screen.getByLabelText('Текст репліки 1'));
  view.unmount();
  expect(clear).not.toHaveBeenCalled();
  expect(await store.read(key)).toEqual(snapshot);
  render(<TranscriptEditor {...props} language="en" />);
  fireEvent.click(await screen.findByRole('button', { name: 'Відновити зміни' }));
  expect(screen.getByLabelText('Текст репліки 1')).toHaveProperty('value', storedLine.text);
});

test('MS3b fix 2 clean server update alone never arms capture on unmount', async () => {
  const store = createMemoryTranscriptHistory();
  const clear = rstest.spyOn(store, 'clear');
  const write = rstest.spyOn(store, 'write');
  const props = {
    lessonId,
    language: 'uk',
    segments,
    documents: [],
    actions: { save: async () => null },
    history: { key: 'author:lesson:track:uk', store },
  };
  const view = render(<TranscriptEditor {...props} />);
  await waitFor(() => expect(store.keys()).resolves.toEqual([]));
  view.rerender(
    <TranscriptEditor
      {...props}
      language="en"
      segments={[
        { id: 'server', startMs: 0, endMs: 1000, documentIds: [], text: 'Server English' },
      ]}
    />,
  );
  view.unmount();
  expect(clear).not.toHaveBeenCalled();
  expect(write).not.toHaveBeenCalled();
});

for (const instanceId of ['lang:uk', 'track:uk']) {
  test(`MS3b fix 5 / M3-R13 legacy restore in ${instanceId} is not a track rename`, async () => {
    await editorWithSnapshot(
      {
        key: 'author:lesson',
        revision: crypto.randomUUID(),
        capturedAt: 1,
        language: 'uk',
        expectedUpdatedAt: undefined,
        lines: [storedLine],
      },
      { instanceId, fixedLanguage: true },
    );
    fireEvent.click(await screen.findByRole('button', { name: 'Відновити зміни' }));
    expect(screen.queryByText('Мова доріжки тепер uk; рядки буде збережено як uk.')).toBeNull();
  });
}

test('MS3b fix 5 a clean track rename followed by editing is not a dirty track rename', async () => {
  const props = {
    lessonId,
    instanceId: 'track:one',
    fixedLanguage: true,
    language: 'uk',
    baseLanguage: 'uk',
    segments,
    documents: [],
    actions: { save: async () => null },
    history: { key: 'author:lesson:track:one', store: createMemoryTranscriptHistory() },
  };
  const view = render(<TranscriptEditor {...props} />);
  // A cached base can remain while the new-language query is in flight.
  view.rerender(<TranscriptEditor {...props} language="en" />);
  fireEvent.change(screen.getByLabelText('Текст репліки 1'), {
    target: { value: 'Edited after rename' },
  });
  expect(screen.queryByText('Мова доріжки тепер en; рядки буде збережено як en.')).toBeNull();
});

const recoveryLockHint = 'Спершу відновіть або відкиньте незбережені зміни';

async function expectRecoveryLocked(save: ReturnType<typeof rstest.fn>) {
  expect(screen.getByText(recoveryLockHint)).toBeTruthy();
  const text = screen.getByLabelText('Текст репліки 1');
  for (const label of ['Текст репліки 1', 'Початок репліки 1, мс', 'Кінець репліки 1, мс']) {
    const field = screen.getByLabelText(label);
    expect(field).toHaveProperty('readOnly', true);
    expect(field).toHaveProperty('disabled', false);
    field.focus();
    expect(document.activeElement).toBe(field);
    fireEvent.change(field, { target: { value: 'Rejected edit' } });
  }
  await userEvent.type(text, 'Rejected typing');
  expect(text).toHaveProperty('value', 'Перша');
  for (const name of ['Додати репліку', 'Видалити репліку 1', 'Зберегти транскрипт']) {
    const button = screen.getByRole('button', { name });
    expect(button.getAttribute('aria-disabled')).toBe('true');
    expect(button).toHaveProperty('disabled', false);
    button.focus();
    expect(document.activeElement).toBe(button);
    fireEvent.click(button);
  }
  const material = screen.getByRole('checkbox', { name: 'Довідка' });
  expect(material.getAttribute('aria-disabled')).toBe('true');
  fireEvent.click(material);
  expect(material).toHaveProperty('checked', false);
  expect(screen.getAllByRole('listitem')).toHaveLength(1);
  expect(save).not.toHaveBeenCalled();
}

for (const resolution of ['restore', 'discard'] as const) {
  for (const stale of [false, true]) {
    test(`MS3b fix2 / M3-R6 own ${stale ? 'stale ' : ''}offer locks author actions; ${resolution} enables new edits surviving remount`, async () => {
      const store = createMemoryTranscriptHistory();
      const key = `author:lesson:track:fix2-${resolution}-${stale}`;
      const snapshot = {
        key,
        revision: crypto.randomUUID(),
        capturedAt: 1,
        language: 'uk',
        expectedUpdatedAt: stale ? 'old' : '2026-09-18T00:00:00Z',
        lines: [storedLine],
      };
      await store.write(snapshot);
      const write = rstest.spyOn(store, 'write');
      const clear = rstest.spyOn(store, 'clear');
      const save = rstest.fn(async () => null);
      const props = { history: { key, store } };
      const view = await editor(save, props);
      await screen.findByRole('button', { name: 'Відновити зміни' });
      await expectRecoveryLocked(save);
      expect(write).not.toHaveBeenCalled();
      expect(clear).not.toHaveBeenCalled();
      expect(await store.read(key)).toEqual(snapshot);
      fireEvent.click(
        screen.getByRole('button', {
          name: resolution === 'restore' ? 'Відновити зміни' : 'Не відновлювати',
        }),
      );
      const text = screen.getByLabelText('Текст репліки 1');
      expect(text).toHaveProperty('readOnly', false);
      expect(text).toHaveProperty('value', resolution === 'restore' ? storedLine.text : 'Перша');
      expect(screen.queryByText(recoveryLockHint)).toBeNull();
      fireEvent.change(text, { target: { value: 'NEW author lines' } });
      fireEvent.blur(text);
      view.unmount();
      expect((await store.read(key))?.lines[0]?.text).toBe('NEW author lines');
      await editor(save, props);
      fireEvent.click(await screen.findByRole('button', { name: 'Відновити зміни' }));
      expect(screen.getByLabelText('Текст репліки 1')).toHaveProperty('value', 'NEW author lines');
    });
  }
}

for (const outcome of ['snapshot', 'none', 'reject'] as const) {
  test(`MS3b fix2 / M3-R12 deferred first read refuses edits and captures until ${outcome}`, async () => {
    const store = createMemoryTranscriptHistory();
    const key = `author:lesson:track:fix2-initial-${outcome}`;
    const snapshot = {
      key,
      revision: crypto.randomUUID(),
      capturedAt: 1,
      expectedUpdatedAt: '2026-09-18T00:00:00Z',
      lines: [storedLine],
    };
    if (outcome === 'snapshot') await store.write(snapshot);
    let resolveRead: (value: typeof snapshot | null) => void = () => {};
    let rejectRead: (reason: Error) => void = () => {};
    const pending = new Promise<typeof snapshot | null>((resolve, reject) => {
      resolveRead = resolve;
      rejectRead = reject;
    });
    const read = rstest.spyOn(store, 'read').mockReturnValueOnce(pending);
    const write = rstest.spyOn(store, 'write');
    const clear = rstest.spyOn(store, 'clear');
    const save = rstest.fn(async () => null);
    const handled = rstest.fn();
    await editor(save, {
      history: { key, store },
      draftOffer: { segments: draftSegments, language: 'uk', updatedAt: 'D1' },
      onDraftOfferHandled: handled,
    });
    await expectRecoveryLocked(save);
    const load = screen.getByRole('button', { name: 'Завантажити чернетку' });
    expect(load.getAttribute('aria-disabled')).toBe('true');
    expect(load).toHaveProperty('disabled', false);
    fireEvent.click(load);
    expect(handled).not.toHaveBeenCalled();
    fireEvent.blur(screen.getByLabelText('Текст репліки 1'));
    expect(write).not.toHaveBeenCalled();
    expect(clear).not.toHaveBeenCalled();
    const focused = screen.getByLabelText('Текст репліки 1');
    focused.focus();
    await act(async () => {
      if (outcome === 'reject') rejectRead(new Error('store refused'));
      else resolveRead(outcome === 'snapshot' ? snapshot : null);
    });
    expect(document.activeElement).toBe(focused);
    expect(read).toHaveBeenCalledTimes(1);
    expect(write).not.toHaveBeenCalled();
    expect(clear).not.toHaveBeenCalled();
    if (outcome === 'snapshot') {
      await expectRecoveryLocked(save);
      expect(await store.read(key)).toEqual(snapshot);
      fireEvent.click(screen.getByRole('button', { name: 'Не відновлювати' }));
    }
    expect(screen.getByLabelText('Текст репліки 1')).toHaveProperty('readOnly', false);
    expect(screen.queryByText(recoveryLockHint)).toBeNull();
    fireEvent.click(screen.getByRole('button', { name: 'Завантажити чернетку' }));
    expect(handled).toHaveBeenCalledTimes(1);
    expect(screen.getByLabelText('Текст репліки 1')).toHaveProperty(
      'value',
      draftSegments[0]?.text,
    );
    fireEvent.change(screen.getByLabelText('Текст репліки 1'), { target: { value: 'After read' } });
    fireEvent.blur(screen.getByLabelText('Текст репліки 1'));
    expect((await store.read(key))?.lines[0]?.text).toBe('After read');
  });
}

test('MS3b fix2 / M3-R6 detached offer keeps editing enabled and captures only the own key', async () => {
  const store = createMemoryTranscriptHistory();
  const key = 'author:lesson:track:fix2-detached-own';
  const snapshot = {
    key: 'author:lesson:track:gone',
    revision: crypto.randomUUID(),
    capturedAt: 1,
    expectedUpdatedAt: undefined,
    lines: [storedLine],
  };
  await store.write(snapshot);
  const write = rstest.spyOn(store, 'write');
  const clear = rstest.spyOn(store, 'clear');
  await editor(async () => null, {
    history: { key, store },
    detachedSnapshots: [{ instanceId: 'track:fix2-detached-own', snapshot, legacy: false }],
  });
  await screen.findByRole('button', { name: 'Відновити зміни' });
  expect(screen.queryByText(recoveryLockHint)).toBeNull();
  const text = screen.getByLabelText('Текст репліки 1');
  expect(text).toHaveProperty('readOnly', false);
  fireEvent.change(text, { target: { value: 'Own new lines' } });
  fireEvent.blur(text);
  expect(write).toHaveBeenCalledTimes(1);
  expect(write.mock.calls[0]?.[0]).toMatchObject({ key, lines: [{ text: 'Own new lines' }] });
  expect(clear).not.toHaveBeenCalled();
  expect(await store.read(snapshot.key)).toEqual(snapshot);
  expect(screen.getByRole('button', { name: 'Відновити зміни' })).toBeTruthy();
});

test('MS3b fix2 / M3-R6 English recovery hint describes the focusable read-only fields', async () => {
  const store = createMemoryTranscriptHistory();
  const key = 'author:lesson:track:fix2-english';
  await store.write({
    key,
    revision: crypto.randomUUID(),
    capturedAt: 1,
    expectedUpdatedAt: undefined,
    lines: [storedLine],
  });
  render(
    <I18nProvider instance={createI18n('en')}>
      <TranscriptEditor
        lessonId={lessonId}
        language="en"
        segments={segments}
        documents={[]}
        actions={{ save: async () => null }}
        history={{ key, store }}
      />
    </I18nProvider>,
  );
  const hint = await screen.findByText('Restore or discard the unsaved changes first');
  const text = screen.getByRole('textbox', { name: 'Line 1 text' });
  expect(text.getAttribute('aria-describedby')?.split(' ')).toContain(hint.id);
  text.focus();
  expect(document.activeElement).toBe(text);
  expect(text).toHaveProperty('readOnly', true);
});

test('MS3b fix2 / M3-R6 Discard without new edits retires only the discarded version across remount', async () => {
  const store = createMemoryTranscriptHistory();
  const key = 'author:lesson:track:fix2-dismiss-old';
  const snapshot = {
    key,
    revision: crypto.randomUUID(),
    capturedAt: 1,
    expectedUpdatedAt: undefined,
    lines: [storedLine],
  };
  await store.write(snapshot);
  const write = rstest.spyOn(store, 'write');
  const props = { history: { key, store } };
  const view = await editor(async () => null, props);
  fireEvent.click(await screen.findByRole('button', { name: 'Не відновлювати' }));
  view.unmount();
  await editor(async () => null, props);
  expect(screen.queryByRole('button', { name: 'Відновити зміни' })).toBeNull();
  expect(screen.getByLabelText('Текст репліки 1')).toHaveProperty('readOnly', false);
  expect(write).not.toHaveBeenCalled();
  expect(await store.read(key)).toBeNull();
});

for (const source of ['own', 'detached'] as const) {
  for (const action of ['Restore changes', 'Discard'] as const) {
    for (const empty of [false, true]) {
      test(`MS3b fix3 recovery focus ${source} keyboard ${action} empty=${empty}`, async () => {
        const store = createMemoryTranscriptHistory();
        const ownKey = `author:focus:${source}:${action}:${empty}`;
        const snapshot = {
          key: source === 'own' ? ownKey : `${ownKey}:track:gone`,
          revision: crypto.randomUUID(),
          capturedAt: 1,
          language: 'uk',
          baseLanguage: 'uk',
          expectedUpdatedAt: '2026-09-18T00:00:00Z',
          lines: empty && action === 'Restore changes' ? [] : [storedLine],
        };
        if (source === 'own') await store.write(snapshot);
        await editor(async () => null, {
          history: { key: ownKey, store },
          segments: empty && action === 'Discard' ? [] : segments,
          detachedSnapshots:
            source === 'detached' ? [{ instanceId: '', legacy: false, snapshot }] : [],
        });
        const button = await screen.findByRole('button', {
          name: action === 'Restore changes' ? 'Відновити зміни' : 'Не відновлювати',
        });
        act(() => button.focus());
        await userEvent.setup().keyboard('{Enter}');
        expect(
          screen.queryByRole('button', {
            name: action === 'Restore changes' ? 'Відновити зміни' : 'Не відновлювати',
          }),
        ).toBeNull();
        expect(document.activeElement).toBe(
          empty
            ? screen.getByRole('button', { name: 'Додати репліку' })
            : screen.getByRole('textbox', { name: 'Текст репліки 1' }),
        );
      });
    }
  }
}

for (const source of ['own', 'detached'] as const) {
  for (const action of ['Restore changes', 'Discard'] as const) {
    test(`MS3b fix3 recovery focus ${source} keyboard ${action} next detached offer`, async () => {
      const store = createMemoryTranscriptHistory();
      const ownKey = `author:focus-next:${source}:${action}`;
      const snapshot = {
        key: source === 'own' ? ownKey : `${ownKey}:track:gone`,
        revision: crypto.randomUUID(),
        capturedAt: 2,
        language: 'uk',
        expectedUpdatedAt: '2026-09-18T00:00:00Z',
        lines: [storedLine],
      };
      if (source === 'own') await store.write(snapshot);
      const next = {
        ...snapshot,
        key: 'author:lesson:track:next',
        revision: crypto.randomUUID(),
        capturedAt: 1,
      };
      await editor(async () => null, {
        history: { key: ownKey, store },
        detachedSnapshots: [
          ...(source === 'detached' ? [{ instanceId: '', legacy: false, snapshot }] : []),
          { instanceId: '', legacy: false, snapshot: next },
        ],
      });
      const button = await screen.findByRole('button', {
        name: action === 'Restore changes' ? 'Відновити зміни' : 'Не відновлювати',
      });
      act(() => button.focus());
      await userEvent.setup().keyboard('{Enter}');
      expect(document.activeElement).toBe(screen.getByRole('button', { name: 'Відновити зміни' }));
    });
  }
}

for (const newerSource of [false, true]) {
  test(`MS3b fix4 detached Restore capture reload own Restore Save two visits newerC=${newerSource}`, async () => {
    const store = createMemoryTranscriptHistory();
    const key = `author:fix4-${newerSource}:track:own`;
    const source = {
      key: `author:fix4-${newerSource}:lang:uk`,
      revision: crypto.randomUUID(),
      capturedAt: 1,
      language: 'uk',
      baseLanguage: 'uk',
      expectedUpdatedAt: '2026-09-18T00:00:00Z',
      lines: [{ ...storedLine, text: 'Accepted A' }],
    };
    await store.write(source);
    const props = async () => {
      const detached = await store.read(source.key);
      return {
        history: { key, store },
        detachedSnapshots:
          detached === null ? [] : [{ instanceId: '', legacy: false, snapshot: detached }],
      };
    };
    const save = rstest.fn(
      async (
        _language: string,
        _token: string | undefined,
        saved: readonly TranscriptSegment[],
      ) => {
        expect(saved[0]?.text).toBe('Accepted A');
        return null;
      },
    );
    const first = await editor(save, await props());
    fireEvent.click(await screen.findByRole('button', { name: 'Відновити зміни' }));
    first.unmount();
    expect((await store.read(key))?.lines[0]?.text).toBe('Accepted A');
    const capturedSource = (await store.read(key))?.acceptedSource;
    const second = await editor(save, await props());
    fireEvent.click(await screen.findByRole('button', { name: 'Відновити зміни' }));
    const newer = {
      ...source,
      revision: crypto.randomUUID(),
      capturedAt: 2,
      lines: [{ ...storedLine, text: 'Newer C' }],
    };
    if (newerSource) await store.write(newer);
    fireEvent.click(screen.getByRole('button', { name: 'Зберегти транскрипт' }));
    await screen.findByText('Транскрипт збережено.');
    expect(save).toHaveBeenCalledTimes(1);
    await waitFor(async () =>
      expect(await store.read(source.key)).toEqual(newerSource ? newer : null),
    );
    expect(await store.read(key)).toBeNull();
    second.unmount();
    for (let visit = 0; visit < 2; visit++) {
      const view = await editor(save, {
        ...(await props()),
        expectedUpdatedAt: 'T2',
        segments: [{ id: 'saved-A', startMs: 0, endMs: 1000, text: 'Accepted A', documentIds: [] }],
      });
      expect(screen.queryByRole('button', { name: 'Відновити зміни' }) !== null).toBe(newerSource);
      expect(await store.read(source.key)).toEqual(newerSource ? newer : null);
      view.unmount();
    }
    expect(capturedSource).toEqual([{ key: source.key, revision: source.revision }]);
  });
}

for (const newer of [false, true]) {
  test(`MS3b fix5 detached Discard two editors newerC=${newer}`, async () => {
    const store = createMemoryTranscriptHistory();
    const oldKey = `author:fix5-discard-${newer}:track:old`;
    const save = rstest.fn<TranscriptEditorActionsSave>(async () => null);
    const old = await editor(save, {
      instanceId: 'old',
      history: { key: oldKey, store },
    });
    const oldFields = within(old.container);
    fireEvent.change(oldFields.getByLabelText('Текст репліки 1'), { target: { value: 'A' } });
    fireEvent.focusOut(oldFields.getByLabelText('Текст репліки 1'));
    const offered = await store.read(oldKey);
    if (offered === null) throw new Error('old editor must capture A');
    const next = { ...offered, key: `${oldKey}:next`, language: 'fr', capturedAt: 1 };
    await store.write(next);
    const receiver = await editor(save, {
      instanceId: 'new',
      history: { key: `${oldKey}:replacement`, store },
      detachedSnapshots: [offered, next].map((snapshot) => ({
        instanceId: 'new',
        legacy: false,
        snapshot,
      })),
    });
    let capturedC = null;
    if (newer) {
      // Distinct versions, without depending on how quickly the host clock ticks.
      const clock = rstest.spyOn(Date, 'now').mockReturnValue(offered.capturedAt + 1000);
      fireEvent.change(oldFields.getByLabelText('Текст репліки 1'), { target: { value: 'C' } });
      fireEvent.focusOut(oldFields.getByLabelText('Текст репліки 1'));
      clock.mockRestore();
      capturedC = await store.read(oldKey);
      expect(capturedC?.lines[0]?.text).toBe('C');
    }
    old.unmount();
    const fields = within(receiver.container);
    const discard = fields.getByRole('button', { name: 'Не відновлювати' });
    act(() => discard.focus());
    await userEvent.setup().keyboard('{Enter}');
    await waitFor(async () => expect(await store.read(oldKey)).toEqual(newer ? capturedC : null));
    const restore = fields.getByRole('button', { name: 'Відновити зміни' });
    expect(document.activeElement).toBe(restore);
    expect(
      fields.getByRole('heading', {
        name: newer
          ? 'Незбережені рядки з іншої вкладки (мова: uk)'
          : 'Незбережені рядки з іншої вкладки (мова: fr)',
      }),
    ).toBeTruthy();
    fireEvent.click(restore);
    expect(fields.getByLabelText('Текст репліки 1')).toHaveProperty('value', newer ? 'C' : 'A');
    expect(save).not.toHaveBeenCalled();
    if (newer) expect(await store.read(oldKey)).toEqual(capturedC);
  });
}

test('MS3b fix6 two windows own Discard through clean blur and remount preserves C', async () => {
  const store = createMemoryTranscriptHistory();
  const key = 'author:fix6-own-discard';
  const save = rstest.fn<TranscriptEditorActionsSave>(async () => null);
  const b = await editor(save, { instanceId: 'B', history: { key, store } });
  const bFields = within(b.container);
  fireEvent.change(bFields.getByLabelText('Текст репліки 1'), { target: { value: 'A' } });
  fireEvent.focusOut(bFields.getByLabelText('Текст репліки 1'));
  const a = await store.read(key);
  if (a === null) throw new Error('B must capture A');
  const receiver = await editor(save, { instanceId: 'A', history: { key, store } });
  const fields = within(receiver.container);
  const clock = rstest.spyOn(Date, 'now').mockReturnValue(a.capturedAt + 1000);
  fireEvent.change(bFields.getByLabelText('Текст репліки 1'), { target: { value: 'C' } });
  fireEvent.focusOut(bFields.getByLabelText('Текст репліки 1'));
  clock.mockRestore();
  const c = await store.read(key);
  b.unmount();
  fireEvent.click(fields.getByRole('button', { name: 'Не відновлювати' }));
  await act(async () => {});
  fireEvent.focusOut(fields.getByLabelText('Текст репліки 1'));
  await act(async () => {});
  expect(await store.read(key)).toEqual(c);
  expect(fields.getByRole('button', { name: 'Відновити зміни' })).toBeTruthy();
  receiver.unmount();
  await editor(save, { history: { key, store } });
  fireEvent.click(await screen.findByRole('button', { name: 'Відновити зміни' }));
  expect(screen.getByLabelText('Текст репліки 1')).toHaveProperty('value', 'C');
  expect(save).not.toHaveBeenCalled();
});

for (const newerAncestor of [false, true]) {
  test(`MS3b fix5 ancestry six-step legacy A to B to C Save revisit newerAncestor=${newerAncestor}`, async () => {
    const store = createMemoryTranscriptHistory();
    const aKey = `author:fix5-ancestry-${newerAncestor}`;
    const bKey = `${aKey}:track:first`;
    const cKey = `${aKey}:track:replacement`;
    const save = rstest.fn<TranscriptEditorActionsSave>(async () => null);
    // Type legacy A before any video is Ready, then promote the first video to B.
    const legacy = await editor(save, { history: { key: aKey, store } });
    fireEvent.change(screen.getByLabelText('Текст репліки 1'), { target: { value: 'Legacy A' } });
    legacy.unmount();
    const a = await store.read(aKey);
    if (a === null) throw new Error('legacy A must be captured');
    const first = await editor(save, {
      history: { key: bKey, store },
      detachedSnapshots: [{ instanceId: '', legacy: true, snapshot: a }],
    });
    fireEvent.click(screen.getByRole('button', { name: 'Відновити зміни' }));
    // Replacing the first video detaches B, including its ownership of A.
    first.unmount();
    const b = await store.read(bKey);
    if (b === null) throw new Error('restored B must be captured');
    const second = await editor(save, {
      history: { key: cKey, store },
      detachedSnapshots: [b, a].map((snapshot) => ({ instanceId: '', legacy: false, snapshot })),
    });
    fireEvent.click(screen.getByRole('button', { name: 'Відновити зміни' }));
    const newer = {
      ...a,
      revision: crypto.randomUUID(),
      capturedAt: a.capturedAt + 1,
      lines: [{ ...storedLine, text: 'Newer ancestor' }],
    };
    if (newerAncestor) await store.write(newer);
    fireEvent.click(screen.getByRole('button', { name: 'Зберегти транскрипт' }));
    await screen.findByText('Транскрипт збережено.');
    expect(save).toHaveBeenCalledTimes(1);
    expect(save.mock.calls[0]?.[2]).toMatchObject([{ text: 'Legacy A' }]);
    await waitFor(async () => {
      expect(await store.read(aKey)).toEqual(newerAncestor ? newer : null);
      expect(await store.read(bKey)).toBeNull();
      expect(await store.read(cKey)).toBeNull();
    });
    second.unmount();
    for (let visit = 0; visit < 2; visit++) {
      const remaining = await store.read(aKey);
      const view = await editor(save, {
        history: { key: cKey, store },
        expectedUpdatedAt: 'T2',
        segments: [
          {
            ...segments[0],
            id: 'saved',
            startMs: 0,
            endMs: 1000,
            text: 'Legacy A',
            documentIds: [],
          },
        ],
        detachedSnapshots:
          remaining === null ? [] : [{ instanceId: '', legacy: true, snapshot: remaining }],
      });
      expect(screen.queryByRole('button', { name: 'Відновити зміни' }) !== null).toBe(
        newerAncestor,
      );
      if (newerAncestor) {
        fireEvent.click(screen.getByRole('button', { name: 'Відновити зміни' }));
        expect(screen.getByLabelText('Текст репліки 1')).toHaveProperty('value', 'Newer ancestor');
      }
      view.unmount();
    }
  });
}

for (const mode of ['deduplicate', 'cap'] as const) {
  test(`MS3b fix5 ancestry ${mode} survives capture and own Restore Save`, async () => {
    const store = createMemoryTranscriptHistory();
    const key = `author:fix5-bound-${mode}:track:own`;
    const source = {
      key: `${key}:source`,
      revision: crypto.randomUUID(),
      capturedAt: 1,
      expectedUpdatedAt: undefined,
      lines: [storedLine],
    };
    const chain =
      mode === 'deduplicate'
        ? [
            { key: source.key, revision: source.revision },
            { key: source.key, revision: source.revision },
          ]
        : Array.from({ length: 8 }, (_, index) => ({
            key: `${key}:ancestor-${index}`,
            revision: crypto.randomUUID(),
          }));
    for (const entry of chain) await store.write({ ...source, ...entry });
    const snapshot = { ...source, acceptedSource: chain };
    await store.write(snapshot);
    const first = await editor(async () => null, {
      history: { key, store },
      detachedSnapshots: [{ instanceId: '', legacy: false, snapshot }],
    });
    fireEvent.click(screen.getByRole('button', { name: 'Відновити зміни' }));
    first.unmount();
    const expected = [
      { key: source.key, revision: source.revision },
      ...(mode === 'cap' ? chain.slice(0, 7) : []),
    ];
    expect((await store.read(key))?.acceptedSource).toEqual(expected);
    await editor(async () => null, { history: { key, store } });
    fireEvent.click(screen.getByRole('button', { name: 'Відновити зміни' }));
    fireEvent.click(screen.getByRole('button', { name: 'Зберегти транскрипт' }));
    await screen.findByText('Транскрипт збережено.');
    await waitFor(async () => {
      for (const entry of expected) expect(await store.read(entry.key)).toBeNull();
    });
    if (mode === 'cap') {
      const overflow = chain.at(-1);
      if (overflow === undefined) throw new Error('cap fixture requires overflow');
      expect(await store.read(overflow.key)).not.toBeNull();
    }
  });
}

test('MS3b fix5 Discard mismatch with refused refresh leaves the key unresolved for retry', async () => {
  const store = createMemoryTranscriptHistory();
  const source = {
    key: 'author:fix5-refused-refresh:old',
    revision: crypto.randomUUID(),
    capturedAt: 1,
    language: 'uk',
    expectedUpdatedAt: undefined,
    lines: [storedLine],
  };
  await store.write(source);
  await editor(async () => null, {
    history: { key: `${source.key}:own`, store },
    detachedSnapshots: [{ instanceId: '', legacy: false, snapshot: source }],
  });
  const c = {
    ...source,
    revision: crypto.randomUUID(),
    capturedAt: 2,
    lines: [{ ...storedLine, text: 'C after read refusal' }],
  };
  await store.write(c);
  const read = rstest.spyOn(store, 'read').mockRejectedValueOnce(new Error('read refused'));
  await act(async () => fireEvent.click(screen.getByRole('button', { name: 'Не відновлювати' })));
  expect(read).toHaveBeenCalledExactlyOnceWith(source.key);
  expect(await store.read(source.key)).toEqual(c);
  fireEvent.click(screen.getByRole('button', { name: 'Не відновлювати' }));
  await act(async () => {});
  fireEvent.click(screen.getByRole('button', { name: 'Відновити зміни' }));
  expect(screen.getByLabelText('Текст репліки 1')).toHaveProperty('value', 'C after read refusal');
});

for (const newer of [false, true]) {
  test(`MS3b fix6 deferred detached Discard preserves moved focus newer=${newer}`, async () => {
    const store = createMemoryTranscriptHistory();
    const source = {
      key: `author:fix6-focus:${newer}:old`,
      revision: crypto.randomUUID(),
      capturedAt: 1,
      language: 'uk',
      expectedUpdatedAt: undefined,
      lines: [storedLine],
    };
    await store.write(source);
    await editor(async () => null, {
      history: { key: `${source.key}:own`, store },
      detachedSnapshots: [{ instanceId: '', legacy: false, snapshot: source }],
    });
    const other = render(<button type="button">Elsewhere</button>);
    const clear = store.clearIfVersion;
    let complete: () => void = () => {};
    const pending = new Promise<void>((resolve) => {
      complete = resolve;
    });
    rstest.spyOn(store, 'clearIfVersion').mockImplementation(async (key, version) => {
      await pending;
      return clear(key, version);
    });
    const discard = screen.getByRole('button', { name: 'Не відновлювати' });
    act(() => discard.focus());
    await userEvent.setup().keyboard('{Enter}');
    const elsewhere = within(other.container).getByRole('button', { name: 'Elsewhere' });
    act(() => elsewhere.focus());
    const replacement = { ...source, revision: crypto.randomUUID(), capturedAt: 2 };
    if (newer) await store.write(replacement);
    await act(async () => complete());
    expect(document.activeElement).toBe(elsewhere);
    expect(await store.read(source.key)).toEqual(newer ? replacement : null);
  });
}

for (const path of ['own Save', 'accepted-source Save'] as const) {
  for (const version of ['same', 'newer'] as const) {
    test(`MS3b fix6 deletion matrix ${path} / ${version}`, async () => {
      const store = createMemoryTranscriptHistory();
      const ownKey = `author:fix6-matrix:${path}:${version}:own`;
      const source = {
        key: `${ownKey}:detached`,
        revision: crypto.randomUUID(),
        capturedAt: 1,
        language: 'uk',
        expectedUpdatedAt: undefined,
        lines: [storedLine],
      };
      if (path === 'accepted-source Save') await store.write(source);
      let finishSave: (value: null) => void = () => {};
      const pending = new Promise<null>((resolve) => {
        finishSave = resolve;
      });
      const save = rstest.fn<TranscriptEditorActionsSave>(() => pending);
      const view = await editor(save, {
        history: { key: ownKey, store },
        detachedSnapshots:
          path === 'accepted-source Save'
            ? [{ instanceId: '', legacy: false, snapshot: source }]
            : [],
      });
      if (path === 'accepted-source Save')
        fireEvent.click(screen.getByRole('button', { name: 'Відновити зміни' }));
      else fireEvent.change(screen.getByLabelText('Текст репліки 1'), { target: { value: 'A' } });
      fireEvent.focusOut(screen.getByLabelText('Текст репліки 1'));
      const key = path === 'own Save' ? ownKey : source.key;
      const owned = await store.read(key);
      if (owned === null) throw new Error('a saved source must exist');
      fireEvent.click(screen.getByRole('button', { name: 'Зберегти транскрипт' }));
      expect(save).toHaveBeenCalledTimes(1);
      const c = {
        ...owned,
        revision: crypto.randomUUID(),
        capturedAt: owned.capturedAt + 1000,
        lines: [{ ...storedLine, text: 'C during Save' }],
      };
      if (version === 'newer') await store.write(c);
      await act(async () => finishSave(null));
      await waitFor(async () =>
        expect(await store.read(key)).toEqual(version === 'same' ? null : c),
      );
      view.unmount();
      const detached = await store.read(source.key);
      await editor(save, {
        history: { key: ownKey, store },
        detachedSnapshots:
          detached === null ? [] : [{ instanceId: '', legacy: false, snapshot: detached }],
      });
      expect(screen.queryByRole('button', { name: 'Відновити зміни' }) !== null).toBe(
        version === 'newer',
      );
      expect(save).toHaveBeenCalledTimes(1);
    });
  }
}

test('MS3b fix6 failed Save preserves a newer capture without recapturing old lines', async () => {
  const store = createMemoryTranscriptHistory();
  const key = 'author:fix6-failed-save';
  let finish: (value: TranscriptErrorCode | null) => void = () => {};
  const pending = new Promise<TranscriptErrorCode | null>((resolve) => {
    finish = resolve;
  });
  await editor(() => pending, { history: { key, store } });
  fireEvent.change(screen.getByLabelText('Текст репліки 1'), { target: { value: 'A' } });
  fireEvent.click(screen.getByRole('button', { name: 'Зберегти транскрипт' }));
  const a = await store.read(key);
  if (a === null) throw new Error('pending lines must be captured before Save');
  const c = {
    ...a,
    revision: crypto.randomUUID(),
    capturedAt: a.capturedAt + 1000,
    lines: [{ ...storedLine, text: 'C after Save began' }],
  };
  await store.write(c);
  const write = rstest.spyOn(store, 'write');
  await act(async () => finish('unavailable'));
  expect(await store.read(key)).toEqual(c);
  expect(write).not.toHaveBeenCalled();
  expect(screen.getByLabelText('Текст репліки 1')).toHaveProperty('value', 'A');
});
