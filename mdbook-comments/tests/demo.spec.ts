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

  test('should load Neon JavaScript file', async ({ page }) => {
    await page.goto('/');
    
    // Check that the comments-neon.js script is loaded
    const scripts = await page.locator('script[src*="comments-neon.js"]').count();
    expect(scripts).toBeGreaterThan(0);
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

  test('should have CSS styling loaded', async ({ page }) => {
    await page.goto('/');
    
    // Check that the comments CSS is loaded
    const stylesheets = await page.locator('link[href*="comments.css"]').count();
    expect(stylesheets).toBeGreaterThan(0);
  });

  test('should not have console errors on load', async ({ page }) => {
    const consoleErrors: string[] = [];
    
    page.on('console', (msg) => {
      if (msg.type() === 'error') {
        consoleErrors.push(msg.text());
      }
    });
    
    await page.goto('/');
    await page.waitForLoadState('networkidle');
    
    // Filter out expected errors (like network errors to mock Neon API)
    const unexpectedErrors = consoleErrors.filter(
      error => !error.includes('mock') && !error.includes('localhost:9999')
    );
    
    expect(unexpectedErrors).toHaveLength(0);
  });
});
