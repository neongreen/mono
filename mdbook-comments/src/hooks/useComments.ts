/**
 * useComments hook - Manages comment loading and state
 */

import { useState, useEffect } from 'preact/hooks';
import type { Comment, BackendAdapter } from '../types';

interface UseCommentsOptions {
  backend: BackendAdapter;
  paragraphId?: string;
  autoLoad?: boolean;
}

interface UseCommentsResult {
  comments: Comment[];
  isLoading: boolean;
  error: Error | null;
  reload: () => Promise<void>;
}

/**
 * Hook for loading and managing comments
 */
export function useComments({
  backend,
  paragraphId,
  autoLoad = true,
}: UseCommentsOptions): UseCommentsResult {
  const [comments, setComments] = useState<Comment[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);

  const loadComments = async () => {
    setIsLoading(true);
    setError(null);

    try {
      const allComments = await backend.loadComments();

      // Filter by paragraph if specified
      if (paragraphId) {
        const filtered = allComments.filter(
          (c) => c.paragraph_id === paragraphId
        );
        setComments(filtered);
      } else {
        setComments(allComments);
      }
    } catch (err) {
      const error = err instanceof Error ? err : new Error(String(err));
      setError(error);
      console.error('Error loading comments:', error);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    if (autoLoad) {
      loadComments();
    }
  }, [paragraphId, autoLoad]);

  return {
    comments,
    isLoading,
    error,
    reload: loadComments,
  };
}
