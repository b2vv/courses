import { z } from 'zod';

const count = z.number().int().nonnegative().max(Number.MAX_SAFE_INTEGER);

const roleSchema = z.object({ student: count, instructor: count, admin: count }).strict();

const usersSchema = z
  .object({
    active: count,
    blocked: count,
    deleted: count,
    byRole: roleSchema,
    verified: count,
    twoFactor: count,
  })
  .strict();

const coursesSchema = z.object({ draft: count, published: count, archived: count }).strict();

const grantsSchema = z.object({ active: count, blocked: count, revoked: count }).strict();

const ordersSchema = z
  .object({ pending: count, paid: count, failed: count, revenueMinorUAH: count })
  .strict();

const reportsSchema = z.object({ open: count, resolved: count }).strict();

const mediaSchema = z
  .object({
    inProgress: count,
    ready: count,
    failed: count,
    totalDurationSeconds: z.number().nonnegative(),
    avgDurationSeconds: z.number().nonnegative().nullable(),
    withDurationCount: count,
  })
  .strict();

const engagementSchema = z
  .object({
    activeLearners30d: count,
    avgProgress: z.number().min(0).max(1).nullable(),
    completions: count,
  })
  .strict();

const instructorSchema = z
  .object({
    instructorId: z.uuid(),
    displayName: z.string(),
    publishedCourses: count,
    activeEnrollments: count,
    revenueMinorUAH: count,
  })
  .strict();

const platformSchema = z.object({ singleInstructor: z.boolean() }).strict();

export const snapshotSchema = z
  .object({
    users: usersSchema,
    courses: coursesSchema,
    grants: grantsSchema,
    orders: ordersSchema,
    reports: reportsSchema,
    media: mediaSchema,
    engagement: engagementSchema,
    instructors: z.array(instructorSchema),
    platform: platformSchema,
  })
  .strict();

export const timeseriesPointSchema = z
  .object({ bucketStart: z.iso.datetime({ offset: true }), value: z.number() })
  .strict();

export const timeseriesSchema = z.array(timeseriesPointSchema);
