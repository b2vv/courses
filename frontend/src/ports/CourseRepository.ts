import type {
  CatalogQuery,
  CourseDetail,
  CoursePage,
  OwnReview,
  PublicReviewPage,
  Result,
} from '../domain/course/course';

export interface CourseRepository {
  listPublished(query: CatalogQuery, signal?: AbortSignal): Promise<Result<CoursePage>>;
  getPublished(slug: string, signal?: AbortSignal): Promise<Result<CourseDetail>>;
  listReviews(
    slug: string,
    cursor?: string,
    signal?: AbortSignal,
  ): Promise<Result<PublicReviewPage>>;
  reportReview(reviewId: string, signal?: AbortSignal): Promise<Result<void>>;
  ownReview(courseId: string, signal?: AbortSignal): Promise<Result<OwnReview>>;
}
