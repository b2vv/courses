import { expect, rstest, test } from '@rstest/core';
import { act, fireEvent, render, screen } from '@testing-library/react';
import type { ComponentProps } from 'react';
import * as actualTranscript from '../../transcript' with { rstest: 'importActual' };
import type { LessonTranscriptField } from '../../transcript';
import { createMemoryTranscriptHistory } from '../../transcript';
import { createI18n, I18nProvider } from '../../../app/providers/i18n';
import { LessonLanguageTabs } from './LessonLanguageTabs';
import type { AudioTrack, CourseEditorVideo } from '../../media';

const probe = rstest.hoisted(() => ({ completion: undefined as Promise<unknown> | undefined }));
// Observe the Save promise at the existing editor callback boundary: a removed editor
// has no DOM left in which to display cancellation. Dialogs and their lifecycle stay real.
rstest.mock('../../transcript', () => {
  return {
    ...actualTranscript,
    LessonTranscriptField: (props: ComponentProps<typeof LessonTranscriptField>) => (
      <button
        type="button"
        onClick={() => {
          probe.completion = props.editorSave?.(props.fixedLanguage ?? 'en', 'SOURCE', [], 'en');
        }}
      >
        Probe Save
      </button>
    ),
  };
});

for (const mode of ['und', 'replacement'] as const) {
  test(`unmount cancels pending ${mode} Save promise without a write`, async () => {
    probe.completion = new Promise<unknown>(() => {});
    const track: AudioTrack = {
      trackId: 'one',
      videoId: 'video',
      kind: 'embedded',
      language: mode === 'und' ? 'und' : 'fr',
      label: 'Audio',
      isDefault: true,
      multilingual: false,
      state: 'ready',
      updatedAt: 'T1',
    };
    const writes: string[] = [];
    const video: CourseEditorVideo = {
      rows: [
        {
          videoId: 'video',
          lessonId: 'lesson',
          state: 'ready',
          generation: 1,
          originalName: 'video',
          declaredBytes: 1,
          receivedBytes: 1,
          updatedAt: 'T1',
        },
      ],
      isLoading: false,
      actions: {
        upload: async () => null,
        resume: async () => null,
        retry: async () => null,
        cancel: async () => null,
      },
      audio: {
        useTracks: () => ({ data: [track], isLoading: false }),
        actions: {
          upload: async () => null,
          resume: async () => null,
          cancel: async () => null,
          rename: async () => {
            writes.push('PATCH');
            return null;
          },
          setDefault: async () => null,
          remove: async () => null,
        },
      },
    };
    const inventory =
      mode === 'und'
        ? []
        : [{ language: 'fr', updatedAt: 'TARGET', stale: false, isPrimary: false }];
    const view = render(
      <I18nProvider instance={createI18n('en')}>
        <LessonLanguageTabs
          lessonId="lesson"
          video={video}
          onTranscriptKnown={() => {}}
          languages={{
            useInventory: () => ({ data: inventory }),
            readInventory: async () => inventory,
            readTracks: async () => [track],
          }}
          transcripts={{
            useLessonTranscript: () => ({ data: undefined, isLoading: false }),
            documents: [],
            history: { store: createMemoryTranscriptHistory(), authorId: 'author' },
            library: { add: async () => null, remove: async () => null },
            save: async () => {
              writes.push('PUT');
              return null;
            },
          }}
        />
      </I18nProvider>,
    );
    fireEvent.click(await screen.findByRole('button', { name: 'Probe Save' }));
    await screen.findByRole('dialog');
    let result: unknown = 'pending';
    void probe.completion?.then((value) => {
      result = value;
    });
    await act(async () => {
      view.unmount();
    });
    expect(result).toBe(mode === 'und' ? 'language_cancelled' : 'replacement_cancelled');
    expect(writes).toEqual([]);
  });
}
