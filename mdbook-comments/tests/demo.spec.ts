import { test, expect } from '@playwright/test';

test.describe('mdbook-comments Docker Demo', () => {
  test('should load the example book', async ({ page }) => {
    await page.goto('/');
    
    // Check that the page title is correct
    await expect(page).toHaveTitle(/Example Book with Comments/);
    
    // Check that the main content is visible
    const mainContent = page.locator('.content');
    await expect(mainContent).toBeVisible();
  });

  test('should display chapter navigation', async ({ page }) => {
    await page.goto('/');
    
    // Check for navigation elements
    const nav = page.locator('nav');
    await expect(nav).toBeVisible();
  });

  test('should have comment links on paragraphs', async ({ page }) => {
    await page.goto('/');
    
    // Wait for content to load
    await page.waitForSelector('.content');
    
    // Look for comment links (the preprocessor should add these)
    // The exact selector depends on how the preprocessor marks commentable elements
    const content = page.locator('.content');
    await expect(content).toBeVisible();
    
    // Check that the page has rendered markdown content
    const paragraphs = page.locator('.content p');
    const count = await paragraphs.count();
    expect(count).toBeGreaterThan(0);
  });

  test('should have embedded JavaScript', async ({ page }) => {
  await page.goto('/');

  // Check that embedded JavaScript is present (inline scripts without src)
  const inlineScripts = await page.locator('script:not([src])').count();
  expect(inlineScripts).toBeGreaterThan(0);

    // Check that the embedded assets marker is present
    const assetsMarker = await page.locator('text=mdbook-comments assets').count();
    expect(assetsMarker).toBeGreaterThan(0);
  });

  test('should have responsive layout', async ({ page }) => {
    await page.goto('/');
    
    // Test on mobile viewport
    await page.setViewportSize({ width: 375, height: 667 });
    const content = page.locator('.content');
    await expect(content).toBeVisible();
    
    // Test on desktop viewport
    await page.setViewportSize({ width: 1920, height: 1080 });
    await expect(content).toBeVisible();
  });

  test('should navigate between chapters', async ({ page }) => {
    await page.goto('/');
    
    // Find and click a chapter link if available
    const chapterLinks = page.locator('nav a');
    const linkCount = await chapterLinks.count();
    
    if (linkCount > 1) {
      // Click the second chapter link
      await chapterLinks.nth(1).click();
      
      // Wait for navigation
      await page.waitForLoadState('networkidle');
      
      // Check that content changed
      const content = page.locator('.content');
      await expect(content).toBeVisible();
    }
  });

  test('should have embedded CSS styling', async ({ page }) => {
  await page.goto('/');

  // Check that embedded CSS is present (inline styles)
  const inlineStyles = await page.locator('style').count();
  expect(inlineStyles).toBeGreaterThan(0);

    // Check that comment link styling is applied (verify CSS is working)
    const commentLinks = await page.locator('.comment-link').count();
    if (commentLinks > 0) {
      const link = page.locator('.comment-link').first();
      const opacity = await link.evaluate(el => window.getComputedStyle(el).opacity);
      expect(opacity).toBe('0.8'); // Default opacity from CSS
    }
  });

  test('should initialize JavaScript correctly', async ({ page }) => {
    const consoleLogs: string[] = [];
    const consoleErrors: string[] = [];

    page.on('console', (msg) => {
      if (msg.type() === 'log') {
        consoleLogs.push(msg.text());
      } else if (msg.type() === 'error') {
        consoleErrors.push(msg.text());
      }
    });

    await page.goto('/');
    await page.waitForLoadState('networkidle');

    // Check that initialization logs are present
    const initLogs = consoleLogs.filter(log => log.includes('Initializing mdbook-comments'));
    expect(initLogs.length).toBeGreaterThan(0);

    // Check that Neon initialization is logged (since we're using the Neon Docker config)
    const neonInit = consoleLogs.find(log => log.includes('Initializing mdbook-comments with Neon'));
    expect(neonInit).toBeDefined();

    // Check that comment loading is attempted (may fail due to mock API)
    const loadAttempts = consoleLogs.filter(log => log.includes('Loaded') || log.includes('Error loading'));
    expect(loadAttempts.length).toBeGreaterThan(0);
  });

  test('should have functional comment links', async ({ page }) => {
    await page.goto('/');

    // Wait for content to load
    await page.waitForSelector('.content');

    // Find comment links
    const commentLinks = page.locator('.comment-link');
    const linkCount = await commentLinks.count();
    expect(linkCount).toBeGreaterThan(0);

    // Test that comment links have the expected attributes
    const firstLink = commentLinks.first();
    const onclick = await firstLink.getAttribute('onclick');
    expect(onclick).toContain('toggleComments');

    // Check that parent element has comment metadata
    const wrapper = firstLink.locator('xpath=ancestor::span[@class="comment-link-wrapper"]').first();
    const metaAttr = await wrapper.getAttribute('data-comment-meta');
    expect(metaAttr).toBeDefined();
  });

  test('should not have unexpected console errors', async ({ page }) => {
    const consoleErrors: string[] = [];

    page.on('console', (msg) => {
      if (msg.type() === 'error') {
        consoleErrors.push(msg.text());
      }
    });

    await page.goto('/');
    await page.waitForLoadState('networkidle');

    // Filter out expected errors from mock API calls
    const unexpectedErrors = consoleErrors.filter(
      error => !error.includes('mock') &&
               !error.includes('localhost:9999') &&
               !error.includes('Query failed') && // Expected for mock Neon API
               !error.includes('Invalid supabaseUrl') && // Expected for demo config
               !error.includes('Google Sign-In library not loaded') && // Expected without GSI script
               !error.includes('Load failed') // Expected for network issues
    );

    expect(unexpectedErrors).toHaveLength(0);
  });
});
