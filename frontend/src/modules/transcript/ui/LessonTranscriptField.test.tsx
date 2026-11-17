import { expect, rstest, test } from '@rstest/core';
import { act, fireEvent, render, screen } from '@testing-library/react';
import { createMemoryTranscriptHistory } from '../adapters/transcript-history-memory';
import type { LessonTranscript } from '../domain/transcript';
import { LessonTranscriptField } from './LessonTranscriptField';

const data: LessonTranscript = {
  lessonId: 'lesson',
  language: 'en',
  updatedAt: 'T1',
  stale: false,
  segments: [{ id: 'line', startMs: 0, endMs: 1000, text: 'Server', documentIds: [] }],
  documents: [],
};

test('GAP first transcript load then renamed-language load retains the dirty editor instance', async () => {
  const store = createMemoryTranscriptHistory();
  let loading = true;
  const useLessonTranscript = () => ({ data: loading ? undefined : data, isLoading: loading });
  const props = {
    lessonId: 'lesson',
    instanceId: 'track:one',
    fixedLanguage: 'en',
    useLessonTranscript,
    documents: [],
    history: { store, authorId: 'author' },
    library: { add: async () => null, remove: async () => null },
    save: async () => null,
  };
  const f = render(<LessonTranscriptField {...props} />);
  expect(screen.queryByLabelText('Текст репліки 1')).toBeNull();
  loading = false;
  f.rerender(<LessonTranscriptField {...props} />);
  await act(async () => {});
  const text = screen.getByLabelText('Текст репліки 1');
  fireEvent.change(text, { target: { value: 'Unsaved' } });
  loading = true;
  f.rerender(<LessonTranscriptField {...props} fixedLanguage="es" />);
  expect(screen.getByLabelText('Текст репліки 1')).toBe(text);
  expect(screen.getByDisplayValue('Unsaved')).toBe(text);
});

test('GAP fixed tab language keeps the returned transcript token attached to its real base language', async () => {
  const store = createMemoryTranscriptHistory();
  const editorSave = rstest.fn(async () => null);
  render(
    <LessonTranscriptField
      lessonId="lesson"
      instanceId="track:one"
      fixedLanguage="es"
      useLessonTranscript={() => ({ data, isLoading: false })}
      documents={[]}
      history={{ store, authorId: 'author' }}
      library={{ add: async () => null, remove: async () => null }}
      save={async () => null}
      editorSave={editorSave}
    />,
  );
  await act(async () => {});
  const text = screen.getByLabelText('Текст репліки 1');
  fireEvent.change(text, { target: { value: 'Spanish edit' } });
  fireEvent.focusOut(text);
  await act(async () => {});
  expect(await store.read('author:lesson:track:one')).toMatchObject({
    language: 'es',
    baseLanguage: 'en',
    expectedUpdatedAt: 'T1',
  });
  fireEvent.click(screen.getByRole('button', { name: 'Зберегти транскрипт' }));
  await act(async () => {});
  expect(editorSave).toHaveBeenCalledWith('es', 'T1', expect.any(Array), 'en');
});
