import { lazy, Suspense } from 'react';
import { createFileRoute, notFound } from '@tanstack/react-router';
import { validEntityId } from '../domain/course/course';
import { LessonDocumentEditorLoading } from '../presentation/admin/LessonDocumentEditorLoading';

// The editor's Tiptap dependency must not reach the shared bundle, so the component is loaded
// lazily here and is never imported by the composition root or the router context. The route sits
// under the admin layout (its role guard and session handling) but not under the course-editor
// route, which renders no Outlet — hence the trailing-underscore, non-nested segment.
const LessonDocumentEditor = lazy(() =>
  import('../presentation/admin/LessonDocumentEditor').then((module) => ({
    default: module.LessonDocumentEditor,
  })),
);

export const Route = createFileRoute('/admin/courses/$courseId_/lessons/$lessonId')({
  ssr: false,
  component: LessonDocumentEditorRoute,
});

function LessonDocumentEditorRoute() {
  const { admin } = Route.useRouteContext();
  const { lessonId } = Route.useParams();
  if (!validEntityId(lessonId)) throw notFound();
  return (
    <Suspense fallback={<LessonDocumentEditorLoading />}>
      <LessonDocumentEditor lessonId={lessonId} documents={admin.documents} />
    </Suspense>
  );
}
