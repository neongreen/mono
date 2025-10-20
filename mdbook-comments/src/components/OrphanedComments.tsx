/**
 * OrphanedComments component - Displays comments that can't be matched to paragraphs
 */

import type { Comment } from '../types';

interface OrphanedCommentsProps {
  comments: Comment[];
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

function OrphanedComment({ comment }: { comment: Comment }) {
  const meta = comment.metadata || ({} as any);
  const context = meta.context || { 'heading-path': [] };
  const content = meta.content || '[Content not available]';
  const headingPath = context['heading-path'] || [];

  return (
    <div class="orphaned-comment">
      <div class="orphaned-comment-context">
        <strong>Original paragraph:</strong>
        <blockquote
          dangerouslySetInnerHTML={{ __html: escapeHtml(content) }}
        />
        {headingPath.length > 0 && (
          <div class="orphaned-comment-location">
            Section: {headingPath.join(' > ')}
          </div>
        )}
      </div>

      <div class="comment-item">
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

        {/* Render replies if present */}
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
      </div>
    </div>
  );
}

export function OrphanedComments({ comments }: OrphanedCommentsProps) {
  if (comments.length === 0) {
    return null;
  }

  return (
    <div class="orphaned-comments-section">
      <h2>Unmapped Comments</h2>
      <p class="orphaned-comments-note">
        The following comments could not be matched to any current paragraph.
        They may refer to content that has been removed or significantly
        changed.
      </p>
      <div class="orphaned-comments-list">
        {comments.map((comment) => (
          <OrphanedComment key={comment.id} comment={comment} />
        ))}
      </div>
    </div>
  );
}
