import { lazy, Suspense } from 'react';
import { useQuery } from '@tanstack/react-query';
import { createFileRoute, notFound } from '@tanstack/react-router';
import { asBrandedId, validEntityId } from '../contracts/course';
import type { CourseTree } from '../contracts/course';
import { LessonDocumentEditorLoading, type LessonMeta } from '../modules/document';

// The editor's Tiptap dependency must not reach the shared bundle, so the component is loaded
// lazily here through a dynamic import of the module's editor file (not the barrel, which the
// composition roots import eagerly for the queries). The route sits under the teach layout (its
// role guard and session handling) but not under the course-editor route, which renders no Outlet —
// hence the trailing-underscore, non-nested segment.
const LessonDocumentEditor = lazy(() =>
  import('../modules/document/ui/LessonDocumentEditor').then((module) => ({
    default: module.LessonDocumentEditor,
  })),
);

// The lesson's own identity for the title block. Resolved from the course tree the course workspace
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

export const Route = createFileRoute('/teach/courses/$courseId_/lessons/$lessonId')({
  ssr: false,
  component: LessonDocumentEditorRoute,
});

function LessonDocumentEditorRoute() {
  const { teach } = Route.useRouteContext();
  const { courseId, lessonId } = Route.useParams();
  if (!validEntityId(lessonId)) throw notFound();
  if (!validEntityId(courseId)) throw notFound();
  const tree = useQuery(teach.detail(asBrandedId(courseId, 'course')));
  const meta = resolveLessonMeta(tree.data, lessonId);
  return (
    <Suspense fallback={<LessonDocumentEditorLoading />}>
      <LessonDocumentEditor lessonId={lessonId} documents={teach.documents} meta={meta} />
    </Suspense>
  );
}
