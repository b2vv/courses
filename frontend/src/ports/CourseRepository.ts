import type { CatalogQuery, CoursePage, PublishedCourse, Result } from '../domain/course/course';

export interface CourseRepository {
  listPublished(query: CatalogQuery, signal?: AbortSignal): Promise<Result<CoursePage>>;
  getPublished(slug: string, signal?: AbortSignal): Promise<Result<PublishedCourse>>;
}
