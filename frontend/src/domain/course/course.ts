export type Result<T, E = CatalogError> =
  | { readonly ok: true; readonly value: T }
  | { readonly ok: false; readonly error: E };
export type CatalogErrorCode =
  | 'invalid_query'
  | 'invalid_slug'
  | 'course_not_found'
  | 'dependency_unavailable'
  | 'invalid_response'
  | 'request_timeout'
  | 'request_cancelled'
  | 'report_own_review'
  | 'authentication_required';
export type CatalogError = { readonly code: CatalogErrorCode };
export type CourseStatus = 'draft' | 'published' | 'archived';
export type CoursePrice = {
  readonly minor: number;
  readonly currency: 'UAH';
};
export type PublishedCourse = {
  readonly id: string;
  readonly slug: string;
  readonly title: string;
  readonly summary: string;
  readonly description: string;
  readonly publishedAt: string;
  readonly price: CoursePrice | null;
  /** Catalog display metadata: true means the course has a price, not that the viewer owns it. */
  readonly paid: boolean;
  readonly rating?: Rating | null;
  readonly lessonCount?: number;
};
export type Rating = { readonly average: number; readonly count: number };
export type Review = {
  readonly id: string;
  readonly authorDisplayName: string | null;
  readonly stars: number;
  readonly body: string;
  readonly completedLessons: number;
  readonly totalLessons: number;
  readonly createdAt: string;
  readonly editedAt: string | null;
  readonly percent: number;
};
export type PublicReviewPage = {
  readonly rating: Rating | null;
  readonly items: readonly Review[];
  readonly nextCursor: string | null;
};
export type OwnReview = { readonly review: Review | null; readonly eligibility: string };
export function reviewPercent(completedLessons: number, totalLessons: number): number {
  return totalLessons === 0
    ? 0
    : Math.floor((completedLessons * 100 + totalLessons / 2) / totalLessons);
}
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

declare const entityIdBrand: unique symbol;
type BrandedId<Kind extends string> = string & { readonly [entityIdBrand]: Kind };

export type CourseId = BrandedId<'course'>;
export type ModuleId = BrandedId<'module'>;
export type LessonId = BrandedId<'lesson'>;
export type AuthoringErrorCode =
  | 'invalid_request'
  | 'authentication_required'
  | 'origin_forbidden'
  | 'forbidden'
  | 'course_forbidden'
  | 'slug_taken'
  | 'slug_immutable'
  | 'invalid_owner'
  | 'invalid_status_transition'
  | 'course_not_deletable'
  | 'course_modified'
  | 'lesson_limit'
  | 'course_not_found'
  | 'dependency_unavailable'
  | 'invalid_response'
  | 'request_timeout'
  | 'request_cancelled';
export type AuthoringError = { readonly code: AuthoringErrorCode };
export type AuthoringResult<T> = Result<T, AuthoringError>;
export type AdminCourse = {
  readonly id: CourseId;
  readonly slug: string;
  readonly title: string;
  readonly status: CourseStatus;
  readonly instructorId: string;
  readonly updatedAt: string;
};
export type AdminCourseQuery = {
  readonly limit: number;
  readonly offset: number;
  readonly status?: CourseStatus;
};
export type AdminCoursePage = {
  readonly items: readonly AdminCourse[];
  readonly pagination: { readonly limit: number; readonly offset: number; readonly total: number };
};
export type InstructorOption = {
  readonly id: string;
  readonly displayName: string;
  readonly role: 'instructor' | 'admin';
};
export type AdminCourseDraft = {
  readonly slug: string;
  readonly title: string;
  readonly summary: string;
  readonly description: string;
  readonly instructorId?: string;
};
export type AuthoringLesson = {
  readonly id: LessonId;
  readonly title: string;
  readonly description: string;
  /** Open to every signed-in viewer of a published paid course. */
  readonly isPreview: boolean;
};
export type AuthoringModule = {
  readonly id: ModuleId;
  readonly title: string;
  readonly lessons: readonly AuthoringLesson[];
};
export type CourseTreeEdit = {
  readonly slug: string;
  readonly title: string;
  readonly summary: string;
  readonly description: string;
  readonly instructorId: string;
  readonly modules: readonly AuthoringModule[];
  /**
   * Omitted means unchanged. The first price (none → some) is refused by the server
   * together with any other change, so it is sent only by `setPrice` over the stored tree.
   */
  readonly price?: CoursePrice | null;
};
export type CourseTree = CourseTreeEdit & {
  readonly price: CoursePrice | null;
  readonly id: CourseId;
  readonly status: CourseStatus;
  readonly createdAt: string;
  readonly updatedAt: string;
  readonly publishedAt: string | null;
};

const uuidPattern = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/;
const adminTimestampPattern = /^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.(\d{6})(?:Z|[+-]\d{2}:\d{2})$/;

/** The only branding escape hatch; call it only after validating an external UUID. */
export function asBrandedId<Kind extends 'course' | 'module' | 'lesson'>(
  value: string,
  _kind: Kind,
): BrandedId<Kind> {
  return value as BrandedId<Kind>;
}

export function newModuleId(): ModuleId {
  return asBrandedId(crypto.randomUUID(), 'module');
}

export function newLessonId(): LessonId {
  return asBrandedId(crypto.randomUUID(), 'lesson');
}

export function validEntityId(value: string): boolean {
  return uuidPattern.test(value) && value !== '00000000-0000-0000-0000-000000000000';
}

export function validAdminTimestamp(value: string): boolean {
  return adminTimestampPattern.test(value) && Number.isFinite(Date.parse(value));
}

function timestampMicros(value: string): bigint | null {
  const match = adminTimestampPattern.exec(value);
  const millis = Date.parse(value);
  if (match?.[1] === undefined || !Number.isFinite(millis)) return null;
  return BigInt(millis) * 1000n + BigInt(match[1].slice(3));
}

function validText(value: string, maximum: number, optional = false): boolean {
  const length = [...value].length;
  return (
    !value.includes('\u0000') &&
    length <= maximum &&
    (optional || (length > 0 && value.trim().length > 0))
  );
}

export function validCourseDraft(input: AdminCourseDraft): boolean {
  return (
    validSlug(input.slug) &&
    validText(input.title, 200) &&
    validText(input.summary, 500) &&
    validText(input.description, 20000) &&
    (input.instructorId === undefined || validEntityId(input.instructorId))
  );
}

/** Minor units of UAH: a positive safe integer. */
export function validCoursePrice(price: CoursePrice): boolean {
  return Number.isSafeInteger(price.minor) && price.minor >= 1 && price.currency === 'UAH';
}

/**
 * An amount typed in hryvnias ("1890", "1890.50", "1 890,5") as minor units. Anything else —
 * negative, more than two decimals, zero, beyond a safe integer — is refused, never rounded.
 */
export function parseUahAmount(text: string): CoursePrice | null {
  const normalized = text.replace(/[\s\u00a0]/g, '').replace(',', '.');
  const match = /^(\d{1,13})(?:\.(\d{1,2}))?$/.exec(normalized);
  if (match?.[1] === undefined) return null;
  const minor = Number(match[1]) * 100 + Number((match[2] ?? '').padEnd(2, '0'));
  const price: CoursePrice = { minor, currency: 'UAH' };
  return validCoursePrice(price) ? price : null;
}

/** The stored tree as an edit: what `setPrice` sends so nothing but the price changes. */
export function unchangedEdit(tree: CourseTree): CourseTreeEdit {
  return {
    slug: tree.slug,
    title: tree.title,
    summary: tree.summary,
    description: tree.description,
    instructorId: tree.instructorId,
    modules: tree.modules,
  };
}

export function validCourseTreeEdit(input: CourseTreeEdit): boolean {
  if (!validCourseDraft(input) || !validEntityId(input.instructorId) || input.modules.length > 40)
    return false;
  if (input.price != null && !validCoursePrice(input.price)) return false;
  const ids = new Set<string>();
  let lessonCount = 0;
  for (const module of input.modules) {
    if (!validEntityId(module.id) || ids.has(module.id) || !validText(module.title, 200))
      return false;
    ids.add(module.id);
    lessonCount += module.lessons.length;
    if (lessonCount > 120) return false;
    for (const lesson of module.lessons) {
      if (
        !validEntityId(lesson.id) ||
        ids.has(lesson.id) ||
        !validText(lesson.title, 200) ||
        !validText(lesson.description, 1000, true)
      )
        return false;
      ids.add(lesson.id);
    }
  }
  return true;
}

export function validCourseTree(input: CourseTree): boolean {
  if (
    !validEntityId(input.id) ||
    !validCourseTreeEdit(input) ||
    !validAdminTimestamp(input.createdAt) ||
    !validAdminTimestamp(input.updatedAt)
  )
    return false;
  const createdAt = timestampMicros(input.createdAt);
  const updatedAt = timestampMicros(input.updatedAt);
  if (createdAt === null || updatedAt === null) return false;
  if (updatedAt < createdAt || (input.status === 'draft' && input.publishedAt !== null))
    return false;
  if (input.status === 'published' && input.publishedAt === null) return false;
  if (input.publishedAt === null) return true;
  if (!validAdminTimestamp(input.publishedAt)) return false;
  const publishedAt = timestampMicros(input.publishedAt);
  if (publishedAt === null) return false;
  return publishedAt >= createdAt && publishedAt <= updatedAt;
}

export function validSlugChange(
  currentSlug: string,
  publishedAt: string | null,
  proposedSlug: string,
): boolean {
  return currentSlug === proposedSlug || publishedAt === null;
}

export function validTransition(from: CourseStatus, to: CourseStatus): boolean {
  return (
    (from === 'draft' && to === 'published') ||
    (from === 'published' && to === 'archived') ||
    (from === 'archived' && to === 'published')
  );
}

export function makeAdminCourseQuery(
  input: Partial<AdminCourseQuery>,
): AuthoringResult<AdminCourseQuery> {
  const limit = input.limit ?? 20;
  const offset = input.offset ?? 0;
  if (
    !Number.isInteger(limit) ||
    limit < 1 ||
    limit > 50 ||
    !Number.isInteger(offset) ||
    offset < 0 ||
    offset > 10000 ||
    (input.status !== undefined && !['draft', 'published', 'archived'].includes(input.status))
  )
    return { ok: false, error: { code: 'invalid_request' } };
  return {
    ok: true,
    value: { limit, offset, ...(input.status === undefined ? {} : { status: input.status }) },
  };
}

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
  price: { minor: number; currency: string } | null;
  paid: boolean;
  rating?: Rating | null;
  lessonCount?: number;
}): Result<PublishedCourse> {
  const validPrice =
    input.price === null ||
    (Number.isSafeInteger(input.price.minor) &&
      input.price.minor > 0 &&
      input.price.currency === 'UAH');
  if (
    input.status !== 'published' ||
    !validSlug(input.slug) ||
    !input.id ||
    !input.title.trim() ||
    !input.publishedAt ||
    !Number.isFinite(Date.parse(input.publishedAt)) ||
    !validPrice ||
    input.paid !== (input.price !== null) ||
    (input.lessonCount !== undefined &&
      (!Number.isSafeInteger(input.lessonCount) || input.lessonCount < 0)) ||
    (input.rating !== undefined &&
      input.rating !== null &&
      (!Number.isFinite(input.rating.average) ||
        input.rating.average < 1 ||
        input.rating.average > 5 ||
        !Number.isSafeInteger(input.rating.count) ||
        input.rating.count < 3))
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
      price: input.price === null ? null : { minor: input.price.minor, currency: 'UAH' },
      paid: input.paid,
      ...(input.rating === undefined
        ? {}
        : { rating: input.rating === null ? null : { ...input.rating } }),
      ...(input.lessonCount === undefined ? {} : { lessonCount: input.lessonCount }),
    },
  };
}
