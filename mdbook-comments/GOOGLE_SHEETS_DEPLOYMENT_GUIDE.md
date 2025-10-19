# Google Sheets Backend Deployment Guide: mdbook-comments

This guide provides step-by-step instructions to deploy the mdbook-comments plugin using Google Sheets as the backend database. This is a lightweight alternative to Supabase that requires no database setup.

## Overview

Using Google Sheets as a backend offers:
- ✅ No database setup required
- ✅ Free (within Google API quotas)
- ✅ Simple authentication via Google
- ✅ Easy to view/manage comments in a spreadsheet
- ✅ No server maintenance

**Limitations:**
- Slower than a traditional database
- Google API rate limits (100 requests per 100 seconds per user)
- Not ideal for high-traffic sites
- Best for internal company handbooks with <1000 comments

## Prerequisites

- [Rust and Cargo](https://rustup.rs/) installed
- [mdbook](https://rust-lang.github.io/mdBook/guide/installation.html) installed
- A Google account
- Basic familiarity with the command line

## Part 1: Set Up Google Sheets Backend

### Step 1: Create a Google Cloud Project

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Click **"Select a project"** → **"New Project"**
3. Enter:
   - **Project Name**: `mdbook-comments` (or any name)
   - **Location**: Your organization (or "No organization")
4. Click **"Create"**
5. Wait for the project to be created (takes a few seconds)

### Step 2: Enable Google Sheets API

1. In the Google Cloud Console, ensure your project is selected
2. Go to **"APIs & Services"** → **"Library"**
3. Search for **"Google Sheets API"**
4. Click on it and click **"Enable"**
5. Also search for **"Google Drive API"** and enable it

### Step 3: Create OAuth 2.0 Credentials

1. Go to **"APIs & Services"** → **"Credentials"**
2. Click **"Create Credentials"** → **"OAuth client ID"**
3. If prompted to configure consent screen:
   - Click **"Configure Consent Screen"**
   - Select **"Internal"** (if using Google Workspace) or **"External"**
   - Fill in:
     - **App name**: `mdbook-comments`
     - **User support email**: your email
     - **Developer contact**: your email
   - Click **"Save and Continue"** through the rest
   - On "Scopes", click **"Add or Remove Scopes"**
   - Add these scopes:
     - `https://www.googleapis.com/auth/spreadsheets`
     - `https://www.googleapis.com/auth/userinfo.email`
   - Click **"Update"** and **"Save and Continue"**
4. Back at "Create OAuth client ID":
   - Application type: **"Web application"**
   - Name: `mdbook-comments`
   - **Authorized JavaScript origins**: Add your site URL
     - For local: `http://localhost:3000`
     - For production: `https://yourbook.com`
   - **Authorized redirect URIs**: Add
     - For local: `http://localhost:3000`
     - For production: `https://yourbook.com`
   - Click **"Create"**
5. Copy the **Client ID** (looks like `xxx.apps.googleusercontent.com`)
6. Download the JSON (optional, but save the Client ID)

### Step 4: Create a Google Sheet

1. Go to [Google Sheets](https://sheets.google.com)
2. Click **"Blank"** to create a new spreadsheet
3. Rename it to **"mdbook-comments"**
4. Set up columns in the first row:
   - **A1**: `id`
   - **B1**: `paragraph_id`
   - **C1**: `metadata`
   - **D1**: `author`
   - **E1**: `text`
   - **F1**: `created`
   - **G1**: `parent_id`
5. Note the **Sheet ID** from the URL:
   - URL format: `https://docs.google.com/spreadsheets/d/{SHEET_ID}/edit`
   - Example: If URL is `https://docs.google.com/spreadsheets/d/1ABC...XYZ/edit`
   - Your Sheet ID is: `1ABC...XYZ`

### Step 5: Share the Sheet (Important!)

1. Click **"Share"** button in the top right
2. Under "General access", select:
   - **"Anyone with the link"** can **"Edit"**
   - OR add specific people's emails who should access
3. Click **"Done"**

**Note**: For a private company handbook, share with your organization only.

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

Edit your `book.toml` and add:

```toml
[book]
title = "My Book with Comments"
authors = ["Your Name"]
language = "en"
multilingual = false
src = "src"

[preprocessor.comments]
# Backend type: "google-sheets"
backend = "google-sheets"

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
# Add Google Sign-In client (comments JS/CSS are now embedded)
additional-js = [
    "https://accounts.google.com/gsi/client"
]
# JavaScript and CSS are now embedded in the preprocessor binary
# No additional-css configuration needed
```

### Step 9: Configure Google Sheets in JavaScript

Edit `js/comments-googlesheets.js` and update the configuration at the top:

```javascript
// Find this section (around line 12)
const GOOGLE_CONFIG = {
    clientId: 'YOUR_CLIENT_ID.apps.googleusercontent.com',
    spreadsheetId: 'YOUR_SPREADSHEET_ID',
    sheetName: 'Sheet1', // or whatever you named your sheet
    apiKey: 'YOUR_API_KEY', // Optional, for read-only access without auth
};
```

Replace:
- `YOUR_CLIENT_ID` with your OAuth 2.0 Client ID from Step 3
- `YOUR_SPREADSHEET_ID` with your Sheet ID from Step 4

**Optional API Key** (for read-only access without sign-in):

1. Go to **"APIs & Services"** → **"Credentials"**
2. Click **"Create Credentials"** → **"API Key"**
3. Copy the key and add it to the config
4. Click **"Edit API key"** and restrict it:
   - **API restrictions**: Select "Restrict key"
   - Check "Google Sheets API"
   - Click **"Save"**

## Part 3: How Google Sheets Backend Works

### Data Structure

Comments are stored as rows in the Google Sheet:

| id | paragraph_id | metadata | author | text | created | parent_id |
|----|--------------|----------|--------|------|---------|-----------|
| uuid-1 | ch1-p1-abc | {...} | user@example.com | Great! | 2024-01-15T10:00:00Z | |
| uuid-2 | ch1-p1-abc | {...} | other@example.com | Agreed | 2024-01-15T11:00:00Z | uuid-1 |

### API Operations

The JavaScript uses Google Sheets API v4 to:

1. **Read comments**: `spreadsheets.values.get`
   - Fetches all rows from the sheet
   - Parses them into comment objects
   - Builds reply structure

2. **Add comment**: `spreadsheets.values.append`
   - Appends a new row to the sheet
   - Generates UUID for comment ID
   - Stores metadata as JSON string

3. **Update comment** (for replies): `spreadsheets.values.update`
   - Updates specific rows
   - Links replies via parent_id

### Authentication Flow

1. User clicks "Sign in to comment"
2. Google Sign-In popup appears
3. User authorizes the app
4. Access token is obtained
5. App can now read/write to the sheet on user's behalf

## Part 4: Build and Test

### Step 11: Build Your Book

```bash
# In your book directory
mdbook build

# The output will be in book/ directory
```

### Step 12: Test Locally

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
6. Verify it appears in the Google Sheet
7. Reload the page - comment should still be there

### Step 13: Check the Google Sheet

1. Open your Google Sheet in another tab
2. You should see a new row with your comment data
3. The metadata column contains JSON with paragraph context
4. This makes it easy to see all comments at a glance

## Part 5: Deploy to Production

### Step 14: Deploy Your Book

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

### Step 15: Update OAuth Settings

After deploying:

1. Go to Google Cloud Console → Credentials → Your OAuth client
2. Edit **"Authorized JavaScript origins"**:
   - Add: `https://yourbook.netlify.app` (or your production URL)
3. Edit **"Authorized redirect URIs"**:
   - Add: `https://yourbook.netlify.app` (or your production URL)
4. Click **"Save"**

### Step 16: Test Production Deployment

1. Visit your production URL
2. Test sign-in and commenting
3. Verify comments appear in Google Sheet
4. Check that comments persist across page reloads

## Part 6: Managing Comments

### Viewing All Comments

Simply open your Google Sheet to see all comments:
- Sort by `created` to see newest first
- Filter by `paragraph_id` to see comments on specific paragraphs
- View `metadata` to see paragraph context
- Use spreadsheet features: charts, pivot tables, etc.

### Moderating Comments

To moderate comments:

1. **Delete a comment**: Delete the row from the sheet
2. **Edit a comment**: Edit the `text` cell
3. **Move a comment**: Change the `paragraph_id`

Changes are reflected immediately on the book site.

### Backing Up Comments

1. **File** → **Download** → **CSV** or **Excel**
2. Or use Google Sheets version history
3. Or enable Google Drive backup

### Exporting Comments

To export comments to another format:

```javascript
// In browser console on your book page
copy(JSON.stringify(allComments, null, 2))
```

Paste into a file or use the spreadsheet export.

## Troubleshooting

### "Sign in failed" or "Authorization error"

- Verify OAuth client ID is correct in `comments-googlesheets.js`
- Check that your site URL is in "Authorized JavaScript origins"
- Ensure Google Sheets API and Drive API are enabled
- Check browser console for specific error messages

### Comments not loading

- Verify Sheet ID is correct
- Check that sheet is shared (at least "Anyone with the link can view")
- Verify the sheet has the correct column headers (A1-G1)
- Check browser console for API errors

### "Failed to add comment"

- Ensure sheet is shared with "Edit" permissions
- Check that you're signed in with an account that has access
- Verify Google Sheets API is enabled
- Check for API rate limits (quota exceeded)

### Rate limit errors

Google Sheets API has quotas:
- 100 requests per 100 seconds per user
- 500 requests per 100 seconds per project

**Solutions:**
- Implement caching in JavaScript
- Reduce frequency of reads
- Use batch operations
- Consider upgrading to paid Google Cloud plan for higher quotas

### Slow performance

Google Sheets is slower than a real database:
- Reads take ~500ms-2s
- Writes take ~1s-3s

**Optimizations:**
- Cache comments in browser localStorage
- Load comments asynchronously
- Show loading indicators
- Limit initial load to current page only

## Advanced Configuration

### Custom Sheet Structure

You can add more columns to track additional data:
- `resolved`: Boolean for resolved comments
- `priority`: Number for comment priority
- `labels`: Tags/categories for comments

Update `comments-googlesheets.js` to read/write these fields.

### Multiple Sheets

To use multiple sheets (e.g., one per chapter):

1. Create multiple sheets in the same spreadsheet
2. Update `sheetName` in config based on current chapter
3. Modify JavaScript to determine which sheet to use

### Backup Automation

Set up automated backups:

1. Use Google Apps Script to copy sheet daily
2. Or use Google Takeout for periodic exports
3. Or use third-party backup services

### Analytics

Use Google Sheets features to analyze comments:

1. Create pivot tables for comment counts per chapter
2. Use formulas to track comment resolution rates
3. Create charts for comment trends over time
4. Use conditional formatting to highlight issues

## Comparison: Google Sheets vs Supabase

| Feature | Google Sheets | Supabase |
|---------|--------------|----------|
| Setup Complexity | Low | Medium |
| Performance | Slow (1-3s latency) | Fast (<100ms) |
| Scalability | Low (~1000 comments) | High (millions) |
| Cost | Free | Free (with limits) |
| Management UI | Excellent (native spreadsheet) | Good (dashboard) |
| Offline Support | No | Possible |
| Rate Limits | Strict (100/100s) | Generous |
| Best For | Small internal docs | Production apps |

## When to Use Google Sheets

**Good fit:**
- Internal company handbook with <50 users
- Low-traffic documentation
- Need easy comment management in spreadsheet
- Want zero infrastructure setup
- Occasional commenting (not real-time chat)

**Not a good fit:**
- High-traffic public documentation
- Real-time commenting requirements
- >1000 comments or >100 concurrent users
- Need sub-second response times
- Frequent comment updates

## Summary

You now have:
- ✅ Google Sheets backend with OAuth authentication
- ✅ mdbook with comment links on every paragraph
- ✅ Zero infrastructure/database setup
- ✅ Easy comment management in spreadsheet
- ✅ Full comment and reply functionality
- ✅ Production-ready deployment

Your users can now leave comments on any paragraph, and all comments are stored in an easy-to-manage Google Sheet!

## Support

If you encounter issues:

1. Check Google Cloud Console logs
2. Review browser console errors
3. Verify Sheet permissions and structure
4. Check API quotas in Cloud Console
5. Test with API Explorer: https://developers.google.com/sheets/api/reference/rest

Happy commenting! 📖💬📊
