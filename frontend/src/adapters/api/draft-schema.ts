import { z } from 'zod';

const id = z.uuid();

const draftState = z.enum(['pending', 'running', 'ready', 'failed', 'no_speech']);

export const lessonDraftSchema = z
  .object({
    lessonId: id,
    state: draftState.nullable(),
    language: z
      .string()
      .min(2)
      .max(35)
      .regex(/^[a-z][a-z0-9]{1,7}(?:-[a-z0-9]{1,8})*$/)
      .nullable(),
    failure: z.string().max(100).nullable(),
    stale: z.boolean(),
    updatedAt: z.string().min(1).nullable(),
    assetId: id.nullable(),
    generation: z.number().int().min(1).nullable(),
    segments: z
      .array(
        z
          .object({
            id,
            startMs: z.number().int().min(0),
            endMs: z.number().int().min(0),
            text: z.string().min(1).max(1000),
            speaker: z.string().min(1).max(100).optional(),
            confidence: z.number().min(0).max(1).optional(),
          })
          .strict()
          .refine((segment) => segment.endMs >= segment.startMs),
      )
      .max(5000),
  })
  .strict();

export const draftErrorSchema = z.object({ error: z.object({ code: z.string() }) }).loose();
