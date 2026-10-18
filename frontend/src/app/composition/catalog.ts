import { createIsomorphicFn } from '@tanstack/react-start';
import { getRequest, getRequestHeader } from '@tanstack/react-start/server';
import { createCatalog } from '../../modules/catalog';
import { createApiCourseRepository } from '../../modules/catalog';
import { createCatalogQueries } from './catalog-queries';

const requestOptions = createIsomorphicFn()
  .server(() => ({ cookie: getRequestHeader('cookie') ?? '', requestSignal: getRequest().signal }))
  .client(() => ({}));

export function createCatalogDependencies() {
  return createCatalogQueries(createCatalog(createApiCourseRepository(requestOptions())));
}
