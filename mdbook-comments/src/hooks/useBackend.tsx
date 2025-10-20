/**
 * useBackend hook - Provides access to backend adapter via context
 */

import { createContext } from 'preact';
import { useContext } from 'preact/hooks';
import type { BackendAdapter } from '../types';

/**
 * Context for backend adapter
 */
export const BackendContext = createContext<BackendAdapter | null>(null);

/**
 * Hook to access backend adapter from context
 * Throws if used outside BackendContext.Provider
 */
export function useBackend(): BackendAdapter {
  const backend = useContext(BackendContext);

  if (!backend) {
    throw new Error('useBackend must be used within BackendContext.Provider');
  }

  return backend;
}
