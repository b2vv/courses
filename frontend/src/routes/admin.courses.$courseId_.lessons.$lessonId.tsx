import { lazy, Suspense } from 'react';
import { useQuery } from '@tanstack/react-query';
import { createFileRoute, notFound } from '@tanstack/react-router';
import { asBrandedId, validEntityId } from '../domain/course/course';
import type { CourseTree } from '../domain/course/course';
import { LessonDocumentEditorLoading } from '../presentation/admin/LessonDocumentEditorLoading';
import type { LessonMeta } from '../presentation/admin/LessonDocumentEditor';

// The editor's Tiptap dependency must not reach the shared bundle, so the component is loaded
// lazily here and is never imported by the composition root or the router context. The route sits
// under the admin layout (its role guard and session handling) but not under the course-editor
// route, which renders no Outlet — hence the trailing-underscore, non-nested segment.
const LessonDocumentEditor = lazy(() =>
  import('../presentation/admin/LessonDocumentEditor').then((module) => ({
    default: module.LessonDocumentEditor,
  })),
);

// The lesson's own identity for the title block. Resolved from the course tree the course editor
// already loads, so it is usually warm in the cache; a miss (still loading, or lesson absent) just
// leaves `meta` null and the title block is omitted rather than invented.
function resolveLessonMeta(tree: CourseTree | undefined, lessonId: string): LessonMeta | null {
  if (tree === undefined) return null;
  for (const module of tree.modules) {
    const lesson = module.lessons.find((candidate) => candidate.id === lessonId);
    if (lesson !== undefined)
      return { lessonTitle: lesson.title, courseTitle: tree.title, moduleLabel: module.title };
  }
  return null;
}

export const Route = createFileRoute('/admin/courses/$courseId_/lessons/$lessonId')({
  ssr: false,
  component: LessonDocumentEditorRoute,
});

function LessonDocumentEditorRoute() {
  const { admin } = Route.useRouteContext();
  const { courseId, lessonId } = Route.useParams();
  if (!validEntityId(lessonId)) throw notFound();
  if (!validEntityId(courseId)) throw notFound();
  const tree = useQuery(admin.detail(asBrandedId(courseId, 'course')));
  const meta = resolveLessonMeta(tree.data, lessonId);
  return (
    <Suspense fallback={<LessonDocumentEditorLoading />}>
      <LessonDocumentEditor lessonId={lessonId} documents={admin.documents} meta={meta} />
    </Suspense>
  );
}
