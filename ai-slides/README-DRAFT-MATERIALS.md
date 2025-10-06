# Context Engineering Talk - Draft Materials

This folder contains rough draft materials for organizing a talk on context engineering. These materials are **for discussion**, not final output.

## What's Here

### 📋 Main Draft Documents

1. **ROUGH-DRAFT-BREAKDOWN.md** - The main breakdown
   - 10 tracks/topics identified from source material
   - Each track has: core idea, slides outline, discussion points
   - Includes alternative organizational schemes
   - ~15,000 words of detailed structure

2. **TRACKS-OVERVIEW.md** - Quick reference guide
   - One-page summary of all 10 tracks
   - Key questions for each track
   - Time estimates
   - Flow options comparison (Linear vs Problem-First vs Three Acts)
   - Identified themes across tracks

3. **ALTERNATIVE-STRUCTURES.md** - Six different ways to organize
   - Structure A: Build from Foundations (current draft)
   - Structure B: Problem-First Hook (recommended)
   - Structure C: Three Acts
   - Structure D: Compare & Contrast
   - Structure E: Five Big Questions
   - Structure F: User Journey
   - Includes pros/cons and best audience for each

4. **KEY-QUOTES-AND-THEMES.md** - Preserved voice and insights
   - Direct quotes from source material organized by theme
   - User pain points in their own words
   - Memorable metaphors (CSS precedence, sponge)
   - Technical insights with context
   - Tone calibration examples

### 📁 Source Materials

- **context-thoughts.md** - Detailed RFC on immutable vs mutable context (287 lines)
- **slides-v0.md** - Simple architectural diagram
- **IMG_6438-6443 Large.jpeg** - Six handwritten notes (OCR attempted but limited success on handwriting)

## The 10 Tracks Identified

1. **What IS Context?** - Foundation and definitions
2. **The Mutable Context Problem** - Why mutation makes reasoning hard
3. **User Preferences (Implicit vs Explicit)** - Capturing what users want
4. **Multi-Model Economics** - 90% discount enables parallel prefixes
5. **Multimodal Inputs** - Don't pre-convert and lose information
6. **Sub-Agents & Tools** - Reusing prefixes for background agents
7. **The Sponge Metaphor** - Functional views over immutable log
8. **Why Immutable Wins** - The strict superset argument
9. **Open Questions & Challenges** - Storage, privacy, evaluation
10. **Closing Thoughts** - Different design space, humble exploration

## Key Themes Discovered

### "Optimizing for the Wrong Thing"
- Optimizing for eval scores, not semantics
- Optimizing for explicit capture, not actual preferences
- Optimizing for universal format, not model capabilities

### "The Power of Never Deleting"
- Provenance & retrospective A/B testing
- Re-evaluation with new criteria
- Strict superset of mutable approaches

### "Economic Incentives Shape Architecture"
- 90% prompt prefix discount is foundational
- Enables multi-prefix approach
- Sub-agents get automatic state continuity

### "User Experience vs System Constraints"
- Don't make users be eval experts
- Users mix tasks in conversations
- Privacy and governance considerations

## Recommended Structure (Modified Structure B)

1. **HOOK:** The Preference Nightmare (3 min) - Start with relatable pain
2. **FOUNDATION:** What IS Context? (5 min) - Quick definitions
3. **PROBLEM:** Why Mutation Fails (7 min) - CSS precedence, semantics
4. **OPPORTUNITY:** The 90% Discount (5 min) - Economic insight
5. **SOLUTION SPACE:** (20 min) Multi-prefix, Sponges, Sub-agents
6. **THE CASE:** Why Immutable Wins (6 min) - Strict superset argument
7. **REALITY:** The Hard Parts (8 min) - Storage, privacy, challenges
8. **CLOSING:** Invitation (3 min) - What am I missing?

Total: ~57 minutes + Q&A

## Next Steps for Discussion

### Questions to Answer

1. **Scope:** 10 tracks too many? Combine some?
2. **Order:** Start with problems or foundations?
3. **Depth:** More concrete examples or stay conceptual?
4. **Audience:** How technical should we assume?
5. **Length:** 60 min? Need cuts?
6. **Metaphors:** "Sponge" working? Other analogies?
7. **Missing:** What from handwritten notes didn't make it?
8. **Emphasis:** Which tracks most important? Which cuttable?

### What Needs Clarification

- Are the handwritten notes adding major themes I'm missing?
- Should we emphasize cost savings more or is that too "sales-y"?
- How much should we compare to existing systems?
- Should we have concrete code examples?

### Potential Cuts if Too Long

- Track 5 (multimodal) - might be cutable
- Track 6 (sub-agents) - could be brief mention  
- Track 9 (open questions) - could be shortened

### What Might Be Missing

- Concrete implementation sketch?
- Comparison table to existing systems?
- Cost/benefit analysis?
- Timeline/feasibility discussion?

## Tone Guidance

This is an **exploration**, not a sales pitch. The talk should feel like:

✅ "Here's what seems questionable to me"  
✅ "If I had to design from scratch, these thoughts occur to me"  
✅ "I'm uncertain about X"  
✅ "Have you considered...?"  
✅ "What am I missing?"

NOT:

❌ "This is clearly the right approach"  
❌ "Current systems are broken"  
❌ "You should implement this"

## How to Use These Materials

1. **Start with:** TRACKS-OVERVIEW.md for quick picture
2. **Dive into:** ROUGH-DRAFT-BREAKDOWN.md for detailed structure
3. **Consider alternatives:** ALTERNATIVE-STRUCTURES.md for different flows
4. **Preserve voice:** KEY-QUOTES-AND-THEMES.md for memorable phrases
5. **Reference source:** context-thoughts.md for full details

## Format Considerations

- **Slides Only:** Clean, professional, good for conference
- **Whiteboard:** Intimate, exploratory, good for small group
- **Mixed:** Slides + whiteboard for flexibility
- **Demo-Driven:** Show working prototype (if we have one)

Recommended: **Medium density** (40-50 slides)
- Title slide
- One overview slide per track
- 3-5 content slides per track
- Transition slides between sections

## Author's Note

These drafts attempt to preserve the exploratory, humble tone from the source material while organizing into a coherent talk structure. The emphasis is on "here are some thoughts" rather than "here's what you should do."

The breakdown is intentionally detailed to facilitate discussion - we can always simplify, but it's harder to expand from sparse notes.

All quoted material preserves the original phrasing to maintain authenticity and voice.

