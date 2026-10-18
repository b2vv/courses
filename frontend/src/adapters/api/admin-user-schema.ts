import { z } from 'zod';
import { validAdminUserId } from '../../domain/admin/administration';

function hasControlCharacter(value: string): boolean {
  for (const character of value) {
    const code = character.codePointAt(0) ?? 0;
    if (code < 0x20 || code === 0x7f) return true;
  }
  return false;
}

function validEmail(value: string): boolean {
  const at = value.indexOf('@');
  const bytes = new TextEncoder().encode(value).byteLength;
  return (
    bytes >= 3 &&
    bytes <= 254 &&
    value === value.trim().toLowerCase() &&
    at > 0 &&
    at === value.lastIndexOf('@') &&
    at < value.length - 1 &&
    !/\s/u.test(value) &&
    !hasControlCharacter(value)
  );
}

const userIdSchema = z.uuid().refine(validAdminUserId);
const emailSchema = z.string().refine(validEmail);
const displayNameSchema = z
  .string()
  .refine(
    (value) =>
      value.length > 0 &&
      value === value.trim() &&
      [...value].length <= 120 &&
      !hasControlCharacter(value),
  );
const timestampSchema = z.iso.datetime({ offset: true, precision: 6 });

export const adminUserSchema = z
  .object({
    id: userIdSchema,
    email: emailSchema,
    displayName: displayNameSchema,
    role: z.enum(['student', 'instructor', 'admin']),
    status: z.enum(['active', 'blocked', 'deleted']),
    updatedAt: timestampSchema,
  })
  .strict();

const paginationSchema = z
  .object({
    limit: z.number().int().min(1).max(50),
    offset: z.number().int().min(0).max(10000),
    total: z.number().int().nonnegative().max(Number.MAX_SAFE_INTEGER),
  })
  .strict();

export const adminUserPageSchema = z
  .object({
    items: z.array(adminUserSchema).max(50),
    pagination: paginationSchema,
  })
  .strict();
