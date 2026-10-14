import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import type {
  LessonColumn,
  LessonDocument,
  LessonImageContentType,
  LessonRichText,
  LessonSection,
} from '../../domain/document/document';
import { imageServeUrl } from '../../domain/document/document';
import {
  DocumentAuthoringError,
  ImageUploadError,
  type createAdminDocumentQueries,
} from '../../app/composition/admin-document-queries';
import { Button, TextField } from '../ui';
import { EmptyState, ErrorState } from '../ui/states';
import { LessonDocumentEditorLoading } from './LessonDocumentEditorLoading';
import { LessonDocumentRichText } from './LessonDocumentRichText';

type AdminDocumentQueries = ReturnType<typeof createAdminDocumentQueries>;

type LessonDocumentEditorProps = {
  readonly lessonId: string;
  readonly documents: AdminDocumentQueries;
};

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

export function LessonDocumentEditor({ lessonId, documents }: LessonDocumentEditorProps) {
  const { t } = useTranslation();
  const client = useQueryClient();
  const draftQuery = useQuery(documents.document(lessonId));
  const save = useMutation(documents.save(client));
  const publish = useMutation(documents.publish(client));
  const upload = useMutation(documents.uploadImage());
  // `edited` overlays the loaded draft once the author touches it; until then the editor renders
  // from the query. Deriving the draft (rather than seeding state in an effect) avoids a cascading
  // render, and a later refetch cannot clobber in-progress edits.
  const [edited, setEdited] = useState<DraftState | null>(null);
  const [error, setError] = useState<string | null>(null);

  const base: DraftState | null =
    draftQuery.data === undefined
      ? null
      : {
          sections: draftQuery.data.document.sections,
          token: draftQuery.data.updatedAt,
          locale: draftQuery.data.document.locale,
        };
  const draft = edited ?? base;

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

  const setSections = (next: readonly LessonSection[]) => setEdited({ ...draft, sections: next });

  const updateSection = (index: number, section: LessonSection) =>
    setSections(draft.sections.map((existing, i) => (i === index ? section : existing)));

  const addSection = (type: LessonSection['type']) =>
    setSections([...draft.sections, blankSection(type)]);

  const removeSection = (index: number) =>
    setSections(draft.sections.filter((_, i) => i !== index));

  // Move keeps the button enabled at the edges (no isDisabled, which would drop focus to the body);
  // a move past the edge is simply a no-op.
  const moveSection = (index: number, delta: -1 | 1) => {
    const target = index + delta;
    if (target < 0 || target >= draft.sections.length) return;
    const next = [...draft.sections];
    const moved = next[index];
    const swapped = next[target];
    if (moved === undefined || swapped === undefined) return;
    next[index] = swapped;
    next[target] = moved;
    setSections(next);
  };

  const uploadInto = (
    file: File,
    apply: (imageId: string, contentType: LessonImageContentType) => void,
  ) => {
    setError(null);
    upload.mutate(
      { lessonId, file },
      {
        onSuccess: (image) => apply(image.imageId, image.contentType),
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
        onSuccess: (result) => setEdited({ ...draft, token: result.updatedAt }),
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
      setEdited({ ...draft, token: saved.updatedAt });
      const published = await publish.mutateAsync({ lessonId, expectedUpdatedAt: saved.updatedAt });
      setEdited({ ...draft, token: published.updatedAt });
    } catch (failure) {
      setError(
        t(
          `administration.document.errors.${failure instanceof DocumentAuthoringError ? failure.code : 'unavailable'}`,
        ),
      );
    }
  };

  const busy = save.isPending || publish.isPending;

  return (
    <section className="lesson-editor" aria-labelledby="lesson-editor-heading">
      <header className="lesson-editor-header">
        <h1 id="lesson-editor-heading">{t('administration.document.heading')}</h1>
        <div className="lesson-editor-actions">
          <Button variant="secondary" onPress={onSave} isPending={busy}>
            {t('administration.document.saveDraft')}
          </Button>
          <Button onPress={() => void onPublish()} isPending={busy}>
            {t('administration.document.publish')}
          </Button>
        </div>
      </header>
      {error !== null && <ErrorState title={error} />}
      {draft.sections.length === 0 && (
        <EmptyState title={t('administration.document.empty')}>
          <p>{t('administration.document.emptyHint')}</p>
        </EmptyState>
      )}
      <ol className="lesson-editor-sections">
        {draft.sections.map((section, index) => (
          <li key={section.id} className="lesson-editor-section" data-type={section.type}>
            <div className="lesson-editor-section-bar">
              <SectionIcon type={section.type} />
              <span className="lesson-editor-section-type">
                {t(`administration.document.section.${section.type}`)}
              </span>
              <span className="lesson-editor-section-index">{index + 1}</span>
              <div className="lesson-editor-section-actions">
                <Button
                  variant="secondary"
                  aria-label={t('administration.document.moveUp')}
                  onPress={() => moveSection(index, -1)}
                >
                  ↑
                </Button>
                <Button
                  variant="secondary"
                  aria-label={t('administration.document.moveDown')}
                  onPress={() => moveSection(index, 1)}
                >
                  ↓
                </Button>
                <Button
                  variant="danger"
                  aria-label={t('administration.document.removeSection')}
                  onPress={() => removeSection(index)}
                >
                  ✕
                </Button>
              </div>
            </div>
            <SectionBody
              section={section}
              onChange={(next) => updateSection(index, next)}
              onUpload={uploadInto}
              uploading={upload.isPending}
            />
          </li>
        ))}
      </ol>
      <fieldset className="lesson-editor-add" aria-label={t('administration.document.addSection')}>
        <legend>{t('administration.document.addSection')}</legend>
        <div className="lesson-editor-palette">
          {sectionTypes
            .filter((type) => type !== 'video' || !hasVideo)
            .map((type) => (
              <Button
                key={type}
                variant="secondary"
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
    </section>
  );
}

type SectionBodyProps = {
  readonly section: LessonSection;
  readonly onChange: (section: LessonSection) => void;
  readonly onUpload: (
    file: File,
    apply: (imageId: string, contentType: LessonImageContentType) => void,
  ) => void;
  readonly uploading: boolean;
};

function SectionBody({ section, onChange, onUpload, uploading }: SectionBodyProps) {
  const { t } = useTranslation();
  switch (section.type) {
    case 'video':
      return <p className="lesson-editor-video">{t('administration.document.videoSlot')}</p>;
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
          uploading={uploading}
          onUpload={(file) =>
            onUpload(file, (imageId, contentType) => onChange({ ...section, imageId, contentType }))
          }
          onAlt={(alt) => onChange({ ...section, alt })}
        />
      );
    case 'row':
      return (
        <RowBody section={section} onChange={onChange} onUpload={onUpload} uploading={uploading} />
      );
  }
}

function RowBody({
  section,
  onChange,
  onUpload,
  uploading,
}: {
  readonly section: Extract<LessonSection, { type: 'row' }>;
  readonly onChange: (section: LessonSection) => void;
  readonly onUpload: SectionBodyProps['onUpload'];
  readonly uploading: boolean;
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
                uploading={uploading}
                onUpload={(file) =>
                  onUpload(file, (imageId, contentType) =>
                    updateColumn(index, { ...column, imageId, contentType }),
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
      <label className="lesson-editor-upload">
        <span>{t('administration.document.uploadImage')}</span>
        <input
          className="ui-input"
          type="file"
          accept="image/jpeg,image/png,image/webp"
          onChange={(event) => {
            const file = event.target.files?.[0];
            if (file !== undefined) onUpload(file);
            event.target.value = '';
          }}
        />
      </label>
      {uploading && <span>{t('administration.document.uploading')}</span>}
      <TextField label={t('administration.document.alt')} value={alt} onChange={onAlt} />
    </div>
  );
}
