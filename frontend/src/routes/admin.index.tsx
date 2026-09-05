import { useSuspenseQuery } from '@tanstack/react-query';
import { createFileRoute } from '@tanstack/react-router';
import { ensureRouteQuery } from '../app/composition/catalog-queries';
import { adminCourseQuery, adminListState, parseAdminSearch } from '../app/router/admin-search';
import type { CourseStatus } from '../domain/course/course';
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
  const { admin } = Route.useRouteContext();
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
    />
  );
}
