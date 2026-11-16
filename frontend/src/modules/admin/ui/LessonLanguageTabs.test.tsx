import { expect, test } from '@rstest/core';
import { fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { createI18n, I18nProvider } from '../../../app/providers/i18n';
import type { AudioTrack, CourseEditorVideo } from '../../media';
import { createMemoryTranscriptHistory } from '../../transcript';
import type {
  CourseEditorDrafts,
  CourseEditorTranscripts,
  LessonDraft,
  LessonTranscript,
  TranscriptInventoryItem,
} from '../../transcript';
import { LessonLanguageTabs, lessonLanguageTabs } from './LessonLanguageTabs';
import type { LessonLanguageServices } from './LessonLanguageTabs';

const lessonId = 'lesson';
const segment = { id: 'source-id', startMs: 0, endMs: 1000, text: 'Original', documentIds: [] };
function track(trackId = 'one', language = 'en', extra: Partial<AudioTrack> = {}): AudioTrack {
  return {
    trackId,
    label: language,
    language,
    videoId: 'video',
    kind: 'embedded',
    isDefault: true,
    multilingual: false,
    state: 'ready',
    updatedAt: 'T1',
    ...extra,
  };
}
function item(language: string, updatedAt = 'T1'): TranscriptInventoryItem {
  return { language, updatedAt, stale: false, isPrimary: false };
}
function transcript(language: string, text = 'Original', updatedAt = 'T1'): LessonTranscript {
  return {
    lessonId,
    language,
    updatedAt,
    stale: false,
    segments: [{ ...segment, text }],
    documents: [],
  };
}

function fixture() {
  const state = {
    tracks: [track()],
    inventory: [item('en')],
    freshTracks: undefined as readonly AudioTrack[] | undefined,
    freshInventory: undefined as readonly TranscriptInventoryItem[] | undefined,
    tracksFailure: false,
    tracksLoading: false,
    inventoryQueryFailed: false,
    inventoryRefreshFailed: false,
    inventoryFailure: false,
    bases: new Map<string, LessonTranscript>([['en', transcript('en')]]),
    puts: [] as Parameters<CourseEditorTranscripts['save']>[],
    reads: [] as string[],
    known: [] as boolean[],
    regenerate: [] as string[],
    putError: null as Awaited<ReturnType<CourseEditorTranscripts['save']>>,
    draft: undefined as LessonDraft | undefined,
  };
  const languages: LessonLanguageServices = {
    useInventory: () =>
      state.inventoryQueryFailed
        ? { data: undefined, isError: true }
        : { data: state.inventory, isError: state.inventoryRefreshFailed },
    readInventory: async () => {
      state.reads.push('inventory');
      if (state.inventoryFailure) throw new Error('timeout');
      return state.freshInventory ?? state.inventory;
    },
    readTracks: async (id) => {
      state.reads.push(id);
      if (state.tracksFailure) throw new Error('503');
      return state.freshTracks ?? state.tracks;
    },
  };
  const historyStore = createMemoryTranscriptHistory();
  const transcripts: CourseEditorTranscripts = {
    useLessonTranscript: (_id, language) => ({
      data: state.bases.get(language ?? 'en') ?? {
        lessonId,
        language: null,
        updatedAt: null,
        stale: false,
        segments: [],
        documents: [],
      },
      isLoading: false,
    }),
    documents: [],
    history: { store: historyStore, authorId: 'author' },
    library: { add: async () => null, remove: async () => null },
    save: async (...args) => {
      state.puts.push(args);
      return state.putError;
    },
  };
  const drafts: CourseEditorDrafts = {
    useLessonDraft: (_id, trackId) => ({
      data: state.draft?.trackId === trackId ? state.draft : undefined,
      isLoading: false,
    }),
    actions: {
      regenerate: async (_id, trackId) => {
        state.regenerate.push(trackId);
        return null;
      },
    },
  };
  const video: CourseEditorVideo = {
    rows: [
      {
        videoId: 'video',
        lessonId,
        state: 'ready',
        generation: 1,
        originalName: 'a.mp4',
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
      useTracks: () =>
        state.tracksLoading
          ? { data: undefined, isLoading: true }
          : { data: state.tracks, isLoading: false },
      actions: {
        upload: async () => null,
        resume: async () => null,
        cancel: async () => null,
        rename: async () => null,
        setDefault: async () => null,
        remove: async () => null,
      },
    },
  };
  const node = (ready = true) => (
    <I18nProvider instance={createI18n('en')}>
      <form data-testid="course-form">
        <LessonLanguageTabs
          lessonId={lessonId}
          video={{ ...video, rows: ready ? video.rows : [] }}
          drafts={drafts}
          transcripts={transcripts}
          languages={languages}
          onTranscriptKnown={(_id, known) => {
            state.known.push(known);
          }}
        />
      </form>
    </I18nProvider>
  );
  return { state, node, historyStore, drafts };
}
const save = () => fireEvent.click(screen.getByRole('button', { name: 'Save the transcript' }));
const edit = () =>
  fireEvent.change(screen.getByRole('textbox', { name: 'Line 1 text' }), {
    target: { value: 'My edit' },
  });

// The browser receives the same rule from main.css; install it for accessibility queries here.
function mount(f: ReturnType<typeof fixture>, ready = true) {
  const view = render(
    <>
      <style>{'.lesson-language-panel[inert] { display: none; }'}</style>
      {f.node(ready)}
    </>,
  );
  return {
    ...view,
    refresh: () =>
      view.rerender(
        <>
          <style>{'.lesson-language-panel[inert] { display: none; }'}</style>
          {f.node(ready)}
        </>,
      ),
  };
}

test('tab order is default, embedded, language/id, then orphan; non-ready tracks are excluded', () => {
  const tabs = lessonLanguageTabs(
    [
      track('z', 'uk', { isDefault: false, kind: 'uploaded' }),
      track('b', 'en', { isDefault: false, kind: 'uploaded' }),
      track('default', 'pl'),
      track('embedded', 'und', { isDefault: false }),
      track('a', 'en', { isDefault: false, kind: 'uploaded' }),
      track('pending', 'fr', { state: 'processing' }),
    ],
    [item('en'), item('de')],
  );
  expect(tabs.map((tab) => tab.id)).toEqual([
    'track:default',
    'track:embedded',
    'track:a',
    'track:b',
    'track:z',
    'lang:de',
  ]);
});

test('no Ready asset keeps the original single editor ids, roles and copy', () => {
  const f = fixture();
  const { container } = mount(f, false);
  expect(screen.queryByRole('tablist')).toBeNull();
  expect(screen.getByRole('heading', { name: 'Lesson transcript' }).id).toBe(
    'transcript-editor-lesson',
  );
  expect(container.querySelectorAll('form')).toHaveLength(1);
});

test('tabs wait for the tracks, so a track language is never first shown as an orphan', () => {
  const f = fixture();
  f.state.tracks = [track('default', 'und')];
  f.state.inventory = [item('uk')];
  f.state.tracksLoading = true;
  const view = mount(f);
  expect(screen.queryByRole('tablist')).toBeNull();
  expect(screen.queryByRole('heading', { name: 'Lesson transcript' })).toBeNull();
  f.state.tracksLoading = false;
  view.refresh();
  const tabs = screen.getAllByRole('tab');
  expect(tabs.map((tab) => tab.getAttribute('aria-selected'))).toEqual(['true', 'false']);
  expect(tabs[0]?.textContent).toContain('Unknown language');
});

test('a failed inventory read shows an alert instead of an empty tab set', () => {
  const f = fixture();
  f.state.inventoryQueryFailed = true;
  mount(f);
  expect(screen.queryByRole('tablist')).toBeNull();
  expect(screen.getByRole('alert').textContent).toContain(
    'Could not load the transcript languages',
  );
});

test('mounted editors have unique ids and inactive inert panels, with names targeting the active editor', () => {
  const f = fixture();
  f.state.tracks.push(track('two', 'de', { isDefault: false, multilingual: true }));
  f.state.inventory.push(item('fr'));
  const { container } = mount(f);
  const ids = [...container.querySelectorAll('[id]')].map((node) => node.id);
  expect(new Set(ids).size).toBe(ids.length);
  expect(container.querySelectorAll('.lesson-language-panel[inert]')).toHaveLength(2);
  expect(screen.getAllByRole('region', { name: 'Lesson transcript' })).toHaveLength(1);
  expect(screen.getByRole('tab', { name: /German.*Multiple languages/ })).toBeTruthy();
  fireEvent.click(screen.getByRole('tab', { name: /German/ }));
  const panel = screen.getByRole('tabpanel');
  const heading = within(panel).getByRole('heading', { name: 'Lesson transcript' });
  expect(heading.id).toBe('transcript-editor-lesson-track:two');
  expect(
    within(panel)
      .getByRole('region', { name: 'Lesson transcript' })
      .getAttribute('aria-labelledby'),
  ).toBe(heading.id);
  expect(container.querySelectorAll('form')).toHaveLength(1);
});

test('save uses lowercased tab language, fresh reads, and new ids on every PUT', async () => {
  const f = fixture();
  f.state.tracks = [track('one', 'EN')];
  mount(f);
  save();
  await waitFor(() => expect(f.state.puts).toHaveLength(1));
  expect(f.state.puts[0]?.slice(0, 3)).toEqual([lessonId, 'en', 'T1']);
  expect(f.state.reads).toEqual(['inventory', 'video']);
  expect(f.state.puts[0]?.[3][0]?.id).not.toBe(segment.id);
  save();
  await waitFor(() => expect(f.state.puts).toHaveLength(2));
  expect(f.state.puts[1]?.[3][0]?.id).not.toBe(f.state.puts[0]?.[3][0]?.id);
});

for (const mode of ['rename', 'delete', 'tracks failure', 'inventory failure'] as const) {
  test(`fresh save check refuses ${mode}, keeps lines and sends zero PUT`, async () => {
    const f = fixture();
    mount(f);
    edit();
    if (mode === 'rename') f.state.freshTracks = [track('one', 'fr')];
    if (mode === 'delete') f.state.freshTracks = [];
    if (mode === 'tracks failure') f.state.tracksFailure = true;
    if (mode === 'inventory failure') f.state.inventoryFailure = true;
    save();
    await screen.findByText(
      mode.includes('failure')
        ? 'Could not verify the track — try again.'
        : 'The track changed or was removed. Reload the tracks and try again.',
    );
    expect(f.state.puts).toHaveLength(0);
    expect(
      screen.getByRole('textbox', { name: 'Line 1 text' }).getAttribute('value') ??
        (screen.getByRole('textbox', { name: 'Line 1 text' }) as HTMLTextAreaElement).value,
    ).toBe('My edit');
    if (mode.includes('failure')) {
      f.state.tracksFailure = false;
      f.state.inventoryFailure = false;
      save();
      await waitFor(() => expect(f.state.puts).toHaveLength(1));
    }
  });
}

test('clean rename replaces the base lines; dirty rename preserves lines and saves a new language without base token', async () => {
  const f = fixture();
  f.state.bases.set('fr', transcript('fr', 'French', 'F1'));
  const view = mount(f);
  f.state.tracks = [track('one', 'fr')];
  view.refresh();
  expect((screen.getByRole('textbox', { name: 'Line 1 text' }) as HTMLTextAreaElement).value).toBe(
    'French',
  );
  edit();
  f.state.tracks = [track('one', 'es')];
  view.refresh();
  expect(
    screen.getByText('The track language is now es; these lines will be saved in es.'),
  ).toBeTruthy();
  expect((screen.getByRole('textbox', { name: 'Line 1 text' }) as HTMLTextAreaElement).value).toBe(
    'My edit',
  );
  save();
  await waitFor(() => expect(f.state.puts).toHaveLength(1));
  expect(f.state.puts[0]?.slice(0, 3)).toEqual([lessonId, 'es', undefined]);
});

test('same-language dirty save never advances its base token from fresh inventory (R13a)', async () => {
  const f = fixture();
  const view = mount(f);
  edit();
  f.state.bases.set('en', transcript('en', 'Other author', 'T2'));
  f.state.freshInventory = [item('en', 'T2')];
  view.refresh();
  f.state.putError = 'transcript_modified';
  save();
  await screen.findByText('The transcript changed while you were editing. Reload the page.');
  expect(f.state.puts[0]?.[2]).toBe('T1');
  expect((screen.getByRole('textbox', { name: 'Line 1 text' }) as HTMLTextAreaElement).value).toBe(
    'My edit',
  );
});

test('renamed save confirms an existing target and uses exactly the displayed token (R13c)', async () => {
  const f = fixture();
  const view = mount(f);
  edit();
  f.state.tracks = [track('one', 'fr')];
  f.state.freshInventory = [item('fr', 'F1')];
  view.refresh();
  save();
  await screen.findByRole('dialog');
  expect(
    screen.getByText(
      'A transcript in fr (updated F1) already exists — replace it with these lines?',
    ),
  ).toBeTruthy();
  expect(f.state.puts).toHaveLength(0);
  f.state.freshInventory = [item('fr', 'F2')];
  f.state.putError = 'transcript_modified';
  fireEvent.click(screen.getByRole('button', { name: 'Replace' }));
  await screen.findByText('The transcript changed while you were editing. Reload the page.');
  expect(f.state.puts[0]?.slice(0, 3)).toEqual([lessonId, 'fr', 'F1']);
});

test('a track renamed while the replace dialog is open sends zero PUT (M3-R10)', async () => {
  const f = fixture();
  const view = mount(f);
  edit();
  f.state.tracks = [track('one', 'fr')];
  f.state.freshInventory = [item('fr', 'F1')];
  view.refresh();
  save();
  await screen.findByRole('dialog');
  f.state.freshTracks = [track('one', 'de')];
  fireEvent.click(screen.getByRole('button', { name: 'Replace' }));
  await screen.findByText('The track changed or was removed. Reload the tracks and try again.');
  expect(f.state.puts).toHaveLength(0);
  expect(screen.getByRole('textbox', { name: 'Line 1 text' })).toHaveProperty('value', 'My edit');
});

test('a failed background refresh keeps the editors and their unsaved lines', () => {
  const f = fixture();
  const view = mount(f);
  edit();
  f.state.inventoryRefreshFailed = true;
  view.refresh();
  expect(screen.getByRole('alert').textContent).toContain(
    'Could not load the transcript languages',
  );
  expect(screen.getByRole('tablist')).toBeTruthy();
  expect(screen.getByRole('textbox', { name: 'Line 1 text' })).toHaveProperty('value', 'My edit');
});

test('a target created after inventory read is PUT without a token and names transcript_exists (R13d)', async () => {
  const f = fixture();
  const view = mount(f);
  edit();
  f.state.tracks = [track('one', 'es')];
  view.refresh();
  f.state.putError = 'transcript_exists';
  save();
  await screen.findByText('A transcript in that language already exists.');
  expect(f.state.puts[0]?.[2]).toBeUndefined();
});

test('orphan save only reads inventory; known state comes from the complete inventory', async () => {
  const f = fixture();
  f.state.inventory = [item('fr')];
  f.state.bases.set('fr', transcript('fr'));
  mount(f);
  expect(f.state.known).toContain(true);
  fireEvent.click(screen.getByRole('tab', { name: 'French' }));
  save();
  await waitFor(() => expect(f.state.puts).toHaveLength(1));
  expect(f.state.reads).toEqual(['inventory']);
  expect(f.state.puts[0]?.[1]).toBe('fr');
});

function readyDraft(extra: Partial<LessonDraft> = {}): LessonDraft {
  return {
    lessonId,
    trackId: 'one',
    state: 'ready',
    trackLanguage: 'und',
    trackMultilingual: false,
    stale: false,
    language: 'en',
    updatedAt: 'D1',
    failure: null,
    assetId: 'video',
    generation: 1,
    segments: [segment],
    ...extra,
  };
}

test('und-family track loads its draft, but Save asks for Audio tracks language and sends zero PUT', async () => {
  const f = fixture();
  f.state.tracks = [track('one', 'und-x')];
  f.state.draft = readyDraft();
  mount(f);
  fireEvent.click(screen.getByRole('button', { name: 'Load draft into editor' }));
  fireEvent.click(screen.getByRole('button', { name: 'Load the draft' }));
  save();
  await screen.findByText("Set the track's language in Audio tracks first");
  expect(f.state.puts).toHaveLength(0);
  expect(screen.getByRole('tab', { name: /Unknown language/ })).toBeTruthy();
});

for (const mismatch of [{ stale: true }, { trackLanguage: 'de' }, { trackMultilingual: true }]) {
  test(`draft guard refuses ${JSON.stringify(mismatch)} with Regenerate instruction`, () => {
    const f = fixture();
    f.state.draft = readyDraft(mismatch);
    mount(f);
    expect(screen.queryByRole('button', { name: 'Load draft into editor' })).toBeNull();
    expect(
      screen.getByText(
        'The draft was recognized for another track language or mode — press Regenerate',
      ),
    ).toBeTruthy();
  });
}

test('Regenerate receives the active track id without a language Select; sibling remains enabled', async () => {
  const f = fixture();
  f.state.tracks.push(track('two', 'fr', { isDefault: false }));
  const { container } = mount(f);
  expect(screen.queryByRole('combobox')).toBeNull();
  fireEvent.click(screen.getByRole('button', { name: 'Generate transcript' }));
  await waitFor(() => expect(f.state.regenerate).toEqual(['one']));
  fireEvent.click(screen.getByRole('tab', { name: 'French' }));
  expect(screen.getByRole('button', { name: 'Generate transcript' }).hasAttribute('disabled')).toBe(
    false,
  );
  expect(container.querySelectorAll('.lesson-language-panel')).toHaveLength(2);
});

test('snapshot recovery uses a tab key and retains the old optimistic token (R13b)', async () => {
  const f = fixture();
  await f.historyStore.write({
    key: 'author:lesson:track:one',
    capturedAt: Date.now(),
    language: 'en',
    expectedUpdatedAt: 'T0',
    lines: [{ id: 'recovery-id', startMs: '0', endMs: '1000', text: 'Recovered', documentIds: [] }],
  });
  f.state.putError = 'transcript_modified';
  mount(f);
  fireEvent.click(await screen.findByRole('button', { name: 'Restore changes' }));
  save();
  await screen.findByText('The transcript changed while you were editing. Reload the page.');
  expect(f.state.puts[0]?.[2]).toBe('T0');
  expect((screen.getByRole('textbox', { name: 'Line 1 text' }) as HTMLTextAreaElement).value).toBe(
    'Recovered',
  );
});

test('recovery keys isolate mounted track and orphan editors', async () => {
  const f = fixture();
  f.state.inventory.push(item('fr'));
  f.state.bases.set('fr', transcript('fr'));
  mount(f);
  edit();
  fireEvent.focusOut(screen.getByRole('textbox', { name: 'Line 1 text' }));
  fireEvent.click(screen.getByRole('tab', { name: 'French' }));
  fireEvent.change(screen.getByRole('textbox', { name: 'Line 1 text' }), {
    target: { value: 'Orphan edit' },
  });
  fireEvent.focusOut(screen.getByRole('textbox', { name: 'Line 1 text' }));
  await waitFor(async () =>
    expect((await f.historyStore.read('author:lesson:lang:fr'))?.lines[0]?.text).toBe(
      'Orphan edit',
    ),
  );
  expect((await f.historyStore.read('author:lesson:track:one'))?.lines[0]?.text).toBe('My edit');
  expect(await f.historyStore.read('author:lesson')).toBeNull();
});

test('per-track pending keeps only the requested Regenerate button busy', async () => {
  const f = fixture();
  f.state.tracks.push(track('two', 'fr', { isDefault: false }));
  let finish: (() => void) | undefined;
  f.drafts.actions.regenerate = async (_id, trackId) => {
    f.state.regenerate.push(trackId);
    await new Promise<void>((resolve) => {
      finish = resolve;
    });
    return null;
  };
  mount(f);
  fireEvent.click(screen.getByRole('button', { name: 'Generate transcript' }));
  expect(screen.getByRole('button', { name: 'Generate transcript' }).hasAttribute('disabled')).toBe(
    true,
  );
  fireEvent.click(screen.getByRole('tab', { name: 'French' }));
  expect(screen.getByRole('button', { name: 'Generate transcript' }).hasAttribute('disabled')).toBe(
    false,
  );
  finish?.();
  await waitFor(() => expect(f.state.regenerate).toEqual(['one']));
});

test('tab keyboard navigation exposes only the selected editor', async () => {
  const f = fixture();
  f.state.tracks.push(track('two', 'fr', { isDefault: false }));
  mount(f);
  const user = userEvent.setup();
  screen.getByRole('tab', { name: /English/ }).focus();
  await user.keyboard('{ArrowRight}');
  expect(screen.getByRole('tab', { name: 'French' }).getAttribute('aria-selected')).toBe('true');
  await user.tab();
  expect(screen.getByRole('tabpanel').contains(document.activeElement)).toBe(true);
});

test('flag changes invalidate an already offered draft without triggering recognition', () => {
  const f = fixture();
  f.state.draft = readyDraft();
  const view = mount(f);
  fireEvent.click(screen.getByRole('button', { name: 'Load draft into editor' }));
  expect(screen.getByRole('button', { name: 'Load the draft' })).toBeTruthy();
  f.state.tracks = [track('one', 'en', { multilingual: true })];
  view.refresh();
  expect(screen.queryByRole('button', { name: 'Load the draft' })).toBeNull();
  expect(screen.queryByRole('button', { name: 'Load draft into editor' })).toBeNull();
  expect(f.state.regenerate).toHaveLength(0);
});
