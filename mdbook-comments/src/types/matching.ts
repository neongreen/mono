/**
 * Fuzzy matching types and interfaces
 */

import type { Comment, ParagraphMetadata } from './comment';

/**
 * Similarity calculation weights
 */
export interface SimilarityWeights {
  /** Weight for content similarity (typically 0.5) */
  content: number;
  /** Weight for previous paragraph similarity (typically 0.2) */
  prevContext: number;
  /** Weight for next paragraph similarity (typically 0.2) */
  nextContext: number;
  /** Weight for heading path similarity (typically 0.1) */
  headingPath: number;
}

/**
 * Default similarity weights
 */
export const DEFAULT_WEIGHTS: SimilarityWeights = {
  content: 0.5,
  prevContext: 0.2,
  nextContext: 0.2,
  headingPath: 0.1,
};

/**
 * Result of similarity calculation
 */
export interface SimilarityResult {
  /** Overall similarity score (0.0 - 1.0) */
  score: number;
  /** Individual component scores */
  components: {
    content?: number;
    prevContext?: number;
    nextContext?: number;
    headingPath?: number;
  };
}

/**
 * Matching strategy
 */
export type MatchingStrategy = 'exact' | 'fuzzy';

/**
 * Matching result
 */
export interface MatchingResult {
  /** The paragraph ID that was matched */
  paragraphId: string;
  /** The comment that was matched */
  comment: Comment;
  /** Matching strategy used */
  strategy: MatchingStrategy;
  /** Similarity score (1.0 for exact matches) */
  similarity: number;
}

/**
 * Paragraph link data (extracted from DOM)
 */
export interface ParagraphLink {
  /** Paragraph ID */
  id: string;
  /** Paragraph metadata */
  metadata: ParagraphMetadata;
  /** DOM element for the comment link */
  element: HTMLElement;
}
