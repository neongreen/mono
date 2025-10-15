-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create comments table
-- No RLS, no auth - trust-based access for local development
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

-- Grant public access (no authentication required)
GRANT ALL ON comments TO anon;
GRANT ALL ON comments TO authenticated;
