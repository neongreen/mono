# mdbook Comment Plugin - Design Specification

## Executive Summary

This document provides a comprehensive design specification for an mdbook plugin that adds paragraph-level commenting functionality similar to Real World Haskell's commenting system.

## Overview

### What is Real World Haskell's Comment System?

Real World Haskell (RWH) features a comment system where:
- Each paragraph in the book has a small "comment" link (usually shown as a speech bubble icon or text)
- Clicking the link opens a comment section specific to that paragraph
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

**Requirement:** Add a visible comment link/button to each commentable block.

**Design Options:**

**Option A: Inline Icon (RWH-style)**
- Small speech bubble icon at the end of each paragraph
- Appears on hover or always visible
- Minimal visual disruption
- **Pros:** Non-intrusive, familiar to RWH users
- **Cons:** May be missed by users, accessibility concerns

**Option B: Margin Icon**
- Icon in the left/right margin
- Appears on hover over the paragraph
- **Pros:** Doesn't disrupt text flow, clear visual indicator
- **Cons:** Requires margin space, may not work on mobile

**Option C: Hover Overlay**
- Button appears in a fixed position when hovering over paragraph
- **Pros:** Very non-intrusive
- **Cons:** Not discoverable, accessibility issues

**Recommended:** Option A with Option B as alternative for different layouts.

### 3. Comment Storage

**Requirement:** Store and retrieve comments associated with paragraph IDs.

**Storage Options:**

**Option A: External Service (like Disqus, utterances, giscus)**
- Use existing comment service
- **Pros:** Maintained by others, feature-rich, handles moderation
- **Cons:** External dependency, privacy concerns, vendor lock-in

**Option B: GitHub Issues/Discussions (like giscus)**
- Each paragraph maps to a GitHub discussion
- **Pros:** Free, version-controlled, familiar to developers
- **Cons:** Requires GitHub account, public repo only

**Option C: Self-hosted Database**
- Custom backend with SQLite/PostgreSQL
- **Pros:** Full control, privacy-friendly
- **Cons:** Requires server, maintenance burden

**Option D: Static Comments (like staticman)**
- Comments stored as JSON/YAML files in the repo
- **Pros:** No external service, version-controlled
- **Cons:** Requires PR workflow, slow feedback loop

**Recommended:** Option B (GitHub Discussions via giscus) for developer-focused books, with extensibility for other backends.

### 4. Comment UI

**Requirement:** Provide an interface for viewing and adding comments.

**Implementation Options:**

**Option A: Inline Expansion**
- Comments expand below the paragraph when clicked
- **Pros:** Contextual, doesn't navigate away
- **Cons:** Can make page very long, layout complexity

**Option B: Modal/Overlay**
- Comments open in a modal window
- **Pros:** Keeps page clean, focused experience
- **Cons:** Loses context, requires closing

**Option C: Sidebar Panel**
- Comments appear in a sliding sidebar
- **Pros:** Maintains context, doesn't break layout
- **Cons:** Screen space, complexity

**Recommended:** Option B (Modal) for simplicity, with Option C as advanced feature.

## Paragraph Identity & Stability

### The Core Challenge

When content is updated, we need to maintain the link between comments and paragraphs even when:
- Paragraphs are reordered
- Paragraphs are deleted
- Paragraphs are merged or split
- Text is rewritten

### Identification Strategy

**Option 1: Hash-Based IDs**
- Generate ID from paragraph content hash
- **Pros:** Automatic, deterministic
- **Cons:** ANY text change breaks the link (too fragile)

**Option 2: Position-Based IDs**
- ID based on location: `chapter-3-para-5`
- **Pros:** Simple to implement
- **Cons:** Breaks when paragraphs are reordered or inserted

**Option 3: Content Prefix + Hash**
- ID includes first N words + hash: `"When-we-define-a-type-abc123"`
- **Pros:** Human-readable, somewhat stable
- **Cons:** Still breaks on text changes

**Option 4: Manual Anchor IDs**
- Author explicitly adds IDs: `{#my-custom-id}`
- **Pros:** Full control, stable across changes
- **Cons:** Manual work, easy to forget

**Option 5: Hybrid Approach (RECOMMENDED)**
- Start with Option 2 (position-based)
- Allow manual override with Option 4
- Use fuzzy matching to reconnect comments when content changes slightly

### Handling Content Updates

#### Scenario 1: Text is Modified (Minor Changes)

**Example:**
```markdown
<!-- Before -->
The quick brown fox jumps over the lazy dog.

<!-- After -->
The quick brown fox leaps over the lazy dog.
```

**Strategy: Fuzzy Matching**
- Calculate similarity score (e.g., Levenshtein distance, cosine similarity)
- If similarity > 80%, maintain the same ID
- Show warning in admin interface: "This paragraph has changed, review comments"

**Implementation Complexity:** Medium
- Requires string similarity algorithm
- Needs threshold tuning
- May need human confirmation

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
- Don't rely on position
- Match based on content hash or fuzzy matching
- Update mapping: old position → new position

**Implementation Complexity:** Medium-High
- Requires storing historical content hashes
- Need migration tool to update mappings

#### Scenario 3: Paragraphs are Merged

**Example:**
```markdown
<!-- Before -->
1. The cat sat on the mat. [has 5 comments]
2. The dog lay on the rug. [has 3 comments]

<!-- After -->
1. The cat sat on the mat, and the dog lay on the rug. [should have 8 comments?]
```

**Strategy Options:**

**Option A: Merge Comments**
- Combine comments from both paragraphs
- Show notice: "Comments from merged paragraphs"
- **Pros:** No comments lost
- **Cons:** May be confusing, comments may not apply to merged text

**Option B: Assign to Best Match**
- Calculate which original paragraph is more similar to merged text
- Assign all comments there, orphan the others
- **Pros:** Clear ownership
- **Cons:** Loses some comments

**Option C: Create "Orphaned Comments" Section**
- Keep comments but mark them as orphaned
- Show in sidebar: "Comments from deleted/merged paragraphs"
- **Pros:** No data loss, transparent
- **Cons:** Extra UI complexity

**Recommended:** Option C for safety, with Option A as user choice.

#### Scenario 4: Paragraphs are Split

**Example:**
```markdown
<!-- Before -->
The cat sat on the mat. The dog lay on the rug. [has 10 comments]

<!-- After -->
1. The cat sat on the mat.
2. The dog lay on the rug.
```

**Strategy:**
- Assign comments to the paragraph with highest similarity
- Typically the first paragraph keeps the comments
- Show notice: "This paragraph was split, review comments"

#### Scenario 5: Paragraphs are Deleted

**Example:**
```markdown
<!-- Before -->
1. Important paragraph. [has 10 comments]
2. Another paragraph.

<!-- After -->
1. Another paragraph.
```

**Strategy:**
- Mark comments as "orphaned"
- Don't delete them from database
- Provide admin interface to:
  - View orphaned comments
  - Reassign to different paragraph
  - Archive permanently
  - Restore if paragraph comes back

#### Scenario 6: Major Rewrite

**Example:**
```markdown
<!-- Before -->
Chapter about functional programming concepts.

<!-- After -->
Complete rewrite about object-oriented programming.
```

**Strategy:**
- Similarity score will be very low
- Treat as "new content"
- Archive old comments with notice
- Start fresh comment threads

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
│   ├── preprocessor.rs        # Rust preprocessor
│   ├── id_generator.rs        # Paragraph ID generation
│   ├── fuzzy_matcher.rs       # Content similarity matching
│   └── migration.rs           # Handle content updates
├── js/
│   ├── comments.js            # Frontend comment UI
│   ├── api.js                 # Backend API integration
│   └── storage.js             # Storage adapter interface
├── css/
│   └── comments.css           # Styling
├── config/
│   └── default.toml           # Default configuration
└── tests/
```

### Configuration

```toml
[preprocessor.comments]
# Storage backend
backend = "giscus"  # or "github-issues", "custom", "staticman"

# Backend configuration
[preprocessor.comments.giscus]
repo = "username/repo"
repo-id = "R_xxx"
category = "Comments"
category-id = "DIC_xxx"

# ID generation
[preprocessor.comments.ids]
strategy = "hybrid"  # position, content-hash, manual, hybrid
auto-fuzzy-match = true
similarity-threshold = 0.8

# UI options
[preprocessor.comments.ui]
icon-position = "inline"  # inline, margin, hover
icon-type = "bubble"      # bubble, text, custom
always-visible = false    # or only on hover

# Commentable elements
[preprocessor.comments.elements]
paragraphs = true
lists = true
blockquotes = true
code-blocks = true
tables = false

# Migration settings
[preprocessor.comments.migration]
enable-fuzzy-matching = true
orphan-comments-action = "preserve"  # preserve, archive, delete
show-warnings = true
```

### Data Model

#### Paragraph Metadata (Stored in Build Artifact)

```json
{
  "chapter": "defining-types",
  "paragraph-id": "chapter-3-para-5-abc123",
  "content-hash": "sha256-of-content",
  "position": {
    "file": "src/chapter3.md",
    "line": 42,
    "block-index": 5
  },
  "manual-id": null,
  "created": "2024-01-15T10:00:00Z",
  "last-modified": "2024-01-20T15:30:00Z"
}
```

#### Comment Data (Stored in Backend)

```json
{
  "comment-id": "comment-uuid",
  "paragraph-id": "chapter-3-para-5-abc123",
  "author": "user@example.com",
  "text": "Great explanation!",
  "created": "2024-01-16T12:00:00Z",
  "replies": [],
  "status": "active"  // active, archived, orphaned
}
```

#### Migration Mapping (Stored During Updates)

```json
{
  "old-paragraph-id": "chapter-3-para-5-old",
  "new-paragraph-id": "chapter-3-para-6-new",
  "reason": "reordered",
  "confidence": 0.95,
  "content-similarity": 0.98,
  "requires-review": false
}
```

## Implementation Phases

### Phase 1: MVP (Minimum Viable Product)
- Basic preprocessor that adds IDs to paragraphs
- Simple position-based IDs
- Manual comment links (open GitHub issue)
- No automatic matching
- **Effort:** 2-3 weeks

### Phase 2: UI Integration
- JavaScript-based comment modal
- Integration with one backend (giscus recommended)
- Basic CSS styling
- **Effort:** 2-3 weeks

### Phase 3: Content Stability
- Implement fuzzy matching for minor text changes
- Basic reordering detection
- Orphaned comment handling
- **Effort:** 3-4 weeks

### Phase 4: Advanced Features
- Migration tool for major content updates
- Admin dashboard for orphaned comments
- Multiple backend support
- Custom styling options
- **Effort:** 4-6 weeks

## Open Questions & Design Choices

### Question 1: What granularity for comments?

**Options:**
- A) Only paragraphs
- B) Paragraphs + list items
- C) Paragraphs + list items + code blocks
- D) Everything except inline elements

**Trade-offs:**
- More granular = more control but more UI clutter
- Less granular = cleaner but less precise feedback

**Recommendation:** Start with B, make C configurable.

### Question 2: How to handle code blocks?

Code blocks are special because:
- They often get updated as code evolves
- Line-by-line comments might be useful
- Syntax highlighting complicates the UI

**Options:**
- A) Treat entire code block as one commentable unit
- B) Allow line-level comments (like GitHub)
- C) Don't allow code block comments at all

**Recommendation:** Start with A, consider B as advanced feature.

### Question 3: Should comments be public or private?

**Public Comments:**
- **Pros:** Community learning, social proof
- **Cons:** Moderation needed, may intimidate users

**Private Comments:**
- **Pros:** Safe space for questions, no moderation
- **Cons:** No community benefit, author gets overwhelmed

**Hybrid:**
- Private by default, author can promote to public
- Or public with option to flag for author-only

**Recommendation:** Public (like RWH), with moderation tools.

### Question 4: What about nested comments (replies)?

**Options:**
- A) Flat comments only (simpler)
- B) One level of replies (standard)
- C) Unlimited nesting (like Reddit)

**Recommendation:** B (one level of replies) - good balance.

### Question 5: How much fuzzy matching is too much?

Setting the similarity threshold is critical:
- Too high (>95%): Fragile, breaks on minor edits
- Too low (<70%): False matches, wrong comments on wrong paragraphs

**Recommendation:**
- Default: 85%
- Configurable by author
- Show confidence score in UI
- Allow manual reassignment

### Question 6: Should the plugin work offline?

**Options:**
- A) Online only (requires backend)
- B) Offline viewing, online commenting
- C) Full offline with local storage

**Recommendation:** B - view cached comments offline, need online to add.

### Question 7: What about translations/multiple versions?

If the book has multiple languages or versions:
- Should comments be per-language?
- Should they be shared across versions?
- How to handle version-specific comments?

**Recommendation:** Per-language comments, with option to share.

### Question 8: How to handle spam and abuse?

**Options:**
- A) Rely on backend's moderation (giscus, GitHub)
- B) Build custom moderation tools
- C) Require authentication only
- D) Rate limiting

**Recommendation:** A + C for MVP.

### Question 9: Anonymous comments?

**Options:**
- A) Require authentication (GitHub, email)
- B) Allow anonymous with optional name/email
- C) Fully anonymous

**Recommendation:** A for quality and moderation.

### Question 10: How to export comments?

Authors may want to:
- Export all comments as PDF/JSON
- Generate summary reports
- Integrate into book (as footnotes, appendix)

**Recommendation:** Provide export tools in Phase 4.

## Comparison: Difficulty of Content Update Options

| Scenario | Strategy | Difficulty | Pros | Cons |
|----------|----------|-----------|------|------|
| Minor text changes | Fuzzy matching (85% threshold) | Medium | Maintains most links | May miss some changes |
| Reordering | Content-based matching | Medium-High | Preserves comments | Requires content hashing |
| Merging paragraphs | Create orphaned section | Low-Medium | No data loss | Extra UI complexity |
| Splitting paragraphs | Assign to best match + notice | Medium | Clear ownership | Some context loss |
| Deleting paragraphs | Orphan comments | Low | Safe, reversible | Requires cleanup UI |
| Major rewrites | Archive old + start fresh | Low | Clean slate | Comments disconnected |

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

## References

- Real World Haskell: https://book.realworldhaskell.org/
- mdbook Documentation: https://rust-lang.github.io/mdBook/
- giscus: https://giscus.app/
- utterances: https://utteranc.es/
- staticman: https://staticman.net/

---

**Document Version:** 1.0  
**Last Updated:** 2024-01-15  
**Author:** AI Assistant
