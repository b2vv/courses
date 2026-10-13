import { Fragment, type ReactNode } from 'react';
import { useTranslation } from 'react-i18next';
import {
  hasRenderableSections,
  imageServeUrl,
  type LessonBlock,
  type LessonColumn,
  type LessonDocument,
  type LessonRichText,
  type LessonSection,
  type LessonSpan,
} from '../../domain/document/document';

type LessonDocumentViewProps = {
  readonly document: LessonDocument;
};

/**
 * The learner-facing lesson body: typed sections rendered as semantic HTML, with no editor
 * dependency so nothing heavy reaches the shared bundle. Video sections are played by the lesson
 * player, not here, so a document that is only a video slot renders nothing — the same as a lesson
 * with no published body.
 */
export function LessonDocumentView({ document }: LessonDocumentViewProps) {
  const { t } = useTranslation();
  if (!hasRenderableSections(document)) return null;
  return (
    <section className="lesson-document" aria-label={t('learn.document.region')}>
      {document.sections.map((section) => renderSection(section))}
    </section>
  );
}

function renderSection(section: LessonSection): ReactNode {
  switch (section.type) {
    case 'video':
      // Played by the lesson player, not the document reader.
      return null;
    case 'text':
      return (
        <section key={section.id} className="lesson-document-text">
          <RichText content={section.content} />
        </section>
      );
    case 'task':
      return (
        <section key={section.id} className="lesson-document-task" data-section="task">
          <RichText content={section.content} />
        </section>
      );
    case 'image':
      return (
        <figure key={section.id} className="lesson-document-image">
          <img src={imageServeUrl(section.imageId, section.contentType)} alt={section.alt} />
        </figure>
      );
    case 'row':
      return (
        <div key={section.id} className="lesson-document-row" data-columns={section.columns.length}>
          {section.columns.map((column) => renderColumn(column))}
        </div>
      );
  }
}

function renderColumn(column: LessonColumn): ReactNode {
  if (column.type === 'text')
    return (
      <div key={column.id} className="lesson-document-column">
        <RichText content={column.content} />
      </div>
    );
  return (
    <figure key={column.id} className="lesson-document-column lesson-document-image">
      <img src={imageServeUrl(column.imageId, column.contentType)} alt={column.alt} />
    </figure>
  );
}

function RichText({ content }: { readonly content: LessonRichText }) {
  return (
    <>
      {content.blocks.map((block, index) => (
        <Block key={index} block={block} />
      ))}
    </>
  );
}

function Block({ block }: { readonly block: LessonBlock }) {
  switch (block.type) {
    case 'paragraph':
      return (
        <p>
          <Spans spans={block.spans} />
        </p>
      );
    case 'heading':
      return block.level === 2 ? (
        <h2>
          <Spans spans={block.spans} />
        </h2>
      ) : (
        <h3>
          <Spans spans={block.spans} />
        </h3>
      );
    case 'list': {
      const items = block.items.map((spans, index) => (
        <li key={index}>
          <Spans spans={spans} />
        </li>
      ));
      return block.ordered ? <ol>{items}</ol> : <ul>{items}</ul>;
    }
  }
}

function Spans({ spans }: { readonly spans: readonly LessonSpan[] }) {
  return (
    <>
      {spans.map((span, index) => (
        <Fragment key={index}>{renderSpan(span)}</Fragment>
      ))}
    </>
  );
}

// The text is placed by React, which escapes it, and the href was validated as http(s) by the
// schema, so neither can inject markup. Marks nest inside the link when both are present.
function renderSpan(span: LessonSpan): ReactNode {
  let node: ReactNode = span.text;
  if (span.marks?.includes('italic')) node = <em>{node}</em>;
  if (span.marks?.includes('bold')) node = <strong>{node}</strong>;
  if (span.href !== undefined)
    node = (
      <a href={span.href} target="_blank" rel="noreferrer noopener">
        {node}
      </a>
    );
  return node;
}
