import { defineConfig } from '@rstest/core';

export default defineConfig({
  testEnvironment: 'happy-dom',
  tools: { swc: { jsc: { transform: { react: { runtime: 'automatic' } } } } },
  include: ['src/**/*.test.{ts,tsx}'],
  setupFiles: ['./src/test-setup.ts'],
});
