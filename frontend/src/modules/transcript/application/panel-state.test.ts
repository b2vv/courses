import { expect, test } from '@rstest/core';
import type { LessonTranscript } from '../domain/transcript';
import { transcriptPanelState } from './panel-state';
const transcript: LessonTranscript = {
  lessonId: 'lesson',
  language: 'EN',
  updatedAt: null,
  stale: false,
  documents: [],
  segments: [{ id: 'segment', startMs: 0, endMs: 1000, text: 'English', documentIds: [] }],
};
const onRetry = () => {};
const input = { language: 'en', data: transcript, isPending: false, isError: false, onRetry };

test.each([
  ['in flight', { isPending: true }, 'loading'],
  ['matching language', {}, 'transcript'],
  ['null language', { data: { ...transcript, language: null } }, 'none-for-track'],
  ['empty segments', { data: { ...transcript, segments: [] } }, 'none-for-track'],
  ['foreign language', { data: { ...transcript, language: 'de' } }, 'none-for-track'],
  ['no data', { data: undefined }, 'none-for-track'],
  ['error', { isError: true }, 'error'],
  ['primary transcript', { language: undefined }, 'transcript'],
  [
    'primary absent',
    { language: undefined, data: { ...transcript, language: null, segments: [] } },
    'none',
  ],
] as const)('M5 %s', (_name, overrides, kind) => {
  const state = transcriptPanelState({ ...input, ...overrides });
  expect(state.kind).toBe(kind);
  if (state.kind === 'error') expect(state.onRetry).toBe(onRetry);
});
