import type {
  DocumentErrorCode,
  DocumentResult,
  LessonBlock,
  LessonColumn,
  LessonDocument,
  LessonRichText,
  LessonSection,
  LessonSpan,
} from '../../domain/document/document';
import type { DocumentRepository } from '../../ports/DocumentRepository';
import { getLearnLessonDocument } from './generated/courses';
import {
  documentErrorSchema,
  lessonDocumentSchema,
  type ParsedBlock,
  type ParsedColumn,
  type ParsedLessonDocument,
  type ParsedRichText,
  type ParsedSection,
  type ParsedSpan,
} from './document-schema';
import { createApiTransport, unnamedStatusCode } from './transport';
import type { TransportOptions, WireResponse } from './transport';

function failure(code: DocumentErrorCode): DocumentResult<never> {
  return { ok: false, error: { code } };
}

function namedError(
  response: WireResponse,
  allowed: Readonly<Record<number, readonly DocumentErrorCode[]>>,
): DocumentErrorCode {
  const parsed = documentErrorSchema.safeParse(response.data);
  if (parsed.success) {
    const code = allowed[response.status]?.find(
      (candidate) => candidate === parsed.data.error.code,
    );
    if (code !== undefined) return code;
  }
  return unnamedStatusCode(response.status) as DocumentErrorCode;
}

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

function mapDocument(parsed: ParsedLessonDocument): LessonDocument {
  return {
    schemaVersion: parsed.schemaVersion,
    locale: parsed.locale,
    sections: (parsed.sections ?? []).map(mapSection),
  };
}

/** One repository per browser/SSR request scope; generated transport never escapes this adapter. */
export function createApiDocumentRepository(options: TransportOptions = {}): DocumentRepository {
  const send = createApiTransport(options);
  return {
    async getLearnLessonDocument(lessonId, locale, signal) {
      const outcome = await send(
        'learn.document',
        (init) =>
          getLearnLessonDocument(lessonId, locale === undefined ? undefined : { locale }, init),
        200,
        signal,
      );
      if (!outcome.ok) return failure(outcome.code as DocumentErrorCode);
      if (outcome.status !== 200)
        return failure(
          namedError(outcome, {
            400: ['invalid_request'],
            401: ['authentication_required'],
            403: ['course_forbidden'],
            404: ['lesson_not_found'],
            503: ['dependency_unavailable'],
          }),
        );
      const parsed = lessonDocumentSchema.safeParse(outcome.data);
      if (!parsed.success) return failure('invalid_request');
      return { ok: true, value: mapDocument(parsed.data) };
    },
  };
}
