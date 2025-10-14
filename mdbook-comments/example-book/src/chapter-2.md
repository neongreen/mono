# Chapter 2: Advanced Topics

This chapter covers more advanced aspects of the commenting system.

## Content Updates

One of the key features of this system is how it handles content updates. When you edit a paragraph, the system tries to match existing comments to the new version.

### Fuzzy Matching

The system uses fuzzy matching to reconnect comments when text changes. If a paragraph is edited but remains similar enough (default: 85% similarity), comments will stay attached.

For example, if you change "The quick brown fox" to "The fast brown fox", the comment will still be connected because the text is very similar.

### Orphaned Comments

When content is removed or rewritten completely, comments become "orphaned". They're not lost - instead, they're displayed at the end of the chapter with their original context.

This ensures no data loss while keeping the main content clean.

## Technical Details

The system stores rich metadata for each paragraph:

- **Position**: File path and block index
- **Content**: The full text of the paragraph
- **Context**: Previous and next paragraphs, heading hierarchy
- **Commit**: Git commit hash when the comment was made

This metadata enables stateless matching - the JavaScript can figure out where comments belong without needing access to the old version of the book.

## Backend Integration

The plugin uses a simple API for comment storage:

```
GET  /api/comments              - Fetch all comments
POST /api/comments              - Create a new comment
POST /api/comments/:id/reply    - Reply to a comment
```

You can implement this backend using:

- Supabase (managed Postgres)
- PocketBase (lightweight Go + SQLite)
- Custom backend (any database + REST API)

The frontend is completely agnostic about backend implementation.

## Summary

The commenting system provides:

1. Simple, obvious UI (text links)
2. Inline expansion (comments appear below paragraphs)
3. Fuzzy matching (handles content updates gracefully)
4. No data loss (orphaned comments are preserved)
5. Flexible backend (multiple implementation options)

This makes it ideal for company handbooks, technical documentation, and educational materials where reader feedback is valuable.
