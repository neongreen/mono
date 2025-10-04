# Automata: Declarative Personal Machine Automation

> **TL;DR:** Like Ansible for your laptop, but simpler and with interactive secret management.

## The Problem

Setting up a new development machine or maintaining your current one is tedious:

❌ **Shell scripts get messy**
```bash
if ! command -v git; then
  if ! command -v brew; then
    # Install Homebrew first
    /bin/bash -c "$(curl ...)"
  fi
  brew install git
fi
# ...100 more lines of conditionals
```

❌ **Manual steps are error-prone**
- Forgot to install a tool?
- Need to remember where to get API keys?
- Different setup for each machine?

❌ **Existing tools are overkill**
- Ansible: Too complex for personal use
- Nix: Steep learning curve
- Homebrew Bundle: Only handles packages

## The Solution: Automata

**Declarative** + **Idempotent** + **Smart** = ✨

### One Config File

```yaml
# machine.yaml
resources:
  - type: package
    name: git
  
  - type: git_repo
    path: ~/projects/myapp
    remote: https://github.com/user/myapp.git
  
  - type: secret
    name: GITHUB_TOKEN
    url: https://github.com/settings/tokens
```

### One Command

```bash
automata apply machine.yaml
```

### Smart Automation

```
Need git? → Need Homebrew? → "Install Homebrew first? [y/n]"
Missing API key? → Shows URL → Prompts for key → Stores securely
Already set up? → Does nothing (idempotent)
```

## Key Features

### 🎯 Declarative
Say **what** you want, not **how** to do it.

### 🔁 Idempotent
Run it 100 times, same result. Always safe.

### 🧠 Smart Dependencies
Figures out prerequisites automatically.

### 💬 Interactive
Guides you through manual steps (API keys, etc.)

### 📦 Multi-Provider
Homebrew, apt, Go, npm, pip - use any package manager.

### 🔒 Secure
Interactive secret management with prompts and URLs.

## Use Cases

### ✅ New Machine Setup
Clone your dotfiles, run one command, everything is ready.

### ✅ Team Onboarding
Share a config file, new developers get consistent setup.

### ✅ Project Bootstrap
Include a config in your repo for easy contributor setup.

### ✅ Dotfiles Management
More than just config files - packages, repos, secrets too.

### ✅ CI/CD Agents
Declaratively configure build machines.

## Quick Comparison

| Feature | Automata | Ansible | Homebrew | Scripts |
|---------|----------|---------|----------|---------|
| Single binary | ✅ | ❌ | ❌ | ✅ |
| Declarative | ✅ | ✅ | ✅ | ❌ |
| Interactive | ✅ | ❌ | ❌ | Manual |
| Packages | ✅ | ✅ | ✅ | ✅ |
| Git repos | ✅ | ✅ | ❌ | ✅ |
| Secrets | ✅ | Vault | ❌ | Manual |
| Learning curve | Low | Medium | Low | Low |

## Why Build This?

**Ansible** - Too heavy for personal use, designed for servers  
**Homebrew Bundle** - Great but only handles packages  
**Nix** - Perfect reproducibility but steep learning curve  
**Shell Scripts** - Flexible but not declarative or idempotent  

**Gap:** Simple, declarative, personal machine automation with human interaction.

## Example Workflows

### Scenario 1: Git Setup

**Config:**
```yaml
resources:
  - type: package
    name: git
  - type: git_repo
    path: ~/projects/notes
```

**First run:**
```
✓ Installed git via Homebrew
✓ Created ~/projects/notes
✓ Initialized git repository
```

**Second run:**
```
✓ git already installed
✓ ~/projects/notes already set up
No changes needed
```

### Scenario 2: API Keys

**Config:**
```yaml
resources:
  - type: secret
    name: GITHUB_TOKEN
    url: https://github.com/settings/tokens
```

**Run:**
```
⚠ GITHUB_TOKEN not found

To create a token:
1. Visit: https://github.com/settings/tokens
2. Click "Generate new token"
3. Select scopes: repo, workflow

Enter token: ********************

✓ Stored in ~/.zshrc
```

### Scenario 3: Complete Setup

**Config:**
```yaml
resources:
  # Tools
  - type: package
    name: git
  - type: package
    name: go
  - type: package
    name: jq
  
  # Projects
  - type: git_repo
    path: ~/projects/webapp
    remote: https://github.com/user/webapp.git
  
  # Secrets
  - type: secret
    name: GITHUB_TOKEN
    url: https://github.com/settings/tokens
```

**Result:** Entire dev environment in one command.

## Technical Highlights

### Architecture
- **Go-based:** Single binary, fast, cross-platform
- **Resource model:** Everything is a resource with desired state
- **Provider system:** Pluggable package managers
- **Dependency resolution:** Automatic ordering

### Resource Types
1. **package** - Install software via any package manager
2. **git_repo** - Manage git repositories
3. **secret** - Interactive secret management
4. **file** - File and directory operations

### Execution Flow
```
Parse → Validate → Resolve Deps → Check State → Plan → Apply → Verify
```

## Current Status

✅ **Design Phase Complete**
- Comprehensive design docs
- Example configurations
- Architecture specification
- Comparison with alternatives

⏳ **Next Steps**
- Review design with stakeholders
- Gather feedback on priorities
- Finalize MVP scope
- Begin implementation

## MVP Scope

**Phase 1: Core (2-4 weeks)**
- YAML parser
- Resource interface
- Basic executor
- File + git + package resources
- Homebrew provider

**Phase 2: Enhancement (4-6 weeks)**
- Dependency resolution
- Interactive prompts
- Secret resource
- Error handling

**Phase 3: Polish (2-3 weeks)**
- Additional providers (apt, go)
- Testing
- Documentation
- Examples

## Success Metrics

✅ **Developer Experience**
- 10-minute setup for new machine
- Single config file
- No manual steps (except secrets)

✅ **Reliability**
- 100% idempotent
- Clear error messages
- Safe defaults

✅ **Usability**
- Intuitive YAML syntax
- Helpful prompts
- Good documentation

## Questions?

### "Why not just use Ansible?"
Too complex for personal use. Automata is optimized for single-machine, personal development environments with interactive workflows.

### "Why not just use shell scripts?"
Not declarative or idempotent. Shell scripts work but require constant maintenance and careful state checking.

### "Why not Nix?"
Nix is amazing for reproducibility but has a steep learning curve. Automata is pragmatic over perfect.

### "What about Windows?"
MVP focuses on macOS/Linux. Windows support is possible but not priority.

### "Can I use this in production?"
Designed for development machines, not production servers. Use Ansible/Terraform for that.

## Get Involved

**📖 Read the docs**
- [README.md](README.md) - Start here
- [DESIGN.md](DESIGN.md) - Core concepts
- [DISCUSSION.md](DISCUSSION.md) - Open questions
- [examples/](examples/) - Sample configs

**💬 Provide feedback**
- Review the design
- Answer questions in [DISCUSSION.md](DISCUSSION.md)
- Suggest priorities
- Share use cases

**🚀 Coming soon**
- Implementation
- Beta testing
- First release

---

## The Vision

**Make personal machine automation:**
- As simple as writing a config file
- As safe as running it multiple times
- As smart as handling prerequisites automatically
- As helpful as guiding you through manual steps

**One command to set up a machine. One config to maintain it all.**

That's automata.

---

**Ready to discuss?** See [DISCUSSION.md](DISCUSSION.md) for specific questions and next steps.
