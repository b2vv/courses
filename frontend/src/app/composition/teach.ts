import { queryOptions } from '@tanstack/react-query';
import { getTeachCourse } from '../../adapters/api/generated/courses';
import { createApiTransport } from '../../adapters/api/transport';
import { mapAuthoringCourseTree, mapAuthoringError } from '../../modules/admin';
import type { AuthoringResult, CourseId, CourseTree } from '../../contracts/course';
import { AuthoringQueryError } from './admin-queries';
import { createIsomorphicFn } from '@tanstack/react-start';
import { getRequest, getRequestHeader } from '@tanstack/react-start/server';
import { createApiTeachRepository } from '../../modules/teach';
import {
  createApiTeachDocumentAuthoringRepository,
  createTeachDocumentQueries,
} from '../../modules/document';
import { createTeachQueries } from './teach-queries';

const requestOptions = createIsomorphicFn()
  .server(() => ({ cookie: getRequestHeader('cookie') ?? '', requestSignal: getRequest().signal }))
  .client(() => ({}));

export function createTeachDependencies() {
  const options = requestOptions();
  const teach = createTeachQueries(createApiTeachRepository(options));
  const documents = createTeachDocumentQueries(createApiTeachDocumentAuthoringRepository(options));
  const send = createApiTransport(options);
  const loadCourse = async (
    id: CourseId,
    signal: AbortSignal,
  ): Promise<AuthoringResult<CourseTree>> => {
    const outcome = await send(
      'teach.courses.load',
      (init) => getTeachCourse(id, init),
      200,
      signal,
    );
    if (!outcome.ok) return { ok: false, error: { code: outcome.code } };
    if (outcome.status === 200) return mapAuthoringCourseTree(outcome.data, id);
    return { ok: false, error: mapAuthoringError(outcome) };
  };
  const readDetail = async (id: CourseId, signal: AbortSignal): Promise<CourseTree> => {
    const result = await loadCourse(id, signal);
    if (!result.ok) throw new AuthoringQueryError(result.error);
    return result.value;
  };
  return {
    ...teach,
    documents,
    /** The owner-scoped read itself, for callers that key and recover it their own way. */
    readDetail,
    detail: (id: CourseId) =>
      queryOptions({
        queryKey: ['teach', 'courses', 'tree', id],
        queryFn: ({ signal }) => readDetail(id, signal),
      }),
  };
}

export type TeachDependencies = ReturnType<typeof createTeachDependencies>;
