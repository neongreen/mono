/**
 * mdbook-comments - Google Sheets backend adapter
 *
 * This adapter implements the backend interface for Google Sheets.
 * It uses Google Identity Services for OAuth2 authentication and
 * Google Sheets API for data storage.
 *
 * Configuration via window.GOOGLE_SHEETS_CONFIG:
 * {
 *   clientId: 'YOUR_CLIENT_ID.apps.googleusercontent.com',
 *   spreadsheetId: 'YOUR_SPREADSHEET_ID',
 *   sheetName: 'Sheet1',
 *   apiKey: '',  // Optional: for read-only access without auth
 *   scopes: 'https://www.googleapis.com/auth/spreadsheets https://www.googleapis.com/auth/userinfo.email'
 * }
 */

(function() {
    'use strict';

    // Configuration (injected or defaults)
    const GOOGLE_CONFIG = window.GOOGLE_SHEETS_CONFIG || {
        clientId: 'YOUR_CLIENT_ID.apps.googleusercontent.com',
        spreadsheetId: 'YOUR_SPREADSHEET_ID',
        sheetName: 'Sheet1',
        apiKey: '',
        scopes: 'https://www.googleapis.com/auth/spreadsheets https://www.googleapis.com/auth/userinfo.email'
    };

    // State
    let currentUser = null;
    let accessToken = null;
    let tokenClient = null;

    /**
     * Generate UUID
     */
    function generateUUID() {
        return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
            const r = Math.random() * 16 | 0;
            const v = c === 'x' ? r : (r & 0x3 | 0x8);
            return v.toString(16);
        });
    }

    /**
     * Get user info from Google
     */
    async function getUserInfo() {
        try {
            const response = await fetch('https://www.googleapis.com/oauth2/v2/userinfo', {
                headers: {
                    'Authorization': `Bearer ${accessToken}`
                }
            });

            if (response.ok) {
                const userInfo = await response.json();
                currentUser = {
                    email: userInfo.email,
                    name: userInfo.name,
                    picture: userInfo.picture
                };
            }
        } catch (error) {
            console.error('Error getting user info:', error);
        }
    }

    /**
     * Sign in with Google
     */
    function signIn() {
        if (tokenClient) {
            tokenClient.requestAccessToken();
        } else {
            throw new Error('Google authentication not initialized');
        }
    }

    /**
     * Parse sheet data into comments
     */
    function parseSheetData(rows) {
        if (!rows || rows.length < 2) {
            return [];
        }

        // First row is headers: id, paragraph_id, metadata, author, text, created, parent_id
        const comments = [];

        for (let i = 1; i < rows.length; i++) {
            const row = rows[i];
            if (!row || row.length < 5) continue;

            try {
                const comment = {
                    id: row[0] || '',
                    paragraph_id: row[1] || '',
                    metadata: row[2] ? JSON.parse(row[2]) : {},
                    author: row[3] || 'Anonymous',
                    text: row[4] || '',
                    created: row[5] || new Date().toISOString(),
                    parent_id: row[6] || null
                };

                comments.push(comment);
            } catch (error) {
                console.error('Error parsing comment row:', error, row);
            }
        }

        return comments;
    }

    /**
     * Append a row to Google Sheet
     */
    async function appendToSheet(values) {
        if (!accessToken) {
            throw new Error('Not authenticated');
        }

        const url = `https://sheets.googleapis.com/v4/spreadsheets/${GOOGLE_CONFIG.spreadsheetId}/values/${GOOGLE_CONFIG.sheetName}:append?valueInputOption=RAW`;

        const response = await fetch(url, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${accessToken}`,
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                values: [values]
            })
        });

        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.error?.message || 'Failed to append to sheet');
        }

        return await response.json();
    }

    /**
     * Google Sheets backend adapter
     */
    const googleSheetsBackend = {
        /**
         * Initialize the backend
         */
        init: async function() {
            console.log('Initializing Google Sheets backend...');

            // Check if Google Sign-In library is loaded
            if (typeof google === 'undefined') {
                console.error('Google Sign-In library not loaded. Add GSI script to book.toml');
                throw new Error('Google Sign-In library not available');
            }

            // Initialize token client for OAuth
            tokenClient = google.accounts.oauth2.initTokenClient({
                client_id: GOOGLE_CONFIG.clientId,
                scope: GOOGLE_CONFIG.scopes,
                callback: async (tokenResponse) => {
                    if (tokenResponse.access_token) {
                        accessToken = tokenResponse.access_token;
                        await getUserInfo();

                        // Notify base module that auth state changed
                        if (window.MdbookComments && window.MdbookComments.onAuthChange) {
                            await window.MdbookComments.onAuthChange();
                        }
                    }
                },
            });

            // Create auth button (will be managed by base module)
            this.createAuthButton();

            // Try to load comments with API key if available (read-only)
            if (GOOGLE_CONFIG.apiKey) {
                try {
                    const url = `https://sheets.googleapis.com/v4/spreadsheets/${GOOGLE_CONFIG.spreadsheetId}/values/${GOOGLE_CONFIG.sheetName}?key=${GOOGLE_CONFIG.apiKey}`;
                    const response = await fetch(url);

                    if (response.ok) {
                        const data = await response.json();
                        return parseSheetData(data.values);
                    }
                } catch (error) {
                    console.log('Could not load with API key:', error);
                }
            }

            return [];
        },

        /**
         * Load all comments from the server
         */
        loadComments: async function() {
            if (!accessToken) {
                console.log('No access token, cannot load comments');
                return [];
            }

            const url = `https://sheets.googleapis.com/v4/spreadsheets/${GOOGLE_CONFIG.spreadsheetId}/values/${GOOGLE_CONFIG.sheetName}`;

            const response = await fetch(url, {
                headers: {
                    'Authorization': `Bearer ${accessToken}`
                }
            });

            if (!response.ok) {
                const error = await response.json();
                throw new Error(`Failed to load comments: ${error.error?.message || response.statusText}`);
            }

            const data = await response.json();
            return parseSheetData(data.values);
        },

        /**
         * Save a new comment
         */
        saveComment: async function(paragraphId, metadata, text, author) {
            if (!currentUser || !accessToken) {
                throw new Error('Please sign in to comment');
            }

            const commentId = generateUUID();
            const actualAuthor = author || currentUser.email || currentUser.name || 'Anonymous';

            await appendToSheet([
                commentId,
                paragraphId,
                JSON.stringify(metadata),
                actualAuthor,
                text,
                new Date().toISOString(),
                '' // parent_id
            ]);

            // Return comment in expected format
            return {
                id: commentId,
                paragraph_id: paragraphId,
                metadata: metadata,
                author: actualAuthor,
                text: text,
                created: new Date().toISOString(),
                parent_id: null
            };
        },

        /**
         * Save a reply to an existing comment
         */
        saveReply: async function(parentCommentId, text, author) {
            if (!currentUser || !accessToken) {
                throw new Error('Please sign in to reply');
            }

            // We need to find the parent comment to get its paragraph_id and metadata
            // This requires loading all comments first - the base module should handle this
            // For now, we'll throw an error if we can't find the parent
            throw new Error('saveReply requires parent comment context - should be handled by base module');
        },

        /**
         * Save a reply with full context (called by base module)
         */
        saveReplyWithContext: async function(parentComment, text, author) {
            if (!currentUser || !accessToken) {
                throw new Error('Please sign in to reply');
            }

            const replyId = generateUUID();
            const actualAuthor = author || currentUser.email || currentUser.name || 'Anonymous';

            await appendToSheet([
                replyId,
                parentComment.paragraph_id,
                JSON.stringify(parentComment.metadata),
                actualAuthor,
                text,
                new Date().toISOString(),
                parentComment.id // parent_id
            ]);

            // Return reply in expected format
            return {
                id: replyId,
                paragraph_id: parentComment.paragraph_id,
                metadata: parentComment.metadata,
                author: actualAuthor,
                text: text,
                created: new Date().toISOString(),
                parent_id: parentComment.id
            };
        },

        /**
         * Get current author name
         */
        getCurrentAuthor: function() {
            return currentUser ? (currentUser.email || currentUser.name) : null;
        },

        /**
         * Check if user is authenticated
         */
        isAuthenticated: function() {
            return !!currentUser && !!accessToken;
        },

        /**
         * Sign in (trigger OAuth flow)
         */
        signIn: function() {
            signIn();
        },

        /**
         * Sign out
         */
        signOut: function() {
            if (accessToken) {
                google.accounts.oauth2.revoke(accessToken, () => {
                    console.log('Token revoked');
                });
                accessToken = null;
                currentUser = null;

                // Reload page to clear state
                window.location.reload();
            }
        },

        /**
         * Create authentication button
         */
        createAuthButton: function() {
            let authButton = document.getElementById('mdbook-auth-button');

            if (!authButton) {
                authButton = document.createElement('button');
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
                document.body.appendChild(authButton);
            }

            this.updateAuthButton();
        },

        /**
         * Update authentication button state
         */
        updateAuthButton: function() {
            const authButton = document.getElementById('mdbook-auth-button');
            if (!authButton) return;

            if (currentUser) {
                authButton.innerHTML = `
                    <span style="margin-right: 8px;">👤</span>
                    ${currentUser.email || currentUser.name || 'User'}
                `.replace(/</g, '&lt;').replace(/>/g, '&gt;');
                authButton.title = 'Click to sign out';
                authButton.onclick = () => {
                    if (confirm('Sign out?')) {
                        this.signOut();
                    }
                };
            } else {
                authButton.innerHTML = `
                    <span style="margin-right: 8px;">🔐</span>
                    Sign in to comment
                `;
                authButton.onclick = () => this.signIn();
            }
        },

        /**
         * Whether to show author input field in forms
         */
        showAuthorInput: false
    };

    // Initialize when DOM is ready
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', () => {
            window.MdbookComments.init(googleSheetsBackend);
        });
    } else {
        window.MdbookComments.init(googleSheetsBackend);
    }

})();
