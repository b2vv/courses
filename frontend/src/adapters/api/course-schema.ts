import { z } from 'zod';
import { priceSchema } from './price-schema';

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
  price: priceSchema.nullable(),
  paid: z.boolean(),
  rating: z
    .object({ average: z.number().min(1).max(5), count: z.number().int().min(3) })
    .nullable(),
  lessonCount: z.number().int().nonnegative(),
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
export const reviewSchema = z
  .object({
    id: z.uuid(),
    authorDisplayName: z.string().nullable(),
    stars: z.number().int().min(1).max(5),
    body: z.string(),
    completedLessons: z.number().int().nonnegative(),
    totalLessons: z.number().int().nonnegative(),
    createdAt: z.iso.datetime({ offset: true }),
    editedAt: z.iso.datetime({ offset: true }).nullable(),
  })
  .refine((review) => review.totalLessons >= review.completedLessons);
export const publicReviewPageSchema = z.object({
  rating: z
    .object({ average: z.number().min(1).max(5), count: z.number().int().min(3) })
    .nullable(),
  items: z.array(reviewSchema).max(20),
  nextCursor: z.string().nullable(),
});
export const ownReviewSchema = z.object({
  review: reviewSchema.nullable(),
  eligibility: z.string(),
});
