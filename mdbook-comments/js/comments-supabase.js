/**
 * mdbook-comments - Client-side comment system with Supabase backend
 * 
 * This script handles:
 * - Loading comments from Supabase
 * - Fuzzy matching of comments to current paragraphs
 * - Displaying inline comment sections
 * - Handling user interactions (expand/collapse, reply)
 * - Authentication via Supabase Auth
 */

(function() {
    'use strict';

    // Supabase configuration - REPLACE THESE WITH YOUR VALUES
    const SUPABASE_URL = 'YOUR_SUPABASE_PROJECT_URL'; // e.g., 'https://xxxxx.supabase.co'
    const SUPABASE_ANON_KEY = 'YOUR_SUPABASE_ANON_KEY'; // Your anon/public key

    // Initialize Supabase client
    let supabaseClient;
    
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
    let currentUser = null;

    /**
     * Initialize the comment system
     */
    async function init() {
        console.log('Initializing mdbook-comments with Supabase...');
        
        // Check if Supabase SDK is loaded
        if (typeof supabase === 'undefined' || !supabase.createClient) {
            console.error('Supabase SDK not loaded. Add the SDK to your book.toml');
            return;
        }
        
        // Initialize Supabase client
        supabaseClient = supabase.createClient(SUPABASE_URL, SUPABASE_ANON_KEY);
        
        // Initialize authentication UI
        await initAuth();
        
        // Load comments for current page
        await loadComments();
        
        // Set up event listeners
        setupEventListeners();
        
        // Add styles
        injectStyles();
    }

    /**
     * Initialize authentication and UI
     */
    async function initAuth() {
        try {
            // Check current auth state
            const { data: { user }, error } = await supabaseClient.auth.getUser();
            
            if (error && error.message !== 'Auth session missing!') {
                console.error('Auth error:', error);
            }
            
            currentUser = user;
            
            // Add auth button to page
            createAuthButton(user);
            
            // Listen for auth state changes
            supabaseClient.auth.onAuthStateChange((event, session) => {
                console.log('Auth state changed:', event);
                if (event === 'SIGNED_IN' || event === 'SIGNED_OUT') {
                    window.location.reload();
                }
            });
        } catch (error) {
            console.error('Error initializing auth:', error);
        }
    }

    /**
     * Create authentication button
     */
    function createAuthButton(user) {
        const authButton = document.createElement('button');
        authButton.id = 'mdbook-auth-button';
        authButton.style.cssText = `
            position: fixed;
            top: 60px;
            right: 20px;
            padding: 8px 16px;
            background: #0066cc;
            color: white;
            border: none;
            border-radius: 4px;
            cursor: pointer;
            z-index: 1000;
            font-size: 14px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.2);
        `;
        
        if (user) {
            const email = user.email || user.id.substring(0, 8);
            authButton.innerHTML = `
                <span style="margin-right: 8px;">👤</span>
                ${escapeHtml(email)}
            `;
            authButton.title = 'Click to sign out';
            authButton.onclick = async () => {
                if (confirm('Sign out?')) {
                    await supabaseClient.auth.signOut();
                    window.location.reload();
                }
            };
        } else {
            authButton.innerHTML = `
                <span style="margin-right: 8px;">🔐</span>
                Sign in to comment
            `;
            authButton.onclick = async () => {
                await supabaseClient.auth.signInWithOAuth({
                    provider: 'google',
                    options: {
                        redirectTo: window.location.href
                    }
                });
            };
        }
        
        document.body.appendChild(authButton);
    }

    /**
     * Load comments from Supabase
     */
    async function loadComments() {
        try {
            // Fetch all comments
            const { data, error } = await supabaseClient
                .from('comments')
                .select('*')
                .order('created', { ascending: true });
            
            if (error) {
                console.error('Failed to load comments:', error);
                return;
            }
            
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
            console.error('Error loading comments:', error);
        }
    }

    /**
     * Build reply structure from flat comment list
     */
    function buildReplyStructure() {
        const commentsById = {};
        
        // First pass: index all comments
        allComments.forEach(comment => {
            comment.replies = [];
            commentsById[comment.id] = comment;
        });
        
        // Second pass: build reply tree
        allComments.forEach(comment => {
            if (comment.parent_id && commentsById[comment.parent_id]) {
                commentsById[comment.parent_id].replies.push(comment);
            }
        });
    }

    /**
     * Match comments to current paragraphs using fuzzy matching
     */
    function matchComments() {
        currentPageComments = [];
        orphanedComments = [];
        
        const commentLinks = document.querySelectorAll('.comment-link-wrapper');
        const usedComments = new Set();
        
        // First pass: exact ID matching
        commentLinks.forEach(link => {
            const paragraphId = link.getAttribute('data-comment-id');
            const metadata = JSON.parse(link.getAttribute('data-comment-meta') || '{}');
            
            const exactMatches = allComments.filter(c => 
                c.metadata && c.metadata.id === paragraphId && 
                !usedComments.has(c.id) &&
                !c.parent_id // Only top-level comments
            );
            
            exactMatches.forEach(comment => {
                currentPageComments.push({ paragraphId, comment, confidence: 1.0 });
                usedComments.add(comment.id);
            });
        });
        
        // Second pass: fuzzy matching for comments without exact match
        commentLinks.forEach(link => {
            const paragraphId = link.getAttribute('data-comment-id');
            const metadata = JSON.parse(link.getAttribute('data-comment-meta') || '{}');
            
            allComments.forEach(comment => {
                if (usedComments.has(comment.id)) return;
                if (!comment.metadata) return;
                if (comment.parent_id) return; // Skip replies
                
                const similarity = calculateSimilarity(metadata, comment.metadata);
                
                if (similarity >= CONFIG.similarityThreshold) {
                    currentPageComments.push({ paragraphId, comment, confidence: similarity });
                    usedComments.add(comment.id);
                }
            });
        });
        
        // Remaining comments are orphaned
        allComments.forEach(comment => {
            if (!usedComments.has(comment.id) && !comment.parent_id) {
                orphanedComments.push(comment);
            }
        });
        
        console.log(`Matched ${currentPageComments.length} comments, ${orphanedComments.length} orphaned`);
    }

    /**
     * Calculate similarity between two paragraph metadata objects
     */
    function calculateSimilarity(meta1, meta2) {
        let score = 0.0;
        let weights = 0.0;
        
        // Content similarity (most important)
        if (meta1.content && meta2.content) {
            const contentSim = textSimilarity(meta1.content, meta2.content);
            score += contentSim * 0.5;
            weights += 0.5;
        }
        
        // Context similarity (prev/next paragraphs)
        if (meta1.context && meta2.context) {
            if (meta1.context.prev && meta2.context.prev) {
                const prevSim = textSimilarity(meta1.context.prev, meta2.context.prev);
                score += prevSim * 0.2;
                weights += 0.2;
            }
            
            if (meta1.context.next && meta2.context.next) {
                const nextSim = textSimilarity(meta1.context.next, meta2.context.next);
                score += nextSim * 0.2;
                weights += 0.2;
            }
            
            // Heading path similarity
            if (meta1.context['heading-path'] && meta2.context['heading-path']) {
                const headingSim = arraySimilarity(
                    meta1.context['heading-path'],
                    meta2.context['heading-path']
                );
                score += headingSim * 0.1;
                weights += 0.1;
            }
        }
        
        return weights > 0 ? score / weights : 0.0;
    }

    /**
     * Calculate text similarity using simple token-based approach
     */
    function textSimilarity(text1, text2) {
        const tokens1 = new Set(tokenize(text1));
        const tokens2 = new Set(tokenize(text2));
        
        const intersection = new Set([...tokens1].filter(x => tokens2.has(x)));
        const union = new Set([...tokens1, ...tokens2]);
        
        return union.size > 0 ? intersection.size / union.size : 0.0;
    }

    /**
     * Tokenize text into words
     */
    function tokenize(text) {
        return text.toLowerCase()
            .replace(/[^\w\s]/g, ' ')
            .split(/\s+/)
            .filter(t => t.length > 2);
    }

    /**
     * Calculate array similarity (Jaccard index)
     */
    function arraySimilarity(arr1, arr2) {
        const set1 = new Set(arr1);
        const set2 = new Set(arr2);
        
        const intersection = new Set([...set1].filter(x => set2.has(x)));
        const union = new Set([...set1, ...set2]);
        
        return union.size > 0 ? intersection.size / union.size : 0.0;
    }

    /**
     * Update comment counts in links
     */
    function updateCommentCounts() {
        document.querySelectorAll('.comment-link-wrapper').forEach(wrapper => {
            const paragraphId = wrapper.getAttribute('data-comment-id');
            const count = currentPageComments.filter(c => c.paragraphId === paragraphId).length;
            
            if (count > 0 && CONFIG.showCommentCount) {
                const link = wrapper.querySelector('.comment-link');
                if (link) {
                    link.textContent = `comment (${count})`;
                }
            }
        });
    }

    /**
     * Display orphaned comments at end of chapter
     */
    function displayOrphanedComments() {
        if (orphanedComments.length === 0) return;
        
        const main = document.querySelector('main') || document.querySelector('#content') || document.body;
        
        const section = document.createElement('div');
        section.className = 'orphaned-comments-section';
        section.innerHTML = `
            <h2>Unmapped Comments</h2>
            <p class="orphaned-comments-note">
                The following comments could not be matched to any current paragraph.
                They may refer to content that has been removed or significantly changed.
            </p>
            <div class="orphaned-comments-list"></div>
        `;
        
        const list = section.querySelector('.orphaned-comments-list');
        
        orphanedComments.forEach(comment => {
            const item = createOrphanedCommentElement(comment);
            list.appendChild(item);
        });
        
        main.appendChild(section);
    }

    /**
     * Create element for orphaned comment
     */
    function createOrphanedCommentElement(comment) {
        const div = document.createElement('div');
        div.className = 'orphaned-comment';
        
        const meta = comment.metadata || {};
        const context = meta.context || {};
        const content = meta.content || '[Content not available]';
        
        div.innerHTML = `
            <div class="orphaned-comment-context">
                <strong>Original paragraph:</strong>
                <blockquote>${escapeHtml(content)}</blockquote>
                ${context['heading-path'] ? `
                    <div class="orphaned-comment-location">
                        Section: ${context['heading-path'].join(' > ')}
                    </div>
                ` : ''}
            </div>
            <div class="comment-item">
                <div class="comment-header">
                    <span class="comment-author">${escapeHtml(comment.author || 'Anonymous')}</span>
                    <span class="comment-date">${formatDate(comment.created)}</span>
                </div>
                <div class="comment-text">${escapeHtml(comment.text)}</div>
                ${comment.replies && comment.replies.length > 0 ? `
                    <div class="comment-replies">
                        ${comment.replies.map(r => createReplyHtml(r)).join('')}
                    </div>
                ` : ''}
            </div>
        `;
        
        return div;
    }

    /**
     * Set up event listeners
     */
    function setupEventListeners() {
        // Event listeners are handled via inline onclick for now
    }

    /**
     * Toggle comments visibility for a paragraph
     */
    window.toggleComments = function(paragraphId) {
        let section = document.getElementById(`comments-${paragraphId}`);
        
        if (section) {
            section.style.display = section.style.display === 'none' ? 'block' : 'none';
            return;
        }
        
        section = createCommentSection(paragraphId);
        
        const wrapper = document.querySelector(`[data-comment-id="${paragraphId}"]`);
        if (wrapper && wrapper.parentNode) {
            wrapper.parentNode.insertBefore(section, wrapper.nextSibling);
        }
    };

    /**
     * Create comment section HTML
     */
    function createCommentSection(paragraphId) {
        const section = document.createElement('div');
        section.id = `comments-${paragraphId}`;
        section.className = 'comment-section';
        
        const comments = currentPageComments
            .filter(c => c.paragraphId === paragraphId)
            .map(c => c.comment);
        
        section.innerHTML = `
            <div class="comment-list">
                ${comments.length > 0 ? 
                    comments.map(c => createCommentHtml(c)).join('') :
                    '<p class="no-comments">No comments yet. Be the first to comment!</p>'
                }
            </div>
            <div class="comment-form">
                ${currentUser ? `
                    <textarea class="comment-input" placeholder="Add a comment..." rows="3"></textarea>
                    <button class="comment-submit" onclick="submitComment('${paragraphId}')">Post Comment</button>
                ` : `
                    <p class="auth-required">Please <a href="#" onclick="document.getElementById('mdbook-auth-button').click(); return false;">sign in</a> to comment.</p>
                `}
            </div>
        `;
        
        return section;
    }

    /**
     * Create HTML for a single comment
     */
    function createCommentHtml(comment) {
        return `
            <div class="comment-item" data-comment-id="${comment.id}">
                <div class="comment-header">
                    <span class="comment-author">${escapeHtml(comment.author || 'Anonymous')}</span>
                    <span class="comment-date">${formatDate(comment.created)}</span>
                </div>
                <div class="comment-text">${escapeHtml(comment.text)}</div>
                ${comment.replies && comment.replies.length > 0 ? `
                    <div class="comment-replies">
                        ${comment.replies.map(r => createReplyHtml(r)).join('')}
                    </div>
                ` : ''}
                ${currentUser ? `
                    <button class="comment-reply-btn" onclick="showReplyForm('${comment.id}')">Reply</button>
                    <div class="reply-form" id="reply-form-${comment.id}" style="display: none;">
                        <textarea class="reply-input" placeholder="Write a reply..." rows="2"></textarea>
                        <button class="reply-submit" onclick="submitReply('${comment.id}')">Post Reply</button>
                    </div>
                ` : ''}
            </div>
        `;
    }

    /**
     * Create HTML for a reply
     */
    function createReplyHtml(reply) {
        return `
            <div class="reply-item">
                <div class="reply-header">
                    <span class="reply-author">${escapeHtml(reply.author || 'Anonymous')}</span>
                    <span class="reply-date">${formatDate(reply.created)}</span>
                </div>
                <div class="reply-text">${escapeHtml(reply.text)}</div>
            </div>
        `;
    }

    /**
     * Show reply form
     */
    window.showReplyForm = function(commentId) {
        const form = document.getElementById(`reply-form-${commentId}`);
        if (form) {
            form.style.display = form.style.display === 'none' ? 'block' : 'none';
        }
    };

    /**
     * Submit a new comment
     */
    window.submitComment = async function(paragraphId) {
        if (!currentUser) {
            alert('Please sign in to comment');
            document.getElementById('mdbook-auth-button').click();
            return;
        }
        
        const wrapper = document.querySelector(`[data-comment-id="${paragraphId}"]`);
        if (!wrapper) return;
        
        const section = document.getElementById(`comments-${paragraphId}`);
        const textarea = section.querySelector('.comment-input');
        const text = textarea.value.trim();
        
        if (!text) return;
        
        const metadata = JSON.parse(wrapper.getAttribute('data-comment-meta') || '{}');
        
        try {
            const { data, error } = await supabaseClient
                .from('comments')
                .insert({
                    paragraph_id: paragraphId,
                    metadata: metadata,
                    author: currentUser.id,
                    text: text,
                    parent_id: null
                })
                .select()
                .single();
            
            if (error) throw error;
            
            // Add user info to comment
            data.author = currentUser.email || currentUser.id.substring(0, 8);
            data.replies = [];
            
            // Add to local state
            allComments.push(data);
            currentPageComments.push({ paragraphId, comment: data, confidence: 1.0 });
            
            // Clear textarea
            textarea.value = '';
            
            // Reload comments display
            section.remove();
            toggleComments(paragraphId);
            
        } catch (error) {
            console.error('Error posting comment:', error);
            alert('Failed to post comment. Please try again.');
        }
    };

    /**
     * Submit a reply to a comment
     */
    window.submitReply = async function(commentId) {
        if (!currentUser) {
            alert('Please sign in to reply');
            document.getElementById('mdbook-auth-button').click();
            return;
        }
        
        const form = document.getElementById(`reply-form-${commentId}`);
        const textarea = form.querySelector('.reply-input');
        const text = textarea.value.trim();
        
        if (!text) return;
        
        // Find parent comment to get paragraph_id and metadata
        const parentComment = allComments.find(c => c.id === commentId);
        if (!parentComment) {
            console.error('Parent comment not found');
            return;
        }
        
        try {
            const { data, error } = await supabaseClient
                .from('comments')
                .insert({
                    paragraph_id: parentComment.paragraph_id,
                    metadata: parentComment.metadata,
                    author: currentUser.id,
                    text: text,
                    parent_id: commentId
                })
                .select()
                .single();
            
            if (error) throw error;
            
            // Add user info
            data.author = currentUser.email || currentUser.id.substring(0, 8);
            
            // Update local comment
            const comment = allComments.find(c => c.id === commentId);
            if (comment) {
                if (!comment.replies) comment.replies = [];
                comment.replies.push(data);
            }
            
            // Add to global list
            allComments.push(data);
            
            // Clear textarea and hide form
            textarea.value = '';
            form.style.display = 'none';
            
            // Reload comments
            const section = form.closest('.comment-section');
            if (section) {
                const paragraphId = section.id.replace('comments-', '');
                section.remove();
                toggleComments(paragraphId);
            }
            
        } catch (error) {
            console.error('Error posting reply:', error);
            alert('Failed to post reply. Please try again.');
        }
    };

    /**
     * Escape HTML to prevent XSS
     */
    function escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }

    /**
     * Format date for display
     */
    function formatDate(dateStr) {
        if (!dateStr) return '';
        const date = new Date(dateStr);
        return date.toLocaleString();
    }

    /**
     * Inject CSS styles
     */
    function injectStyles() {
        if (document.getElementById('mdbook-comments-styles')) return;
        
        const style = document.createElement('style');
        style.id = 'mdbook-comments-styles';
        style.textContent = `
            .comment-link-wrapper {
                display: inline;
                margin-left: 0.5em;
            }
            
            .comment-link {
                font-size: 0.85em;
                color: #0066cc;
                text-decoration: underline;
                cursor: pointer;
            }
            
            .comment-link:hover {
                color: #0052a3;
            }
            
            .comment-section {
                margin: 1em 0;
                padding: 1em;
                background: #f5f5f5;
                border-left: 3px solid #0066cc;
                border-radius: 4px;
            }
            
            .comment-list {
                margin-bottom: 1em;
            }
            
            .comment-item {
                background: white;
                padding: 0.75em;
                margin-bottom: 0.75em;
                border-radius: 4px;
                box-shadow: 0 1px 3px rgba(0,0,0,0.1);
            }
            
            .comment-header {
                display: flex;
                justify-content: space-between;
                margin-bottom: 0.5em;
                font-size: 0.9em;
                color: #666;
            }
            
            .comment-author {
                font-weight: bold;
                color: #333;
            }
            
            .comment-text {
                line-height: 1.5;
                white-space: pre-wrap;
            }
            
            .comment-replies {
                margin-top: 0.75em;
                margin-left: 1.5em;
                border-left: 2px solid #ddd;
                padding-left: 0.75em;
            }
            
            .reply-item {
                background: #fafafa;
                padding: 0.5em;
                margin-bottom: 0.5em;
                border-radius: 3px;
            }
            
            .reply-header {
                display: flex;
                justify-content: space-between;
                margin-bottom: 0.25em;
                font-size: 0.85em;
                color: #666;
            }
            
            .reply-author {
                font-weight: bold;
                color: #333;
            }
            
            .reply-text {
                font-size: 0.95em;
                line-height: 1.4;
                white-space: pre-wrap;
            }
            
            .comment-reply-btn {
                margin-top: 0.5em;
                padding: 0.25em 0.75em;
                font-size: 0.85em;
                background: #f0f0f0;
                border: 1px solid #ddd;
                border-radius: 3px;
                cursor: pointer;
            }
            
            .comment-reply-btn:hover {
                background: #e0e0e0;
            }
            
            .comment-form, .reply-form {
                margin-top: 0.75em;
            }
            
            .comment-input, .reply-input {
                width: 100%;
                padding: 0.5em;
                border: 1px solid #ddd;
                border-radius: 3px;
                font-family: inherit;
                font-size: 0.95em;
                resize: vertical;
            }
            
            .comment-submit, .reply-submit {
                margin-top: 0.5em;
                padding: 0.5em 1em;
                background: #0066cc;
                color: white;
                border: none;
                border-radius: 3px;
                cursor: pointer;
                font-size: 0.95em;
            }
            
            .comment-submit:hover, .reply-submit:hover {
                background: #0052a3;
            }
            
            .no-comments {
                color: #999;
                font-style: italic;
            }
            
            .auth-required {
                color: #666;
                font-style: italic;
            }
            
            .auth-required a {
                color: #0066cc;
                text-decoration: underline;
            }
            
            .orphaned-comments-section {
                margin-top: 3em;
                padding-top: 2em;
                border-top: 2px solid #ddd;
            }
            
            .orphaned-comments-note {
                color: #666;
                font-style: italic;
                margin-bottom: 1.5em;
            }
            
            .orphaned-comment {
                margin-bottom: 2em;
                padding: 1em;
                background: #fff9e6;
                border-left: 3px solid #ffcc00;
                border-radius: 4px;
            }
            
            .orphaned-comment-context {
                margin-bottom: 1em;
                padding-bottom: 1em;
                border-bottom: 1px solid #ddd;
            }
            
            .orphaned-comment-context blockquote {
                margin: 0.5em 0;
                padding: 0.5em;
                background: white;
                border-left: 3px solid #ddd;
                font-style: italic;
            }
            
            .orphaned-comment-location {
                font-size: 0.9em;
                color: #666;
                margin-top: 0.5em;
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
