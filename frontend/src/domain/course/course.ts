export type Result<T> =
  | { readonly ok: true; readonly value: T }
  | { readonly ok: false; readonly error: CatalogError };
export type CatalogErrorCode =
  | 'invalid_query'
  | 'invalid_slug'
  | 'course_not_found'
  | 'dependency_unavailable'
  | 'invalid_response'
  | 'request_timeout'
  | 'request_cancelled';
export type CatalogError = { readonly code: CatalogErrorCode };
export type CourseStatus = 'draft' | 'published' | 'archived';
export type PublishedCourse = {
  readonly id: string;
  readonly slug: string;
  readonly title: string;
  readonly summary: string;
  readonly description: string;
  readonly publishedAt: string;
};
export type ProgramLesson = { readonly title: string };
export type ProgramModule = {
  readonly title: string;
  readonly lessons: readonly ProgramLesson[];
};
export type CourseDetail = PublishedCourse & { readonly program: readonly ProgramModule[] };
export type CatalogQuery = {
  readonly q: string;
  readonly sort: 'recent' | 'title';
  readonly limit: number;
  readonly offset: number;
};
export type CoursePage = {
  readonly items: readonly PublishedCourse[];
  readonly pagination: { readonly limit: number; readonly offset: number; readonly total: number };
};

export function validSlug(slug: string): boolean {
  return slug.length <= 120 && /^[a-z0-9]+(?:-[a-z0-9]+)*$/.test(slug);
}

export function makeCatalogQuery(input: Partial<CatalogQuery>): Result<CatalogQuery> {
  const query = {
    q: (input.q ?? '').trim(),
    sort: input.sort ?? 'recent',
    limit: input.limit ?? 20,
    offset: input.offset ?? 0,
  };
  if (
    !Number.isInteger(query.limit) ||
    query.limit < 1 ||
    query.limit > 50 ||
    !Number.isInteger(query.offset) ||
    query.offset < 0 ||
    query.offset > 10000 ||
    [...query.q].length > 100 ||
    query.q.includes('\u0000') ||
    !['recent', 'title'].includes(query.sort)
  )
    return { ok: false, error: { code: 'invalid_query' } };
  return { ok: true, value: query };
}

export function makePublishedCourse(input: {
  id: string;
  slug: string;
  title: string;
  summary: string;
  description: string;
  status: string;
  publishedAt: string | null;
}): Result<PublishedCourse> {
  if (
    input.status !== 'published' ||
    !validSlug(input.slug) ||
    !input.id ||
    !input.title.trim() ||
    !input.publishedAt ||
    !Number.isFinite(Date.parse(input.publishedAt))
  )
    return { ok: false, error: { code: 'invalid_response' } };
  return {
    ok: true,
    value: {
      id: input.id,
      slug: input.slug,
      title: input.title,
      summary: input.summary,
      description: input.description,
      publishedAt: input.publishedAt,
    },
  };
}
