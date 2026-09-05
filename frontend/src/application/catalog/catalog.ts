import { makeCatalogQuery, validSlug } from '../../domain/course/course';
import type { CatalogQuery, CourseDetail, CoursePage, Result } from '../../domain/course/course';
import type { CourseRepository } from '../../ports/CourseRepository';

export function createCatalog(repository: CourseRepository) {
  return {
    list(input: Partial<CatalogQuery>, signal?: AbortSignal): Promise<Result<CoursePage>> {
      const result = makeCatalogQuery(input);
      return result.ok ? repository.listPublished(result.value, signal) : Promise.resolve(result);
    },
    get(slug: string, signal?: AbortSignal): Promise<Result<CourseDetail>> {
      return validSlug(slug)
        ? repository.getPublished(slug, signal)
        : Promise.resolve({ ok: false, error: { code: 'invalid_slug' } });
    },
  };
}
