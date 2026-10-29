import { defineConfig } from 'orval';

export default defineConfig({
  courses: {
    input: '../backend/openapi/openapi.yaml',
    output: {
      target: './src/adapters/api/generated/courses.ts',
      schemas: './src/adapters/api/generated/models',
      // The whole generated/ tree is Orval-owned, so clean it each run: a schema the backend
      // stops emitting (e.g. a component the overlay inlines away) must not linger as a dead
      // model file and barrel export. Without this, orphaned files survive regeneration.
      clean: true,
      client: 'fetch',
      mode: 'split',
      baseUrl: {
        runtime: 'apiOrigin()',
        imports: [{ name: 'apiOrigin', importPath: '../api-origin' }],
      },
      override: { fetch: { includeHttpResponseReturnType: true } },
    },
  },
});
