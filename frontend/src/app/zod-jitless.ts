import { z } from 'zod';

// Enforcing CSP forbids Zod's new Function capability probe and JIT validation.
z.config({ jitless: true });
