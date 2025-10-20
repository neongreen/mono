/**
 * useAuth hook - Manages authentication state
 */

import { useState, useEffect } from 'preact/hooks';
import type { BackendAdapter } from '../types';

interface UseAuthResult {
  isAuthenticated: boolean;
  currentAuthor: string | null;
  signIn?: () => void;
  signOut?: () => void;
}

/**
 * Hook for managing authentication state
 * Works with backends that support authentication
 */
export function useAuth(backend: BackendAdapter): UseAuthResult {
  const [isAuthenticated, setIsAuthenticated] = useState(
    backend.isAuthenticated ? backend.isAuthenticated() : false
  );
  const [currentAuthor, setCurrentAuthor] = useState(
    backend.getCurrentAuthor?.() || null
  );

  // Update state when backend changes
  useEffect(() => {
    setIsAuthenticated(
      backend.isAuthenticated ? backend.isAuthenticated() : false
    );
    setCurrentAuthor(backend.getCurrentAuthor?.() || null);

    // Listen for auth changes if backend supports it
    if (backend.onAuthChange) {
      backend.onAuthChange(() => {
        setIsAuthenticated(
          backend.isAuthenticated ? backend.isAuthenticated() : false
        );
        setCurrentAuthor(backend.getCurrentAuthor?.() || null);
      });
    }
  }, [backend]);

  return {
    isAuthenticated,
    currentAuthor,
    signIn: backend.signIn,
    signOut: backend.signOut,
  };
}
