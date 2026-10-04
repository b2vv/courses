import { z } from 'zod';
import { priceSchema } from './price-schema';

const id = z.uuid();
const title = z
  .string()
  .min(1)
  .max(200)
  .refine((value) => value.trim().length > 0);

export const learningCourseSchema = z
  .object({
    id,
    slug: z
      .string()
      .min(1)
      .max(120)
      .regex(/^[a-z0-9]+(?:-[a-z0-9]+)*$/),
    title,
    summary: z.string().min(1).max(500),
    description: z.string().min(1).max(20_000),
    status: z.enum(['draft', 'published', 'archived']),
    createdAt: z.iso.datetime({ offset: true }),
    updatedAt: z.iso.datetime({ offset: true }),
    publishedAt: z.iso.datetime({ offset: true }).nullable(),
    // OpenAPI LearningCourse (payment): required and accepted at the boundary;
    // the learning model does not map them.
    price: priceSchema.nullable(),
    paid: z.boolean(),
    modules: z
      .array(
        z
          .object({
            id,
            title,
            lessons: z
              .array(
                z
                  .object({
                    id,
                    title,
                    videoState: z.enum(['processing', 'ready', 'failed']).nullable(),
                    position: z.number().finite().nonnegative(),
                    completed: z.boolean(),
                    clientSeq: z.number().int().min(0).max(2_147_483_647),
                    isPreview: z.boolean(),
                    locked: z.boolean(),
                  })
                  .strict(),
              )
              .max(120),
          })
          .strict(),
      )
      .max(40),
    continueLessonId: id.nullable(),
    canPurchase: z.boolean(),
    // The API sends the catalog summary on the learning course too (T04); a strict schema without these
    // fields rejected every real learning course as invalid_response.
    rating: z
      .object({ average: z.number().min(1).max(5), count: z.number().int().min(3) })
      .strict()
      .nullable(),
    lessonCount: z.number().int().min(0),
  })
  .strict();

export const playbackSessionSchema = z
  .object({
    playbackId: z.string().regex(/^[0-9a-f]{32}$/),
    mediaPath: z
      .string()
      .regex(
        /^\/media\/p\/[0-9a-f]{32}\/hls\/[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}\/master\.m3u8$/,
      ),
    clientSeq: z.number().int().min(0).max(2_147_483_647),
    heartbeatSeconds: z.number().int().positive(),
    ownershipTtlSeconds: z.number().int().min(3),
  })
  .strict()
  .refine((value) => value.mediaPath.startsWith(`/media/p/${value.playbackId}/`))
  .refine((value) => value.heartbeatSeconds * 3 <= value.ownershipTtlSeconds);

export const playbackErrorSchema = z
  .object({
    error: z.object({ code: z.string(), message: z.string(), requestId: z.string() }).strict(),
  })
  .strict();
