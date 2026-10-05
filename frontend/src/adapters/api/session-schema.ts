import { z } from 'zod';

export const authenticatedUserSchema = z
  .object({
    id: z.uuid(),
    email: z.string().min(3).max(254),
    displayName: z.string().min(1).max(120),
    emailVerified: z.boolean(),
    twoFactorEnabled: z.boolean(),
    role: z.enum(['student', 'instructor', 'admin']),
  })
  .strict();

/** Empty success bodies are still a wire contract, not an unchecked special case. */
export const emptyResponseSchema = z.undefined();
export const twoFactorRequiredSchema = z
  .object({ status: z.literal('two_factor_required'), pendingId: z.string().min(1).max(128) })
  .strict();
export const twoFactorEnrollmentSchema = z
  .object({ secret: z.string().min(16), otpauthUri: z.url() })
  .strict();
export const recoveryCodesSchema = z
  .object({ recoveryCodes: z.array(z.string().min(1).max(128)).length(10) })
  .strict();
