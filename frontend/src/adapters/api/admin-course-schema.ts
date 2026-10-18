import { z } from 'zod';
import { priceSchema } from './price-schema';
import { asBrandedId, validCourseTree, validEntityId } from '../../contracts/course';

const entityIdSchema = z.uuid().refine(validEntityId);
const courseIdSchema = entityIdSchema.transform((value) => asBrandedId(value, 'course'));
const moduleIdSchema = entityIdSchema.transform((value) => asBrandedId(value, 'module'));
const lessonIdSchema = entityIdSchema.transform((value) => asBrandedId(value, 'lesson'));
const statusSchema = z.enum(['draft', 'published', 'archived']);
const requiredText = (maximum: number) =>
  z
    .string()
    .min(1)
    .max(maximum)
    .refine((value) => value.trim().length > 0 && !value.includes('\u0000'));

export const adminTimestampSchema = z.iso.datetime({ offset: true, precision: 6 });

const courseCountSchema = z.number().int().nonnegative().max(Number.MAX_SAFE_INTEGER);

export const adminCourseSchema = z
  .object({
    id: courseIdSchema,
    slug: z
      .string()
      .min(1)
      .max(120)
      .regex(/^[a-z0-9]+(?:-[a-z0-9]+)*$/),
    title: requiredText(200),
    status: statusSchema,
    instructorId: entityIdSchema,
    updatedAt: adminTimestampSchema,
    lessonCount: courseCountSchema,
    studentCount: courseCountSchema,
    // Required and nullable in the contract (null = free course).
    price: priceSchema.nullable(),
  })
  .strict();

const paginationSchema = z
  .object({
    limit: z.number().int().min(1).max(50),
    offset: z.number().int().min(0).max(10000),
    total: z.number().int().nonnegative().max(Number.MAX_SAFE_INTEGER),
  })
  .strict();

const statusCountsSchema = z
  .object({
    total: courseCountSchema,
    draft: courseCountSchema,
    published: courseCountSchema,
    archived: courseCountSchema,
  })
  .strict();

export const adminCoursePageSchema = z
  .object({
    items: z.array(adminCourseSchema).max(50),
    pagination: paginationSchema,
    statusCounts: statusCountsSchema,
  })
  .strict();

const adminLessonSchema = z
  .object({
    id: lessonIdSchema,
    title: requiredText(200),
    // Optional in the contract with default false; saves always state it.
    isPreview: z.boolean().default(false),
  })
  .strict();

const adminModuleSchema = z
  .object({
    id: moduleIdSchema,
    title: requiredText(200),
    lessons: z.array(adminLessonSchema).max(120),
  })
  .strict();

export const adminCourseTreeSchema = z
  .object({
    id: courseIdSchema,
    slug: z
      .string()
      .min(1)
      .max(120)
      .regex(/^[a-z0-9]+(?:-[a-z0-9]+)*$/),
    title: requiredText(200),
    summary: requiredText(500),
    description: requiredText(20000),
    status: statusSchema,
    instructorId: entityIdSchema,
    createdAt: adminTimestampSchema,
    updatedAt: adminTimestampSchema,
    publishedAt: adminTimestampSchema.nullable(),
    // Required and nullable in the contract (null = free). The course form omits it
    // ("unchanged"); only `setPrice` sends it.
    price: priceSchema.nullable(),
    // Active enrolments, derived server-side; read-only, shown in the editor mini-stats.
    studentCount: z.number().int().min(0),
    modules: z.array(adminModuleSchema).max(40),
  })
  .strict()
  .refine(validCourseTree);

const instructorOptionSchema = z
  .object({
    id: entityIdSchema,
    displayName: requiredText(120),
    role: z.enum(['instructor', 'admin']),
  })
  .strict();

export const instructorOptionsSchema = z.array(instructorOptionSchema).max(200);
