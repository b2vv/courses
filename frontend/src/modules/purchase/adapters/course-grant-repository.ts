import { z } from 'zod';
import type { CourseGrant, GrantErrorCode, GrantResult } from '../domain/grant';
import type { CourseGrantRepository } from '../ports/CourseGrantRepository';
import { grantCourseAccess } from '../../../adapters/api/generated/courses';
import { createApiTransport, unnamedStatusCode } from '../../../adapters/api/transport';
import type { TransportOptions, WireResponse } from '../../../adapters/api/transport';

const grantSchema = z
  .object({
    id: z.uuid(),
    userId: z.uuid(),
    courseId: z.uuid(),
    source: z.literal('admin'),
    state: z.literal('active'),
    grantedAt: z.iso.datetime({ offset: true }),
    grantedBy: z.uuid().nullable(),
  })
  .strict();

const errorSchema = z.object({ error: z.object({ code: z.string() }).loose() }).loose();

const grantErrors = {
  400: ['invalid_request'],
  401: ['authentication_required'],
  403: ['grant_forbidden', 'grant_self_forbidden', 'admin_not_learner', 'origin_forbidden'],
  409: ['course_already_owned'],
} as const satisfies Readonly<Record<number, readonly GrantErrorCode[]>>;

function failure(code: GrantErrorCode): GrantResult<never> {
  return { ok: false, error: { code } };
}

function namedError(response: WireResponse): GrantErrorCode {
  const parsed = errorSchema.safeParse(response.data);
  const code = parsed.success ? parsed.data.error.code : undefined;
  const allowed: Readonly<Record<number, readonly GrantErrorCode[]>> = grantErrors;
  return (
    allowed[response.status]?.find((candidate) => candidate === code) ??
    unnamedStatusCode(response.status)
  );
}

/** One repository per router request; generated transport types never leave this adapter. */
export function createApiCourseGrantRepository(
  options: TransportOptions = {},
): CourseGrantRepository {
  const send = createApiTransport(options);
  return {
    async grant(courseId, userId, reason, signal) {
      const outcome = await send(
        'admin.courses.grant',
        (init) => grantCourseAccess(courseId, { userId, reason }, init),
        201,
        signal,
      );
      if (!outcome.ok) return failure(outcome.code);
      if (outcome.status !== 201) return failure(namedError(outcome));
      const parsed = grantSchema.safeParse(outcome.data);
      // The grant must be the one asked for: this user, this course.
      if (!parsed.success || parsed.data.userId !== userId || parsed.data.courseId !== courseId)
        return failure('invalid_response');
      const grant: CourseGrant = {
        id: parsed.data.id,
        userId: parsed.data.userId,
        courseId: parsed.data.courseId,
        grantedAt: parsed.data.grantedAt,
      };
      return { ok: true, value: grant };
    },
  };
}
