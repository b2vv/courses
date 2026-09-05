import type { CatalogQuery, CourseDetail, CoursePage, Result } from '../domain/course/course';

export interface CourseRepository {
  listPublished(query: CatalogQuery, signal?: AbortSignal): Promise<Result<CoursePage>>;
  getPublished(slug: string, signal?: AbortSignal): Promise<Result<CourseDetail>>;
}
