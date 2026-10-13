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
import { Button } from '../ui';
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
        <h1 id="lesson-editor-heading">{t('administration.document.heading')}</h1>
        <p role="alert">{t('administration.document.loadError')}</p>
        <Button variant="secondary" onPress={() => void draftQuery.refetch()}>
          {t('administration.document.retry')}
        </Button>
      </section>
    );
  if (draft === null)
    return (
      <section className="lesson-editor" aria-labelledby="lesson-editor-heading">
        <h1 id="lesson-editor-heading">{t('administration.document.heading')}</h1>
        <p>{t('administration.document.loading')}</p>
      </section>
    );

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
      {error !== null && <p role="alert">{error}</p>}
      <ol className="lesson-editor-sections">
        {draft.sections.map((section, index) => (
          <li key={section.id} className="lesson-editor-section" data-type={section.type}>
            <div className="lesson-editor-section-bar">
              <span className="lesson-editor-section-type">
                {t(`administration.document.section.${section.type}`)}
              </span>
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
        {!hasVideo && (
          <Button variant="secondary" onPress={() => addSection('video')}>
            {t('administration.document.section.video')}
          </Button>
        )}
        <Button variant="secondary" onPress={() => addSection('text')}>
          {t('administration.document.section.text')}
        </Button>
        <Button variant="secondary" onPress={() => addSection('image')}>
          {t('administration.document.section.image')}
        </Button>
        <Button variant="secondary" onPress={() => addSection('row')}>
          {t('administration.document.section.row')}
        </Button>
        <Button variant="secondary" onPress={() => addSection('task')}>
          {t('administration.document.section.task')}
        </Button>
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
            {section.columns.length > 1 && (
              <Button
                variant="danger"
                aria-label={t('administration.document.removeColumn')}
                onPress={() => removeColumn(index)}
              >
                ✕
              </Button>
            )}
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
          <Button variant="secondary" onPress={() => addColumn('text')}>
            {t('administration.document.section.text')}
          </Button>
          <Button variant="secondary" onPress={() => addColumn('image')}>
            {t('administration.document.section.image')}
          </Button>
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
      <label>
        {t('administration.document.uploadImage')}
        <input
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
      <label>
        {t('administration.document.alt')}
        <input type="text" value={alt} onChange={(event) => onAlt(event.target.value)} />
      </label>
    </div>
  );
}
