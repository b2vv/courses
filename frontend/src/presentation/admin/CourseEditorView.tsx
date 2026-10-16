import { useCallback, useEffect, useRef, useState } from 'react';
import { flushSync } from 'react-dom';
import type { ReactNode } from 'react';
import { useFieldArray, useForm, useController, useWatch } from 'react-hook-form';
import type {
  Control,
  FieldPath,
  FieldPathByValue,
  FieldValues,
  UseFormReturn,
} from 'react-hook-form';
import { useTranslation } from 'react-i18next';
import { AuthoringQueryError } from '../../app/composition/admin-queries';
import { newLessonId, newModuleId } from '../../domain/course/course';
import type {
  AdminCourseDraft,
  AuthoringErrorCode,
  CoursePrice,
  CourseStatus,
  CourseTree,
  CourseTreeEdit,
  InstructorOption,
} from '../../domain/course/course';
import type { DraftSegment } from '../../domain/transcript/draft';
import type { AuthenticatedUser } from '../../domain/identity/identity';
import {
  Button,
  Checkbox,
  Dialog,
  Link,
  PageHeader,
  SegmentedControl,
  Select,
  Skeleton,
  TextArea,
  TextField,
} from '../ui';
import { CoursePriceField } from './CoursePriceField';
import { LessonDraftField } from './LessonDraftField';
import type { CourseEditorDrafts } from './LessonDraftField';
import { LessonVideoField } from './LessonVideoField';
import { DocumentLibrary } from './DocumentLibrary';
import { LessonTranscriptField } from './LessonTranscriptField';
import type { CourseEditorTranscripts } from './LessonTranscriptField';
import type { TranscriptDraftOffer } from './transcript-draft-offer';
import type { CourseEditorVideo } from './LessonVideoField';

const statusVariant: Record<CourseStatus, string> = {
  draft: 'warning',
  published: 'success',
  archived: 'neutral',
};

type CourseFormValues = {
  slug: string;
  title: string;
  summary: string;
  description: string;
  instructorId: string;
  modules: Array<{
    id: CourseTreeEdit['modules'][number]['id'];
    title: string;
    lessons: Array<{
      id: CourseTreeEdit['modules'][number]['lessons'][number]['id'];
      title: string;
      isPreview: boolean;
    }>;
  }>;
};

type CourseEditorViewProps = {
  course: CourseTree;
  actor: AuthenticatedUser;
  ownerOptions: readonly InstructorOption[] | null;
  onSave: (current: CourseTree, edit: CourseTreeEdit) => Promise<CourseTree>;
  onSetPrice?: (current: CourseTree, price: CoursePrice | null) => Promise<CourseTree>;
  onTransition: (current: CourseTree, status: CourseStatus) => Promise<CourseTree>;
  onDelete: (current: CourseTree) => Promise<void>;
  onAuthenticationRequired: () => void;
  video?: CourseEditorVideo;
  drafts?: CourseEditorDrafts;
  transcripts?: CourseEditorTranscripts;
  /** Admin-only sections, rendered after the course's own (e.g. granting access). */
  adminTools?: ReactNode;
};

type CourseCreateDialogProps = {
  actor: AuthenticatedUser;
  ownerOptions: readonly InstructorOption[] | null;
  onCreate: (draft: AdminCourseDraft) => Promise<CourseTree>;
  onSlugTaken: (slug: string) => Promise<CourseTree['id'] | undefined>;
  onAuthenticationRequired: () => void;
};

type DraftFormValues = {
  slug: string;
  title: string;
  summary: string;
  description: string;
  instructorId: string;
};

type StringFieldProps<Values extends FieldValues> = {
  control: Control<Values>;
  name: FieldPathByValue<Values, string>;
  label: string;
  maximum: number;
  required?: boolean;
  slug?: boolean;
  multiline?: boolean;
  isDisabled?: boolean;
  description?: string;
};

function StringField<Values extends FieldValues>({
  control,
  name,
  label,
  maximum,
  required = false,
  slug = false,
  multiline = false,
  isDisabled = false,
  description,
}: StringFieldProps<Values>) {
  const { t } = useTranslation();
  const { field, fieldState } = useController({
    control,
    name,
    rules: {
      validate: (raw) => {
        const value = String(raw);
        if (required && value.trim().length === 0) return t('administration.validation.required');
        if ([...value].length > maximum) return t('administration.validation.maximum', { maximum });
        if (value.includes('\u0000')) return t('administration.validation.invalidText');
        if (slug && !/^[a-z0-9]+(?:-[a-z0-9]+)*$/.test(value))
          return t('administration.validation.slug');
        return true;
      },
    },
  });
  const props = {
    label,
    ref: field.ref,
    value: String(field.value),
    name: field.name,
    onChange: field.onChange,
    onBlur: field.onBlur,
    isDisabled,
    isRequired: required,
    isInvalid: fieldState.invalid,
    ...(fieldState.error?.message === undefined ? {} : { errorMessage: fieldState.error.message }),
    ...(description === undefined ? {} : { description }),
  };
  return multiline ? <TextArea {...props} /> : <TextField {...props} />;
}

const courseTextFields = ['title', 'slug', 'summary', 'description'] as const;
const courseFieldLabels = {
  title: 'administration.courseTitle',
  slug: 'administration.slug',
  summary: 'administration.summary',
  description: 'administration.description',
} as const;

function FormErrorSummary<Values extends FieldValues>({
  form,
  fields,
  onSelect,
}: {
  form: UseFormReturn<Values>;
  fields: Array<{ name: FieldPath<Values>; label: string }>;
  /** Overrides the default focus so the parent can reveal a field on a hidden tab first. */
  onSelect?: (name: FieldPath<Values>) => void;
}) {
  const { t } = useTranslation();
  const summary = useRef<HTMLDivElement>(null);
  const { errors, submitCount } = form.formState;
  const invalidFields = fields.filter(
    ({ name }) => form.getFieldState(name, { ...form.formState, errors }).invalid,
  );
  useEffect(() => {
    if (submitCount > 0) summary.current?.focus();
  }, [submitCount]);
  if (submitCount === 0 || invalidFields.length === 0) return null;
  return (
    <div ref={summary} className="ui-error-summary" role="alert" tabIndex={-1}>
      <strong>{t('administration.validationSummary')}</strong>
      <ul>
        {invalidFields.map(({ name, label }) => (
          <li key={name}>
            <Link onPress={() => (onSelect ? onSelect(name) : form.setFocus(name))}>{label}</Link>
          </li>
        ))}
      </ul>
    </div>
  );
}

function formValues(course: CourseTree): CourseFormValues {
  return {
    slug: course.slug,
    title: course.title,
    summary: course.summary,
    description: course.description,
    instructorId: course.instructorId,
    modules: course.modules.map((module) => ({
      id: module.id,
      title: module.title,
      lessons: module.lessons.map((lesson) => ({ ...lesson })),
    })),
  };
}

function ownerName(
  instructorId: string,
  actor: AuthenticatedUser,
  options: readonly InstructorOption[] | null,
) {
  return (
    options?.find((option) => option.id === instructorId)?.displayName ??
    (actor.id === instructorId ? actor.displayName : instructorId)
  );
}

function errorMessage(code: AuthoringErrorCode, t: ReturnType<typeof useTranslation>['t']) {
  switch (code) {
    case 'authentication_required':
      return t('administration.error.authenticationRequired');
    case 'origin_forbidden':
      return t('administration.error.originForbidden');
    case 'forbidden':
    case 'course_forbidden':
      return t('administration.error.forbidden');
    case 'slug_taken':
      return t('administration.error.slugTaken');
    case 'slug_immutable':
      return t('administration.error.slugImmutable');
    case 'invalid_owner':
      return t('administration.error.invalidOwner');
    case 'invalid_status_transition':
      return t('administration.error.invalidStatusTransition');
    case 'course_not_deletable':
      return t('administration.error.courseNotDeletable');
    case 'lesson_limit':
      return t('administration.error.lessonLimit');
    case 'course_modified':
      return t('administration.conflict');
    case 'course_not_found':
      return t('administration.error.courseNotFound');
    case 'invalid_request':
      return t('administration.error.invalidRequest');
    case 'request_cancelled':
      return t('administration.error.requestCancelled');
    case 'request_timeout':
      return t('administration.error.requestTimeout');
    case 'dependency_unavailable':
    case 'invalid_response':
      return t('administration.error.unavailable');
  }
}

function errorCode(error: unknown): AuthoringErrorCode | undefined {
  return error instanceof AuthoringQueryError ? error.code : undefined;
}

function OwnerField({
  form,
  actor,
  options,
}: {
  form: UseFormReturn<CourseFormValues>;
  actor: AuthenticatedUser;
  options: readonly InstructorOption[] | null;
}) {
  const { t } = useTranslation();
  const instructorId = form.watch('instructorId');
  if (actor.role !== 'admin' || options === null || options.length === 0)
    return (
      <div className="admin-readonly-field">
        <span>{t('administration.owner')}</span>
        <strong>{ownerName(instructorId, actor, options)}</strong>
        {actor.role === 'admin' && <small>{t('administration.ownerPickerUnavailable')}</small>}
      </div>
    );

  return (
    <Select
      label={t('administration.owner')}
      options={options.map((option) => ({ id: option.id, label: option.displayName }))}
      selectedKey={instructorId}
      onSelectionChange={(key) => form.setValue('instructorId', String(key), { shouldDirty: true })}
      placeholder={t('ui.selectPlaceholder')}
    />
  );
}

/**
 * A draft segment carries no `documentIds` — ASR has no way to know about the course's document
 * library — so a loaded draft always starts with none, exactly like a line the author adds by
 * hand.
 */
function toTranscriptSegment(segment: DraftSegment) {
  return { ...segment, documentIds: [] };
}

function PreviewField({
  control,
  name,
  label,
}: {
  control: Control<CourseFormValues>;
  name: FieldPathByValue<CourseFormValues, boolean>;
  label: string;
}) {
  const { field } = useController({ control, name });
  return (
    <Checkbox isSelected={field.value} onChange={field.onChange} onBlur={field.onBlur}>
      {label}
    </Checkbox>
  );
}

function LessonFields({
  form,
  moduleIndex,
  savedLessonIds,
  video,
  drafts,
  transcripts,
  onTranscriptKnown,
}: {
  form: UseFormReturn<CourseFormValues>;
  moduleIndex: number;
  savedLessonIds: ReadonlySet<string>;
  video: CourseEditorVideo | undefined;
  drafts: CourseEditorDrafts | undefined;
  transcripts: CourseEditorTranscripts | undefined;
  onTranscriptKnown: (lessonId: string, hasTranscript: boolean) => void;
}) {
  const { t } = useTranslation();
  const lessons = useFieldArray({
    control: form.control,
    name: `modules.${moduleIndex}.lessons`,
    keyName: 'formKey',
  });
  const lessonCount = form
    .watch('modules')
    .reduce((total, module) => total + module.lessons.length, 0);

  // A draft the author asked to load, keyed by lesson: `LessonDraftField` and
  // `LessonTranscriptField` are siblings below, and this is the shortest-lived state that can
  // pass one's callback to the other's prop without either knowing about the other.
  const [loadedDrafts, setLoadedDrafts] = useState<Record<string, TranscriptDraftOffer>>({});
  const [draftTokens, setDraftTokens] = useState<Record<string, string | undefined>>({});

  return (
    <div className="admin-lessons">
      {lessons.fields.map((lesson, lessonIndex) => {
        const draftOffer = loadedDrafts[lesson.id];
        const currentDraftUpdatedAt = draftTokens[lesson.id];
        return (
          <fieldset className="admin-lesson" key={lesson.formKey}>
            <legend>{t('administration.lessonNumber', { number: lessonIndex + 1 })}</legend>
            <div className="admin-editor-grid">
              <StringField
                control={form.control}
                name={`modules.${moduleIndex}.lessons.${lessonIndex}.title`}
                label={t('administration.lessonTitle', {
                  lesson: lessonIndex + 1,
                  module: moduleIndex + 1,
                })}
                maximum={200}
                required
              />
              <PreviewField
                control={form.control}
                name={`modules.${moduleIndex}.lessons.${lessonIndex}.isPreview`}
                label={t('administration.lessonPreview', {
                  lesson: lessonIndex + 1,
                  module: moduleIndex + 1,
                })}
              />
            </div>
            {video !== undefined && (
              <LessonVideoField
                lessonId={lesson.id}
                isSaved={savedLessonIds.has(lesson.id)}
                {...video}
              />
            )}
            {drafts !== undefined && savedLessonIds.has(lesson.id) && (
              <LessonDraftField
                lessonId={lesson.id}
                {...drafts}
                onLoadDraft={(segments, language, updatedAt) =>
                  setLoadedDrafts((current) => ({
                    ...current,
                    [lesson.id]: {
                      segments: segments.map(toTranscriptSegment),
                      language,
                      updatedAt,
                    },
                  }))
                }
                onDraftTokenChange={(updatedAt) =>
                  setDraftTokens((current) =>
                    current[lesson.id] === updatedAt
                      ? current
                      : { ...current, [lesson.id]: updatedAt },
                  )
                }
              />
            )}
            {transcripts !== undefined && savedLessonIds.has(lesson.id) && (
              <LessonTranscriptField
                lessonId={lesson.id}
                {...transcripts}
                {...(draftOffer === undefined ? {} : { draftOffer })}
                {...(currentDraftUpdatedAt === undefined
                  ? {}
                  : { draftUpdatedAt: currentDraftUpdatedAt })}
                onTranscriptKnown={onTranscriptKnown}
                onDraftOfferHandled={() =>
                  setLoadedDrafts((current) => {
                    if (current[lesson.id] === undefined) return current;
                    const next = { ...current };
                    delete next[lesson.id];
                    return next;
                  })
                }
              />
            )}
            <div className="admin-row-actions">
              <Button
                type="button"
                variant="secondary"
                isDisabled={lessonIndex === 0}
                onPress={() => lessons.move(lessonIndex, lessonIndex - 1)}
              >
                {t('administration.moveLessonUp', {
                  lesson: lessonIndex + 1,
                  module: moduleIndex + 1,
                })}
              </Button>
              <Button
                type="button"
                variant="secondary"
                isDisabled={lessonIndex === lessons.fields.length - 1}
                onPress={() => lessons.move(lessonIndex, lessonIndex + 1)}
              >
                {t('administration.moveLessonDown', {
                  lesson: lessonIndex + 1,
                  module: moduleIndex + 1,
                })}
              </Button>
              <Button type="button" variant="danger" onPress={() => lessons.remove(lessonIndex)}>
                {t('administration.deleteLesson', {
                  lesson: lessonIndex + 1,
                  module: moduleIndex + 1,
                })}
              </Button>
            </div>
          </fieldset>
        );
      })}
      <Button
        type="button"
        variant="secondary"
        isDisabled={lessonCount >= 120}
        onPress={() => lessons.append({ id: newLessonId(), title: '', isPreview: false })}
      >
        {t('administration.addLesson', { number: moduleIndex + 1 })}
      </Button>
    </div>
  );
}

function ModuleFields({
  form,
  savedLessonIds,
  video,
  drafts,
  transcripts,
  onTranscriptKnown,
}: {
  form: UseFormReturn<CourseFormValues>;
  savedLessonIds: ReadonlySet<string>;
  video: CourseEditorVideo | undefined;
  drafts: CourseEditorDrafts | undefined;
  transcripts: CourseEditorTranscripts | undefined;
  onTranscriptKnown: (lessonId: string, hasTranscript: boolean) => void;
}) {
  const { t } = useTranslation();
  const modules = useFieldArray({ control: form.control, name: 'modules', keyName: 'formKey' });

  return (
    <section className="admin-program ui-card" aria-labelledby="course-program-heading">
      <div className="admin-section-heading">
        <div>
          <p className="eyebrow">{t('administration.structure')}</p>
          <h2 id="course-program-heading">{t('administration.program')}</h2>
        </div>
        <span>{t('administration.moduleCount', { count: modules.fields.length })}</span>
      </div>
      <div className="admin-modules">
        {modules.fields.map((module, moduleIndex) => (
          <fieldset className="admin-module" key={module.formKey}>
            <legend>{t('administration.moduleNumber', { number: moduleIndex + 1 })}</legend>
            <StringField
              control={form.control}
              name={`modules.${moduleIndex}.title`}
              label={t('administration.moduleTitle', { number: moduleIndex + 1 })}
              maximum={200}
              required
            />
            <LessonFields
              form={form}
              moduleIndex={moduleIndex}
              savedLessonIds={savedLessonIds}
              video={video}
              drafts={drafts}
              transcripts={transcripts}
              onTranscriptKnown={onTranscriptKnown}
            />
            <div className="admin-row-actions">
              <Button
                type="button"
                variant="secondary"
                isDisabled={moduleIndex === 0}
                onPress={() => modules.move(moduleIndex, moduleIndex - 1)}
              >
                {t('administration.moveModuleUp', { number: moduleIndex + 1 })}
              </Button>
              <Button
                type="button"
                variant="secondary"
                isDisabled={moduleIndex === modules.fields.length - 1}
                onPress={() => modules.move(moduleIndex, moduleIndex + 1)}
              >
                {t('administration.moveModuleDown', { number: moduleIndex + 1 })}
              </Button>
              <Button type="button" variant="danger" onPress={() => modules.remove(moduleIndex)}>
                {t('administration.deleteModule', { number: moduleIndex + 1 })}
              </Button>
            </div>
          </fieldset>
        ))}
      </div>
      <Button
        type="button"
        variant="secondary"
        isDisabled={modules.fields.length >= 40}
        onPress={() => modules.append({ id: newModuleId(), title: '', lessons: [] })}
      >
        {t('administration.addModule')}
      </Button>
    </section>
  );
}

function nextStatus(status: CourseStatus): CourseStatus {
  return status === 'published' ? 'archived' : 'published';
}

function lessonIds(course: CourseTree): Set<string> {
  return new Set(course.modules.flatMap((module) => module.lessons.map((lesson) => lesson.id)));
}

export function CourseEditorView({
  course,
  actor,
  ownerOptions,
  onSave,
  onSetPrice,
  onTransition,
  onDelete,
  onAuthenticationRequired,
  video,
  drafts,
  transcripts,
  adminTools,
}: CourseEditorViewProps) {
  const { t } = useTranslation();
  const [message, setMessage] = useState<string>();
  const [messageKind, setMessageKind] = useState<'status' | 'alert'>('status');
  const [lifecyclePending, setLifecyclePending] = useState(false);
  const [transcriptKnowledge, setTranscriptKnowledge] = useState<Record<string, boolean>>({});
  const [savedLessonIds, setSavedLessonIds] = useState(() => lessonIds(course));
  const [tab, setTab] = useState<'overview' | 'lessons' | 'access'>('overview');
  const form = useForm<CourseFormValues>({
    defaultValues: formValues(course),
    mode: 'onBlur',
    shouldFocusError: false,
  });
  const summaryFields: Array<{ name: FieldPath<CourseFormValues>; label: string }> =
    courseTextFields.map((name) => ({ name, label: t(courseFieldLabels[name]) }));
  const currentModules = useWatch({ control: form.control, name: 'modules' });
  currentModules.forEach((module, moduleIndex) => {
    summaryFields.push({
      name: `modules.${moduleIndex}.title`,
      label: t('administration.moduleTitle', { number: moduleIndex + 1 }),
    });
    module.lessons.forEach((_lesson, lessonIndex) => {
      summaryFields.push({
        name: `modules.${moduleIndex}.lessons.${lessonIndex}.title`,
        label: t('administration.lessonTitle', {
          lesson: lessonIndex + 1,
          module: moduleIndex + 1,
        }),
      });
    });
  });
  const resetVersion = useRef(course.updatedAt);

  useEffect(() => {
    if (resetVersion.current === course.updatedAt) return;
    resetVersion.current = course.updatedAt;
    form.reset(formValues(course));
    setSavedLessonIds(lessonIds(course));
    setTranscriptKnowledge({});
  }, [course, form]);

  const onTranscriptKnown = useCallback((lessonId: string, hasTranscript: boolean) => {
    setTranscriptKnowledge((current) =>
      current[lessonId] === hasTranscript ? current : { ...current, [lessonId]: hasTranscript },
    );
  }, []);
  const hasTranscriptNotice =
    course.status === 'draft' &&
    savedLessonIds.size > 0 &&
    (transcripts === undefined ||
      [...savedLessonIds].some(
        (lessonId) => transcriptKnowledge[lessonId] === undefined || !transcriptKnowledge[lessonId],
      ));

  function showError(error: unknown) {
    const code = errorCode(error);
    setMessageKind('alert');
    setMessage(code === undefined ? t('administration.error.unavailable') : errorMessage(code, t));
    if (code === 'authentication_required') onAuthenticationRequired();
  }

  async function save(values: CourseFormValues) {
    setMessage(undefined);
    try {
      const saved = await onSave(course, values);
      form.reset(formValues(saved));
      setSavedLessonIds(lessonIds(saved));
      setMessageKind('status');
      setMessage(t('administration.saved'));
    } catch (error) {
      showError(error);
    }
  }

  async function setPrice(current: CourseTree, price: CoursePrice | null) {
    if (onSetPrice === undefined) return;
    setMessage(undefined);
    try {
      await onSetPrice(current, price);
      setMessageKind('status');
      setMessage(t(price === null ? 'administration.price.removed' : 'administration.price.saved'));
    } catch (error) {
      showError(error);
    }
  }

  async function transition() {
    setLifecyclePending(true);
    setMessage(undefined);
    try {
      const changed = await onTransition(course, nextStatus(course.status));
      form.reset(formValues(changed));
      setMessageKind('status');
      setMessage(t('administration.statusChanged'));
    } catch (error) {
      showError(error);
    } finally {
      setLifecyclePending(false);
    }
  }

  function focusField(name: FieldPath<CourseFormValues>) {
    // Module and lesson fields live on the Lessons tab; details on Overview. An error-summary
    // link may point at a field on the tab that is not showing, so switch there synchronously
    // (flushSync un-hides the panel before focus) — otherwise setFocus lands on a hidden input.
    const target = String(name).startsWith('modules') ? 'lessons' : 'overview';
    if (target !== tab) flushSync(() => setTab(target));
    form.setFocus(name);
  }

  async function removeCourse() {
    setLifecyclePending(true);
    setMessage(undefined);
    try {
      await onDelete(course);
    } catch (error) {
      showError(error);
      setLifecyclePending(false);
    }
  }

  const statusPill = (
    <span
      className={`ui-pill ui-pill--${statusVariant[course.status]}`}
      data-status={course.status}
    >
      {t(`administration.${course.status}`)}
    </span>
  );
  const lessonCount = course.modules.reduce((total, module) => total + module.lessons.length, 0);
  const transitionLabel =
    course.status === 'published'
      ? t('administration.archive')
      : course.status === 'archived'
        ? t('administration.restore')
        : t('administration.publish');

  return (
    <article className="admin-editor">
      <Link href="/admin">{t('administration.backToCourses')}</Link>
      <PageHeader
        eyebrow={t('administration.editorEyebrow')}
        title={course.title}
        description={t('administration.editorLede')}
        actions={
          <>
            {statusPill}
            {course.status === 'published' && (
              <Link href={`/courses/${course.slug}`}>{t('administration.previewCourse')}</Link>
            )}
          </>
        }
      />
      <SegmentedControl
        label={t('administration.tabsLabel')}
        options={[
          { id: 'overview', label: t('administration.tabOverview') },
          { id: 'lessons', label: t('administration.tabLessons') },
          { id: 'access', label: t('administration.tabAccess') },
        ]}
        selectedKeys={new Set([tab])}
        onSelectionChange={(keys) => {
          const next = String([...keys][0] ?? 'overview');
          if (next === 'overview' || next === 'lessons' || next === 'access') setTab(next);
        }}
      />
      {message && (
        <div className="admin-editor-message" role={messageKind}>
          {message}
        </div>
      )}
      <form
        className="admin-editor-form"
        onSubmit={form.handleSubmit(save)}
        noValidate
        aria-busy={form.formState.isSubmitting}
        hidden={tab === 'access'}
      >
        <FormErrorSummary form={form} fields={summaryFields} onSelect={focusField} />
        <div className="course-form-layout" hidden={tab !== 'overview'}>
          <section
            className="admin-editor-details ui-card course-details-form"
            aria-labelledby="course-details-heading"
          >
            <div className="card-heading">
              <h2 id="course-details-heading">{t('administration.courseDetails')}</h2>
            </div>
            <div className="admin-editor-grid">
              <StringField
                control={form.control}
                name="title"
                label={t('administration.courseTitle')}
                maximum={200}
                required
              />
              <StringField
                control={form.control}
                name="slug"
                label={t('administration.slug')}
                maximum={120}
                required
                slug
                isDisabled={course.publishedAt !== null}
                {...(course.publishedAt === null
                  ? {}
                  : { description: t('administration.slugLocked') })}
              />
              <StringField
                control={form.control}
                name="summary"
                label={t('administration.summary')}
                maximum={500}
                required
                multiline
              />
              <StringField
                control={form.control}
                name="description"
                label={t('administration.description')}
                maximum={20000}
                required
                multiline
              />
              <OwnerField form={form} actor={actor} options={ownerOptions} />
            </div>
          </section>
          <div className="course-side-panels">
            <section className="ui-card" aria-labelledby="course-lifecycle-heading">
              <div className="card-heading">
                <h2 id="course-lifecycle-heading">{t('administration.courseStatus')}</h2>
                {statusPill}
              </div>
              <p>{t('administration.statusHint')}</p>
              {hasTranscriptNotice && (
                <output className="admin-editor-message">
                  {t('administration.transcriptsPending')}
                </output>
              )}
              <Button
                type="button"
                variant="secondary"
                isDisabled={lifecyclePending || form.formState.isSubmitting}
                onPress={() => void transition()}
              >
                {transitionLabel}
              </Button>
            </section>
            <section className="ui-card" aria-labelledby="course-stats-heading">
              <div className="card-heading">
                <h2 id="course-stats-heading">{t('administration.miniStatsLabel')}</h2>
              </div>
              <div className="course-mini-stats">
                <div>
                  <span>{t('administration.statLessons')}</span>
                  <strong>{lessonCount}</strong>
                </div>
                <div>
                  <span>{t('administration.students')}</span>
                  <strong>{course.studentCount}</strong>
                </div>
                <div>
                  <span>{t('administration.updated')}</span>
                  <strong>{course.updatedAt.slice(0, 10)}</strong>
                </div>
              </div>
            </section>
            {course.status === 'draft' && (
              <section className="ui-card danger-panel" aria-labelledby="course-danger-heading">
                <p className="eyebrow">{t('administration.dangerZone')}</p>
                <h2 id="course-danger-heading">{t('administration.deleteCourse')}</h2>
                <p>{t('administration.dangerHint')}</p>
                <Dialog
                  triggerLabel={t('administration.deleteCourse')}
                  trigger={
                    <Button
                      type="button"
                      variant="danger"
                      isDisabled={lifecyclePending || form.formState.isSubmitting}
                    >
                      {t('administration.deleteCourse')}
                    </Button>
                  }
                  title={t('administration.deleteCourse')}
                  closeLabel={t('common.cancel')}
                >
                  {(close) => (
                    <Button
                      variant="danger"
                      onPress={() => {
                        void removeCourse();
                        close();
                      }}
                    >
                      {t('administration.confirmDeleteCourse')}
                    </Button>
                  )}
                </Dialog>
              </section>
            )}
          </div>
        </div>
        <div className="course-lessons-panel" hidden={tab !== 'lessons'}>
          {transcripts !== undefined && (
            <DocumentLibrary documents={transcripts.documents} actions={transcripts.library} />
          )}
          <ModuleFields
            form={form}
            savedLessonIds={savedLessonIds}
            video={video}
            drafts={drafts}
            transcripts={transcripts}
            onTranscriptKnown={onTranscriptKnown}
          />
        </div>
        <div className="admin-editor-save">
          <Button type="submit" isDisabled={form.formState.isSubmitting || lifecyclePending}>
            {form.formState.isSubmitting ? t('administration.saving') : t('administration.save')}
          </Button>
        </div>
      </form>
      {onSetPrice !== undefined && (
        <div hidden={tab !== 'overview'}>
          <CoursePriceField
            key={course.price?.minor ?? 'free'}
            course={course}
            hasUnsavedChanges={form.formState.isDirty}
            isBusy={lifecyclePending || form.formState.isSubmitting}
            onSetPrice={setPrice}
          />
        </div>
      )}
      <div hidden={tab !== 'access'}>{adminTools}</div>
    </article>
  );
}

function DraftOwnerField({
  form,
  actor,
  options,
}: {
  form: UseFormReturn<DraftFormValues>;
  actor: AuthenticatedUser;
  options: readonly InstructorOption[] | null;
}) {
  const { t } = useTranslation();
  const instructorId = form.watch('instructorId');
  if (actor.role !== 'admin' || options === null || options.length === 0)
    return (
      <div className="admin-readonly-field">
        <span>{t('administration.owner')}</span>
        <strong>{actor.displayName}</strong>
        {actor.role === 'admin' && <small>{t('administration.ownerPickerUnavailable')}</small>}
      </div>
    );
  return (
    <Select
      label={t('administration.owner')}
      options={options.map((option) => ({ id: option.id, label: option.displayName }))}
      selectedKey={instructorId}
      onSelectionChange={(key) => form.setValue('instructorId', String(key))}
      placeholder={t('ui.selectPlaceholder')}
    />
  );
}

export function CourseCreateDialog({
  actor,
  ownerOptions,
  onCreate,
  onSlugTaken,
  onAuthenticationRequired,
}: CourseCreateDialogProps) {
  const { t } = useTranslation();
  const [message, setMessage] = useState<string>();
  const [existingCourseId, setExistingCourseId] = useState<CourseTree['id']>();
  const form = useForm<DraftFormValues>({
    mode: 'onBlur',
    shouldFocusError: false,
    defaultValues: {
      slug: '',
      title: '',
      summary: '',
      description: '',
      instructorId: actor.id,
    },
  });

  async function create(values: DraftFormValues) {
    setMessage(undefined);
    setExistingCourseId(undefined);
    const fields = {
      slug: values.slug,
      title: values.title,
      summary: values.summary,
      description: values.description,
    };
    const draft =
      actor.role === 'admin' ? { ...fields, instructorId: values.instructorId } : fields;
    try {
      await onCreate(draft);
    } catch (error) {
      const code = errorCode(error);
      if (code === 'slug_taken') {
        setExistingCourseId(await onSlugTaken(values.slug));
        setMessage(t('administration.createConflict'));
      } else {
        setMessage(
          code === undefined ? t('administration.error.unavailable') : errorMessage(code, t),
        );
      }
      if (code === 'authentication_required') onAuthenticationRequired();
    }
  }

  return (
    <Dialog
      triggerLabel={t('administration.createCourse')}
      triggerVariant="primary"
      title={t('administration.createCourse')}
      closeLabel={t('common.close')}
    >
      {message && (
        <div className="admin-editor-message" role="alert">
          <p>{message}</p>
          {existingCourseId && (
            <Link href={`/admin/courses/${existingCourseId}`}>
              {t('administration.openExisting')}
            </Link>
          )}
        </div>
      )}
      <form
        className="admin-create-form"
        onSubmit={form.handleSubmit(create)}
        noValidate
        aria-busy={form.formState.isSubmitting}
      >
        <FormErrorSummary
          form={form}
          fields={courseTextFields.map((name) => ({ name, label: t(courseFieldLabels[name]) }))}
        />
        <StringField
          control={form.control}
          name="title"
          label={t('administration.courseTitle')}
          maximum={200}
          required
        />
        <StringField
          control={form.control}
          name="slug"
          label={t('administration.slug')}
          maximum={120}
          required
          slug
        />
        <StringField
          control={form.control}
          name="summary"
          label={t('administration.summary')}
          maximum={500}
          required
          multiline
        />
        <StringField
          control={form.control}
          name="description"
          label={t('administration.description')}
          maximum={20000}
          required
          multiline
        />
        <DraftOwnerField form={form} actor={actor} options={ownerOptions} />
        <Button type="submit" isDisabled={form.formState.isSubmitting}>
          {form.formState.isSubmitting ? t('administration.creating') : t('administration.create')}
        </Button>
      </form>
    </Dialog>
  );
}

export function CourseEditorPending() {
  const { t } = useTranslation();
  return <Skeleton label={t('administration.loadingCourse')} lines={7} />;
}
