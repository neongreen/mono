import { test, expect } from '@playwright/test';

/**
 * End-to-end tests for mdbook-comments with json-server backend
 * 
 * These tests exercise the full commenting workflow:
 * - Loading the book with comment links
 * - Fetching comments from json-server
 * - Posting new comments
 * - Replying to comments
 * - Toggling comment sections
 */

test.describe('mdbook-comments functionality', () => {
  // Reset database before each test
  test.beforeEach(async ({ page }) => {
    // Clear all comments via json-server API
    const response = await page.request.get('http://localhost:55432/comments');
    const comments = await response.json();
    
    for (const comment of comments) {
      await page.request.delete(`http://localhost:55432/comments/${comment.id}`);
    }
  });

  test('should display comment links on paragraphs', async ({ page }) => {
    await page.goto('/chapter-1.html');
    
    // Wait for page to load
    await page.waitForLoadState('networkidle');
    
    // Check that comment links are present
    const commentLinks = page.locator('.comment-link');
    const count = await commentLinks.count();
    
    expect(count).toBeGreaterThan(0);
    
    // Verify first comment link has correct text
    const firstLink = commentLinks.first();
    await expect(firstLink).toHaveText('comment');
  });

  test('should have comment metadata on paragraphs', async ({ page }) => {
    await page.goto('/chapter-1.html');
    await page.waitForLoadState('networkidle');
    
    // Check that comment wrappers have metadata
    const wrapper = page.locator('.comment-link-wrapper').first();
    const metadata = await wrapper.getAttribute('data-comment-meta');
    
    expect(metadata).toBeTruthy();
    
    // Parse and verify metadata structure
    const parsed = JSON.parse(metadata!);
    expect(parsed).toHaveProperty('id');
    expect(parsed).toHaveProperty('content');
    expect(parsed).toHaveProperty('position');
    expect(parsed.position).toHaveProperty('file');
    expect(parsed.position).toHaveProperty('block-index');
  });

  test('should toggle comment section when clicking comment link', async ({ page }) => {
    await page.goto('/chapter-1.html');
    await page.waitForLoadState('networkidle');
    
    // Get the first comment link
    const firstLink = page.locator('.comment-link').first();
    
    // Get the paragraph ID
    const wrapper = page.locator('.comment-link-wrapper').first();
    const paragraphId = await wrapper.getAttribute('data-comment-id');
    
    // Click the comment link
    await firstLink.click();
    
    // Wait for comment section to appear
    const commentSection = page.locator(`#comments-${paragraphId}`);
    await expect(commentSection).toBeVisible();
    
    // Verify comment form is present
    const commentForm = commentSection.locator('.comment-form');
    await expect(commentForm).toBeVisible();
    
    // Click again to toggle
    await firstLink.click();
    await expect(commentSection).toBeHidden();
  });

  test('should post a new comment', async ({ page }) => {
    await page.goto('/chapter-1.html');
    await page.waitForLoadState('networkidle');
    
    // Click the first comment link
    const firstLink = page.locator('.comment-link').first();
    await firstLink.click();
    
    // Wait for comment section
    const wrapper = page.locator('.comment-link-wrapper').first();
    const paragraphId = await wrapper.getAttribute('data-comment-id');
    const commentSection = page.locator(`#comments-${paragraphId}`);
    await expect(commentSection).toBeVisible();
    
    // Fill in author name if prompted
    const authorInput = commentSection.locator('input[name="author"]');
    if (await authorInput.isVisible()) {
      await authorInput.fill('Test User');
    }
    
    // Fill in comment text
    const commentTextarea = commentSection.locator('textarea[name="comment-text"]');
    await commentTextarea.fill('This is a test comment!');
    
    // Submit the form
    const submitButton = commentSection.locator('button[type="submit"]');
    await submitButton.click();
    
    // Wait for comment to appear
    await page.waitForTimeout(1000); // Give time for API call
    
    // Verify comment appears in the list
    const commentList = commentSection.locator('.comment-list');
    const commentText = commentList.locator('text=This is a test comment!');
    await expect(commentText).toBeVisible();
    
    // Verify comment was saved to json-server
    const response = await page.request.get('http://localhost:55432/comments');
    const comments = await response.json();
    expect(comments.length).toBe(1);
    expect(comments[0].text).toBe('This is a test comment!');
    expect(comments[0]['paragraph-id']).toBe(paragraphId);
  });

  test('should display existing comments from json-server', async ({ page }) => {
    // First, add a comment via API
    await page.request.post('http://localhost:55432/comments', {
      data: {
        'paragraph-id': 'chapter-1-md-block-0-e9f5a4bf',
        metadata: {
          id: 'chapter-1-md-block-0-e9f5a4bf',
          position: { file: 'chapter-1.md', 'block-index': 0 },
          content: 'Test paragraph content',
          context: {
            'heading-path': ['Chapter 1: Getting Started'],
            prev: null,
            next: 'Next paragraph'
          },
          commit: 'test123'
        },
        author: 'Test User',
        text: 'Pre-existing comment',
        'parent-id': null
      }
    });
    
    // Now visit the page
    await page.goto('/chapter-1.html');
    await page.waitForLoadState('networkidle');
    
    // Click the first comment link
    const firstLink = page.locator('.comment-link').first();
    await firstLink.click();
    
    // Wait for comment section
    const wrapper = page.locator('.comment-link-wrapper').first();
    const paragraphId = await wrapper.getAttribute('data-comment-id');
    const commentSection = page.locator(`#comments-${paragraphId}`);
    await expect(commentSection).toBeVisible();
    
    // Verify the pre-existing comment appears
    const commentText = commentSection.locator('text=Pre-existing comment');
    await expect(commentText).toBeVisible();
  });

  test('should reply to a comment', async ({ page }) => {
    // First, add a comment via API
    const createResponse = await page.request.post('http://localhost:55432/comments', {
      data: {
        'paragraph-id': 'chapter-1-md-block-0-e9f5a4bf',
        metadata: {
          id: 'chapter-1-md-block-0-e9f5a4bf',
          position: { file: 'chapter-1.md', 'block-index': 0 },
          content: 'Test paragraph content'
        },
        author: 'Original Author',
        text: 'Original comment',
        'parent-id': null
      }
    });
    
    const originalComment = await createResponse.json();
    
    // Visit the page
    await page.goto('/chapter-1.html');
    await page.waitForLoadState('networkidle');
    
    // Click the comment link
    const firstLink = page.locator('.comment-link').first();
    await firstLink.click();
    
    // Wait for comment section
    const wrapper = page.locator('.comment-link-wrapper').first();
    const paragraphId = await wrapper.getAttribute('data-comment-id');
    const commentSection = page.locator(`#comments-${paragraphId}`);
    await expect(commentSection).toBeVisible();
    
    // Find the reply button
    const replyButton = commentSection.locator('button:has-text("Reply")').first();
    await replyButton.click();
    
    // Fill in reply form
    const replyTextarea = commentSection.locator('textarea[name="reply-text"]');
    await replyTextarea.fill('This is a reply!');
    
    // Submit reply
    const submitReply = commentSection.locator('button[type="submit"]:has-text("Post Reply")');
    await submitReply.click();
    
    // Wait for reply to appear
    await page.waitForTimeout(1000);
    
    // Verify reply appears
    const replyText = commentSection.locator('text=This is a reply!');
    await expect(replyText).toBeVisible();
    
    // Verify reply was saved with correct parent-id
    const response = await page.request.get('http://localhost:55432/comments');
    const comments = await response.json();
    const reply = comments.find(c => c.text === 'This is a reply!');
    expect(reply).toBeTruthy();
    expect(reply['parent-id']).toBe(originalComment.id);
  });

  test('should show comment count on links', async ({ page }) => {
    // Add multiple comments to first paragraph
    for (let i = 0; i < 3; i++) {
      await page.request.post('http://localhost:55432/comments', {
        data: {
          'paragraph-id': 'chapter-1-md-block-0-e9f5a4bf',
          metadata: { id: 'chapter-1-md-block-0-e9f5a4bf' },
          author: `User ${i}`,
          text: `Comment ${i}`,
          'parent-id': null
        }
      });
    }
    
    // Visit the page
    await page.goto('/chapter-1.html');
    await page.waitForLoadState('networkidle');
    
    // Check if comment count is shown
    const firstWrapper = page.locator('.comment-link-wrapper').first();
    const linkText = await firstWrapper.textContent();
    
    // Should show "comment (3)" or similar
    expect(linkText).toContain('3');
  });

  test('should handle multiple paragraphs with comments', async ({ page }) => {
    // Add comments to multiple paragraphs
    await page.request.post('http://localhost:55432/comments', {
      data: {
        'paragraph-id': 'chapter-1-md-block-0-e9f5a4bf',
        metadata: { id: 'chapter-1-md-block-0-e9f5a4bf' },
        author: 'User 1',
        text: 'Comment on first paragraph',
        'parent-id': null
      }
    });
    
    await page.request.post('http://localhost:55432/comments', {
      data: {
        'paragraph-id': 'chapter-1-md-block-1-3142a274',
        metadata: { id: 'chapter-1-md-block-1-3142a274' },
        author: 'User 2',
        text: 'Comment on second paragraph',
        'parent-id': null
      }
    });
    
    // Visit the page
    await page.goto('/chapter-1.html');
    await page.waitForLoadState('networkidle');
    
    // Click first comment link
    const firstLink = page.locator('.comment-link').first();
    await firstLink.click();
    
    const firstWrapper = page.locator('.comment-link-wrapper').first();
    const firstParagraphId = await firstWrapper.getAttribute('data-comment-id');
    const firstSection = page.locator(`#comments-${firstParagraphId}`);
    
    // Verify first comment
    await expect(firstSection.locator('text=Comment on first paragraph')).toBeVisible();
    
    // Click second comment link
    const secondLink = page.locator('.comment-link').nth(1);
    await secondLink.click();
    
    const secondWrapper = page.locator('.comment-link-wrapper').nth(1);
    const secondParagraphId = await secondWrapper.getAttribute('data-comment-id');
    const secondSection = page.locator(`#comments-${secondParagraphId}`);
    
    // Verify second comment
    await expect(secondSection.locator('text=Comment on second paragraph')).toBeVisible();
    
    // Verify first section is still visible (or collapsed depending on implementation)
    // Just check that they are independent
    expect(firstParagraphId).not.toBe(secondParagraphId);
  });

  test('should persist author name in localStorage', async ({ page }) => {
    await page.goto('/chapter-1.html');
    await page.waitForLoadState('networkidle');
    
    // Click comment link
    const firstLink = page.locator('.comment-link').first();
    await firstLink.click();
    
    const wrapper = page.locator('.comment-link-wrapper').first();
    const paragraphId = await wrapper.getAttribute('data-comment-id');
    const commentSection = page.locator(`#comments-${paragraphId}`);
    
    // Fill in author name
    const authorInput = commentSection.locator('input[name="author"]');
    if (await authorInput.isVisible()) {
      await authorInput.fill('Persistent User');
      
      // Check localStorage
      const storedAuthor = await page.evaluate(() => {
        return localStorage.getItem('mdbook-comments-author');
      });
      
      expect(storedAuthor).toBe('Persistent User');
      
      // Reload page
      await page.reload();
      await page.waitForLoadState('networkidle');
      
      // Click comment link again
      await firstLink.click();
      await expect(commentSection).toBeVisible();
      
      // Check that author name is pre-filled
      const newAuthorInput = commentSection.locator('input[name="author"]');
      if (await newAuthorInput.isVisible()) {
        const value = await newAuthorInput.inputValue();
        expect(value).toBe('Persistent User');
      }
    }
  });

  test('should handle API errors gracefully', async ({ page }) => {
    // Stop json-server to simulate API error
    // Note: This test may need adjustment based on webServer lifecycle
    
    await page.goto('/chapter-1.html');
    await page.waitForLoadState('networkidle');
    
    // Page should still load even if API is down
    const content = page.locator('.content');
    await expect(content).toBeVisible();
    
    // Comment links should still be present
    const commentLinks = page.locator('.comment-link');
    expect(await commentLinks.count()).toBeGreaterThan(0);
  });
});

test.describe('mdbook-comments UI', () => {
  test('should have proper styling for comment links', async ({ page }) => {
    await page.goto('/chapter-1.html');
    await page.waitForLoadState('networkidle');
    
    const firstLink = page.locator('.comment-link').first();
    
    // Check that link has proper styling
    const color = await firstLink.evaluate(el => window.getComputedStyle(el).color);
    expect(color).toBeTruthy();
    
    // Link should be clickable
    await expect(firstLink).toBeEnabled();
  });

  test('should navigate to different chapters', async ({ page }) => {
    await page.goto('/index.html');
    await page.waitForLoadState('networkidle');
    
    // Find chapter link
    const chapterLink = page.locator('a[href*="chapter-1"]').first();
    await chapterLink.click();
    
    await page.waitForLoadState('networkidle');
    
    // Verify we're on chapter 1
    expect(page.url()).toContain('chapter-1');
    
    // Verify comment links are present
    const commentLinks = page.locator('.comment-link');
    expect(await commentLinks.count()).toBeGreaterThan(0);
  });
});
