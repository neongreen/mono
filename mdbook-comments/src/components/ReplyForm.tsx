/**
 * ReplyForm component - Form for submitting replies to comments
 */

import { useState } from 'preact/hooks';
import type { BackendAdapter } from '../types';

interface ReplyFormProps {
  parentCommentId: string;
  backend: BackendAdapter;
  onReplySubmitted: () => void;
  onCancel: () => void;
}

export function ReplyForm({
  parentCommentId,
  backend,
  onReplySubmitted,
  onCancel,
}: ReplyFormProps) {
  const [text, setText] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleSubmit = async (e: Event) => {
    e.preventDefault();

    if (!text.trim()) {
      alert('Please enter a reply');
      return;
    }

    setIsSubmitting(true);

    try {
      const author = backend.getCurrentAuthor?.() || 'Anonymous';
      await backend.saveReply(parentCommentId, text.trim(), author);

      // Clear form and notify parent
      setText('');
      onReplySubmitted();
    } catch (error) {
      console.error('Error posting reply:', error);
      alert('Failed to post reply. Please try again.');
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div class="reply-form">
      <form onSubmit={handleSubmit}>
        <textarea
          class="reply-input"
          placeholder="Write a reply..."
          rows={2}
          value={text}
          onInput={(e) => setText((e.target as HTMLTextAreaElement).value)}
          disabled={isSubmitting}
          autoFocus
        />
        <div class="reply-form-actions">
          <button
            type="submit"
            class="reply-submit"
            disabled={isSubmitting}
          >
            {isSubmitting ? 'Posting...' : 'Post Reply'}
          </button>
          <button
            type="button"
            class="reply-cancel"
            onClick={onCancel}
            disabled={isSubmitting}
          >
            Cancel
          </button>
        </div>
      </form>
    </div>
  );
}
