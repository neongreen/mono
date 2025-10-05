# Design Options for User Decision

Based on the feedback, here are concrete options to choose from.

## Option 1: Tool Name

### A. `want`
```bash
want jujutsu
want github.com/user/repo
want setup-jj-here
```
**Pros:** Natural, conversational ("I want X")  
**Cons:** Might sound too casual

### B. `need`
```bash
need jujutsu
need github.com/user/repo
need setup-jj-here
```
**Pros:** More imperative, clearer intent  
**Cons:** Less friendly

### C. `get`
```bash
get jujutsu
get github.com/user/repo
get setup-jj-here
```
**Pros:** Simple, direct, familiar (like `apt-get`)  
**Cons:** Common word, might conflict with git/other commands

### D. `ensure`
```bash
ensure jujutsu
ensure github.com/user/repo
ensure setup-jj-here
```
**Pros:** Clear about idempotency  
**Cons:** Longer, more formal

### E. Other suggestions?

## Option 2: MVP Scope

### A. Minimal (2-3 weeks)
**Focus:** Just get packages working well
- `want <package>` - Install via Homebrew/mise
- Interactive provider selection
- Basic preferences
- `want list`, `want check`

**Example:**
```bash
want jujutsu
want ripgrep
want list
```

### B. Medium (4-5 weeks)
**Add:** Git repositories
- Everything from Minimal
- `want <github-url>` - Clone repositories
- Smart "don't reset branch" logic
- Repository location prompts

**Example:**
```bash
want jujutsu
want github.com/user/repo
want list
```

### C. Ambitious (6-8 weeks)
**Add:** Advanced use cases
- Everything from Medium
- `want binary <url>` - Download binaries
- `want setup-X-here` - Setup tools in current directory
- Multiple provider paths with preferences

**Example:**
```bash
want jujutsu
want github.com/user/repo
want binary https://example.com/tool
want setup-jj-here
```

## Option 3: Preference System

### A. Explicit (simpler to implement)
User sets preferences manually:
```bash
want preferences set jujutsu mise
want preferences set default mise,homebrew,github
```

### B. Learning (smarter, more complex)
Tool learns from choices:
```bash
want jujutsu
# Shows options: 1. mise, 2. homebrew
# You pick: 1
# Tool remembers: "User prefers mise for packages"
```

Next time:
```bash
want ripgrep
# Automatically tries mise first (learned preference)
```

### C. Hybrid
Learns from choices, but you can override:
```bash
# Tool learns automatically
want jujutsu
# Pick mise -> remembered

# But you can be explicit
want ripgrep --via homebrew

# Or edit preferences
want preferences
```

## Option 4: Temporary vs Persistent

### A. Everything is Persistent
Every `want` command is recorded:
```bash
want temp-tool
# Recorded in ~/.config/want/wants.yaml forever
```

**Simple, but clutters the wants list**

### B. Explicit Flag for Temporary
```bash
want temp-tool          # Recorded
want --temp temp-tool   # Not recorded
```

**Clear, but verbose for temporary stuff**

### C. Smart Detection
```bash
want jujutsu            # Recorded (common tool)
want setup-jj-here      # Not recorded (one-time action)
want binary https://... # Recorded (specific requirement)
```

**Smart, but might surprise users**

### D. Always Ask
```bash
want temp-tool
# Installed successfully
# Record this requirement? [Y/n]:
```

**Most flexible, but interrupts flow**

## Option 5: Multiple Provider Handling

### A. Always Interactive
Every time there are multiple options, prompt:
```bash
want jujutsu
# 1. mise
# 2. homebrew
# 3. github-release
# Which? [1]:
```

**Pros:** User always in control  
**Cons:** Interrupts flow even when obvious

### B. Use Preferences
If preference exists, use it. Otherwise, prompt:
```bash
# First time
want jujutsu
# [Interactive prompt] -> Choose mise

# Next time
want ripgrep
# Automatically tries mise (from preferences)
# If not available via mise, then prompts
```

**Pros:** Smart, learns over time  
**Cons:** More complex logic

### C. Default with Override
Always use a default order (mise → homebrew → github), but allow override:
```bash
want jujutsu           # Uses mise (default first choice)
want jujutsu --via homebrew  # Forces homebrew
```

**Pros:** No interruptions for common cases  
**Cons:** Less transparent

## Option 6: Repository Handling

### A. Just Clone (simplest)
```bash
want github.com/user/repo
# Clones to ~/projects/repo
# That's it - no branch management, no pulls
```

### B. Clone + Location Choice
```bash
want github.com/user/repo
# Where to clone?
# 1. ~/projects/repo
# 2. ~/code/repo
# 3. Custom
```

### C. Clone + Smart Updates
```bash
want github.com/user/repo
# If exists: verify only, don't touch
# If not: clone and ask about location
# Later: `want check` can offer to pull updates
```

## Recommendation Summary

Based on the feedback, I recommend:

| Aspect | Recommendation | Why |
|--------|---------------|-----|
| **Name** | `want` | Most natural: "I want X" |
| **MVP Scope** | Medium (4-5 weeks) | Packages + git repos covers main use cases |
| **Preferences** | Hybrid | Learn from choices, allow overrides |
| **Temporary** | Explicit flag | Clear: `--temp` when needed |
| **Multiple providers** | Use preferences | Smart default, no interruptions |
| **Repositories** | Clone + location choice | Balance of simplicity and flexibility |

## Quick Start Examples (Recommended Approach)

```bash
# Install a tool
want jujutsu
# First time: Interactive prompt to choose provider
# Remembers your choice

# Get a repository
want github.com/neongreen/mono
# Prompts for location if not exists
# Doesn't touch if already cloned

# Temporary action
want --temp setup-jj-here
# Does the setup, doesn't record permanently

# Check what you have
want list
want check

# Manage preferences
want preferences
```

## Questions for User

1. **Name:** Which name feels right? `want`, `need`, `get`, or something else?

2. **MVP:** Should we start with just packages (Minimal), or include git repos (Medium)?

3. **Interactivity:** How much prompting is acceptable? Always interactive, or learn preferences?

4. **Temporary wants:** Should temporary things require `--temp` flag, or detect automatically?

5. **Other use cases:** Which specific use cases are highest priority?
   - Binary downloads from URLs
   - MCP server setup
   - Folder restoration from backups
   - Setup commands (like `setup-jj-here`)

---

**Ready for your input on these choices!**
