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
