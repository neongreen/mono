# mdbook-comments

An mdbook preprocessor that adds paragraph-level commenting functionality similar to Real World Haskell's commenting system.

## Features

- **Simple UI**: Text links saying "comment" at the end of each paragraph
- **Inline Comments**: Comments expand below paragraphs when clicked
- **Rich Context Storage**: Stores position, content, and surrounding context for each paragraph
- **Fuzzy Matching**: Automatically matches comments to paragraphs even after minor edits
- **Orphaned Comments**: Shows unmatched comments at the end of chapters with full context
- **No Data Loss**: All comments are preserved, even when content changes significantly

## Quick Start with Docker

Try mdbook-comments with Neon backend instantly using Docker:

```bash
docker run -p 3000:3000 \
  -e NEON_API_URL="https://ep-xxx-xxx.us-east-2.aws.neon.tech/v2/query" \
  -e NEON_API_KEY="neon_api_key_xxxxxxxxxxxxxxxxxxxxx" \
  ghcr.io/neongreen/mono/mdbook-comments-demo:latest
```

See [DOCKER_DEMO.md](DOCKER_DEMO.md) for complete setup instructions including Neon database configuration.

## Installation

### Building from Source

```bash
cd mdbook-comments
cargo build --release
```

The binary will be at `target/release/mdbook-comments`.

### Installing

```bash
cargo install --path .
```

Or add to your system PATH:

```bash
export PATH="$PATH:/path/to/mdbook-comments/target/release"
```

## Configuration

Add to your `book.toml`:

```toml
[preprocessor.comments]
# API endpoint for comment storage
api-url = "http://localhost:3000/api"

# Authentication type (cookie, bearer-token, oauth)
auth-type = "cookie"

# Similarity threshold for fuzzy matching (0.0 - 1.0)
similarity-threshold = 0.85

# Where to show orphaned comments (end-of-chapter, end-of-page)
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
```

## Backend Options

The plugin supports multiple backend options:

### Option 1: Supabase (Recommended for production)
- Full-featured database with PostgreSQL
- Real-time capabilities
- Built-in authentication
- See `SUPABASE_DEPLOYMENT_GUIDE.md` for setup instructions

### Option 2: Neon (Serverless PostgreSQL)
- Serverless PostgreSQL with generous free tier
- Fast performance with auto-scaling
- Database branching for development
- HTTP Data API for direct queries
- See `NEON_DEPLOYMENT_GUIDE.md` for setup instructions

### Option 3: Google Sheets (Simple alternative)
- Uses Google Sheets as database
- No server setup required
- Easy comment management in spreadsheet
- See `GOOGLE_SHEETS_DEPLOYMENT_GUIDE.md` for setup instructions

### Option 4: Custom Backend API

The plugin requires a backend API with the following endpoints:

### GET /api/comments
Returns all comments for the book.

Response:
```json
[
  {
    "id": "comment-uuid",
    "paragraph-id": "chapter-1-block-3-abc12345",
    "metadata": {
      "id": "chapter-1-block-3-abc12345",
      "position": { "file": "chapter1.md", "block-index": 3 },
      "content": "The paragraph text...",
      "context": {
        "prev": "Previous paragraph...",
        "next": "Next paragraph...",
        "heading-path": ["Chapter 1", "Introduction"]
      },
      "commit": "7a3716b"
    },
    "author": "user@example.com",
    "text": "Great explanation!",
    "created": "2024-01-16T12:00:00Z",
    "parent-id": null,
    "replies": []
  }
]
```

### POST /api/comments
Create a new comment.

Request:
```json
{
  "paragraph-id": "chapter-1-block-3-abc12345",
  "metadata": { /* paragraph metadata */ },
  "text": "Comment text"
}
```

Response: The created comment object.

### POST /api/comments/:id/reply
Reply to a comment.

Request:
```json
{
  "text": "Reply text"
}
```

Response: The created reply object.

## Backend Implementation

Choose one of the following backends:

- **Supabase**: Managed Postgres with built-in API (see `SUPABASE_DEPLOYMENT_GUIDE.md`)
- **Neon**: Serverless PostgreSQL with Data API (see `NEON_DEPLOYMENT_GUIDE.md`)
- **Google Sheets**: Use spreadsheet as database (see `GOOGLE_SHEETS_DEPLOYMENT_GUIDE.md`)
- **PocketBase**: Lightweight Go backend with SQLite
- **Custom**: Any database with a simple REST API

See `MDBOOK_COMMENT_PLUGIN_DESIGN.md` for more details on backend requirements.

### Which Backend to Choose?

**Use Supabase if:**
- Building a production application
- Need built-in authentication (Google, GitHub, email)
- Want real-time features
- Need advanced features (webhooks, edge functions)
- Prefer integrated solution

**Use Neon if:**
- Want serverless PostgreSQL with generous free tier
- Prefer pure PostgreSQL without abstractions
- Need database branching for dev/test
- Want fast cold starts with auto-suspend
- Value simplicity and PostgreSQL compatibility

**Use Google Sheets if:**
- Building internal documentation
- Low traffic (<50 users)
- Want zero infrastructure setup
- Need easy comment management in spreadsheet
- Occasional commenting (not real-time)

## How It Works

### 1. Preprocessing

The Rust preprocessor:
1. Parses markdown content
2. Identifies commentable blocks (paragraphs, lists, code blocks, etc.)
3. Generates unique IDs based on position and content hash
4. Extracts context (previous/next paragraphs, heading hierarchy)
5. Injects metadata and comment links into the HTML

### 2. Client-Side Matching

The JavaScript:
1. Loads all comments from the API
2. Matches comments to current paragraphs using:
   - Exact ID matching (for unchanged content)
   - Fuzzy matching based on content similarity (for edited content)
   - Context matching (prev/next paragraphs, headings)
3. Displays matched comments inline
4. Shows orphaned comments at the end of the chapter

### 3. User Interaction

- Clicking "comment" expands the comment section
- Users can add comments and replies
- Comments are saved via API calls
- Authentication is handled by the backend (cookies, OAuth, etc.)

## Development

### Running Tests

```bash
cargo test
```

### Building Documentation

```bash
cargo doc --open
```

## Design Philosophy

This plugin follows the "Real World Haskell" approach:

1. **Obvious UI**: Plain text links, no fancy icons
2. **Inline Display**: Comments appear below paragraphs, not in modals
3. **Stateless Matching**: JavaScript matches comments without needing diffs
4. **No Data Loss**: Orphaned comments are always preserved and displayed
5. **Simple Backend**: Minimal API requirements, implementation flexible

## License

MIT

## See Also

- [MDBOOK_COMMENT_PLUGIN_DESIGN.md](../MDBOOK_COMMENT_PLUGIN_DESIGN.md) - Full design specification
- [mdbook Documentation](https://rust-lang.github.io/mdBook/)
