# Automata Design Summary

**Status:** 📋 Design Phase - Ready for Discussion

This is a quick reference guide to the design documentation.

## What is Automata?

A declarative system automation tool for personal machine setup and programming tasks - think "Ansible for your laptop" with interactive secret management.

```yaml
# Example: One command to set up a new machine
resources:
  - type: package
    name: git
  - type: git_repo
    path: ~/projects/myapp
  - type: secret
    name: GITHUB_TOKEN
    url: https://github.com/settings/tokens
```

## Documentation Structure

### 1. 📖 [README.md](README.md)
**Start here** - Overview, motivation, and current status
- What is automata?
- Use cases
- Why another tool?
- Example configuration

### 2. 🎨 [DESIGN.md](DESIGN.md)
**Core concepts** - Resource types, execution model, and configuration format
- Declarative state management
- Resource types (git, package, secret, file)
- Execution phases (parse, plan, apply, verify)
- Key design decisions

### 3. 🏗️ [ARCHITECTURE.md](ARCHITECTURE.md)
**Technical details** - Implementation architecture
- Component diagram
- Resource interface
- Provider system
- File structure
- Development roadmap

### 4. 💬 [DISCUSSION.md](DISCUSSION.md)
**Questions for you** - Specific decisions to make
- Scope for MVP
- Resource priorities
- Configuration format
- State management
- Platform support
- 10 key questions to answer

### 5. ⚖️ [COMPARISON.md](COMPARISON.md)
**vs Similar Tools** - How automata differs
- vs Ansible (lighter, personal focus)
- vs Homebrew Bundle (broader scope)
- vs Nix (pragmatic vs pure)
- vs Shell Scripts (declarative vs imperative)
- Feature matrix

### 6. 📝 [examples/](examples/)
**Sample configs** - Concrete use cases
- `simple-git.yaml` - Basic git setup
- `personal-machine.yaml` - Full machine setup
- `project-setup.yaml` - Project bootstrap
- `secrets-management.yaml` - Secret handling

## Quick Navigation

**Want to understand the concept?** → [README.md](README.md)

**Ready to dive into design?** → [DESIGN.md](DESIGN.md)

**Need technical details?** → [ARCHITECTURE.md](ARCHITECTURE.md)

**Have opinions on design?** → [DISCUSSION.md](DISCUSSION.md)

**Compare with other tools?** → [COMPARISON.md](COMPARISON.md)

**See concrete examples?** → [examples/](examples/)

## Key Features

### ✅ Declarative
Describe what you want, not how to achieve it:
```yaml
- type: git_repo
  path: ~/projects/myapp
```
Not: "check if directory exists, then check if .git exists, then git init..."

### ✅ Idempotent
Safe to run multiple times - only changes what's needed:
```bash
automata apply config.yaml  # First run: creates repo
automata apply config.yaml  # Second run: does nothing
```

### ✅ Smart Dependencies
Auto-installs prerequisites:
```yaml
- type: git_repo  # Needs git
  path: ~/projects/myapp
```
Result: "Git not found. Install Homebrew first? [y/n]"

### ✅ Interactive Secrets
Human-in-the-loop for API keys:
```yaml
- type: secret
  name: GITHUB_TOKEN
  url: https://github.com/settings/tokens
```
Result: Shows URL, prompts for token, stores securely

## Design Highlights

### Resource Types

| Type | Purpose | Example |
|------|---------|---------|
| `git_repo` | Manage git repositories | Init, set remote |
| `package` | Install software | via brew, apt, go, npm |
| `secret` | Interactive secret mgmt | API keys with prompts |
| `file` | File/directory operations | Create, copy, permissions |

### Execution Flow

```
1. Parse YAML
2. Validate config
3. Resolve dependencies
4. Check current state
5. Generate plan ───→ [Show to user]
6. Apply changes
7. Verify results
```

### Provider System

Multiple package manager support:
- **Homebrew** (macOS)
- **apt** (Ubuntu/Debian)
- **go** (Go toolchain)
- **npm** (Node packages)
- **pip** (Python packages)

Auto-installs providers if missing (with permission).

## Example Workflow

### Setting up a new machine:

```bash
# 1. Install automata
go install github.com/neongreen/mono/automata@latest

# 2. Create config
cat > machine.yaml << 'EOF'
resources:
  - type: package
    name: git
  - type: git_repo
    path: ~/projects/myapp
    remote: https://github.com/user/myapp.git
  - type: secret
    name: GITHUB_TOKEN
    store: env
    url: https://github.com/settings/tokens
EOF

# 3. Apply configuration
automata apply machine.yaml

# Output:
# ✓ git: already installed
# ✓ ~/projects/myapp: initialized git repository
# ✓ ~/projects/myapp: set remote 'origin'
# ⚠ GITHUB_TOKEN: not found
#   Get your token from: https://github.com/settings/tokens
#   Enter token: ***************
# ✓ GITHUB_TOKEN: stored in ~/.zshrc
```

## Open Questions

Key decisions to make (see [DISCUSSION.md](DISCUSSION.md) for details):

1. **MVP Scope:** File + git + package, or start simpler?
2. **Package Providers:** Homebrew only, or multi-platform?
3. **Dependency Resolution:** Manual ordering or automatic?
4. **State Management:** Stateless or track changes?
5. **Interactive Features:** Essential or can wait?
6. **Platform Support:** macOS only or cross-platform?
7. **Testing:** Unit tests only or integration tests?
8. **Naming:** Keep "automata" or choose different name?

## Next Steps

1. **Review** the design documents
2. **Provide feedback** on [DISCUSSION.md](DISCUSSION.md) questions
3. **Prioritize** features for MVP
4. **Decide** on key design questions
5. **Approve** or refine the design
6. **Begin implementation** once aligned

## Timeline Estimates

### Option A - Minimal MVP (2 weeks)
- Core framework
- File resource only
- No dependency resolution
- No interactive features

### Option B - Useful MVP (4 weeks)
- Core framework
- File + git + package resources
- Basic dependency resolution
- Homebrew provider only

### Option C - Feature Complete (8 weeks)
- All core resources
- Multiple providers
- Full dependency resolution
- Interactive secrets
- Comprehensive tests

## Questions?

**Unclear about something?** Ask in the PR discussion!

**Have suggestions?** Comment on [DISCUSSION.md](DISCUSSION.md)

**Want to see more examples?** Check [examples/](examples/)

**Ready to start coding?** Let's finalize the design first!

---

## File Checklist

Design documentation:
- [x] README.md - Overview and motivation
- [x] DESIGN.md - Core design and concepts
- [x] ARCHITECTURE.md - Technical architecture
- [x] DISCUSSION.md - Questions and decisions
- [x] COMPARISON.md - vs similar tools
- [x] SUMMARY.md - This quick reference
- [x] examples/ - Sample configurations

Ready for:
- [ ] Design review
- [ ] Feedback incorporation
- [ ] MVP scope definition
- [ ] Implementation planning
- [ ] Coding!

---

**Let's discuss! 💬**
