import type { SearchSchemaInput } from '@tanstack/react-router';
import type { CatalogQuery } from '../../domain/course/course';
import { z } from 'zod';
import { makeCatalogQuery } from '../../domain/course/course';
import { CatalogQueryError } from '../composition/catalog-queries';

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
export function parseCatalogSearch(search: SearchSchemaInput & Partial<CatalogQuery>) {
  return validateCatalogSearch(search);
}

/** Router's typed navigation input still arrives from an untrusted URL at runtime. */
export function validateCatalogSearch(search: unknown): CatalogQuery {
  const parsed = searchSchema.safeParse(search);
  if (!parsed.success) throw new CatalogQueryError({ code: 'invalid_query' });
  const result = makeCatalogQuery({
    q: parsed.data.q ?? '',
    sort: parsed.data.sort ?? 'recent',
    limit: parsed.data.limit ?? 20,
    offset: parsed.data.offset ?? 0,
  });
  if (!result.ok) throw new CatalogQueryError(result.error);
  return result.value;
}
