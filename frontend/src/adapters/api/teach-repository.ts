import {
  dismissCommentReports,
  listTeachCourseComments,
  listTeachModerationQueue,
} from './generated/courses';
import { createApiTransport, unnamedStatusCode } from './transport';
import type { TransportOptions, WireResponse } from './transport';
import type { TeachRepository } from '../../ports/TeachRepository';
import type {
  DismissReportsInput,
  ModerationQueuePage,
  TeachComment,
  TeachCommentPage,
  TeachErrorCode,
  TeachResult,
} from '../../domain/teach/teach';

function record(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null;
}
function uuid(value: unknown): value is string {
  return (
    typeof value === 'string' &&
    /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i.test(value)
  );
}
function date(value: unknown): value is string {
  return typeof value === 'string' && !Number.isNaN(Date.parse(value));
}
function comment(value: unknown): value is TeachComment {
  if (
    !record(value) ||
    !uuid(value.id) ||
    !uuid(value.lessonId) ||
    (value.parentId !== null && !uuid(value.parentId)) ||
    !record(value.author) ||
    (value.author.displayName !== null && typeof value.author.displayName !== 'string') ||
    (value.body !== null && typeof value.body !== 'string') ||
    (value.positionMs !== null &&
      (typeof value.positionMs !== 'number' ||
        !Number.isInteger(value.positionMs) ||
        value.positionMs < 0)) ||
    !date(value.createdAt) ||
    (value.editedAt !== null && !date(value.editedAt)) ||
    typeof value.removed !== 'boolean'
  )
    return false;
  return true;
}
function page(value: unknown): value is TeachCommentPage {
  return (
    record(value) &&
    Array.isArray(value.comments) &&
    value.comments.length <= 50 &&
    typeof value.nextCursor === 'string' &&
    value.comments.every(comment)
  );
}
function queuePage(value: unknown): value is ModerationQueuePage {
  return (
    record(value) &&
    Array.isArray(value.items) &&
    value.items.length <= 50 &&
    typeof value.nextCursor === 'string' &&
    value.items.every(
      (item) =>
        record(item) &&
        typeof item.openReports === 'number' &&
        Number.isInteger(item.openReports) &&
        item.openReports > 0 &&
        comment(item.comment),
    )
  );
}
function failure(code: TeachErrorCode): TeachResult<never> {
  return { ok: false, error: { code } };
}
function namedError(
  response: WireResponse,
  allowed: Readonly<Record<number, readonly TeachErrorCode[]>>,
): TeachErrorCode {
  const data = response.data;
  const error = record(data) && record(data.error) ? data.error.code : undefined;
  const named = allowed[response.status]?.find((candidate) => candidate === error);
  return named ?? (unnamedStatusCode(response.status) as TeachErrorCode);
}
const readErrors = {
  400: ['invalid_request'],
  401: ['authentication_required'],
  403: ['course_forbidden'],
  404: ['course_not_found'],
} as const satisfies Readonly<Record<number, readonly TeachErrorCode[]>>;
const dismissErrors = {
  400: ['invalid_request'],
  401: ['authentication_required'],
  403: ['course_forbidden', 'comment_forbidden'],
  404: ['comment_not_found'],
} as const satisfies Readonly<Record<number, readonly TeachErrorCode[]>>;

export function createApiTeachRepository(options: TransportOptions = {}): TeachRepository {
  const send = createApiTransport(options);
  return {
    async listCourseComments(courseId, cursor, signal) {
      const outcome = await send(
        'teach.comments.list',
        (init) =>
          listTeachCourseComments(courseId, cursor === undefined ? undefined : { cursor }, init),
        200,
        signal,
      );
      if (!outcome.ok) return failure(outcome.code as TeachErrorCode);
      if (outcome.status !== 200) return failure(namedError(outcome, readErrors));
      if (!page(outcome.data)) return failure('invalid_response');
      return {
        ok: true,
        value: {
          ...outcome.data,
          nextCursor: outcome.data.nextCursor === '' ? null : outcome.data.nextCursor,
        },
      };
    },
    async listModerationQueue(courseId, cursor, signal) {
      const outcome = await send(
        'teach.moderation.list',
        (init) =>
          listTeachModerationQueue(courseId, cursor === undefined ? undefined : { cursor }, init),
        200,
        signal,
      );
      if (!outcome.ok) return failure(outcome.code as TeachErrorCode);
      if (outcome.status !== 200) return failure(namedError(outcome, readErrors));
      if (!queuePage(outcome.data)) return failure('invalid_response');
      return {
        ok: true,
        value: {
          items: outcome.data.items,
          nextCursor: outcome.data.nextCursor === '' ? null : outcome.data.nextCursor,
        },
      };
    },
    async dismissReport(commentId, input: DismissReportsInput, signal) {
      const outcome = await send(
        'teach.moderation.dismiss',
        (init) => dismissCommentReports(commentId, input, init),
        204,
        signal,
      );
      if (!outcome.ok) return failure(outcome.code as TeachErrorCode);
      if (outcome.status !== 204) return failure(namedError(outcome, dismissErrors));
      return { ok: true, value: undefined };
    },
  };
}
