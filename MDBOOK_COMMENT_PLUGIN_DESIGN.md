# mdbook Comment Plugin - Design Specification

## Executive Summary

This document provides a comprehensive design specification for an mdbook plugin that adds paragraph-level commenting functionality similar to Real World Haskell's commenting system.

## Overview

### What is Real World Haskell's Comment System?

Real World Haskell (RWH) features a comment system where:
- Each paragraph in the book has a small text link that says "comment" at the end
- The link is slightly smaller than the body text and blue (standard link styling)
- Clicking the link expands a comment section below the paragraph
- Users can read existing comments and add new ones
- Comments are persistent and tied to specific paragraphs

### Goal

Create an mdbook preprocessor/renderer plugin that adds similar functionality to any mdbook project.

## Core Requirements

### 1. Paragraph Identification

**Requirement:** Each commentable block must have a stable, unique identifier.

**What qualifies as a "commentable block"?**
- Paragraphs (`<p>` tags)
- List items (`<li>` tags)
- Blockquotes (`<blockquote>` tags)
- Code blocks (`<pre><code>` tags)
- Tables (entire table or individual cells - TBD)

**Exclusions (not commentable):**
- Headings (they're structural, not content)
- Inline elements (links, bold, italic, etc.)
- Images by themselves (unless in a paragraph)

### 2. UI Elements

**Requirement:** Add a visible comment link to each commentable block.

**Design Choice: Text Link (RWH-style)**
- Small text link that says "comment" at the end of each paragraph
- Slightly smaller font size than body text
- Standard link styling (blue, underlined)
- Always visible (not just on hover)
- **Pros:** 
  - Extremely obvious and discoverable
  - No confusion about what it does
  - Simple to implement
  - Works everywhere (desktop, mobile, screen readers)
- **Cons:** 
  - Visible in text flow (but this is actually a feature)
  - May be considered "ugly" by some designers

**Rationale:** While a speech bubble icon might be more aesthetically pleasing, the text link is proven to work well in Real World Haskell. It's completely unambiguous and requires no user education.

### 3. Comment Storage

**Requirement:** Store and retrieve comments associated with paragraph IDs.

**Design Choice: Simple API with Backend Database**

The plugin will use a simple HTTP API to store and retrieve comments. The frontend (static HTML + JavaScript) will make API calls to:
- `GET /api/comments?paragraph_id=...` - Retrieve comments for a paragraph
- `POST /api/comments` - Add a new comment
- `POST /api/comments/:id/reply` - Reply to a comment

**Backend Implementation (TBD):**
The actual backend implementation is intentionally left flexible. Possible options:
- **Supabase:** Managed Postgres with built-in API
- **PocketBase:** Lightweight Go backend with SQLite
- **Custom Postgres:** Direct database access with a thin API layer

**Key Requirements:**
- Must support authentication (e.g., OAuth, session cookies)
- Must be self-hostable for private use cases
- API should be simple and well-documented for easy backend swapping

**Why NOT GitHub Issues/External Services:**
- Need full control for private company handbooks
- Want to avoid external dependencies
- Need simple, fast API without platform lock-in

### 4. Comment UI

**Requirement:** Provide an interface for viewing and adding comments.

**Design Choice: Inline Expansion (RWH-style)**

When a user clicks the "comment" link:
- The comment section expands directly below the paragraph
- Existing comments are displayed in chronological order
- A reply form is shown at the bottom
- The page dynamically grows to accommodate the comments
- Clicking the link again collapses the comments

**Features:**
- One level of nesting (replies to comments)
- Timestamp and author name for each comment
- Simple text formatting (markdown in comments)
- "Resolve" or "Mark as read" functionality (optional)

**Rationale:** 
- Keeps comments in context with the paragraph
- Simple, proven UX from Real World Haskell
- No modal complexity or navigation away from content
- Making the page longer is not a problem - it's better than losing context

## Paragraph Identity & Stability

### The Core Challenge

When content is updated, we need to maintain the link between comments and paragraphs even when:
- Paragraphs are reordered
- Paragraphs are deleted
- Paragraphs are merged or split
- Text is rewritten

### Identification Strategy

**Design Choice: Rich Context Storage**

Each comment will be associated with extensive metadata about its paragraph:

1. **Position Information:**
   - File path (e.g., `src/chapter-3.md`)
   - Position in file (block index within chapter)
   - Position within section (paragraph number within current heading section)

2. **Content Signature:**
   - The complete text of the paragraph
   - The text of the preceding paragraph (if exists)
   - The text of the following paragraph (if exists)

3. **Structural Context:**
   - Section heading hierarchy (e.g., `["Chapter 3", "Defining Types", "Pattern Matching"]`)
   - Closest heading above the paragraph
   - Commit hash of the book version when comment was made

4. **Generated ID:**
   - Combination of position + content hash: `chapter-3-para-5-abc123`
   - Used for initial matching only

**Rationale:**
- No manual IDs required (authors won't maintain them)
- Rich context enables fuzzy matching when content changes
- Stateless: JavaScript has everything it needs to match comments to current content
- Philosophy: Store enough context that matching is possible even without historical diffs

### Handling Content Updates

#### Scenario 1: Text is Modified (Minor Changes)

**Example:**
```markdown
<!-- Before -->
The quick brown fox jumps over the lazy dog.

<!-- After -->
The quick brown fox leaps over the lazy dog.
```

**Strategy: Client-Side Fuzzy Matching**
- JavaScript calculates similarity score when loading the page
- Compares stored paragraph text with current paragraph text
- If similarity is high enough, displays comments
- No admin interface - happens automatically on page load

**Implementation Complexity:** Medium
- Requires string similarity algorithm in JavaScript
- Needs threshold tuning
- Fully automatic, no human intervention

#### Scenario 2: Paragraphs are Reordered

**Example:**
```markdown
<!-- Before -->
1. First paragraph about X.
2. Second paragraph about Y.
3. Third paragraph about Z.

<!-- After -->
1. Third paragraph about Z.
2. First paragraph about X.
3. Second paragraph about Y.
```

**Strategy: Content-Based Matching**
- JavaScript matches based on paragraph content, not position
- Uses stored paragraph text + prev/next paragraph context
- Finds best match in current document structure
- Displays comments at the matched location

**Implementation Complexity:** Medium
- Content matching algorithm in JavaScript
- No server-side migration needed (stateless approach)

#### Scenario 3: Paragraphs are Merged

**Example:**
```markdown
<!-- Before -->
1. The cat sat on the mat. [has 5 comments]
2. The dog lay on the rug. [has 3 comments]

<!-- After -->
1. The cat sat on the mat, and the dog lay on the rug. [should have 8 comments?]
```

**Strategy: Merge Comments**
- JavaScript combines comments from all matching source paragraphs
- Displays them all on the merged paragraph
- Shows a notice: "Comments from 2 paragraphs"
- **Pros:** No comments lost, simple and transparent
- **Cons:** May have many comments on one paragraph

**If no good match:** Comments become orphaned (see below)

#### Scenario 4: Paragraphs are Split

**Example:**
```markdown
<!-- Before -->
The cat sat on the mat. The dog lay on the rug. [has 10 comments]

<!-- After -->
1. The cat sat on the mat.
2. The dog lay on the rug.
```

**Strategy: Assign to Best Match**
- JavaScript calculates similarity for each resulting paragraph
- Assigns comments to the most similar one
- Typically the first paragraph if content is evenly distributed
- No special notice needed (comments just appear on one)

#### Scenario 5: Paragraphs are Deleted

**Example:**
```markdown
<!-- Before -->
1. Important paragraph. [has 10 comments]
2. Another paragraph.

<!-- After -->
1. Another paragraph.
```

**Strategy: Display as Orphaned**
- Comments that can't be matched appear at the end of the chapter/page
- Section titled "Unmapped Comments" or similar
- Each orphaned comment shows its stored context:
  - Original paragraph text
  - Section heading it was under
  - Surrounding paragraphs
- Users can still read and reply to orphaned comments
- Comments stay in database forever (no data loss)

**No Admin Interface:**
- The system is stateless - just displays what it can match
- Orphaned comments resolve themselves if matching content returns
- Authors can see orphaned comments just like readers

#### Scenario 6: Major Rewrite

**Example:**
```markdown
<!-- Before -->
Chapter about functional programming concepts.

<!-- After -->
Complete rewrite about object-oriented programming.
```

**Strategy: Become Orphaned**
- Similarity score will be very low
- Comments can't be matched to any current paragraph
- Appear in "Unmapped Comments" section at end of chapter
- Readers can see what was discussed in previous versions
- New paragraphs start with no comments (fresh slate)

## Technical Architecture

### Plugin Type

mdbook supports:
1. **Preprocessors:** Modify markdown before rendering
2. **Renderers:** Custom output format
3. **Backends:** Alternative rendering engines

**Recommended:** Preprocessor + Custom JavaScript

**Why?**
- Preprocessor adds IDs and comment anchors to markdown
- JavaScript handles UI interactions and API calls
- Works with all mdbook themes
- Easy to distribute

### Component Breakdown

```
mdbook-comments/
├── src/
│   ├── preprocessor.rs        # Main preprocessor entry point
│   ├── element_finder.rs      # Find commentable elements in AST
│   ├── id_generator.rs        # Generate IDs from position + content
│   ├── metadata_extractor.rs  # Extract context (prev/next, headings)
│   └── html_injector.rs       # Inject metadata and links into HTML
├── js/
│   ├── comments.js            # Main UI controller
│   ├── matcher.js             # Fuzzy matching algorithm
│   ├── api.js                 # HTTP API client
│   └── ui.js                  # DOM manipulation for comment display
├── css/
│   └── comments.css           # Styling for links and comment sections
└── tests/
    ├── integration/           # End-to-end tests
    └── unit/                  # Unit tests for matcher
```

**No Backend Code:** Backend implementation (Supabase/PocketBase/Postgres) is separate and pluggable.

### Configuration

```toml
[preprocessor.comments]
# API endpoint for comment storage
api-url = "https://comments.example.com/api"

# Authentication (cookies/headers passed through)
auth-type = "cookie"  # or "bearer-token", "oauth"

# Matching configuration
similarity-threshold = 0.85  # How similar content must be to match
orphaned-comments-location = "end-of-chapter"  # or "end-of-page"

# Commentable elements
[preprocessor.comments.elements]
paragraphs = true
lists = true
blockquotes = true
code-blocks = true
tables = true  # entire tables
headings = false  # headings are structural, not commentable

# UI customization
[preprocessor.comments.ui]
link-text = "comment"  # text for comment links
show-comment-count = true  # show "(3)" after "comment" if there are 3 comments
```

### Data Model

#### Paragraph Metadata (Injected into HTML by Preprocessor)

```json
{
  "id": "chapter-3-para-5-abc123",
  "position": {
    "file": "src/chapter3.md",
    "block-index": 5,
    "section-index": 2
  },
  "content": "The quick brown fox jumps over the lazy dog.",
  "context": {
    "prev": "This is the preceding paragraph.",
    "next": "This is the following paragraph.",
    "heading-path": ["Chapter 3", "Defining Types", "Pattern Matching"]
  },
  "commit": "7a3716b"
}
```

This metadata is embedded as a `data-comment-meta` attribute on each commentable element.

#### Comment Data (Stored in Backend Database)

```json
{
  "id": "comment-uuid-1",
  "paragraph-id": "chapter-3-para-5-abc123",
  "metadata": {
    "position": {
      "file": "src/chapter3.md",
      "block-index": 5
    },
    "content": "The quick brown fox jumps over the lazy dog.",
    "context": {
      "prev": "This is the preceding paragraph.",
      "next": "This is the following paragraph.",
      "heading-path": ["Chapter 3", "Defining Types", "Pattern Matching"]
    },
    "commit": "7a3716b"
  },
  "author": "user@example.com",
  "text": "Great explanation!",
  "created": "2024-01-16T12:00:00Z",
  "parent-id": null,
  "replies": ["comment-uuid-2"]
}
```

**Key Design Decision:** Comments store their own context. This enables stateless matching - the JavaScript doesn't need access to old book versions or diffs.

## Implementation Phases

### Phase 1: Basic Infrastructure
- Rust preprocessor that:
  - Identifies all commentable elements
  - Generates IDs based on position + content hash
  - Injects metadata into HTML as data attributes
  - Adds "comment" links with appropriate styling
- **Effort:** 1-2 weeks

### Phase 2: Comment UI
- JavaScript that:
  - Fetches comments from API for current page
  - Displays inline comment sections
  - Handles expand/collapse of comments
  - Provides reply form
- Basic CSS styling
- **Effort:** 2-3 weeks

### Phase 3: Matching Algorithm
- Implement fuzzy matching in JavaScript:
  - String similarity calculation
  - Context-based matching (prev/next paragraphs)
  - Position-based matching as fallback
- Orphaned comment display at end of chapter
- **Effort:** 2-3 weeks

### Phase 4: Backend Implementation
- Choose backend (Supabase/PocketBase/Postgres)
- Implement simple API endpoints
- Add authentication integration
- **Effort:** 1-2 weeks (depending on choice)

## Design Decisions (Finalized)

### 1. Commentable Elements
**Decision:** Everything except inline elements and headings
- Paragraphs ✓
- List items ✓
- Blockquotes ✓
- Code blocks ✓ (entire block as one unit)
- Tables ✓
- Headings ✗ (structural, not content)

### 2. Comment Visibility
**Decision:** Public comments with authentication required
- Use case: Company handbook, internal documentation
- Authentication required (e.g., Google OAuth, session cookies)
- No anonymous comments
- No spam concerns (private, authenticated site)

### 3. Reply Nesting
**Decision:** One level of replies
- Top-level comments on paragraphs
- One level of replies to comments
- No deeply nested threads

### 4. Fuzzy Matching Threshold
**Decision:** 85% similarity (configurable)
- Balances between false positives and false negatives
- Can be adjusted per-deployment based on usage patterns
- No UI for confidence scores (keep it simple)

### 5. Offline Support
**Decision:** Online only
- Comments require API access
- No offline caching or viewing
- Acceptable for the use case (internal handbooks)

### 6. Translations/Versions
**Decision:** Out of scope
- Single language per deployment
- No multi-version comment tracking
- Can be added later if needed

### 7. Comment Export
**Decision:** Not needed
- Comments stay in database
- Can be queried directly if needed
- No special export tools

## Comparison: Content Update Handling

| Scenario | Strategy | Difficulty | Result |
|----------|----------|-----------|--------|
| Minor text changes | Fuzzy matching (85% threshold) | Medium | Comments stay attached |
| Reordering | Content + context matching | Medium | Comments follow paragraph |
| Merging paragraphs | Merge all comments | Low | All comments on merged paragraph |
| Splitting paragraphs | Assign to best match | Low | Comments on most similar paragraph |
| Deleting paragraphs | Show as orphaned | Low | Comments at end of chapter with context |
| Major rewrites | Show as orphaned | Low | Comments at end of chapter with context |

**Key Principle:** Never lose data. If matching fails, show comments as orphaned with full context.

## Success Criteria

The plugin will be considered successful if:

1. **Functionality:**
   - ✓ Comments can be added to paragraphs
   - ✓ Comments persist across builds
   - ✓ Comments remain connected through minor text edits (>80% success rate)
   
2. **Usability:**
   - ✓ Comment UI is intuitive and non-intrusive
   - ✓ Works on mobile and desktop
   - ✓ Accessible (keyboard navigation, screen readers)
   
3. **Maintainability:**
   - ✓ Configuration is simple
   - ✓ Works with existing mdbook themes
   - ✓ Minimal maintenance burden

4. **Performance:**
   - ✓ Build time increase < 10%
   - ✓ Page load time increase < 100ms
   - ✓ Comment loading is asynchronous

## Next Steps

1. **Get Feedback:** Review this design with stakeholders
2. **Prototype:** Build a proof-of-concept for Phase 1
3. **User Testing:** Get early feedback on UI/UX
4. **Iterate:** Refine based on real usage

## Key Design Principles

1. **Simple and Obvious:** Text link that says "comment" - no icons, no hover states
2. **Stateless Matching:** All context stored with comments, no migration tools needed
3. **No Data Loss:** Orphaned comments displayed with context, never deleted
4. **Inline UI:** Comments expand below paragraphs, no modals or sidebars
5. **Backend Agnostic:** Simple API, implementation details TBD (Supabase/PocketBase/Postgres)
6. **Private Use Case:** Authenticated users only, no spam concerns

## References

- Real World Haskell: https://book.realworldhaskell.org/
- mdbook Documentation: https://rust-lang.github.io/mdBook/
- Supabase: https://supabase.com/
- PocketBase: https://pocketbase.io/

---

**Document Version:** 2.0  
**Last Updated:** 2025-10-10  
**Status:** Design Finalized
