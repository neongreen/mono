/**
 * CommentForm component - Form for submitting new comments
 */

import { useState } from 'preact/hooks';
import type { BackendAdapter, ParagraphMetadata } from '../types';

interface CommentFormProps {
  paragraphId: string;
  metadata: ParagraphMetadata;
  backend: BackendAdapter;
  onCommentSubmitted: () => void;
}

export function CommentForm({
  paragraphId,
  metadata,
  backend,
  onCommentSubmitted,
}: CommentFormProps) {
  const [text, setText] = useState('');
  const [author, setAuthor] = useState(
    backend.getCurrentAuthor?.() || ''
  );
  const [isSubmitting, setIsSubmitting] = useState(false);

  const showAuthorInput = backend.showAuthorInput && !backend.getCurrentAuthor?.();

  const handleSubmit = async (e: Event) => {
    e.preventDefault();

    if (!text.trim()) {
      alert('Please enter a comment');
      return;
    }

    if (showAuthorInput && !author.trim()) {
      alert('Please enter your name');
      return;
    }

    setIsSubmitting(true);

    try {
      // Save author for future use
      if (author && backend.setCurrentAuthor) {
        backend.setCurrentAuthor(author);
      }

      await backend.saveComment(
        paragraphId,
        metadata,
        text.trim(),
        author.trim() || 'Anonymous'
      );

      // Clear form and notify parent
      setText('');
      onCommentSubmitted();
    } catch (error) {
      console.error('Error posting comment:', error);
      alert('Failed to post comment. Please try again.');
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div class="comment-form">
      <form onSubmit={handleSubmit}>
        {showAuthorInput && (
          <input
            type="text"
            class="author-input"
            name="author"
            placeholder="Your name"
            value={author}
            onInput={(e) => {
              const value = (e.target as HTMLInputElement).value;
              setAuthor(value);
              // Save to backend on input for localStorage-based backends
              if (value && backend.setCurrentAuthor) {
                backend.setCurrentAuthor(value);
              }
            }}
            disabled={isSubmitting}
          />
        )}
        <textarea
          class="comment-input"
          name="comment-text"
          placeholder="Add a comment..."
          rows={3}
          value={text}
          onInput={(e) => setText((e.target as HTMLTextAreaElement).value)}
          disabled={isSubmitting}
        />
        <div class="markdown-help">
          <small>
            Supports Markdown: **bold**, *italic*, `code`, [links](url), and lists
          </small>
        </div>
        <button
          type="submit"
          class="comment-submit"
          disabled={isSubmitting}
        >
          {isSubmitting ? 'Submitting...' : 'Submit'}
        </button>
      </form>
    </div>
  );
}
