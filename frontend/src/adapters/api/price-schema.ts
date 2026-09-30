import { z } from 'zod';

/**
 * OpenAPI `Price`: a positive amount in minor units; UAH is the only currency
 * for now. One schema for every response that carries a price (catalog, admin,
 * learning), so a new currency is added in one place and cannot drift.
 */
export const priceSchema = z
  .object({
    minor: z.number().int().min(1).max(Number.MAX_SAFE_INTEGER),
    currency: z.enum(['UAH']),
  })
  .strict();
