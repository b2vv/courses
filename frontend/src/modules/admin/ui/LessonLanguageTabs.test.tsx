import { expect, test, rstest } from '@rstest/core';
import { act, fireEvent, render, screen, waitFor, within } from '@testing-library/react';
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
    transcriptLoading: false,
    inventoryQueryFailed: false,
    inventoryRefreshFailed: false,
    inventoryFailure: false,
    bases: new Map<string, LessonTranscript>([['en', transcript('en')]]),
    patches: [] as Parameters<NonNullable<CourseEditorVideo['audio']>['actions']['rename']>[],
    operations: [] as string[],
    patchError: null as Awaited<
      ReturnType<NonNullable<CourseEditorVideo['audio']>['actions']['rename']>
    >,
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
      data: state.transcriptLoading
        ? undefined
        : (state.bases.get(language ?? 'en') ?? {
            lessonId,
            language: null,
            updatedAt: null,
            stale: false,
            segments: [],
            documents: [],
          }),
      isLoading: state.transcriptLoading,
    }),
    documents: [],
    history: { store: historyStore, authorId: 'author' },
    library: { add: async () => null, remove: async () => null },
    save: async (...args) => {
      state.operations.push('PUT');
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
        rename: async (...args) => {
          state.operations.push('PATCH');
          state.patches.push(args);
          return state.patchError;
        },
        setDefault: async () => null,
        remove: async () => null,
      },
    },
  };
  const node = (ready = true, locale: 'en' | 'uk' = 'en') => (
    <I18nProvider instance={createI18n(locale)}>
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
async function mount(f: ReturnType<typeof fixture>, ready = true) {
  const view = render(
    <>
      <style>{'.lesson-language-panel[inert] { display: none; }'}</style>
      {f.node(ready)}
    </>,
  );
  await act(async () => {});
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

test('tab order is default, embedded, language/id, then orphan; non-ready tracks are excluded', async () => {
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

test('no Ready asset keeps the original single editor ids, roles and copy', async () => {
  const f = fixture();
  const { container } = await mount(f, false);
  expect(screen.queryByRole('tablist')).toBeNull();
  expect(screen.getByRole('heading', { name: 'Lesson transcript' }).id).toBe(
    'transcript-editor-lesson',
  );
  expect(container.querySelectorAll('form')).toHaveLength(1);
});

test('tabs wait for the tracks, so a track language is never first shown as an orphan', async () => {
  const f = fixture();
  f.state.tracks = [track('default', 'und')];
  f.state.inventory = [item('uk')];
  f.state.tracksLoading = true;
  const view = await mount(f);
  expect(screen.queryByRole('tablist')).toBeNull();
  expect(screen.queryByRole('heading', { name: 'Lesson transcript' })).toBeNull();
  f.state.tracksLoading = false;
  view.refresh();
  const tabs = screen.getAllByRole('tab');
  expect(tabs.map((tab) => tab.getAttribute('aria-selected'))).toEqual(['true', 'false']);
  expect(tabs[0]?.textContent).toContain('Unknown language');
});

test('a failed inventory read shows an alert instead of an empty tab set', async () => {
  const f = fixture();
  f.state.inventoryQueryFailed = true;
  await mount(f);
  expect(screen.queryByRole('tablist')).toBeNull();
  expect(screen.getByRole('alert').textContent).toContain(
    'Could not load the transcript languages',
  );
});

test('mounted editors have unique ids and inactive inert panels, with names targeting the active editor', async () => {
  const f = fixture();
  f.state.tracks.push(track('two', 'de', { isDefault: false, multilingual: true }));
  f.state.inventory.push(item('fr'));
  const { container } = await mount(f);
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
  await mount(f);
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
    await mount(f);
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
  const view = await mount(f);
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
  const view = await mount(f);
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
  const view = await mount(f);
  edit();
  f.state.tracks = [track('one', 'fr')];
  f.state.freshInventory = [item('fr', 'F1')];
  view.refresh();
  save();
  await screen.findByRole('dialog');
  expect(
    screen.getByText(
      'A transcript in French (updated F1) already exists — replace it with these lines?',
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
  const view = await mount(f);
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

test('a failed background refresh keeps the editors and their unsaved lines', async () => {
  const f = fixture();
  const view = await mount(f);
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
  const view = await mount(f);
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
  await mount(f);
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

function undFixture(detected = 'en') {
  const f = fixture();
  f.state.tracks = [track('one', 'und-x', { label: 'Original audio', multilingual: true })];
  f.state.inventory = [];
  f.state.draft = readyDraft({ language: detected, trackMultilingual: true });
  return f;
}

function loadUndDraft() {
  fireEvent.click(screen.getByRole('button', { name: 'Load draft into editor' }));
  fireEvent.click(screen.getByRole('button', { name: 'Load the draft' }));
}

test('M3-R4 und-family Save preselects detected draft language; Cancel restores focus and keeps lines with zero writes', async () => {
  const f = undFixture();
  await mount(f);
  loadUndDraft();
  const user = userEvent.setup();
  const button = screen.getByRole('button', { name: 'Save the transcript' });
  await user.click(button);
  const dialog = await screen.findByRole('dialog', { name: 'Confirm the transcript language' });
  expect(within(dialog).getByRole('button', { name: /Track language/ }).textContent).toContain(
    'English',
  );
  expect(dialog.contains(document.activeElement)).toBe(true);
  expect(
    within(dialog).getByText('To attach an existing transcript, rename the track in Audio tracks'),
  ).toBeTruthy();
  await user.click(within(dialog).getByRole('button', { name: 'Cancel' }));
  expect(await screen.findByText('Language not changed; your lines are still here.')).toBeTruthy();
  await waitFor(() => expect(document.activeElement).toBe(button));
  expect(f.state.patches).toHaveLength(0);
  expect(f.state.puts).toHaveLength(0);
  expect(screen.getByRole('textbox', { name: 'Line 1 text' })).toHaveProperty('value', 'Original');
});

test('M3-R4 options use fresh track/inventory reads, excluding und and taken languages case-insensitively; no invalid preselection', async () => {
  const f = undFixture('en');
  f.state.freshTracks = [...f.state.tracks, track('other', 'EN')];
  f.state.freshInventory = [item('UK'), item('pl'), item('de')];
  await mount(f);
  loadUndDraft();
  save();
  const dialog = await screen.findByRole('dialog');
  const user = userEvent.setup();
  expect(within(dialog).getByText('Choose a language')).toBeTruthy();
  const confirm = within(dialog).getByRole('button', { name: 'Rename track and save' });
  await user.click(confirm);
  expect(within(dialog).getByRole('alert').textContent).toBe('Choose a language before saving.');
  expect(document.activeElement).toBe(confirm);
  expect(f.state.operations).toEqual([]);
  await user.click(within(dialog).getByRole('button', { name: /Track language/ }));
  expect(screen.getAllByRole('option').map((option) => option.textContent)).toEqual([
    'French',
    'Spanish',
  ]);
  await user.click(screen.getByRole('option', { name: 'Spanish' }));
  await user.click(confirm);
  await waitFor(() => expect(f.state.puts).toHaveLength(1));
  expect(f.state.operations).toEqual(['PATCH', 'PUT']);
  expect(f.state.puts[0]?.slice(0, 3)).toEqual([lessonId, 'es', undefined]);
});

for (const failure of ['tracksFailure', 'inventoryFailure'] as const) {
  test(`M3-R4 ${failure} means no dialog, zero PATCH and PUT`, async () => {
    const f = undFixture();
    f.state[failure] = true;
    await mount(f);
    loadUndDraft();
    save();
    await screen.findByText('Could not verify the track — try again.');
    expect(screen.queryByRole('dialog')).toBeNull();
    expect(f.state.operations).toEqual([]);
  });
}

test('M3-R4 success sends exactly PATCH then PUT, preserving label/mode, chosen language, no token and fresh ids', async () => {
  const f = undFixture();
  await mount(f);
  loadUndDraft();
  save();
  await screen.findByRole('dialog');
  fireEvent.click(screen.getByRole('button', { name: 'Rename track and save' }));
  await screen.findByText('Transcript saved.');
  expect(f.state.operations).toEqual(['PATCH', 'PUT']);
  expect(f.state.patches[0]).toEqual([f.state.tracks[0], 'en', 'Original audio', true]);
  expect(f.state.puts[0]?.slice(0, 3)).toEqual([lessonId, 'en', undefined]);
  expect(f.state.puts[0]?.[3][0]?.id).not.toBe(segment.id);
  expect(f.state.puts[0]?.[3][0]?.id).toMatch(/^[0-9a-f-]{36}$/);
});

for (const error of ['duplicate_language', 'dependency_unavailable'] as const) {
  test(`MS3b fix2 / M3-R4 PATCH ${error} keeps ordinary capture under own key, no PUT/clear/finalize/move`, async () => {
    const f = undFixture();
    f.state.patchError = error;
    await mount(f);
    loadUndDraft();
    edit();
    fireEvent.focusOut(screen.getByRole('textbox', { name: 'Line 1 text' }));
    const captured = await f.historyStore.read('author:lesson:track:one');
    expect(captured?.lines[0]?.text).toBe('My edit');
    save();
    await screen.findByRole('dialog');
    const write = rstest.spyOn(f.historyStore, 'write');
    const clear = rstest.spyOn(f.historyStore, 'clear');
    fireEvent.click(screen.getByRole('button', { name: 'Rename track and save' }));
    await screen.findByText(
      error === 'duplicate_language'
        ? 'A track in that language already exists.'
        : 'Could not rename the track — try again.',
    );
    expect(f.state.puts).toHaveLength(0);
    expect(write).not.toHaveBeenCalled();
    expect(clear).not.toHaveBeenCalled();
    expect(await f.historyStore.read('author:lesson:track:one')).toEqual(captured);
    expect(screen.getByRole('textbox', { name: 'Line 1 text' })).toHaveProperty('value', 'My edit');
  });
}

test('M3-R4 PATCH succeeds and PUT fails: same mounted tab and typed line survive refetch, next Save is ordinary PUT', async () => {
  const f = undFixture();
  f.state.putError = 'unavailable';
  const view = await mount(f);
  loadUndDraft();
  edit();
  const line = screen.getByRole('textbox', { name: 'Line 1 text' });
  save();
  await screen.findByRole('dialog');
  fireEvent.click(screen.getByRole('button', { name: 'Rename track and save' }));
  await screen.findByText('The service is unavailable. Try again.');
  f.state.tracks = [track('one', 'en', { label: 'Original audio', multilingual: true })];
  f.state.putError = null;
  f.state.transcriptLoading = true;
  view.refresh();
  expect(screen.getByRole('textbox', { name: 'Line 1 text' })).toBe(line);
  f.state.transcriptLoading = false;
  view.refresh();
  expect(screen.getByRole('textbox', { name: 'Line 1 text' })).toBe(line);
  expect(line).toHaveProperty('value', 'My edit');
  expect(screen.getByRole('tab', { name: /English/ })).toBeTruthy();
  save();
  await screen.findByText('Transcript saved.');
  expect(screen.queryByRole('dialog')).toBeNull();
  expect(f.state.operations).toEqual(['PATCH', 'PUT', 'PUT']);
  expect(f.state.puts[1]?.slice(0, 3)).toEqual([lessonId, 'en', undefined]);
  expect(f.state.puts[1]?.[3][0]?.id).not.toBe(f.state.puts[0]?.[3][0]?.id);
  expect(f.state.reads).toEqual(['inventory', 'video', 'inventory', 'video']);
});

test('M3-R4 no options offers only Cancel; Escape returns focus to Save', async () => {
  const f = undFixture();
  f.state.freshInventory = ['uk', 'en', 'pl', 'de', 'fr', 'es'].map((code) => item(code));
  await mount(f);
  loadUndDraft();
  const user = userEvent.setup();
  const button = screen.getByRole('button', { name: 'Save the transcript' });
  await user.click(button);
  const dialog = await screen.findByRole('dialog');
  expect(
    within(dialog)
      .getAllByRole('button')
      .map((button) => button.textContent),
  ).toEqual(['Cancel']);
  expect(within(dialog).getByText(/No languages are available/)).toBeTruthy();
  await user.keyboard('{Escape}');
  await waitFor(() => expect(screen.queryByRole('dialog')).toBeNull());
  expect(document.activeElement).toBe(button);
  expect(f.state.operations).toEqual([]);
  expect(await screen.findByText('Language not changed; your lines are still here.')).toBeTruthy();
});

for (const mismatch of [{ stale: true }, { trackLanguage: 'de' }, { trackMultilingual: true }]) {
  test(`draft guard refuses ${JSON.stringify(mismatch)} with Regenerate instruction`, async () => {
    const f = fixture();
    f.state.draft = readyDraft(mismatch);
    await mount(f);
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
  const { container } = await mount(f);
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
    revision: crypto.randomUUID(),
    capturedAt: Date.now(),
    language: 'en',
    baseLanguage: 'en',
    expectedUpdatedAt: 'T0',
    lines: [{ id: 'recovery-id', startMs: '0', endMs: '1000', text: 'Recovered', documentIds: [] }],
  });
  f.state.putError = 'transcript_modified';
  await mount(f);
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
  await mount(f);
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
  await mount(f);
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
  await mount(f);
  const user = userEvent.setup();
  screen.getByRole('tab', { name: /English/ }).focus();
  await user.keyboard('{ArrowRight}');
  expect(screen.getByRole('tab', { name: 'French' }).getAttribute('aria-selected')).toBe('true');
  await user.tab();
  expect(screen.getByRole('tabpanel').contains(document.activeElement)).toBe(true);
});

test('flag changes invalidate an already offered draft without triggering recognition', async () => {
  const f = fixture();
  f.state.draft = readyDraft();
  const view = await mount(f);
  fireEvent.click(screen.getByRole('button', { name: 'Load draft into editor' }));
  expect(screen.getByRole('button', { name: 'Load the draft' })).toBeTruthy();
  f.state.tracks = [track('one', 'en', { multilingual: true })];
  view.refresh();
  expect(screen.queryByRole('button', { name: 'Load the draft' })).toBeNull();
  expect(screen.queryByRole('button', { name: 'Load draft into editor' })).toBeNull();
  expect(f.state.regenerate).toHaveLength(0);
});

for (const targetExists of [false, true]) {
  test(`M3-R13 ${targetExists ? 'f' : 'e'} restores (en,T1) into es, recaptures pair and uses target token only after confirmation`, async () => {
    const f = fixture();
    f.state.tracks = [track('one', 'es')];
    if (targetExists) f.state.freshInventory = [item('es', 'E2')];
    await f.historyStore.write({
      key: 'author:lesson:track:one',
      revision: crypto.randomUUID(),
      capturedAt: 1,
      language: 'es',
      baseLanguage: 'en',
      expectedUpdatedAt: 'T1',
      lines: [
        { id: 'snapshot-id', startMs: '0', endMs: '1000', text: 'Recovered', documentIds: [] },
      ],
    });
    await mount(f);
    fireEvent.click(await screen.findByRole('button', { name: 'Restore changes' }));
    fireEvent.focusOut(screen.getByRole('textbox', { name: 'Line 1 text' }));
    expect(await f.historyStore.read('author:lesson:track:one')).toMatchObject({
      language: 'es',
      baseLanguage: 'en',
      expectedUpdatedAt: 'T1',
    });
    save();
    if (targetExists) {
      await screen.findByRole('dialog', { name: /transcript in Spanish.*E2/ });
      expect(f.state.puts).toHaveLength(0);
      fireEvent.click(screen.getByRole('button', { name: 'Replace' }));
    }
    await waitFor(() => expect(f.state.puts).toHaveLength(1));
    expect(f.state.puts[0]?.slice(0, 3)).toEqual([lessonId, 'es', targetExists ? 'E2' : undefined]);
  });
}

for (const targetExists of [false, true]) {
  test(`M3-R13 g legacy snapshot never reuses base token even for its own save language (${targetExists ? 'present' : 'absent'})`, async () => {
    const f = fixture();
    f.state.freshInventory = targetExists ? [item('en', 'T2')] : [];
    await f.historyStore.write({
      key: 'author:lesson:track:one',
      revision: crypto.randomUUID(),
      capturedAt: 1,
      language: 'en',
      expectedUpdatedAt: 'T0',
      lines: [{ id: 'snapshot-id', startMs: '0', endMs: '1000', text: 'Legacy', documentIds: [] }],
    });
    await mount(f);
    fireEvent.click(await screen.findByRole('button', { name: 'Restore changes' }));
    fireEvent.focusOut(screen.getByRole('textbox', { name: 'Line 1 text' }));
    expect((await f.historyStore.read('author:lesson:track:one'))?.baseLanguage).toBe('');
    save();
    if (targetExists) {
      await screen.findByRole('dialog', { name: /transcript in English.*T2/ });
      fireEvent.click(screen.getByRole('button', { name: 'Replace' }));
    }
    await waitFor(() => expect(f.state.puts).toHaveLength(1));
    expect(f.state.puts[0]?.[2]).toBe(targetExists ? 'T2' : undefined);
  });
}

function recoverySnapshot(
  key: string,
  text: string,
  capturedAt = 1,
  language: string | null = 'en',
) {
  return {
    key,
    revision: crypto.randomUUID(),
    capturedAt,
    ...(language === null ? {} : { language }),
    baseLanguage: 'en',
    expectedUpdatedAt: 'T1',
    lines: [{ ...segment, startMs: '0', endMs: '1000', text }],
  };
}

test('M3-R6 own offer first, then detached newest one at a time; Discard clears only the detached key', async () => {
  const f = fixture();
  await f.historyStore.write(recoverySnapshot('author:lesson:track:one', 'own'));
  await f.historyStore.write(recoverySnapshot('author:lesson', 'old legacy', 1));
  await f.historyStore.write(recoverySnapshot('author:lesson:track:gone', 'new detached', 2));
  await mount(f);
  await screen.findByRole('heading', { name: 'Unsaved changes' });
  expect(screen.queryByText(/Unsaved lines/)).toBeNull();
  fireEvent.click(screen.getByRole('button', { name: 'Restore changes' }));
  await screen.findByRole('heading', { name: 'Unsaved lines from another tab (language: en)' });
  expect(screen.getByRole('textbox', { name: 'Line 1 text' })).toHaveProperty('value', 'own');
  expect(screen.queryByRole('heading', { name: /before tabs/ })).toBeNull();
  fireEvent.click(screen.getByRole('button', { name: 'Discard' }));
  await screen.findByRole('heading', { name: 'Unsaved lines from before tabs (language: en)' });
  expect(await f.historyStore.read('author:lesson:track:gone')).toBeNull();
  expect(await f.historyStore.read('author:lesson')).not.toBeNull();
  expect(await f.historyStore.read('author:lesson:track:one')).not.toBeNull();
});

test('M3-R6 Restore + successful Save clears accepted detached AND finalizes own key', async () => {
  const f = fixture();
  const source = recoverySnapshot('author:lesson:track:gone', 'recovered');
  await f.historyStore.write(source);
  const write = rstest.spyOn(f.historyStore, 'write');
  const clear = rstest.spyOn(f.historyStore, 'clear');
  const clearIfVersion = rstest.spyOn(f.historyStore, 'clearIfVersion');
  const view = await mount(f);
  fireEvent.click(await screen.findByRole('button', { name: 'Restore changes' }));
  fireEvent.focusOut(screen.getByRole('textbox', { name: 'Line 1 text' }));
  expect(await f.historyStore.read('author:lesson:track:one')).not.toBeNull();
  expect(await f.historyStore.read('author:lesson:track:gone')).not.toBeNull();
  save();
  await screen.findByText('Transcript saved.');
  expect(f.state.puts[0]?.[3][0]?.text).toBe('recovered');
  expect(f.state.puts[0]?.[3][0]?.id).not.toBe(segment.id);
  expect(await f.historyStore.read('author:lesson:track:one')).toBeNull();
  expect(await f.historyStore.read('author:lesson:track:gone')).toBeNull();
  expect(write.mock.calls.every(([s]) => s.key === 'author:lesson:track:one')).toBe(true);
  expect(clear).not.toHaveBeenCalled();
  expect(clearIfVersion).toHaveBeenCalledWith(
    'author:lesson:track:one',
    write.mock.calls.at(-1)?.[0].revision,
  );
  expect(clearIfVersion).toHaveBeenCalledWith('author:lesson:track:gone', source.revision);
  view.unmount();
  expect(await f.historyStore.read('author:lesson:track:one')).toBeNull();
});

for (const newerCapture of [false, true]) {
  test(`MS3b fix4 parent Restore remount own Restore Save two visits newerC=${newerCapture}`, async () => {
    const f = fixture();
    const sourceKey = 'author:lesson:track:gone';
    const ownKey = 'author:lesson:track:one';
    const source = recoverySnapshot(sourceKey, 'recovered');
    await f.historyStore.write(source);
    const first = await mount(f);
    fireEvent.click(await screen.findByRole('button', { name: 'Restore changes' }));
    first.unmount();
    expect((await f.historyStore.read(ownKey))?.lines[0]?.text).toBe('recovered');
    expect(await f.historyStore.read(sourceKey)).toEqual(source);
    const second = await mount(f);
    await screen.findByRole('heading', { name: 'Unsaved changes' });
    fireEvent.click(screen.getByRole('button', { name: 'Restore changes' }));
    expect(
      screen.queryByRole('heading', { name: 'Unsaved lines from another tab (language: en)' }),
    ).toBeNull();
    const newer = recoverySnapshot(sourceKey, 'Newer C', 2);
    if (newerCapture) await f.historyStore.write(newer);
    save();
    await screen.findByText('Transcript saved.');
    expect(f.state.patches).toHaveLength(0);
    expect(f.state.puts).toHaveLength(1);
    expect(f.state.puts[0]?.[3][0]?.text).toBe('recovered');
    expect(await f.historyStore.read(ownKey)).toBeNull();
    await waitFor(async () =>
      expect(await f.historyStore.read(sourceKey)).toEqual(newerCapture ? newer : null),
    );
    // The fixture server must return the successful saved result on later visits.
    f.state.bases.set('en', transcript('en', 'recovered', 'T2'));
    f.state.inventory = [item('en', 'T2')];
    second.unmount();
    for (let visit = 0; visit < 2; visit++) {
      const view = await mount(f);
      const offered = screen.queryByRole('heading', {
        name: 'Unsaved lines from another tab (language: en)',
      });
      expect(offered !== null).toBe(newerCapture);
      expect(screen.queryByRole('heading', { name: 'Unsaved changes' })).toBeNull();
      expect(await f.historyStore.read(sourceKey)).toEqual(newerCapture ? newer : null);
      view.unmount();
    }
  });
}

test('M3-R6 legacy unknown-language copy is offered only in the default panel; one keys read', async () => {
  const f = fixture();
  f.state.tracks.push(track('two', 'uk', { isDefault: false }));
  await f.historyStore.write(recoverySnapshot('author:lesson', 'unknown', 1, null));
  const keys = rstest.spyOn(f.historyStore, 'keys');
  const view = await mount(f);
  await screen.findByRole('heading', { name: 'Unsaved lines from before tabs (language unknown)' });
  expect(keys).toHaveBeenCalledTimes(1);
  fireEvent.click(screen.getByRole('tab', { name: 'Ukrainian' }));
  expect(screen.queryByRole('button', { name: 'Restore changes' })).toBeNull();
  view.refresh();
  expect(keys).toHaveBeenCalledTimes(1);
  const ids = [...document.querySelectorAll('[id]')].map((node) => node.id);
  expect(new Set(ids).size).toBe(ids.length);
});

test('M3-R6 legacy uk snapshot belongs in uk track and never another lesson prefix', async () => {
  const f = fixture();
  f.state.tracks.push(track('two', 'uk', { isDefault: false }));
  await f.historyStore.write(recoverySnapshot('author:lesson', 'uk legacy', 1, 'UK'));
  await f.historyStore.write(recoverySnapshot('author:lesson-other', 'foreign', 2));
  const read = rstest.spyOn(f.historyStore, 'read');
  await mount(f);
  await waitFor(() => expect(read.mock.calls.some(([key]) => key === 'author:lesson')).toBe(true));
  expect(screen.queryByRole('button', { name: 'Restore changes' })).toBeNull();
  fireEvent.click(screen.getByRole('tab', { name: 'Ukrainian' }));
  await screen.findByRole('heading', { name: 'Unsaved lines from before tabs (language: UK)' });
  expect(read.mock.calls.some(([key]) => key === 'author:lesson-other')).toBe(false);
});

test('M3-R12 c dirty orphan is flushed on attachment and offered in the renamed track', async () => {
  const f = fixture();
  f.state.inventory.push(item('uk'));
  f.state.bases.set('uk', transcript('uk'));
  const view = await mount(f);
  fireEvent.click(screen.getByRole('tab', { name: 'Ukrainian' }));
  edit();
  f.state.tracks = [track('one', 'uk')];
  view.refresh();
  await screen.findByRole('heading', { name: 'Unsaved lines from another tab (language: uk)' });
  expect((await f.historyStore.read('author:lesson:lang:uk'))?.lines[0]?.text).toBe('My edit');
  fireEvent.click(screen.getByRole('button', { name: 'Restore changes' }));
  expect(screen.getByRole('textbox', { name: 'Line 1 text' })).toHaveProperty('value', 'My edit');
});

test('M3-R12 single editor pending lines flush before first ready tabs read the legacy snapshot', async () => {
  const f = fixture();
  const view = await mount(f, false);
  edit();
  view.rerender(
    <>
      <style>{'.lesson-language-panel[inert] { display: none; }'}</style>
      {f.node(true)}
    </>,
  );
  await screen.findByRole('heading', { name: 'Unsaved lines from before tabs (language: en)' });
  expect((await f.historyStore.read('author:lesson'))?.lines[0]?.text).toBe('My edit');
});

test('M3-R6 single editor receives detached track snapshots', async () => {
  const f = fixture();
  await f.historyStore.write(recoverySnapshot('author:lesson:track:gone', 'detached'));
  await mount(f, false);
  await screen.findByRole('heading', { name: 'Unsaved lines from another tab (language: en)' });
  fireEvent.click(screen.getByRole('button', { name: 'Restore changes' }));
  expect(screen.getByRole('textbox', { name: 'Line 1 text' })).toHaveProperty('value', 'detached');
});

for (const targetExists of [false, true]) {
  test(`M3-R6 detached restore into another language uses target token rules (exists=${targetExists})`, async () => {
    const f = fixture();
    f.state.tracks = [track('one', 'es')];
    f.state.inventory = targetExists ? [item('es', 'TARGET')] : [];
    f.state.bases.clear();
    if (targetExists) f.state.bases.set('es', transcript('es', 'target', 'TARGET'));
    await f.historyStore.write(recoverySnapshot('author:lesson:track:gone', 'restored', 1, 'en'));
    await mount(f);
    fireEvent.click(await screen.findByRole('button', { name: 'Restore changes' }));
    save();
    if (targetExists) {
      await screen.findByRole('dialog');
      expect(f.state.puts).toHaveLength(0);
      fireEvent.click(screen.getByRole('button', { name: 'Replace' }));
    }
    await screen.findByText('Transcript saved.');
    expect(f.state.puts[0]?.slice(0, 3)).toEqual([
      'lesson',
      'es',
      targetExists ? 'TARGET' : undefined,
    ]);
  });
}

test('M3-R6 failed Save with pending capture writes current lines once, retains detached, unmount flushes nothing stale', async () => {
  const f = fixture();
  f.state.putError = 'unavailable';
  await f.historyStore.write(recoverySnapshot('author:lesson:track:gone', 'restored'));
  const view = await mount(f);
  fireEvent.click(await screen.findByRole('button', { name: 'Restore changes' }));
  edit();
  const write = rstest.spyOn(f.historyStore, 'write');
  const clear = rstest.spyOn(f.historyStore, 'clear');
  save();
  await screen.findByText('The service is unavailable. Try again.');
  expect(write).toHaveBeenCalledTimes(1);
  expect((await f.historyStore.read('author:lesson:track:one'))?.lines[0]?.text).toBe('My edit');
  expect(await f.historyStore.read('author:lesson:track:gone')).not.toBeNull();
  expect(clear).not.toHaveBeenCalled();
  view.unmount();
  expect(write).toHaveBeenCalledTimes(1);
});

for (const unknownBase of [false, true]) {
  test(`M3-R6 single editor detached restore never carries a foreign/unknown base token (${unknownBase})`, async () => {
    const f = fixture();
    const stored = recoverySnapshot('author:lesson:track:gone', 'recovered', 1, 'es');
    const { baseLanguage: _baseLanguage, ...legacy } = stored;
    await f.historyStore.write(unknownBase ? legacy : stored);
    await mount(f, false);
    fireEvent.click(await screen.findByRole('button', { name: 'Restore changes' }));
    save();
    await screen.findByText('Transcript saved.');
    expect(f.state.puts[0]?.slice(0, 3)).toEqual(['lesson', 'es', undefined]);
  });
}

for (const locale of ['en', 'uk'] as const) {
  test(`MS3b fix 6 replacement heading localizes language and date-time in ${locale}`, async () => {
    const f = fixture();
    f.state.inventory = [];
    const view = render(f.node(true, locale));
    await act(async () => {});
    fireEvent.change(
      screen.getByRole('textbox', { name: locale === 'en' ? 'Line 1 text' : 'Текст репліки 1' }),
      { target: { value: 'My edit' } },
    );
    f.state.tracks = [track('one', 'uk')];
    const updatedAt = '2026-11-16T16:11:47.927508Z';
    f.state.freshInventory = [item('uk', updatedAt)];
    view.rerender(f.node(true, locale));
    fireEvent.click(
      screen.getByRole('button', {
        name: locale === 'en' ? 'Save the transcript' : 'Зберегти транскрипт',
      }),
    );
    const dialog = await screen.findByRole('dialog');
    const time = new Intl.DateTimeFormat(locale, {
      dateStyle: 'medium',
      timeStyle: 'short',
    }).format(new Date(updatedAt));
    expect(within(dialog).getByRole('heading').textContent).toBe(
      locale === 'en'
        ? `A transcript in Ukrainian (updated ${time}) already exists — replace it with these lines?`
        : `Транскрипт мовою Українська (оновлено ${time}) уже є — замінити його цими рядками?`,
    );
  });
}

for (const mode of ['und', 'replace'] as const) {
  test(`MS3b fix 4 ${mode} keyboard Enter then Escape restores Save before the next animation frame`, async () => {
    const f = mode === 'und' ? undFixture() : fixture();
    const view = await mount(f);
    if (mode === 'replace') {
      edit();
      f.state.tracks = [track('one', 'fr')];
      f.state.freshInventory = [item('fr', '2026-11-16T16:11:47Z')];
      view.refresh();
    }
    const button = screen.getByRole('button', { name: 'Save the transcript' });
    const user = userEvent.setup();
    button.focus();
    await user.keyboard('{Enter}');
    const dialog = await screen.findByRole('dialog');
    expect(dialog.contains(document.activeElement)).toBe(true);
    // FocusScope's default restore is deferred to RAF. Pin the close-commit contract,
    // so this check cannot pass merely because happy-dom ran that frame early.
    const frame = rstest.spyOn(window, 'requestAnimationFrame').mockImplementation(() => 0);
    try {
      await user.keyboard('{Escape}');
      await waitFor(() => expect(screen.queryByRole('dialog')).toBeNull());
      expect(document.activeElement).toBe(button);
      expect(f.state.puts).toHaveLength(0);
      expect(f.state.patches).toHaveLength(0);
    } finally {
      frame.mockRestore();
    }
  });
}

for (const newerCapture of [true, false]) {
  test(`MS3b fix3 mounted orphan relinquishes detached deletion, rename-back Save remount (newer=${newerCapture})`, async () => {
    const f = fixture();
    const key = 'author:lesson:lang:uk';
    f.state.inventory.push(item('uk'));
    f.state.bases.set('uk', transcript('uk', 'Server uk'));
    await f.historyStore.write({ ...recoverySnapshot(key, 'A', 1, 'uk'), baseLanguage: 'uk' });
    const view = await mount(f);
    f.state.tracks = [track('one', 'uk')];
    view.refresh();
    fireEvent.click(await screen.findByRole('button', { name: 'Restore changes' }));
    expect(screen.getByRole('textbox', { name: 'Line 1 text' })).toHaveProperty('value', 'A');
    f.state.tracks = [track('one', 'en')];
    view.refresh();
    fireEvent.click(screen.getByRole('tab', { name: 'Ukrainian' }));
    await screen.findByRole('button', { name: 'Restore changes' });
    if (newerCapture) {
      fireEvent.click(screen.getByRole('button', { name: 'Restore changes' }));
      fireEvent.change(screen.getByRole('textbox', { name: 'Line 1 text' }), {
        target: { value: 'C' },
      });
      fireEvent.focusOut(screen.getByRole('textbox', { name: 'Line 1 text' }));
    }
    const preserved = await f.historyStore.read(key);
    expect(preserved?.lines[0]?.text).toBe(newerCapture ? 'C' : 'A');
    const clear = rstest.spyOn(f.historyStore, 'clear');
    fireEvent.click(screen.getByRole('tab', { name: /English/ }));
    save();
    const dialog = await screen.findByRole('dialog');
    fireEvent.click(within(dialog).getByRole('button', { name: 'Replace' }));
    await screen.findByText('Transcript saved.');
    expect(f.state.puts).toHaveLength(1);
    expect(f.state.puts[0]?.[1]).toBe('en');
    expect(f.state.puts[0]?.[3][0]?.text).toBe('A');
    expect((await f.historyStore.read(key))?.lines).toEqual(preserved?.lines);
    expect(clear.mock.calls.map(([cleared]) => cleared)).not.toContain(key);
    view.unmount();
    await mount(f);
    fireEvent.click(screen.getByRole('tab', { name: 'Ukrainian' }));
    fireEvent.click(await screen.findByRole('button', { name: 'Restore changes' }));
    expect(screen.getByRole('textbox', { name: 'Line 1 text' })).toHaveProperty(
      'value',
      newerCapture ? 'C' : 'A',
    );
  });
}
