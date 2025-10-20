/**
 * Comment component - Displays a single comment with replies
 */

import { useState } from 'preact/hooks';
import { ReplyForm } from './ReplyForm';
import type { Comment as CommentType, BackendAdapter } from '../types';

interface CommentProps {
  comment: CommentType;
  backend: BackendAdapter;
  onUpdate: () => void;
}

function escapeHtml(text: string): string {
  const div = document.createElement('div');
  div.textContent = text;
  return div.innerHTML;
}

function formatDate(dateString: string): string {
  const date = new Date(dateString);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffMins = Math.floor(diffMs / 60000);
  const diffHours = Math.floor(diffMs / 3600000);
  const diffDays = Math.floor(diffMs / 86400000);

  if (diffMins < 1) return 'just now';
  if (diffMins < 60) return `${diffMins} minute${diffMins > 1 ? 's' : ''} ago`;
  if (diffHours < 24)
    return `${diffHours} hour${diffHours > 1 ? 's' : ''} ago`;
  if (diffDays < 7) return `${diffDays} day${diffDays > 1 ? 's' : ''} ago`;

  return date.toLocaleDateString();
}

export function Comment({ comment, backend, onUpdate }: CommentProps) {
  const [showReplyForm, setShowReplyForm] = useState(false);

  const handleReplySubmitted = () => {
    setShowReplyForm(false);
    onUpdate();
  };

  return (
    <div class="comment-item" data-comment-id={comment.id}>
      <div class="comment-header">
        <span class="comment-author">
          {escapeHtml(comment.author || 'Anonymous')}
        </span>
        <span class="comment-date">{formatDate(comment.created)}</span>
      </div>
      <div
        class="comment-text"
        dangerouslySetInnerHTML={{ __html: escapeHtml(comment.text) }}
      />

      {/* Nested replies */}
      {comment.replies && comment.replies.length > 0 && (
        <div class="comment-replies">
          {comment.replies.map((reply) => (
            <div key={reply.id} class="reply-item">
              <div class="reply-header">
                <span class="reply-author">
                  {escapeHtml(reply.author || 'Anonymous')}
                </span>
                <span class="reply-date">{formatDate(reply.created)}</span>
              </div>
              <div
                class="reply-text"
                dangerouslySetInnerHTML={{ __html: escapeHtml(reply.text) }}
              />
            </div>
          ))}
        </div>
      )}

      {/* Reply button */}
      <button
        class="comment-reply-btn"
        onClick={() => setShowReplyForm(!showReplyForm)}
      >
        {showReplyForm ? 'Cancel' : 'Reply'}
      </button>

      {/* Reply form */}
      {showReplyForm && (
        <ReplyForm
          parentCommentId={comment.id}
          backend={backend}
          onReplySubmitted={handleReplySubmitted}
          onCancel={() => setShowReplyForm(false)}
        />
      )}
    </div>
  );
}
