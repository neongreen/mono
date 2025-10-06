# Immutable-vs-Mutable Context — RFC DRAFT v2

> Goal: articulate why an **“infinite immutable log + functional views”** beats a **mutable-compaction pipeline**, and surface every nuance (preferences, multimodal, multi-model prefixes, governance, etc.).
> Note: This commentary is based only on my messages, the `@context-thoughts.md` notes, and the diagram from yesterday’s call. I’m new to the team and might be misinterpreting parts of the proposal (?). Please read with that caveat.

---

## 0 · Primer — What We Mean by “Context”

- **Context object** = *everything the system “knows” at time t*  
  - All user messages (any project, any session)  
  - Code files & diffs  
  - Tool outputs, DB look-ups, analytics  
  - Knowledge bases, rule files  
- **Key property (our proposal)**: append-only, never destroyed  
- Contrast with diagram from last call:  
  - They create a *Context → Compaction → Mutated-Context* loop (mutable)

- Wording anchors (to avoid ambiguity later):
  - "User messages and conversations" — this can be a subpoint or not even a subpoint; either way, it must live inside the context object explicitly (no silent elision).
  - "Append-only" here is literal: we never rewrite history; we only derive views.
    - > "Context is an append-only log ... By definition, it's immutable - everything you know."
    - No reason ever to mutate context; when you learn more, old knowledge remains accessible.
  - "Compaction" (in our framing) = view-building, not destructive mutation.
    - > "Compaction could be a mutable operation, or a view of the context, or a pure function from the context to whatever. Could be a smaller context. Could be a string."

---

## 1 · Multiple Problems We’re Tackling (keep separate!)

1. **Mutable compaction layer**  
   - Hard to reason about *semantics of the context type*  
     - Say which semantics: semantics of the **context data type** itself.  
     - In the immutable model, context = "everything we know" (subject to practical size caps) and you can still point to older knowledge like "How did this file look before?"  
       - > "When learning more, you still want to refer to old knowledge — 'How did this file look before?' 'When did the user do the revert?'"
     - In the mutable model, once a compactor mutates the context, the meaning becomes: "whatever the compaction procedure (approved by the evals team) decided to keep."  
       - > "If the context thing is mutated by the compactor, then suddenly the semantics of the context type are whatever compaction process ends up being approved by the evals team or maximizes the eval given by the evals team. We don't know how to work with this. The semantics exist in a way you cannot reason about."
       - Decision-making burden shifts: you're trying to forecast the future and read the mind of the evals team.
     - This leads to "semantics we can't reason about" — we are effectively optimizing for scores, not meaning.  
       - > "Semantics become: 'stuff we ended up while running this loop' ... Whatever ends up after loop iterations and iterations ... Something that the evil team is happy with."
   - **Loss of provenance & history** (expanded)  
     - Post-compaction, some artefacts are irretrievably gone → you cannot reconstruct what was known at the time of an answer.  
     - Prevents instant A/B re-evaluation over past conversations because only the outputs of compaction remain, not the inputs.  
     - Semantics drift into "stuff we ended up with after loop iterations that seemed to work."  
   - **CSS precedence problem** (your metaphor)  
     - Adding new importance to a memory requires rebaking old choices ("!important" stacking).  
       - > "You end up with a sort of CSS precedence problem where you have a bunch of stuff, and then you want to add another thing, and you think it's important, so you say, you know, you say exclamation mark important. But then it turns out that there's a third thing that's actually even more important you've discovered, and well, there's only one level of exclamation mark important."
     - When you discover an even-more-important rule, there's no more headroom. Balancing becomes fragile and global.
       - > "Doing eval on this or doing any sort of A/B on this where this is adding the right amount of caps lock and important and exclamation marks to the context ... it makes efficient A/B, I think, very hard."
       - > "Every hack you add like this, where instead of recording something, you record richer information that you're going to translate into the prompts later, it all brings you to the immutable, the infinite immutable context model."
2. **User-preference capture**  
   - Reality: users repeat rants across projects — e.g., "Use pnpm, not npm", "Prefer Python scripts over bash", "Don't write inline CSS", "No TODO-later comments", "Don't use default-chatty style."  
     - > "I never want npm. I never want a default ChatGPT style. This style of talking is terrible. I never want to write bash scripts because it's also terrible. I want Python scripts. I never want to leave comments like 'to do later.' I never want to write CSS in inline HTML."
     - > "If I rant about a thing in three projects, I'm starting to dread starting a new project because I know I'll have to repeat the same things over and over or fight the same battles over and over."
   - Cursor-style explicit prompts ("Should I remember this?") cause hesitation — I don't want an angry rant to follow me forever; I'm not sure in-the-moment.  
     - > "As a user, do I want to say that when Cursor tells me, 'Oh, here's a memory. Should I remember it? Do I want to remember it?' Am I comfortable telling Cursor that, 'Yeah, my angry rant right now should follow always forever?' Nah, I'm feeling queasy about it."
     - > "I don't want to commit to actually making this my preference. I don't know how much Cursor is going to stick to this preference. Maybe it's going to treat it like an ironclad rule and it's going to be bad."
     - User might have valid preferences but be hesitant to commit explicitly:
       - > "I might have preferences that the agent should take into account, but that if asked explicitly, I would be hesitant to commit to."
   - ChatGPT-style implicit capture overshoots and undershoots: it may remember a one-off role-play ("therapist", "reply in Norwegian") as a permanent preference while missing genuine preferences expressed via reverts and repeated actions.  
     - > "I sometimes have a conversation and I say, 'Now we're going to pretend you're a therapist,' and then I look at the memories and I see, 'Oh, the user prefers ChatGPT to be a therapist or a Spanish woman or whoever.' And I'm like, 'No, I'm not going to use this again because apparently you're going to remember that you should reply in Norwegian because I wanted to practice Norwegian.'"
     - > "Simultaneously, it overshoots. It takes a thing which is not a preference at all but from a single chat or a single rant. Maybe the user even misunderstood. Like the user says, 'Oh, you should use this component library,' and then eventually it turns out the component library is shit and the user rolls back. But you have somehow decided, for example, to retain the first preference but not the second."
   - Strength/weight is unknowable in mutable memory: everything looks equally strong unless the user perfectly annotates exceptions (A/B/C), which is unrealistic.
     - > "Once you have recorded something, you have no way of telling if you really recorded it for very valid reasons or for kind of not very reliable reasons."
     - > "You cannot distinguish those. You cannot notice which are strong preferences unless the user has had the foresight to say that this thing is a really, really, incredibly strong preference. It must always be obeyed."
     - Users can't anticipate edge cases:
       - > "'I never want to X,' except obviously for cases A, B, C. The user thought it was obvious, but eh."  
3. **Model diversity & shared-prefix economics**  
   - 90 % cost discount for tokens in common prefixes  
   - Different LLMs need different compactions; one shared prefix is a straitjacket
     - > "Now you definitely want different contexts for different models. You're doing compaction. Compaction depends on how big the prompt is. Compaction results in different contexts for different models."
   - Key constraint this creates:
     - > "You cannot be freely reshuffling the prompt. You cannot do the ideal thing, which is for each request to the LLM say that, 'Hey, you know, now the top 15 most important things are here, here, here,' because, well, then there is no shared prefix."  
4. **Multimodal & long-form inputs**  
   - Voice, video, screenshots: some models can swallow full media, others only summaries.  
     - > "I found that voice suffers from the same lossy, spur-of-the-moment problem. I know that my voice recording—I know that multimodal models can actually handle voice better than purely voice recognition models if it's in context."
   - The current workaround is a non-local "kebab" of tool calls (model asks tool, tool asks Gemini about video, ships back a summary for another model) — hard to make robust and evaluate.
     - > "You have this whole package. While if you have a shared prefix for Gemini and for Claude and they're different ... you're going to be using O3 for this request ... You don't have to think about other models when doing eval."  
5. **Background / sub-agents**  
   - Security-scanner, auto-doc, lint-fixer each pay full prompt cost on every invocation  
   - Reusing their *own* prefix would be 10× cheaper and more capable
   - Background agents map naturally to this:
     - > "If we're talking about background agents, then this very naturally maps to background agents because the background agent still wants to have probably a prefix ... You still want to have a 90% discount here, right? Probably. Surely."
   - Long-running agents like security scanners or code-improvement suggesters need their own evolving prefix, not throwaway prompts each time.

---

## 2 · Two Context Models

### 2.1 · Model A — *Local & Session-Bound*
- Scope = current conversation / project (?)  
- Typically mutable or at least truncated (?)  
- Works fine for: quick chats, throwaway tasks  
- Drawbacks  
  - Re-teach prefs each new project  
  - No cross-pollination of knowledge  
  - Compaction heuristics become whack-a-mole
- Between A and B: ChatGPT/Cursor hybrid approach
  - > "Somewhere between Model A and B is the ChatGPT model, where they have a thing which tries to extract preferences. Cursor has it now as well, right? But it is one-time. They don't try to be declarative or functional about it."
  - > "They say, 'Oh, look at what happened just now. Try to guess if it's relevant and add it to a database.' So they do mutable context. They have to guess in each moment what's relevant, which is—I think it's worse."

### 2.2 · Model B — *Global Immutable Log (preferred)*
- Append-only, covers all artefacts (see §0)  
- Materialised *views* (not mutations) feed prompts  
- Allows retrospective re-evaluation (new metrics, new extractors)  
- Storage can be compressed; logic stays pure

- Your phrasing anchors (Model A vs B):
  - **Model A** = "standard current conversation / current projects model" (?).  
  - **Model B** = taken to its logical conclusion; separate header; includes:  
    - User messages; conversations in other conversations in the same project; even reverted messages; even different projects by the same user.  
      - > "User messages, conversations, even in other conversations in the same project, even reverted messages, even different projects by the same user."
    - Justification for cross-project use: avoid re-stating obvious preferences every time ("pnpm not npm", style, etc.).
      - > "The justification, for example, for using different projects from the same user as context is that it's annoying to have to tell it in every new project every time that, 'Hey, I want to use pnpm, not npm. I want to use... I prefer things to be done this way.'"
      - > "It's like ridiculously low-hanging fruit that a rules file in one project doesn't get taken into account when working in another project."  

---

## 3 · Functional Views = “Context Sponges”

- **Definition:** Pure functions over the log; output *derived contexts*  
- Examples  
  - **User-Preferences Sponge**  
    - Extract rants, reverts, repeated actions  
    - Assign *strength* via heuristics (e.g. ×10 if seen in 3 projects)  
    - Can surface implicit preferences the user didn't articulate:
      - > "Often, for example, I say that you should rewrite this or that, or let's say I revert and do something. When I revert, it's probably because I didn't like something. But it would be tricky for me to explain what I liked."
  - **Error-Pattern Sponge** – "Stop generating 300-line hashes"  
    - > "Stop generating 300-line hashes or like stop repeating the function name in the comment. Like function says 'generate eval,' comment says 'this function generates eval.' That's a useless comment."
  - **Security-Scanner View** – remembers past findings & fixes  
- Benefits  
  1. No mutation — provenance intact  
  2. Multiple sponges compose; each can have its own prompt budget  
  3. Easy A/B: toggle a sponge and replay history instantly
     - > "The model where you remember stuff is—I don't even think you could say it's an optimization, but the model of immutable context forever is strictly more powerful because you could just as well be simply running some sort of salience filter, salience marker."
     - Instead of extracting a "memory" to store, you mark salient messages with flags; summary determines salience. Then you look at starred/flagged messages from the infinite context.
     - > "It's literally the same as what you would have gotten if you were adding those memories into a mutable database, right? But so it's strictly as powerful, but it is also more powerful because you can do other stuff instead."

- Your metaphors / constraints:  
  - "Not a tool call so much as a context filter or squeeze — like a sponge."  
    - > "You could have sub-agents as sort of extractors from the context of fun stuff. So there are views of the context that get the entire context ... But it only extracts stuff like user rants and user reverts and messages like, 'No, no, no, do this, not that.'"
    - > "So sort of squeezing out stuff from the context. It's not like a tool call as in 'give me, use this tool,' as much as it's a context filter or squeeze. Like a sponge—a context sponge."
  - Given the entire context; only *absorbs* things like user rants, reverts, and "No, do this, not that."  
    - > "Go through the context and absorb what is about the area of user preferences or things that the Lovable thing seems to be bad at, or whatever."
    - Each sponge is specialized: one agent is like "extract user preferences," given the entire context because everything is given the entire context.
  - Can also focus on "areas Lovable is bad at" inferred from history.

---

## 4 · Preference-Recording Landscape

| System          | Mechanism                    | Over-Capture                 | Under-Capture               | User Burden                             |
| --------------- | ---------------------------- | ---------------------------- | --------------------------- | --------------------------------------- |
| **Cursor**      | Asks “Remember this?” pop-up | Low                          | High (missed rants)         | Explicit confirmation causes hesitation |
| **ChatGPT**     | Implicit heuristics          | High (role-play “therapist”) | High (ignores subtle prefs) | None                                    |
| **Claude Code** | `claude.md` rules file       | N/A                          | Very High                   | User must be evaluation expert          |

- Immutable log + preference sponge solves both extremes: captures *everything*, lets view decide salience.
- Strength signals can be recomputed any time (reverts, recency, project count, etc.).

- Your concrete pain points to preserve:  
  - Naïve textual prefs backfire ("be concise" → ultra-telegraphic).  
    - > "If I try some naive approach like 'oh, you should be more concise,' it might backfire. I might get a super telegraphic style which I also dislike. I've been burned by this enough times that I don't want to just express my preferences in words and then have the agent perverse them somehow."
  - Memory banks can't distinguish strong from weak prefs; team downstream can neither over- nor under-weight safely.  
    - > "Simultaneously they cannot give too much weight to this because there might be bullshit preferences there. And if they give little weight to this, then things that are actually very strong preferences get lost."
  - Users shouldn't have to be the eval team; most will underuse rule files.
    - > "Well, it's not a good expectation for the user to be good at evals. The user is definitely not going to be good at evals ... I think there's an art to writing the rules files, and I think it's very hard for people because now the user has to be the eval team."
    - > "I do that thing sometimes with the sessions where the final sessions, but I know for sure that many of my preferences that I do angry rants about end up not being recorded in code and rule files."

---

## 5 · Multi-Model, Multi-Prefix Strategy

- Maintain **parallel shared prefixes**, one per model family  
  - `Claude-Prefix` – smaller, XML policy tags  
  - `Gemini-Prefix` – bigger, can include raw video / voice  
  - `Fast-Grok-Prefix` – ultra-trimmed, code-edit focus  
- Router chooses model per user ask → still gets 90 % discount
- Removes need to shoehorn one compaction for all
- o3: Worth measuring latency of keeping 2-3 prefixes warm in memory?

- Additional claims from transcript:  
  - Sub-agents are one instance, but we can maintain several shared prefixes simultaneously without throwing away the main one.  
    - > "We are not gonna throw away the main prefix. We're gonna use the model with temporarily a different prefix, but then we're gonna go back to the good shared built prefix that we all know and love."
    - > "I claim that you can simply have several shared prompt prefixes going on at the same time. I don't see why you have to limit yourself."
  - Decouples eval: changing Gemini's compaction doesn't perturb Claude's.  
    - > "Shared prefixes decouple the eval. You don't have to think about other models when doing eval—less spooky effect at a distance. That's good."
    - > "And balancing things is hard. Balancing things, it—like, the eval of balance changes. It's hard to see. Like, I just think it's hard, okay? I just think it's hard."
  - Different families prefer different policy encodings (e.g., Claude's XML-important vs other encodings).  
    - > "Different families of models have different guidelines on how policies should be encoded. Like, maybe if something is super important, Claude has been trained to interpret the XML tags—like the policy XML tag—as its equivalent of a super important marker. And Gemini maybe has some different way of expressing the same thing."
  - Enables easy A/B and side-by-side answers without rebuilding context each time.  
    - > "Now that you have two parallel prefixes going on, what you can do is you can A/B your model decision process for the ask. You can say, 'Okay, here's an ask,' and we are not sure which model is better for this action. So we're going to show the user two responses, one from Claude, one from Gemini, and let the user pick."
  - Users often mix tasks in one conversation (docs tangent, comment sweeps) → having multiple active prefixes avoids "context pollution" while still getting discounts.  
    - > "In the same conversation I can say—and users, they love saying probably everything in the same conversation ... Like, context pollution—they again, they don't probably, I don't think they have an intuitive feel for context pollution."
    - Example: working on build issues, then "Oh, you know, comments—we need to add comments everywhere and write documentation help pages" — different task, shouldn't dominate the context.
  - Assumption (?): discount applies when a new message shares a prefix; we can exploit this per model concurrently.
    - > "When we send a prompt that shares a prefix with one of the previous prompts, we get a 90% discount on that prefix."

---

## 6 · Multimodal Inputs

- Voice dictation, screen recordings, images
  - Views decide:  
    - Gemini gets full video in the prompt  
    - Claude gets 4 keyframes + auto-summary  
- Avoids fragile on-the-fly tool-call orchestration

- Real-world flow to keep:  
  - I may attach a voice recording or a screen walkthrough of me clicking around the generated UI and narrating dislikes.  
    - > "Nowadays, most of my instructions to language models are via voice. I'm dictating all of this. I'm not typing it."
    - > "I can literally record my screen as I'm going with the—I'm pressing the buttons in the generated interface that Lovable created. And I want to just attach my video of literally me going around pressing the buttons and commenting on what I dislike about my app and want Lovable to change, right?"
  - Only Gemini might handle raw video cheaply; others need distilled artefacts.  
    - > "Gemini can handle video, I think, cheaply. You can give it the whole video. And other models, I think, you cannot. You have to extract the salient screenshots or frames from the video."
  - If we pre-condense to a universal summary, future asks run "in the dark."  
    - > "It's tricky. It's tricky to know what's going to be relevant later, what the user actually wants to do, what's important in the video for later asks from the user."
    - Gemini would have to make a summary without knowing future context, then all models work from that potentially-wrong summary.
  - Multi-prefix solves this: video sits in Gemini-prefix; summary + keyframes in Claude-prefix; maybe only a terse summary in expensive models (O3).
    - > "If you have a shared prefix for Gemini and for Claude and they're different, and you have, let's say, the whole video in the shared prefix for Gemini, you have the summary of the video or whatever in the shared prefix for Claude, or maybe you have a couple of screenshots."
    - Voice has the same issue: multimodal models handle raw voice better with context, but we're forced to transcribe upfront today.

---

## 7 · Sub-Agents as First-Class Tools

- Calling a sub-agent = LLM-tool-call with its own prefix  
- Persist last sub-agent prompt → next call reuses as prefix (90 % cheaper)  
- Patterns:  
  1. **Iterative linter** – remembers prior fixes, only diffs new errors  
  2. **Doc generator** – tracks which files already documented  
- Implementation: treat each sub-agent view as another derived context

- Clarifications in your words:  
  - A sub-agent is essentially a tool call whose tool is an LLM, and whose parameter is a prompt surface.  
    - > "A sub-agent, use of sub-agent by the main agent, is a tool call where the tool is LLM, possibly a different LLM, and the parameters are prompt, some prompt surface. Cool stuff. That's the tool call. That's the sub-agent use."
  - After a sub-agent finishes, its final sent prompt becomes the next call's prefix; instant discount and continuity of thought.  
    - > "When you run sub-agent 'find security issues,' and after it's done and after it is dead, you have what's left is its final prompt that was sent, right? That's your prompt prefix for the next call of the sub-agent, which you get a 90% discount on."
    - > "Just use your 90% discount ... You automatically get them, and you automatically get a 10x discount on them if you preserve that previous prompt and say that your next call of the same sub-agent can reuse it."
  - Great for compile-error fixers: the history of compile errors and fixes is already "in the prefix," no bespoke state-passing needed.
    - > "If you have a sub-agent for fixing lints or fixing compile errors, what do you want to give to this sub-agent? You probably want to give it the previous compile errors and how they were fixed, right? And where are they stored? In the previous call of the sub-agent's prompt prefix that was sent."
    - Security agent example: can naturally say "Oh, this got fixed. That's nice. This was pointed out and didn't get fixed. That's not nice." without manual memory passing.

---

## 8 · Why Immutable Wins (detailed)

1. **Strict Superset Power** – any mutable scheme = particular view of immutable log  
2. **Audit / Debug** – replay exact knowledge state at T₀  
3. **Easier Experiments** – toggle new features on old data; no cold-start  
4. **User Trust** – never forced to decide “forever?” in the moment  
5. **Model Flexibility** – compactions per model, per task  
6. **Cheaper A/B** – no waiting for fresh data; instant backfill

- Additional reasoning threads to retain:  
  - Recording "importance numbers" inside a mutable store is just another brittle heuristic; as soon as you want to change what "importance" means, you wish you had the full immutable territory and could derive a better map.  
    - > "You want to record the number, which is importance, which is actually not importance—it's what your maybe stupid model thought was important at some point in time, taking into account something that now maybe you think was stupid to take into account."
    - > "Recording the entire territory, you can always come up with any sort of map out of this."
  - Retroactive signals (e.g., whether a stated preference was later obliterated by actions/reverts) can immediately re-rank preferences when the full log is available; in mutable memory you can only apply to future entries, slowing A/B to a crawl.  
    - > "You suddenly realize it's a good thing to check if a recorded preference was later completely reverted. Like, if a user states a preference, then the agent follows it, and then the user goes on and types a long message saying, 'Shit, no, I want it like this, no, completely something else,' right?"
    - > "In the mutable preference recording model, where you have to decide at each point if something is a preference, you are now—you cannot do it for past preferences, only for future ... It only applies to future recordings of preferences, which means that you have to enable this signal for some users and then see if—and then monitor the situation over some time."
    - Contrast: "With immutable derived preferences, you could just, if you wanted, you could literally have a toggle."
  - Developer ergonomics: add a toggle ("also consider preference reverts") and instantly see the new sponge output against your whole history; dogfooding becomes possible.
    - > "You can even have a debug view with this toggle on that says, 'See what the preference filter sponge gives now,' and let's mark it up ... And now you can see as a developer—well, thankfully, you can see based on your actions in the past if the current sponge gives you valid results or if it gives bonkers results."
    - > "You can dogfood it. You can literally read through a list of what the sponge gives out now, and you can say, 'Well, is it true that I want the agent to be a therapist? Not really, okay. No, no, it's not my preference.'"

---

## 9 · Open Questions & TODOs

- **Storage & Cost**  
  - Compression, pruning policies for huge binaries  
  - Indexing for fast view queries
- **Governance / Privacy**  
  - Who can query full log?  
  - PII redaction pipelines?
- **Evaluation Framework**  
  - Benchmarks for new sponges vs. old heuristics  
  - Metrics: user revert rate, preference satisfaction
- **Multi-prefix Infra**  
  - How many prefixes before latency balloons?  
  - Hot-swap strategy when router picks new model
- **o3: Versioning of Views**  
  - Should views be pure functions at `git-sha` X to guarantee reproducibility?
- **o3: Conflict Resolution**  
  - If two sponges emit contradictory prefs, how do we rank? Tie-breaker = recency? strength? user vote?

