import { z } from 'zod';

export const courseSchema = z.object({
  id: z.uuid(),
  slug: z.string(),
  title: z.string(),
  summary: z.string(),
  description: z.string(),
  status: z.enum(['draft', 'published', 'archived']),
  createdAt: z.iso.datetime({ offset: true }),
  updatedAt: z.iso.datetime({ offset: true }),
  publishedAt: z.iso.datetime({ offset: true }).nullable(),
});
export const pageSchema = z.object({
  items: z.array(courseSchema).max(50),
  pagination: z.object({
    limit: z.number().int().min(1).max(50),
    offset: z.number().int().min(0).max(10000),
    total: z.number().int().nonnegative().max(Number.MAX_SAFE_INTEGER),
  }),
});
export const errorSchema = z.object({
  error: z.object({ code: z.string(), message: z.string(), requestId: z.string() }),
});
