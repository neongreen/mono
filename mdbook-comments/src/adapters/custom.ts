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
  deleted_at?: string | null;
  deletedAt?: string | null;
  reactions?: {
    thumbs_up: number;
    thumbs_down: number;
  };
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
        deleted_at: c.deleted_at || c.deletedAt,
        reactions: c.reactions,
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
      edited_at: c.edited_at || c.editedAt,
      deleted_at: c.deleted_at || c.deletedAt,
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
      deleted_at: c.deleted_at || c.deletedAt,
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
      deleted_at: c.deleted_at || c.deletedAt,
      parent_id: c['parent-id'] || c.parent_id || c.parentId || null,
      replies: c.replies?.map(normalizeComment),
    });
    
    return normalizeComment(comment);
  },

  /**
   * Delete a comment (soft deletion)
   */
  async deleteComment(commentId: string): Promise<Comment> {
    const response = await fetch(`${config.url}/comments/${commentId}`, {
      method: 'PATCH',
      headers: {
        'Content-Type': 'application/json',
      },
      credentials: 'include',
      body: JSON.stringify({
        deleted_at: new Date().toISOString(),
      }),
    });

    if (!response.ok) {
      throw new Error('Failed to delete comment');
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
      deleted_at: c.deleted_at || c.deletedAt,
      parent_id: c['parent-id'] || c.parent_id || c.parentId || null,
      replies: c.replies?.map(normalizeComment),
    });
    
    return normalizeComment(comment);
  },

  /**
   * Add or update a reaction to a comment
   */
  async addReaction(commentId: string, reactionType: 'thumbs_up' | 'thumbs_down'): Promise<Comment> {
    // Get current comment first
    const getCurrentResponse = await fetch(`${config.url}/comments/${commentId}`, {
      credentials: 'include',
    });

    if (!getCurrentResponse.ok) {
      throw new Error('Failed to get current comment');
    }

    const currentComment = (await getCurrentResponse.json()) as CustomBackendComment;
    const currentReactions = currentComment.reactions || { thumbs_up: 0, thumbs_down: 0 };

    // Handle user reaction tracking
    const userReactionsKey = `comment-reactions-${commentId}`;
    const existingReaction = localStorage.getItem(userReactionsKey);

    let newReactions = { ...currentReactions };

    if (existingReaction === reactionType) {
      // Toggle off
      newReactions[reactionType] = Math.max(0, newReactions[reactionType] - 1);
      localStorage.removeItem(userReactionsKey);
    } else {
      // Remove old reaction if exists
      if (existingReaction && existingReaction !== reactionType) {
        const oldType = existingReaction as 'thumbs_up' | 'thumbs_down';
        newReactions[oldType] = Math.max(0, newReactions[oldType] - 1);
      }
      
      // Add new reaction
      newReactions[reactionType] = newReactions[reactionType] + 1;
      localStorage.setItem(userReactionsKey, reactionType);
    }

    const response = await fetch(`${config.url}/comments/${commentId}`, {
      method: 'PATCH',
      headers: {
        'Content-Type': 'application/json',
      },
      credentials: 'include',
      body: JSON.stringify({
        reactions: newReactions,
      }),
    });

    if (!response.ok) {
      throw new Error('Failed to update reaction');
    }

    const comment = (await response.json()) as CustomBackendComment;

    const normalizeComment = (c: CustomBackendComment): Comment => ({
      id: c.id,
      paragraph_id: c['paragraph-id'] || c.paragraph_id || c.paragraphId || '',
      metadata: c.metadata,
      author: c.author,
      text: c.text,
      created: c.created,
      edited_at: c.edited_at || c.editedAt,
      deleted_at: c.deleted_at || c.deletedAt,
      reactions: c.reactions,
      parent_id: c['parent-id'] || c.parent_id || c.parentId || null,
      replies: c.replies?.map(normalizeComment),
    });
    
    return normalizeComment(comment);
  },

  /**
   * Remove a reaction from a comment
   */
  async removeReaction(commentId: string, reactionType: 'thumbs_up' | 'thumbs_down'): Promise<Comment> {
    // Get current comment first
    const getCurrentResponse = await fetch(`${config.url}/comments/${commentId}`, {
      credentials: 'include',
    });

    if (!getCurrentResponse.ok) {
      throw new Error('Failed to get current comment');
    }

    const currentComment = (await getCurrentResponse.json()) as CustomBackendComment;
    const currentReactions = currentComment.reactions || { thumbs_up: 0, thumbs_down: 0 };

    const newReactions = { ...currentReactions };
    newReactions[reactionType] = Math.max(0, newReactions[reactionType] - 1);

    // Remove from localStorage
    const userReactionsKey = `comment-reactions-${commentId}`;
    if (localStorage.getItem(userReactionsKey) === reactionType) {
      localStorage.removeItem(userReactionsKey);
    }

    const response = await fetch(`${config.url}/comments/${commentId}`, {
      method: 'PATCH',
      headers: {
        'Content-Type': 'application/json',
      },
      credentials: 'include',
      body: JSON.stringify({
        reactions: newReactions,
      }),
    });

    if (!response.ok) {
      throw new Error('Failed to remove reaction');
    }

    const comment = (await response.json()) as CustomBackendComment;

    const normalizeComment = (c: CustomBackendComment): Comment => ({
      id: c.id,
      paragraph_id: c['paragraph-id'] || c.paragraph_id || c.paragraphId || '',
      metadata: c.metadata,
      author: c.author,
      text: c.text,
      created: c.created,
      edited_at: c.edited_at || c.editedAt,
      deleted_at: c.deleted_at || c.deletedAt,
      reactions: c.reactions,
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
