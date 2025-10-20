/**
 * mdbook-comments - Supabase backend adapter
 *
 * This adapter implements the backend interface for Supabase.
 * It uses the Supabase JavaScript client for database operations.
 *
 * Configuration via window.SUPABASE_CONFIG:
 * {
 *   url: 'https://xxxxx.supabase.co',
 *   anonKey: 'your-anon-key'
 * }
 */

(function() {
    'use strict';

    // Configuration (injected or defaults)
    const supabaseConfig = window.SUPABASE_CONFIG || {
        url: 'YOUR_SUPABASE_PROJECT_URL',
        anonKey: 'YOUR_SUPABASE_ANON_KEY'
    };

    const SUPABASE_URL = supabaseConfig.url;
    const SUPABASE_ANON_KEY = supabaseConfig.anonKey;

    // State
    let supabaseClient;
    let currentAuthor = null;

    /**
     * Supabase backend adapter
     */
    const supabaseBackend = {
        /**
         * Initialize the backend
         */
        init: async function() {
            console.log('Initializing Supabase backend...');
            console.log('Supabase URL:', SUPABASE_URL);

            // Check if Supabase SDK is loaded
            if (typeof supabase === 'undefined' || !supabase.createClient) {
                console.error('Supabase SDK not loaded. Add the SDK to your book.toml');
                throw new Error('Supabase SDK not available');
            }

            // Initialize Supabase client
            supabaseClient = supabase.createClient(SUPABASE_URL, SUPABASE_ANON_KEY);

            // Load author name from localStorage
            currentAuthor = localStorage.getItem('mdbook-comments-author') || '';
        },

        /**
         * Load all comments from the server
         */
        loadComments: async function() {
            const { data, error } = await supabaseClient
                .from('comments')
                .select('*')
                .order('created', { ascending: true });

            if (error) {
                throw new Error(`Failed to load comments: ${error.message}`);
            }

            return data || [];
        },

        /**
         * Save a new comment
         */
        saveComment: async function(paragraphId, metadata, text, author) {
            const { data, error } = await supabaseClient
                .from('comments')
                .insert({
                    paragraph_id: paragraphId,
                    metadata: metadata,
                    author: author || 'Anonymous',
                    text: text,
                    parent_id: null
                })
                .select()
                .single();

            if (error) {
                throw new Error('Failed to post comment');
            }

            return data;
        },

        /**
         * Save a reply to an existing comment
         */
        saveReply: async function(parentCommentId, text, author) {
            // We need to find the parent comment to get its paragraph_id and metadata
            const { data: parentComment, error: fetchError } = await supabaseClient
                .from('comments')
                .select('paragraph_id, metadata')
                .eq('id', parentCommentId)
                .single();

            if (fetchError) {
                throw new Error('Failed to find parent comment');
            }

            const { data, error } = await supabaseClient
                .from('comments')
                .insert({
                    paragraph_id: parentComment.paragraph_id,
                    metadata: parentComment.metadata,
                    author: author || 'Anonymous',
                    text: text,
                    parent_id: parentCommentId
                })
                .select()
                .single();

            if (error) {
                throw new Error('Failed to post reply');
            }

            return data;
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
            window.MdbookComments.init(supabaseBackend);
        });
    } else {
        window.MdbookComments.init(supabaseBackend);
    }

})();
