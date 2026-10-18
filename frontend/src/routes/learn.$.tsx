import { useCallback, useMemo, useState } from 'react';
import {
  useInfiniteQuery,
  useMutation,
  useQuery,
  useQueryClient,
  useSuspenseQuery,
} from '@tanstack/react-query';
import { createFileRoute, notFound, redirect, useNavigate } from '@tanstack/react-router';
import { ensureRouteQuery } from '../app/composition/catalog-queries';
import { learningCourseQueryKey } from '../app/composition/learn-queries';
import { parseCourseRoute, parseLessonRoute } from '../app/router/learn-route';
import { lessonInCourse } from '../modules/playback';
import type { LearningCourse, LearningLesson } from '../modules/playback';
import { PurchaseQueryError } from '../app/composition/purchase-queries';
import { LockedLessonPanel } from '../presentation/learn/LockedLessonPanel';
import { CourseLearningView } from '../presentation/learn/CourseLearningView';
import { LessonPlayer } from '../modules/playback';
import { LessonDocumentView } from '../modules/document';
import { TranscriptPanel } from '../presentation/learn/TranscriptPanel';
import { DiscussionPanel } from '../modules/discussion';
import { LearnRouteFailure } from '../modules/identity';
import { toWebVTT } from '../domain/transcript/transcript';
import { LearnEmptyState } from './learn';
import { DiscussionQueryError } from '../app/composition/discussion-queries';
import type { DiscussionErrorCode, DiscussionPage } from '../modules/discussion';
import { ReviewBlock } from '../presentation/learn/ReviewBlock';
import {
  EmailVerificationGate,
  shouldGateLessonPlayback,
} from '../presentation/learn/EmailVerificationGate';

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
  const { data: course } = useSuspenseQuery(learn.course(slug));
  const found = lessonInCourse(course, lessonId);
  if (!found.ok) throw notFound();
  // A locked lesson mounts no player, transcript or discussion: their requests would be
  // refused by the server anyway, and the viewer is shown the way to buy instead.
  if (found.value.locked)
    return (
      <CourseLearningView course={course} currentLessonId={lessonId}>
        <LockedLesson course={course} lesson={found.value} />
      </CourseLearningView>
    );
  return <OpenLesson slug={slug} course={course} lesson={found.value} />;
}

function LockedLesson({
  course,
  lesson,
}: {
  readonly course: LearningCourse;
  readonly lesson: LearningLesson;
}) {
  const { learn } = Route.useRouteContext();
  const queryClient = useQueryClient();
  const navigate = useNavigate();
  const start = useMutation(learn.purchase.start(queryClient));
  const error =
    start.error instanceof PurchaseQueryError
      ? start.error.code
      : start.error === null
        ? undefined
        : 'dependency_unavailable';
  const buy = () =>
    start.mutate(
      { courseId: course.id },
      {
        onSuccess: (order) =>
          void navigate({
            to: '/orders/$orderId',
            params: { orderId: order.id },
            search: { course: course.slug },
          }),
        onError: (failure) => {
          // Bought in another tab, or by a grant: re-read the outline so the lessons unlock.
          if (failure instanceof PurchaseQueryError && failure.code === 'course_already_owned')
            void queryClient.invalidateQueries({
              queryKey: learningCourseQueryKey(course.slug),
              exact: true,
            });
        },
      },
    );
  return (
    <LockedLessonPanel
      courseTitle={course.title}
      lessonTitle={lesson.title}
      price={course.price}
      isStarting={start.isPending}
      {...(error === undefined || error === 'course_already_owned' ? {} : { error })}
      onBuy={buy}
    />
  );
}

function OpenLesson({
  slug,
  course,
  lesson,
}: {
  readonly slug: string;
  readonly course: LearningCourse;
  readonly lesson: LearningLesson;
}) {
  const { learn, identity, identityDependencies } = Route.useRouteContext();
  const queryClient = useQueryClient();
  const navigate = useNavigate();
  const startPurchase = useMutation(learn.purchase.start(queryClient));
  const resendVerification = useMutation({
    mutationFn: () => identityDependencies.session.resend({}).then(identityDependencies.unwrap),
  });
  const lessonId = lesson.id;
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
  // The document body is additive and not suspended on either: a reader failure must not stop the
  // lesson, and an absent publication returns an empty document that renders as nothing.
  const { data: lessonDocument } = useQuery(learn.lessonDocument(lessonId));
  const discussionQuery = useInfiniteQuery(learn.discussion.lesson(lessonId));
  const ownReview = useQuery(
    learn.ownReview(course.id, identity.status === 'signed-in' ? identity.user.id : 'anonymous'),
  );
  const reviewsQuery = useInfiniteQuery(learn.reviews(course.id));
  const putReview = useMutation(learn.putReview(queryClient));
  const deleteReview = useMutation(learn.deleteReview(queryClient));
  const [discussionMutationError, setDiscussionMutationError] = useState<DiscussionErrorCode>();
  const publishDiscussion = useMutation(learn.discussion.publish(queryClient));
  const updateDiscussion = useMutation(learn.discussion.update(queryClient));
  const removeDiscussion = useMutation(learn.discussion.remove(queryClient));
  const reportDiscussion = useMutation(learn.discussion.report(queryClient));
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
    <CourseLearningView
      course={course}
      currentLessonId={lessonId}
      buying={startPurchase.isPending}
      buyError={
        startPurchase.error instanceof PurchaseQueryError
          ? startPurchase.error.code
          : startPurchase.error === null
            ? undefined
            : 'dependency_unavailable'
      }
      onBuy={() =>
        startPurchase.mutate(
          { courseId: course.id },
          {
            onSuccess: (order) =>
              void navigate({
                to: '/orders/$orderId',
                params: { orderId: order.id },
                search: { course: course.slug },
              }),
          },
        )
      }
    >
      {shouldGateLessonPlayback(identity) ? (
        <EmailVerificationGate
          isPending={resendVerification.isPending}
          onResend={() => resendVerification.mutate()}
        />
      ) : (
        <LessonPlayer
          courseTitle={course.title}
          lesson={lesson}
          actions={actions}
          createMedia={learn.createMedia}
          {...(captions === undefined ? {} : { captions })}
          {...(transcript?.language == null ? {} : { captionsLanguage: transcript.language })}
          onPlaybackChannel={onPlaybackChannel}
        />
      )}
      {lessonDocument === undefined ? null : <LessonDocumentView document={lessonDocument} />}
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
        onReport={(commentId) => reportDiscussion.mutateAsync({ lessonId, commentId })}
      />
      <ReviewBlock
        own={ownReview.data}
        ownError={ownReview.isError}
        onRetryOwn={() => void ownReview.refetch()}
        reviews={
          reviewsQuery.data
            ? {
                rating: reviewsQuery.data.pages[0]?.rating ?? null,
                items: reviewsQuery.data.pages.flatMap((page) => page.items),
                nextCursor: reviewsQuery.data.pages.at(-1)?.nextCursor ?? null,
              }
            : undefined
        }
        // Section error only without data: TanStack v5 also sets an error for a failed next page, and that
        // must keep the reviews already shown (nextPageError carries the inline retry).
        reviewsError={reviewsQuery.isError && reviewsQuery.data === undefined}
        hasMore={reviewsQuery.hasNextPage}
        loadingMore={reviewsQuery.isFetchingNextPage}
        nextPageError={reviewsQuery.isFetchNextPageError}
        onMore={() => void reviewsQuery.fetchNextPage()}
        onRetry={() => void reviewsQuery.refetch()}
        onRetryMore={() => void reviewsQuery.fetchNextPage()}
        onSave={(stars, body) =>
          putReview.mutateAsync({ courseId: course.id, stars, body }).then(() => undefined)
        }
        onDelete={() => deleteReview.mutateAsync(course.id)}
      />
    </CourseLearningView>
  );
}
