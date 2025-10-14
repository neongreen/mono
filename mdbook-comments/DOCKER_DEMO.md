# Docker Demo for mdbook-comments with Neon

This Docker image provides a complete demo of mdbook-comments with Neon database backend. It allows you to quickly spin up a working instance of the example book with commenting functionality connected to your Neon database.

## Prerequisites

1. A Neon account and database set up (see [NEON_DEPLOYMENT_GUIDE.md](NEON_DEPLOYMENT_GUIDE.md))
2. Docker installed on your machine
3. Your Neon API credentials:
   - `NEON_API_URL` - Your Neon endpoint URL (e.g., `https://ep-xxx-xxx.us-east-2.aws.neon.tech/v2/query`)
   - `NEON_API_KEY` - Your Neon API key (starts with `neon_api_key_...`)

## Quick Start

Pull and run the Docker image with one command:

```bash
docker run -p 3000:3000 \
  -e NEON_API_URL="https://ep-xxx-xxx.us-east-2.aws.neon.tech/v2/query" \
  -e NEON_API_KEY="neon_api_key_xxxxxxxxxxxxxxxxxxxxx" \
  ghcr.io/neongreen/mono/mdbook-comments-demo:latest
```

Then open your browser to http://localhost:3000

## Environment Variables

### Required

- `NEON_API_URL` - Your Neon API endpoint URL
  - Format: `https://[endpoint].neon.tech/v2/query`
  - Get this from your Neon project dashboard

- `NEON_API_KEY` - Your Neon API key
  - Format: `neon_api_key_xxxxxxxxxxxxxxxxxxxxx`
  - Generate this in Neon dashboard under Settings → API Keys

### Optional

- `NEON_DATABASE` - Database name (default: `neondb`)
  - Only needed if you're using a different database name

## Setting Up Neon Database

Before running the Docker image, you need to set up your Neon database with the required tables:

1. Create a Neon account at https://neon.tech/
2. Create a new project
3. Run the following SQL in the Neon SQL Editor:

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
```

4. Get your API credentials from the Neon dashboard

## Examples

### Using docker run

```bash
docker run -p 3000:3000 \
  -e NEON_API_URL="https://ep-cool-darkness-123456.us-east-2.aws.neon.tech/v2/query" \
  -e NEON_API_KEY="neon_api_key_abc123def456" \
  ghcr.io/neongreen/mono/mdbook-comments-demo:latest
```

### Using docker-compose

Create a `docker-compose.yml` file:

```yaml
version: '3.8'

services:
  mdbook-comments:
    image: ghcr.io/neongreen/mono/mdbook-comments-demo:latest
    ports:
      - "3000:3000"
    environment:
      - NEON_API_URL=https://ep-cool-darkness-123456.us-east-2.aws.neon.tech/v2/query
      - NEON_API_KEY=neon_api_key_abc123def456
      - NEON_DATABASE=neondb
```

Then run:

```bash
docker-compose up
```

### Using environment file

Create a `.env` file:

```
NEON_API_URL=https://ep-cool-darkness-123456.us-east-2.aws.neon.tech/v2/query
NEON_API_KEY=neon_api_key_abc123def456
NEON_DATABASE=neondb
```

Then run:

```bash
docker run -p 3000:3000 --env-file .env ghcr.io/neongreen/mono/mdbook-comments-demo:latest
```

## What's Included

The Docker image contains:

- mdbook (book generator)
- mdbook-comments preprocessor (built from source)
- Example book with sample content
- Neon backend JavaScript for comment storage
- All necessary CSS and JavaScript assets

## Customization

If you want to use your own book content instead of the example:

1. Clone this repository
2. Modify the files in `mdbook-comments/example-book/`
3. Build your own Docker image:

```bash
cd mdbook-comments
docker build -t my-mdbook-comments .
```

4. Run your custom image:

```bash
docker run -p 3000:3000 \
  -e NEON_API_URL="your-url" \
  -e NEON_API_KEY="your-key" \
  my-mdbook-comments
```

## Troubleshooting

### Error: NEON_API_URL environment variable is required

Make sure you're passing the environment variables correctly. Check that:
- The variable names are correct (NEON_API_URL, not NEON_URL)
- You're using `-e` flag for each variable
- The values are properly quoted if they contain special characters

### Comments not loading

1. Check that your Neon database is running and accessible
2. Verify your API key is correct
3. Check the browser console for JavaScript errors
4. Ensure the comments table exists in your database

### Port already in use

If port 3000 is already in use, map to a different port:

```bash
docker run -p 8080:3000 -e NEON_API_URL="..." -e NEON_API_KEY="..." ghcr.io/neongreen/mono/mdbook-comments-demo:latest
```

Then access at http://localhost:8080

## Building from Source

To build the Docker image yourself:

```bash
cd mdbook-comments
docker build -t mdbook-comments-demo .
```

## Security Notes

- The Neon API key is embedded in the JavaScript file at container startup
- This is a demo setup - for production use, implement proper authentication
- Consider using environment-specific API keys
- Don't expose your Neon credentials in public repositories

## See Also

- [NEON_DEPLOYMENT_GUIDE.md](NEON_DEPLOYMENT_GUIDE.md) - Complete Neon setup guide
- [README.md](README.md) - mdbook-comments documentation
- [Neon Documentation](https://neon.tech/docs)
