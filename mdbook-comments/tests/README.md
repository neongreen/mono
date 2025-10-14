# Playwright Tests for mdbook-comments Docker Demo

This directory contains end-to-end tests for the mdbook-comments Docker demo using Playwright.

## Prerequisites

- Node.js 20 or higher
- Docker
- npm

## Setup

Install dependencies:

```bash
npm install
```

Install Playwright browsers:

```bash
npx playwright install chromium
```

## Running Tests

### Build the Docker image

Before running tests, build the Docker image:

```bash
docker build -t mdbook-comments-demo:test ..
```

### Run all tests

```bash
npm test
```

### Run tests with UI mode

```bash
npm run test:ui
```

### Run tests in headed mode

```bash
npm run test:headed
```

### Debug tests

```bash
npm run test:debug
```

## What the Tests Cover

The test suite verifies:

1. **Page Loading** - Example book loads correctly
2. **Navigation** - Chapter navigation works
3. **Content** - Markdown content is rendered properly
4. **Comment System** - Comment links and JavaScript are loaded
5. **Responsive Design** - Layout works on different viewport sizes
6. **Asset Loading** - CSS and JavaScript files are properly loaded
7. **No Console Errors** - No unexpected JavaScript errors

## CI/CD

These tests run automatically in GitHub Actions on every push and pull request. The workflow:

1. Builds the Docker image
2. Installs Node.js and Playwright
3. Runs the test suite
4. Uploads test reports as artifacts

## Test Configuration

The tests are configured in `playwright.config.ts`. Key settings:

- **Base URL**: `http://localhost:3000`
- **Web Server**: Automatically starts Docker container for testing
- **Browser**: Chromium (headless)
- **Retries**: 2 retries on CI, 0 locally
- **Screenshots**: Captured on failure
- **Trace**: Captured on first retry

## Writing New Tests

Add new test files in this directory with the pattern `*.spec.ts`. Example:

```typescript
import { test, expect } from '@playwright/test';

test('my new test', async ({ page }) => {
  await page.goto('/');
  // Your test code here
});
```

## Troubleshooting

### Port 3000 already in use

If port 3000 is already in use, stop the Docker container:

```bash
docker ps | grep mdbook-comments-demo
docker stop <container-id>
```

### Tests timeout

Increase the timeout in `playwright.config.ts`:

```typescript
timeout: 60 * 1000, // 60 seconds
```

### Docker image not found

Make sure you've built the Docker image with the correct tag:

```bash
docker build -t mdbook-comments-demo:test ..
```
