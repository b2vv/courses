import { z } from 'zod';
import {
  maximumVideoBytes,
  validMediaId,
  videoChunkBytes,
  videoFailureReasons,
  videoStates,
} from '../domain/video';
import { adminTimestampSchema } from '../../../adapters/api/admin-course-schema';

const videoIdSchema = z.uuid().refine(validMediaId);
const byteCountSchema = z.number().int().min(0).max(maximumVideoBytes);
const originalNameSchema = z
  .string()
  .refine(
    (value) => [...value].length >= 1 && [...value].length <= 255 && !value.includes('\u0000'),
  );

export const lessonVideoSchema = z
  .object({
    lessonId: videoIdSchema,
    videoId: videoIdSchema,
    state: z.enum(videoStates),
    generation: z.number().int().min(1),
    originalName: originalNameSchema,
    receivedBytes: byteCountSchema,
    declaredBytes: byteCountSchema.min(1),
    updatedAt: adminTimestampSchema,
    failureReason: z.enum(videoFailureReasons).optional(),
  })
  .strict()
  .refine((video) => video.receivedBytes <= video.declaredBytes);

export const courseVideosSchema = z.object({ items: z.array(lessonVideoSchema) }).strict();

export const videoUploadTicketSchema = z
  .object({
    videoId: videoIdSchema,
    uploadUrl: z.string(),
    chunkBytes: z.literal(videoChunkBytes),
    maxBytes: z.literal(maximumVideoBytes),
    expiresAt: adminTimestampSchema,
  })
  .strict()
  .refine((ticket) => ticket.uploadUrl === `/uploads/${ticket.videoId}`);
