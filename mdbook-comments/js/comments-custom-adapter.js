/**
 * mdbook-comments - Custom backend adapter
 *
 * This adapter provides a generic REST API implementation that can work
 * with any backend following the standard API contract.
 *
 * Configuration via window.MDBOOK_COMMENTS_CONFIG:
 * {
 *   apiUrl: 'http://your-api.com/comments',
 *   similarityThreshold: 0.85,       // optional
 *   orphanedLocation: 'end-of-chapter'  // optional
 * }
 */

(function() {
    'use strict';

    // Configuration from window or defaults
    const CONFIG = window.MDBOOK_COMMENTS_CONFIG || {
        apiUrl: 'http://localhost:3000/api',
        similarityThreshold: 0.85,
        orphanedLocation: 'end-of-chapter'
    };

    const API_URL = CONFIG.apiUrl;

    /**
     * Custom backend adapter
     */
    const customBackend = {
        /**
         * Initialize the backend
         */
        init: async function() {
            console.log('Initializing custom backend...');
            console.log('API URL:', API_URL);
        },

        /**
         * Load all comments from the server
         */
        loadComments: async function() {
            const response = await fetch(`${API_URL}/comments`, {
                credentials: 'include' // Include cookies for authentication
            });

            if (!response.ok) {
                throw new Error(`Failed to load comments: ${response.statusText}`);
            }

            const comments = await response.json();

            // Normalize any naming variations to internal format
            return comments.map(comment => ({
                ...comment,
                paragraph_id: comment['paragraph-id'] || comment.paragraph_id,
                parent_id: comment['parent-id'] || comment.parent_id
            }));
        },

        /**
         * Save a new comment
         */
        saveComment: async function(paragraphId, metadata, text, author) {
            const response = await fetch(`${API_URL}/comments`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                credentials: 'include',
                body: JSON.stringify({
                    'paragraph-id': paragraphId,
                    metadata: metadata,
                    text: text,
                    author: author || 'Anonymous'
                })
            });

            if (!response.ok) {
                throw new Error('Failed to post comment');
            }

            const comment = await response.json();

            // Normalize response
            return {
                ...comment,
                paragraph_id: comment['paragraph-id'] || comment.paragraph_id,
                parent_id: comment['parent-id'] || comment.parent_id
            };
        },

        /**
         * Save a reply to an existing comment
         */
        saveReply: async function(parentCommentId, text, author) {
            const response = await fetch(`${API_URL}/comments/${parentCommentId}/reply`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                credentials: 'include',
                body: JSON.stringify({
                    text: text,
                    author: author || 'Anonymous'
                })
            });

            if (!response.ok) {
                throw new Error('Failed to post reply');
            }

            const reply = await response.json();

            // Normalize response
            return {
                ...reply,
                paragraph_id: reply['paragraph-id'] || reply.paragraph_id,
                parent_id: reply['parent-id'] || reply.parent_id
            };
        },

        /**
         * Get current author name (no persistence in generic custom backend)
         */
        getCurrentAuthor: function() {
            return null;
        },

        /**
         * Whether to show author input field in forms
         */
        showAuthorInput: true
    };

    // Initialize when DOM is ready
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', () => {
            window.MdbookComments.init(customBackend);
        });
    } else {
        window.MdbookComments.init(customBackend);
    }

})();
