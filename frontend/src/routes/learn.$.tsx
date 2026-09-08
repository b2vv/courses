import { useMemo } from 'react';
import { useQueryClient, useSuspenseQuery } from '@tanstack/react-query';
import { createFileRoute, notFound } from '@tanstack/react-router';
import { ensureRouteQuery } from '../app/composition/catalog-queries';
import { learningCourseQueryKey } from '../app/composition/learn-queries';
import { parseLessonRoute } from '../app/router/learn-route';
import { lessonInCourse } from '../domain/playback/playback';
import { LessonPlayer } from '../presentation/learn/LessonPlayer';
import { LearnEmptyState } from './learn';

export const Route = createFileRoute('/learn/$')({
  ssr: false,
  beforeLoad: async ({ context, params, abortController }) => {
    const lessonRoute = parseLessonRoute(params._splat);
    if (lessonRoute === null) return { lessonRoute: null };
    const course = await ensureRouteQuery(
      context.queryClient,
      context.learn.course(lessonRoute.slug),
      abortController.signal,
    );
    if (!lessonInCourse(course, lessonRoute.lessonId).ok) throw notFound();
    return { lessonRoute };
  },
  component: LessonRoute,
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
  return (
    <LessonPlayer
      courseTitle={course.title}
      lesson={found.value}
      actions={actions}
      createMedia={learn.createMedia}
    />
  );
}
