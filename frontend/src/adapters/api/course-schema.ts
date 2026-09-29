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
  price: z
    .object({
      minor: z.number().int().positive().max(Number.MAX_SAFE_INTEGER),
      currency: z.literal('UAH'),
    })
    .strict()
    .nullable(),
  paid: z.boolean(),
});
const programTitleSchema = z
  .string()
  .min(1)
  .max(200)
  .refine((title) => title.trim().length > 0);
const programLessonSchema = z.object({ title: programTitleSchema }).strict();
const programModuleSchema = z
  .object({
    title: programTitleSchema,
    lessons: z.array(programLessonSchema).max(120),
  })
  .strict();
export const courseDetailSchema = courseSchema.extend({
  program: z.array(programModuleSchema).max(40),
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
