import { z } from 'zod';

const id = z.uuid();

/**
 * The ceilings mirror the contract, which mirrors the Go domain, which owns them. This is the
 * outermost mirror: it refuses a payload the server should never have produced rather than
 * rendering it. The numbers are duplicated deliberately — a browser cannot import a Go constant —
 * and the pin that keeps contract and domain equal is what stops them drifting silently.
 */
export const lessonTranscriptSchema = z
  .object({
    lessonId: id,
    language: z
      .string()
      .min(2)
      .max(35)
      .regex(/^[a-z][a-z0-9]{1,7}(?:-[a-z0-9]{1,8})*$/)
      .nullable(),
    updatedAt: z.string().min(1).nullable(),
    stale: z.boolean(),
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
            documentIds: z.array(id).max(3),
          })
          .strict()
          // Ordered times are a domain rule, so a response that breaks them is refused rather
          // than highlighted: a segment ending before it starts can never be active, and would
          // silently disappear from the transcript instead of announcing a broken payload.
          .refine((segment) => segment.endMs >= segment.startMs),
      )
      .max(5000),
    documents: z
      .array(z.object({ id, label: z.string().min(1).max(200), url: z.url() }).strict())
      .max(200),
  })
  .strict();

export const transcriptErrorSchema = z.object({ error: z.object({ code: z.string() }) }).loose();

export const savedTranscriptSchema = z.object({ updatedAt: z.string().min(1) }).loose();

export const courseDocumentsSchema = z
  .object({
    documents: z
      .array(z.object({ id, label: z.string().min(1).max(200), url: z.url() }).strict())
      .max(200),
  })
  .strict();
