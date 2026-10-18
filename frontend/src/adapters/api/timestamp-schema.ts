import { z } from 'zod';

// Shared ISO-8601 timestamp wire schema: offset-aware, microsecond precision. The API emits these
// for every entity's audit fields, so it lives in the shared adapter zone rather than any one
// feature module (both admin course rows and media lease rows parse with it).
export const isoTimestampSchema = z.iso.datetime({ offset: true, precision: 6 });
