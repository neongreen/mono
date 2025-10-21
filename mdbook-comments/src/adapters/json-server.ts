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
