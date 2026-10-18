import {
  QueryObserver,
  infiniteQueryOptions,
  mutationOptions,
  queryOptions,
} from '@tanstack/react-query';
import type { FetchQueryOptions, QueryClient, QueryKey } from '@tanstack/react-query';
import type { createCatalog } from '../../modules/catalog';
import type { CatalogError, CatalogQuery, Result } from '../../contracts/course';

export class CatalogQueryError extends Error {
  readonly code: CatalogError['code'];
  constructor(error: CatalogError) {
    super(error.code);
    this.name = 'CatalogQueryError';
    this.code = error.code;
  }
}
function unwrap<T>(result: Result<T>): T {
  if (!result.ok) throw new CatalogQueryError(result.error);
  return result.value;
}

/**
 * A pending route is a Query consumer before its React component mounts.
 * An observer gives each loader its own lifetime: canceling one preload does
 * not cancel another consumer, while removing the last one aborts the read.
 * Keep loader signals out of cached query functions so later loads get fresh signals.
 */
export function ensureRouteQuery<T, TKey extends QueryKey>(
  client: QueryClient,
  options: FetchQueryOptions<T, Error, T, TKey>,
  signal: AbortSignal,
): Promise<T> {
  if (signal.aborted) return Promise.reject(signal.reason);
  const observer = new QueryObserver(client, { ...options, enabled: false });
  const unsubscribe = observer.subscribe(() => {});
  return new Promise<T>((resolve, reject) => {
    let settled = false;
    const finish = (complete: () => void) => {
      if (settled) return;
      settled = true;
      signal.removeEventListener('abort', abort);
      unsubscribe();
      complete();
    };
    const abort = () => {
      const reason: unknown = signal.reason;
      finish(() => reject(reason));
    };
    signal.addEventListener('abort', abort, { once: true });
    void client.ensureQueryData(options).then(
      (data) => finish(() => resolve(data)),
      (error: unknown) => finish(() => reject(error)),
    );
  });
}

export function createCatalogQueries(catalog: ReturnType<typeof createCatalog>) {
  return {
    list: (query: CatalogQuery) =>
      queryOptions({
        queryKey: ['catalog', query.q, query.sort, query.limit, query.offset],
        queryFn: async ({ signal }) => unwrap(await catalog.list(query, signal)),
      }),
    detail: (slug: string) =>
      queryOptions({
        queryKey: ['course', slug],
        queryFn: async ({ signal }) => unwrap(await catalog.get(slug, signal)),
      }),
    reviews: (slug: string) =>
      infiniteQueryOptions({
        queryKey: ['course', slug, 'reviews'],
        initialPageParam: undefined as string | undefined,
        queryFn: async ({ signal, pageParam }) =>
          unwrap(await catalog.reviews(slug, pageParam, signal)),
        getNextPageParam: (lastPage) => lastPage.nextCursor ?? undefined,
      }),
    ownReview: (courseId: string, userId: string) =>
      queryOptions({
        queryKey: ['course', courseId, 'own-review', userId],
        queryFn: async ({ signal }) => unwrap(await catalog.ownReview(courseId, signal)),
      }),
    report: () =>
      mutationOptions({
        mutationFn: async ({ reviewId }: { reviewId: string }) =>
          unwrap(await catalog.report(reviewId)),
      }),
  };
}
