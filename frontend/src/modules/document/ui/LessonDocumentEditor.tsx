import { useEffect, useRef, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import type {
  LessonColumn,
  LessonDocument,
  LessonImageContentType,
  LessonRichText,
  LessonSection,
} from '../domain/document';
import { imageServeUrl } from '../domain/document';
import { hasRenderableSections } from '../domain/document';
import {
  DocumentAuthoringError,
  ImageUploadError,
  type createAdminDocumentQueries,
} from '../application/admin-document-queries';
import { Button, FileUpload, SegmentedControl, TextField, Toast } from '../../../presentation/ui';
import { EmptyState, ErrorState } from '../../../presentation/ui/states';
import { LessonDocumentView } from './LessonDocumentView';
import { LessonDocumentEditorLoading } from './LessonDocumentEditorLoading';
import { LessonDocumentRichText } from './LessonDocumentRichText';

/** Lesson identity resolved from the course tree, shown in the title block. */
export type LessonMeta = {
  readonly lessonTitle: string;
  readonly courseTitle: string;
  readonly moduleLabel: string;
};

type AdminDocumentQueries = ReturnType<typeof createAdminDocumentQueries>;

type LessonDocumentEditorProps = {
  readonly lessonId: string;
  readonly documents: AdminDocumentQueries;
  readonly meta?: LessonMeta | null;
};

const actionIcons = {
  up: 'm6 10 6-6 6 6M12 4v16',
  down: 'm6 14 6 6 6-6M12 4v16',
  trash: 'M5 7h14M9 7V4h6v3M7 7l1 13h8l1-13M10 10v7m4-7v7',
  chevron: 'm9 5 7 7-7 7',
  editor: 'M4 5h16M12 5v15M8 20h8',
  preview: 'M2 12s4-7 10-7 10 7 10 7-4 7-10 7S2 12 2 12zM12 9a3 3 0 1 0 0 6 3 3 0 0 0 0-6',
  drag: 'M9 5h.01M15 5h.01M9 12h.01M15 12h.01M9 19h.01M15 19h.01',
} as const;

function ActionIcon({ path }: { readonly path: string }) {
  return (
    <svg
      viewBox="0 0 24 24"
      width="18"
      height="18"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.5"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
      focusable="false"
    >
      <path d={path} />
    </svg>
  );
}

type DraftState = {
  readonly sections: readonly LessonSection[];
  readonly token: string | null;
  readonly locale: 'uk' | 'en';
};

let localIdCounter = 0;
function localId(): string {
  localIdCounter += 1;
  return `s-${Date.now().toString(36)}-${localIdCounter}`;
}

const sectionTypes = ['video', 'text', 'image', 'row', 'task'] as const;

function SectionIcon({ type }: { readonly type: LessonSection['type'] }) {
  const paths: Record<LessonSection['type'], string> = {
    video: 'M9 6l9 6-9 6Z',
    text: 'M4 5h16M12 5v14M8 19h8',
    image: 'M3 3h18v18H3ZM3 17l6-6 4 4 3-3 5 5M7 7h1',
    row: 'M3 4h18v16H3ZM12 4v16',
    task: 'M9 6h12M9 12h12M9 18h12M3 6l1 1 2-2M3 12l1 1 2-2M3 18l1 1 2-2',
  };
  return (
    <span className="lesson-editor-icon-tile" aria-hidden="true">
      <svg
        viewBox="0 0 24 24"
        width="20"
        height="20"
        fill="none"
        stroke="currentColor"
        strokeWidth="1.5"
        strokeLinecap="round"
        strokeLinejoin="round"
        aria-hidden="true"
        focusable="false"
      >
        <path d={paths[type]} />
      </svg>
    </span>
  );
}

function blankSection(type: LessonSection['type']): LessonSection {
  const id = localId();
  switch (type) {
    case 'video':
      return { type: 'video', id };
    case 'text':
      return { type: 'text', id, content: { blocks: [] } };
    case 'task':
      return { type: 'task', id, content: { blocks: [] } };
    case 'image':
      return { type: 'image', id, imageId: '', contentType: 'image/png', alt: '' };
    case 'row':
      return {
        type: 'row',
        id,
        columns: [{ type: 'text', id: localId(), content: { blocks: [] } }],
      };
  }
}

export function LessonDocumentEditor({ lessonId, documents, meta }: LessonDocumentEditorProps) {
  const { t } = useTranslation();
  const client = useQueryClient();
  const draftQuery = useQuery(documents.document(lessonId));
  const save = useMutation(documents.save(client));
  const publish = useMutation(documents.publish(client));
  const upload = useMutation(documents.uploadImage());
  const [mode, setMode] = useState<'edit' | 'preview'>('edit');
  // Tracks whether the draft holds edits not yet written back, driving the save-status indicator.
  const [dirty, setDirty] = useState(false);
  // `edited` overlays the loaded draft once the author touches it; until then the editor renders
  // from the query. Deriving the draft (rather than seeding state in an effect) avoids a cascading
  // render, and a later refetch cannot clobber in-progress edits.
  const [edited, setEdited] = useState<DraftState | null>(null);
  const [uploadingKey, setUploadingKey] = useState<string | null>(null);
  const [status, setStatus] = useState<{ kind: 'saved' | 'published' } | null>(null);
  const [error, setError] = useState<string | null>(null);
  const editorRef = useRef<HTMLElement>(null);
  const pendingFocus = useRef<{ sectionId: string; selector: string } | { palette: true } | null>(
    null,
  );
  const [announcement, setAnnouncement] = useState({ text: '', sequence: 0 });

  const base: DraftState | null =
    draftQuery.data === undefined
      ? null
      : {
          sections: draftQuery.data.document.sections,
          token: draftQuery.data.updatedAt,
          locale: draftQuery.data.document.locale,
        };
  const draft = edited ?? base;

  const sections = draft?.sections;
  useEffect(() => {
    const target = pendingFocus.current;
    const root = editorRef.current;
    if (target === null || root === null) return;
    const focusTarget = () => {
      const control =
        'palette' in target
          ? root.querySelector<HTMLElement>('[data-add-section="text"]')
          : root
              .querySelector(`[data-section-id="${CSS.escape(target.sectionId)}"]`)
              ?.querySelector<HTMLElement>(target.selector);
      if (control == null) return false;
      control.focus();
      pendingFocus.current = null;
      return true;
    };
    if (focusTarget()) return;
    // Tiptap mounts its textbox after the section when immediatelyRender is false.
    const observer = new MutationObserver(() => {
      if (focusTarget()) observer.disconnect();
    });
    observer.observe(root, { childList: true, subtree: true });
    return () => observer.disconnect();
  }, [sections]);

  const hasVideo = draft?.sections.some((section) => section.type === 'video') ?? false;

  if (draftQuery.isError)
    return (
      <section className="lesson-editor" aria-labelledby="lesson-editor-heading">
        <header className="lesson-editor-header">
          <h1 id="lesson-editor-heading">{t('administration.document.heading')}</h1>
        </header>
        <ErrorState
          title={t('administration.document.loadError')}
          onRetry={() => void draftQuery.refetch()}
          retryLabel={t('administration.document.retry')}
        />
      </section>
    );
  if (draft === null) return <LessonDocumentEditorLoading />;

  const setSections = (
    next:
      | readonly LessonSection[]
      | ((sections: readonly LessonSection[]) => readonly LessonSection[]),
  ) => {
    setDirty(true);
    setEdited((cur) => {
      const b = cur ?? draft;
      return { ...b, sections: typeof next === 'function' ? next(b.sections) : next };
    });
  };

  const applyImage: SectionBodyProps['applyImage'] = (sectionId, columnId, imageId, contentType) =>
    setSections((prev) =>
      prev.map((section) => {
        if (section.id !== sectionId) return section;
        if (section.type === 'image' && columnId === null)
          return { ...section, imageId, contentType };
        if (section.type === 'row')
          return {
            ...section,
            columns: section.columns.map((column) =>
              column.id === columnId && column.type === 'image'
                ? { ...column, imageId, contentType }
                : column,
            ),
          };
        return section;
      }),
    );

  const updateSection = (index: number, section: LessonSection) =>
    setSections(draft.sections.map((existing, i) => (i === index ? section : existing)));

  const addSection = (type: LessonSection['type']) => {
    const section = blankSection(type);
    pendingFocus.current = {
      sectionId: section.id,
      selector:
        type === 'text' || type === 'task'
          ? '[role="textbox"]'
          : type === 'image'
            ? // The visible upload button carries .ui-button; RAC's hidden DropZone button (first in
              // DOM, aria-label="DropZone") does not, so target .ui-button to skip it.
              '.lesson-editor-image .ui-button'
            : '.lesson-editor-section-actions button',
    };
    setSections([...draft.sections, section]);
  };

  const removeSection = (index: number) => {
    const neighbor = draft.sections[index - 1] ?? draft.sections[index + 1];
    pendingFocus.current =
      neighbor === undefined
        ? { palette: true }
        : { sectionId: neighbor.id, selector: '.lesson-editor-section-actions button' };
    setSections(draft.sections.filter((_, i) => i !== index));
  };

  const announce = (text: string) =>
    setAnnouncement((previous) => ({ text, sequence: previous.sequence + 1 }));

  // Move keeps the button enabled at the edges (no isDisabled, which would drop focus to the body);
  // a move past the edge is simply a no-op.
  const moveSection = (index: number, delta: -1 | 1) => {
    const target = index + delta;
    if (target < 0 || target >= draft.sections.length) {
      announce(
        t('administration.document.moveBlocked', {
          edge: t(
            target < 0 ? 'administration.document.edgeFirst' : 'administration.document.edgeLast',
          ),
        }),
      );
      return;
    }
    const next = [...draft.sections];
    const moved = next[index];
    const swapped = next[target];
    if (moved === undefined || swapped === undefined) return;
    next[index] = swapped;
    next[target] = moved;
    setSections(next);
    announce(t('administration.document.movedTo', { pos: target + 1, total: next.length }));
  };

  const uploadInto = (
    key: string,
    file: File,
    apply: (imageId: string, contentType: LessonImageContentType) => void,
  ) => {
    setError(null);
    setUploadingKey(key);
    upload.mutate(
      { lessonId, file },
      {
        onSuccess: (image) => apply(image.imageId, image.contentType),
        onSettled: () => setUploadingKey((k) => (k === key ? null : k)),
        onError: (failure) =>
          setError(
            t(
              `administration.document.errors.${failure instanceof ImageUploadError ? failure.code : 'unavailable'}`,
            ),
          ),
      },
    );
  };

  const document: LessonDocument = {
    schemaVersion: 1,
    locale: draft.locale,
    sections: draft.sections,
  };

  const onSave = () => {
    setError(null);
    save.mutate(
      { lessonId, document, expectedUpdatedAt: draft.token },
      {
        onSuccess: (result) => {
          setEdited((cur) => ({ ...(cur ?? draft), token: result.updatedAt }));
          setDirty(false);
          setStatus({ kind: 'saved' });
        },
        onError: (failure) =>
          setError(
            t(
              `administration.document.errors.${failure instanceof DocumentAuthoringError ? failure.code : 'unavailable'}`,
            ),
          ),
      },
    );
  };

  const onPublish = async () => {
    setError(null);
    try {
      // Save the current edits first, then publish exactly what was saved.
      const saved = await save.mutateAsync({ lessonId, document, expectedUpdatedAt: draft.token });
      setEdited((cur) => ({ ...(cur ?? draft), token: saved.updatedAt }));
      const published = await publish.mutateAsync({ lessonId, expectedUpdatedAt: saved.updatedAt });
      setEdited((cur) => ({ ...(cur ?? draft), token: published.updatedAt }));
      setDirty(false);
      setStatus({ kind: 'published' });
    } catch (failure) {
      setError(
        t(
          `administration.document.errors.${failure instanceof DocumentAuthoringError ? failure.code : 'unavailable'}`,
        ),
      );
    }
  };

  const busy = save.isPending || publish.isPending;
  const lastIndex = draft.sections.length - 1;

  return (
    <section ref={editorRef} className="lesson-editor" aria-labelledby="lesson-editor-heading">
      <header className="lesson-editor-header">
        <div className="lesson-editor-headings">
          <span className="lesson-editor-eyebrow">{t('administration.document.eyebrow')}</span>
          <h1 id="lesson-editor-heading">{t('administration.document.heading')}</h1>
          <p className="lesson-editor-intro">{t('administration.document.intro')}</p>
        </div>
        <div className="lesson-editor-actions">
          <Button variant="secondary" onPress={onSave} isPending={busy}>
            {t('administration.document.saveDraft')}
          </Button>
          <Button onPress={() => void onPublish()} isPending={busy}>
            {t('administration.document.publish')}
            <ActionIcon path={actionIcons.chevron} />
          </Button>
        </div>
      </header>
      <div className="lesson-editor-meta">
        <SegmentedControl
          label={t('administration.document.modeLabel')}
          options={[
            {
              id: 'edit',
              label: t('administration.document.modeEditor'),
              icon: <ActionIcon path={actionIcons.editor} />,
            },
            {
              id: 'preview',
              label: t('administration.document.modePreview'),
              icon: <ActionIcon path={actionIcons.preview} />,
            },
          ]}
          selectedKeys={[mode]}
          onSelectionChange={(keys) => {
            const next = [...keys][0];
            if (next === 'edit' || next === 'preview') setMode(next);
          }}
        />
        <span className="lesson-save-status">
          <span
            className={dirty ? 'lesson-status-dot is-unsaved' : 'lesson-status-dot is-saved'}
            aria-hidden="true"
          />
          {t(
            dirty ? 'administration.document.unsavedStatus' : 'administration.document.savedStatus',
          )}
          {` · ${draft.sections.length}`}
        </span>
      </div>
      {meta != null && (
        <div className="lesson-title-block">
          <span className="lesson-editor-eyebrow">{t('administration.document.titleLabel')}</span>
          <h2 className="lesson-title-display">{meta.lessonTitle}</h2>
          <span className="lesson-title-caption">{`${meta.courseTitle} · ${meta.moduleLabel}`}</span>
        </div>
      )}
      {status !== null && (
        <Toast
          message={t(
            status.kind === 'saved'
              ? 'administration.document.savedToast'
              : 'administration.document.publishedToast',
          )}
          onDismiss={() => setStatus(null)}
        />
      )}
      {error !== null && <ErrorState title={error} />}
      <output className="lesson-editor-live" aria-live="polite" aria-atomic="true">
        <span key={announcement.sequence}>{announcement.text}</span>
      </output>
      {mode === 'preview' ? (
        <div className="lesson-editor-preview lesson-document-column">
          {hasRenderableSections(document) ? (
            <LessonDocumentView document={document} />
          ) : (
            <EmptyState title={t('administration.document.previewEmpty')} />
          )}
        </div>
      ) : (
        <>
          {draft.sections.length === 0 && (
            <EmptyState title={t('administration.document.empty')}>
              <p>{t('administration.document.emptyHint')}</p>
            </EmptyState>
          )}
          <ol className="lesson-editor-sections">
            {draft.sections.map((section, index) => (
              <li
                key={section.id}
                className="lesson-editor-section"
                data-type={section.type}
                data-section-id={section.id}
              >
                <div className="lesson-editor-section-bar">
                  <div className="section-identity">
                    <span className="lesson-editor-future-handle" aria-hidden="true">
                      <ActionIcon path={actionIcons.drag} />
                    </span>
                    <SectionIcon type={section.type} />
                    <span className="lesson-editor-section-type">
                      {t(`administration.document.section.${section.type}`)}
                    </span>
                    <small className="lesson-editor-section-index">
                      {String(index + 1).padStart(2, '0')}
                    </small>
                  </div>
                  <div className="lesson-editor-section-actions">
                    <Button
                      variant="secondary"
                      className="lesson-editor-action-button"
                      aria-label={t('administration.document.moveUp')}
                      {...(index === 0 ? { 'aria-disabled': true } : {})}
                      onPress={() => moveSection(index, -1)}
                    >
                      <ActionIcon path={actionIcons.up} />
                    </Button>
                    <Button
                      variant="secondary"
                      className="lesson-editor-action-button"
                      aria-label={t('administration.document.moveDown')}
                      {...(index === lastIndex ? { 'aria-disabled': true } : {})}
                      onPress={() => moveSection(index, 1)}
                    >
                      <ActionIcon path={actionIcons.down} />
                    </Button>
                    <Button
                      variant="secondary"
                      className="lesson-editor-action-button is-destructive"
                      aria-label={t('administration.document.removeSection')}
                      onPress={() => removeSection(index)}
                    >
                      <ActionIcon path={actionIcons.trash} />
                    </Button>
                  </div>
                </div>
                <SectionBody
                  section={section}
                  onChange={(next) => updateSection(index, next)}
                  onUpload={uploadInto}
                  uploadingKey={uploadingKey}
                  applyImage={applyImage}
                />
              </li>
            ))}
          </ol>
          <fieldset
            className="lesson-editor-add"
            aria-label={t('administration.document.addSection')}
          >
            <legend>{t('administration.document.addSection')}</legend>
            <div className="lesson-editor-palette">
              {sectionTypes
                .filter((type) => type !== 'video' || !hasVideo)
                .map((type) => (
                  <Button
                    key={type}
                    data-add-section={type}
                    variant="secondary"
                    className="lesson-editor-palette-tile"
                    aria-describedby={`lesson-palette-desc-${type}`}
                    onPress={() => addSection(type)}
                  >
                    <SectionIcon type={type} />
                    <span className="lesson-editor-palette-label">
                      {t(`administration.document.section.${type}`)}
                    </span>
                    <span
                      id={`lesson-palette-desc-${type}`}
                      className="lesson-editor-palette-description"
                      aria-hidden="true"
                    >
                      {t(`administration.document.sectionDescription.${type}`)}
                    </span>
                  </Button>
                ))}
            </div>
          </fieldset>
        </>
      )}
    </section>
  );
}

type SectionBodyProps = {
  readonly section: LessonSection;
  readonly onChange: (section: LessonSection) => void;
  readonly onUpload: (
    key: string,
    file: File,
    apply: (imageId: string, contentType: LessonImageContentType) => void,
  ) => void;
  readonly uploadingKey: string | null;
  readonly applyImage: (
    sectionId: string,
    columnId: string | null,
    imageId: string,
    contentType: LessonImageContentType,
  ) => void;
};

function SectionBody({ section, onChange, onUpload, uploadingKey, applyImage }: SectionBodyProps) {
  const { t } = useTranslation();
  switch (section.type) {
    case 'video':
      return (
        <div className="lesson-editor-video">
          <SectionIcon type="video" />
          <span className="lesson-editor-eyebrow">
            {t('administration.document.section.video')}
          </span>
          <p>{t('administration.document.videoSlot')}</p>
        </div>
      );
    case 'text':
    case 'task':
      return (
        <LessonDocumentRichText
          value={section.content}
          label={t(`administration.document.section.${section.type}`)}
          onChange={(content: LessonRichText) => onChange({ ...section, content })}
        />
      );
    case 'image':
      return (
        <ImageField
          imageId={section.imageId}
          contentType={section.contentType}
          alt={section.alt}
          uploading={uploadingKey === section.id}
          onUpload={(file) =>
            onUpload(section.id, file, (imageId, contentType) =>
              applyImage(section.id, null, imageId, contentType),
            )
          }
          onAlt={(alt) => onChange({ ...section, alt })}
        />
      );
    case 'row':
      return (
        <RowBody
          section={section}
          onChange={onChange}
          onUpload={onUpload}
          uploadingKey={uploadingKey}
          applyImage={applyImage}
        />
      );
  }
}

function RowBody({
  section,
  onChange,
  onUpload,
  uploadingKey,
  applyImage,
}: {
  readonly section: Extract<LessonSection, { type: 'row' }>;
  readonly onChange: (section: LessonSection) => void;
  readonly onUpload: SectionBodyProps['onUpload'];
  readonly uploadingKey: string | null;
  readonly applyImage: SectionBodyProps['applyImage'];
}) {
  const { t } = useTranslation();
  const setColumns = (columns: readonly LessonColumn[]) => onChange({ ...section, columns });
  const updateColumn = (index: number, column: LessonColumn) =>
    setColumns(section.columns.map((existing, i) => (i === index ? column : existing)));
  const addColumn = (type: 'text' | 'image') =>
    setColumns([
      ...section.columns,
      type === 'text'
        ? { type: 'text', id: localId(), content: { blocks: [] } }
        : { type: 'image', id: localId(), imageId: '', contentType: 'image/png', alt: '' },
    ]);
  const removeColumn = (index: number) => setColumns(section.columns.filter((_, i) => i !== index));
  return (
    <div className="lesson-editor-row">
      <div className="lesson-editor-columns" data-columns={section.columns.length}>
        {section.columns.map((column, index) => (
          <div key={column.id} className="lesson-editor-column">
            <div className="lesson-editor-column-bar">
              <span className="lesson-editor-section-type">
                {t(`administration.document.section.${column.type}`)}
              </span>
              {section.columns.length > 1 && (
                <Button
                  variant="danger"
                  aria-label={t('administration.document.removeColumn')}
                  onPress={() => removeColumn(index)}
                >
                  ✕
                </Button>
              )}
            </div>
            {column.type === 'text' ? (
              <LessonDocumentRichText
                value={column.content}
                label={t('administration.document.section.text')}
                onChange={(content) => updateColumn(index, { ...column, content })}
              />
            ) : (
              <ImageField
                imageId={column.imageId}
                contentType={column.contentType}
                alt={column.alt}
                uploading={uploadingKey === column.id}
                onUpload={(file) =>
                  onUpload(column.id, file, (imageId, contentType) =>
                    applyImage(section.id, column.id, imageId, contentType),
                  )
                }
                onAlt={(alt) => updateColumn(index, { ...column, alt })}
              />
            )}
          </div>
        ))}
      </div>
      {section.columns.length < 3 && (
        <fieldset
          className="lesson-editor-add-column"
          aria-label={t('administration.document.addColumn')}
        >
          <legend>{t('administration.document.addColumn')}</legend>
          <div className="lesson-editor-palette">
            <Button variant="secondary" onPress={() => addColumn('text')}>
              {t('administration.document.section.text')}
            </Button>
            <Button variant="secondary" onPress={() => addColumn('image')}>
              {t('administration.document.section.image')}
            </Button>
          </div>
        </fieldset>
      )}
    </div>
  );
}

function ImageField({
  imageId,
  contentType,
  alt,
  uploading,
  onUpload,
  onAlt,
}: {
  readonly imageId: string;
  readonly contentType: LessonImageContentType;
  readonly alt: string;
  readonly uploading: boolean;
  readonly onUpload: (file: File) => void;
  readonly onAlt: (alt: string) => void;
}) {
  const { t } = useTranslation();
  return (
    <div className="lesson-editor-image">
      {imageId !== '' && (
        <img
          src={imageServeUrl(imageId, contentType)}
          alt={alt}
          className="lesson-editor-image-preview"
        />
      )}
      <FileUpload
        label={t('administration.document.uploadImage')}
        acceptedFileTypes={['image/jpeg', 'image/png', 'image/webp']}
        onSelect={(files) => {
          const file = files?.[0];
          if (file !== undefined) onUpload(file);
        }}
        dropLabel={t('administration.document.dropHint')}
        onDrop={onUpload}
      />
      {uploading && <span>{t('administration.document.uploading')}</span>}
      <TextField label={t('administration.document.alt')} value={alt} onChange={onAlt} />
    </div>
  );
}
