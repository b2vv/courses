import { describe, expect, it } from '@rstest/core';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import axe from 'axe-core';
import { createI18n, I18nProvider } from '../../../app/providers/i18n';
import type { DraftErrorCode, DraftSegment, LessonDraft } from '../domain/draft';
import { LessonDraftField, type LessonDraftFieldActions } from './LessonDraftField';

const lessonId = '20000000-0000-4000-8000-000000000001';

function draftActions(overrides: Partial<LessonDraftFieldActions> = {}): LessonDraftFieldActions {
  return {
    regenerate: async () => null,
    ...overrides,
  };
}

function draft(state: LessonDraft['state'], overrides: Partial<LessonDraft> = {}): LessonDraft {
  return {
    lessonId,
    state,
    language: null,
    failure: null,
    stale: false,
    updatedAt: new Date().toISOString(),
    assetId: null,
    generation: null,
    segments: [],
    ...overrides,
  };
}

function segments(count: number): DraftSegment[] {
  return Array.from({ length: count }, (_, i) => ({
    id: `seg-${i}`,
    startMs: i * 1000,
    endMs: (i + 1) * 1000,
    text: `Segment ${i}`,
  }));
}

function renderField(
  fieldDraft: LessonDraft | undefined,
  fieldActions: LessonDraftFieldActions = draftActions(),
  options: {
    isLoading?: boolean;
    onLoadDraft?: (segs: readonly DraftSegment[], lang: string) => void;
  } = {},
) {
  const hookDraft = fieldDraft;
  const hookLoading = options.isLoading ?? false;
  return render(
    <I18nProvider instance={createI18n('en')}>
      <main>
        <LessonDraftField
          lessonId={lessonId}
          useLessonDraft={() => ({ data: hookDraft, isLoading: hookLoading })}
          actions={fieldActions}
          {...(options.onLoadDraft === undefined ? {} : { onLoadDraft: options.onLoadDraft })}
        />
      </main>
    </I18nProvider>,
  );
}

describe('lesson draft field', () => {
  it('renders nothing while loading', () => {
    const { container } = renderField(undefined, draftActions(), { isLoading: true });
    expect(container.querySelector('.admin-draft-field')).toBeNull();
  });

  it('shows generate button when no draft exists and passes axe', async () => {
    const { container } = renderField(draft(null));
    expect(screen.getByRole('button', { name: 'Generate transcript' })).toBeTruthy();
    expect(screen.queryByRole('button', { name: 'Regenerate transcript' })).toBeNull();
    expect((await axe.run(container)).violations).toEqual([]);
  });

  it('shows regenerate button when a draft exists', () => {
    renderField(draft('ready', { segments: segments(2), language: 'en' }));
    expect(screen.getByRole('button', { name: 'Regenerate transcript' })).toBeTruthy();
    expect(screen.queryByRole('button', { name: 'Generate transcript' })).toBeNull();
  });

  it('shows indeterminate progress bar while running', () => {
    renderField(draft('running'));
    expect(screen.getByRole('progressbar', { name: 'Recognizing speech' })).toBeTruthy();
    expect(screen.queryByRole('button', { name: 'Generate transcript' })).toBeNull();
    expect(screen.queryByRole('button', { name: 'Regenerate transcript' })).toBeNull();
  });

  it('shows indeterminate progress bar while pending', () => {
    renderField(draft('pending'));
    expect(screen.getByRole('progressbar', { name: 'Recognizing speech' })).toBeTruthy();
  });

  it('shows state message for each state', () => {
    const { rerender } = renderField(draft('pending'));
    expect(screen.getByText('Queued for recognition')).toBeTruthy();

    rerender(
      <I18nProvider instance={createI18n('en')}>
        <main>
          <LessonDraftField
            lessonId={lessonId}
            useLessonDraft={() => ({ data: draft('running'), isLoading: false })}
            actions={draftActions()}
          />
        </main>
      </I18nProvider>,
    );
    expect(screen.getByText('Recognizing speech…')).toBeTruthy();
  });

  it('shows ready state and load-draft button when segments exist', () => {
    const loadDraft = () => {};
    renderField(draft('ready', { segments: segments(3), language: 'uk' }), draftActions(), {
      onLoadDraft: loadDraft,
    });
    expect(screen.getByText('Draft ready')).toBeTruthy();
    expect(screen.getByRole('button', { name: 'Load draft into editor' })).toBeTruthy();
  });

  it('hides load-draft button when no onLoadDraft callback', () => {
    renderField(draft('ready', { segments: segments(1), language: 'uk' }));
    expect(screen.queryByRole('button', { name: 'Load draft into editor' })).toBeNull();
  });

  it('calls onLoadDraft with segments and language when load-draft button is pressed', async () => {
    const user = userEvent.setup();
    const segs = segments(2);
    let receivedSegments: readonly DraftSegment[] = [];
    let receivedLanguage = '';
    renderField(draft('ready', { segments: segs, language: 'en' }), draftActions(), {
      onLoadDraft: (s, l) => {
        receivedSegments = s;
        receivedLanguage = l;
      },
    });

    await user.click(screen.getByRole('button', { name: 'Load draft into editor' }));
    expect(receivedSegments).toEqual(segs);
    expect(receivedLanguage).toBe('en');
  });

  it('shows failed state with named failure reason', () => {
    renderField(draft('failed', { failure: 'model_error' }));
    expect(screen.getByText('Recognition failed: model_error.')).toBeTruthy();
  });

  it('shows generic failed state when no failure reason', () => {
    renderField(draft('failed'));
    expect(screen.getByText('Recognition failed')).toBeTruthy();
  });

  it('shows no_speech state', () => {
    renderField(draft('no_speech'));
    expect(screen.getByText('No speech detected in the video')).toBeTruthy();
  });

  it('C06 shows stale indicator when draft.stale is true', () => {
    renderField(draft('ready', { stale: true, segments: segments(1), language: 'uk' }));
    expect(
      screen.getByText(
        'This draft was made from an older video or another audio track. A new draft is being prepared automatically; you can also regenerate.',
      ),
    ).toBeTruthy();
  });

  it('hides stale indicator when draft.stale is false', () => {
    renderField(draft('ready', { stale: false, segments: segments(1), language: 'uk' }));
    expect(
      screen.queryByText(
        'This draft was made from an older video or another audio track. A new draft is being prepared automatically; you can also regenerate.',
      ),
    ).toBeNull();
  });

  it('C11 auto language sends undefined to regenerate, never the string "auto"', async () => {
    const user = userEvent.setup();
    let capturedLanguage: string | undefined = 'NOT_CALLED';
    renderField(
      draft(null),
      draftActions({
        regenerate: async (_lessonId, language) => {
          capturedLanguage = language;
          return null;
        },
      }),
    );

    await user.click(screen.getByRole('button', { name: 'Generate transcript' }));
    await waitFor(() => expect(capturedLanguage).not.toBe('NOT_CALLED'));
    expect(capturedLanguage).toBeUndefined();
  });

  it('sends selected language code when a specific language is chosen', async () => {
    const user = userEvent.setup();
    let capturedLanguage: string | undefined = 'NOT_CALLED';
    renderField(
      draft(null),
      draftActions({
        regenerate: async (_lessonId, language) => {
          capturedLanguage = language;
          return null;
        },
      }),
    );

    const select = screen.getByRole('button', { name: /Recognition language/i });
    await user.click(select);
    const option = screen.getByRole('option', { name: 'English' });
    await user.click(option);
    await user.click(screen.getByRole('button', { name: 'Generate transcript' }));
    await waitFor(() => expect(capturedLanguage).not.toBe('NOT_CALLED'));
    expect(capturedLanguage).toBe('en');
  });

  it('shows action error from regenerate and clears it on retry', async () => {
    const user = userEvent.setup();
    let calls = 0;
    renderField(
      draft(null),
      draftActions({
        regenerate: async () => {
          calls++;
          return calls === 1 ? ('already_running' as DraftErrorCode) : null;
        },
      }),
    );

    await user.click(screen.getByRole('button', { name: 'Generate transcript' }));
    await waitFor(() =>
      expect(screen.getByRole('alert').textContent).toContain(
        'Recognition is already running for this lesson.',
      ),
    );

    await user.click(screen.getByRole('button', { name: 'Generate transcript' }));
    await waitFor(() => expect(screen.queryByRole('alert')).toBeNull());
  });

  it('shows dependency_unavailable when regenerate throws', async () => {
    const user = userEvent.setup();
    renderField(
      draft(null),
      draftActions({
        regenerate: async () => {
          throw new Error('network error');
        },
      }),
    );

    await user.click(screen.getByRole('button', { name: 'Generate transcript' }));
    await waitFor(() =>
      expect(screen.getByRole('alert').textContent).toContain(
        'Recognition service is temporarily unavailable. Try again later.',
      ),
    );
  });

  it('sets data-draft-state attribute on state output', () => {
    renderField(draft('ready', { segments: segments(1), language: 'uk' }));
    const output = screen.getByText('Draft ready');
    expect(output.getAttribute('data-draft-state')).toBe('ready');
  });
});
