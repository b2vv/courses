import type { SearchSchemaInput } from '@tanstack/react-router';
import { z } from 'zod';

const schema = z.object({ token: z.string().min(1).optional() });
export type TokenSearch = { readonly token?: string };
export function parseTokenSearch(search: SearchSchemaInput & Partial<TokenSearch>): TokenSearch {
  const parsed = schema.safeParse(search);
  return parsed.success && parsed.data.token !== undefined ? { token: parsed.data.token } : {};
}
