import type {
  LessonBlock,
  LessonColumn,
  LessonDocument,
  LessonRichText,
  LessonSection,
  LessonSpan,
} from '../../domain/document/document';
import type {
  ParsedBlock,
  ParsedColumn,
  ParsedLessonDocument,
  ParsedRichText,
  ParsedSection,
  ParsedSpan,
} from './document-schema';
import type {
  LessonBlock as WireBlock,
  LessonDocument as WireDocument,
  LessonSection as WireSection,
  LessonSpan as WireSpan,
} from './generated/models';

// The lesson document crosses the wire as the generated, all-optional OpenAPI shape; the reader and
// the editor work with the domain discriminated union. These two pure maps are the only place that
// translates between them, shared so the reader and the authoring adapter cannot drift.

function mapSpan(span: ParsedSpan): LessonSpan {
  return {
    text: span.text,
    ...(span.marks === undefined ? {} : { marks: span.marks }),
    ...(span.href === undefined ? {} : { href: span.href }),
  };
}

function mapBlock(block: ParsedBlock): LessonBlock {
  switch (block.type) {
    case 'paragraph':
      return { type: 'paragraph', spans: (block.spans ?? []).map(mapSpan) };
    case 'heading':
      return { type: 'heading', level: block.level, spans: (block.spans ?? []).map(mapSpan) };
    case 'list':
      return {
        type: 'list',
        ordered: block.ordered,
        items: block.items.map((item) => (item ?? []).map(mapSpan)),
      };
  }
}

function mapRichText(content: ParsedRichText): LessonRichText {
  return { blocks: (content.blocks ?? []).map(mapBlock) };
}

function mapColumn(column: ParsedColumn): LessonColumn {
  if (column.type === 'text')
    return { type: 'text', id: column.id, content: mapRichText(column.content) };
  return {
    type: 'image',
    id: column.id,
    imageId: column.imageId,
    contentType: column.contentType,
    alt: column.alt,
  };
}

function mapSection(section: ParsedSection): LessonSection {
  switch (section.type) {
    case 'video':
      return { type: 'video', id: section.id };
    case 'text':
      return { type: 'text', id: section.id, content: mapRichText(section.content) };
    case 'task':
      return { type: 'task', id: section.id, content: mapRichText(section.content) };
    case 'image':
      return {
        type: 'image',
        id: section.id,
        imageId: section.imageId,
        contentType: section.contentType,
        alt: section.alt,
      };
    case 'row':
      return { type: 'row', id: section.id, columns: section.columns.map(mapColumn) };
  }
}

export function mapDocument(parsed: ParsedLessonDocument): LessonDocument {
  return {
    schemaVersion: parsed.schemaVersion,
    locale: parsed.locale,
    sections: (parsed.sections ?? []).map(mapSection),
  };
}

function toWireSpan(span: LessonSpan): WireSpan {
  return {
    text: span.text,
    ...(span.marks === undefined ? {} : { marks: [...span.marks] }),
    ...(span.href === undefined ? {} : { href: span.href }),
  };
}

function toWireBlock(block: LessonBlock): WireBlock {
  switch (block.type) {
    case 'paragraph':
      return { type: 'paragraph', spans: block.spans.map(toWireSpan) };
    case 'heading':
      return { type: 'heading', level: block.level, spans: block.spans.map(toWireSpan) };
    case 'list':
      return {
        type: 'list',
        ordered: block.ordered,
        items: block.items.map((item) => item.map(toWireSpan)),
      };
  }
}

function toWireSection(section: LessonSection): WireSection {
  switch (section.type) {
    case 'video':
      return { id: section.id, type: 'video' };
    case 'text':
      return {
        id: section.id,
        type: 'text',
        content: { blocks: section.content.blocks.map(toWireBlock) },
      };
    case 'task':
      return {
        id: section.id,
        type: 'task',
        content: { blocks: section.content.blocks.map(toWireBlock) },
      };
    case 'image':
      return {
        id: section.id,
        type: 'image',
        imageId: section.imageId,
        contentType: section.contentType,
        alt: section.alt,
      };
    case 'row':
      return {
        id: section.id,
        type: 'row',
        columns: section.columns.map((column) =>
          column.type === 'text'
            ? {
                id: column.id,
                type: 'text',
                content: { blocks: column.content.blocks.map(toWireBlock) },
              }
            : {
                id: column.id,
                type: 'image',
                imageId: column.imageId,
                contentType: column.contentType,
                alt: column.alt,
              },
        ),
      };
  }
}

export function toWireDocument(document: LessonDocument): WireDocument {
  return {
    schemaVersion: document.schemaVersion,
    locale: document.locale,
    sections: document.sections.map(toWireSection),
  };
}
