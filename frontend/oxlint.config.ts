import { defineConfig } from 'oxlint';

export default defineConfig({
  plugins: ['typescript', 'react', 'import', 'jsx-a11y'],
  env: { browser: true, node: true },
  ignorePatterns: ['src/routeTree.gen.ts', 'src/adapters/api/generated/**', 'dist/**'],
  rules: {
    'typescript/no-explicit-any': 'error',
    'typescript/no-non-null-assertion': 'error',
    'import/no-cycle': 'error',
    'react-hooks/rules-of-hooks': 'error',
    'react-hooks/exhaustive-deps': 'error',
  },
});
