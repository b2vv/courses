import type { SearchSchemaInput } from '@tanstack/react-router';
import { z } from 'zod';
import type { LearnerFilters } from '../../domain/teach/teach';

const uuidPattern = /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i;
const courseIdSchema = z.string().refine((value) => uuidPattern.test(value));
const stateSchema = z.enum(['active', 'blocked']);
const querySchema = z
  .string()
  .transform((value) => value.trim())
  .refine((value) => value.length >= 1 && value.length <= 80 && !value.includes('\u0000'));

/** Raw known URL fields. Each stays unknown until its own parser runs. */
export type LearnerSearch = {
  readonly courseId?: unknown;
  readonly state?: unknown;
  readonly q?: unknown;
};

export function parseLearnerSearch(
  search: SearchSchemaInput & Partial<LearnerSearch>,
): LearnerSearch {
  return validateLearnerSearch(search);
}

/**
 * Echoes known parameters exactly as received — defaults and refusal do not belong in
 * validateSearch, or the router would rewrite or reject an otherwise usable address.
 */
export function validateLearnerSearch(search: unknown): LearnerSearch {
  if (typeof search !== 'object' || search === null) return {};
  return {
    ...(Object.hasOwn(search, 'courseId') ? { courseId: Reflect.get(search, 'courseId') } : {}),
    ...(Object.hasOwn(search, 'state') ? { state: Reflect.get(search, 'state') } : {}),
    ...(Object.hasOwn(search, 'q') ? { q: Reflect.get(search, 'q') } : {}),
  };
}

/** Drops each unusable field without affecting the others; the cursor is not a URL field. */
export function learnerFilters(search: LearnerSearch): LearnerFilters {
  const courseId = courseIdSchema.safeParse(search.courseId);
  const state = stateSchema.safeParse(search.state);
  const q = querySchema.safeParse(search.q);
  return {
    ...(courseId.success ? { courseId: courseId.data } : {}),
    ...(state.success ? { state: state.data } : {}),
    ...(q.success ? { q: q.data } : {}),
  };
}
