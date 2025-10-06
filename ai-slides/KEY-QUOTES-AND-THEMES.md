# Key Quotes and Themes - Context Engineering

> Memorable phrases and insights from the source material to preserve in the talk

---

## Core Philosophical Statements

### On Context Definition
- "Context is an append-only log ... By definition, it's immutable - everything you know."
- "Context object = everything the system 'knows' at time t"
- "When you learn more, old knowledge remains accessible."

### On Semantics
- **The key insight:** "If the context thing is mutated by the compactor, then suddenly the semantics of the context type are whatever compaction process ends up being approved by the evals team or maximizes the eval given by the evals team. We don't know how to work with this. The semantics exist in a way you cannot reason about."

- "Semantics become: 'stuff we ended up while running this loop'"
- "We're optimizing for scores, not meaning"

### On Recording vs Territory
- "You want to record the number, which is importance, which is actually not importance—it's what your maybe stupid model thought was important at some point in time"
- "Recording the entire territory, you can always come up with any sort of map out of this."

---

## Memorable Metaphors

### The CSS Precedence Problem
> "You end up with a sort of CSS precedence problem where you have a bunch of stuff, and then you want to add another thing, and you think it's important, so you say, you know, you say exclamation mark important. But then it turns out that there's a third thing that's actually even more important you've discovered, and well, there's only one level of exclamation mark important."

**Why it matters:** Captures the cascading priority problem in mutable systems

### The Sponge
> "So sort of squeezing out stuff from the context. It's not like a tool call as in 'give me, use this tool,' as much as it's a context filter or squeeze. Like a sponge—a context sponge."

**Why it matters:** Intuitive metaphor for functional views

---

## User Pain Points (In Their Own Words)

### The Repetition Problem
> "I never want npm. I never want a default ChatGPT style. This style of talking is terrible. I never want to write bash scripts because it's also terrible. I want Python scripts. I never want to leave comments like 'to do later.' I never want to write CSS in inline HTML."

> "If I rant about a thing in three projects, I'm starting to dread starting a new project because I know I'll have to repeat the same things over and over or fight the same battles over and over."

**Use this:** Opens Track 3 - very relatable

### The Commitment Anxiety
> "As a user, do I want to say that when Cursor tells me, 'Oh, here's a memory. Should I remember it? Do I want to remember it?' Am I comfortable telling Cursor that, 'Yeah, my angry rant right now should follow always forever?' Nah, I'm feeling queasy about it."

> "I don't want to commit to actually making this my preference. I don't know how much Cursor is going to stick to this preference. Maybe it's going to treat it like an ironclad rule and it's going to be bad."

**Use this:** Shows why explicit fails

### The ChatGPT Overshoot
> "I sometimes have a conversation and I say, 'Now we're going to pretend you're a therapist,' and then I look at the memories and I see, 'Oh, the user prefers ChatGPT to be a therapist or a Spanish woman or whoever.' And I'm like, 'No, I'm not going to use this again because apparently you're going to remember that you should reply in Norwegian because I wanted to practice Norwegian.'"

**Use this:** Shows why implicit fails

### The Strength Problem
> "Once you have recorded something, you have no way of telling if you really recorded it for very valid reasons or for kind of not very reliable reasons."

> "'I never want to X,' except obviously for cases A, B, C. The user thought it was obvious, but eh."

**Use this:** The core preference capture problem

---

## Technical Insights

### Shared Prefix Economics
> "When we send a prompt that shares a prefix with one of the previous prompts, we get a 90% discount on that prefix."

> "I claim that you can simply have several shared prompt prefixes going on at the same time. I don't see why you have to limit yourself."

**Use this:** The economic foundation of the architecture

### Decoupled Evaluation
> "Shared prefixes decouple the eval. You don't have to think about other models when doing eval—less spooky effect at a distance. That's good."

> "Now you definitely want different contexts for different models. You're doing compaction. Compaction depends on how big the prompt is. Compaction results in different contexts for different models."

**Use this:** Key benefit of multi-prefix

### The Prefix Constraint
> "You cannot be freely reshuffling the prompt. You cannot do the ideal thing, which is for each request to the LLM say that, 'Hey, you know, now the top 15 most important things are here, here, here,' because, well, then there is no shared prefix."

**Use this:** Shows the tradeoff current systems face

---

## Multimodal Insights

### The Voice Reality
> "Nowadays, most of my instructions to language models are via voice. I'm dictating all of this. I'm not typing it."

> "I found that voice suffers from the same lossy, spur-of-the-moment problem. I know that my voice recording—I know that multimodal models can actually handle voice better than purely voice recognition models if it's in context."

### The Video Use Case
> "I can literally record my screen as I'm going with the—I'm pressing the buttons in the generated interface that Lovable created. And I want to just attach my video of literally me going around pressing the buttons and commenting on what I dislike about my app and want Lovable to change, right?"

> "Gemini can handle video, I think, cheaply. You can give it the whole video. And other models, I think, you cannot. You have to extract the salient screenshots or frames from the video."

### The Dark Problem
> "It's tricky. It's tricky to know what's going to be relevant later, what the user actually wants to do, what's important in the video for later asks from the user."

**Use this:** Shows why premature conversion is bad

---

## Sub-Agent Pattern

### The Definition
> "A sub-agent, use of sub-agent by the main agent, is a tool call where the tool is LLM, possibly a different LLM, and the parameters are prompt, some prompt surface. Cool stuff. That's the tool call. That's the sub-agent use."

### The Persistence Trick
> "When you run sub-agent 'find security issues,' and after it's done and after it is dead, you have what's left is its final prompt that was sent, right? That's your prompt prefix for the next call of the sub-agent, which you get a 90% discount on."

> "Just use your 90% discount ... You automatically get them, and you automatically get a 10x discount on them if you preserve that previous prompt and say that your next call of the same sub-agent can reuse it."

### Natural Memory
> "If you have a sub-agent for fixing lints or fixing compile errors, what do you want to give to this sub-agent? You probably want to give it the previous compile errors and how they were fixed, right? And where are they stored? In the previous call of the sub-agent's prompt prefix that was sent."

**Use this:** Shows elegance of the pattern

---

## The Immutable Power Argument

### Strict Superset
> "The model where you remember stuff is—I don't even think you could say it's an optimization, but the model of immutable context forever is strictly more powerful because you could just as well be simply running some sort of salience filter, salience marker."

> "It's literally the same as what you would have gotten if you were adding those memories into a mutable database, right? But so it's strictly as powerful, but it is also more powerful because you can do other stuff instead."

### Retrospective Power
> "You suddenly realize it's a good thing to check if a recorded preference was later completely reverted. Like, if a user states a preference, then the agent follows it, and then the user goes on and types a long message saying, 'Shit, no, I want it like this, no, completely something else,' right?"

> "In the mutable preference recording model, where you have to decide at each point if something is a preference, you are now—you cannot do it for past preferences, only for future ... It only applies to future recordings of preferences, which means that you have to enable this signal for some users and then see if—and then monitor the situation over some time."

**Contrast:** "With immutable derived preferences, you could just, if you wanted, you could literally have a toggle."

### Developer Experience
> "You can even have a debug view with this toggle on that says, 'See what the preference filter sponge gives now,' and let's mark it up ... And now you can see as a developer—well, thankfully, you can see based on your actions in the past if the current sponge gives you valid results or if it gives bonkers results."

> "You can dogfood it. You can literally read through a list of what the sponge gives out now, and you can say, 'Well, is it true that I want the agent to be a therapist? Not really, okay. No, no, it's not my preference.'"

**Use this:** Concrete developer benefit

---

## On Preferences and Rules Files

### The Naive Approach Fails
> "If I try some naive approach like 'oh, you should be more concise,' it might backfire. I might get a super telegraphic style which I also dislike. I've been burned by this enough times that I don't want to just express my preferences in words and then have the agent perverse them somehow."

### Users Aren't Eval Experts
> "Well, it's not a good expectation for the user to be good at evals. The user is definitely not going to be good at evals ... I think there's an art to writing the rules files, and I think it's very hard for people because now the user has to be the eval team."

> "I do that thing sometimes with the sessions where the final sessions, but I know for sure that many of my preferences that I do angry rants about end up not being recorded in code and rule files."

**Use this:** Shows why manual rules files fail

### The Implicit Capture Problem
> "Simultaneously, it overshoots. It takes a thing which is not a preference at all but from a single chat or a single rant. Maybe the user even misunderstood. Like the user says, 'Oh, you should use this component library,' and then eventually it turns out the component library is shit and the user rolls back. But you have somehow decided, for example, to retain the first preference but not the second."

**Use this:** Both over and under capture

---

## Context Pollution and Task Mixing

> "In the same conversation I can say—and users, they love saying probably everything in the same conversation ... Like, context pollution—they again, they don't probably, I don't think they have an intuitive feel for context pollution."

> "Working on build issues, then 'Oh, you know, comments—we need to add comments everywhere and write documentation help pages' — different task, shouldn't dominate the context."

**Use this:** Why multi-prefix helps with mixed tasks

---

## The Compaction Landscape

### Between Models
> "Somewhere between Model A and B is the ChatGPT model, where they have a thing which tries to extract preferences. Cursor has it now as well, right? But it is one-time. They don't try to be declarative or functional about it."

> "They say, 'Oh, look at what happened just now. Try to guess if it's relevant and add it to a database.' So they do mutable context. They have to guess in each moment what's relevant, which is—I think it's worse."

### Different Models Need Different Compactions
> "Different families of models have different guidelines on how policies should be encoded. Like, maybe if something is super important, Claude has been trained to interpret the XML tags—like the policy XML tag—as its equivalent of a super important marker. And Gemini maybe has some different way of expressing the same thing."

**Use this:** Why one-size-fits-all fails

---

## Practical Examples

### Error Prevention
> "Stop generating 300-line hashes or like stop repeating the function name in the comment. Like function says 'generate eval,' comment says 'this function generates eval.' That's a useless comment."

**Use this:** Concrete error-pattern sponge example

### Implicit Preferences
> "Often, for example, I say that you should rewrite this or that, or let's say I revert and do something. When I revert, it's probably because I didn't like something. But it would be tricky for me to explain what I liked."

**Use this:** Shows why implicit signals matter

---

## Humble Framing

### The Uncertainty
> "I'm new to the team and might be misinterpreting parts of the proposal (?). Please read with that caveat."

**Use this tone:** Throughout the talk

### The Exploration Stance
> "If I had to come up with a design from scratch, then like those are the thoughts that occur to me."

> "Those common practices seem questionable and here's why."

**Use this framing:** Sets expectation correctly

---

## Model Comparison & A/B

### Easy Side-by-Side
> "Now that you have two parallel prefixes going on, what you can do is you can A/B your model decision process for the ask. You can say, 'Okay, here's an ask,' and we are not sure which model is better for this action. So we're going to show the user two responses, one from Claude, one from Gemini, and let the user pick."

### No Rebuild Cost
> "Can A/B your model decision process for the ask"
> "No waiting for fresh data; instant backfill"

**Use this:** Concrete benefit of multi-prefix

---

## Background Agents

> "If we're talking about background agents, then this very naturally maps to background agents because the background agent still wants to have probably a prefix ... You still want to have a 90% discount here, right? Probably. Surely."

> "We are not gonna throw away the main prefix. We're gonna use the model with temporarily a different prefix, but then we're gonna go back to the good shared built prefix that we all know and love."

**Use this:** Shows pattern extends naturally

---

## Balancing is Hard

> "And balancing things is hard. Balancing things, it—like, the eval of balance changes. It's hard to see. Like, I just think it's hard, okay? I just think it's hard."

> "Simultaneously they cannot give too much weight to this because there might be bullshit preferences there. And if they give little weight to this, then things that are actually very strong preferences get lost."

**Use this:** Shows empathy for difficulty of current approaches

---

## Questions to Preserve

### For Track 1:
- "What do we mean by 'context'?"
- "Should we ever delete anything?"

### For Track 2:
- "Can we reason about mutated context?"
- "How did this file look before?"
- "When did the user do the revert?"

### For Track 3:
- "Should I remember this angry rant forever?"
- "How do we distinguish strong from weak preferences?"

### For Track 4:
- "Why must all models share the same compaction?"
- "How many prefixes before overhead hits?"

### For Track 8:
- "What would prove this approach wrong?"

---

## Tone Examples

### Good
✅ "This seems questionable to me"
✅ "Here's what occurs to me if I think about it"
✅ "I'm uncertain about X"
✅ "Have you considered...?"

### Too Strong
❌ "This is clearly the right approach"
❌ "Current systems are broken"

### Too Weak
❌ "Maybe nothing I'm saying makes sense"
❌ "This is probably all wrong"

---

## Visual Metaphors to Consider

1. **The Compaction Loop** (Track 2)
   - Context → Compaction → Mutated Context → Use
   - Shows the circular dependency

2. **CSS Precedence Stack** (Track 2)
   - Visual: !important, !!important, !!!important
   - Shows absurdity of priority escalation

3. **Preference Landscape** (Track 3)
   - Cursor vs ChatGPT vs Claude Code
   - Three different failure modes

4. **Parallel Prefixes** (Track 4)
   - Three swim lanes: Claude, Gemini, Fast
   - Router in middle directing traffic

5. **The Sponge Pipeline** (Track 7)
   - Immutable log → Multiple sponges → Derived contexts
   - Shows filtering/extraction metaphor

6. **Mutable vs Immutable** (Track 8)
   - Mutable: X → Y (lost X)
   - Immutable: X + Y + Z... → view₁, view₂, view₃

---

## Call to Action (Closing)

### Not This:
❌ "You should implement this"
❌ "This solves everything"

### But This:
✅ "What experiments would validate/invalidate this?"
✅ "What am I missing?"
✅ "Which challenges are showstoppers?"
✅ "Is this solving a real problem?"