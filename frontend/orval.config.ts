import { defineConfig } from 'orval';

export default defineConfig({
  courses: {
    input: '../backend/openapi/openapi.yaml',
    output: {
      target: './src/adapters/api/generated/courses.ts',
      schemas: './src/adapters/api/generated/models',
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
