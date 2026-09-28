import { useCallback, useMemo, useState } from 'react';
import {
  useInfiniteQuery,
  useMutation,
  useQuery,
  useQueryClient,
  useSuspenseQuery,
} from '@tanstack/react-query';
import { createFileRoute, notFound, redirect } from '@tanstack/react-router';
import { ensureRouteQuery } from '../app/composition/catalog-queries';
import { learningCourseQueryKey } from '../app/composition/learn-queries';
import { parseCourseRoute, parseLessonRoute } from '../app/router/learn-route';
import { lessonInCourse } from '../domain/playback/playback';
import { CourseLearningView } from '../presentation/learn/CourseLearningView';
import { LessonPlayer } from '../presentation/learn/LessonPlayer';
import { TranscriptPanel } from '../presentation/learn/TranscriptPanel';
import { DiscussionPanel } from '../presentation/learn/DiscussionPanel';
import { LearnRouteFailure } from '../presentation/identity/ProtectedRouteStates';
import { toWebVTT } from '../domain/transcript/transcript';
import { LearnEmptyState } from './learn';
import { DiscussionQueryError } from '../app/composition/discussion-queries';
import type { DiscussionErrorCode, DiscussionPage } from '../domain/discussion/discussion';

export const Route = createFileRoute('/learn/$')({
  ssr: false,
  beforeLoad: async ({ context, params, abortController }) => {
    const courseSlug = parseCourseRoute(params._splat);
    if (courseSlug !== null) {
      const course = await ensureRouteQuery(
        context.queryClient,
        context.learn.course(courseSlug),
        abortController.signal,
      );
      if (course.continueLessonId === null) return { lessonRoute: null };
      throw redirect({
        to: '/learn/$',
        params: { _splat: `${course.slug}/${course.continueLessonId}` },
        replace: true,
      });
    }
    const lessonRoute = parseLessonRoute(params._splat);
    if (lessonRoute === null) return { lessonRoute: null };
    context.learn.primeAv1Probe();
    const course = await ensureRouteQuery(
      context.queryClient,
      context.learn.course(lessonRoute.slug),
      abortController.signal,
    );
    if (!lessonInCourse(course, lessonRoute.lessonId).ok) throw notFound();
    return { lessonRoute };
  },
  component: LessonRoute,
  errorComponent: LearnRouteFailure,
});

function LessonRoute() {
  const { lessonRoute } = Route.useRouteContext();
  if (lessonRoute === null) return <LearnEmptyState />;
  return <LessonRoutePlayer slug={lessonRoute.slug} lessonId={lessonRoute.lessonId} />;
}

function LessonRoutePlayer({
  slug,
  lessonId,
}: {
  readonly slug: string;
  readonly lessonId: string;
}) {
  const { learn } = Route.useRouteContext();
  const queryClient = useQueryClient();
  const { data: course } = useSuspenseQuery(learn.course(slug));
  const found = lessonInCourse(course, lessonId);
  if (!found.ok) throw notFound();
  const actions = useMemo(
    () => ({
      play: () => learn.playback.play(lessonId),
      heartbeat: (heartbeat: Parameters<typeof learn.playback.heartbeat>[0]) =>
        learn.playback.heartbeat(heartbeat),
      refresh: async () => {
        await queryClient.invalidateQueries({
          queryKey: learningCourseQueryKey(slug),
          exact: true,
        });
        await queryClient.refetchQueries({
          queryKey: learningCourseQueryKey(slug),
          exact: true,
          type: 'active',
        });
      },
    }),
    [learn, lessonId, queryClient, slug],
  );
  // The transcript is not suspended on: a lesson must stay watchable when its transcript fails
  // to load. A11 says the refusal does not stop playback, so this query is allowed to be absent.
  const { data: transcript } = useQuery(learn.lesson(lessonId));
  const discussionQuery = useInfiniteQuery(learn.discussion.lesson(lessonId));
  const [discussionMutationError, setDiscussionMutationError] = useState<DiscussionErrorCode>();
  const publishDiscussion = useMutation(learn.discussion.publish(queryClient));
  const updateDiscussion = useMutation(learn.discussion.update(queryClient));
  const removeDiscussion = useMutation(learn.discussion.remove(queryClient));
  const discussionPage = useMemo<DiscussionPage | undefined>(() => {
    const pages = discussionQuery.data?.pages;
    if (pages === undefined || pages.length === 0) return undefined;
    return {
      threads: pages.flatMap((page) => page.threads),
      nextCursor: pages[pages.length - 1]?.nextCursor ?? null,
    };
  }, [discussionQuery.data]);
  const runDiscussionMutation = useCallback(async (action: () => Promise<unknown>) => {
    setDiscussionMutationError(undefined);
    try {
      await action();
    } catch (error) {
      setDiscussionMutationError(
        error instanceof DiscussionQueryError ? error.code : 'dependency_unavailable',
      );
    }
  }, []);
  const captions = useMemo(
    () =>
      transcript === undefined || transcript.segments.length === 0
        ? undefined
        : toWebVTT(transcript.segments),
    [transcript],
  );
  const [channel, setChannel] = useState<{
    positionMs: number;
    seek: ((positionMs: number) => void) | null;
  }>({ positionMs: 0, seek: null });
  const onPlaybackChannel = useCallback(
    (next: { positionMs: number; seek: ((positionMs: number) => void) | null }) => {
      setChannel((current) =>
        current.positionMs === next.positionMs && current.seek === next.seek ? current : next,
      );
    },
    [],
  );
  const discussionError =
    discussionMutationError ??
    (discussionQuery.error instanceof DiscussionQueryError
      ? discussionQuery.error.code
      : discussionQuery.error === null || discussionQuery.error === undefined
        ? undefined
        : 'dependency_unavailable');
  return (
    <CourseLearningView course={course} currentLessonId={lessonId}>
      <LessonPlayer
        courseTitle={course.title}
        lesson={found.value}
        actions={actions}
        createMedia={learn.createMedia}
        {...(captions === undefined ? {} : { captions })}
        {...(transcript?.language == null ? {} : { captionsLanguage: transcript.language })}
        onPlaybackChannel={onPlaybackChannel}
      />
      {transcript === undefined ? null : (
        <TranscriptPanel
          transcript={transcript}
          positionMs={channel.positionMs}
          {...(channel.seek === null ? {} : { onSeek: channel.seek })}
        />
      )}
      <DiscussionPanel
        page={discussionPage}
        {...(discussionError === undefined ? {} : { error: discussionError })}
        isLoading={discussionQuery.isLoading}
        isLoadingMore={discussionQuery.isFetchingNextPage}
        onRetry={() => void discussionQuery.refetch()}
        {...(discussionQuery.hasNextPage
          ? { onLoadMore: () => void discussionQuery.fetchNextPage() }
          : {})}
        positionMs={channel.positionMs}
        {...(channel.seek === null ? {} : { onSeek: channel.seek })}
        onPublish={(input) =>
          runDiscussionMutation(() => publishDiscussion.mutateAsync({ lessonId, input }))
        }
        onUpdate={(commentId, body) =>
          runDiscussionMutation(() =>
            updateDiscussion.mutateAsync({ lessonId, input: { commentId, body } }),
          )
        }
        onRemove={(commentId) =>
          runDiscussionMutation(() => removeDiscussion.mutateAsync({ lessonId, commentId }))
        }
      />
    </CourseLearningView>
  );
}
