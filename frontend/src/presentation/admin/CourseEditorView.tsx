import { useEffect, useRef, useState } from 'react';
import { useFieldArray, useForm, useController } from 'react-hook-form';
import type { Control, FieldPathByValue, FieldValues, UseFormReturn } from 'react-hook-form';
import { useTranslation } from 'react-i18next';
import { AuthoringQueryError } from '../../app/composition/admin-queries';
import { newLessonId, newModuleId } from '../../domain/course/course';
import type {
  AdminCourseDraft,
  AuthoringErrorCode,
  CourseStatus,
  CourseTree,
  CourseTreeEdit,
  InstructorOption,
} from '../../domain/course/course';
import type { AuthenticatedUser } from '../../domain/identity/identity';
import { Button, Dialog, Link, Select, Skeleton, TextArea, TextField } from '../ui';
import { LessonVideoField } from './LessonVideoField';
import type { CourseEditorVideo } from './LessonVideoField';

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
      description: string;
    }>;
  }>;
};

type CourseEditorViewProps = {
  course: CourseTree;
  actor: AuthenticatedUser;
  ownerOptions: readonly InstructorOption[] | null;
  onSave: (current: CourseTree, edit: CourseTreeEdit) => Promise<CourseTree>;
  onTransition: (current: CourseTree, status: CourseStatus) => Promise<CourseTree>;
  onDelete: (current: CourseTree) => Promise<void>;
  onAuthenticationRequired: () => void;
  video?: CourseEditorVideo;
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

function LessonFields({
  form,
  moduleIndex,
  savedLessonIds,
  video,
}: {
  form: UseFormReturn<CourseFormValues>;
  moduleIndex: number;
  savedLessonIds: ReadonlySet<string>;
  video: CourseEditorVideo | undefined;
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

  return (
    <div className="admin-lessons">
      {lessons.fields.map((lesson, lessonIndex) => (
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
            <StringField
              control={form.control}
              name={`modules.${moduleIndex}.lessons.${lessonIndex}.description`}
              label={t('administration.lessonDescription')}
              maximum={1000}
              multiline
            />
          </div>
          {video !== undefined && (
            <LessonVideoField
              lessonId={lesson.id}
              isSaved={savedLessonIds.has(lesson.id)}
              {...video}
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
      ))}
      <Button
        type="button"
        variant="secondary"
        isDisabled={lessonCount >= 120}
        onPress={() => lessons.append({ id: newLessonId(), title: '', description: '' })}
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
}: {
  form: UseFormReturn<CourseFormValues>;
  savedLessonIds: ReadonlySet<string>;
  video: CourseEditorVideo | undefined;
}) {
  const { t } = useTranslation();
  const modules = useFieldArray({ control: form.control, name: 'modules', keyName: 'formKey' });

  return (
    <section className="admin-program" aria-labelledby="course-program-heading">
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
  onTransition,
  onDelete,
  onAuthenticationRequired,
  video,
}: CourseEditorViewProps) {
  const { t } = useTranslation();
  const [message, setMessage] = useState<string>();
  const [messageKind, setMessageKind] = useState<'status' | 'alert'>('status');
  const [lifecyclePending, setLifecyclePending] = useState(false);
  const [savedLessonIds, setSavedLessonIds] = useState(() => lessonIds(course));
  const form = useForm<CourseFormValues>({ defaultValues: formValues(course) });
  const resetVersion = useRef(course.updatedAt);

  useEffect(() => {
    if (resetVersion.current === course.updatedAt) return;
    resetVersion.current = course.updatedAt;
    form.reset(formValues(course));
    setSavedLessonIds(lessonIds(course));
  }, [course, form]);

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

  return (
    <article className="admin-editor">
      <Link href="/admin">{t('administration.backToCourses')}</Link>
      <header className="admin-editor-heading">
        <div>
          <p className="eyebrow">{t('administration.editor')}</p>
          <h1>{course.title}</h1>
        </div>
        <span className="admin-course-status" data-status={course.status}>
          {t(`administration.${course.status}`)}
        </span>
      </header>
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
      >
        <section className="admin-editor-details" aria-labelledby="course-details-heading">
          <div className="admin-section-heading">
            <div>
              <p className="eyebrow">{t('administration.details')}</p>
              <h2 id="course-details-heading">{t('administration.courseDetails')}</h2>
            </div>
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
        <ModuleFields form={form} savedLessonIds={savedLessonIds} video={video} />
        <div className="admin-editor-save">
          <Button type="submit" isDisabled={form.formState.isSubmitting || lifecyclePending}>
            {form.formState.isSubmitting ? t('administration.saving') : t('administration.save')}
          </Button>
        </div>
      </form>
      <section className="admin-lifecycle" aria-labelledby="course-lifecycle-heading">
        <div>
          <p className="eyebrow">{t('administration.lifecycle')}</p>
          <h2 id="course-lifecycle-heading">{t('administration.courseStatus')}</h2>
        </div>
        <div className="admin-row-actions">
          <Button
            type="button"
            variant="secondary"
            isDisabled={lifecyclePending || form.formState.isSubmitting}
            onPress={() => void transition()}
          >
            {course.status === 'published'
              ? t('administration.archive')
              : course.status === 'archived'
                ? t('administration.restore')
                : t('administration.publish')}
          </Button>
          {course.status === 'draft' && (
            <Button
              type="button"
              variant="danger"
              isDisabled={lifecyclePending || form.formState.isSubmitting}
              onPress={() => void removeCourse()}
            >
              {t('administration.deleteCourse')}
            </Button>
          )}
        </div>
      </section>
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
