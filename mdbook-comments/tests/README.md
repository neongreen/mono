# Playwright Tests for mdbook-comments

This directory contains end-to-end tests for mdbook-comments using Playwright with a json-server backend.

## Prerequisites

- Node.js 20 or higher
- Cargo (Rust toolchain)
- mdbook installed (`cargo install mdbook`)

## Setup

Install dependencies:

```bash
npm install
```

Install Playwright browsers:

```bash
npx playwright install chromium
```

Build the mdbook-comments preprocessor:

```bash
cargo build --release
```

Build the example book:

```bash
cd example-book && PATH="../target/release:$PATH" mdbook build && cd ..
```

## Running Tests

The test suite automatically starts:
1. json-server on port 55432 with test database
2. mdbook serve on port 3300

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

The test suite exercises full commenting functionality:

1. **Comment Links** - Verifies comment links are present on paragraphs
2. **Comment Metadata** - Checks that paragraphs have proper metadata attributes
3. **Toggle Comments** - Tests expanding/collapsing comment sections
4. **Post Comments** - Tests posting new comments via the UI
5. **Load Comments** - Verifies comments are loaded from json-server
6. **Reply to Comments** - Tests the reply functionality
7. **Comment Counts** - Verifies comment counts are shown on links
8. **Multiple Paragraphs** - Tests independent comment sections
9. **Author Persistence** - Tests localStorage for author names
10. **Error Handling** - Tests graceful degradation when API is unavailable

## Architecture

The tests use:

- **json-server**: Simple REST API for testing comment CRUD operations
- **mdbook serve**: Local HTTP server for the book
- **Playwright**: Browser automation for E2E testing

This setup allows testing the full comment workflow without external dependencies like Supabase or Neon.

## Test Database

Tests use json-server with `db.json` as the database. Each test:
1. Clears all comments before running (via DELETE API calls)
2. Sets up test data via POST API calls
3. Exercises the UI functionality
4. Verifies both UI state and database state

## CI/CD

These tests run automatically in GitHub Actions. The workflow:

1. Builds the mdbook-comments preprocessor
2. Builds the example book
3. Installs Node.js and Playwright
4. Runs the test suite with json-server and mdbook serve
5. Uploads test reports as artifacts on failure

## Test Configuration

The tests are configured in `playwright.config.ts`. Key settings:

- **Base URL**: `http://localhost:3300`
- **Web Servers**: json-server (55432) and mdbook serve (3300)
- **Browser**: Chromium (headless)
- **Workers**: 1 (to avoid database conflicts)
- **Retries**: 2 retries on CI, 0 locally
- **Screenshots**: Captured on failure
- **Trace**: Captured on first retry

## Writing New Tests

Add new tests to `tests/comments.spec.ts` or create new spec files. Example:

```typescript
import { test, expect } from '@playwright/test';

test('my new test', async ({ page }) => {
  await page.goto('/chapter-1.html');
  
  // Test comment functionality
  const commentLink = page.locator('.comment-link').first();
  await commentLink.click();
  
  // Your test code here
});
```

## Troubleshooting

### Port conflicts

If ports 3300 or 55432 are in use:

```bash
# Find and kill processes
lsof -i :3300
lsof -i :55432
kill <pid>
```

### Tests timeout

Increase timeout in `playwright.config.ts`:

```typescript
timeout: 60000, // 60 seconds
```

### Build failures

Ensure preprocessor is built:

```bash
cargo build --release
```

Ensure example book is built:

```bash
cd example-book && PATH="../target/release:$PATH" mdbook build
```
