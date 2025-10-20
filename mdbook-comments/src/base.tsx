/**
 * mdbook-comments - Shared base module
 *
 * This module contains all common functionality shared across backend implementations:
 * - Fuzzy matching algorithms
 * - Comment matching and orphaned comment detection
 * - DOM manipulation and UI rendering
 * - CSS injection
 * - Utility functions
 *
 * Backend adapters must implement the BackendAdapter interface.
 */

import { render } from 'preact';
import { CommentSection, OrphanedComments } from './components';
import type {
  Comment,
  BackendAdapter,
  ParagraphMetadata,
  MatchedComment,
} from './types';

/**
 * Configuration
 */
interface Config {
  similarityThreshold: number;
  orphanedLocation: 'end-of-chapter' | 'end-of-page';
  showCommentCount: boolean;
}

const CONFIG: Config = {
  similarityThreshold: 0.85,
  orphanedLocation: 'end-of-chapter',
  showCommentCount: true,
};

/**
 * State
 */
interface State {
  backend: BackendAdapter | null;
  allComments: Comment[];
  currentPageComments: MatchedComment[];
  orphanedComments: Comment[];
}

const state: State = {
  backend: null,
  allComments: [],
  currentPageComments: [],
  orphanedComments: [],
};

/**
 * Main initialization function
 * Called by backend adapter with backend implementation
 */
export async function init(backendAdapter: BackendAdapter): Promise<void> {
  state.backend = backendAdapter;

  console.log('Initializing mdbook-comments base module...');

  // Initialize backend
  if (state.backend.init) {
    await state.backend.init();
  }

  // Set up auth change listener if supported
  if ('onAuthChange' in state.backend && state.backend.onAuthChange) {
    state.backend.onAuthChange(() => {
      // Refresh UI when auth state changes
      refreshAllCommentSections();
    });
  }

  // Load comments
  await loadComments();

  // Inject styles
  injectStyles();
}

/**
 * Load comments from backend
 */
async function loadComments(): Promise<void> {
  if (!state.backend) {
    throw new Error('Backend not initialized');
  }

  try {
    state.allComments = await state.backend.loadComments();
    console.log(`Loaded ${state.allComments.length} comments`);

    // Build reply structure if comments don't already have it
    buildReplyStructure();

    // Match comments to current paragraphs
    matchComments();

    // Update comment counts in links
    updateCommentCounts();

    // Display orphaned comments if any
    displayOrphanedComments();
  } catch (error) {
    console.error('Error loading comments:', error);
  }
}

/**
 * Build reply structure from flat comment list
 */
function buildReplyStructure(): void {
  // Create a map to collect replies for each parent
  const repliesByParent: Record<string, Comment[]> = {};

  // First, create new comment objects (immutable)
  const newComments = state.allComments.map((c) => ({ ...c, replies: [] }));

  // Collect replies for each parent ID (using new objects)
  newComments.forEach((comment) => {
    if (comment.parent_id) {
      if (!repliesByParent[comment.parent_id]) {
        repliesByParent[comment.parent_id] = [];
      }
      repliesByParent[comment.parent_id].push(comment);
    }
  });

  // Update the new comments with their replies
  state.allComments = newComments.map((comment) => ({
    ...comment,
    replies: repliesByParent[comment.id] || [],
  }));
}

/**
 * Match comments to current paragraphs using fuzzy matching
 */
function matchComments(): void {
  state.currentPageComments = [];
  state.orphanedComments = [];

  const commentLinks = document.querySelectorAll('.comment-link-wrapper');
  const usedComments = new Set<string>();

  // First pass: exact ID matching
  commentLinks.forEach((link) => {
    const paragraphId = link.getAttribute('data-comment-id');

    if (!paragraphId) return;

    const exactMatches = state.allComments.filter(
      (c) =>
        c.metadata &&
        c.metadata.id === paragraphId &&
        !usedComments.has(c.id) &&
        !c.parent_id // Only top-level comments
    );

    exactMatches.forEach((comment) => {
      state.currentPageComments.push({
        paragraphId,
        comment,
        confidence: 1.0,
      });
      usedComments.add(comment.id);
    });
  });

  // Second pass: fuzzy matching for comments without exact match
  commentLinks.forEach((link) => {
    const paragraphId = link.getAttribute('data-comment-id');
    const metaStr = link.getAttribute('data-comment-meta') || '{}';
    const metadata: ParagraphMetadata = JSON.parse(metaStr);

    if (!paragraphId) return;

    state.allComments.forEach((comment) => {
      if (usedComments.has(comment.id)) return;
      if (!comment.metadata) return;
      if (comment.parent_id) return; // Skip replies

      const similarity = calculateSimilarity(metadata, comment.metadata);

      if (similarity >= CONFIG.similarityThreshold) {
        state.currentPageComments.push({
          paragraphId,
          comment,
          confidence: similarity,
        });
        usedComments.add(comment.id);
      }
    });
  });

  // Remaining comments are orphaned
  state.allComments.forEach((comment) => {
    if (!usedComments.has(comment.id) && !comment.parent_id) {
      state.orphanedComments.push(comment);
    }
  });

  console.log(
    `Matched ${state.currentPageComments.length} comments, ${state.orphanedComments.length} orphaned`
  );
}

/**
 * Calculate similarity between two paragraph metadata objects
 */
function calculateSimilarity(
  meta1: ParagraphMetadata,
  meta2: ParagraphMetadata
): number {
  let score = 0.0;
  let weights = 0.0;

  // Content similarity (most important)
  if (meta1.content && meta2.content) {
    const contentSim = textSimilarity(meta1.content, meta2.content);
    score += contentSim * 0.5;
    weights += 0.5;
  }

  // Context similarity (prev/next paragraphs)
  if (meta1.context && meta2.context) {
    if (meta1.context.prev && meta2.context.prev) {
      const prevSim = textSimilarity(meta1.context.prev, meta2.context.prev);
      score += prevSim * 0.2;
      weights += 0.2;
    }

    if (meta1.context.next && meta2.context.next) {
      const nextSim = textSimilarity(meta1.context.next, meta2.context.next);
      score += nextSim * 0.2;
      weights += 0.2;
    }

    // Heading path similarity
    if (meta1.context['heading-path'] && meta2.context['heading-path']) {
      const headingSim = arraySimilarity(
        meta1.context['heading-path'],
        meta2.context['heading-path']
      );
      score += headingSim * 0.1;
      weights += 0.1;
    }
  }

  return weights > 0 ? score / weights : 0.0;
}

/**
 * Calculate text similarity using simple token-based approach (Jaccard similarity)
 */
function textSimilarity(text1: string, text2: string): number {
  const tokens1 = new Set(tokenize(text1));
  const tokens2 = new Set(tokenize(text2));

  const intersection = new Set(
    Array.from(tokens1).filter((x) => tokens2.has(x))
  );
  const union = new Set([...Array.from(tokens1), ...Array.from(tokens2)]);

  return union.size > 0 ? intersection.size / union.size : 0.0;
}

/**
 * Tokenize text into words
 */
function tokenize(text: string): string[] {
  return text
    .toLowerCase()
    .replace(/[^\w\s]/g, ' ')
    .split(/\s+/)
    .filter((t) => t.length > 2);
}

/**
 * Calculate array similarity (Jaccard index)
 */
function arraySimilarity(arr1: string[], arr2: string[]): number {
  const set1 = new Set(arr1);
  const set2 = new Set(arr2);

  const intersection = new Set(Array.from(set1).filter((x) => set2.has(x)));
  const union = new Set([...Array.from(set1), ...Array.from(set2)]);

  return union.size > 0 ? intersection.size / union.size : 0.0;
}

/**
 * Update comment counts in links
 */
function updateCommentCounts(): void {
  document.querySelectorAll('.comment-link-wrapper').forEach((wrapper) => {
    const paragraphId = wrapper.getAttribute('data-comment-id');
    if (!paragraphId) return;

    const count = state.currentPageComments.filter(
      (c) => c.paragraphId === paragraphId
    ).length;

    if (count > 0 && CONFIG.showCommentCount) {
      const link = wrapper.querySelector('.comment-link');
      if (link) {
        link.textContent = `comment (${count})`;
      }
    }
  });
}

/**
 * Display orphaned comments at end of chapter
 */
function displayOrphanedComments(): void {
  // Remove existing orphaned comments section if present
  const existing = document.querySelector('.orphaned-comments-section');
  if (existing) {
    existing.remove();
  }

  if (state.orphanedComments.length === 0) return;

  const main =
    document.querySelector('main') ||
    document.querySelector('#content') ||
    document.body;

  const container = document.createElement('div');
  main.appendChild(container);

  render(
    <OrphanedComments comments={state.orphanedComments} />,
    container
  );
}

/**
 * Toggle comments visibility for a paragraph
 */
export function toggleComments(paragraphId: string): void {
  let section = document.getElementById(`comments-${paragraphId}`);

  if (section) {
    section.style.display = section.style.display === 'none' ? 'block' : 'none';
    return;
  }

  section = createCommentSection(paragraphId);

  const wrapper = document.querySelector(
    `[data-comment-id="${paragraphId}"]`
  );
  if (wrapper && wrapper.parentNode) {
    wrapper.parentNode.insertBefore(section, wrapper.nextSibling);
  }
}

/**
 * Create comment section using Preact
 */
function createCommentSection(paragraphId: string): HTMLElement {
  if (!state.backend) {
    throw new Error('Backend not initialized');
  }

  // Get metadata for this paragraph
  const wrapper = document.querySelector(
    `[data-comment-id="${paragraphId}"]`
  );
  const metaStr = wrapper?.getAttribute('data-comment-meta') || '{}';
  const metadata: ParagraphMetadata = JSON.parse(metaStr);

  // Get comments for this paragraph
  const comments = state.currentPageComments.filter(
    (c) => c.paragraphId === paragraphId
  );

  // Create wrapper div that will be inserted into DOM and kept as render target
  const wrapperDiv = document.createElement('div');
  wrapperDiv.style.display = 'contents'; // Make wrapper transparent in layout

  // Render Preact component with update callback that reloads and re-renders
  const handleUpdate = async () => {
    await loadComments();
    // Re-render this section with updated comments
    const updatedComments = state.currentPageComments.filter(
      (c) => c.paragraphId === paragraphId
    );
    // Re-render into the same wrapper (which is now in the DOM)
    render(
      <CommentSection
        paragraphId={paragraphId}
        metadata={metadata}
        comments={updatedComments}
        backend={state.backend}
        onUpdate={handleUpdate}
      />,
      wrapperDiv
    );
  };

  render(
    <CommentSection
      paragraphId={paragraphId}
      metadata={metadata}
      comments={comments}
      backend={state.backend}
      onUpdate={handleUpdate}
    />,
    wrapperDiv
  );

  // Return the wrapper itself (not firstElementChild), so we can re-render into it
  return wrapperDiv;
}

/**
 * Show reply form
 */
export function showReplyForm(commentId: string): void {
  const form = document.getElementById(`reply-form-${commentId}`);
  if (form) {
    form.style.display = form.style.display === 'none' ? 'block' : 'none';
  }
}

/**
 * Submit a new comment
 * NOTE: This function is kept for backward compatibility.
 * The CommentForm component handles submission internally.
 */
export async function submitComment(paragraphId: string): Promise<void> {
  if (!state.backend) {
    throw new Error('Backend not initialized');
  }

  const wrapper = document.querySelector(
    `[data-comment-id="${paragraphId}"]`
  );
  if (!wrapper) return;

  const section = document.getElementById(`comments-${paragraphId}`);
  if (!section) return;

  const form = section.querySelector('.comment-form');
  if (!form) return;

  // Get author if needed
  let author = state.backend.getCurrentAuthor
    ? state.backend.getCurrentAuthor()
    : null;
  if (state.backend.showAuthorInput && !author) {
    const authorInput = form.querySelector('.author-input') as HTMLInputElement;
    if (authorInput) {
      author = authorInput.value.trim();
      if (!author) {
        alert('Please enter your name');
        return;
      }
      // Save author for future comments
      if (state.backend.setCurrentAuthor) {
        state.backend.setCurrentAuthor(author);
      }
    }
  }

  const textarea = form.querySelector('.comment-input') as HTMLTextAreaElement;
  if (!textarea) return;

  const text = textarea.value.trim();

  if (!text) {
    alert('Please enter a comment');
    return;
  }

  const metaStr = wrapper.getAttribute('data-comment-meta') || '{}';
  const metadata: ParagraphMetadata = JSON.parse(metaStr);

  try {
    const newComment = await state.backend.saveComment(
      paragraphId,
      metadata,
      text,
      author || 'Anonymous'
    );

    // Add to local state
    newComment.replies = newComment.replies || [];
    state.allComments.push(newComment);
    state.currentPageComments.push({
      paragraphId,
      comment: newComment,
      confidence: 1.0,
    });

    // Clear textarea
    textarea.value = '';

    // Reload comments and re-render section
    await loadComments();
  } catch (error) {
    console.error('Error posting comment:', error);
    alert('Failed to post comment. Please try again.');
  }
}

/**
 * Submit a reply to a comment
 * NOTE: This function is kept for backward compatibility.
 * The ReplyForm component handles submission internally.
 */
export async function submitReply(commentId: string): Promise<void> {
  if (!state.backend) {
    throw new Error('Backend not initialized');
  }

  // Get author if needed
  let author = state.backend.getCurrentAuthor
    ? state.backend.getCurrentAuthor()
    : null;

  const form = document.getElementById(`reply-form-${commentId}`);
  if (!form) return;

  const textarea = form.querySelector('.reply-input') as HTMLTextAreaElement;
  if (!textarea) return;

  const text = textarea.value.trim();

  if (!text) {
    alert('Please enter a reply');
    return;
  }

  try {
    const newReply = await state.backend.saveReply(
      commentId,
      text,
      author || 'Anonymous'
    );

    // Update local comment
    const comment = state.allComments.find((c) => c.id === commentId);
    if (comment) {
      if (!comment.replies) comment.replies = [];
      comment.replies.push(newReply);
    }

    // Add to global list
    state.allComments.push(newReply);

    // Clear textarea and hide form
    textarea.value = '';
    form.style.display = 'none';

    // Reload comments and re-render
    await loadComments();
  } catch (error) {
    console.error('Error posting reply:', error);
    alert('Failed to post reply. Please try again.');
  }
}

/**
 * Refresh all open comment sections (e.g., after auth change)
 */
async function refreshAllCommentSections(): Promise<void> {
  // Reload comments first
  await loadComments();

  // Re-render all visible comment sections
  document.querySelectorAll('.comment-section').forEach((section) => {
    const paragraphId = section.getAttribute('data-paragraph-id');
    const htmlSection = section as HTMLElement;
    if (paragraphId && htmlSection.style.display !== 'none') {
      // Get the parent container and re-render
      const parent = section.parentElement;
      if (parent && state.backend) {
        const wrapper = document.querySelector(
          `[data-comment-id="${paragraphId}"]`
        );
        const metaStr = wrapper?.getAttribute('data-comment-meta') || '{}';
        const metadata: ParagraphMetadata = JSON.parse(metaStr);

        const comments = state.currentPageComments.filter(
          (c) => c.paragraphId === paragraphId
        );

        render(
          <CommentSection
            paragraphId={paragraphId}
            metadata={metadata}
            comments={comments}
            backend={state.backend}
            onUpdate={() => loadComments()}
          />,
          parent,
          section as Element
        );
      }
    }
  });
}

/**
 * Escape HTML to prevent XSS
 * Exported for use by components and utilities
 */
export function escapeHtml(text: string): string {
  const div = document.createElement('div');
  div.textContent = text;
  return div.innerHTML;
}

/**
 * Format date for display
 * Exported for use by components and utilities
 */
export function formatDate(dateStr: string): string {
  if (!dateStr) return '';
  const date = new Date(dateStr);
  return date.toLocaleString();
}

/**
 * Inject CSS styles
 */
function injectStyles(): void {
  if (document.getElementById('mdbook-comments-styles')) return;

  const style = document.createElement('style');
  style.id = 'mdbook-comments-styles';
  style.textContent = `
            .comment-link-wrapper {
                display: inline;
                margin-left: 0.5em;
            }

            .comment-link {
                font-size: 0.85em;
                color: #0066cc;
                text-decoration: underline;
                cursor: pointer;
            }

            .comment-link:hover {
                color: #0052a3;
            }

            .comment-section {
                margin: 1em 0;
                padding: 1em;
                background: #f5f5f5;
                border-left: 3px solid #0066cc;
                border-radius: 4px;
            }

            .comment-list {
                margin-bottom: 1em;
            }

            .comment-item {
                background: white;
                padding: 0.75em;
                margin-bottom: 0.75em;
                border-radius: 4px;
                box-shadow: 0 1px 3px rgba(0,0,0,0.1);
            }

            .comment-header {
                display: flex;
                justify-content: space-between;
                margin-bottom: 0.5em;
                font-size: 0.9em;
                color: #666;
            }

            .comment-author {
                font-weight: bold;
                color: #333;
            }

            .comment-text {
                line-height: 1.5;
                white-space: pre-wrap;
            }

            .comment-replies {
                margin-top: 0.75em;
                margin-left: 1.5em;
                border-left: 2px solid #ddd;
                padding-left: 0.75em;
            }

            .reply-item {
                background: #fafafa;
                padding: 0.5em;
                margin-bottom: 0.5em;
                border-radius: 3px;
            }

            .reply-header {
                display: flex;
                justify-content: space-between;
                margin-bottom: 0.25em;
                font-size: 0.85em;
                color: #666;
            }

            .reply-author {
                font-weight: bold;
                color: #333;
            }

            .reply-text {
                font-size: 0.95em;
                line-height: 1.4;
                white-space: pre-wrap;
            }

            .comment-reply-btn {
                margin-top: 0.5em;
                padding: 0.25em 0.75em;
                font-size: 0.85em;
                background: #f0f0f0;
                border: 1px solid #ddd;
                border-radius: 3px;
                cursor: pointer;
            }

            .comment-reply-btn:hover {
                background: #e0e0e0;
            }

            .comment-form, .reply-form {
                margin-top: 0.75em;
            }

            .author-input {
                width: 100%;
                padding: 0.5em;
                margin-bottom: 0.5em;
                border: 1px solid #ddd;
                border-radius: 3px;
                font-family: inherit;
                font-size: 0.95em;
            }

            .comment-input, .reply-input {
                width: 100%;
                padding: 0.5em;
                border: 1px solid #ddd;
                border-radius: 3px;
                font-family: inherit;
                font-size: 0.95em;
                resize: vertical;
            }

            .comment-submit, .reply-submit {
                margin-top: 0.5em;
                padding: 0.5em 1em;
                background: #0066cc;
                color: white;
                border: none;
                border-radius: 3px;
                cursor: pointer;
                font-size: 0.95em;
            }

            .comment-submit:hover, .reply-submit:hover {
                background: #0052a3;
            }

            .no-comments {
                color: #999;
                font-style: italic;
            }

            .orphaned-comments-section {
                margin-top: 3em;
                padding-top: 2em;
                border-top: 2px solid #ddd;
            }

            .orphaned-comments-note {
                color: #666;
                font-style: italic;
                margin-bottom: 1.5em;
            }

            .orphaned-comment {
                margin-bottom: 2em;
                padding: 1em;
                background: #fff9e6;
                border-left: 3px solid #ffcc00;
                border-radius: 4px;
            }

            .orphaned-comment-context {
                margin-bottom: 1em;
                padding-bottom: 1em;
                border-bottom: 1px solid #ddd;
            }

            .orphaned-comment-context blockquote {
                margin: 0.5em 0;
                padding: 0.5em;
                background: white;
                border-left: 3px solid #ddd;
                font-style: italic;
            }

            .orphaned-comment-location {
                font-size: 0.9em;
                color: #666;
                margin-top: 0.5em;
            }
        `;

  document.head.appendChild(style);
}

/**
 * Public API exposed on window for backwards compatibility
 */
export interface MdbookCommentsAPI {
  init: typeof init;
}

declare global {
  interface Window {
    MdbookComments: MdbookCommentsAPI;
    toggleComments: typeof toggleComments;
    submitComment: typeof submitComment;
    submitReply: typeof submitReply;
    showReplyForm: typeof showReplyForm;
  }
}

// Export public API
window.MdbookComments = {
  init: init,
};

// Export global functions for backwards compatibility
window.toggleComments = toggleComments;
window.submitComment = submitComment;
window.submitReply = submitReply;
window.showReplyForm = showReplyForm;
