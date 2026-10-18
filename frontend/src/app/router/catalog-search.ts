import { notFound } from '@tanstack/react-router';
import type { SearchSchemaInput } from '@tanstack/react-router';
import type { CatalogQuery } from '../../contracts/course';
import { z } from 'zod';
import { makeCatalogQuery } from '../../contracts/course';

const searchSchema = z.object({
  q: z.string().optional(),
  sort: z.enum(['recent', 'title']).optional(),
  limit: z
    .union([
      z.number(),
      z
        .string()
        .regex(/^[0-9]+$/)
        .transform(Number),
    ])
    .optional(),
  offset: z
    .union([
      z.number(),
      z
        .string()
        .regex(/^[0-9]+$/)
        .transform(Number),
    ])
    .optional(),
});

/**
 * The address as it arrived, still untrusted. Values stay `unknown` because
 * this type exists to keep the URL intact, not to describe a usable query:
 * {@link catalogQuery} is what narrows it.
 */
export type CatalogSearch = {
  readonly q?: unknown;
  readonly sort?: unknown;
  readonly limit?: unknown;
  readonly offset?: unknown;
};

export function parseCatalogSearch(
  search: SearchSchemaInput & Partial<CatalogSearch>,
): CatalogSearch {
  return validateCatalogSearch(search);
}

/**
 * Echoes the four known parameters verbatim and drops everything else. It
 * neither defaults nor refuses, and that is deliberate on both counts.
 *
 * The router rebuilds the canonical URL from this output and redirects when it
 * differs from the request, so emitting a default would answer every plain
 * `/courses` with a 307 to `/courses?q=&sort=recent&limit=20&offset=0`. And a
 * value thrown from here never reaches the route's own handling: the router
 * answers it with a hardcoded 500, which claims a server fault for what is a
 * malformed address. Refusal happens in the route, through {@link catalogQuery}.
 */
export function validateCatalogSearch(search: unknown): CatalogSearch {
  if (typeof search !== 'object' || search === null) return {};
  const source = search as Record<string, unknown>;
  return {
    ...(Object.hasOwn(source, 'q') ? { q: source['q'] } : {}),
    ...(Object.hasOwn(source, 'sort') ? { sort: source['sort'] } : {}),
    ...(Object.hasOwn(source, 'limit') ? { limit: source['limit'] } : {}),
    ...(Object.hasOwn(source, 'offset') ? { offset: source['offset'] } : {}),
  };
}

/**
 * The complete query the catalog runs, with defaults applied once. An
 * unusable address is refused as not found, which the route lane answers with
 * 404 rather than the 500 a plain throw would produce.
 */
export function catalogQuery(search: CatalogSearch): CatalogQuery {
  const parsed = searchSchema.safeParse(search);
  if (!parsed.success) throw notFound();
  const result = makeCatalogQuery({
    q: parsed.data.q ?? '',
    sort: parsed.data.sort ?? 'recent',
    limit: parsed.data.limit ?? 20,
    offset: parsed.data.offset ?? 0,
  });
  if (!result.ok) throw notFound();
  return result.value;
}
