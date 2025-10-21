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

function formatRelativeDate(dateString: string): string {
  const date = new Date(dateString);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffMins = Math.floor(diffMs / 60000);
  const diffHours = Math.floor(diffMs / 3600000);
  const diffDays = Math.floor(diffMs / 86400000);
  const diffWeeks = Math.floor(diffDays / 7);
  const diffMonths = Math.floor(diffDays / 30);
  const diffYears = Math.floor(diffDays / 365);

  if (diffMins < 1) return 'just now';
  if (diffMins < 60) return `${diffMins} minute${diffMins > 1 ? 's' : ''} ago`;
  if (diffHours < 24) return `${diffHours} hour${diffHours > 1 ? 's' : ''} ago`;
  if (diffDays < 7) return `${diffDays} day${diffDays > 1 ? 's' : ''} ago`;
  if (diffWeeks < 4) return `${diffWeeks} week${diffWeeks > 1 ? 's' : ''} ago`;
  if (diffMonths < 12) return `${diffMonths} month${diffMonths > 1 ? 's' : ''} ago`;
  
  return `${diffYears} year${diffYears > 1 ? 's' : ''} ago`;
}

function formatAbsoluteDate(dateString: string): string {
  const date = new Date(dateString);
  return date.toLocaleString('en-US', {
    year: 'numeric',
    month: 'long', 
    day: 'numeric',
    hour: 'numeric',
    minute: '2-digit',
    hour12: true
  });
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
        <span 
          class="comment-date" 
          title={formatAbsoluteDate(comment.created)}
        >
          {formatRelativeDate(comment.created)}
        </span>
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
                <span 
                  class="reply-date" 
                  title={formatAbsoluteDate(reply.created)}
                >
                  {formatRelativeDate(reply.created)}
                </span>
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
