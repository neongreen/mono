import { defineConfig, devices } from '@playwright/test';

/**
 * Playwright configuration for testing mdbook-comments with json-server
 */
export default defineConfig({
  testDir: './tests',
  fullyParallel: false, // Run tests serially to avoid json-server race conditions
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: 1, // Single worker to avoid database conflicts
  reporter: 'line',
  timeout: 30000, // 30 seconds per test
  use: {
    baseURL: 'http://localhost:3300',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
  },

  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],

  // Start json-server and mdbook serve before tests
  webServer: [
    {
      command: 'npx json-server db.json --port 54322 --middlewares json-server-middleware.js --routes routes.json',
      port: 54322,
      timeout: 30000,
      reuseExistingServer: !process.env.CI,
    },
    {
      command: 'cd example-book && PATH="../target/release:$PATH" mdbook serve --port 3300',
      port: 3300,
      timeout: 30000,
      reuseExistingServer: !process.env.CI,
    },
  ],
});
