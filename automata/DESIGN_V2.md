# Want/Need - Interactive Task Fulfillment Tool (Design V2)

> **Based on feedback:** This design reflects a fundamental shift from declarative state management to an interactive, task-oriented tool.

## Core Concept

**Want/Need** is an interactive tool that helps you get things you need on your system. It's more like an intelligent assistant than a state manager.

### Key Principle: Task Fulfillment, Not State Management

❌ **Not this:** "My system must always be in this exact state"  
✅ **This:** "I need X right now, make it available to me"

## Primary Interaction Model

### CLI-First, Not YAML-First

Users interact primarily through commands, like `pnpm add` or `mise use`:

```bash
# Get things you need
want jujutsu
want github.com/user/repo
want binary https://example.com/tool

# The tool records these in a config file, but editing YAML is not the primary workflow
```

### Interactive Decision Making

When there are multiple ways to satisfy a requirement:

```bash
$ want jujutsu

Multiple options available:
  1. Install via mise (recommended)
  2. Install via Homebrew
  3. Download from GitHub releases
  4. Build from source (requires Rust)

Which would you prefer? [1]: 
```

The tool remembers your preferences for next time.

## Core Philosophy

### 1. Under-Constrain, Don't Over-Constrain

**Bad:** "Install jujutsu 0.23.0 from Homebrew specifically"  
**Good:** "Make jujutsu available" (however works best)

When you say `want jujutsu`, you mean:
- ✅ I want to be able to run `jj` command
- ❌ NOT: Install this specific version from this specific source

### 2. Make Things Available, Don't Reset State

**Bad:** Check out repo and force to main branch  
**Good:** Ensure repo is cloned, but don't touch existing work

Example:
```bash
$ want github.com/user/myproject

Checking github.com/user/myproject...
✓ Repository already cloned at ~/projects/myproject
✓ Currently on branch: feature-work
  (Not changing branch - you're already working here)
```

### 3. Temporary vs Persistent Wants

Not everything needs to be recorded permanently:

```bash
# Temporary - just do it now
want --temp setup-jujutsu-in ~/projects/myrepo

# Persistent - record as a system requirement
want --persist jujutsu
```

### 4. Respect User Preferences

The tool learns your preferences:

```yaml
# ~/.config/want/preferences.yaml (generated, rarely edited manually)
preferences:
  package_sources:
    default: [mise, homebrew, github-release, build-from-source]
  jujutsu:
    prefer: mise
  ocr:
    tool: tesseract
    prefer: homebrew
```

## Use Cases

### Use Case 1: Need a Tool Available

```bash
$ want jujutsu

Analyzing requirement: jujutsu
Found options:
  • Via mise: jujutsu 0.23.0
  • Via Homebrew: jujutsu 0.22.0
  • Via GitHub releases: jujutsu 0.23.0
  • Build from source: latest

Recommendation: mise (matches your preference)
Install via mise? [Y/n]: y

Installing jujutsu via mise...
✓ jujutsu available at /Users/you/.local/share/mise/installs/jujutsu/0.23.0/bin/jj
✓ Added to PATH

Recorded preference: jujutsu via mise
```

### Use Case 2: Need a Repository

```bash
$ want github.com/neongreen/mono

Checking github.com/neongreen/mono...
  Repository not found locally

Where should I clone it?
  1. ~/projects/mono (default)
  2. ~/code/mono
  3. Custom path

Choice [1]: 

Cloning github.com/neongreen/mono to ~/projects/mono...
✓ Repository cloned
✓ Currently on: main branch

Setup complete. Repository available at ~/projects/mono
```

### Use Case 3: Need Setup in a Repo

```bash
$ cd ~/projects/myrepo
$ want jujutsu-setup-here

This will set up Jujutsu in the current repository.

Prerequisites:
  • jujutsu (available ✓)

Actions:
  1. Initialize Jujutsu: jj git init --colocate
  2. Configure git push behavior

Proceed? [Y/n]: y

✓ Jujutsu initialized in ~/projects/myrepo
✓ Configured git push
```

### Use Case 4: Need a Binary from URL

```bash
$ want binary https://github.com/user/tool/releases/download/v1.0/tool-macos

Downloading tool from https://github.com/user/tool/releases/download/v1.0/tool-macos...
✓ Downloaded to ~/.local/bin/tool
✓ Made executable
✓ Added to PATH

tool is now available. Try: tool --help
```

### Use Case 5: Need MCP Server Available

```bash
$ want mcp-server @modelcontextprotocol/server-filesystem

Analyzing requirement...

To install MCP server @modelcontextprotocol/server-filesystem:
  • Need: npx (via node)
  • Will: Update Claude config at ~/Library/Application Support/Claude/claude_desktop_config.json

Prerequisites check:
  • node: available via Homebrew ✓
  • npx: available ✓

Add MCP server to Claude config? [Y/n]: y

✓ Added @modelcontextprotocol/server-filesystem to Claude config
✓ Server will be available after restarting Claude

MCP server configured successfully.
```

### Use Case 6: Multiple Paths with Preferences

```bash
$ want some-rare-tool

Multiple installation paths available:

Path 1: Homebrew
  • Status: Homebrew has old version (1.0.0)
  • Pros: Easy updates
  • Cons: Outdated

Path 2: GitHub releases
  • Status: Latest version available (2.5.0)
  • Pros: Latest version
  • Cons: Manual updates

Path 3: Build from source
  • Status: Requires Rust toolchain
  • Pros: Always latest, optimized for your system
  • Cons: Slow, requires additional tools

Recommended: GitHub releases (latest version, simple)

Which path? [2]: 2

Downloading from GitHub releases...
✓ some-rare-tool 2.5.0 installed

Remember this choice for next time? [Y/n]: y
✓ Preference saved: some-rare-tool via github-release
```

## Architecture

### Components

```
want
├── cli/              # Command-line interface
│   ├── commands.go   # want, list, forget, preferences
│   └── ui.go         # Interactive prompts
├── resolver/         # Requirement resolution
│   ├── resolver.go   # Find ways to satisfy requirements
│   ├── options.go    # Evaluate different paths
│   └── preferences.go # User preference handling
├── providers/        # Different ways to get things
│   ├── mise.go
│   ├── homebrew.go
│   ├── github.go
│   ├── build.go
│   └── download.go
├── checker/          # Check if requirements are satisfied
│   └── checker.go
└── config/           # Configuration storage
    ├── wants.go      # What user has asked for
    └── preferences.go # User preferences
```

### State Files

```yaml
# ~/.config/want/wants.yaml - What you've asked for
wants:
  - name: jujutsu
    satisfied_by: mise
    installed_at: 2024-01-15T10:30:00Z
    
  - name: github.com/user/repo
    satisfied_by: git-clone
    path: ~/projects/repo
    installed_at: 2024-01-15T11:00:00Z

# ~/.config/want/preferences.yaml - Your preferences
preferences:
  package_sources:
    priority: [mise, homebrew, github-release]
  
  specific:
    jujutsu:
      prefer: mise
    tesseract:
      prefer: homebrew
```

## Commands

### `want <requirement>`

Primary command to get something you need.

```bash
want jujutsu                    # Get a tool
want github.com/user/repo       # Get a repository
want binary https://...         # Get a binary from URL
want mcp-server @scope/name     # Setup MCP server
want folder-from-backup ~/data  # Restore from backup
```

Options:
- `--temp` - Don't record this permanently
- `--path <path>` - Specific installation path
- `--via <provider>` - Force specific provider

### `want list`

Show what you've asked for and their status.

```bash
$ want list

Your requirements:
  ✓ jujutsu (via mise, v0.23.0)
  ✓ github.com/user/repo (at ~/projects/repo)
  ✗ old-tool (no longer available via homebrew)
  ⚠ binary-tool (outdated, v1.0 available)
```

### `want check [requirement]`

Check if requirements are satisfied.

```bash
$ want check

Checking all requirements...
  ✓ jujutsu: available at ~/.local/share/mise/bin/jj
  ✓ github.com/user/repo: cloned at ~/projects/repo
  ✗ removed-tool: not found

1 requirement needs attention.
Run 'want removed-tool' to reinstall.
```

### `want forget <requirement>`

Remove a recorded requirement (doesn't uninstall).

```bash
$ want forget old-tool

Removed 'old-tool' from your wants.
(Tool is still installed, just not tracked)
```

### `want preferences`

Manage preferences interactively.

```bash
$ want preferences

Current preferences:
  1. Package source priority: mise, homebrew, github-release
  2. jujutsu: prefer mise
  3. tesseract: prefer homebrew

Options:
  [e] Edit priority
  [r] Remove specific preference
  [s] Show all
  [q] Quit
```

## Design Decisions

### 1. CLI Over YAML

**Primary workflow:**
```bash
want jujutsu
want github.com/user/repo
```

**Not:**
```yaml
# edit wants.yaml manually
wants:
  - jujutsu
  - github.com/user/repo
```

The YAML exists, but it's **generated** by the CLI commands.

### 2. Interactive Over Declarative

When multiple options exist, prompt the user:
```
Which installation method?
  1. mise
  2. Homebrew
  3. GitHub release
```

Don't force them to specify in advance.

### 3. Preferences Over Strict Configuration

Instead of:
```yaml
jujutsu:
  provider: mise
  version: "0.23.0"
  path: /exact/path
```

Just:
```yaml
jujutsu:
  prefer: mise
```

Let the tool figure out the rest.

### 4. Under-Constrain by Default

When you say `want X`:
- ✅ Care: X is available
- ❌ Don't care: Exact version, exact path, exact method

Unless you explicitly say otherwise:
```bash
want jujutsu --via homebrew --version 0.22.0
```

### 5. Don't Touch Existing Work

If a repository is already cloned:
- ✅ Verify it's there
- ❌ Don't checkout a specific branch
- ❌ Don't pull changes
- ❌ Don't reset state

The user is probably working in it!

## Comparison with Original Design

| Aspect | Original (Automata) | New (Want/Need) |
|--------|---------------------|-----------------|
| **Philosophy** | State management | Task fulfillment |
| **Primary UI** | Edit YAML | CLI commands |
| **Approach** | Declarative | Interactive |
| **Constraints** | Specify exactly | Under-specify |
| **Multi-machine** | Sync state | Not a goal |
| **Existing work** | Reset to spec | Preserve |
| **Name** | automata | want/need |
| **Like** | Ansible | pnpm/mise |

## Implementation Priorities

### Phase 1: Core (2-3 weeks)
- [ ] `want` command for packages via Homebrew
- [ ] `want` command for git repositories
- [ ] Basic preferences system
- [ ] `want list` and `want check`

### Phase 2: Smart Resolution (2-3 weeks)
- [ ] Multiple provider support (mise, Homebrew, GitHub)
- [ ] Interactive option selection
- [ ] Preference learning

### Phase 3: Advanced Use Cases (3-4 weeks)
- [ ] Binary downloads from URLs
- [ ] MCP server setup
- [ ] Build-from-source option
- [ ] Folder restoration from backups

### Phase 4: Polish (1-2 weeks)
- [ ] Error handling and recovery
- [ ] Help text and documentation
- [ ] Testing
- [ ] Real-world usage refinement

## Open Questions

1. **Name:** `want` or `need`? Or something else?
   - `want jujutsu` - sounds natural
   - `need jujutsu` - more imperative
   - `get jujutsu` - simple, clear

2. **Config location:** `~/.config/want/` or `~/.want/`?

3. **Uninstall behavior:** Should `want forget` also uninstall, or just stop tracking?

4. **Update behavior:** Should `want check` offer to update outdated tools?

5. **Backup integration:** How to handle "want folder-from-backup"? UI for selecting backup dates?

## Next Steps

1. Validate this revised design with user feedback
2. Choose the tool name (want/need/get)
3. Define MVP scope precisely
4. Create implementation plan
5. Start with Phase 1

---

**This is a fundamentally different tool from the original design.** It's an interactive assistant for getting things you need, not a declarative state manager.
