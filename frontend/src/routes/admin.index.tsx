import { useMutation, useQuery, useQueryClient, useSuspenseQuery } from '@tanstack/react-query';
import { createFileRoute } from '@tanstack/react-router';
import { adminCourseListKey } from '../app/composition/admin-queries';
import { ensureRouteQuery } from '../app/composition/catalog-queries';
import { sessionQueryKey } from '../app/composition/identity-queries';
import { adminCourseQuery, adminListState, parseAdminSearch } from '../app/router/admin-search';
import type {
  AdminCourseDraft,
  AdminCoursePage,
  CourseStatus,
  CourseTree,
} from '../domain/course/course';
import type { AuthenticatedUser } from '../domain/identity/identity';
import { CourseCreateDialog } from '../presentation/admin/CourseEditorView';
import {
  CourseListFailure,
  CourseListPending,
  CourseListView,
} from '../presentation/admin/CourseListView';

export const Route = createFileRoute('/admin/')({
  ssr: false,
  validateSearch: parseAdminSearch,
  loaderDeps: ({ search }) => adminListState(search),
  loader: ({ context, deps, abortController }) =>
    ensureRouteQuery(
      context.queryClient,
      context.admin.list(adminCourseQuery(deps)),
      abortController.signal,
    ),
  head: () => ({ meta: [{ title: 'Адміністрування курсів — CourseForge' }] }),
  component: AdminCourseListRoute,
  pendingComponent: CourseListPending,
  errorComponent: CourseListFailure,
});

function AdminCourseListRoute() {
  const state = adminListState(Route.useSearch());
  const { admin, adminActor } = Route.useRouteContext();
  const { data } = useSuspenseQuery(admin.list(adminCourseQuery(state)));
  const navigate = Route.useNavigate();

  function changeStatus(status: CourseStatus | undefined) {
    void navigate({
      search: status === undefined ? {} : { status },
    });
  }

  function changePage(page: number) {
    void navigate({
      search: {
        ...(state.status === undefined ? {} : { status: state.status }),
        ...(page === 1 ? {} : { page }),
      },
    });
  }

  return (
    <CourseListView
      page={data}
      state={state}
      onStatusChange={changeStatus}
      onPageChange={changePage}
      createAction={<CourseCreateAction actor={adminActor} />}
    />
  );
}

function CourseCreateAction({ actor }: { actor: AuthenticatedUser }) {
  const { admin } = Route.useRouteContext();
  const queryClient = useQueryClient();
  const navigate = Route.useNavigate();
  const create = useMutation(admin.create(queryClient));

  async function createCourse(draft: AdminCourseDraft): Promise<CourseTree> {
    const course = await create.mutateAsync(draft);
    await navigate({
      to: '/admin/courses/$courseId',
      params: { courseId: course.id },
    });
    return course;
  }

  async function findExisting(slug: string) {
    await queryClient.invalidateQueries({ queryKey: adminCourseListKey });
    return queryClient
      .getQueriesData<AdminCoursePage>({ queryKey: adminCourseListKey })
      .flatMap(([, page]) => page?.items ?? [])
      .find((course) => course.slug === slug)?.id;
  }

  const props = {
    actor,
    onCreate: createCourse,
    onSlugTaken: findExisting,
    onAuthenticationRequired: () => queryClient.setQueryData(sessionQueryKey, null),
  };
  return actor.role === 'admin' ? (
    <AdminCourseCreateDialog {...props} />
  ) : (
    <CourseCreateDialog {...props} ownerOptions={null} />
  );
}

function AdminCourseCreateDialog({
  actor,
  onCreate,
  onSlugTaken,
  onAuthenticationRequired,
}: {
  actor: AuthenticatedUser;
  onCreate: (draft: AdminCourseDraft) => Promise<CourseTree>;
  onSlugTaken: (slug: string) => Promise<CourseTree['id'] | undefined>;
  onAuthenticationRequired: () => void;
}) {
  const { admin } = Route.useRouteContext();
  const instructors = useQuery(admin.instructors());
  return (
    <CourseCreateDialog
      actor={actor}
      ownerOptions={instructors.data ?? null}
      onCreate={onCreate}
      onSlugTaken={onSlugTaken}
      onAuthenticationRequired={onAuthenticationRequired}
    />
  );
}
