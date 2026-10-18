import { afterAll, afterEach, beforeAll, expect, test } from '@rstest/core';
import { focusManager, QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { createMemoryHistory, createRouter, RouterProvider } from '@tanstack/react-router';
import { render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { http, HttpResponse } from 'msw';
import { setupServer } from 'msw/node';
import { apiOrigin } from '../adapters/api/api-origin';
import { createApiSessionRepository } from '../modules/identity';
import { createAuthoring } from '../modules/admin';
import { createLessonVideos } from '../modules/media';
import { createUserAdministration } from '../modules/admin';
import { createCatalog } from '../modules/catalog';
import { createSession } from '../modules/identity';
import { createAdminQueries } from '../app/composition/admin-queries';
import { createAdminUserQueries } from '../app/composition/admin-user-queries';
import { createAdminVideoQueries } from '../app/composition/admin-video-queries';
import { createCatalogQueries } from '../app/composition/catalog-queries';
import { createSessionQueries, sessionQueryKey } from '../app/composition/identity-queries';
import { createLearnQueries, learningCourseQueryKey } from '../app/composition/learn-queries';
import { createMemoryTranscriptHistory } from '../modules/transcript';
import { createDraftQueries } from '../app/composition/admin-draft-queries';
import {
  createAdminDocumentQueries,
  createDocumentQueries,
  type DocumentRepository,
} from '../modules/document';
import { createTranscriptQueries } from '../app/composition/transcript-queries';
import { createDiscussionQueries } from '../app/composition/discussion-queries';
import { createTeachQueries } from '../app/composition/teach-queries';
import { createPurchaseQueries } from '../app/composition/purchase-queries';
import { createEnrollmentQueries } from '../app/composition/enrollment-queries';
import { createAdminGrantQueries } from '../app/composition/admin-grant-queries';
import { createAdminStatsQueries } from '../app/composition/admin-stats-queries';
import type { PurchaseRepository } from '../modules/purchase';
import type { TeachRepository } from '../modules/teach';
import type { TranscriptRepository } from '../modules/transcript';
import type { DiscussionRepository } from '../modules/discussion';
import type { CreatePlaybackMedia } from '../modules/playback';
import type { LessonTranscript } from '../modules/transcript';
import type { LearningCourse } from '../modules/playback';
import type { PurchaseOrder } from '../modules/purchase';

// Route guards never touch transcripts; a proxy keeps this stub from drifting every time the port
// grows a method.
const transcriptStub = new Proxy({} as TranscriptRepository, {
  get() {
    return () => {
      throw new Error('Transcripts are not used by route guard tests');
    };
  },
});
const documentStub = new Proxy({} as DocumentRepository, {
  get() {
    return () => {
      throw new Error('Lesson documents are not used by route guard tests');
    };
  },
});
const discussionStub = new Proxy({} as DiscussionRepository, {
  get() {
    return () => {
      throw new Error('Discussion is not used by route guard tests');
    };
  },
});
const teachStub = new Proxy({} as TeachRepository, {
  get() {
    return () => {
      throw new Error('Teach is not used by route guard tests');
    };
  },
});
import type { Locale } from '../app/composition/locale';
import { createI18n, I18nProvider } from '../app/providers/i18n';
import { asBrandedId } from '../contracts/course';
import type {
  AdminCourse,
  AdminCourseQuery,
  CourseTree,
  CourseTreeEdit,
} from '../contracts/course';
import type { AdminUser, AdminUserQuery } from '../modules/admin';
import type { AdminCourseRepository } from '../modules/admin';
import type { AdminUserRepository } from '../modules/admin';
import type { CourseRepository } from '../modules/catalog';
import type { LessonVideoRepository } from '../modules/media';
import type { PlaybackRepository } from '../modules/playback';
import type { AuthenticatedUser, Role } from '../modules/identity';
import { routeTree } from '../routeTree.gen';

const currentUserUrl = `${apiOrigin()}/api/v1/auth/me`;
const logoutUrl = `${apiOrigin()}/api/v1/auth/logout`;
const authenticationRequired = {
  error: { code: 'authentication_required', message: 'x', requestId: 'a'.repeat(32) },
};
const dependencyUnavailable = {
  error: { code: 'dependency_unavailable', message: 'x', requestId: 'b'.repeat(32) },
};
const server = setupServer(
  http.get(currentUserUrl, () => HttpResponse.json(authenticationRequired, { status: 401 })),
);
const purchaseStub: PurchaseRepository = {
  start: async () => ({ ok: false, error: { code: 'dependency_unavailable' } }),
  status: async () => ({ ok: false, error: { code: 'dependency_unavailable' } }),
  checkout: async () => ({ ok: false, error: { code: 'dependency_unavailable' } }),
};
const observedAdminQueries: AdminCourseQuery[] = [];
const observedAdminUserQueries: AdminUserQuery[] = [];

beforeAll(() => server.listen({ onUnhandledRequest: 'error' }));

afterEach(() => {
  focusManager.setFocused(undefined);
  observedAdminQueries.length = 0;
  observedAdminUserQueries.length = 0;
  server.resetHandlers();
});

afterAll(() => server.close());

function user(role: Role): AuthenticatedUser {
  return {
    id: `40000000-0000-4000-8000-00000000000${role === 'student' ? '1' : role === 'instructor' ? '2' : '3'}`,
    email: `${role}@courseforge.local`,
    displayName: role,
    emailVerified: true,
    twoFactorEnabled: false,
    role,
  };
}

function createTestRouter(
  href: string,
  cachedSession: AuthenticatedUser | null,
  locale: Locale = 'uk',
  adminCourses: readonly AdminCourse[] = [],
  adminOverrides: Partial<AdminCourseRepository> = {},
  adminUsers: readonly AdminUser[] = [],
  adminUserOverrides: Partial<AdminUserRepository> = {},
  playbackOverrides: Partial<PlaybackRepository> = {},
  transcriptRepository: TranscriptRepository = transcriptStub,
  discussionRepository: DiscussionRepository = discussionStub,
  createMediaOverride?: CreatePlaybackMedia,
  purchaseRepository: PurchaseRepository = purchaseStub,
) {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  const i18n = createI18n(locale);
  const identityDependencies = {
    ...createSessionQueries(createSession(createApiSessionRepository())),
    shouldLoadCurrentUser: true,
  };
  const catalogRepository: CourseRepository = {
    listPublished: async (query) => ({
      ok: true,
      value: {
        items: [],
        pagination: { limit: query.limit, offset: query.offset, total: 0 },
      },
    }),
    getPublished: async () => ({ ok: false, error: { code: 'course_not_found' } }),
    listReviews: async () => ({ ok: true, value: { rating: null, items: [], nextCursor: null } }),
    reportReview: async () => ({ ok: true, value: undefined }),
    putReview: async () => ({ ok: false, error: { code: 'dependency_unavailable' } }),
    deleteReview: async () => ({ ok: true, value: undefined }),
    listLearningReviews: async () => ({
      ok: true,
      value: { rating: null, items: [], nextCursor: null },
    }),
    ownReview: async () => ({
      ok: true,
      value: { eligibility: 'review_not_started', review: null },
    }),
  };
  const adminRepository: AdminCourseRepository = {
    list: async (query) => {
      observedAdminQueries.push(query);
      const filtered = adminCourses.filter(
        (course) =>
          (query.status === undefined || course.status === query.status) &&
          (query.search === '' || course.title.toLowerCase().includes(query.search.toLowerCase())),
      );
      return {
        ok: true,
        value: {
          items: filtered.slice(query.offset, query.offset + query.limit),
          pagination: { limit: query.limit, offset: query.offset, total: filtered.length },
          statusCounts: {
            total: adminCourses.length,
            draft: adminCourses.filter((course) => course.status === 'draft').length,
            published: adminCourses.filter((course) => course.status === 'published').length,
            archived: adminCourses.filter((course) => course.status === 'archived').length,
          },
        },
      };
    },
    load: async () => ({ ok: false, error: { code: 'course_not_found' } }),
    listInstructors: async () => ({ ok: true, value: [] }),
    create: async () => ({ ok: false, error: { code: 'dependency_unavailable' } }),
    save: async () => ({ ok: false, error: { code: 'dependency_unavailable' } }),
    transition: async () => ({ ok: false, error: { code: 'dependency_unavailable' } }),
    delete: async () => ({ ok: false, error: { code: 'dependency_unavailable' } }),
    ...adminOverrides,
  };
  const adminUserRepository: AdminUserRepository = {
    list: async (query) => {
      observedAdminUserQueries.push(query);
      const searched = adminUsers.filter(
        (listed) =>
          (query.status === undefined || listed.status === query.status) &&
          (query.q === '' ||
            listed.email.toLowerCase().includes(query.q.toLowerCase()) ||
            listed.displayName.toLowerCase().includes(query.q.toLowerCase())),
      );
      return {
        ok: true,
        value: {
          items: searched.slice(query.offset, query.offset + query.limit),
          pagination: { limit: query.limit, offset: query.offset, total: searched.length },
        },
      };
    },
    setStatus: async () => ({ ok: true, value: null }),
    delete: async () => ({ ok: true, value: null }),
    resetPassword: async () => ({ ok: true, value: null }),
    ...adminUserOverrides,
  };
  const lessonVideoRepository: LessonVideoRepository = {
    mint: async () => ({ ok: false, error: { code: 'dependency_unavailable' } }),
    listByCourse: async () => ({ ok: true, value: [] }),
    retry: async () => ({ ok: false, error: { code: 'video_not_found' } }),
    transfer: async () => ({ ok: false, error: { code: 'dependency_unavailable' } }),
    cancelUpload: async () => ({ ok: false, error: { code: 'dependency_unavailable' } }),
  };
  const playbackRepository: PlaybackRepository = {
    getCourse: async () => ({ ok: false, error: { code: 'course_not_found' } }),
    play: async () => ({ ok: false, error: { code: 'lesson_not_found' } }),
    heartbeat: async () => ({ ok: false, error: { code: 'playback_session_expired' } }),
    ...playbackOverrides,
  };
  const router = createRouter({
    routeTree,
    history: createMemoryHistory({ initialEntries: [href] }),
    context: {
      queryClient,
      admin: {
        ...createAdminQueries(createAuthoring(adminRepository)),
        users: createAdminUserQueries(createUserAdministration(adminUserRepository)),
        videos: createAdminVideoQueries(createLessonVideos(lessonVideoRepository)),
        transcripts: createTranscriptQueries(transcriptStub),
        grants: createAdminGrantQueries({
          grant: async () => ({ ok: false, error: { code: 'dependency_unavailable' } }),
        }),
        drafts: createDraftQueries({
          getDraft: async () => ({
            ok: true,
            value: {
              lessonId: '',
              state: null,
              language: null,
              failure: null,
              stale: false,
              updatedAt: null,
              assetId: null,
              generation: null,
              segments: [],
            },
          }),
          regenerate: async () => ({ ok: true, value: undefined }),
        }),
        stats: createAdminStatsQueries({
          snapshot: async () => ({ ok: false, error: { code: 'dependency_unavailable' } }),
          timeseries: async () => ({ ok: false, error: { code: 'dependency_unavailable' } }),
        }),
        documents: createAdminDocumentQueries({
          getDraft: async () => ({ ok: false, error: { code: 'dependency_unavailable' } }),
          saveDraft: async () => ({ ok: false, error: { code: 'dependency_unavailable' } }),
          publish: async () => ({ ok: false, error: { code: 'dependency_unavailable' } }),
          uploadImage: async () => ({ ok: false, error: { code: 'dependency_unavailable' } }),
        }),
      },
      // Route guards never open the editor, so the in-memory store is the whole truth here.
      transcriptHistory: createMemoryTranscriptHistory(),
      identityDependencies,
      catalog: createCatalogQueries(createCatalog(catalogRepository)),
      learn: {
        ...createLearnQueries(playbackRepository, catalogRepository),
        ...createTranscriptQueries(transcriptRepository),
        ...createDocumentQueries(documentStub),
        discussion: createDiscussionQueries(discussionRepository),
        purchase: createPurchaseQueries(purchaseRepository),
        primeAv1Probe: () => undefined,
        createMedia:
          createMediaOverride ??
          ((_video, _callbacks) => ({
            kind: 'native',
            load: () => undefined,
            stopLoad: () => undefined,
            selectLevel: () => undefined,
            destroy: () => undefined,
          })),
      },
      teach: createTeachQueries(teachStub),
      enrollment: createEnrollmentQueries({
        membership: async () => ({ ok: true, value: { member: false, state: null } }),
        enroll: async () => ({ ok: false, error: { code: 'dependency_unavailable' } }),
        myMemberships: async () => ({ ok: true, value: [] }),
        roster: async () => ({ ok: true, value: { entries: [], nextCursor: null } }),
        block: async () => ({ ok: true, value: undefined }),
        unblock: async () => ({ ok: true, value: undefined }),
        reportLecturer: async () => ({ ok: true, value: undefined }),
        reportStudent: async () => ({ ok: true, value: undefined }),
        participantReports: async () => ({ ok: true, value: { items: [], nextCursor: null } }),
        resolveReport: async () => ({ ok: true, value: undefined }),
        revoke: async () => ({ ok: true, value: undefined }),
        refundQueue: async () => ({ ok: true, value: { items: [] } }),
        refundOrder: async () => ({ ok: true, value: undefined }),
        declineRefund: async () => ({ ok: true, value: undefined }),
        revokeRefund: async () => ({ ok: true, value: undefined }),
      }),
    },
    defaultPendingMs: 0,
    defaultPendingMinMs: 0,
    Wrap: ({ children }) => <I18nProvider instance={i18n}>{children}</I18nProvider>,
  });
  queryClient.setQueryData(sessionQueryKey, cachedSession);
  return router;
}

function adminTree(owner: AuthenticatedUser): CourseTree {
  return {
    id: asBrandedId('10000000-0000-4000-8000-000000000001', 'course'),
    slug: 'owned-course',
    title: 'Власний курс',
    summary: 'Практичний курс',
    description: 'Опис курсу',
    status: 'draft',
    instructorId: owner.id,
    createdAt: '2026-09-05T10:15:30.000000Z',
    updatedAt: '2026-09-05T10:15:30.000001Z',
    publishedAt: null,
    price: null,
    studentCount: 0,
    modules: [],
  };
}

function renderRouter(router: ReturnType<typeof createTestRouter>) {
  return render(
    <QueryClientProvider client={router.options.context.queryClient}>
      <RouterProvider router={router} />
    </QueryClientProvider>,
  );
}

function answerWith(current: AuthenticatedUser | null) {
  server.use(
    http.get(currentUserUrl, () =>
      current === null
        ? HttpResponse.json(authenticationRequired, { status: 401 })
        : HttpResponse.json(current),
    ),
  );
}

test.each([
  ['/learn', '/login?returnTo=%2Flearn'],
  ['/admin', '/login?returnTo=%2Fadmin'],
])('G01/G02/G13 redirects an anonymous visit to %s', async (href, expected) => {
  answerWith(null);
  const router = createTestRouter(href, null);

  await router.load();

  expect(router.state.location.href).toBe(expected);
});

test.each(['student', 'instructor', 'admin'] satisfies Role[])(
  'R36 shows the %s role their «Моє навчання» list at /learn (enrollment replaced the catalog redirect)',
  async (role) => {
    const signedIn = user(role);
    answerWith(signedIn);
    const router = createTestRouter('/learn', signedIn);

    await router.load();
    renderRouter(router);

    expect(router.state.location.href).toBe('/learn');
    // The harness wires myMemberships → empty, so the list renders its empty state.
    expect(screen.getByText('Ви ще не записалися на жоден курс.')).toBeTruthy();
    expect(screen.queryByText('Сторінку не знайдено')).toBeNull();
    expect(screen.queryByText('Ваші викладацькі курси')).toBeNull();
  },
);

test('R24 keeps the generated route tree equal to the complete expected path list', () => {
  const router = createTestRouter('/', null);
  const routes = Object.values(router.routesById)
    .filter((route) => route.id !== '__root__')
    .map((route) => ({ id: route.id, path: route.path, fullPath: route.fullPath }))
    .sort((left, right) => left.id.localeCompare(right.id));

  expect(routes).toEqual(
    [
      { id: '/', path: '/', fullPath: '/' },
      { id: '/admin', path: 'admin', fullPath: '/admin' },
      { id: '/admin/', path: '/', fullPath: '/admin/' },
      { id: '/admin/$', path: '$', fullPath: '/admin/$' },
      { id: '/admin/analytics', path: 'analytics', fullPath: '/admin/analytics' },
      {
        id: '/admin/courses/$courseId',
        path: 'courses/$courseId',
        fullPath: '/admin/courses/$courseId',
      },
      {
        id: '/admin/courses/$courseId_/lessons/$lessonId',
        path: 'courses/$courseId/lessons/$lessonId',
        fullPath: '/admin/courses/$courseId/lessons/$lessonId',
      },
      {
        id: '/admin/participant-reports',
        path: 'participant-reports',
        fullPath: '/admin/participant-reports',
      },
      { id: '/admin/learners', path: 'learners', fullPath: '/admin/learners' },
      { id: '/admin/refunds', path: 'refunds', fullPath: '/admin/refunds' },
      { id: '/admin/users/', path: 'users/', fullPath: '/admin/users/' },
      { id: '/courses/', path: 'courses/', fullPath: '/courses/' },
      { id: '/courses/$slug', path: 'courses/$slug', fullPath: '/courses/$slug' },
      { id: '/learn', path: 'learn', fullPath: '/learn' },
      { id: '/learn/$', path: '$', fullPath: '/learn/$' },
      { id: '/login', path: 'login', fullPath: '/login' },
      { id: '/register', path: 'register', fullPath: '/register' },
      { id: '/verify-email', path: 'verify-email', fullPath: '/verify-email' },
      { id: '/password/forgot', path: 'password/forgot', fullPath: '/password/forgot' },
      { id: '/password/reset', path: 'password/reset', fullPath: '/password/reset' },
      {
        id: '/email/change/confirm',
        path: 'email/change/confirm',
        fullPath: '/email/change/confirm',
      },
      { id: '/settings/account', path: 'settings/account', fullPath: '/settings/account' },
      { id: '/orders/$orderId', path: 'orders/$orderId', fullPath: '/orders/$orderId' },
      { id: '/teach', path: 'teach', fullPath: '/teach' },
      {
        id: '/teach/courses/$courseId',
        path: 'courses/$courseId',
        fullPath: '/teach/courses/$courseId',
      },
      { id: '/teach/learners', path: 'learners', fullPath: '/teach/learners' },
    ].sort((left, right) => left.id.localeCompare(right.id)),
  );
});

test('R31 keeps the instructor owner read-only without mounting the owner query', async () => {
  const instructor = user('instructor');
  const course = adminTree(instructor);
  let ownerRequests = 0;
  answerWith(instructor);
  const router = createTestRouter(`/admin/courses/${course.id}`, instructor, 'uk', [], {
    load: async () => ({ ok: true, value: course }),
    listInstructors: async () => {
      ownerRequests += 1;
      return { ok: true, value: [] };
    },
  });

  await router.load();
  renderRouter(router);

  expect(await screen.findByRole('heading', { name: 'Власний курс' })).toBeTruthy();
  expect(screen.getAllByText('instructor').length).toBeGreaterThan(0);
  expect(screen.queryByRole('button', { name: /Власник/ })).toBeNull();
  expect(ownerRequests).toBe(0);
});

test('R31 mounts the owner picker query for an administrator', async () => {
  const admin = user('admin');
  const course = adminTree(admin);
  let ownerRequests = 0;
  answerWith(admin);
  const router = createTestRouter(`/admin/courses/${course.id}`, admin, 'uk', [], {
    load: async () => ({ ok: true, value: course }),
    listInstructors: async () => {
      ownerRequests += 1;
      return {
        ok: true,
        value: [{ id: admin.id, displayName: admin.displayName, role: 'admin' }],
      };
    },
  });

  await router.load();
  renderRouter(router);

  expect(await screen.findByRole('button', { name: /Власник/ })).toBeTruthy();
  expect(ownerRequests).toBe(1);
});

test('R32 degrades an unavailable admin owner picker without blocking save', async () => {
  const admin = user('admin');
  const owner = user('instructor');
  const course = adminTree(owner);
  let saves = 0;
  answerWith(admin);
  const router = createTestRouter(`/admin/courses/${course.id}`, admin, 'uk', [], {
    load: async () => ({ ok: true, value: course }),
    listInstructors: async () => ({
      ok: false,
      error: { code: 'dependency_unavailable' },
    }),
    save: async (_id, _updatedAt, edit: CourseTreeEdit) => {
      saves += 1;
      return {
        ok: true,
        value: { ...course, ...edit, updatedAt: '2026-09-05T10:15:30.000002Z' },
      };
    },
  });

  await router.load();
  renderRouter(router);
  expect(await screen.findByText(owner.id)).toBeTruthy();
  expect(screen.queryByRole('button', { name: /Власник/ })).toBeNull();
  await userEvent.click(screen.getByRole('button', { name: 'Зберегти зміни' }));

  await waitFor(() => expect(saves).toBe(1));
});

test('R35 returns to login with the editor address when the session ends mid-edit', async () => {
  const admin = user('admin');
  const course = adminTree(admin);
  answerWith(admin);
  const router = createTestRouter(`/admin/courses/${course.id}`, admin, 'uk', [], {
    load: async () => ({ ok: true, value: course }),
    listInstructors: async () => ({
      ok: true,
      value: [{ id: admin.id, displayName: admin.displayName, role: 'admin' }],
    }),
    save: async () => ({ ok: false, error: { code: 'authentication_required' } }),
  });

  await router.load();
  renderRouter(router);
  await userEvent.click(await screen.findByRole('button', { name: 'Зберегти зміни' }));

  await waitFor(() =>
    expect(router.state.location.href).toBe(`/login?returnTo=%2Fadmin%2Fcourses%2F${course.id}`),
  );
});

test('G04 gives a student the same not-found answer for /admin as an unknown address', async () => {
  const student = user('student');
  answerWith(student);
  const adminRouter = createTestRouter('/admin', student);
  await adminRouter.load();
  const adminScreen = renderRouter(adminRouter);
  const adminAnswer = screen.getByRole('heading', { name: 'Сторінку не знайдено' }).textContent;
  // Refused, the student stays in the client app: its header (with their name)
  // and no admin sidebar — not a bare page with no way back.
  expect(screen.getByText(student.displayName)).toBeTruthy();
  expect(screen.queryByRole('complementary')).toBeNull();
  adminScreen.unmount();

  const unknownRouter = createTestRouter('/unknown', student);
  await unknownRouter.load();
  renderRouter(unknownRouter);

  expect(screen.getByRole('heading', { name: 'Сторінку не знайдено' }).textContent).toBe(
    adminAnswer,
  );
});

test('W guard gives a student the same not-found answer for /teach as an unknown address', async () => {
  const student = user('student');
  answerWith(student);
  const teachRouter = createTestRouter('/teach', student);
  await teachRouter.load();
  const teachScreen = renderRouter(teachRouter);
  const teachAnswer = screen.getByRole('heading', { name: 'Сторінку не знайдено' }).textContent;
  expect(screen.getByText(student.displayName)).toBeTruthy();
  expect(screen.queryByRole('complementary')).toBeNull();
  teachScreen.unmount();

  const unknownRouter = createTestRouter('/unknown', student);
  await unknownRouter.load();
  renderRouter(unknownRouter);
  expect(screen.getByRole('heading', { name: 'Сторінку не знайдено' }).textContent).toBe(
    teachAnswer,
  );
});

test.each(['instructor', 'admin'] satisfies Role[])(
  'G05 opens /admin for the %s role',
  async (role) => {
    const signedIn = user(role);
    answerWith(signedIn);
    const router = createTestRouter('/admin', signedIn);

    await router.load();
    renderRouter(router);

    expect(router.state.location.href).toBe('/admin');
    expect(screen.getByRole('heading', { name: 'Курси' })).toBeTruthy();
    expect(screen.getByRole('button', { name: 'Створити курс' })).toBeTruthy();
    expect(screen.queryByText('Сторінку не знайдено')).toBeNull();
  },
);

test('P16 renders an owner-scoped course row through the guarded memory route', async () => {
  const instructor = user('instructor');
  answerWith(instructor);
  const course: AdminCourse = {
    id: asBrandedId('10000000-0000-4000-8000-000000000001', 'course'),
    slug: 'owned-course',
    title: 'Власний курс',
    status: 'published',
    instructorId: instructor.id,
    updatedAt: '2026-09-05T10:15:30.000000Z',
    lessonCount: 6,
    studentCount: 4,
    price: null,
  };
  const router = createTestRouter('/admin?status=published', instructor, 'uk', [course]);

  await router.load();
  renderRouter(router);

  const table = screen.getByRole('table', { name: 'Курси' });
  expect(table).toBeTruthy();
  expect(
    within(table)
      .getByRole('link', { name: /^Власний курс/ })
      .getAttribute('href'),
  ).toBe('/admin/courses/10000000-0000-4000-8000-000000000001');
  expect(within(table).getByText('Опубліковано')).toBeTruthy();
  expect(observedAdminQueries).toEqual([
    { status: 'published', limit: 20, offset: 0, search: '', sort: 'recent' },
  ]);
});

test('B02 renders the admin list empty state without weakening the shell guard', async () => {
  const admin = user('admin');
  answerWith(admin);
  const router = createTestRouter('/admin', admin);

  await router.load();
  renderRouter(router);

  expect(screen.getByRole('heading', { name: 'Курси' })).toBeTruthy();
  expect(screen.getByRole('heading', { name: 'Курсів ще немає' })).toBeTruthy();
  expect(screen.queryByRole('table')).toBeNull();
  // The empty slot replaces the table and its pager (proto's editorial empty state).
  expect(screen.queryByRole('navigation', { name: 'Сторінки курсів' })).toBeNull();
  expect(observedAdminQueries).toEqual([{ limit: 20, offset: 0, search: '', sort: 'recent' }]);
});

test('R34 keeps unusable admin search text in the URL and queries with independent defaults', async () => {
  const admin = user('admin');
  answerWith(admin);
  const href = '/admin?status=nonsense&page=abc&x=1';
  const router = createTestRouter(href, admin);

  await router.load();
  renderRouter(router);

  expect(screen.getByRole('heading', { name: 'Курси' })).toBeTruthy();
  expect(screen.queryByText('Не вдалося відкрити сторінку')).toBeNull();
  expect(router.state.location.href).toBe(href);
  expect(observedAdminQueries).toEqual([{ limit: 20, offset: 0, search: '', sort: 'recent' }]);
});

test('R21 keeps unusable user search text in the URL and loads default directory filters', async () => {
  const admin = user('admin');
  answerWith(admin);
  const href = '/admin/users?status=nonsense&page=abc';
  const router = createTestRouter(href, admin);

  await router.load();
  renderRouter(router);

  expect(screen.getByRole('heading', { name: 'Користувачі' })).toBeTruthy();
  expect(screen.queryByText('Не вдалося відкрити сторінку')).toBeNull();
  expect(router.state.location.href).toBe(href);
  expect(observedAdminUserQueries).toEqual([{ q: '', limit: 20, offset: 0, sort: 'name' }]);
});

test('R24 gives an instructor not-found at /admin/users before loading the directory', async () => {
  const instructor = user('instructor');
  answerWith(instructor);
  const router = createTestRouter('/admin/users', instructor);

  await router.load();
  renderRouter(router);

  expect(screen.getByRole('heading', { name: 'Сторінку не знайдено' })).toBeTruthy();
  expect(screen.queryByRole('heading', { name: 'Користувачі' })).toBeNull();
  expect(observedAdminUserQueries).toEqual([]);
});

test.each([
  { role: null, expected: '/login?returnTo=%2Fadmin%3Fstatus%3Dnonsense%26page%3Dabc' },
  { role: 'student', expected: '/admin?status=nonsense&page=abc' },
] satisfies ReadonlyArray<{ role: Role | null; expected: string }>)(
  'R34 keeps the shell refusal for the $role role with unusable search',
  async ({ role, expected }) => {
    const signedIn = role === null ? null : user(role);
    answerWith(signedIn);
    const router = createTestRouter('/admin?status=nonsense&page=abc', signedIn);

    await router.load();
    if (role !== null) {
      renderRouter(router);
      expect(screen.getByRole('heading', { name: 'Сторінку не знайдено' })).toBeTruthy();
    }

    expect(router.state.location.href).toBe(expected);
    expect(observedAdminQueries).toEqual([]);
  },
);

test.each([
  ['/learn/go/intro?lesson=3', null, '/login?returnTo=%2Flearn%2Fgo%2Fintro%3Flesson%3D3'],
  ['/admin/courses', null, '/login?returnTo=%2Fadmin%2Fcourses'],
] as const)('G06 redirects anonymous nested address %s', async (href, cached, expected) => {
  answerWith(null);
  const router = createTestRouter(href, cached);

  await router.load();

  expect(router.state.location.href).toBe(expected);
});

test('G06 applies the parent role matrix to signed-in nested addresses', async () => {
  const student = user('student');
  answerWith(student);
  const learnRouter = createTestRouter('/learn/go/intro', student);
  await learnRouter.load();
  const learnScreen = renderRouter(learnRouter);
  expect(screen.getByRole('heading', { name: 'Ваше навчання' })).toBeTruthy();
  learnScreen.unmount();

  const adminRouter = createTestRouter('/admin/courses', student);
  await adminRouter.load();
  renderRouter(adminRouter);
  expect(screen.getByRole('heading', { name: 'Сторінку не знайдено' })).toBeTruthy();
});

test('R11 renders course access denial instead of a session failure', async () => {
  const student = user('student');
  answerWith(student);
  const router = createTestRouter(
    '/learn/foreign-draft',
    student,
    'uk',
    [],
    {},
    [],
    {},
    {
      getCourse: async () => ({ ok: false, error: { code: 'course_forbidden' } }),
    },
  );

  await router.load();
  renderRouter(router);

  expect(screen.getByRole('heading', { name: 'Немає доступу до курсу' })).toBeTruthy();
  expect(screen.getByText('Цей курс недоступний для вашого облікового запису.')).toBeTruthy();
  expect(screen.getByRole('link', { name: 'Перейти до каталогу' }).getAttribute('href')).toBe(
    '/courses',
  );
  expect(screen.queryByRole('link', { name: 'Спробувати ще раз' })).toBeNull();
  expect(screen.queryByRole('heading', { name: 'Не вдалося перевірити сеанс' })).toBeNull();
});

test('K18 keeps the lesson usable when discussion loading fails', async () => {
  const student = user('student');
  const lessonId = '30000000-0000-4000-8000-000000000001';
  const course: LearningCourse = {
    id: '10000000-0000-4000-8000-000000000001',
    slug: 'go',
    title: 'Go basics',
    summary: 'Practice',
    description: 'Learn Go.',
    status: 'published',
    price: null,
    continueLessonId: lessonId,
    modules: [
      {
        id: '20000000-0000-4000-8000-000000000001',
        title: 'Basics',
        lessons: [
          {
            id: lessonId,
            title: 'Ownership',
            videoState: 'ready',
            position: 0,
            completed: false,
            clientSeq: 0,
            isPreview: false,
            locked: false,
          },
        ],
      },
    ],
  };
  const transcript: LessonTranscript = {
    lessonId,
    language: 'en',
    updatedAt: '2026-09-28T10:00:00Z',
    stale: false,
    segments: [
      {
        id: '70000000-0000-4000-8000-000000000001',
        startMs: 0,
        endMs: 1_000,
        text: 'The first line',
        documentIds: [],
      },
    ],
    documents: [],
  };
  const transcriptRepository: TranscriptRepository = {
    getLessonTranscript: async () => ({ ok: true, value: transcript }),
    saveLessonTranscript: async () => ({ ok: false, error: { code: 'dependency_unavailable' } }),
    listCourseDocuments: async () => ({ ok: true, value: [] }),
    addCourseDocument: async () => ({ ok: false, error: { code: 'dependency_unavailable' } }),
    removeCourseDocument: async () => ({ ok: false, error: { code: 'dependency_unavailable' } }),
  };
  const discussionRepository: DiscussionRepository = {
    list: async () => ({ ok: false, error: { code: 'dependency_unavailable' } }),
    publish: async () => ({ ok: false, error: { code: 'dependency_unavailable' } }),
    update: async () => ({ ok: false, error: { code: 'dependency_unavailable' } }),
    remove: async () => ({ ok: false, error: { code: 'dependency_unavailable' } }),
  };
  answerWith(student);
  const router = createTestRouter(
    `/learn/${course.slug}/${lessonId}`,
    student,
    'en',
    [],
    {},
    [],
    {},
    { getCourse: async () => ({ ok: true, value: course }) },
    transcriptRepository,
    discussionRepository,
  );

  await router.load();
  renderRouter(router);

  expect(await screen.findByRole('heading', { name: 'Ownership' })).toBeTruthy();
  expect(await screen.findByRole('heading', { name: 'Transcript' })).toBeTruthy();
  const panel = screen.getByRole('region', { name: 'Discussion' });
  expect(await within(panel).findByText('The service is unavailable. Try again.')).toBeTruthy();
  expect(screen.queryByRole('heading', { name: 'Could not verify your session' })).toBeNull();
});

test('G08 shows the route pending state while protected navigation revalidates', async () => {
  const admin = user('admin');
  let release: (() => void) | undefined;
  let requestStarted = false;
  const responseGate = new Promise<void>((resolve) => {
    release = resolve;
  });
  server.use(
    http.get(currentUserUrl, async () => {
      requestStarted = true;
      await responseGate;
      return HttpResponse.json(admin);
    }),
  );
  const router = createTestRouter('/', admin);
  await router.load();
  renderRouter(router);

  const navigation = router.navigate({ to: '/admin' });
  await waitFor(() => expect(requestStarted).toBe(true));
  expect(await screen.findByText('Перевіряємо сеанс')).toBeTruthy();
  if (release === undefined) throw new Error('The delayed identity response was not initialized');
  release();
  await navigation;

  expect(await screen.findByRole('heading', { name: 'Курси' })).toBeTruthy();
});

test('G09 renders a localized retryable route error when identity is unavailable', async () => {
  const admin = user('admin');
  server.use(
    http.get(currentUserUrl, () => HttpResponse.json(dependencyUnavailable, { status: 503 })),
  );
  const router = createTestRouter('/admin', admin);

  await router.load();
  renderRouter(router);

  expect(screen.getByRole('heading', { name: 'Не вдалося перевірити сеанс' })).toBeTruthy();
  expect(screen.getByRole('link', { name: 'Спробувати ще раз' }).getAttribute('href')).toBe(
    '/admin',
  );
  expect(router.state.location.href).toBe('/admin');

  answerWith(admin);
  await router.invalidate({ sync: true });

  expect(await screen.findByRole('heading', { name: 'Курси' })).toBeTruthy();
});

test('G11 revalidates a cached identity and redirects when the session has ended', async () => {
  const admin = user('admin');
  answerWith(null);
  const router = createTestRouter('/admin', admin);

  await router.load();

  expect(router.state.location.href).toBe('/login?returnTo=%2Fadmin');
  expect(router.options.context.queryClient.getQueryData(sessionQueryKey)).toBe(null);
});

test('G12 leaves a mounted protected route when focus revalidation ends the session', async () => {
  const admin = user('admin');
  let current: AuthenticatedUser | null = admin;
  server.use(
    http.get(currentUserUrl, () =>
      current === null
        ? HttpResponse.json(authenticationRequired, { status: 401 })
        : HttpResponse.json(current),
    ),
  );
  const router = createTestRouter('/admin', admin);
  await router.load();
  renderRouter(router);
  expect(screen.getByRole('heading', { name: 'Курси' })).toBeTruthy();

  current = null;
  focusManager.setFocused(false);
  focusManager.setFocused(true);

  await waitFor(() => expect(router.state.location.href).toBe('/login?returnTo=%2Fadmin'));
});

test('G24 signs out, clears the cached identity and navigates to the catalog', async () => {
  const admin = user('admin');
  let current: AuthenticatedUser | null = admin;
  let currentUserRequests = 0;
  let logoutRequests = 0;
  server.use(
    http.get(currentUserUrl, () => {
      currentUserRequests += 1;
      return current === null
        ? HttpResponse.json(authenticationRequired, { status: 401 })
        : HttpResponse.json(current);
    }),
    http.post(logoutUrl, () => {
      logoutRequests += 1;
      current = null;
      return new HttpResponse(null, { status: 204 });
    }),
  );
  const router = createTestRouter('/learn', admin);
  await router.load();
  renderRouter(router);
  expect(currentUserRequests).toBe(1);

  await userEvent.click(screen.getByRole('button', { name: 'Вийти' }));

  await waitFor(() => expect(router.state.location.href).toBe('/courses'));
  await waitFor(() => expect(router.state.status).toBe('idle'));
  expect(logoutRequests).toBe(1);
  expect(router.options.context.queryClient.getQueryData(sessionQueryKey)).toBe(null);
  expect(screen.getByRole('link', { name: 'Увійти' })).toBeTruthy();
  expect(screen.queryByText('admin')).toBeNull();

  await router.navigate({ to: '/learn' });
  expect(router.state.location.href).toBe('/login?returnTo=%2Flearn');
  expect(currentUserRequests).toBe(2);
});

test('G25 ignores a second sign-out while the first request is pending', async () => {
  const admin = user('admin');
  answerWith(null);
  const browser = userEvent.setup();
  let logoutRequests = 0;
  let release: (() => void) | undefined;
  const responseGate = new Promise<void>((resolve) => {
    release = resolve;
  });
  server.use(
    http.post(logoutUrl, async () => {
      logoutRequests += 1;
      await responseGate;
      return new HttpResponse(null, { status: 204 });
    }),
  );
  const router = createTestRouter('/', admin);
  await router.load();
  renderRouter(router);

  await browser.click(screen.getByRole('button', { name: 'Вийти' }));
  const pendingButton = await screen.findByRole('button', { name: 'Виходимо…' });
  expect(pendingButton.hasAttribute('disabled')).toBe(true);
  await browser.click(pendingButton);
  expect(logoutRequests).toBe(1);

  if (release === undefined) throw new Error('The delayed logout response was not initialized');
  release();
  await waitFor(() => expect(router.state.location.href).toBe('/courses'));
  await waitFor(() => expect(router.state.status).toBe('idle'));
});

test.each([
  {
    locale: 'uk',
    signOut: 'Вийти',
    login: 'Увійти',
    message: 'Сеанс у браузері завершено, але сервер не підтвердив вихід.',
  },
  {
    locale: 'en',
    signOut: 'Sign out',
    login: 'Log in',
    message: 'Your browser session ended, but the server did not confirm sign-out.',
  },
] satisfies ReadonlyArray<{ locale: Locale; signOut: string; login: string; message: string }>)(
  'G26/G31 ends the $locale local session and localizes the revocation failure',
  async ({ locale, signOut, login, message }) => {
    const admin = user('admin');
    answerWith(null);
    server.use(
      http.post(logoutUrl, () => HttpResponse.json(dependencyUnavailable, { status: 503 })),
    );
    const router = createTestRouter('/', admin, locale);
    await router.load();
    renderRouter(router);

    await userEvent.click(screen.getByRole('button', { name: signOut }));

    await waitFor(() => expect(router.state.location.href).toBe('/courses'));
    await waitFor(() => expect(router.state.status).toBe('idle'));
    expect(router.options.context.queryClient.getQueryData(sessionQueryKey)).toBe(null);
    expect(screen.getByRole('status').textContent).toContain(message);
    expect(screen.getByRole('link', { name: login })).toBeTruthy();
  },
);

function paidCourse(): LearningCourse {
  const lesson = (id: string, title: string, isPreview: boolean, locked: boolean) => ({
    id,
    title,
    videoState: 'ready' as const,
    position: 0,
    completed: false,
    clientSeq: 0,
    isPreview,
    locked,
  });
  return {
    id: '10000000-0000-4000-8000-000000000009',
    slug: 'paid-go',
    title: 'Paid Go',
    summary: 'Practice',
    description: 'Learn Go.',
    status: 'published',
    price: { minor: 189000, currency: 'UAH' },
    continueLessonId: null,
    modules: [
      {
        id: '20000000-0000-4000-8000-000000000009',
        title: 'Basics',
        lessons: [
          lesson('30000000-0000-4000-8000-000000000091', 'Locked first', false, true),
          lesson('30000000-0000-4000-8000-000000000092', 'Open preview', true, false),
        ],
      },
    ],
  };
}

const pendingOrder: PurchaseOrder = {
  id: '60000000-0000-4000-8000-000000000009',
  courseId: '10000000-0000-4000-8000-000000000009',
  amount: { minor: 189000, currency: 'UAH' },
  status: 'pending',
  createdAt: '2026-10-01T10:00:00Z',
  paidAt: null,
};

function recordingPurchase(
  keys: string[],
  orders: readonly PurchaseOrder[],
  status: () => PurchaseOrder,
): PurchaseRepository {
  let next = 0;
  return {
    start: async (_courseId, key) => {
      keys.push(key);
      const order = orders[next++] ?? orders[orders.length - 1];
      return order === undefined
        ? { ok: false, error: { code: 'dependency_unavailable' } }
        : { ok: true, value: order };
    },
    status: async () => ({ ok: true, value: status() }),
    checkout: async () => ({ ok: false, error: { code: 'dependency_unavailable' } }),
  };
}

test('Pay11 a locked lesson offers the purchase and requests no lesson content', async () => {
  const student = user('student');
  const course = paidCourse();
  const lockedId = '30000000-0000-4000-8000-000000000091';
  const contentRequests: string[] = [];
  const keys: string[] = [];
  answerWith(student);
  const router = createTestRouter(
    `/learn/${course.slug}/${lockedId}`,
    student,
    'uk',
    [],
    {},
    [],
    {},
    {
      getCourse: async () => ({ ok: true, value: course }),
      play: async () => {
        contentRequests.push('play');
        return { ok: false, error: { code: 'playback_forbidden' } };
      },
    },
    new Proxy({} as TranscriptRepository, {
      get: () => () => {
        contentRequests.push('transcript');
        return Promise.resolve({ ok: false, error: { code: 'dependency_unavailable' } });
      },
    }),
    new Proxy({} as DiscussionRepository, {
      get: () => () => {
        contentRequests.push('discussion');
        return Promise.resolve({ ok: false, error: { code: 'dependency_unavailable' } });
      },
    }),
    undefined,
    recordingPurchase(keys, [pendingOrder], () => pendingOrder),
  );

  await router.load();
  renderRouter(router);

  expect(await screen.findByRole('heading', { level: 1, name: 'Locked first' })).toBeTruthy();
  expect(screen.getByText(/1\s?890/)).toBeTruthy();
  const outline = screen.getByRole('navigation', { name: 'Зміст курсу' });
  expect(within(outline).getByRole('link', { name: /Locked first.*Після покупки/ })).toBeTruthy();
  expect(within(outline).getByRole('link', { name: /Open preview.*Відкритий урок/ })).toBeTruthy();

  await userEvent.click(screen.getByRole('button', { name: 'Купити курс' }));

  await waitFor(() => expect(router.state.location.pathname).toBe(`/orders/${pendingOrder.id}`));
  expect(router.state.location.search).toEqual({ course: course.slug });
  expect(
    await screen.findByText('Чекаємо підтвердження оплати. Сторінка оновиться сама.'),
  ).toBeTruthy();
  expect(keys).toHaveLength(1);
  expect(contentRequests).toEqual([]);
});

test('O05 a failed order retries under a new idempotency key (assumptions, O05 prerequisites)', async () => {
  const student = user('student');
  const failed: PurchaseOrder = { ...pendingOrder, status: 'failed' };
  const second: PurchaseOrder = { ...pendingOrder, id: '60000000-0000-4000-8000-00000000000a' };
  const keys: string[] = [];
  answerWith(student);
  const router = createTestRouter(
    `/orders/${failed.id}?course=paid-go`,
    student,
    'uk',
    [],
    {},
    [],
    {},
    {},
    transcriptStub,
    discussionStub,
    undefined,
    recordingPurchase(keys, [second], () => (keys.length === 0 ? failed : second)),
  );

  await router.load();
  renderRouter(router);

  expect(await screen.findByText('Оплата не пройшла. Можна спробувати ще раз.')).toBeTruthy();
  await userEvent.click(screen.getByRole('button', { name: 'Спробувати ще раз' }));

  await waitFor(() => expect(router.state.location.pathname).toBe(`/orders/${second.id}`));
  expect(keys).toHaveLength(1);
  expect(keys[0]).toMatch(/^[0-9a-f-]{36}$/);
});

test('a paid order survives the checkout-409 then already-owned retry race', async () => {
  const student = user('student');
  const keys: string[] = [];
  let statusCalls = 0;
  answerWith(student);
  const router = createTestRouter(
    `/orders/${pendingOrder.id}?course=paid-go`,
    student,
    'uk',
    [],
    {},
    [],
    {},
    {},
    transcriptStub,
    discussionStub,
    undefined,
    {
      start: async (_courseId, key) => {
        keys.push(key);
        return { ok: false, error: { code: 'course_already_owned' as const } };
      },
      status: async () => {
        statusCalls += 1;
        return statusCalls === 1
          ? { ok: true, value: pendingOrder }
          : {
              ok: true,
              value: { ...pendingOrder, status: 'paid', paidAt: '2026-10-01T10:01:00Z' },
            };
      },
      checkout: async () => ({ ok: false, error: { code: 'checkout_already_opened' as const } }),
    },
  );

  await router.load();
  renderRouter(router);

  await userEvent.click(await screen.findByRole('button', { name: 'Перейти до оплати' }));
  expect(
    await screen.findByText('Не вдалося зв’язатися з сервером. Спробуйте ще раз.'),
  ).toBeTruthy();
  expect(await screen.findByText('Оплату зараховано — курс відкрито.')).toBeTruthy();
  expect(screen.getByRole('link', { name: 'Перейти до курсу' })).toBeTruthy();
  expect(router.state.location.pathname).toBe(`/orders/${pendingOrder.id}`);
  expect(keys).toHaveLength(1);
});

test('Pay13 a paid order re-reads the course outline so its lessons unlock', async () => {
  const student = user('student');
  const paid: PurchaseOrder = { ...pendingOrder, status: 'paid', paidAt: '2026-10-01T10:01:00Z' };
  answerWith(student);
  const router = createTestRouter(
    `/orders/${paid.id}?course=paid-go`,
    student,
    'uk',
    [],
    {},
    [],
    {},
    {
      getCourse: async () => ({ ok: true, value: paidCourse() }),
    },
    transcriptStub,
    discussionStub,
    undefined,
    recordingPurchase([], [], () => paid),
  );
  await router.load();
  // The outline was cached before the purchase, as it is after visiting a locked lesson.
  const { queryClient, learn } = router.options.context;
  await queryClient.prefetchQuery(learn.course('paid-go'));
  renderRouter(router);

  expect(await screen.findByText('Оплату зараховано — курс відкрито.')).toBeTruthy();
  expect(screen.getByRole('link', { name: 'Перейти до курсу' }).getAttribute('href')).toBe(
    '/learn/paid-go',
  );
  // Not mounted here, so it is marked stale and re-read when the course opens again.
  await waitFor(() =>
    expect(queryClient.getQueryState(learningCourseQueryKey('paid-go'))?.isInvalidated).toBe(true),
  );
});

test('an unsafe course slug on the order page is dropped from the way back', async () => {
  const student = user('student');
  answerWith(student);
  const router = createTestRouter(
    `/orders/${pendingOrder.id}?course=..%2Fadmin`,
    student,
    'uk',
    [],
    {},
    [],
    {},
    {},
    transcriptStub,
    discussionStub,
    undefined,
    recordingPurchase([], [], () => pendingOrder),
  );
  await router.load();
  renderRouter(router);
  expect(await screen.findByRole('link', { name: 'Повернутися до курсу' })).toBeTruthy();
  expect(screen.getByRole('link', { name: 'Повернутися до курсу' }).getAttribute('href')).toBe(
    '/courses',
  );
});
