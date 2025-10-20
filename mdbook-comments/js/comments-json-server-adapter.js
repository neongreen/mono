/**
 * mdbook-comments - JSON Server backend adapter
 *
 * This adapter implements the backend interface for json-server.
 * It provides simple REST API calls to a json-server instance.
 */

(function() {
    'use strict';

    // Configuration (injected by preprocessor or JSON server config)
    const jsonServerConfig = window.JSON_SERVER_CONFIG || {
        url: 'http://localhost:54322'
    };
    const API_URL = jsonServerConfig.url;

    // State
    let currentAuthor = null;

    /**
     * JSON Server backend adapter
     */
    const jsonServerBackend = {
        /**
         * Initialize the backend
         */
        init: async function() {
            console.log('Initializing json-server backend...');
            console.log('API URL:', API_URL);

            // Load author name from localStorage
            currentAuthor = localStorage.getItem('mdbook-comments-author') || '';
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

            // Normalize kebab-case to snake_case for internal consistency
            return comments.map(comment => ({
                ...comment,
                paragraph_id: comment['paragraph-id'] || comment.paragraph_id,
                parent_id: comment['parent-id'] || comment.parent_id,
                // Keep both formats for compatibility
                'paragraph-id': comment['paragraph-id'],
                'parent-id': comment['parent-id']
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

            return await response.json();
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

            return await response.json();
        },

        /**
         * Get current author name
         */
        getCurrentAuthor: function() {
            return currentAuthor;
        },

        /**
         * Set current author name
         */
        setCurrentAuthor: function(author) {
            currentAuthor = author;
            localStorage.setItem('mdbook-comments-author', author);
        },

        /**
         * Whether to show author input field in forms
         */
        showAuthorInput: true
    };

    // Initialize when DOM is ready
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', () => {
            window.MdbookComments.init(jsonServerBackend);
        });
    } else {
        window.MdbookComments.init(jsonServerBackend);
    }

})();
