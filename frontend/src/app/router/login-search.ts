import type { SearchSchemaInput } from '@tanstack/react-router';
import { z } from 'zod';
import { safeReturnTo } from '../../modules/identity';

const searchSchema = z.object({
  returnTo: z.string().optional(),
  notice: z.enum(['password-reset']).optional(),
});

export type LoginSearch = { readonly returnTo?: string; readonly notice?: 'password-reset' };

/**
 * The return target arrives from the URL, so it is narrowed to an internal path
 * here rather than trusted anywhere later. An absent or unusable value produces
 * no key at all: emitting a default would make the router rewrite a plain
 * `/login` into `/login?returnTo=...` on every visit.
 */
export function parseLoginSearch(search: SearchSchemaInput & Partial<LoginSearch>): LoginSearch {
  return validateLoginSearch(search);
}

/** Router's typed navigation input still arrives from an untrusted URL at runtime. */
export function validateLoginSearch(search: unknown): LoginSearch {
  const parsed = searchSchema.safeParse(search);
  const candidate = parsed.success ? parsed.data.returnTo : undefined;
  const notice = parsed.success ? parsed.data.notice : undefined;
  if (candidate === undefined) return notice === undefined ? {} : { notice };
  const safe = safeReturnTo(candidate, '');
  return safe === ''
    ? notice === undefined
      ? {}
      : { notice }
    : { returnTo: safe, ...(notice === undefined ? {} : { notice }) };
}
