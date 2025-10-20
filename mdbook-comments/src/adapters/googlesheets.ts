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

import type { BackendAdapter, Comment, ParagraphMetadata } from '../types';

/**
 * Google Sheets configuration
 */
interface GoogleSheetsConfig {
  clientId: string;
  spreadsheetId: string;
  sheetName: string;
  apiKey?: string;
  scopes: string;
}

/**
 * Google user info
 */
interface GoogleUserInfo {
  email: string;
  name: string;
  picture: string;
}

/**
 * Google OAuth2 token response
 */
interface TokenResponse {
  access_token?: string;
  error?: string;
}

/**
 * Google OAuth2 token client
 */
interface TokenClient {
  requestAccessToken(): void;
}

/**
 * Google Sheets API response
 */
interface SheetsApiResponse {
  values?: string[][];
}

/**
 * Google API error response
 */
interface GoogleApiError {
  error?: {
    message?: string;
  };
}

/**
 * Google accounts OAuth2 API
 */
interface GoogleAccountsOAuth2 {
  initTokenClient(config: {
    client_id: string;
    scope: string;
    callback: (response: TokenResponse) => void;
  }): TokenClient;
  revoke(token: string, callback: () => void): void;
}

/**
 * Google Identity Services API
 */
interface Google {
  accounts: {
    oauth2: GoogleAccountsOAuth2;
  };
}

// Global type declarations
declare global {
  interface Window {
    GOOGLE_SHEETS_CONFIG?: GoogleSheetsConfig;
    google?: Google;
  }
}

// Configuration (injected or defaults)
const GOOGLE_CONFIG: GoogleSheetsConfig = window.GOOGLE_SHEETS_CONFIG || {
  clientId: 'YOUR_CLIENT_ID.apps.googleusercontent.com',
  spreadsheetId: 'YOUR_SPREADSHEET_ID',
  sheetName: 'Sheet1',
  apiKey: '',
  scopes: 'https://www.googleapis.com/auth/spreadsheets https://www.googleapis.com/auth/userinfo.email',
};

// State
let currentUser: GoogleUserInfo | null = null;
let accessToken: string | null = null;
let tokenClient: TokenClient | null = null;

/**
 * Generate UUID
 */
function generateUUID(): string {
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, (c) => {
    const r = (Math.random() * 16) | 0;
    const v = c === 'x' ? r : (r & 0x3) | 0x8;
    return v.toString(16);
  });
}

/**
 * Get user info from Google
 */
async function getUserInfo(): Promise<void> {
  if (!accessToken) {
    return;
  }

  try {
    const response = await fetch('https://www.googleapis.com/oauth2/v2/userinfo', {
      headers: {
        Authorization: `Bearer ${accessToken}`,
      },
    });

    if (response.ok) {
      const userInfo = (await response.json()) as GoogleUserInfo;
      currentUser = {
        email: userInfo.email,
        name: userInfo.name,
        picture: userInfo.picture,
      };
    }
  } catch (error) {
    console.error('Error getting user info:', error);
  }
}

/**
 * Sign in with Google
 */
function signIn(): void {
  if (tokenClient) {
    tokenClient.requestAccessToken();
  } else {
    throw new Error('Google authentication not initialized');
  }
}

/**
 * Parse sheet data into comments
 */
function parseSheetData(rows: string[][] | undefined): Comment[] {
  if (!rows || rows.length < 2) {
    return [];
  }

  // First row is headers: id, paragraph_id, metadata, author, text, created, parent_id
  const comments: Comment[] = [];

  for (let i = 1; i < rows.length; i++) {
    const row = rows[i];
    if (!row || row.length < 5) continue;

    try {
      const comment: Comment = {
        id: row[0] || '',
        paragraph_id: row[1] || '',
        metadata: row[2] ? (JSON.parse(row[2]) as ParagraphMetadata) : ({} as ParagraphMetadata),
        author: row[3] || 'Anonymous',
        text: row[4] || '',
        created: row[5] || new Date().toISOString(),
        parent_id: row[6] || null,
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
async function appendToSheet(values: string[]): Promise<SheetsApiResponse> {
  if (!accessToken) {
    throw new Error('Not authenticated');
  }

  const url = `https://sheets.googleapis.com/v4/spreadsheets/${GOOGLE_CONFIG.spreadsheetId}/values/${GOOGLE_CONFIG.sheetName}:append?valueInputOption=RAW`;

  const response = await fetch(url, {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${accessToken}`,
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      values: [values],
    }),
  });

  if (!response.ok) {
    const error = (await response.json()) as GoogleApiError;
    throw new Error(error.error?.message || 'Failed to append to sheet');
  }

  return (await response.json()) as SheetsApiResponse;
}

/**
 * Google Sheets backend adapter
 */
const googleSheetsBackend: BackendAdapter = {
  /**
   * Initialize the backend
   */
  async init(): Promise<void> {
    console.log('Initializing Google Sheets backend...');

    // Check if Google Sign-In library is loaded
    if (typeof window.google === 'undefined') {
      console.error('Google Sign-In library not loaded. Add GSI script to book.toml');
      throw new Error('Google Sign-In library not available');
    }

    // Initialize token client for OAuth
    tokenClient = window.google.accounts.oauth2.initTokenClient({
      client_id: GOOGLE_CONFIG.clientId,
      scope: GOOGLE_CONFIG.scopes,
      callback: async (tokenResponse: TokenResponse) => {
        if (tokenResponse.access_token) {
          accessToken = tokenResponse.access_token;
          await getUserInfo();

          // Update auth button to reflect new state
          updateAuthButton();

          // Notify base module that auth state changed
          const mdbook = window.MdbookComments as any;
          if (mdbook?.onAuthChange) {
            await mdbook.onAuthChange();
          }
        }
      },
    });

    // Create auth button (will be managed by base module)
    createAuthButton();

    // Try to load comments with API key if available (read-only)
    if (GOOGLE_CONFIG.apiKey) {
      try {
        const url = `https://sheets.googleapis.com/v4/spreadsheets/${GOOGLE_CONFIG.spreadsheetId}/values/${GOOGLE_CONFIG.sheetName}?key=${GOOGLE_CONFIG.apiKey}`;
        const response = await fetch(url);

        if (response.ok) {
          // Note: We can't return comments from init() as the interface requires Promise<void>
          // The initial comments will be loaded via loadComments()
          await response.json();
        }
      } catch (error) {
        console.log('Could not load with API key:', error);
      }
    }
  },

  /**
   * Load all comments from the server
   */
  async loadComments(): Promise<Comment[]> {
    if (!accessToken) {
      console.log('No access token, cannot load comments');
      return [];
    }

    const url = `https://sheets.googleapis.com/v4/spreadsheets/${GOOGLE_CONFIG.spreadsheetId}/values/${GOOGLE_CONFIG.sheetName}`;

    const response = await fetch(url, {
      headers: {
        Authorization: `Bearer ${accessToken}`,
      },
    });

    if (!response.ok) {
      const error = (await response.json()) as GoogleApiError;
      throw new Error(`Failed to load comments: ${error.error?.message || response.statusText}`);
    }

    const data = (await response.json()) as SheetsApiResponse;
    return parseSheetData(data.values);
  },

  /**
   * Save a new comment
   */
  async saveComment(
    paragraphId: string,
    metadata: ParagraphMetadata,
    text: string,
    author: string
  ): Promise<Comment> {
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
      '', // parent_id
    ]);

    // Return comment in expected format
    return {
      id: commentId,
      paragraph_id: paragraphId,
      metadata: metadata,
      author: actualAuthor,
      text: text,
      created: new Date().toISOString(),
      parent_id: null,
    };
  },

  /**
   * Save a reply to an existing comment
   */
  async saveReply(parentCommentId: string, text: string, author: string): Promise<Comment> {
    if (!currentUser || !accessToken) {
      throw new Error('Please sign in to reply');
    }

    // Load all comments to find the parent
    const allComments = await this.loadComments();
    const parentComment = allComments.find((c) => c.id === parentCommentId);

    if (!parentComment) {
      throw new Error('Parent comment not found');
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
      parentComment.id, // parent_id
    ]);

    // Return reply in expected format
    return {
      id: replyId,
      paragraph_id: parentComment.paragraph_id,
      metadata: parentComment.metadata,
      author: actualAuthor,
      text: text,
      created: new Date().toISOString(),
      parent_id: parentComment.id,
    };
  },

  /**
   * Get current author name
   */
  getCurrentAuthor(): string | null {
    return currentUser ? currentUser.email || currentUser.name : null;
  },

  /**
   * Check if user is authenticated
   */
  isAuthenticated(): boolean {
    return !!currentUser && !!accessToken;
  },

  /**
   * Sign in (trigger OAuth flow)
   */
  signIn(): void {
    signIn();
  },

  /**
   * Sign out
   */
  signOut(): void {
    if (accessToken && window.google) {
      window.google.accounts.oauth2.revoke(accessToken, () => {
        console.log('Token revoked');
      });
      accessToken = null;
      currentUser = null;

      // Reload page to clear state
      window.location.reload();
    }
  },

  /**
   * Whether to show author input field in forms
   */
  showAuthorInput: false,
};

/**
 * Escape HTML to prevent XSS
 */
function escapeHtml(text: string): string {
  return text.replace(/</g, '&lt;').replace(/>/g, '&gt;');
}

/**
 * Create authentication button
 */
function createAuthButton(): void {
  let authButton = document.getElementById('mdbook-auth-button') as HTMLButtonElement | null;

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

  updateAuthButton();
}

/**
 * Update authentication button state
 */
function updateAuthButton(): void {
  const authButton = document.getElementById('mdbook-auth-button') as HTMLButtonElement | null;
  if (!authButton) return;

  if (currentUser) {
    authButton.innerHTML = `
      <span style="margin-right: 8px;">👤</span>
      ${escapeHtml(currentUser.email || currentUser.name || 'User')}
    `;
    authButton.title = 'Click to sign out';
    authButton.onclick = () => {
      if (confirm('Sign out?')) {
        googleSheetsBackend.signOut?.();
      }
    };
  } else {
    authButton.innerHTML = `
      <span style="margin-right: 8px;">🔐</span>
      Sign in to comment
    `;
    authButton.onclick = () => googleSheetsBackend.signIn?.();
  }
}

// Initialize when DOM is ready
if (document.readyState === 'loading') {
  document.addEventListener('DOMContentLoaded', () => {
    (window.MdbookComments as any)?.init(googleSheetsBackend);
  });
} else {
  (window.MdbookComments as any)?.init(googleSheetsBackend);
}

// Export for testing and use
export { googleSheetsBackend };
