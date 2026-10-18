import { useMutation, useQuery, useQueryClient, useSuspenseQuery } from '@tanstack/react-query';
import { createFileRoute } from '@tanstack/react-router';
import { adminCourseListKey } from '../app/composition/admin-queries';
import { ensureRouteQuery } from '../app/composition/catalog-queries';
import { sessionQueryKey } from '../app/composition/identity-queries';
import { adminCourseQuery, adminListState, parseAdminSearch } from '../app/router/admin-search';
import type {
  AdminCourseDraft,
  AdminCoursePage,
  AdminCourseSort,
  CourseStatus,
  CourseTree,
} from '../domain/course/course';
import type { AuthenticatedUser } from '../modules/identity';
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
  const queryClient = useQueryClient();
  const deleteCourse = useMutation(admin.deleteListed(queryClient));

  // One place to merge a change over the current state, dropping every field at its default
  // so the URL stays clean and no filter is silently lost when another changes.
  function go(next: {
    status?: CourseStatus | undefined;
    search?: string;
    sort?: AdminCourseSort;
    page?: number;
  }) {
    const merged = { ...state, ...next };
    void navigate({
      search: {
        ...(merged.status === undefined ? {} : { status: merged.status }),
        ...(merged.search === '' ? {} : { search: merged.search }),
        ...(merged.sort === 'recent' ? {} : { sort: merged.sort }),
        ...(merged.page === 1 ? {} : { page: merged.page }),
      },
    });
  }

  // Changing a filter returns to page 1; paging keeps the filters.
  const changeStatus = (status: CourseStatus | undefined) => go({ status, page: 1 });
  const changeSearch = (search: string) => go({ search, page: 1 });
  const changeSort = (sort: AdminCourseSort) => go({ sort, page: 1 });
  const changePage = (page: number) => go({ page });
  // Drop search and status in one navigation so a filtered-empty list clears to the full list.
  const clearFilters = () => go({ status: undefined, search: '', page: 1 });

  return (
    <CourseListView
      page={data}
      state={state}
      onStatusChange={changeStatus}
      onSearch={changeSearch}
      onSortChange={changeSort}
      onPageChange={changePage}
      onClearFilters={clearFilters}
      onDelete={async (course) => {
        await deleteCourse.mutateAsync(course);
      }}
      onAuthenticationRequired={() => queryClient.setQueryData(sessionQueryKey, null)}
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
