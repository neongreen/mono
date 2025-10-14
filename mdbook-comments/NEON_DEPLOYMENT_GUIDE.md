# Neon Database Deployment Guide: mdbook-comments

This guide provides step-by-step instructions to deploy the mdbook-comments plugin with Neon as the backend database. Neon is a serverless Postgres database with a modern developer experience.

## Overview

Using Neon as a backend offers:
- ✅ Serverless PostgreSQL database
- ✅ Generous free tier (512 MB storage, 1 compute hour)
- ✅ Instant database provisioning
- ✅ Built-in connection pooling
- ✅ Branching for development/testing
- ✅ Auto-scaling and auto-suspend
- ✅ HTTP API for direct database queries

**Advantages:**
- Fast performance (similar to Supabase)
- Simple PostgreSQL - no custom API to learn
- Generous free tier
- Built-in Data API for direct HTTP access
- Great for both development and production

## Prerequisites

- [Rust and Cargo](https://rustup.rs/) installed
- [mdbook](https://rust-lang.github.io/mdBook/guide/installation.html) installed
- A [Neon](https://neon.tech/) account (free tier works fine)
- Basic familiarity with the command line
- Basic SQL knowledge

## Part 1: Set Up Neon Database

### Step 1: Create a Neon Account

1. Go to https://neon.tech/
2. Click **"Sign Up"**
3. Sign up with GitHub, Google, or email
4. Verify your email if needed

### Step 2: Create a Neon Project

1. After signing in, click **"New Project"**
2. Fill in:
   - **Project Name**: `mdbook-comments` (or any name)
   - **Region**: Choose closest to your users (e.g., US East, EU Central)
   - **PostgreSQL Version**: 16 (recommended, latest)
3. Click **"Create Project"**
4. Wait a few seconds - Neon will provision your database instantly

### Step 3: Get Your Connection Details

After the project is created, you'll see the connection details. Save these:

1. **Connection String**: 
   ```
   postgresql://[user]:[password]@[endpoint].neon.tech/neondb?sslmode=require
   ```

2. **API Key** (for Data API):
   - Go to **"Settings"** → **"API Keys"**
   - Click **"Generate New Key"**
   - Name it `mdbook-comments-api`
   - Copy the key (starts with `neon_api_key_...`)
   - Save it securely - you won't be able to see it again

3. **Project ID** and **Endpoint ID**:
   - Found in project settings
   - Used for constructing API URLs

### Step 4: Enable Data API

Neon provides a Data API that allows you to query your database via HTTP requests:

1. Go to your project dashboard
2. Click **"Settings"** → **"Beta Features"**
3. Enable **"Data API"** (if not already enabled)
4. Note the API endpoint URL:
   ```
   https://[endpoint].neon.tech/v2/query
   ```

### Step 5: Create Database Tables

You have two options to create tables:

#### Option A: Using Neon SQL Editor (Recommended)

1. In your Neon project, click **"SQL Editor"** in the left sidebar
2. Paste the following SQL:

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

-- Create indexes for faster lookups
CREATE INDEX idx_comments_paragraph_id ON comments(paragraph_id);
CREATE INDEX idx_comments_parent_id ON comments(parent_id);
CREATE INDEX idx_comments_created ON comments(created DESC);

-- Add comment for documentation
COMMENT ON TABLE comments IS 'Stores mdbook paragraph-level comments';
COMMENT ON COLUMN comments.paragraph_id IS 'Unique identifier for the paragraph';
COMMENT ON COLUMN comments.metadata IS 'Paragraph context (position, content, headings)';
```

3. Click **"Run"** to execute the SQL
4. Verify the table was created:
   - Click **"Tables"** in the left sidebar
   - You should see the `comments` table

#### Option B: Using psql or Any PostgreSQL Client

```bash
# Install psql if needed (on macOS)
brew install postgresql

# Connect using your connection string
psql "postgresql://[user]:[password]@[endpoint].neon.tech/neondb?sslmode=require"

# Paste and run the SQL above
```

### Step 6: Test the Database

Test that you can query the database:

```sql
-- Insert a test comment
INSERT INTO comments (paragraph_id, metadata, author, text)
VALUES (
    'test-paragraph-1',
    '{"id": "test-paragraph-1", "position": {"file": "test.md", "block-index": 1}, "content": "Test content", "context": {"prev": null, "next": null, "heading-path": ["Test"]}}',
    'test@example.com',
    'This is a test comment'
);

-- Query comments
SELECT * FROM comments;

-- Delete test comment
DELETE FROM comments WHERE paragraph_id = 'test-paragraph-1';
```

## Part 2: Install and Configure mdbook-comments

### Step 7: Install the Preprocessor

```bash
# Navigate to the mdbook-comments directory
cd /path/to/mono/mdbook-comments

# Install the preprocessor
cargo install --path .

# Verify installation
mdbook-comments --version
```

### Step 8: Set Up Your mdbook Project

If you have an existing mdbook project, skip to Step 9. Otherwise:

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

### Step 9: Configure book.toml

Edit your `book.toml` and add the preprocessor configuration:

```toml
[book]
title = "My Book with Comments"
authors = ["Your Name"]
language = "en"
multilingual = false
src = "src"

[preprocessor.comments]
# Backend type
backend = "neon"

# Your Neon Data API endpoint
api-url = "https://[endpoint].neon.tech/v2/query"

# Authentication type
auth-type = "api-key"

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
# Add JavaScript and CSS for comments
additional-js = ["js/comments-neon.js"]
additional-css = ["css/comments.css"]
```

**Important**: Replace `[endpoint]` with your actual Neon endpoint (e.g., `ep-cool-darkness-123456.us-east-2.aws.neon.tech`).

### Step 10: Copy Comment Assets

Copy the JavaScript and CSS files to your book directory:

```bash
# Create directories if they don't exist
mkdir -p js css

# Copy the Neon JavaScript client
cp /path/to/mono/mdbook-comments/js/comments-neon.js js/

# Copy the comment CSS
cp /path/to/mono/mdbook-comments/css/comments.css css/
```

### Step 11: Configure Neon in JavaScript

Edit `js/comments-neon.js` and update the configuration at the top:

```javascript
// Find this section (around line 16)
const NEON_CONFIG = {
    apiUrl: 'YOUR_NEON_API_URL',
    apiKey: 'YOUR_NEON_API_KEY',
    database: 'neondb',
    authProvider: null
};
```

Replace with your actual values:

```javascript
const NEON_CONFIG = {
    apiUrl: 'https://[endpoint].neon.tech/v2/query',
    apiKey: 'neon_api_key_xxxxxxxxxxxxxxxxxxxxx', // Your API key from Step 3
    database: 'neondb',
    authProvider: null
};
```

## Part 3: Implement Authentication

Neon doesn't provide built-in authentication like Supabase. You need to implement your own auth system. Here are a few options:

### Option A: Simple JWT Authentication (Recommended for MVP)

For a simple implementation, you can use a lightweight auth service:

1. **Create an auth service** (e.g., using Cloudflare Workers or Vercel Edge Functions):

```javascript
// auth-endpoint.js - Deploy as edge function
export default async function handler(req) {
    if (req.method === 'POST') {
        const { email, password } = await req.json();
        
        // Verify credentials (implement your logic)
        if (isValidUser(email, password)) {
            // Generate JWT token
            const token = generateJWT({ email, id: generateUserId(email) });
            return new Response(JSON.stringify({ token, user: { email } }), {
                headers: { 'Content-Type': 'application/json' }
            });
        }
    }
    
    return new Response('Unauthorized', { status: 401 });
}
```

2. **Update comments-neon.js** to use your auth endpoint:

```javascript
async function authenticate(email, password) {
    const response = await fetch('https://your-auth-endpoint.com/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, password })
    });
    
    if (!response.ok) {
        throw new Error('Authentication failed');
    }
    
    const data = await response.json();
    return data.token;
}
```

### Option B: OAuth with Auth0 or Clerk

For production use, integrate a third-party auth provider:

1. **Sign up** for [Auth0](https://auth0.com/) or [Clerk](https://clerk.dev/)
2. **Configure OAuth** application
3. **Update comments-neon.js** to use the provider's SDK:

```javascript
// Example with Auth0
import { Auth0Client } from '@auth0/auth0-spa-js';

const auth0 = new Auth0Client({
    domain: 'your-domain.auth0.com',
    client_id: 'your-client-id',
    redirect_uri: window.location.origin
});

async function authenticate() {
    await auth0.loginWithPopup();
    const user = await auth0.getUser();
    const token = await auth0.getTokenSilently();
    return { user, token };
}
```

### Option C: Email-Only (No Password)

For internal documentation, you might use email-only auth:

```javascript
// Simple email-only implementation
function createAuthButton(user) {
    authButton.onclick = async () => {
        const email = prompt('Enter your work email:');
        if (email && email.endsWith('@yourcompany.com')) {
            currentUser = { email, id: email };
            localStorage.setItem('neon_user', JSON.stringify(currentUser));
            window.location.reload();
        } else {
            alert('Please use your company email address');
        }
    };
}
```

For this guide, we'll use **Option C** (email-only) for simplicity. Update `comments-neon.js`:

```javascript
// Replace the authenticate and createAuthButton functions
async function initAuth() {
    try {
        const userJson = localStorage.getItem('neon_user');
        if (userJson) {
            currentUser = JSON.parse(userJson);
        }
        createAuthButton(currentUser);
    } catch (error) {
        console.error('Error initializing auth:', error);
    }
}

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
        authButton.innerHTML = `
            <span style="margin-right: 8px;">👤</span>
            ${escapeHtml(user.email)}
        `;
        authButton.title = 'Click to sign out';
        authButton.onclick = () => {
            if (confirm('Sign out?')) {
                localStorage.removeItem('neon_user');
                window.location.reload();
            }
        };
    } else {
        authButton.innerHTML = `
            <span style="margin-right: 8px;">🔐</span>
            Sign in to comment
        `;
        authButton.onclick = () => {
            const email = prompt('Enter your email to comment:');
            if (email && email.includes('@')) {
                currentUser = { email, id: email };
                localStorage.setItem('neon_user', JSON.stringify(currentUser));
                window.location.reload();
            } else if (email) {
                alert('Please enter a valid email address');
            }
        };
    }
    
    document.body.appendChild(authButton);
}
```

## Part 4: Build and Test

### Step 12: Build Your Book

```bash
# In your book directory
mdbook build

# The output will be in book/ directory
```

### Step 13: Test Locally

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
3. Enter your email address
4. Click a "comment" link
5. Add a comment
6. Verify it appears and persists after page reload

### Step 14: Verify in Neon Console

1. Open your Neon project
2. Go to **"SQL Editor"**
3. Run: `SELECT * FROM comments;`
4. You should see your test comments

## Part 5: Deploy to Production

### Step 15: Deploy Your Book

Deploy your built book (`book/` directory) to:

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

### Step 16: Configure CORS (Important!)

By default, Neon allows connections from any origin. However, for security in production:

1. Consider using **connection pooling** with Neon's built-in proxy
2. Implement **API key rotation** regularly
3. Use **environment variables** for sensitive data
4. Consider a **backend proxy** to hide your API key

**Recommended Setup**: Create a simple backend proxy:

```javascript
// backend-proxy.js - Deploy as serverless function
export default async function handler(req) {
    const NEON_API_KEY = process.env.NEON_API_KEY;
    const NEON_ENDPOINT = process.env.NEON_ENDPOINT;
    
    // Only allow specific origins
    const origin = req.headers.get('origin');
    if (!origin.endsWith('.yourdomain.com')) {
        return new Response('Forbidden', { status: 403 });
    }
    
    // Forward query to Neon
    const response = await fetch(`${NEON_ENDPOINT}/v2/query`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${NEON_API_KEY}`
        },
        body: req.body
    });
    
    return response;
}
```

Then update your book to use this proxy instead of direct Neon access.

## Part 6: Managing Comments

### Viewing All Comments

Option 1: **Neon SQL Editor**
```sql
-- View all comments
SELECT 
    author,
    text,
    created,
    metadata->>'id' as paragraph_id
FROM comments
ORDER BY created DESC;

-- View comments by paragraph
SELECT * FROM comments 
WHERE paragraph_id = 'chapter-1-block-3-abc12345'
ORDER BY created ASC;

-- Count comments per paragraph
SELECT 
    paragraph_id,
    COUNT(*) as comment_count
FROM comments
GROUP BY paragraph_id
ORDER BY comment_count DESC;
```

Option 2: **Export to CSV**
```sql
-- In psql or Neon SQL Editor
\copy (SELECT * FROM comments) TO 'comments.csv' WITH CSV HEADER
```

### Moderating Comments

Delete a comment:
```sql
DELETE FROM comments WHERE id = 'comment-uuid-here';
```

Edit a comment:
```sql
UPDATE comments 
SET text = 'Updated comment text'
WHERE id = 'comment-uuid-here';
```

### Backing Up Comments

Neon provides automatic backups, but you can also:

1. **Export database**:
   ```bash
   pg_dump "postgresql://[user]:[password]@[endpoint].neon.tech/neondb" > backup.sql
   ```

2. **Restore from backup**:
   ```bash
   psql "postgresql://[user]:[password]@[endpoint].neon.tech/neondb" < backup.sql
   ```

3. **Use Neon's branching feature**:
   - Go to **"Branches"** in Neon console
   - Create a branch for testing or backups

## Troubleshooting

### "Failed to load comments"

- Check that your Neon API key is correct in `comments-neon.js`
- Verify the endpoint URL is correct
- Check Neon project is not suspended (free tier auto-suspends after inactivity)
- Look at browser console for specific error messages

### "Failed to post comment"

- Verify the user is "authenticated" (email entered)
- Check that the `comments` table exists
- Verify API key has write permissions
- Check Neon project logs in the console

### Performance Issues

Neon is fast, but if you experience slowness:
- Ensure indexes are created (see Step 5)
- Use Neon's read replicas for heavy read workloads
- Consider caching comments in browser localStorage
- Check if you're on free tier and hitting compute limits

### Database Suspended

Free tier databases auto-suspend after 5 minutes of inactivity:
- First query after suspension takes 2-3 seconds to wake up
- Subsequent queries are fast
- Upgrade to paid tier for always-on compute

### API Key Exposed

If you accidentally commit your API key:
1. Go to Neon console → Settings → API Keys
2. Delete the compromised key
3. Generate a new key
4. Update your application
5. Rotate any other affected credentials

## Advanced Configuration

### Read Replicas

For high-traffic sites, use Neon's read replicas:

```javascript
const NEON_CONFIG = {
    writeApiUrl: 'https://[endpoint].neon.tech/v2/query',
    readApiUrl: 'https://[endpoint]-read.neon.tech/v2/query',
    apiKey: 'neon_api_key_xxxxxxxxxxxxxxxxxxxxx'
};

// Use read replica for loading comments
async function loadComments() {
    const rows = await executeQuery(sql, [], NEON_CONFIG.readApiUrl);
    // ...
}

// Use primary for writes
async function submitComment() {
    const rows = await executeQuery(sql, params, NEON_CONFIG.writeApiUrl);
    // ...
}
```

### Database Branching

Use branches for development:

1. Create a branch in Neon console: `dev-branch`
2. Get the branch connection string
3. Use it in development environment
4. Test changes without affecting production
5. Delete branch when done

### Connection Pooling

For high concurrency, use Neon's built-in pooler:

```javascript
const NEON_CONFIG = {
    // Use pooled connection endpoint
    apiUrl: 'https://[endpoint]-pooler.neon.tech/v2/query',
    apiKey: 'neon_api_key_xxxxxxxxxxxxxxxxxxxxx'
};
```

### Monitoring and Analytics

Track comment metrics:

```sql
-- Daily comment counts
SELECT 
    DATE(created) as date,
    COUNT(*) as comments
FROM comments
GROUP BY DATE(created)
ORDER BY date DESC;

-- Most commented paragraphs
SELECT 
    paragraph_id,
    metadata->>'content' as content,
    COUNT(*) as comment_count
FROM comments
GROUP BY paragraph_id, metadata->>'content'
ORDER BY comment_count DESC
LIMIT 10;

-- Active commenters
SELECT 
    author,
    COUNT(*) as comment_count,
    MAX(created) as last_comment
FROM comments
GROUP BY author
ORDER BY comment_count DESC;
```

## Cost Considerations

### Free Tier

Neon's free tier includes:
- 512 MB storage per project
- 1 compute hour per month (shared compute)
- Auto-suspend after 5 minutes of inactivity
- Unlimited projects

**Perfect for:**
- Personal documentation
- Small team wikis
- Learning/experimentation
- Low-traffic sites (<1000 comments)

### Paid Tiers

If you need more:
- **Launch**: $19/month
  - 10 GB storage
  - Always-on compute
  - Read replicas
  - Point-in-time restore
  
- **Scale**: $69/month
  - 50 GB storage
  - Autoscaling compute
  - Multiple read replicas
  - Advanced monitoring

### When to Upgrade

Upgrade when:
- Storage exceeds 500 MB
- Compute hours exceed limit
- Need always-on database (no cold starts)
- Need read replicas for performance
- Need longer data retention

## Comparison with Other Backends

| Feature | Neon | Supabase | Google Sheets |
|---------|------|----------|---------------|
| **Setup** | Easy | Medium | Easy |
| **Performance** | Fast (50-100ms) | Fast (50-100ms) | Slow (1-3s) |
| **Free Tier** | 512 MB, 1 compute hour | 500 MB, 2GB bandwidth | Unlimited |
| **Scalability** | High | High | Low |
| **Auth Built-in** | No | Yes | Yes (Google) |
| **Management UI** | SQL Editor | Dashboard | Spreadsheet |
| **Branching** | Yes | Preview branches | No |
| **Auto-suspend** | Yes (free tier) | No | N/A |
| **Best For** | PostgreSQL fans, high performance | Full-featured apps | Simple internal docs |

## Summary

You now have:
- ✅ Serverless PostgreSQL database with Neon
- ✅ Fast comment loading and posting
- ✅ mdbook with comment links on every paragraph
- ✅ Simple authentication system
- ✅ Production-ready deployment
- ✅ Automatic backups and branching

Your users can now leave comments on any paragraph, and all comments are stored in a fast, scalable PostgreSQL database!

## Support

If you encounter issues:

1. Check [Neon documentation](https://neon.tech/docs)
2. Review browser console errors
3. Check Neon project logs in the console
4. Verify connection strings and API keys
5. Test queries in Neon SQL Editor

Happy commenting! 📖💬⚡
