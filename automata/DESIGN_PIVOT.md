# Design Pivot: From State Management to Task Fulfillment

This document explains the fundamental shift in design based on user feedback.

## The Shift

### Original Vision (DESIGN.md)
**Declarative State Manager** - Like Ansible for personal machines
- Edit YAML files to describe system state
- Tool brings system to match YAML
- Idempotent, repeatable
- Focus: Identical state across machines

### New Vision (DESIGN_V2.md)
**Interactive Task Fulfillment** - Like pnpm/mise for everything
- Run commands to get what you need
- Tool figures out how to satisfy requirements
- Interactive, flexible
- Focus: Get things done right now

## Key Differences

### 1. Primary Interaction

**Before:**
```yaml
# Edit config.yaml
resources:
  - type: package
    name: jujutsu
    provider: homebrew
```
```bash
automata apply config.yaml
```

**After:**
```bash
want jujutsu

# Interactive:
# Multiple options available:
#   1. Install via mise
#   2. Install via Homebrew
# Which would you prefer? [1]:
```

### 2. Philosophy

**Before:** "Describe the complete state you want"
- System state must match config
- Everything is specified
- Force system to match

**After:** "Say what you need, we'll figure it out"
- Under-specify by default
- Multiple ways to satisfy
- Don't force changes to existing work

### 3. Constraints

**Before:** Over-constrain
```yaml
- type: git_repo
  path: ~/projects/myapp
  remote: https://github.com/user/myapp.git
  branch: main  # Forces checkout of main
```

**After:** Under-constrain
```bash
want github.com/user/myapp
# Just ensures it's cloned
# Doesn't touch your current branch if already there
```

### 4. Use Cases

**Before:**
- New machine setup (clone everything at once)
- Maintain identical environments
- Team synchronization
- Comprehensive system configuration

**After:**
- Need a tool right now → `want tool`
- Need a repo right now → `want repo`
- Need something set up → `want setup-x`
- Get things done quickly

### 5. State Management

**Before:**
- Config file is the source of truth
- System must match config exactly
- Changes: Edit YAML, then apply

**After:**
- CLI commands are primary
- Config is generated from commands
- Changes: Run new `want` command

### 6. Repository Handling

**Before:**
```yaml
git_repo:
  path: ~/projects/myapp
  branch: main
  remote: origin
```
→ If you're on `feature-branch`, tool resets to `main`

**After:**
```bash
want github.com/user/myapp
```
→ If already cloned on `feature-branch`, tool says "Already there" and doesn't touch it

### 7. Multiple Paths

**Before:**
- Choose one provider in config
- No alternatives presented
- Rigid

**After:**
- Tool shows all options
- User picks interactively
- Preferences remembered

### 8. Naming

**Before:** `automata` (suggests automation/state machines)

**After:** `want` or `need` (suggests getting what you need)

## Why This Changes Everything

### Original Design Problems (based on feedback)

1. **Too verbose** - Specifying every detail is tedious
2. **Too rigid** - Forces specific providers/versions/paths
3. **Too destructive** - Resets things that shouldn't be reset
4. **Wrong workflow** - YAML editing instead of CLI commands
5. **Wrong goal** - State management instead of task fulfillment

### New Design Benefits

1. **Natural workflow** - `want X` is how you think
2. **Flexible** - Multiple ways to satisfy requirements
3. **Safe** - Doesn't touch existing work
4. **Interactive** - Presents options when needed
5. **Learns** - Remembers your preferences

## What Stays the Same

- Written in Go
- macOS focus (initially)
- Can install prerequisites automatically
- Tracks what you've asked for
- Smart about dependencies

## What Changes Completely

| Aspect | Before | After |
|--------|--------|-------|
| **Core concept** | State manager | Task fulfiller |
| **User flow** | Edit → Apply | Want → Get |
| **Config role** | Source of truth | Record of wants |
| **When to use** | Setup new machine | Get things as needed |
| **Philosophy** | Over-specify | Under-specify |
| **Interaction** | Batch (apply all) | Interactive (one want) |

## Examples Side-by-Side

### Example 1: Get Jujutsu

**Before (automata):**
```yaml
# Edit config.yaml
resources:
  - type: package
    name: jujutsu
    provider: homebrew
```
```bash
automata apply config.yaml
```

**After (want):**
```bash
want jujutsu
# Interactive prompt if multiple options
# Tool remembers your choice
```

### Example 2: Get a Repository

**Before (automata):**
```yaml
# Edit config.yaml
resources:
  - type: git_repo
    path: ~/projects/myapp
    remote: https://github.com/user/myapp.git
    branch: main
```
```bash
automata apply config.yaml
# If repo exists on feature-branch, gets reset to main!
```

**After (want):**
```bash
want github.com/user/myapp
# If repo exists, just verifies and doesn't touch branch
# If not, clones it
```

### Example 3: Need Multiple Things

**Before (automata):**
```yaml
# Edit config.yaml
resources:
  - type: package
    name: git
  - type: package
    name: jq
  - type: package
    name: ripgrep
```
```bash
automata apply config.yaml
# Installs all at once
```

**After (want):**
```bash
want git
want jq
want ripgrep
# Or: want git jq ripgrep
# Interactive for each if needed
# Can do them as you need them, not all at once
```

## Implementation Impact

### What Can Be Reused

From the original design:
- Provider system concept (Homebrew, mise, etc.)
- Dependency resolution
- Checking if things are available
- Go as implementation language

### What Needs Complete Rework

- CLI interface (commands vs config files)
- User interaction model (interactive vs batch)
- State representation (wants vs resources)
- Philosophy of operation

### New Code Needed

- Interactive prompts and option selection
- Preference learning system
- Smart under-constraining logic
- "Don't touch existing work" guards

## Migration Path

The original design documents remain useful for:
- Understanding package providers
- Resource type concepts
- Technical architecture ideas

But the user-facing design is fundamentally different.

## Decision Points

Based on this pivot, we need to decide:

1. **Name:** `want`, `need`, `get`, or something else?
2. **Scope:** Start with just packages, or also git repos?
3. **MVP:** How minimal can we make the first version?
4. **Preferences:** How much learning vs manual configuration?

## Conclusion

This isn't a minor design change - it's a complete pivot from:
- **"Ansible for personal machines"** (declarative state management)

To:
- **"pnpm/mise for everything"** (interactive task fulfillment)

The technical implementation may share some concepts, but the user experience and philosophy are entirely different.

---

**Next:** Review DESIGN_V2.md and validate this new direction before implementation.
