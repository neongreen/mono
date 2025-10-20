/**
 * mdbook-comments - TypeScript entry point
 *
 * This is the main entry point that exports the base module functionality.
 * The base module provides all shared functionality for comment rendering
 * and management across different backend implementations.
 */

// Re-export everything from base module
export * from './base';

// Re-export types for external use
export type * from './types';
