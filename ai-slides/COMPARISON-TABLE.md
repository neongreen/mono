# Context Engineering Approaches - Comparison Table

> Quick reference for comparing different approaches to context management

---

## Current Systems Comparison

| System | Approach | User Interface | Over-Capture | Under-Capture | Main Problem |
|--------|----------|----------------|--------------|---------------|--------------|
| **Cursor** | Mutable memory with explicit capture | "Remember this?" popup | Low | High | Users hesitant to commit |
| **ChatGPT** | Mutable memory with implicit extraction | Silent automatic | High (role-play as prefs) | High (misses subtle prefs) | Both overshoots and undershoots |
| **Claude Code** | Manual rules file | User edits `.claude.md` | N/A | Very High | User must be eval expert |
| **Proposed** | Immutable log with functional views | Automatic with retroactive re-eval | None (keeps everything) | None (derives views) | Storage & privacy challenges |

---

## Architecture Comparison

### Mutable Context Model (Current)

```
User Input → Context → Compaction → Mutated Context → Model
                ↑                           |
                └───────── Loop ────────────┘
```

**Characteristics:**
- ❌ Semantics unclear ("whatever compaction kept")
- ❌ Can't reconstruct past states
- ❌ Can't do retrospective A/B
- ❌ Priority balancing fragile (CSS precedence problem)
- ✅ Lower storage costs
- ✅ Simpler implementation

### Immutable Log Model (Proposed)

```
User Input → Immutable Log → [Sponge₁, Sponge₂, ..., Spongeₙ] → Views → Models
                                   ↓         ↓            ↓
                               User-Prefs  Errors    Security
```

**Characteristics:**
- ✅ Clear semantics ("everything we know")
- ✅ Full provenance & replay
- ✅ Retrospective re-evaluation
- ✅ Views can be versioned
- ✅ Multiple sponges compose
- ❌ Higher storage costs
- ❌ More complex infrastructure

---

## Multi-Model Strategy Comparison

| Aspect | Single Prefix (Current) | Multi-Prefix (Proposed) |
|--------|-------------------------|-------------------------|
| **Compaction** | One size fits all models | Per-model optimization |
| **Multimodal** | Convert to universal format upfront | Each model gets optimal format |
| **Evaluation** | Changes affect all models | Decoupled per model |
| **Cost** | 90% discount on one prefix | 90% discount on each prefix |
| **A/B Testing** | Must rebuild context | Instant side-by-side |
| **Task Mixing** | Context pollution | Separate prefixes per task |

---

## Preference Capture Comparison

### Explicit (Cursor)
**Mechanism:** Ask "Remember this?"  
**Pros:** User control, no surprises  
**Cons:** Commitment anxiety, interrupts flow, misses implicit patterns  
**Quote:** "Should I remember this angry rant forever? Feels queasy."

### Implicit (ChatGPT)
**Mechanism:** Automatic extraction from behavior  
**Pros:** No user burden  
**Cons:** Captures wrong things (role-play), misses right things (reverts), can't distinguish strength  
**Quote:** "User wants to be a therapist? Wants replies in Norwegian?"

### Manual (Claude Code)
**Mechanism:** User writes rules file  
**Pros:** Precise control  
**Cons:** Most users won't do it, requires eval expertise, doesn't capture implicit preferences  
**Quote:** "User has to be the eval team. Most preferences never get written down."

### Derived (Proposed)
**Mechanism:** Functional views over full log  
**Pros:** No commitment needed, captures everything, can infer strength, retrospective re-eval  
**Cons:** Requires full log storage, complex view logic  
**Quote:** "Record territory, derive any map. Retroactively recompute with new signals."

---

## Problem Space Matrix

| Problem | Mutable | Immutable |
|---------|---------|-----------|
| **User repeats prefs across projects** | Re-state each time | Automatically detected |
| **User reverts a preference** | Stuck in memory | Can mark as reverted retroactively |
| **Need to A/B new extraction logic** | Wait for new data | Apply to entire history |
| **Different models need different context** | Forced compromise | Parallel prefixes |
| **Multimodal input** | Pre-convert to universal | Each model gets optimal |
| **Sub-agent needs state** | Manual passing | Reuse prefix automatically |
| **Want to audit past decision** | Can't reconstruct | Full replay |
| **Storage costs** | ✅ Lower | ❌ Higher |
| **Privacy & governance** | ✅ Simpler | ❌ More complex |

---

## Economic Impact

### Current Approach
```
Single Prefix (all models) → 90% discount
Sub-agent calls → No discount (throwaway context)
A/B testing → Full cost (rebuild context)
```

### Proposed Approach
```
Multiple Prefixes (per model) → 90% discount each
Sub-agent calls → 90% discount (reuse prefix)
A/B testing → No rebuild cost (prefixes ready)
```

**Key Insight:** 90% discount changes from constraint to enabler

---

## Implementation Complexity

| Component | Mutable | Immutable |
|-----------|---------|-----------|
| **Storage** | Simple (current state) | Complex (full log + indexing) |
| **Compaction Logic** | Complex (one-time decision) | Simple (pure function, versioned) |
| **Multi-Model** | Complex (single compromise) | Simpler (independent views) |
| **Evaluation** | Complex (spooky action) | Simpler (decoupled per model) |
| **Debugging** | Hard (can't replay) | Easy (full history) |
| **Privacy** | Simpler (less data) | Complex (more data + governance) |

---

## User Experience Comparison

### Scenario: New Project

**Mutable:**
1. User starts project
2. Agent starts from zero
3. User re-states: "Use pnpm", "No inline CSS", etc.
4. Agent "remembers" some, misses others
5. Next project: repeat

**Immutable:**
1. User starts project
2. Sponge detects: "Stated 'use pnpm' in 3 projects"
3. Sponge infers: High-strength preference
4. Agent knows this automatically
5. Next project: preference already there

### Scenario: User Rant

**Cursor (Explicit):**
- Popup: "Remember this?"
- User: "Uh... maybe? Forever? Not sure..."
- Hesitates, might decline
- Preference not captured

**ChatGPT (Implicit):**
- Silently extracts
- Might overshoot (role-play → permanent pref)
- Might undershoot (ignore genuine rant)
- User surprised later

**Immutable (Derived):**
- Everything recorded
- Sponge checks: "Rant + action taken + not reverted?"
- If confirmed multiple times: high strength
- If reverted: low strength
- Can recompute anytime

### Scenario: Mixed Tasks

**Single Prefix:**
- User: "Fix build errors" (adds to context)
- User: "Now write all the docs" (adds to context)
- User: "Back to build" (context polluted with docs)
- Must manage manually

**Multi-Prefix:**
- User: "Fix build" → Fast-Model prefix
- User: "Write docs" → Different prefix (or sub-agent)
- User: "Back to build" → Resume original prefix
- Automatic separation

---

## Trade-off Summary

### When Mutable Makes Sense
- ✅ Storage is very constrained
- ✅ Privacy/compliance requires minimal retention
- ✅ Simple use cases (single session)
- ✅ No need for retrospective analysis
- ✅ Evaluation is stable (not evolving)

### When Immutable Makes Sense
- ✅ Multi-project, long-term usage
- ✅ Need to learn user preferences over time
- ✅ Evaluation logic evolving rapidly
- ✅ Multiple models with different needs
- ✅ Debugging & auditing important
- ✅ A/B testing is frequent

---

## Open Questions

| Question | Impact | Answer Needed By |
|----------|--------|------------------|
| How much does full log storage cost? | Critical | Phase 1 |
| Can we get 90% discount on multiple concurrent prefixes? | Critical | Phase 1 |
| What's the latency with N active prefixes? | Important | Phase 2 |
| How to handle PII in immutable log? | Critical | Phase 1 |
| Can views be versioned at git-sha for reproducibility? | Nice-to-have | Phase 3 |
| How to resolve conflicting sponge outputs? | Important | Phase 2 |
| At what log size does query performance degrade? | Important | Phase 2 |

---

## Decision Framework

### Choose Mutable If:
- [ ] Storage/privacy is primary concern
- [ ] Single-session use cases dominate
- [ ] Evaluation is stable/mature
- [ ] One model, one task at a time
- [ ] No need for historical analysis

### Choose Immutable If:
- [ ] Multi-project, long-term usage
- [ ] User preference learning is key
- [ ] Multiple models with different needs
- [ ] Evaluation logic rapidly evolving
- [ ] Debugging & auditing important
- [ ] A/B testing frequent

### Hybrid Approach:
- [ ] Immutable log for X days/weeks
- [ ] Compaction after threshold
- [ ] Keep derived summaries forever
- [ ] Balances both concerns

---

## Summary Scores

| Criterion | Mutable | Immutable |
|-----------|---------|-----------|
| **Storage Cost** | ⭐⭐⭐⭐⭐ | ⭐⭐ |
| **Implementation Simplicity** | ⭐⭐⭐⭐ | ⭐⭐⭐ |
| **Privacy/Governance** | ⭐⭐⭐⭐ | ⭐⭐ |
| **User Preference Learning** | ⭐⭐ | ⭐⭐⭐⭐⭐ |
| **Multi-Model Support** | ⭐⭐ | ⭐⭐⭐⭐⭐ |
| **Debugging/Auditing** | ⭐⭐ | ⭐⭐⭐⭐⭐ |
| **A/B Testing** | ⭐⭐ | ⭐⭐⭐⭐⭐ |
| **Evaluation Evolution** | ⭐⭐ | ⭐⭐⭐⭐⭐ |
| **Clear Semantics** | ⭐⭐ | ⭐⭐⭐⭐⭐ |
| **Developer Experience** | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ |

**Overall:** Immutable wins on capability, mutable wins on constraints

---

## Visualization Ideas

### For Talk Slides

1. **The Compaction Loop** (Track 2)
```
Context → [?] Compaction [?] → ??? Context → Model
    ↑________________________________|
           "What's left?"
```

2. **CSS Precedence Problem** (Track 2)
```
Priority 1: !important
Priority 2: !!important  ← Running out of headroom
Priority 3: !!!important ← This is absurd
```

3. **Preference Capture Landscape** (Track 3)
```
           Over-Capture
                ↑
   ChatGPT ●    |
                |
Explicit ←──────┼──────→ Implicit
                |
      Cursor ●  |
                ↓
           Under-Capture
```

4. **Multi-Prefix Architecture** (Track 4)
```
               Router
                 |
    ┌────────────┼────────────┐
    ↓            ↓            ↓
Claude-Prefix Gemini-Prefix Fast-Prefix
    90%          90%          90%
```

5. **Sponge Pipeline** (Track 7)
```
Immutable Log → [Sponge₁] → User Prefs View
             → [Sponge₂] → Error Pattern View
             → [Sponge₃] → Security View
```

6. **Mutable vs Immutable** (Track 8)
```
Mutable:   X → Y → Z  (lost X and Y)
Immutable: X + Y + Z → [view₁, view₂, view₃]
```

