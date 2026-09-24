import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import type { DraftErrorCode, DraftSegment, LessonDraft } from '../../domain/transcript/draft';
import { Button, ProgressBar, Select } from '../ui';
import type { ChoiceOption } from '../ui';

export type LessonDraftFieldActions = {
  regenerate(lessonId: string, language: string | undefined): Promise<DraftErrorCode | null>;
};

export type CourseEditorDrafts = {
  readonly useLessonDraft: (lessonId: string) => {
    readonly data: LessonDraft | undefined;
    readonly isLoading: boolean;
  };
  readonly actions: LessonDraftFieldActions;
};

type LessonDraftFieldProps = CourseEditorDrafts & {
  readonly lessonId: string;
  readonly onLoadDraft?: (
    segments: readonly DraftSegment[],
    language: string,
    updatedAt: string | undefined,
  ) => void;
  readonly onDraftTokenChange?: (updatedAt: string | undefined) => void;
};

const languageOptions: readonly ChoiceOption[] = [
  { id: 'auto', label: 'Auto' },
  { id: 'en', label: 'English' },
  { id: 'uk', label: 'Українська' },
];

function stateMessage(
  draft: LessonDraft,
  t: ReturnType<typeof useTranslation>['t'],
): string | null {
  if (draft.state === null) return null;
  if (draft.state === 'failed' && draft.failure !== null)
    return t('administration.asr.failureNamed', { reason: draft.failure });
  return t(`administration.asr.state.${draft.state}`);
}

function errorMessage(code: DraftErrorCode, t: ReturnType<typeof useTranslation>['t']): string {
  switch (code) {
    case 'authentication_required':
      return t('administration.asr.error.authenticationRequired');
    case 'origin_forbidden':
      return t('administration.asr.error.originForbidden');
    case 'course_forbidden':
      return t('administration.asr.error.forbidden');
    case 'lesson_not_found':
    case 'asset_not_found':
      return t('administration.asr.error.notFound');
    case 'already_running':
      return t('administration.asr.error.alreadyRunning');
    case 'invalid_request':
      return t('administration.asr.error.invalidRequest');
    case 'request_cancelled':
      return t('administration.asr.error.requestCancelled');
    case 'request_timeout':
    case 'dependency_unavailable':
    case 'invalid_response':
      return t('administration.asr.error.unavailable');
  }
}

export function LessonDraftField({
  lessonId,
  useLessonDraft,
  actions,
  onLoadDraft,
  onDraftTokenChange,
}: LessonDraftFieldProps) {
  const { t } = useTranslation();
  const { data: draft, isLoading } = useLessonDraft(lessonId);
  const [selectedLanguage, setSelectedLanguage] = useState<string>('auto');
  const [isBusy, setIsBusy] = useState(false);
  const [actionError, setActionError] = useState<DraftErrorCode>();

  useEffect(() => {
    onDraftTokenChange?.(draft?.updatedAt ?? undefined);
  }, [draft?.updatedAt, onDraftTokenChange]);

  async function regenerate() {
    setIsBusy(true);
    setActionError(undefined);
    try {
      const language = selectedLanguage === 'auto' ? undefined : selectedLanguage;
      const code = await actions.regenerate(lessonId, language);
      if (code !== null) setActionError(code);
    } catch {
      setActionError('dependency_unavailable');
    } finally {
      setIsBusy(false);
    }
  }

  if (isLoading) return null;

  const isRunning = draft?.state === 'running' || draft?.state === 'pending';
  const isReady = draft?.state === 'ready';
  const hasDraft = draft?.state !== null && draft?.state !== undefined;
  const message = draft !== undefined ? stateMessage(draft, t) : null;

  return (
    <section className="admin-draft-field" aria-busy={isBusy || isRunning}>
      <h3>{t('administration.asr.title')}</h3>
      {actionError !== undefined && (
        <p className="admin-draft-error" role="alert">
          {errorMessage(actionError, t)}
        </p>
      )}
      {hasDraft && message !== null && (
        <output className="admin-draft-state" data-draft-state={draft.state}>
          {message}
        </output>
      )}
      {isRunning && <ProgressBar label={t('administration.asr.recognizing')} isIndeterminate />}
      {draft?.stale === true && (
        <p className="admin-draft-stale">{t('administration.asr.stale')}</p>
      )}
      {isReady && draft.segments.length > 0 && onLoadDraft !== undefined && (
        <Button
          type="button"
          variant="secondary"
          onPress={() =>
            onLoadDraft(draft.segments, draft.language ?? 'uk', draft.updatedAt ?? undefined)
          }
        >
          {t('administration.asr.loadDraft')}
        </Button>
      )}
      {!isRunning && (
        <div className="admin-draft-regenerate">
          <Select
            label={t('administration.asr.language')}
            options={[...languageOptions]}
            selectedKey={selectedLanguage}
            onSelectionChange={(key) => setSelectedLanguage(String(key))}
          />
          <Button
            type="button"
            variant="secondary"
            isDisabled={isBusy}
            onPress={() => void regenerate()}
          >
            {hasDraft ? t('administration.asr.regenerate') : t('administration.asr.generate')}
          </Button>
        </div>
      )}
    </section>
  );
}
