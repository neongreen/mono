# Context Engineering Talk - Tracks Overview

> Quick reference for discussing the rough draft breakdown

## The 10 Tracks (with key questions)

### 🎯 TRACK 1: What IS Context?
**Time:** ~5 min  
**Key Idea:** Define the problem space before solving it  
**Question:** Should context ever be deleted?

---

### 🔄 TRACK 2: The Mutable Context Problem
**Time:** ~8 min  
**Key Idea:** Mutation makes semantics unclear  
**Metaphor:** CSS precedence problem  
**Question:** Can we reason about mutated context?

---

### 👤 TRACK 3: User Preferences (Implicit vs Explicit)
**Time:** ~8 min  
**Key Idea:** Both Cursor and ChatGPT fail differently  
**Pain Point:** "I dread starting new projects"  
**Question:** How to capture preference strength?

---

### 💰 TRACK 4: Multi-Model Economics
**Time:** ~7 min  
**Key Idea:** 90% discount unlocks parallel prefixes  
**Benefit:** Decoupled evaluation  
**Question:** How many prefixes before overhead hits?

---

### 🎬 TRACK 5: Multimodal Inputs
**Time:** ~6 min  
**Key Idea:** Don't pre-convert and lose information  
**Example:** Video in Gemini, summary in Claude  
**Question:** Worth the storage cost?

---

### 🤖 TRACK 6: Sub-Agents & Tools
**Time:** ~6 min  
**Key Idea:** Sub-agents reuse their own prefix  
**Pattern:** Security scanner, linter, doc generator  
**Question:** When vs simple tool calls?

---

### 🧽 TRACK 7: The Sponge Metaphor
**Time:** ~8 min  
**Key Idea:** Derive views, don't mutate  
**Examples:** Preference sponge, error-pattern sponge  
**Power:** Retrospective re-evaluation  
**Question:** Is "sponge" good metaphor?

---

### ✅ TRACK 8: Why Immutable Wins
**Time:** ~7 min  
**Key Idea:** Strict superset of mutable  
**Benefits:** Provenance, debugging, experimentation, trust  
**Question:** Is storage cost worth it?

---

### ❓ TRACK 9: Open Questions & Challenges
**Time:** ~6 min  
**Key Idea:** Not free lunch - hard parts  
**Issues:** Storage, privacy, evaluation, conflicts  
**Question:** Which are showstoppers?

---

### 🎬 TRACK 10: Closing - Different Design Space
**Time:** ~5 min  
**Key Idea:** Questioning assumptions opens possibilities  
**Tone:** Humble, explorative, uncertain  
**Question:** What would prove this wrong?

---

## Flow Options

### Option A: Linear (as above)
Build up from foundations → applications → challenges

**Pros:** Logical progression  
**Cons:** Might lose audience before payoff

### Option B: Problem-First
Start with Track 3 (user pain), then explain why (Tracks 1-2), then solutions (4-8), then reality (9-10)

**Pros:** Hook audience immediately  
**Cons:** Less systematic

### Option C: Three Acts
- **Act 1:** The Problem Space (Tracks 1-3)
- **Act 2:** The Alternative Architecture (Tracks 4-7)
- **Act 3:** Reality Check (Tracks 8-10)

**Pros:** Story-like structure  
**Cons:** Track 8 feels like it should be in Act 2

---

## Key Themes Across Tracks

### Theme: "Optimizing for the Wrong Thing"
- Track 2: Optimizing for eval scores, not semantics
- Track 3: Optimizing for explicit capture, not actual preferences
- Track 5: Optimizing for universal format, not model capabilities

### Theme: "The Power of Never Deleting"
- Track 2: Provenance & A/B testing
- Track 7: Retrospective re-evaluation
- Track 8: Strict superset argument

### Theme: "Economic Incentives Shape Architecture"
- Track 4: 90% discount enables multi-prefix
- Track 6: Sub-agents reuse prefixes
- Track 9: Storage costs are limiting factor

### Theme: "User Experience vs System Constraints"
- Track 3: Don't make users be eval experts
- Track 5: Users mix tasks in conversations
- Track 9: Privacy & governance

---

## Potential Demo/Concrete Examples

If we want to make it less abstract:

### Example 1: The Rant Scenario
"In Project A, I say 'use pnpm not npm'  
In Project B, I say 'use pnpm not npm' again  
In Project C, I say it a third time  
Now I start Project D..."

**Question:** Should the system learn? How?

### Example 2: The Revert Pattern
"User: Add component library X  
Agent: Added library X  
User: Actually, revert that, library X is bad  
Agent: Reverted  
*Later in Project B*  
Should system suggest library X? Or remember it failed?"

### Example 3: The Video Input
"User records 2-minute walkthrough clicking UI  
Gemini: Can process full video  
Claude: Needs summary + keyframes  
O3: Too expensive, just summary  
Do we force everyone to work from same summary?"

---

## Visual/Diagram Opportunities

1. **Track 2:** Mutable compaction loop diagram
2. **Track 4:** Parallel prefixes architecture
3. **Track 7:** Sponge/filter pipeline visualization
4. **Track 8:** Immutable log with multiple views
5. **Comparison table:** Cursor vs ChatGPT vs Claude Code preferences

---

## Tone Calibration

### Too Confident
❌ "This is clearly the right approach"  
❌ "Current systems are broken"  
❌ "You should implement this"

### Right Level
✅ "This seems questionable to me"  
✅ "I'm uncertain about X"  
✅ "Have you considered...?"  
✅ "What am I missing?"  
✅ "Here's what occurs to me"

### Too Uncertain
❌ "Maybe nothing I'm saying makes sense"  
❌ "This is probably all wrong"  
❌ "I don't really know anything"

---

## Discussion Questions for Review

1. **Scope:** Is 10 tracks too many? Should we combine some?

2. **Order:** Start with problems or start with foundations?

3. **Depth:** More concrete examples or keep conceptual?

4. **Audience:** How technical? (Engineers? Product? Mixed?)

5. **Length:** 60 min total? Need cuts?

6. **Metaphors:** "Sponge" working? "CSS precedence"? Other analogies?

7. **Missing:** What major themes from handwritten notes didn't make it?

8. **Emphasis:** Which tracks are most important? Which cuttable?

9. **Format:** Slides? Whiteboard? Demo? Mix?

10. **Goal:** Pure exploration or gentle advocacy?

