/**
 * mdbook-comments - JSON Server backend adapter
 *
 * This adapter implements the backend interface for json-server.
 * It provides simple REST API calls to a json-server instance.
 */

import type { BackendAdapter, Comment, ParagraphMetadata } from '../types';

/**
 * JSON Server configuration
 */
interface JsonServerConfig {
  url: string;
}

/**
 * Raw comment data as received from JSON Server API
 * Uses kebab-case field names as stored in json-server
 */
interface JsonServerComment {
  id: string;
  'paragraph-id': string;
  metadata: ParagraphMetadata;
  author: string;
  text: string;
  created: string;
  edited_at?: string | null;
  deleted_at?: string | null;
  reactions?: {
    thumbs_up: number;
    thumbs_down: number;
  };
  resolved_at?: string | null;
  resolved_by?: string | null;
  'parent-id': string | null;
  replies?: JsonServerComment[];
}

// Configuration (injected via window or use defaults)
declare global {
  interface Window {
    JSON_SERVER_CONFIG?: JsonServerConfig;
  }
}

const config: JsonServerConfig = window.JSON_SERVER_CONFIG || {
  url: 'http://localhost:54322',
};

const API_URL = config.url;

// State
let currentAuthor: string | null = null;

/**
 * JSON Server backend adapter implementation
 */
const jsonServerBackend: BackendAdapter = {
  /**
   * Initialize the backend
   */
  async init(): Promise<void> {
    console.log('Initializing json-server backend...');
    console.log('API URL:', API_URL);

    // Load author name from localStorage
    currentAuthor = localStorage.getItem('mdbook-comments-author') || '';
  },

  /**
   * Load all comments from the server
   */
  async loadComments(): Promise<Comment[]> {
    // Add cache-busting parameter to ensure fresh data
    const timestamp = new Date().getTime();
    const response = await fetch(`${API_URL}/comments?_=${timestamp}`, {
      credentials: 'include', // Include cookies for authentication
      cache: 'no-store', // Prevent caching
    });

    if (!response.ok) {
      throw new Error(`Failed to load comments: ${response.statusText}`);
    }

    const comments = (await response.json()) as JsonServerComment[];

    // Normalize kebab-case to snake_case for internal consistency
    return comments.map((comment: JsonServerComment): Comment => {
      const normalized: Comment = {
        id: comment.id,
        paragraph_id: comment['paragraph-id'],
        metadata: comment.metadata,
        author: comment.author,
        text: comment.text,
        created: comment.created,
        reactions: comment.reactions,
        resolved_at: comment.resolved_at,
        resolved_by: comment.resolved_by,
        parent_id: comment['parent-id'],
        replies: comment.replies?.map(reply => ({
          id: reply.id,
          paragraph_id: reply['paragraph-id'],
          metadata: reply.metadata,
          author: reply.author,
          text: reply.text,
          created: reply.created,
          edited_at: reply.edited_at,
          deleted_at: reply.deleted_at,
          reactions: reply.reactions,
          resolved_at: reply.resolved_at,
          resolved_by: reply.resolved_by,
          parent_id: reply['parent-id'],
        })),
      };
      return normalized;
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

    const comment = (await response.json()) as JsonServerComment;
    
    // Normalize the response to internal Comment format
    return {
      id: comment.id,
      paragraph_id: comment['paragraph-id'],
      metadata: comment.metadata,
      author: comment.author,
      text: comment.text,
      created: comment.created,
      parent_id: comment['parent-id'],
      replies: comment.replies?.map(reply => ({
        id: reply.id,
        paragraph_id: reply['paragraph-id'],
        metadata: reply.metadata,
        author: reply.author,
        text: reply.text,
        created: reply.created,
        parent_id: reply['parent-id'],
      })),
    };
  },

  /**
   * Save a reply to an existing comment
   */
  async saveReply(
    parentCommentId: string,
    text: string,
    author: string
  ): Promise<Comment> {
    // Fetch the parent comment to get its paragraph-id and metadata
    const parentResponse = await fetch(`${API_URL}/comments?id=${parentCommentId}`, {
      credentials: 'include',
    });

    if (!parentResponse.ok) {
      throw new Error('Failed to fetch parent comment');
    }

    const parents = (await parentResponse.json()) as JsonServerComment[];
    const parent = parents[0];

    if (!parent) {
      throw new Error('Parent comment not found');
    }

    // Post reply as a regular comment with parent-id set
    const response = await fetch(`${API_URL}/comments`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      credentials: 'include',
      body: JSON.stringify({
        'paragraph-id': parent['paragraph-id'],
        metadata: parent.metadata || {},
        text: text,
        author: author || 'Anonymous',
        'parent-id': parentCommentId,
      }),
    });

    if (!response.ok) {
      throw new Error('Failed to post reply');
    }

    const reply = (await response.json()) as JsonServerComment;
    
    // Normalize the response to internal Comment format
    return {
      id: reply.id,
      paragraph_id: reply['paragraph-id'],
      metadata: reply.metadata,
      author: reply.author,
      text: reply.text,
      created: reply.created,
      reactions: reply.reactions,
      resolved_at: reply.resolved_at,
      resolved_by: reply.resolved_by,
      parent_id: reply['parent-id'],
      replies: reply.replies?.map(r => ({
        id: r.id,
        paragraph_id: r['paragraph-id'],
        metadata: r.metadata,
        author: r.author,
        text: r.text,
        created: r.created,
        parent_id: r['parent-id'],
      })),
    };
  },

  /**
   * Update an existing comment's text
   */
  async updateComment(commentId: string, newText: string): Promise<Comment> {
    const response = await fetch(`${API_URL}/comments/${commentId}`, {
      method: 'PATCH',
      headers: {
        'Content-Type': 'application/json',
      },
      credentials: 'include',
      body: JSON.stringify({
        text: newText,
        'edited_at': new Date().toISOString(),
      }),
    });

    if (!response.ok) {
      throw new Error('Failed to update comment');
    }

    const comment = (await response.json()) as JsonServerComment;
    
    // Normalize the response to internal Comment format
    return {
      id: comment.id,
      paragraph_id: comment['paragraph-id'],
      metadata: comment.metadata,
      author: comment.author,
      text: comment.text,
      created: comment.created,
      edited_at: comment.edited_at,
      deleted_at: comment.deleted_at,
      reactions: comment.reactions,
      parent_id: comment['parent-id'],
      replies: comment.replies?.map(r => ({
        id: r.id,
        paragraph_id: r['paragraph-id'],
        metadata: r.metadata,
        author: r.author,
        text: r.text,
        created: r.created,
        edited_at: r.edited_at,
        deleted_at: r.deleted_at,
        reactions: r.reactions,
        parent_id: r['parent-id'],
      })),
    };
  },

  /**
   * Delete a comment (soft deletion)
   */
  async deleteComment(commentId: string): Promise<Comment> {
    const response = await fetch(`${API_URL}/comments/${commentId}`, {
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

    const comment = (await response.json()) as JsonServerComment;
    
    // Normalize the response to internal Comment format
    return {
      id: comment.id,
      paragraph_id: comment['paragraph-id'],
      metadata: comment.metadata,
      author: comment.author,
      text: comment.text,
      created: comment.created,
      edited_at: comment.edited_at,
      deleted_at: comment.deleted_at,
      reactions: comment.reactions,
      parent_id: comment['parent-id'],
      replies: comment.replies?.map(r => ({
        id: r.id,
        paragraph_id: r['paragraph-id'],
        metadata: r.metadata,
        author: r.author,
        text: r.text,
        created: r.created,
        edited_at: r.edited_at,
        deleted_at: r.deleted_at,
        reactions: r.reactions,
        parent_id: r['parent-id'],
      })),
    };
  },

  /**
   * Add or update a reaction to a comment
   */
  async addReaction(commentId: string, reactionType: 'thumbs_up' | 'thumbs_down'): Promise<Comment> {
    // First, get the current comment to read current reactions
    const getCurrentResponse = await fetch(`${API_URL}/comments/${commentId}`, {
      credentials: 'include',
    });

    if (!getCurrentResponse.ok) {
      throw new Error('Failed to get current comment');
    }

    const currentComment = (await getCurrentResponse.json()) as JsonServerComment;
    const currentReactions = currentComment.reactions || { thumbs_up: 0, thumbs_down: 0 };

    // Get user's current reaction from localStorage to prevent double-voting
    const userReactionsKey = `comment-reactions-${commentId}`;
    const existingReaction = localStorage.getItem(userReactionsKey);

    let newReactions = { ...currentReactions };

    // If user already has this reaction, remove it (toggle off)
    if (existingReaction === reactionType) {
      newReactions[reactionType] = Math.max(0, newReactions[reactionType] - 1);
      localStorage.removeItem(userReactionsKey);
    } else {
      // If user has a different reaction, remove old and add new
      if (existingReaction && existingReaction !== reactionType) {
        const oldType = existingReaction as 'thumbs_up' | 'thumbs_down';
        newReactions[oldType] = Math.max(0, newReactions[oldType] - 1);
      }
      
      // Add new reaction
      newReactions[reactionType] = newReactions[reactionType] + 1;
      localStorage.setItem(userReactionsKey, reactionType);
    }

    // Update the comment with new reaction counts
    const response = await fetch(`${API_URL}/comments/${commentId}`, {
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

    const comment = (await response.json()) as JsonServerComment;
    
    // Normalize the response to internal Comment format
    return {
      id: comment.id,
      paragraph_id: comment['paragraph-id'],
      metadata: comment.metadata,
      author: comment.author,
      text: comment.text,
      created: comment.created,
      edited_at: comment.edited_at,
      deleted_at: comment.deleted_at,
      reactions: comment.reactions,
      parent_id: comment['parent-id'],
      replies: comment.replies?.map(r => ({
        id: r.id,
        paragraph_id: r['paragraph-id'],
        metadata: r.metadata,
        author: r.author,
        text: r.text,
        created: r.created,
        edited_at: r.edited_at,
        deleted_at: r.deleted_at,
        reactions: r.reactions,
        parent_id: r['parent-id'],
      })),
    };
  },

  /**
   * Remove a reaction from a comment
   */
  async removeReaction(commentId: string, reactionType: 'thumbs_up' | 'thumbs_down'): Promise<Comment> {
    // This is handled by addReaction when toggling off, but we'll implement it for completeness
    const getCurrentResponse = await fetch(`${API_URL}/comments/${commentId}`, {
      credentials: 'include',
    });

    if (!getCurrentResponse.ok) {
      throw new Error('Failed to get current comment');
    }

    const currentComment = (await getCurrentResponse.json()) as JsonServerComment;
    const currentReactions = currentComment.reactions || { thumbs_up: 0, thumbs_down: 0 };

    // Remove reaction
    const newReactions = { ...currentReactions };
    newReactions[reactionType] = Math.max(0, newReactions[reactionType] - 1);

    // Remove from localStorage
    const userReactionsKey = `comment-reactions-${commentId}`;
    if (localStorage.getItem(userReactionsKey) === reactionType) {
      localStorage.removeItem(userReactionsKey);
    }

    const response = await fetch(`${API_URL}/comments/${commentId}`, {
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

    const comment = (await response.json()) as JsonServerComment;
    
    return {
      id: comment.id,
      paragraph_id: comment['paragraph-id'],
      metadata: comment.metadata,
      author: comment.author,
      text: comment.text,
      created: comment.created,
      edited_at: comment.edited_at,
      deleted_at: comment.deleted_at,
      reactions: comment.reactions,
      parent_id: comment['parent-id'],
      replies: comment.replies?.map(r => ({
        id: r.id,
        paragraph_id: r['paragraph-id'],
        metadata: r.metadata,
        author: r.author,
        text: r.text,
        created: r.created,
        edited_at: r.edited_at,
        deleted_at: r.deleted_at,
        reactions: r.reactions,
        parent_id: r['parent-id'],
      })),
    };
  },

  /**
   * Mark a comment as resolved
   */
  async resolveComment(commentId: string): Promise<Comment> {
    const currentAuthor = this.getCurrentAuthor() || 'Anonymous';
    
    const response = await fetch(`${API_URL}/comments/${commentId}`, {
      method: 'PATCH',
      headers: {
        'Content-Type': 'application/json',
      },
      credentials: 'include',
      body: JSON.stringify({
        resolved_at: new Date().toISOString(),
        resolved_by: currentAuthor,
      }),
    });

    if (!response.ok) {
      throw new Error('Failed to resolve comment');
    }

    const comment = (await response.json()) as JsonServerComment;
    
    // Normalize the response to internal Comment format
    return {
      id: comment.id,
      paragraph_id: comment['paragraph-id'],
      metadata: comment.metadata,
      author: comment.author,
      text: comment.text,
      created: comment.created,
      edited_at: comment.edited_at,
      deleted_at: comment.deleted_at,
      reactions: comment.reactions,
      resolved_at: comment.resolved_at,
      resolved_by: comment.resolved_by,
      parent_id: comment['parent-id'],
      replies: comment.replies?.map(r => ({
        id: r.id,
        paragraph_id: r['paragraph-id'],
        metadata: r.metadata,
        author: r.author,
        text: r.text,
        created: r.created,
        edited_at: r.edited_at,
        deleted_at: r.deleted_at,
        reactions: r.reactions,
        resolved_at: r.resolved_at,
        resolved_by: r.resolved_by,
        parent_id: r['parent-id'],
      })),
    };
  },

  /**
   * Mark a comment as unresolved
   */
  async unresolveComment(commentId: string): Promise<Comment> {
    const response = await fetch(`${API_URL}/comments/${commentId}`, {
      method: 'PATCH',
      headers: {
        'Content-Type': 'application/json',
      },
      credentials: 'include',
      body: JSON.stringify({
        resolved_at: null,
        resolved_by: null,
      }),
    });

    if (!response.ok) {
      throw new Error('Failed to unresolve comment');
    }

    const comment = (await response.json()) as JsonServerComment;
    
    return {
      id: comment.id,
      paragraph_id: comment['paragraph-id'],
      metadata: comment.metadata,
      author: comment.author,
      text: comment.text,
      created: comment.created,
      edited_at: comment.edited_at,
      deleted_at: comment.deleted_at,
      reactions: comment.reactions,
      resolved_at: comment.resolved_at,
      resolved_by: comment.resolved_by,
      parent_id: comment['parent-id'],
      replies: comment.replies?.map(r => ({
        id: r.id,
        paragraph_id: r['paragraph-id'],
        metadata: r.metadata,
        author: r.author,
        text: r.text,
        created: r.created,
        edited_at: r.edited_at,
        deleted_at: r.deleted_at,
        reactions: r.reactions,
        resolved_at: r.resolved_at,
        resolved_by: r.resolved_by,
        parent_id: r['parent-id'],
      })),
    };
  },

  /**
   * Get current author name
   */
  getCurrentAuthor(): string | null {
    return currentAuthor;
  },

  /**
   * Set current author name
   */
  setCurrentAuthor(author: string): void {
    currentAuthor = author;
    localStorage.setItem('mdbook-comments-author', author);
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
      window.MdbookComments.init(jsonServerBackend);
    }
  });
} else {
  if (window.MdbookComments) {
    window.MdbookComments.init(jsonServerBackend);
  }
}

// Export for testing
export { jsonServerBackend };
