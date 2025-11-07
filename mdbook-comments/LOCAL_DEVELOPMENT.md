# Local Development Guide

This guide explains how to run mdbook-comments locally with a local Supabase database, allowing you to develop and test without any external services.

## Overview

The local development setup uses:
- **Supabase Local** - Full PostgreSQL database with REST API running in Docker
- **mdbook serve** - Serves the example book with live reload
- **No authentication** - Trust-based commenting (anyone can comment with just a name)

Everything runs on your machine with no internet connection required.

## Prerequisites

1. **Rust and Cargo** - For building the preprocessor
   ```bash
   # Install via rustup: https://rustup.rs/
   curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
   ```

2. **mdbook** - For serving the book
   ```bash
   cargo install mdbook
   ```

3. **Supabase CLI** - For local database
   ```bash
   # macOS
   brew install supabase/tap/supabase
   
   # Linux/Windows
   # See: https://supabase.com/docs/guides/cli/getting-started
   ```

4. **Docker** - Required by Supabase CLI
   - Install Docker Desktop from https://www.docker.com/products/docker-desktop

## Quick Start

### First Time Setup

```bash
cd mdbook-comments

# One-time setup
mise run dev:setup
```

This will:
- Check for Supabase CLI
- Initialize Supabase project
- Create configuration files

### Daily Development

```bash
# Start everything
mise run dev
```

This will:
1. Start Supabase (PostgreSQL + REST API + Studio)
2. Extract credentials and configure the book
3. Start mdbook serve

The command will output:
```
📚 Book will be available at: http://localhost:3000
🔧 Supabase Studio at: http://localhost:54323
```

**Open your browser:**
- http://localhost:3000 - Your book with comments
- http://localhost:54323 - Supabase Studio (database UI)

### Stopping

Press `Ctrl+C` to stop mdbook.

To stop Supabase:
```bash
mise run dev:stop
```

## How It Works

### Architecture

```
┌─────────────────┐
│  mdbook serve   │ ← Serves book at :3000
│  :3000          │
└────────┬────────┘
         │
         │ JavaScript loads comments
         ▼
┌─────────────────┐
│ Supabase Local  │
│                 │
│ PostgreSQL      │ ← Comments stored here
│ PostgREST API   │ ← REST API at :54321
│ Studio UI       │ ← Admin UI at :54323
└─────────────────┘
```

### No Authentication

Unlike production Supabase, local development uses a simplified trust-based approach:

- No login required
- Users enter their name in a text field
- Name is saved in localStorage
- Anyone can read/write comments
- Perfect for development and testing

### Configuration

The `comments-supabase.js` file works for both local and remote:

1. **Local mode**: Reads config from `window.SUPABASE_CONFIG` (auto-generated)
2. **Remote mode**: Uses hardcoded values in the JS file

The `scripts/configure-supabase.sh` script:
- Extracts local Supabase credentials (`supabase status`)
- Generates `example-book/supabase-config.js`
- Book loads this config before the main JS

## Available Commands

### Core Development

```bash
mise run dev          # Start local development
mise run dev:stop     # Stop Supabase
mise run dev:reset    # Reset database (delete all data)
```

### Database Management

```bash
mise run dev:studio   # Open Supabase Studio in browser
mise run dev:status   # Show what's running
mise run dev:logs     # View Supabase logs
```

### Building

TODO: outdated

```bash
mise run build        # Build the preprocessor
mise run test         # Run tests
mise run format-rust  # Format code
mise run check        # Clippy checks
```

## Viewing and Managing Comments

### Using Supabase Studio

Open http://localhost:54323 to access Supabase Studio where you can:

- View all comments in the `comments` table
- Edit or delete comments
- See metadata and relationships
- Run SQL queries

### Using SQL Queries

In Supabase Studio's SQL Editor:

```sql
-- View all comments
SELECT * FROM comments ORDER BY created DESC;

-- View comments for a specific paragraph
SELECT * FROM comments 
WHERE paragraph_id = 'your-paragraph-id'
ORDER BY created;

-- Count total comments
SELECT COUNT(*) FROM comments;

-- Delete all comments
TRUNCATE comments;
```

## Database Schema

The comments table:

```sql
CREATE TABLE comments (
    id UUID PRIMARY KEY,
    paragraph_id TEXT NOT NULL,
    metadata JSONB NOT NULL,
    author TEXT NOT NULL,        -- Simple name string
    text TEXT NOT NULL,
    created TIMESTAMPTZ DEFAULT NOW(),
    parent_id UUID REFERENCES comments(id)
);
```

Key points:
- `author` is just a text field (not a user ID)
- `metadata` contains paragraph context for fuzzy matching
- `parent_id` links replies to parent comments
- No RLS policies (public read/write)

## Troubleshooting

### Supabase CLI not found

```
❌ Supabase CLI not found.
```

Install it:
```bash
brew install supabase/tap/supabase  # macOS
```

### Docker not running

```
Error: Cannot connect to the Docker daemon
```

Start Docker Desktop or your Docker daemon.

### Port already in use

If port 3000 or 54321 is already in use:

1. Stop the conflicting service
2. Or modify the port in `book.toml` (for mdbook)
3. Supabase ports can be configured in `supabase/config.toml`

### Comments not appearing

1. Check browser console for errors
2. Verify Supabase is running: `mise run dev:status`
3. Check that `supabase-config.js` was generated
4. Open http://localhost:54323 to verify the database

### Database issues

Reset the database:
```bash
mise run dev:reset
```

This will delete all data and recreate the schema.

## Tips

### Adding Sample Data

Edit `supabase/seed.sql` to add test comments:

```sql
INSERT INTO comments (paragraph_id, metadata, author, text) VALUES
('chapter-1-block-0-abc123', 
 '{"id": "chapter-1-block-0-abc123", "position": {"file": "chapter1.md", "block-index": 0}, "content": "Sample text", "context": {}}',
 'Test User',
 'This is a sample comment');
```

Then reset the database: `mise run dev:reset`

### Working Offline

Once started, Supabase Local works completely offline. All data is stored in Docker volumes.

### Data Persistence

Your comments persist across restarts. Supabase stores data in Docker volumes.

To clear data:
```bash
mise run dev:reset
```

To completely remove Supabase:
```bash
supabase stop --no-backup
rm -rf supabase/.branches/
```

## Deploying to Production

When you're ready to deploy, you can use the same JavaScript with a remote Supabase:

1. Create a Supabase project at https://supabase.com
2. Run the migration SQL (from `supabase/migrations/`)
3. Update `comments-supabase.js` with your production URL and key
4. Add proper authentication (see `SUPABASE_DEPLOYMENT_GUIDE.md`)

The same `comments-supabase.js` works for both!

## See Also

- [README.md](README.md) - Main project documentation
- [SUPABASE_DEPLOYMENT_GUIDE.md](SUPABASE_DEPLOYMENT_GUIDE.md) - Production deployment
- [Supabase CLI Documentation](https://supabase.com/docs/guides/cli)
