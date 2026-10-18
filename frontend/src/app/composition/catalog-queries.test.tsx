import { describe, expect, it } from '@rstest/core';
import { QueryClient, QueryClientProvider, useQuery } from '@tanstack/react-query';
import {
  createMemoryHistory,
  createRootRoute,
  createRoute,
  createRouter,
} from '@tanstack/react-router';
import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { createCatalog } from '../../application/catalog/catalog';
import type { CatalogQuery, CoursePage, Result } from '../../contracts/course';
import type { CourseRepository } from '../../ports/CourseRepository';
import { createCatalogQueries, ensureRouteQuery } from './catalog-queries';

const baseQuery: CatalogQuery = { q: 'Go', sort: 'recent', limit: 20, offset: 0 };

function coursePage(title: string): CoursePage {
  return {
    items: [
      {
        id: title,
        slug: 'go-architecture',
        title,
        summary: 'Практичний курс',
        description: 'Програма курсу',
        publishedAt: '2026-09-01T00:00:00Z',
        price: null,
        paid: false,
      },
    ],
    pagination: { limit: 20, offset: 0, total: 1 },
  };
}

function clientForTest() {
  return new QueryClient({
    defaultOptions: { queries: { retry: false, staleTime: Infinity, gcTime: Infinity } },
  });
}

describe('catalog query boundary', () => {
  it('cancels an actual pending route loader when navigation moves to another route', async () => {
    const client = clientForTest();
    const delayed = delayedCatalog();
    const root = createRootRoute();
    const courses = createRoute({
      getParentRoute: () => root,
      path: '/courses',
      loader: ({ abortController }) =>
        ensureRouteQuery(client, delayed.queries.list(baseQuery), abortController.signal),
    });
    const login = createRoute({ getParentRoute: () => root, path: '/login' });
    const router = createRouter({
      routeTree: root.addChildren([courses, login]),
      history: createMemoryHistory({ initialEntries: ['/courses'] }),
    });
    try {
      const loading = router.load();
      await waitFor(() => expect(delayed.pending.length).toBe(1));
      const request = delayed.pending[0];
      if (!request?.signal) throw new Error('The loader must forward a cancellation signal');
      await router.navigate({ to: '/login' });
      await waitFor(() => expect(request.signal?.aborted).toBe(true));
      await loading;
      expect(router.state.location.pathname).toBe('/login');
    } finally {
      client.clear();
    }
  });

  it('keeps a shared read alive when only its preload consumer is canceled', async () => {
    const client = clientForTest();
    const delayed = delayedCatalog();
    const preload = new AbortController();
    const navigation = new AbortController();
    try {
      const preloadResult = ensureRouteQuery(
        client,
        delayed.queries.list(baseQuery),
        preload.signal,
      );
      const navigationResult = ensureRouteQuery(
        client,
        delayed.queries.list(baseQuery),
        navigation.signal,
      );
      const canceled = expect(preloadResult).rejects.toMatchObject({ name: 'AbortError' });
      preload.abort();
      await canceled;
      expect(delayed.pending.length).toBe(1);
      const request = delayed.pending[0];
      if (!request?.signal) throw new Error('The shared read must provide its signal');
      expect(request.signal.aborted).toBe(false);
      request.resolve({ ok: true, value: coursePage('Навігація завершена') });
      expect((await navigationResult).items[0]?.title).toBe('Навігація завершена');
    } finally {
      client.clear();
    }
  });

  it('starts a fresh read after a canceled preload and does not reuse its aborted signal', async () => {
    const client = clientForTest();
    const delayed = delayedCatalog();
    const preload = new AbortController();
    try {
      const preloadResult = ensureRouteQuery(
        client,
        delayed.queries.list(baseQuery),
        preload.signal,
      );
      const canceled = expect(preloadResult).rejects.toMatchObject({ name: 'AbortError' });
      const original = delayed.pending[0];
      if (!original?.signal) throw new Error('The preload must provide its signal');
      preload.abort();
      await canceled;
      expect(original.signal.aborted).toBe(true);

      const fresh = ensureRouteQuery(
        client,
        delayed.queries.list(baseQuery),
        new AbortController().signal,
      );
      expect(delayed.pending.length).toBe(2);
      const replacement = delayed.pending[1];
      if (!replacement?.signal) throw new Error('Navigation must create a fresh read');
      expect(replacement.signal.aborted).toBe(false);
      replacement.resolve({ ok: true, value: coursePage('Свіжий результат') });
      expect((await fresh).items[0]?.title).toBe('Свіжий результат');
      original.resolve({ ok: true, value: coursePage('Запізнілий preload') });
      expect((await client.fetchQuery(delayed.queries.list(baseQuery))).items[0]?.title).toBe(
        'Свіжий результат',
      );
    } finally {
      client.clear();
    }
  });

  it('aborts obsolete A and keeps B visible when A completes later', async () => {
    type PendingRead = {
      query: CatalogQuery;
      signal: AbortSignal | undefined;
      resolve: (result: Result<CoursePage>) => void;
    };
    const pending: PendingRead[] = [];
    const repository: CourseRepository = {
      listPublished: (query, signal) =>
        new Promise((resolve) => {
          pending.push({ query, signal, resolve });
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
    const queries = createCatalogQueries(createCatalog(repository));
    const client = clientForTest();
    const wrapper = ({ children }: { children: ReactNode }) => (
      <QueryClientProvider client={client}>{children}</QueryClientProvider>
    );
    const hook = renderHook(({ query }: { query: CatalogQuery }) => useQuery(queries.list(query)), {
      initialProps: { query: baseQuery },
      wrapper,
    });

    try {
      await waitFor(() => expect(pending.length).toBe(1));
      const a = pending[0];
      if (!a?.signal) throw new Error('Query cancellation signal must reach the repository port');
      expect(hook.result.current.isPending).toBe(true);
      hook.rerender({ query: { q: 'TypeScript', sort: 'title', limit: 10, offset: 0 } });
      await waitFor(() => expect(pending.length).toBe(2));
      const b = pending[1];
      if (!b) throw new Error('The new filter must request its own result');
      expect(a.signal.aborted).toBe(true);

      await act(async () => {
        b.resolve({ ok: true, value: coursePage('TypeScript — активний результат') });
      });
      await waitFor(() =>
        expect(hook.result.current.data?.items[0]?.title).toBe('TypeScript — активний результат'),
      );

      // The external dependency intentionally ignores cancellation and finishes late.
      // Query must still protect the current screen from this obsolete response.
      await act(async () => {
        a.resolve({ ok: true, value: coursePage('Go — застарілий результат') });
      });
      expect(hook.result.current.data?.items[0]?.title).toBe('TypeScript — активний результат');
      expect(hook.result.current.isSuccess).toBe(true);
    } finally {
      hook.unmount();
      client.clear();
    }
  });

  const alternatives: { name: string; query: CatalogQuery }[] = [
    { name: 'search', query: { ...baseQuery, q: 'TypeScript' } },
    { name: 'sort', query: { ...baseQuery, sort: 'title' } },
    { name: 'limit', query: { ...baseQuery, limit: 10 } },
    { name: 'offset', query: { ...baseQuery, offset: 20 } },
  ];

  it.each(alternatives)('does not reuse cached results when $name changes', async ({ query }) => {
    let reads = 0;
    const repository: CourseRepository = {
      listPublished: async () => ({
        ok: true,
        value: coursePage(++reads === 1 ? 'Перший набір' : 'Новий набір'),
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
    const queries = createCatalogQueries(createCatalog(repository));
    const client = clientForTest();
    try {
      const original = await client.fetchQuery(queries.list(baseQuery));
      const changed = await client.fetchQuery(queries.list(query));
      const originalAgain = await client.fetchQuery(queries.list(baseQuery));

      expect(original.items[0]?.title).toBe('Перший набір');
      expect(changed.items[0]?.title).toBe('Новий набір');
      expect(originalAgain.items[0]?.title).toBe('Перший набір');
      expect(reads).toBe(2);
    } finally {
      client.clear();
    }
  });

  it('keeps two request-scoped QueryClient caches independent for an identical URL', async () => {
    function queriesFor(title: string) {
      const repository: CourseRepository = {
        listPublished: async () => ({ ok: true, value: coursePage(title) }),
        getPublished: async () => ({ ok: false, error: { code: 'course_not_found' } }),
        listReviews: async () => ({
          ok: true,
          value: { rating: null, items: [], nextCursor: null },
        }),
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
      return createCatalogQueries(createCatalog(repository));
    }
    const clientA = clientForTest();
    const clientB = clientForTest();
    const queriesA = queriesFor('Результат запиту A');
    const queriesB = queriesFor('Результат запиту B');
    try {
      const [a, b] = await Promise.all([
        clientA.fetchQuery(queriesA.list(baseQuery)),
        clientB.fetchQuery(queriesB.list(baseQuery)),
      ]);
      expect(a.items[0]?.title).toBe('Результат запиту A');
      expect(b.items[0]?.title).toBe('Результат запиту B');

      clientA.clear();
      expect((await clientB.fetchQuery(queriesB.list(baseQuery))).items[0]?.title).toBe(
        'Результат запиту B',
      );
    } finally {
      clientA.clear();
      clientB.clear();
    }
  });

  it('exposes a typed domain failure as Query error without caching it as successful data', async () => {
    const repository: CourseRepository = {
      listPublished: async () => ({ ok: false, error: { code: 'dependency_unavailable' } }),
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
    const queries = createCatalogQueries(createCatalog(repository));
    const client = clientForTest();
    try {
      await expect(client.fetchQuery(queries.list(baseQuery))).rejects.toMatchObject({
        code: 'dependency_unavailable',
      });
      expect(client.getQueryState(queries.list(baseQuery).queryKey)?.status).toBe('error');
      expect(client.getQueryData(queries.list(baseQuery).queryKey)).toBeUndefined();
      await expect(client.fetchQuery(queries.detail('missing-course'))).rejects.toMatchObject({
        code: 'course_not_found',
      });
    } finally {
      client.clear();
    }
  });
});

function delayedCatalog() {
  const pending: {
    signal: AbortSignal | undefined;
    resolve: (result: Result<CoursePage>) => void;
  }[] = [];
  const repository: CourseRepository = {
    listPublished: (_query, signal) =>
      new Promise((resolve) => {
        pending.push({ signal, resolve });
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
  return { queries: createCatalogQueries(createCatalog(repository)), pending };
}
