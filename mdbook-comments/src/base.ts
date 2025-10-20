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
  const commentsById: Record<string, Comment> = {};

  // First pass: index all comments and initialize replies array
  state.allComments.forEach((comment) => {
    if (!comment.replies) {
      comment.replies = [];
    }
    commentsById[comment.id] = comment;
  });

  // Second pass: build reply tree
  state.allComments.forEach((comment) => {
    if (comment.parent_id && commentsById[comment.parent_id]) {
      const parent = commentsById[comment.parent_id];
      if (parent && parent.replies) {
        parent.replies.push(comment);
      }
    }
  });
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

  const section = document.createElement('div');
  section.className = 'orphaned-comments-section';
  section.innerHTML = `
            <h2>Unmapped Comments</h2>
            <p class="orphaned-comments-note">
                The following comments could not be matched to any current paragraph.
                They may refer to content that has been removed or significantly changed.
            </p>
            <div class="orphaned-comments-list"></div>
        `;

  const list = section.querySelector('.orphaned-comments-list');
  if (!list) return;

  state.orphanedComments.forEach((comment) => {
    const item = createOrphanedCommentElement(comment);
    list.appendChild(item);
  });

  main.appendChild(section);
}

/**
 * Create element for orphaned comment
 */
function createOrphanedCommentElement(comment: Comment): HTMLElement {
  const div = document.createElement('div');
  div.className = 'orphaned-comment';

  const meta = comment.metadata || ({} as ParagraphMetadata);
  const context = meta.context || { 'heading-path': [] };
  const content = meta.content || '[Content not available]';

  div.innerHTML = `
            <div class="orphaned-comment-context">
                <strong>Original paragraph:</strong>
                <blockquote>${escapeHtml(content)}</blockquote>
                ${
                  context['heading-path'] && context['heading-path'].length > 0
                    ? `
                    <div class="orphaned-comment-location">
                        Section: ${context['heading-path'].join(' > ')}
                    </div>
                `
                    : ''
                }
            </div>
            <div class="comment-item">
                <div class="comment-header">
                    <span class="comment-author">${escapeHtml(
                      comment.author || 'Anonymous'
                    )}</span>
                    <span class="comment-date">${formatDate(
                      comment.created
                    )}</span>
                </div>
                <div class="comment-text">${escapeHtml(comment.text)}</div>
                ${
                  comment.replies && comment.replies.length > 0
                    ? `
                    <div class="comment-replies">
                        ${comment.replies.map((r) => createReplyHtml(r)).join('')}
                    </div>
                `
                    : ''
                }
            </div>
        `;

  return div;
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
 * Create comment section HTML
 */
function createCommentSection(paragraphId: string): HTMLElement {
  const section = document.createElement('div');
  section.id = `comments-${paragraphId}`;
  section.className = 'comment-section';
  section.setAttribute('data-paragraph-id', paragraphId);

  const comments = state.currentPageComments
    .filter((c) => c.paragraphId === paragraphId)
    .map((c) => c.comment);

  const commentList = document.createElement('div');
  commentList.className = 'comment-list';

  if (comments.length > 0) {
    comments.forEach((comment) => {
      const commentElement = createCommentElement(comment);
      commentList.appendChild(commentElement);
    });
  } else {
    commentList.innerHTML =
      '<p class="no-comments">No comments yet. Be the first to comment!</p>';
  }

  section.appendChild(commentList);

  // Create comment form
  const form = createCommentForm(paragraphId);
  section.appendChild(form);

  return section;
}

/**
 * Create comment form
 */
function createCommentForm(
  paragraphId: string | null,
  parentCommentId: string | null = null
): HTMLElement {
  if (!state.backend) {
    throw new Error('Backend not initialized');
  }

  const form = document.createElement('div');
  form.className = parentCommentId ? 'reply-form' : 'comment-form';

  const isReply = !!parentCommentId;
  const currentAuthor = state.backend.getCurrentAuthor
    ? state.backend.getCurrentAuthor()
    : '';

  // Show author input if backend requires it
  if (state.backend.showAuthorInput && !currentAuthor) {
    const authorInput = document.createElement('input');
    authorInput.type = 'text';
    authorInput.className = 'author-input';
    authorInput.name = 'author';
    authorInput.placeholder = 'Your name';
    authorInput.value = currentAuthor || '';

    // Save author to backend on input (for localStorage-based backends)
    authorInput.addEventListener('input', (e) => {
      const target = e.target as HTMLInputElement;
      const value = target.value.trim();
      if (value && state.backend && state.backend.setCurrentAuthor) {
        state.backend.setCurrentAuthor(value);
      }
    });

    form.appendChild(authorInput);
  }

  const textarea = document.createElement('textarea');
  textarea.className = isReply ? 'reply-input' : 'comment-input';
  textarea.name = isReply ? 'reply-text' : 'comment-text';
  textarea.placeholder = isReply ? 'Write a reply...' : 'Add a comment...';
  textarea.rows = isReply ? 2 : 3;
  form.appendChild(textarea);

  const submitButton = document.createElement('button');
  submitButton.type = 'submit';
  submitButton.className = isReply ? 'reply-submit' : 'comment-submit';
  submitButton.textContent = isReply ? 'Post Reply' : 'Submit';
  submitButton.onclick = (e) => {
    e.preventDefault();
    if (isReply && parentCommentId) {
      submitReply(parentCommentId);
    } else if (paragraphId) {
      submitComment(paragraphId);
    }
  };
  form.appendChild(submitButton);

  return form;
}

/**
 * Create DOM element for a single comment
 */
function createCommentElement(comment: Comment): HTMLElement {
  const div = document.createElement('div');
  div.className = 'comment-item';
  div.setAttribute('data-comment-id', comment.id);

  div.innerHTML = `
            <div class="comment-header">
                <span class="comment-author">${escapeHtml(
                  comment.author || 'Anonymous'
                )}</span>
                <span class="comment-date">${formatDate(comment.created)}</span>
            </div>
            <div class="comment-text">${escapeHtml(comment.text)}</div>
        `;

  // Add replies if present
  if (comment.replies && comment.replies.length > 0) {
    const repliesDiv = document.createElement('div');
    repliesDiv.className = 'comment-replies';
    comment.replies.forEach((reply) => {
      repliesDiv.innerHTML += createReplyHtml(reply);
    });
    div.appendChild(repliesDiv);
  }

  // Add reply button
  const replyBtn = document.createElement('button');
  replyBtn.className = 'comment-reply-btn';
  replyBtn.textContent = 'Reply';
  replyBtn.onclick = () => showReplyForm(comment.id);
  div.appendChild(replyBtn);

  // Add hidden reply form
  const replyForm = createCommentForm(null, comment.id);
  replyForm.id = `reply-form-${comment.id}`;
  replyForm.style.display = 'none';
  div.appendChild(replyForm);

  return div;
}

/**
 * Create HTML for a reply
 */
function createReplyHtml(reply: Comment): string {
  return `
            <div class="reply-item">
                <div class="reply-header">
                    <span class="reply-author">${escapeHtml(
                      reply.author || 'Anonymous'
                    )}</span>
                    <span class="reply-date">${formatDate(reply.created)}</span>
                </div>
                <div class="reply-text">${escapeHtml(reply.text)}</div>
            </div>
        `;
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

    // Reload comments display
    section.remove();
    toggleComments(paragraphId);
  } catch (error) {
    console.error('Error posting comment:', error);
    alert('Failed to post comment. Please try again.');
  }
}

/**
 * Submit a reply to a comment
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

  // Find comment section to get paragraph ID
  const commentSection = form.closest('.comment-section') as HTMLElement | null;
  const paragraphId = commentSection
    ? commentSection.getAttribute('data-paragraph-id')
    : null;

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

    // Reload comments
    if (commentSection && paragraphId) {
      commentSection.remove();
      toggleComments(paragraphId);
    }
  } catch (error) {
    console.error('Error posting reply:', error);
    alert('Failed to post reply. Please try again.');
  }
}

/**
 * Refresh all open comment sections (e.g., after auth change)
 */
function refreshAllCommentSections(): void {
  document.querySelectorAll('.comment-section').forEach((section) => {
    const paragraphId = section.getAttribute('data-paragraph-id');
    const htmlSection = section as HTMLElement;
    if (paragraphId && htmlSection.style.display !== 'none') {
      section.remove();
      toggleComments(paragraphId);
    }
  });
}

/**
 * Escape HTML to prevent XSS
 */
function escapeHtml(text: string): string {
  const div = document.createElement('div');
  div.textContent = text;
  return div.innerHTML;
}

/**
 * Format date for display
 */
function formatDate(dateStr: string): string {
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
