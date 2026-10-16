import { z } from 'zod';
import { lessonDocumentSchema } from './document-schema';

export const lessonDocumentDraftSchema = z
  .object({ document: lessonDocumentSchema, updatedAt: z.string().min(1).nullable() })
  .loose();

export const lessonDocumentUpdatedSchema = z.object({ updatedAt: z.string().min(1) }).loose();

export const imageUploadSchema = z
  .object({
    id: z.uuid(),
    contentType: z.enum(['image/jpeg', 'image/png', 'image/webp']),
    width: z.number().int().positive(),
    height: z.number().int().positive(),
  })
  .loose();
