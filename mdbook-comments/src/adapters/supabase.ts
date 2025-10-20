/**
 * mdbook-comments - Supabase backend adapter
 *
 * This adapter implements the backend interface for Supabase.
 * It uses the Supabase JavaScript client for database operations.
 *
 * Configuration via window.SUPABASE_CONFIG:
 * {
 *   url: 'https://xxxxx.supabase.co',
 *   anonKey: 'your-anon-key'
 * }
 */

import type { BackendAdapter, Comment, ParagraphMetadata } from '../types';

/**
 * Supabase configuration
 */
interface SupabaseConfig {
  url: string;
  anonKey: string;
}

/**
 * Supabase client interface (minimal typing for external SDK)
 */
interface SupabaseClient {
  from(table: string): {
    select(columns: string): any;
    insert(data: any): any;
  };
}

/**
 * Global Supabase SDK
 */
declare global {
  interface Window {
    SUPABASE_CONFIG?: SupabaseConfig;
    supabase?: {
      createClient(url: string, key: string): SupabaseClient;
    };
  }
}

// Configuration (injected or defaults)
const supabaseConfig: SupabaseConfig = window.SUPABASE_CONFIG || {
  url: 'YOUR_SUPABASE_PROJECT_URL',
  anonKey: 'YOUR_SUPABASE_ANON_KEY',
};

const SUPABASE_URL = supabaseConfig.url;
const SUPABASE_ANON_KEY = supabaseConfig.anonKey;

// State
let supabaseClient: SupabaseClient | null = null;
let currentAuthor: string | null = null;

/**
 * Supabase backend adapter implementation
 */
const supabaseBackend: BackendAdapter = {
  /**
   * Initialize the backend
   */
  async init(): Promise<void> {
    console.log('Initializing Supabase backend...');
    console.log('Supabase URL:', SUPABASE_URL);

    // Check if Supabase SDK is loaded
    if (!window.supabase || !window.supabase.createClient) {
      console.error('Supabase SDK not loaded. Add the SDK to your book.toml');
      throw new Error('Supabase SDK not available');
    }

    // Initialize Supabase client
    supabaseClient = window.supabase.createClient(
      SUPABASE_URL,
      SUPABASE_ANON_KEY
    );

    // Load author name from localStorage
    currentAuthor = localStorage.getItem('mdbook-comments-author') || '';
  },

  /**
   * Load all comments from the server
   */
  async loadComments(): Promise<Comment[]> {
    if (!supabaseClient) {
      throw new Error('Supabase client not initialized');
    }

    const { data, error } = await supabaseClient
      .from('comments')
      .select('*')
      .order('created', { ascending: true });

    if (error) {
      throw new Error(`Failed to load comments: ${error.message}`);
    }

    return (data as Comment[]) || [];
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
    if (!supabaseClient) {
      throw new Error('Supabase client not initialized');
    }

    const { data, error } = await supabaseClient
      .from('comments')
      .insert({
        paragraph_id: paragraphId,
        metadata: metadata,
        author: author || 'Anonymous',
        text: text,
        parent_id: null,
      })
      .select()
      .single();

    if (error) {
      throw new Error('Failed to post comment');
    }

    return data as Comment;
  },

  /**
   * Save a reply to an existing comment
   */
  async saveReply(
    parentCommentId: string,
    text: string,
    author: string
  ): Promise<Comment> {
    if (!supabaseClient) {
      throw new Error('Supabase client not initialized');
    }

    // We need to find the parent comment to get its paragraph_id and metadata
    const { data: parentComment, error: fetchError } = await supabaseClient
      .from('comments')
      .select('paragraph_id, metadata')
      .eq('id', parentCommentId)
      .single();

    if (fetchError) {
      throw new Error('Failed to find parent comment');
    }

    const { data, error } = await supabaseClient
      .from('comments')
      .insert({
        paragraph_id: (parentComment as any).paragraph_id,
        metadata: (parentComment as any).metadata,
        author: author || 'Anonymous',
        text: text,
        parent_id: parentCommentId,
      })
      .select()
      .single();

    if (error) {
      throw new Error('Failed to post reply');
    }

    return data as Comment;
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
      window.MdbookComments.init(supabaseBackend);
    }
  });
} else {
  if (window.MdbookComments) {
    window.MdbookComments.init(supabaseBackend);
  }
}

// Export for testing
export { supabaseBackend };
