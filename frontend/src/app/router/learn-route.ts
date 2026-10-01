const courseSlug = /^[a-z0-9]+(?:-[a-z0-9]+)*$/;
const entityId = /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i;

export type LessonRoute = { readonly slug: string; readonly lessonId: string };

export function parseCourseRoute(splat: string | undefined): string | null {
  if (splat === undefined || splat.includes('/')) return null;
  return courseSlug.test(splat) ? splat : null;
}

export function parseLessonRoute(splat: string | undefined): LessonRoute | null {
  if (splat === undefined) return null;
  const parts = splat.split('/');
  if (parts.length !== 2) return null;
  const [slug, lessonId] = parts;
  if (slug === undefined || lessonId === undefined) return null;
  return courseSlug.test(slug) && entityId.test(lessonId) ? { slug, lessonId } : null;
}

export function parseOrderId(value: string): string | null {
  return entityId.test(value) ? value : null;
}

export type OrderSearch = { readonly course?: string | undefined };

/**
 * The course slug only builds the way back to the course after payment; it arrives from
 * the URL, so an unusable value is dropped instead of trusted. The key is always returned:
 * the router merges the raw search under a route's validated one, so omitting it would let
 * the unvalidated value through.
 */
export function parseOrderSearch(search: Record<string, unknown>): OrderSearch {
  const course = search.course;
  return { course: typeof course === 'string' && courseSlug.test(course) ? course : undefined };
}
