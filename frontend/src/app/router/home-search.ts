import type { SearchSchemaInput } from '@tanstack/react-router';
import { z } from 'zod';

const schema = z.object({ notice: z.enum(['verify-email']).optional() });
export type HomeSearch = { readonly notice?: 'verify-email' };
export function parseHomeSearch(search: SearchSchemaInput & Partial<HomeSearch>): HomeSearch {
  const parsed = schema.safeParse(search);
  return parsed.success && parsed.data.notice !== undefined ? { notice: parsed.data.notice } : {};
}
