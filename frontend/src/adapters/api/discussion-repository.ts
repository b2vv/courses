import type {
  DiscussionComment,
  DiscussionErrorCode,
  DiscussionPage,
  DiscussionResult,
} from '../../domain/discussion/discussion';
import type { DiscussionRepository } from '../../ports/DiscussionRepository';
import type {
  PublishDiscussionInput,
  UpdateDiscussionInput,
} from '../../domain/discussion/discussion';
import {
  listLessonComments,
  publishLessonComment,
  removeLessonComment,
  updateLessonComment,
} from './generated/courses';
import { createApiTransport, unnamedStatusCode } from './transport';
import type { TransportOptions, WireResponse } from './transport';

function failure(code: DiscussionErrorCode): DiscussionResult<never> {
  return { ok: false, error: { code } };
}

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

function hasOnly(value: Record<string, unknown>, keys: readonly string[]): boolean {
  return Object.keys(value).every((key) => keys.includes(key));
}

function isComment(value: unknown): value is DiscussionComment {
  if (
    !record(value) ||
    !hasOnly(value, [
      'id',
      'author',
      'body',
      'positionMs',
      'createdAt',
      'editedAt',
      'removed',
      'canEdit',
      'canRemove',
    ]) ||
    !uuid(value.id) ||
    !record(value.author) ||
    !hasOnly(value.author, ['displayName']) ||
    (value.author.displayName !== null && typeof value.author.displayName !== 'string') ||
    (value.body !== null && typeof value.body !== 'string') ||
    (value.positionMs !== null &&
      (typeof value.positionMs !== 'number' ||
        !Number.isInteger(value.positionMs) ||
        value.positionMs < 0)) ||
    !date(value.createdAt) ||
    (value.editedAt !== null && !date(value.editedAt)) ||
    typeof value.removed !== 'boolean' ||
    typeof value.canEdit !== 'boolean' ||
    typeof value.canRemove !== 'boolean'
  ) {
    return false;
  }
  return true;
}

function isPage(value: unknown): value is DiscussionPage {
  if (
    !record(value) ||
    !hasOnly(value, ['threads', 'nextCursor']) ||
    !Array.isArray(value.threads)
  ) {
    return false;
  }
  return (
    value.threads.length <= 50 &&
    typeof value.nextCursor === 'string' &&
    value.threads.every(
      (thread) =>
        record(thread) &&
        hasOnly(thread, ['root', 'replies']) &&
        isComment(thread.root) &&
        Array.isArray(thread.replies) &&
        thread.replies.every(isComment),
    )
  );
}

function namedError(
  response: WireResponse,
  allowed: Readonly<Record<number, readonly DiscussionErrorCode[]>>,
): DiscussionErrorCode {
  const data = response.data;
  const error = record(data) ? data.error : undefined;
  const code = record(error) ? error.code : undefined;
  const named = allowed[response.status]?.find((candidate) => candidate === code);
  return named ?? (unnamedStatusCode(response.status) as DiscussionErrorCode);
}

const readErrors = {
  400: ['comment_cursor_invalid'],
  401: ['authentication_required'],
  403: ['course_forbidden'],
  404: ['lesson_not_found'],
} as const satisfies Readonly<Record<number, readonly DiscussionErrorCode[]>>;

const publishErrors = {
  400: [
    'invalid_request',
    'comment_body_invalid',
    'comment_position_invalid',
    'comment_parent_invalid',
  ],
  401: ['authentication_required'],
  403: ['origin_forbidden', 'course_forbidden'],
  404: ['lesson_not_found'],
  409: ['comment_thread_full'],
} as const satisfies Readonly<Record<number, readonly DiscussionErrorCode[]>>;

const updateErrors = {
  400: ['invalid_request', 'comment_body_invalid'],
  401: ['authentication_required'],
  403: ['comment_forbidden'],
  404: ['comment_not_found'],
  409: ['comment_removed'],
} as const satisfies Readonly<Record<number, readonly DiscussionErrorCode[]>>;

const removeErrors = {
  401: ['authentication_required'],
  403: ['comment_forbidden'],
  404: ['comment_not_found'],
} as const satisfies Readonly<Record<number, readonly DiscussionErrorCode[]>>;

export function createApiDiscussionRepository(
  options: TransportOptions = {},
): DiscussionRepository {
  const send = createApiTransport(options);
  return {
    async list(lessonId, cursor, signal) {
      const outcome = await send(
        'learn.discussion.list',
        (init) => listLessonComments(lessonId, cursor === undefined ? undefined : { cursor }, init),
        200,
        signal,
      );
      if (!outcome.ok) return failure(outcome.code as DiscussionErrorCode);
      if (outcome.status !== 200) return failure(namedError(outcome, readErrors));
      if (!isPage(outcome.data)) return failure('invalid_response');
      return {
        ok: true,
        value: {
          ...outcome.data,
          nextCursor: outcome.data.nextCursor === '' ? null : outcome.data.nextCursor,
        },
      };
    },

    async publish(lessonId, input: PublishDiscussionInput, signal) {
      const outcome = await send(
        'learn.discussion.publish',
        (init) => publishLessonComment(lessonId, input, init),
        201,
        signal,
      );
      if (!outcome.ok) return failure(outcome.code as DiscussionErrorCode);
      if (outcome.status !== 201) return failure(namedError(outcome, publishErrors));
      if (!isComment(outcome.data)) return failure('invalid_response');
      return { ok: true, value: outcome.data };
    },

    async update(input: UpdateDiscussionInput, signal) {
      const outcome = await send(
        'learn.discussion.update',
        (init) => updateLessonComment(input.commentId, { body: input.body }, init),
        200,
        signal,
      );
      if (!outcome.ok) return failure(outcome.code as DiscussionErrorCode);
      if (outcome.status !== 200) return failure(namedError(outcome, updateErrors));
      if (!isComment(outcome.data)) return failure('invalid_response');
      return { ok: true, value: outcome.data };
    },

    async remove(commentId, signal) {
      const outcome = await send(
        'learn.discussion.remove',
        (init) => removeLessonComment(commentId, init),
        204,
        signal,
      );
      if (!outcome.ok) return failure(outcome.code as DiscussionErrorCode);
      if (outcome.status !== 204) return failure(namedError(outcome, removeErrors));
      return { ok: true, value: undefined };
    },
  };
}
