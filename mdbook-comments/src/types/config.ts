/**
 * Configuration types for mdbook-comments
 */

/**
 * Configuration provided via window.MDBOOK_COMMENTS_CONFIG
 */
export interface CommentsConfig {
  /** Similarity threshold for fuzzy matching (0.0 - 1.0) */
  similarityThreshold?: number;
  /** Where to display orphaned comments */
  orphanedLocation?: 'end-of-chapter' | 'end-of-page';
  /** Whether to show comment count on links */
  showCommentCount?: boolean;
  /** Custom link text (default: "comment") */
  linkText?: string;
}

/**
 * Backend-specific configuration
 */
export interface BackendConfig {
  /** Backend type identifier */
  type: 'json-server' | 'supabase' | 'google-sheets' | 'custom';
  /** Backend-specific options */
  [key: string]: unknown;
}

/**
 * JSON Server backend configuration
 */
export interface JsonServerConfig extends BackendConfig {
  type: 'json-server';
  /** API base URL (default: http://localhost:55432) */
  url?: string;
}

/**
 * Supabase backend configuration
 */
export interface SupabaseConfig extends BackendConfig {
  type: 'supabase';
  /** Supabase project URL */
  url: string;
  /** Supabase anon/public key */
  anonKey: string;
}

/**
 * Google Sheets backend configuration
 */
export interface GoogleSheetsConfig extends BackendConfig {
  type: 'google-sheets';
  /** Google OAuth client ID */
  clientId: string;
  /** Spreadsheet ID */
  spreadsheetId: string;
  /** Sheet name (default: "Sheet1") */
  sheetName?: string;
  /** Optional API key for read-only access */
  apiKey?: string;
  /** OAuth scopes */
  scopes?: string;
}

/**
 * Custom backend configuration
 */
export interface CustomBackendConfig extends BackendConfig {
  type: 'custom';
  /** API base URL */
  apiUrl: string;
  /** Additional backend-specific options */
  [key: string]: unknown;
}

/**
 * Default configuration values
 */
export const DEFAULT_CONFIG: Required<CommentsConfig> = {
  similarityThreshold: 0.85,
  orphanedLocation: 'end-of-chapter',
  showCommentCount: true,
  linkText: 'comment',
};
