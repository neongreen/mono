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
 * Raw comment data as received from custom backend API
 * Flexible format that normalizes various naming conventions
 */
interface CustomBackendComment {
  id: string;
  'paragraph-id'?: string;
  paragraph_id?: string;
  paragraphId?: string;
  metadata: ParagraphMetadata;
  author: string;
  text: string;
  created: string;
  edited_at?: string | null;
  editedAt?: string | null;
  'parent-id'?: string | null;
  parent_id?: string | null;
  parentId?: string | null;
  replies?: CustomBackendComment[];
}

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

    const comments = (await response.json()) as CustomBackendComment[];

    // Normalize any naming variations to internal format
    return comments.map((comment: CustomBackendComment): Comment => {
      const normalizeComment = (c: CustomBackendComment): Comment => ({
        id: c.id,
        paragraph_id: c['paragraph-id'] || c.paragraph_id || c.paragraphId || '',
        metadata: c.metadata,
        author: c.author,
        text: c.text,
        created: c.created,
        edited_at: c.edited_at || c.editedAt,
        parent_id: c['parent-id'] || c.parent_id || c.parentId || null,
        replies: c.replies?.map(normalizeComment),
      });
      return normalizeComment(comment);
    });
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

    const comment = (await response.json()) as CustomBackendComment;

    // Normalize response
    const normalizeComment = (c: CustomBackendComment): Comment => ({
      id: c.id,
      paragraph_id: c['paragraph-id'] || c.paragraph_id || c.paragraphId || '',
      metadata: c.metadata,
      author: c.author,
      text: c.text,
      created: c.created,
      parent_id: c['parent-id'] || c.parent_id || c.parentId || null,
      replies: c.replies?.map(normalizeComment),
    });
    
    return normalizeComment(comment);
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

    const reply = (await response.json()) as CustomBackendComment;

    // Normalize response
    const normalizeComment = (c: CustomBackendComment): Comment => ({
      id: c.id,
      paragraph_id: c['paragraph-id'] || c.paragraph_id || c.paragraphId || '',
      metadata: c.metadata,
      author: c.author,
      text: c.text,
      created: c.created,
      edited_at: c.edited_at || c.editedAt,
      parent_id: c['parent-id'] || c.parent_id || c.parentId || null,
      replies: c.replies?.map(normalizeComment),
    });
    
    return normalizeComment(reply);
  },

  /**
   * Update an existing comment's text
   */
  async updateComment(commentId: string, newText: string): Promise<Comment> {
    const response = await fetch(`${config.url}/comments/${commentId}`, {
      method: 'PATCH',
      headers: {
        'Content-Type': 'application/json',
      },
      credentials: 'include',
      body: JSON.stringify({
        text: newText,
        edited_at: new Date().toISOString(),
      }),
    });

    if (!response.ok) {
      throw new Error('Failed to update comment');
    }

    const comment = (await response.json()) as CustomBackendComment;

    // Normalize response
    const normalizeComment = (c: CustomBackendComment): Comment => ({
      id: c.id,
      paragraph_id: c['paragraph-id'] || c.paragraph_id || c.paragraphId || '',
      metadata: c.metadata,
      author: c.author,
      text: c.text,
      created: c.created,
      edited_at: c.edited_at || c.editedAt,
      parent_id: c['parent-id'] || c.parent_id || c.parentId || null,
      replies: c.replies?.map(normalizeComment),
    });
    
    return normalizeComment(comment);
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
