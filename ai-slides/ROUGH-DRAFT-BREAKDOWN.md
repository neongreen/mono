# Context Engineering - Rough Draft Breakdown

> **Note:** This is a ROUGH DRAFT for discussion. This breakdown organizes the AI slides material into tracks/topics for an exploratory talk about context engineering thoughts.

## Overview

This talk explores context engineering from a "design from scratch" perspective - questioning common practices and presenting alternative approaches with humility. Not trying to convince, just sharing observations and thoughts.

---

## TRACK 1: The Fundamental Question - What IS Context?

**Core idea:** Before we optimize context, let's be clear about what we mean.

### Slides:
1. **Opening Question**
   - "What do we mean by 'context' in AI systems?"
   - Not just current conversation
   - Everything the system "knows" at time T

2. **Context Object Anatomy**
   - User messages (across all projects/sessions)
   - Code files & diffs
   - Tool outputs
   - Knowledge bases, rules
   - Analytics, DB lookups
   - Key question: Should we ever delete any of this?

3. **Two Mental Models**
   - **Model A:** Session-bound, mutable, gets compacted
   - **Model B:** Infinite log, immutable, derive views
   - Most systems pick A by default. Why?

### Discussion points:
- Is the choice between A and B actually arbitrary, or are there fundamental trade-offs?
- What are the hidden assumptions in "context management"?

---

## TRACK 2: The Mutable Context Problem

**Core idea:** When context can be mutated, we lose the ability to reason about it.

### Slides:
1. **The Compaction Loop**
   - Diagram: Context → Compaction → Mutated Context → Use
   - Seems efficient, but what's the cost?
   - Once compacted, can you reconstruct what was known?

2. **Loss of Provenance**
   - Example: "Why did the AI give this answer?"
   - Post-compaction: Cannot reconstruct the full knowledge state
   - Prevents retrospective A/B testing
   - Can't replay with different eval criteria

3. **The Semantics Problem**
   - In immutable model: Context = "everything we know"
   - In mutable model: Context = "whatever the compaction process kept"
   - Quote: *"The semantics of the context type become whatever the evals team approves"*
   - We're optimizing for scores, not meaning

4. **The CSS Precedence Metaphor**
   - First important thing: mark it "!important"
   - Second more important thing: now what?
   - Running out of headroom for priority
   - Every hack pushes you toward immutable model anyway

### Discussion points:
- Is this a real problem or just theoretical?
- Are there scenarios where mutable context is actually better?

---

## TRACK 3: User Preferences - The Implicit vs Explicit Problem

**Core idea:** How do we capture what users actually want without making them uncomfortable?

### Slides:
1. **The Preference Capture Landscape**
   - Three approaches:
     - **Cursor:** Explicit "Remember this?" popups
     - **ChatGPT:** Implicit heuristic extraction
     - **Claude Code:** Manual `.claude.md` rules files
   - Each has different failure modes

2. **The Rant Problem**
   - Users repeat the same preferences across projects
   - "Use pnpm not npm", "No default ChatGPT style", "Python scripts not bash"
   - Gets tiring: dread starting new projects
   - But: Hesitant to commit to explicit rules

3. **Why Explicit Fails**
   - "Should I remember this angry rant forever?"
   - Users feel queasy about commitment
   - Don't know how strongly the rule will be enforced
   - Might be wrong, might need exceptions

4. **Why Implicit Fails**
   - ChatGPT: Remembers role-play as permanent preference
   - "Oh, user wants me to be a therapist forever"
   - "User wants replies in Norwegian" (was just practice)
   - Misses genuine preferences from reverts

5. **The Strength Problem**
   - Once recorded, how strong is this preference?
   - All preferences look equally strong
   - Users can't annotate "except cases A, B, C"
   - Can't distinguish weak hunches from strong convictions

### Discussion points:
- Is there a middle way between explicit and implicit?
- Can we infer preference strength from behavior patterns?

---

## TRACK 4: Multi-Model Economics - The Shared Prefix Opportunity

**Core idea:** 90% cost discount for shared prefixes changes everything.

### Slides:
1. **The Pricing Mechanic**
   - When prompt shares prefix with previous: 90% off
   - This is not a minor optimization
   - Fundamentally changes architecture possibilities

2. **The Single-Prefix Straitjacket**
   - Current approach: One compaction for all models
   - But different models need different contexts
   - Claude: smaller, XML tags
   - Gemini: can handle full video
   - Fast models: ultra-trimmed
   - Forcing one size to fit all

3. **Multi-Prefix Architecture**
   - Maintain parallel shared prefixes
   - Claude-Prefix, Gemini-Prefix, Fast-Prefix
   - Router picks model per request
   - Still get 90% discount on each
   - No more lowest-common-denominator

4. **Decoupled Evaluation**
   - Changing Gemini's compaction doesn't affect Claude
   - No spooky action at a distance
   - Each model family optimized independently
   - Easier A/B testing

5. **Enabling Easy Comparisons**
   - Two prefixes ready: show user both answers
   - "Not sure which model is better for this"
   - User picks, system learns
   - No rebuild cost

### Discussion points:
- How many parallel prefixes before overhead becomes problem?
- What's the latency impact?

---

## TRACK 5: Multimodal Inputs - The Lossy Conversion Problem

**Core idea:** Pre-converting multimodal to text throws away information.

### Slides:
1. **The Reality of Modern Input**
   - Voice dictation
   - Screen recordings
   - Screenshots
   - Video walkthroughs
   - Most instructions now via voice/video

2. **The Premature Optimization**
   - Current: Convert to universal format upfront
   - Voice → text transcription
   - Video → summary + keyframes
   - Why? To fit in all models
   - Problem: Running "in the dark"

3. **The Context Availability Problem**
   - Don't know what future questions will be
   - Summary might miss crucial details
   - Gemini could handle full video cheaply
   - But we force it to use degraded summary

4. **Multi-Prefix Solution**
   - Gemini-Prefix: full video
   - Claude-Prefix: summary + keyframes
   - O3-Prefix: terse summary only
   - Each model gets optimal format
   - No premature information loss

### Discussion points:
- Storage costs for keeping raw media?
- Is this over-engineering?

---

## TRACK 6: Sub-Agents & Tools - Rethinking the Abstraction

**Core idea:** Sub-agents are just LLM tool calls with their own prefixes.

### Slides:
1. **What's a Sub-Agent?**
   - Tool call where tool = LLM
   - Parameter = prompt surface
   - Gets its own context view
   - Examples: linter, doc generator, security scanner

2. **The Persistent Prefix Trick**
   - Sub-agent finishes: its final prompt remains
   - Next invocation: reuse as prefix
   - Instant 90% discount
   - Automatic state continuity

3. **Natural Use Cases**
   - **Iterative linter:** Remembers previous fixes
   - **Security scanner:** Tracks what was fixed vs ignored
   - **Doc generator:** Knows which files documented
   - No manual state-passing needed

4. **The Background Agent Pattern**
   - Long-running improvement suggesters
   - Code quality monitors
   - Each maintains its own evolving prefix
   - Not throwaway prompts each time

### Discussion points:
- When do sub-agents make sense vs simple tool calls?
- How to handle sub-agent context divergence?

---

## TRACK 7: The Sponge Metaphor - Functional Views

**Core idea:** Instead of mutating context, derive views from immutable log.

### Slides:
1. **What's a Sponge?**
   - Pure function over the immutable log
   - "Absorbs" relevant information
   - Outputs derived context
   - Many sponges can coexist

2. **Example Sponges**
   - **User-Preferences Sponge:** Rants, reverts, repeated actions
   - **Error-Pattern Sponge:** "Stop generating 300-line hashes"
   - **Security-Scanner View:** Past findings & fixes
   - Each specialized for different concern

3. **Strength Inference**
   - Preference seen in 3 projects: ×10 weight
   - Reverted within session: ×0.1 weight
   - Reinforced over time: boost
   - All computable from immutable log

4. **The Composition Benefit**
   - Sponges don't interfere
   - Each has own prompt budget
   - Easy to toggle on/off for A/B
   - No global rebalancing

5. **Retrospective Re-evaluation**
   - New idea: "Check if preference was later reverted"
   - Mutable model: Only applies to future
   - Immutable model: Apply to entire history instantly
   - Immediate validation

### Discussion points:
- Is "sponge" a good metaphor or confusing?
- How many sponges before it's too complex?

---

## TRACK 8: Why Immutable Wins (The Case Summary)

**Core idea:** Not just a preference - it's strictly more powerful.

### Slides:
1. **The Strict Superset Argument**
   - Any mutable scheme = particular view of immutable log
   - Can simulate mutable by choosing to only look at "latest"
   - But also can do things mutable can't
   - Therefore: strictly more powerful

2. **Six Concrete Benefits**
   - **Provenance:** Replay exact state at any T₀
   - **Debugging:** See what system knew when it answered
   - **Experimentation:** Toggle features on old data
   - **User Trust:** Never forced to decide "forever?" in moment
   - **Model Flexibility:** Different compactions per model
   - **Fast A/B:** No waiting for fresh data

3. **The Developer Experience Win**
   - Add toggle: "also consider preference reverts"
   - See output against whole history immediately
   - Dogfooding becomes possible
   - "Is it true I want agent to be therapist? Nope."

4. **The "Record Importance" Trap**
   - Mutable: Store "importance number"
   - But that's what your *old model* thought was important
   - When you change importance definition: stuck
   - Immutable: Re-derive from territory anytime

### Discussion points:
- Is the storage cost worth it?
- When do we actually need to delete data?

---

## TRACK 9: Open Questions & Challenges

**Core idea:** This approach isn't free lunch - here are the hard parts.

### Slides:
1. **Storage & Cost**
   - How much raw data can we keep?
   - Compression strategies
   - Pruning policies for huge binaries
   - Indexing for fast view queries

2. **Governance & Privacy**
   - Who can query the full log?
   - PII redaction pipelines
   - Right to be forgotten
   - Cross-project data sharing consent

3. **Evaluation Framework**
   - How to benchmark new sponges?
   - Metrics: revert rate, preference satisfaction
   - Comparing to existing systems
   - Avoiding local maxima

4. **Multi-Prefix Infrastructure**
   - How many prefixes before latency balloons?
   - Hot-swap when router picks new model
   - Cache invalidation strategies
   - Memory constraints

5. **Conflict Resolution**
   - Two sponges emit contradictory preferences
   - How to rank/prioritize?
   - Recency vs strength vs user vote?
   - Making conflicts visible to user?

6. **Versioning of Views**
   - Should sponges be versioned at git-sha?
   - Reproducibility vs evolution
   - How to handle view schema changes?

### Discussion points:
- Which challenges are showstoppers vs solvable?
- What would prove this approach wrong?

---

## TRACK 10: Closing Thoughts - A Different Design Space

**Core idea:** Stepping back and questioning assumptions opens new possibilities.

### Slides:
1. **The Journey**
   - Started: "Context management is expensive"
   - Asked: "What if we never delete anything?"
   - Discovered: Fundamentally different architecture
   - Not just optimization - different paradigm

2. **Common Practice Questioned**
   - "Obviously we must compact context" - must we?
   - "Users must explicitly state preferences" - do they?
   - "One context for all models" - why?
   - "Process multimodal upfront" - really?

3. **The Humility Angle**
   - These are explorations, not proven solutions
   - Many unknowns remain
   - Storage costs might be prohibitive
   - But: worth exploring systematically

4. **What Would Change Your Mind?**
   - If storage costs grow unbounded?
   - If privacy constraints make it impossible?
   - If latency with multiple prefixes too high?
   - If users hate seeing old mistakes resurface?

5. **Next Steps (For Someone)**
   - Prototype one sponge
   - Measure storage growth over time
   - Test multi-prefix latency
   - User study: implicit vs explicit preferences
   - Build eval framework

### Discussion points:
- What experiments would be most valuable?
- Is this solving a real problem or inventing one?

---

## Meta-Structure for the Talk

### Opening (5 min)
- Personal context: Why I'm thinking about this
- Not here to convince, here to explore
- "If I had to design from scratch" perspective

### Core Tracks (40-50 min)
- Work through tracks 1-9 above
- Each track: problem → current approaches → alternative thinking
- Leave space for questions/reactions

### Closing (5-10 min)
- Track 10: Summary and open questions
- Invitation for feedback and critique
- "What am I missing?"

### Tone Throughout
- Humble, exploratory
- "Here's what seems questionable to me"
- Acknowledge uncertainties
- Not "you should do this" but "have you considered...?"

---

## Alternative Organizational Schemes

If the 10-track structure feels too fragmented, could reorganize as:

### 3 Big Themes:
1. **The Immutable Foundation** (Tracks 1, 2, 8)
2. **User Preferences & Multi-Model Economics** (Tracks 3, 4, 5)
3. **Implementation Patterns** (Tracks 6, 7, 9, 10)

### 5 Questions Format:
1. What is context really?
2. Why does mutation cause problems?
3. How should we capture preferences?
4. Can we exploit multi-model economics?
5. What are the hard parts?

---

## Notes for Discussion

**Things that need clarification:**
- Are the handwritten notes adding major themes I'm missing?
- Is the "sponge" metaphor helpful or should we find better framing?
- Should we emphasize cost savings more or is that too "sales-y"?
- How technical should the audience be assumed to be?

**Things I'm uncertain about:**
- Does Track 4 (multi-prefix) belong earlier as it's foundational?
- Is Track 6 (sub-agents) a distraction from main themes?
- Should we have more concrete code examples vs conceptual discussion?

**Potential cuts if too long:**
- Track 5 (multimodal) might be cutable
- Track 6 (sub-agents) could be brief mention
- Track 9 (open questions) could be shortened

**What's missing:**
- Concrete implementation sketch?
- Comparison table to existing systems?
- Cost/benefit analysis?
- Timeline/feasibility discussion?

