import { z } from 'zod';

export const authenticatedUserSchema = z.object({
  id: z.uuid(),
  email: z.string().min(3).max(254),
  displayName: z.string().min(1).max(120),
  role: z.enum(['student', 'instructor', 'admin']),
});
