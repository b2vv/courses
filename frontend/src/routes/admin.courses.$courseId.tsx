import { useMutation, useQuery, useQueryClient, useSuspenseQuery } from '@tanstack/react-query';
import { createFileRoute, notFound } from '@tanstack/react-router';
import { useEffect, useState } from 'react';
import { AuthoringQueryError } from '../app/composition/admin-queries';
import { VideoQueryError } from '../app/composition/admin-video-queries';
import { ensureRouteQuery } from '../app/composition/catalog-queries';
import { sessionQueryKey } from '../app/composition/identity-queries';
import { asBrandedId, validEntityId } from '../domain/course/course';
import type { CourseId, CourseStatus, CourseTree, CourseTreeEdit } from '../domain/course/course';
import type { AuthenticatedUser } from '../domain/identity/identity';
import type {
  VideoErrorCode,
  VideoTransferProgress,
  VideoTransferTarget,
} from '../domain/media/video';
import { CourseEditorPending, CourseEditorView } from '../presentation/admin/CourseEditorView';
import type {
  CourseEditorVideo,
  LessonVideoFieldActions,
  LessonVideoFieldActionResult,
} from '../presentation/admin/LessonVideoField';
import { CourseListFailure } from '../presentation/admin/CourseListView';

function courseId(value: string): CourseId {
  if (!validEntityId(value)) throw notFound();
  return asBrandedId(value, 'course');
}

function videoErrorCode(error: unknown): VideoErrorCode | undefined {
  return error instanceof VideoQueryError
    ? error.code
    : error === null || error === undefined
      ? undefined
      : 'dependency_unavailable';
}

export const Route = createFileRoute('/admin/courses/$courseId')({
  ssr: false,
  loader: async ({ context, params, abortController }) => {
    try {
      return await ensureRouteQuery(
        context.queryClient,
        context.admin.detail(courseId(params.courseId)),
        abortController.signal,
      );
    } catch (error) {
      if (
        error instanceof AuthoringQueryError &&
        ['course_not_found', 'invalid_request'].includes(error.code)
      )
        throw notFound();
      throw error;
    }
  },
  head: ({ loaderData }) => ({
    meta: [
      {
        title: loaderData ? `${loaderData.title} — CourseForge` : 'Редактор курсу — CourseForge',
      },
    ],
  }),
  component: AdminCourseEditorRoute,
  pendingComponent: CourseEditorPending,
  errorComponent: CourseListFailure,
});

function AdminCourseEditorRoute() {
  const { admin, adminActor } = Route.useRouteContext();
  const { courseId: rawCourseId } = Route.useParams();
  const { data } = useSuspenseQuery(admin.detail(courseId(rawCourseId)));
  return <CourseEditorActions course={data} actor={adminActor} />;
}

function CourseEditorActions({ course, actor }: { course: CourseTree; actor: AuthenticatedUser }) {
  const { admin } = Route.useRouteContext();
  const queryClient = useQueryClient();
  const navigate = Route.useNavigate();
  const save = useMutation(admin.save(queryClient));
  const transition = useMutation(admin.transition(queryClient));
  const remove = useMutation(admin.delete(queryClient));
  const mintVideo = useMutation(admin.videos.mint(queryClient));
  const retryVideo = useMutation(admin.videos.retry(queryClient));
  const transferVideo = useMutation(admin.videos.transfer(queryClient));
  const cancelVideo = useMutation(admin.videos.cancelUpload(queryClient));
  const [activeTransfers, setActiveTransfers] = useState(0);
  const videos = useQuery(admin.videos.list(course.id, activeTransfers > 0));
  const videosError = videoErrorCode(videos.error);
  const authenticationRequired = () => queryClient.setQueryData(sessionQueryKey, null);

  useEffect(() => {
    if (videoErrorCode(videos.error) === 'authentication_required')
      queryClient.setQueryData(sessionQueryKey, null);
  }, [queryClient, videos.error]);

  async function videoAction(action: () => Promise<void>): Promise<LessonVideoFieldActionResult> {
    try {
      await action();
      return null;
    } catch (error) {
      const code = videoErrorCode(error) ?? 'dependency_unavailable';
      if (code === 'authentication_required') authenticationRequired();
      return code;
    }
  }

  async function transferFile(
    target: VideoTransferTarget,
    file: File,
    onProgress: (progress: VideoTransferProgress) => void,
  ) {
    setActiveTransfers((count) => count + 1);
    try {
      await transferVideo.mutateAsync({ courseId: course.id, target, file, onProgress });
    } finally {
      setActiveTransfers((count) => count - 1);
    }
  }

  const videoActions: LessonVideoFieldActions = {
    upload: (lessonId, file, onProgress) =>
      videoAction(async () => {
        const ticket = await mintVideo.mutateAsync({
          courseId: course.id,
          lessonId,
          declaredBytes: file.size,
          originalName: file.name,
        });
        await transferFile({ kind: 'new', ticket }, file, onProgress);
      }),
    resume: (video, file, onProgress) =>
      videoAction(async () => {
        await transferFile(
          {
            kind: 'resume',
            videoId: video.videoId,
            declaredBytes: video.declaredBytes,
          },
          file,
          onProgress,
        );
      }),
    retry: (video) =>
      videoAction(async () => {
        await retryVideo.mutateAsync({ courseId: course.id, videoId: video.videoId });
      }),
    cancel: (video) =>
      videoAction(async () => {
        await cancelVideo.mutateAsync({ courseId: course.id, videoId: video.videoId });
      }),
  };
  const actions = {
    onSave: (current: CourseTree, edit: CourseTreeEdit) => save.mutateAsync({ current, edit }),
    onTransition: (current: CourseTree, status: CourseStatus) =>
      transition.mutateAsync({ current, status }),
    onDelete: async (current: CourseTree) => {
      await remove.mutateAsync(current);
      await navigate({ to: '/admin' });
    },
    onAuthenticationRequired: authenticationRequired,
    video: {
      rows: videos.data ?? [],
      isLoading: videos.isPending,
      ...(videosError === undefined ? {} : { error: videosError }),
      actions: videoActions,
    },
  };

  return actor.role === 'admin' ? (
    <AdminOwnerEditor course={course} actor={actor} {...actions} />
  ) : (
    <CourseEditorView course={course} actor={actor} ownerOptions={null} {...actions} />
  );
}

function AdminOwnerEditor({
  course,
  actor,
  onSave,
  onTransition,
  onDelete,
  onAuthenticationRequired,
  video,
}: {
  course: CourseTree;
  actor: AuthenticatedUser;
  onSave: (current: CourseTree, edit: CourseTreeEdit) => Promise<CourseTree>;
  onTransition: (current: CourseTree, status: CourseStatus) => Promise<CourseTree>;
  onDelete: (current: CourseTree) => Promise<void>;
  onAuthenticationRequired: () => void;
  video: CourseEditorVideo;
}) {
  const { admin } = Route.useRouteContext();
  const transcripts = admin.transcripts;
  const instructors = useQuery(admin.instructors());
  const queryClient = useQueryClient();
  const documentsQuery = transcripts.documents(course.id);
  const documentsKey = documentsQuery.queryKey;
  const documents = useQuery(documentsQuery);
  // Declared as a hook because it is one: it runs inside the field's render, and a name that hid
  // that would also hide it from the rules that verify hook usage.
  function useLessonTranscript(lessonId: string) {
    const query = useQuery(transcripts.lesson(lessonId));
    return { data: query.data, isLoading: query.isLoading };
  }
  return (
    <CourseEditorView
      course={course}
      actor={actor}
      ownerOptions={instructors.data ?? null}
      onSave={onSave}
      onTransition={onTransition}
      onDelete={onDelete}
      onAuthenticationRequired={onAuthenticationRequired}
      video={video}
      transcripts={{
        // Loaded per lesson rather than per course: a course may hold 120 lessons of 5000 lines
        // each, and the editor needs exactly one of them at a time.
        useLessonTranscript,
        documents: documents.data ?? [],
        library: {
          add: async (document) => {
            const result = await transcripts.transcript.addCourseDocument(course.id, document);
            if (result.ok) await queryClient.invalidateQueries({ queryKey: documentsKey });
            return result.ok ? null : result.error.code;
          },
          remove: async (documentId) => {
            const result = await transcripts.transcript.removeCourseDocument(course.id, documentId);
            if (result.ok) await queryClient.invalidateQueries({ queryKey: documentsKey });
            return result.ok ? null : result.error.code;
          },
        },
        save: async (lessonId, language, expectedUpdatedAt, segments) => {
          const result = await transcripts.transcript.saveLessonTranscript(lessonId, {
            language,
            ...(expectedUpdatedAt === undefined ? {} : { expectedUpdatedAt }),
            segments,
          });
          return result.ok ? null : result.error.code;
        },
      }}
    />
  );
}
