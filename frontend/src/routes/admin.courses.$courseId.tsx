import {
  keepPreviousData,
  useMutation,
  useQuery,
  useQueryClient,
  useSuspenseQuery,
} from '@tanstack/react-query';
import { createFileRoute, notFound } from '@tanstack/react-router';
import { useEffect, useState } from 'react';
import { DraftQueryError } from '../app/composition/admin-draft-queries';
import { AuthoringQueryError } from '../app/composition/admin-queries';
import { VideoQueryError } from '../app/composition/admin-video-queries';
import { ensureRouteQuery } from '../app/composition/catalog-queries';
import { sessionQueryKey } from '../app/composition/identity-queries';
import { asBrandedId, validEntityId } from '../domain/course/course';
import type {
  CourseId,
  CoursePrice,
  CourseStatus,
  CourseTree,
  CourseTreeEdit,
} from '../domain/course/course';
import type { DraftErrorCode } from '../domain/transcript/draft';
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
import { CourseGrantPanel } from '../presentation/admin/CourseGrantPanel';
import { GrantQueryError } from '../app/composition/admin-grant-queries';
import type { GrantErrorCode } from '../domain/purchase/grant';

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
  const setPrice = useMutation(admin.setPrice(queryClient));
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
    onSetPrice: (current: CourseTree, price: CoursePrice | null) =>
      setPrice.mutateAsync({ current, price }),
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
  onSetPrice,
  onTransition,
  onDelete,
  onAuthenticationRequired,
  video,
}: {
  course: CourseTree;
  actor: AuthenticatedUser;
  onSave: (current: CourseTree, edit: CourseTreeEdit) => Promise<CourseTree>;
  onSetPrice: (current: CourseTree, price: CoursePrice | null) => Promise<CourseTree>;
  onTransition: (current: CourseTree, status: CourseStatus) => Promise<CourseTree>;
  onDelete: (current: CourseTree) => Promise<void>;
  onAuthenticationRequired: () => void;
  video: CourseEditorVideo;
}) {
  const { admin, transcriptHistory } = Route.useRouteContext();
  const transcripts = admin.transcripts;
  const draftQueries = admin.drafts;
  const instructors = useQuery(admin.instructors());
  const queryClient = useQueryClient();
  const regenerateDraft = useMutation(draftQueries.regenerate(queryClient));
  const documentsQuery = transcripts.documents(course.id);
  const documentsKey = documentsQuery.queryKey;
  const documents = useQuery(documentsQuery);
  function useLessonTranscript(lessonId: string) {
    const query = useQuery(transcripts.lesson(lessonId));
    return { data: query.data, isLoading: query.isLoading };
  }
  function useLessonDraft(lessonId: string) {
    const query = useQuery(draftQueries.lesson(lessonId, queryClient));
    return { data: query.data, isLoading: query.isLoading };
  }
  async function draftRegenerate(
    lessonId: string,
    language: string | undefined,
  ): Promise<DraftErrorCode | null> {
    try {
      await regenerateDraft.mutateAsync({ lessonId, language });
      return null;
    } catch (error) {
      const code =
        error instanceof DraftQueryError ? error.code : ('dependency_unavailable' as const);
      if (code === 'authentication_required') onAuthenticationRequired();
      return code;
    }
  }
  return (
    <CourseEditorView
      course={course}
      actor={actor}
      ownerOptions={instructors.data ?? null}
      adminTools={
        <CourseGrantTool courseId={course.id} onAuthenticationRequired={onAuthenticationRequired} />
      }
      onSave={onSave}
      onSetPrice={onSetPrice}
      onTransition={onTransition}
      onDelete={onDelete}
      onAuthenticationRequired={onAuthenticationRequired}
      video={video}
      drafts={{
        useLessonDraft,
        actions: { regenerate: draftRegenerate },
      }}
      transcripts={{
        // Loaded per lesson rather than per course: a course may hold 120 lessons of 5000 lines
        // each, and the editor needs exactly one of them at a time.
        useLessonTranscript,
        documents: documents.data ?? [],
        history: { store: transcriptHistory, authorId: actor.id },
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

const userSearchDebounceMs = 250;

/** Rendered only in the admin branch; the server refuses everyone else with grant_forbidden. */
function CourseGrantTool({
  courseId,
  onAuthenticationRequired,
}: {
  courseId: string;
  onAuthenticationRequired: () => void;
}) {
  const { admin } = Route.useRouteContext();
  const [typed, setTyped] = useState('');
  const [search, setSearch] = useState('');
  // One request per pause in typing, not per keystroke: a burst of keystrokes is a burst of
  // requests against a rate-limited API.
  useEffect(() => {
    const timer = setTimeout(() => setSearch(typed.trim()), userSearchDebounceMs);
    return () => clearTimeout(timer);
  }, [typed]);
  const users = useQuery({
    ...admin.users.list({ q: search, limit: 20, offset: 0, status: 'active' }),
    placeholderData: keepPreviousData,
  });
  const grant = useMutation(admin.grants.grant());
  const options = (users.data?.items ?? []).map((user) => ({
    id: user.id,
    label: `${user.displayName} — ${user.email}`,
  }));
  async function onGrant(userId: string, reason: string): Promise<GrantErrorCode | null> {
    try {
      await grant.mutateAsync({ courseId, userId, reason });
      return null;
    } catch (error) {
      const code = error instanceof GrantQueryError ? error.code : 'dependency_unavailable';
      if (code === 'authentication_required') onAuthenticationRequired();
      return code;
    }
  }
  return <CourseGrantPanel users={options} onSearch={setTyped} onGrant={onGrant} />;
}
