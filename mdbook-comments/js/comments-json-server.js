/**
 * mdbook-comments - Client-side comment system with JSON Server backend
 *
 * This script handles:
 * - Loading comments from JSON Server (local development)
 * - Fuzzy matching of comments to current paragraphs
 * - Displaying inline comment sections
 * - Handling user interactions (expand/collapse, reply)
 *
 * Configuration can be provided via:
 * 1. window.JSON_SERVER_CONFIG = { url: '...' }
 * 2. Fallback to default localhost:54321
 */

(function() {
    'use strict';

    // Get JSON Server configuration from window or use defaults
    const jsonServerConfig = window.JSON_SERVER_CONFIG || {
        url: 'http://localhost:54322'
    };

    const JSON_SERVER_URL = jsonServerConfig.url;

    // Configuration
    const CONFIG = {
        similarityThreshold: 0.85,
        orphanedLocation: 'end-of-chapter',
        showCommentCount: true
    };

    // State
    let allComments = [];
    let currentPageComments = [];
    let orphanedComments = [];
    let currentAuthor = null; // Simple name string, stored in localStorage

    /**
     * Initialize the comment system
     */
    async function init() {
        console.log('Initializing mdbook-comments with JSON Server...');
        console.log('JSON Server URL:', JSON_SERVER_URL);

        // Load author name from localStorage
        currentAuthor = localStorage.getItem('mdbook-comments-author') || '';

        // Load comments for current page
        await loadComments();

        // Set up event listeners
        setupEventListeners();

        // Add styles
        injectStyles();
    }

    /**
     * Load comments from JSON Server
     */
    async function loadComments() {
        try {
            const response = await fetch(`${JSON_SERVER_URL}/comments`);
            if (!response.ok) {
                console.error('Failed to load comments:', response.status, response.statusText);
                return;
            }

            const data = await response.json();
            allComments = data || [];
            console.log(`Loaded ${allComments.length} comments`);

            // Build reply structure
            buildReplyStructure();

            // Match comments to current paragraphs
            matchComments();

            // Update comment counts in links
            updateCommentCounts();

            // Display orphaned comments if any
            displayOrphanedComments();

        } catch (error) {
            console.error('Failed to load comments:', error);
        }
    }

    /**
     * Save a new comment to JSON Server
     */
    async function saveComment(paragraphId, metadata, text, parentId = null) {
        try {
            const commentData = {
                'paragraph-id': paragraphId,
                metadata: metadata,
                author: currentAuthor || 'Anonymous',
                text: text,
                'parent-id': parentId
            };

            const response = await fetch(`${JSON_SERVER_URL}/comments`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(commentData)
            });

            if (!response.ok) {
                console.error('Failed to save comment:', response.status, response.statusText);
                return null;
            }

            const newComment = await response.json();
            console.log('Comment saved:', newComment);

            // Add to local state
            allComments.push(newComment);

            // Rebuild structures
            buildReplyStructure();
            matchComments();
            updateCommentCounts();

            return newComment;

        } catch (error) {
            console.error('Failed to save comment:', error);
            return null;
        }
    }

    /**
     * Save a reply to JSON Server (same as saving a comment with parent-id)
     */
    async function saveReply(parentCommentId, paragraphId, metadata, text) {
        return await saveComment(paragraphId, metadata, text, parentCommentId);
    }

    /**
     * Build reply structure by linking parent/child comments
     */
    function buildReplyStructure() {
        // Group comments by parent
        const commentMap = new Map();
        const rootComments = [];

        // Initialize all comments with empty replies array
        allComments.forEach(comment => {
            comment.replies = [];
            commentMap.set(comment.id, comment);
        });

        // Build hierarchy
        allComments.forEach(comment => {
            if (comment['parent-id']) {
                const parent = commentMap.get(comment['parent-id']);
                if (parent) {
                    parent.replies.push(comment);
                }
            } else {
                rootComments.push(comment);
            }
        });

        // Update allComments to only include root comments with nested replies
        allComments = rootComments;
    }

    /**
     * Match comments to paragraphs on the current page using fuzzy matching
     */
    function matchComments() {
        // This is the same logic as the Supabase version
        // Find all commentable elements
        const commentableElements = document.querySelectorAll('[data-comment-id]');

        currentPageComments = [];
        orphanedComments = [];

        commentableElements.forEach(element => {
            const paragraphId = element.getAttribute('data-comment-id');
            const paragraphText = element.textContent.trim();

            // Find matching comments
            const matchingComments = findMatchingComments(paragraphId, paragraphText);

            if (matchingComments.length > 0) {
                currentPageComments.push({
                    element: element,
                    paragraphId: paragraphId,
                    comments: matchingComments
                });
            }
        });

        // Find orphaned comments (comments that don't match any current paragraph)
        allComments.forEach(comment => {
            let isMatched = false;
            currentPageComments.forEach(pageComment => {
                if (pageComment.comments.includes(comment)) {
                    isMatched = true;
                }
            });
            if (!isMatched) {
                orphanedComments.push(comment);
            }
        });
    }

    /**
     * Find comments that match a paragraph
     */
    function findMatchingComments(paragraphId, paragraphText) {
        const matches = [];

        allComments.forEach(comment => {
            // Exact ID match
            if (comment['paragraph-id'] === paragraphId) {
                matches.push(comment);
                return;
            }

            // Fuzzy content matching
            if (comment.metadata && comment.metadata.content) {
                const similarity = calculateSimilarity(paragraphText, comment.metadata.content);
                if (similarity >= CONFIG.similarityThreshold) {
                    matches.push(comment);
                }
            }
        });

        return matches;
    }

    /**
     * Calculate text similarity (simple implementation)
     */
    function calculateSimilarity(text1, text2) {
        // Simple Levenshtein distance based similarity
        const longer = text1.length > text2.length ? text1 : text2;
        const shorter = text1.length > text2.length ? text2 : text1;

        if (longer.length === 0) return 1.0;

        const distance = levenshteinDistance(longer, shorter);
        return (longer.length - distance) / longer.length;
    }

    /**
     * Levenshtein distance calculation
     */
    function levenshteinDistance(str1, str2) {
        const matrix = [];

        for (let i = 0; i <= str2.length; i++) {
            matrix[i] = [i];
        }

        for (let j = 0; j <= str1.length; j++) {
            matrix[0][j] = j;
        }

        for (let i = 1; i <= str2.length; i++) {
            for (let j = 1; j <= str1.length; j++) {
                if (str2.charAt(i - 1) === str1.charAt(j - 1)) {
                    matrix[i][j] = matrix[i - 1][j - 1];
                } else {
                    matrix[i][j] = Math.min(
                        matrix[i - 1][j - 1] + 1, // substitution
                        matrix[i][j - 1] + 1,     // insertion
                        matrix[i - 1][j] + 1      // deletion
                    );
                }
            }
        }

        return matrix[str2.length][str1.length];
    }

    /**
     * Update comment count displays
     */
    function updateCommentCounts() {
        // This is the same logic as the Supabase version
        currentPageComments.forEach(pageComment => {
            const element = pageComment.element;
            const comments = pageComment.comments;
            const totalComments = countTotalComments(comments);

            // Find the comment link
            const commentLink = element.querySelector('.comment-link');
            if (commentLink && CONFIG.showCommentCount && totalComments > 0) {
                commentLink.textContent = `comment (${totalComments})`;
            }
        });
    }

    /**
     * Count total comments including replies
     */
    function countTotalComments(comments) {
        let count = 0;
        comments.forEach(comment => {
            count += 1 + countTotalComments(comment.replies || []);
        });
        return count;
    }

    /**
     * Display orphaned comments
     */
    function displayOrphanedComments() {
        if (orphanedComments.length === 0) return;

        // Find where to display orphaned comments
        let container;
        if (CONFIG.orphanedLocation === 'end-of-page') {
            container = document.body;
        } else { // end-of-chapter
            // Find the last chapter
            const chapters = document.querySelectorAll('.chapter');
            container = chapters[chapters.length - 1];
        }

        if (!container) return;

        // Create orphaned comments section
        const orphanedSection = document.createElement('div');
        orphanedSection.className = 'orphaned-comments-section';
        orphanedSection.innerHTML = `
            <h3>Orphaned Comments</h3>
            <p>These comments could not be matched to current content:</p>
        `;

        orphanedComments.forEach(comment => {
            const commentDiv = createCommentElement(comment, true);
            orphanedSection.appendChild(commentDiv);
        });

        container.appendChild(orphanedSection);
    }

    /**
     * Set up event listeners
     */
    function setupEventListeners() {
        // Listen for clicks on comment links and expand/collapse buttons
        document.addEventListener('click', handleClick);
    }

    /**
     * Handle clicks
     */
    function handleClick(event) {
        const target = event.target;

        // Comment link clicked
        if (target.classList.contains('comment-link')) {
            event.preventDefault();
            const paragraphElement = target.closest('[data-comment-id]');
            if (paragraphElement) {
                toggleCommentSection(paragraphElement);
            }
        }

        // Reply button clicked
        if (target.classList.contains('reply-button')) {
            event.preventDefault();
            const commentElement = target.closest('.comment');
            if (commentElement) {
                showReplyForm(commentElement);
            }
        }

        // Submit comment button clicked
        if (target.classList.contains('submit-comment')) {
            event.preventDefault();
            const form = target.closest('.comment-form');
            if (form) {
                submitComment(form);
            }
        }

        // Cancel button clicked
        if (target.classList.contains('cancel-comment')) {
            event.preventDefault();
            const form = target.closest('.comment-form');
            if (form) {
                form.remove();
            }
        }
    }

    /**
     * Toggle comment section visibility
     */
    function toggleCommentSection(paragraphElement) {
        const paragraphId = paragraphElement.getAttribute('data-comment-id');
        let commentSection = paragraphElement.querySelector('.comment-section');

        if (commentSection) {
            // Toggle visibility
            commentSection.classList.toggle('hidden');
        } else {
            // Create new comment section
            commentSection = document.createElement('div');
            commentSection.id = `comments-${paragraphId}`;
            commentSection.className = 'comment-section';
            commentSection.setAttribute('data-paragraph-id', paragraphId); // Store for reply forms

            // Create comment list container
            const commentList = document.createElement('div');
            commentList.className = 'comment-list';

            // Find matching comments
            const pageComment = currentPageComments.find(pc => pc.paragraphId === paragraphId);
            if (pageComment) {
                pageComment.comments.forEach(comment => {
                    const commentElement = createCommentElement(comment);
                    commentList.appendChild(commentElement);
                });
            }

            // Append comment list to section
            commentSection.appendChild(commentList);

            // Add new comment form
            const newCommentForm = createCommentForm(paragraphId, null);
            commentSection.appendChild(newCommentForm);

            paragraphElement.appendChild(commentSection);
        }
    }

    /**
     * Create a comment element
     */
    function createCommentElement(comment, isOrphaned = false) {
        const commentDiv = document.createElement('div');
        commentDiv.className = 'comment';
        commentDiv.setAttribute('data-comment-id', comment.id);

        const metadata = comment.metadata || {};
        const created = new Date(comment.created).toLocaleString();

        commentDiv.innerHTML = `
            <div class="comment-header">
                <strong>${comment.author}</strong>
                <span class="comment-date">${created}</span>
            </div>
            <div class="comment-content">${comment.text}</div>
            <div class="comment-actions">
                <button class="reply-button">Reply</button>
            </div>
        `;

        // Add replies
        if (comment.replies && comment.replies.length > 0) {
            const repliesDiv = document.createElement('div');
            repliesDiv.className = 'comment-replies';
            comment.replies.forEach(reply => {
                const replyElement = createCommentElement(reply);
                repliesDiv.appendChild(replyElement);
            });
            commentDiv.appendChild(repliesDiv);
        }

        return commentDiv;
    }

    /**
     * Create a comment form
     */
    function createCommentForm(paragraphId, parentCommentId) {
        const form = document.createElement('form');
        form.className = 'comment-form';

        // Determine if this is a reply or a new comment
        const isReply = !!parentCommentId;
        const textareaName = isReply ? 'reply-text' : 'comment-text';
        const placeholder = isReply ? 'Write a reply...' : 'Write a comment...';
        const submitText = isReply ? 'Post Reply' : 'Submit';

        // Author field (only show if not set)
        let authorField = '';
        if (!currentAuthor) {
            authorField = `
                <input type="text" name="author" class="author-input" placeholder="Your name" required>
            `;
        }

        form.innerHTML = `
            ${authorField}
            <textarea name="${textareaName}" class="comment-textarea" placeholder="${placeholder}" required></textarea>
            <div class="comment-buttons">
                <button type="submit" class="submit-comment">${submitText}</button>
                <button type="button" class="cancel-comment">Cancel</button>
            </div>
        `;

        // Store metadata for submission
        form.setAttribute('data-paragraph-id', paragraphId);
        if (parentCommentId) {
            form.setAttribute('data-parent-id', parentCommentId);
        }

        // Add event listener for author input to save to localStorage immediately
        if (!currentAuthor) {
            const authorInput = form.querySelector('.author-input');
            if (authorInput) {
                authorInput.addEventListener('input', (e) => {
                    const value = e.target.value.trim();
                    if (value) {
                        currentAuthor = value;
                        localStorage.setItem('mdbook-comments-author', value);
                    }
                });
            }
        }

        return form;
    }

    /**
     * Show reply form
     */
    function showReplyForm(commentElement) {
        const commentId = commentElement.getAttribute('data-comment-id');
        console.log('[showReplyForm] commentId:', commentId);

        // Check if reply form already exists
        if (commentElement.querySelector('.comment-form')) {
            console.log('[showReplyForm] Reply form already exists');
            return;
        }

        // Get paragraph ID from the comment section
        // We can't reliably use closest() because the comment section div
        // might be moved out of the inline span by the browser's HTML parser
        const commentSection = commentElement.closest('.comment-section');
        console.log('[showReplyForm] commentSection:', commentSection);
        const paragraphId = commentSection ? commentSection.getAttribute('data-paragraph-id') : null;
        console.log('[showReplyForm] paragraphId:', paragraphId);

        const replyForm = createCommentForm(paragraphId, commentId);
        commentElement.appendChild(replyForm);
        console.log('[showReplyForm] Reply form appended');
    }

    /**
     * Submit a comment
     */
    async function submitComment(form) {
        const paragraphId = form.getAttribute('data-paragraph-id');
        const parentId = form.getAttribute('data-parent-id');

        // Get author
        let author = currentAuthor;
        const authorInput = form.querySelector('.author-input');
        if (authorInput) {
            author = authorInput.value.trim();
            if (author) {
                currentAuthor = author;
                localStorage.setItem('mdbook-comments-author', author);
            }
        }

        // Get text
        const textarea = form.querySelector('.comment-textarea');
        const text = textarea.value.trim();

        if (!text) return;

        // Get metadata (would need to be stored in the form or retrieved)
        const metadata = {}; // Simplified for now

        let result;
        if (parentId) {
            // This is a reply
            result = await saveReply(parentId, paragraphId, metadata, text);
        } else {
            // This is a new comment
            result = await saveComment(paragraphId, metadata, text);
        }

        if (result) {
            // Remove form and refresh display
            form.remove();

            // Re-render the comment section
            const paragraphElement = document.querySelector(`[data-comment-id="${paragraphId}"]`);
            if (paragraphElement) {
                const commentSection = paragraphElement.querySelector('.comment-section');
                if (commentSection) {
                    commentSection.remove();
                    toggleCommentSection(paragraphElement);
                }
            }
        }
    }

    /**
     * Inject required CSS styles
     */
    function injectStyles() {
        // This is the same as the Supabase version
        const style = document.createElement('style');
        style.textContent = `
            /* Comment links */
            .comment-link-wrapper {
                display: inline;
                margin-left: 0.5em;
            }

            .comment-link {
                font-size: 0.85em;
                text-decoration: underline;
                cursor: pointer;
                opacity: 0.8;
                transition: opacity 0.2s;
            }

            .comment-link:hover {
                opacity: 1;
            }

            /* Comment sections */
            .comment-section {
                margin: 1.5em 0;
                animation: fadeIn 0.3s;
            }

            .comment-section.hidden {
                display: none;
            }

            @keyframes fadeIn {
                from {
                    opacity: 0;
                    transform: translateY(-10px);
                }
                to {
                    opacity: 1;
                    transform: translateY(0);
                }
            }

            /* Comment elements */
            .comment {
                margin: 1em 0;
                padding: 1em;
                border-left: 3px solid #007acc;
                background: #f8f9fa;
            }

            .comment-header {
                display: flex;
                justify-content: space-between;
                margin-bottom: 0.5em;
            }

            .comment-date {
                color: #666;
                font-size: 0.9em;
            }

            .comment-content {
                margin-bottom: 0.5em;
            }

            .comment-actions {
                margin-top: 0.5em;
            }

            .reply-button {
                background: none;
                border: none;
                color: #007acc;
                cursor: pointer;
                font-size: 0.9em;
            }

            .reply-button:hover {
                text-decoration: underline;
            }

            /* Comment replies */
            .comment-replies {
                margin-left: 2em;
                margin-top: 1em;
            }

            .comment-replies .comment {
                border-left-color: #28a745;
                background: #f8fff8;
            }

            /* Comment forms */
            .comment-form {
                margin: 1em 0;
                padding: 1em;
                background: #f8f9fa;
                border-radius: 4px;
            }

            .author-input {
                display: block;
                width: 100%;
                padding: 8px;
                margin-bottom: 8px;
                border: 1px solid #ccc;
                border-radius: 4px;
                font-size: 14px;
                box-sizing: border-box;
            }

            .comment-textarea {
                display: block;
                width: 100%;
                min-height: 80px;
                padding: 8px;
                margin-bottom: 8px;
                border: 1px solid #ccc;
                border-radius: 4px;
                font-size: 14px;
                box-sizing: border-box;
                resize: vertical;
            }

            .comment-buttons {
                text-align: right;
            }

            .submit-comment {
                background: #007acc;
                color: white;
                border: none;
                padding: 8px 16px;
                border-radius: 4px;
                cursor: pointer;
                margin-right: 8px;
            }

            .submit-comment:hover {
                background: #0056b3;
            }

            .cancel-comment {
                background: #6c757d;
                color: white;
                border: none;
                padding: 8px 16px;
                border-radius: 4px;
                cursor: pointer;
            }

            .cancel-comment:hover {
                background: #545b62;
            }

            /* Orphaned comments */
            .orphaned-comments-section {
                margin: 2em 0;
                padding: 1em;
                border: 1px solid #ffc107;
                background: #fffbf0;
            }

            .orphaned-comments-section h3 {
                margin-top: 0;
                color: #856404;
            }
        `;
        document.head.appendChild(style);
    }

    // Initialize when DOM is ready
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', init);
    } else {
        init();
    }

})();
