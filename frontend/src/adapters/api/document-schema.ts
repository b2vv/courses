import { z } from 'zod';

/**
 * The outermost mirror of the lesson-document contract (OpenAPI LessonDocument), which mirrors the
 * Go domain that owns it. It refuses a payload the server should never have produced rather than
 * rendering it — above all an href that is not http(s) and an image id that is not a UUID, both of
 * which the reader puts into the DOM. The ceilings are duplicated from the domain deliberately: a
 * browser cannot import a Go constant.
 */

const trimmed = (value: string) => value.trim() !== '';
const sectionId = z.string().min(1).max(200).refine(trimmed);

function httpUrl(value: string): boolean {
  try {
    const url = new URL(value);
    return url.protocol === 'http:' || url.protocol === 'https:';
  } catch {
    return false;
  }
}

const spanSchema = z
  .object({
    text: z.string().max(20000),
    marks: z
      .array(z.enum(['bold', 'italic']))
      .max(2)
      .optional(),
    href: z.string().max(2048).refine(httpUrl).optional(),
  })
  .strict();

const blockSchema = z.discriminatedUnion('type', [
  z
    .object({ type: z.literal('paragraph'), spans: z.array(spanSchema).max(200).optional() })
    .strict(),
  z
    .object({
      type: z.literal('heading'),
      level: z.union([z.literal(2), z.literal(3)]),
      spans: z.array(spanSchema).max(200).optional(),
    })
    .strict(),
  z
    .object({
      type: z.literal('list'),
      ordered: z.boolean(),
      items: z.array(z.array(spanSchema).max(200).nullable()).min(1).max(200),
    })
    .strict(),
]);

const richTextSchema = z.object({ blocks: z.array(blockSchema).max(200).nullable() }).strict();

const imageShape = {
  id: sectionId,
  type: z.literal('image'),
  imageId: z.uuid(),
  contentType: z.enum(['image/jpeg', 'image/png', 'image/webp']),
  alt: z.string().min(1).max(2000).refine(trimmed),
} as const;

const columnSchema = z.discriminatedUnion('type', [
  z.object({ id: sectionId, type: z.literal('text'), content: richTextSchema }).strict(),
  z.object(imageShape).strict(),
]);

const sectionSchema = z.discriminatedUnion('type', [
  z.object({ id: sectionId, type: z.literal('video'), content: z.null().optional() }).strict(),
  z.object({ id: sectionId, type: z.literal('text'), content: richTextSchema }).strict(),
  z.object({ id: sectionId, type: z.literal('task'), content: richTextSchema }).strict(),
  z.object(imageShape).strict(),
  z
    .object({ id: sectionId, type: z.literal('row'), columns: z.array(columnSchema).min(1).max(3) })
    .strict(),
]);

export const lessonDocumentSchema = z
  .object({
    schemaVersion: z.literal(1),
    locale: z.enum(['uk', 'en']),
    sections: z.array(sectionSchema).max(200).nullable(),
  })
  .strict();

export const documentErrorSchema = z.object({ error: z.object({ code: z.string() }) }).loose();

export type ParsedLessonDocument = z.infer<typeof lessonDocumentSchema>;
export type ParsedSpan = z.infer<typeof spanSchema>;
export type ParsedSection = z.infer<typeof sectionSchema>;
export type ParsedColumn = z.infer<typeof columnSchema>;
export type ParsedRichText = z.infer<typeof richTextSchema>;
export type ParsedBlock = z.infer<typeof blockSchema>;
