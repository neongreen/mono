# Complete Deployment Guide: mdbook-comments with Supabase

This guide provides step-by-step instructions to deploy the mdbook-comments plugin with Supabase as the backend. Follow these steps to get a fully working book with commenting functionality.

## Prerequisites

- [Rust and Cargo](https://rustup.rs/) installed
- [mdbook](https://rust-lang.github.io/mdBook/guide/installation.html) installed
- A [Supabase](https://supabase.com/) account (free tier works fine)
- Basic familiarity with the command line

## Part 1: Set Up Supabase Backend

### Step 1: Create a Supabase Project

1. Go to https://supabase.com/ and sign in
2. Click **"New Project"**
3. Fill in:
   - **Project Name**: `mdbook-comments` (or any name you prefer)
   - **Database Password**: Choose a strong password (save it!)
   - **Region**: Choose closest to your users
   - **Pricing Plan**: Free tier is sufficient
4. Click **"Create new project"**
5. Wait 2-3 minutes for the project to be set up

### Step 2: Get Your Supabase Credentials

1. In your Supabase project dashboard, click **"Settings"** (gear icon) in the left sidebar
2. Click **"API"** under Settings
3. Note down these values (you'll need them later):
   - **Project URL**: `https://xxxxxxxxxxxxx.supabase.co`
   - **anon/public key**: `eyJhbGc...` (long string)

### Step 3: Create Database Tables

1. In Supabase dashboard, click **"SQL Editor"** in the left sidebar
2. Click **"New query"**
3. Paste the following SQL and click **"Run"**:

```sql
-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create comments table
CREATE TABLE comments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    paragraph_id TEXT NOT NULL,
    metadata JSONB NOT NULL,
    author TEXT NOT NULL,
    text TEXT NOT NULL,
    created TIMESTAMPTZ DEFAULT NOW(),
    parent_id UUID REFERENCES comments(id) ON DELETE CASCADE,
    CONSTRAINT valid_metadata CHECK (
        metadata ? 'id' AND
        metadata ? 'position' AND
        metadata ? 'content' AND
        metadata ? 'context'
    )
);

-- Create index for faster lookups
CREATE INDEX idx_comments_paragraph_id ON comments(paragraph_id);
CREATE INDEX idx_comments_parent_id ON comments(parent_id);
CREATE INDEX idx_comments_created ON comments(created DESC);

-- Enable Row Level Security (RLS)
ALTER TABLE comments ENABLE ROW LEVEL SECURITY;

-- Policy: Anyone can read comments
CREATE POLICY "Allow public read access"
ON comments FOR SELECT
TO public
USING (true);

-- Policy: Authenticated users can insert comments
CREATE POLICY "Allow authenticated insert"
ON comments FOR INSERT
TO authenticated
WITH CHECK (true);

-- Policy: Users can update their own comments
CREATE POLICY "Allow users to update own comments"
ON comments FOR UPDATE
TO authenticated
USING (auth.uid()::text = author)
WITH CHECK (auth.uid()::text = author);

-- Policy: Users can delete their own comments
CREATE POLICY "Allow users to delete own comments"
ON comments FOR DELETE
TO authenticated
USING (auth.uid()::text = author);
```

4. Verify the table was created:
   - Click **"Table Editor"** in the left sidebar
   - You should see the `comments` table

### Step 4: Set Up Authentication (Google OAuth)

1. In Supabase dashboard, click **"Authentication"** in the left sidebar
2. Click **"Providers"**
3. Find **"Google"** and toggle it **ON**
4. You'll need to set up Google OAuth:

#### Set up Google OAuth Console:

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select existing one
3. Go to **"APIs & Services"** → **"Credentials"**
4. Click **"Create Credentials"** → **"OAuth client ID"**
5. If prompted, configure the OAuth consent screen:
   - User Type: **External**
   - App name: `mdbook-comments` (or your book name)
   - User support email: your email
   - Developer contact: your email
   - Click **"Save and Continue"** through the rest
6. Back at Create OAuth client:
   - Application type: **Web application**
   - Name: `mdbook-comments`
   - **Authorized redirect URIs**: Add your Supabase callback URL
     - Format: `https://xxxxxxxxxxxxx.supabase.co/auth/v1/callback`
     - (Replace with your Project URL from Step 2)
   - Click **"Create"**
7. Copy the **Client ID** and **Client secret**

#### Configure in Supabase:

1. Back in Supabase → Authentication → Providers → Google
2. Paste the **Client ID** and **Client Secret**
3. Click **"Save"**

### Step 5: Set Up CORS (for local development)

1. In Supabase dashboard, go to **"Settings"** → **"API"**
2. Scroll down to **"API Settings"**
3. Under **"Additional Settings"**, add your development URL:
   - For local testing: `http://localhost:3000`
   - For production: your book's domain

## Part 2: Install and Configure mdbook-comments

### Step 6: Install the Preprocessor

```bash
# Navigate to the mdbook-comments directory
cd /path/to/mono/mdbook-comments

# Install the preprocessor
cargo install --path .

# Verify installation
mdbook-comments --version
```

### Step 7: Set Up Your mdbook Project

If you have an existing mdbook project, skip to Step 8. Otherwise:

```bash
# Create a new mdbook project
mkdir my-book
cd my-book
mdbook init

# This creates:
# ├── book.toml
# └── src/
#     ├── SUMMARY.md
#     └── chapter_1.md
```

### Step 8: Configure book.toml

Edit your `book.toml` and add the preprocessor configuration:

```toml
[book]
title = "My Book with Comments"
authors = ["Your Name"]
language = "en"
multilingual = false
src = "src"

[preprocessor.comments]
# Your Supabase project URL
api-url = "https://xxxxxxxxxxxxx.supabase.co"

# Authentication type
auth-type = "supabase"

# Similarity threshold for fuzzy matching
similarity-threshold = 0.85

# Where to show orphaned comments
orphaned-comments-location = "end-of-chapter"

# Commentable elements
[preprocessor.comments.elements]
paragraphs = true
lists = true
blockquotes = true
code-blocks = true
tables = true
headings = false

# UI customization
[preprocessor.comments.ui]
link-text = "comment"
show-comment-count = true

[output.html]
# JavaScript and CSS are now embedded in the preprocessor binary
# No additional-js or additional-css configuration needed
```

**Important**: Replace `https://xxxxxxxxxxxxx.supabase.co` with your actual Supabase Project URL from Step 2.

### Step 9: Configure Supabase in JavaScript

Edit `js/comments.js` and update the configuration at the top:

```javascript
// Find this section (around line 12)
const CONFIG = window.MDBOOK_COMMENTS_CONFIG || {
    apiUrl: 'http://localhost:3000/api',
    similarityThreshold: 0.85,
    orphanedLocation: 'end-of-chapter'
};
```

Replace it with:

```javascript
// Supabase configuration
const SUPABASE_URL = 'https://xxxxxxxxxxxxx.supabase.co';
const SUPABASE_ANON_KEY = 'eyJhbGc...'; // Your anon key from Step 2

// Initialize Supabase client
const { createClient } = supabase;
const supabaseClient = createClient(SUPABASE_URL, SUPABASE_ANON_KEY);

const CONFIG = window.MDBOOK_COMMENTS_CONFIG || {
    apiUrl: SUPABASE_URL,
    similarityThreshold: 0.85,
    orphanedLocation: 'end-of-chapter'
};
```

### Step 11: Add Supabase JavaScript SDK

Edit your `book.toml` to include the Supabase SDK:

```toml
[output.html]
additional-js = [
    "https://cdn.jsdelivr.net/npm/@supabase/supabase-js@2"
]
# JavaScript and CSS are now embedded in the preprocessor binary
# No additional-css configuration needed
```

## Part 3: Adapt JavaScript for Supabase

### Step 12: Update comments.js for Supabase API

Replace the API functions in `js/comments.js`. Find these functions and replace them:

#### Replace `loadComments()`:

```javascript
async function loadComments() {
    try {
        // Fetch all comments from Supabase
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
```

#### Replace `submitComment()`:

```javascript
window.submitComment = async function(paragraphId) {
    const wrapper = document.querySelector(`[data-comment-id="${paragraphId}"]`);
    if (!wrapper) return;
    
    const section = document.getElementById(`comments-${paragraphId}`);
    const textarea = section.querySelector('.comment-input');
    const text = textarea.value.trim();
    
    if (!text) return;
    
    // Check if user is authenticated
    const { data: { user } } = await supabaseClient.auth.getUser();
    if (!user) {
        alert('Please sign in to comment');
        // Trigger authentication
        await supabaseClient.auth.signInWithOAuth({
            provider: 'google',
            options: {
                redirectTo: window.location.href
            }
        });
        return;
    }
    
    const metadata = JSON.parse(wrapper.getAttribute('data-comment-meta') || '{}');
    
    try {
        const { data, error } = await supabaseClient
            .from('comments')
            .insert({
                paragraph_id: paragraphId,
                metadata: metadata,
                author: user.id,
                text: text,
                parent_id: null
            })
            .select()
            .single();
        
        if (error) throw error;
        
        // Add user info to comment
        data.author = user.email || user.id;
        
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
```

#### Replace `submitReply()`:

```javascript
window.submitReply = async function(commentId) {
    const form = document.getElementById(`reply-form-${commentId}`);
    const textarea = form.querySelector('.reply-input');
    const text = textarea.value.trim();
    
    if (!text) return;
    
    // Check if user is authenticated
    const { data: { user } } = await supabaseClient.auth.getUser();
    if (!user) {
        alert('Please sign in to reply');
        await supabaseClient.auth.signInWithOAuth({
            provider: 'google',
            options: {
                redirectTo: window.location.href
            }
        });
        return;
    }
    
    try {
        // Find parent comment to get paragraph_id and metadata
        const parentComment = allComments.find(c => c.id === commentId);
        if (!parentComment) throw new Error('Parent comment not found');
        
        const { data, error } = await supabaseClient
            .from('comments')
            .insert({
                paragraph_id: parentComment.paragraph_id,
                metadata: parentComment.metadata,
                author: user.id,
                text: text,
                parent_id: commentId
            })
            .select()
            .single();
        
        if (error) throw error;
        
        // Add user info
        data.author = user.email || user.id;
        
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
```

#### Add Authentication UI:

Add this function to handle authentication status:

```javascript
async function initAuth() {
    // Check current auth state
    const { data: { user } } = await supabaseClient.auth.getUser();
    
    // Add auth button to page
    const authButton = document.createElement('button');
    authButton.id = 'auth-button';
    authButton.style.cssText = `
        position: fixed;
        top: 10px;
        right: 10px;
        padding: 8px 16px;
        background: #0066cc;
        color: white;
        border: none;
        border-radius: 4px;
        cursor: pointer;
        z-index: 1000;
    `;
    
    if (user) {
        authButton.textContent = `Signed in as ${user.email || 'User'}`;
        authButton.onclick = async () => {
            await supabaseClient.auth.signOut();
            window.location.reload();
        };
    } else {
        authButton.textContent = 'Sign in to comment';
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
    
    // Listen for auth state changes
    supabaseClient.auth.onAuthStateChange((event, session) => {
        if (event === 'SIGNED_IN' || event === 'SIGNED_OUT') {
            window.location.reload();
        }
    });
}
```

Update the `init()` function to call `initAuth()`:

```javascript
function init() {
    console.log('Initializing mdbook-comments...');
    
    // Initialize authentication
    initAuth();
    
    // Load comments for current page
    loadComments();
    
    // Set up event listeners
    setupEventListeners();
    
    // Add styles
    injectStyles();
}
```

## Part 4: Build and Test

### Step 13: Build Your Book

```bash
# In your book directory
mdbook build

# The output will be in book/ directory
```

### Step 14: Test Locally

```bash
# Serve the book locally
mdbook serve

# Or use a simple HTTP server
cd book
python3 -m http.server 3000
```

Open your browser to `http://localhost:3000` and test:

1. You should see "comment" links at the end of paragraphs
2. Click "Sign in to comment" button in the top right
3. Sign in with your Google account
4. Click a "comment" link
5. Add a comment
6. Verify it appears and persists after page reload

## Part 5: Deploy to Production

### Step 15: Deploy Your Book

You can deploy your built book (`book/` directory) to:

#### Option A: GitHub Pages

```bash
# In your book directory
git init
git add book/
git commit -m "Deploy book"
git branch -M gh-pages
git remote add origin https://github.com/username/repo.git
git push -u origin gh-pages
```

Then enable GitHub Pages in your repository settings.

#### Option B: Netlify

1. Create account at https://netlify.com
2. Click "Add new site" → "Deploy manually"
3. Drag and drop your `book/` directory
4. Done!

#### Option C: Vercel

```bash
# Install Vercel CLI
npm install -g vercel

# Deploy
cd book/
vercel
```

### Step 16: Update Supabase CORS Settings

After deploying:

1. Go to Supabase dashboard → Settings → API
2. Under "Additional Settings", add your production URL
3. Example: `https://mybook.netlify.app`

### Step 17: Update Redirect URLs

1. In Google Cloud Console → Credentials → Your OAuth client
2. Add your production URL to "Authorized redirect URIs":
   - Format: `https://xxxxxxxxxxxxx.supabase.co/auth/v1/callback`
3. Also add your book's URL to "Authorized JavaScript origins":
   - Example: `https://mybook.netlify.app`

## Troubleshooting

### Comments not loading

- Check browser console for errors (F12 → Console)
- Verify Supabase URL and anon key are correct
- Check that CORS is configured in Supabase

### Authentication not working

- Verify Google OAuth credentials are correct
- Check redirect URLs include your site
- Ensure `auth.users` table has proper permissions

### Comments not saving

- Check if user is authenticated (look for user ID in console)
- Verify RLS policies are correct in Supabase
- Check Supabase logs: Dashboard → Logs → Database

### Orphaned comments not showing

- This is normal if no content has changed
- Try editing a paragraph significantly and rebuilding

## Advanced Configuration

### Custom Styling

Edit `css/comments.css` to customize appearance:

```css
/* Make comment links more prominent */
.comment-link {
    font-size: 0.9em;
    font-weight: bold;
    color: #0066cc;
}

/* Customize comment sections */
.comment-section {
    background: #f8f9fa;
    border-left: 4px solid #0066cc;
}
```

### Email Notifications

To get email notifications when comments are added:

1. In Supabase dashboard, go to Database → Functions
2. Create a new function to send emails when comments are inserted
3. Or use Supabase's built-in email templates (Authentication → Email Templates)

### Moderation

To enable comment moderation:

1. Add a `status` column to comments table:
   ```sql
   ALTER TABLE comments ADD COLUMN status TEXT DEFAULT 'pending';
   ```
2. Update RLS policies to only show approved comments
3. Create an admin interface to approve/reject comments

## Summary

You now have:
- ✅ Supabase backend with database and authentication
- ✅ mdbook with comment links on every paragraph
- ✅ Google OAuth for user authentication
- ✅ Full comment and reply functionality
- ✅ Orphaned comment handling
- ✅ Production-ready deployment

Your users can now leave comments on any paragraph in your book, and the comments will persist across page reloads and be visible to all readers!

## Support

If you encounter issues:

1. Check the [Supabase documentation](https://supabase.com/docs)
2. Review browser console errors
3. Check Supabase dashboard logs
4. Verify all configuration values are correct

Happy commenting! 📖💬
