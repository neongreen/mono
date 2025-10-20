/**
 * mdbook-comments - Custom backend adapter
 *
 * This adapter provides a generic REST API implementation that can work
 * with any backend following the standard API contract.
 *
 * Configuration via window.MDBOOK_COMMENTS_CONFIG:
 * {
 *   apiUrl: 'http://your-api.com/comments',
 *   similarityThreshold: 0.85,       // optional
 *   orphanedLocation: 'end-of-chapter'  // optional
 * }
 */

import type { BackendAdapter, Comment, ParagraphMetadata } from '../types';

/**
 * Custom backend configuration
 */
interface CustomBackendConfig {
  apiUrl: string;
  similarityThreshold?: number;
  orphanedLocation?: string;
}

// Configuration from window or defaults
declare global {
  interface Window {
    MDBOOK_COMMENTS_CONFIG?: CustomBackendConfig;
  }
}

const CONFIG: CustomBackendConfig = window.MDBOOK_COMMENTS_CONFIG || {
  apiUrl: 'http://localhost:3000/api',
  similarityThreshold: 0.85,
  orphanedLocation: 'end-of-chapter',
};

const API_URL = CONFIG.apiUrl;

/**
 * Custom backend adapter implementation
 */
const customBackend: BackendAdapter = {
  /**
   * Initialize the backend
   */
  async init(): Promise<void> {
    console.log('Initializing custom backend...');
    console.log('API URL:', API_URL);
  },

  /**
   * Load all comments from the server
   */
  async loadComments(): Promise<Comment[]> {
    const response = await fetch(`${API_URL}/comments`, {
      credentials: 'include', // Include cookies for authentication
    });

    if (!response.ok) {
      throw new Error(`Failed to load comments: ${response.statusText}`);
    }

    const comments = (await response.json()) as unknown[];

    // Normalize any naming variations to internal format
    return comments.map((comment: any) => ({
      ...comment,
      paragraph_id: comment['paragraph-id'] || comment.paragraph_id,
      parent_id: comment['parent-id'] || comment.parent_id,
    })) as Comment[];
  },

  /**
   * Save a new comment
   */
  async saveComment(
    paragraphId: string,
    metadata: ParagraphMetadata,
    text: string,
    author: string
  ): Promise<Comment> {
    const response = await fetch(`${API_URL}/comments`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      credentials: 'include',
      body: JSON.stringify({
        'paragraph-id': paragraphId,
        metadata: metadata,
        text: text,
        author: author || 'Anonymous',
      }),
    });

    if (!response.ok) {
      throw new Error('Failed to post comment');
    }

    const comment = (await response.json()) as any;

    // Normalize response
    return {
      ...comment,
      paragraph_id: comment['paragraph-id'] || comment.paragraph_id,
      parent_id: comment['parent-id'] || comment.parent_id,
    } as Comment;
  },

  /**
   * Save a reply to an existing comment
   */
  async saveReply(
    parentCommentId: string,
    text: string,
    author: string
  ): Promise<Comment> {
    const response = await fetch(
      `${API_URL}/comments/${parentCommentId}/reply`,
      {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        credentials: 'include',
        body: JSON.stringify({
          text: text,
          author: author || 'Anonymous',
        }),
      }
    );

    if (!response.ok) {
      throw new Error('Failed to post reply');
    }

    const reply = (await response.json()) as any;

    // Normalize response
    return {
      ...reply,
      paragraph_id: reply['paragraph-id'] || reply.paragraph_id,
      parent_id: reply['parent-id'] || reply.parent_id,
    } as Comment;
  },

  /**
   * Get current author name (no persistence in generic custom backend)
   */
  getCurrentAuthor(): string | null {
    return null;
  },

  /**
   * Whether to show author input field in forms
   */
  showAuthorInput: true,
};

// Initialize when DOM is ready
if (document.readyState === 'loading') {
  document.addEventListener('DOMContentLoaded', () => {
    if (window.MdbookComments) {
      window.MdbookComments.init(customBackend);
    }
  });
} else {
  if (window.MdbookComments) {
    window.MdbookComments.init(customBackend);
  }
}

// Export for testing
export { customBackend };
