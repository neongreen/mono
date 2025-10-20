/**
 * Type definitions for mdbook-comments
 *
 * This module exports all type definitions used throughout the application.
 */

// Comment types
export type {
  Comment,
  ParagraphMetadata,
  ParagraphPosition,
  ParagraphContext,
  MatchedComment,
} from './comment';

// Backend types
export type {
  BackendAdapter,
} from './backend';

export {
  BackendInitError,
  BackendOperationError,
} from './backend';

// Configuration types
export type {
  CommentsConfig,
  BackendConfig,
  JsonServerConfig,
  SupabaseConfig,
  GoogleSheetsConfig,
  CustomBackendConfig,
} from './config';

export {
  DEFAULT_CONFIG,
} from './config';

// Matching types
export type {
  SimilarityWeights,
  SimilarityResult,
  MatchingStrategy,
  MatchingResult,
  ParagraphLink,
} from './matching';

export {
  DEFAULT_WEIGHTS,
} from './matching';
