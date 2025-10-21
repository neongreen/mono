/**
 * Backend adapter interface and types
 */

import type { Comment, ParagraphMetadata } from './comment';

/**
 * Backend adapter interface
 *
 * All backend implementations (json-server, Supabase, Google Sheets, custom)
 * must implement this interface to integrate with the comment system.
 */
export interface BackendAdapter {
  /**
   * Initialize the backend connection
   *
   * This is called once when the page loads. Use it to:
   * - Initialize clients (Supabase, Google Sheets API, etc.)
   * - Check authentication status
   * - Load initial configuration
   *
   * @returns Promise that resolves when initialization is complete
   */
  init(): Promise<void>;

  /**
   * Load all comments from the backend
   *
   * @returns Promise resolving to array of comments (empty array if none)
   * @throws Error if loading fails
   */
  loadComments(): Promise<Comment[]>;

  /**
   * Save a new top-level comment
   *
   * @param paragraphId - ID of the paragraph being commented on
   * @param metadata - Full paragraph metadata
   * @param text - Comment text
   * @param author - Author name/identifier
   * @returns Promise resolving to the created comment
   * @throws Error if save fails
   */
  saveComment(
    paragraphId: string,
    metadata: ParagraphMetadata,
    text: string,
    author: string
  ): Promise<Comment>;

  /**
   * Save a reply to an existing comment
   *
   * @param parentCommentId - ID of the parent comment
   * @param text - Reply text
   * @param author - Author name/identifier
   * @returns Promise resolving to the created reply
   * @throws Error if save fails
   */
  saveReply(
    parentCommentId: string,
    text: string,
    author: string
  ): Promise<Comment>;

  /**
   * Update an existing comment's text
   */
  updateComment(commentId: string, newText: string): Promise<Comment>;

  /**
   * Delete a comment (soft deletion)
   */
  deleteComment(commentId: string): Promise<Comment>;

  /**
   * Add or update a reaction to a comment
   */
  addReaction(commentId: string, reactionType: 'thumbs_up' | 'thumbs_down'): Promise<Comment>;

  /**
   * Remove a reaction from a comment
   */
  removeReaction(commentId: string, reactionType: 'thumbs_up' | 'thumbs_down'): Promise<Comment>;

  /**
   * Mark a comment as resolved
   */
  resolveComment(commentId: string): Promise<Comment>;

  /**
   * Mark a comment as unresolved
   */
  unresolveComment(commentId: string): Promise<Comment>;

  /**
   * Get the current author name
   *
   * This is used to pre-fill the author field in forms.
   * Can return null if no author is set.
   *
   * @returns Current author name or null
   */
  getCurrentAuthor(): string | null;

  /**
   * Whether to show the author input field in comment forms
   *
   * - true: Show author input (json-server, custom)
   * - false: Hide author input (authenticated backends like Supabase, Google Sheets)
   */
  showAuthorInput: boolean;

  /**
   * Optional: Set the current author name
   *
   * Used by backends that manage author in localStorage
   */
  setCurrentAuthor?(author: string): void;

  /**
   * Optional: Check if user is authenticated
   *
   * Used by backends with authentication (Supabase, Google Sheets)
   */
  isAuthenticated?(): boolean;

  /**
   * Optional: Trigger sign-in flow
   *
   * Used by backends with authentication
   */
  signIn?(): void;

  /**
   * Optional: Sign out
   *
   * Used by backends with authentication
   */
  signOut?(): void;

  /**
   * Optional: Register callback for authentication state changes
   *
   * @param callback - Function to call when auth state changes
   */
  onAuthChange?(callback: (user: unknown) => void): void;
}

/**
 * Backend initialization error
 */
export class BackendInitError extends Error {
  constructor(message: string, public cause?: Error) {
    super(message);
    this.name = 'BackendInitError';
  }
}

/**
 * Backend operation error
 */
export class BackendOperationError extends Error {
  constructor(message: string, public cause?: Error) {
    super(message);
    this.name = 'BackendOperationError';
  }
}
