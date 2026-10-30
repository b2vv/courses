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
  return { ...teach, documents };
}

export type TeachDependencies = ReturnType<typeof createTeachDependencies>;
