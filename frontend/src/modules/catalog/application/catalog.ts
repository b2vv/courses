import { makeCatalogQuery, validSlug } from '../../../contracts/course';
import type {
  CatalogQuery,
  CourseDetail,
  CoursePage,
  OwnReview,
  PublicReviewPage,
  Result,
} from '../../../contracts/course';
import type { CourseRepository } from '../ports/CourseRepository';

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
    reviews(
      slug: string,
      cursor?: string,
      signal?: AbortSignal,
    ): Promise<Result<PublicReviewPage>> {
      return validSlug(slug)
        ? repository.listReviews(slug, cursor, signal)
        : Promise.resolve({ ok: false, error: { code: 'invalid_slug' } });
    },
    report(reviewId: string, signal?: AbortSignal): Promise<Result<void>> {
      return repository.reportReview(reviewId, signal);
    },
    ownReview(courseId: string, signal?: AbortSignal): Promise<Result<OwnReview>> {
      return repository.ownReview(courseId, signal);
    },
  };
}
