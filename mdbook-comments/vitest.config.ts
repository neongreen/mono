import { defineConfig } from 'vitest/config';

export default defineConfig({
  test: {
    environment: 'jsdom',
    globals: true,
    setupFiles: [],
    include: ['src/**/*.test.ts', 'src/**/*.test.tsx'],
    exclude: ['tests/**/*', 'node_modules/**/*'],
  },
  esbuild: {
    jsx: 'automatic',
    jsxImportSource: 'preact',
  },
});