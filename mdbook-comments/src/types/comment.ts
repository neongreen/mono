/**
 * Core comment data types
 */

/**
 * Position information for a commentable paragraph
 */
export interface ParagraphPosition {
  /** Source file path */
  file: string;
  /** Block index within the file */
  'block-index': number;
  /** Section index within current heading context */
  'section-index': number;
}

/**
 * Context information for a commentable paragraph
 */
export interface ParagraphContext {
  /** Content of previous paragraph (for fuzzy matching) */
  prev?: string;
  /** Content of next paragraph (for fuzzy matching) */
  next?: string;
  /** Hierarchical heading path (e.g., ["Chapter 1", "Introduction"]) */
  'heading-path': string[];
}

/**
 * Complete metadata for a commentable paragraph
 */
export interface ParagraphMetadata {
  /** Unique identifier for this paragraph */
  id: string;
  /** Position within the document */
  position: ParagraphPosition;
  /** Full content of the paragraph */
  content: string;
  /** Surrounding context for fuzzy matching */
  context: ParagraphContext;
  /** Git commit hash (if available) */
  commit?: string;
}

/**
 * A comment or reply
 */
export interface Comment {
  /** Unique comment ID (UUID or backend-generated) */
  id: string;
  /** ID of the paragraph this comment is attached to */
  paragraph_id: string;
  /** Metadata of the paragraph (stored with comment for orphaned comment display) */
  metadata: ParagraphMetadata;
  /** Author name or identifier */
  author: string;
  /** Comment text content */
  text: string;
  /** ISO 8601 timestamp of creation */
  created: string;
  /** Parent comment ID (null for top-level comments) */
  parent_id: string | null;
  /** Nested replies to this comment */
  replies?: Comment[];
}

/**
 * Comment with fuzzy matching metadata
 */
export interface MatchedComment {
  /** The paragraph ID this comment was matched to */
  paragraphId: string;
  /** The comment data */
  comment: Comment;
  /** Similarity confidence (0.0 - 1.0) */
  confidence: number;
}
