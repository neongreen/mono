/**
 * CommentSection component - Main container for paragraph comments
 */

import { Comment } from './Comment';
import { CommentForm } from './CommentForm';
import type {
  BackendAdapter,
  ParagraphMetadata,
  MatchedComment,
} from '../types';

interface CommentSectionProps {
  paragraphId: string;
  metadata: ParagraphMetadata;
  comments: MatchedComment[];
  backend: BackendAdapter;
  onUpdate: () => void;
}

export function CommentSection({
  paragraphId,
  metadata,
  comments,
  backend,
  onUpdate,
}: CommentSectionProps) {
  // Get top-level comments (not replies)
  const topLevelComments = comments
    .filter((mc) => mc.paragraphId === paragraphId)
    .map((mc) => mc.comment);

  return (
    <div
      id={`comments-${paragraphId}`}
      class="comment-section"
      data-paragraph-id={paragraphId}
    >
      <div class="comment-list">
        {topLevelComments.length > 0 ? (
          topLevelComments.map((comment) => (
            <Comment
              key={`${comment.id}-replies-${comment.replies?.length || 0}`}
              comment={comment}
              backend={backend}
              onUpdate={onUpdate}
            />
          ))
        ) : (
          <p class="no-comments">No comments yet. Be the first to comment!</p>
        )}
      </div>

      <CommentForm
        paragraphId={paragraphId}
        metadata={metadata}
        backend={backend}
        onCommentSubmitted={onUpdate}
      />
    </div>
  );
}
