# Automata Workflow Examples

This document provides visual workflows showing how automata works in practice.

## Basic Workflow

```
┌─────────────────────────────────────────────────────────────┐
│ User writes config.yaml                                     │
│                                                             │
│   resources:                                                │
│     - type: package                                         │
│       name: git                                             │
│     - type: git_repo                                        │
│       path: ~/projects/myapp                                │
└─────────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────┐
│ automata apply config.yaml                                  │
└─────────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────┐
│ 1. Parse YAML                                               │
│    ✓ Syntax valid                                           │
│    ✓ Schema valid                                           │
└─────────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────┐
│ 2. Validate Resources                                       │
│    ✓ package 'git' - valid                                  │
│    ✓ git_repo '~/projects/myapp' - valid                    │
└─────────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────┐
│ 3. Resolve Dependencies                                     │
│    git_repo depends on package:git                          │
│    Order: [package:git] → [git_repo:myapp]                  │
└─────────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────┐
│ 4. Check Current State                                      │
│    package:git → not installed                              │
│    git_repo:myapp → not exists                              │
└─────────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────┐
│ 5. Generate Plan                                            │
│    • Install package 'git' via homebrew                     │
│    • Create directory ~/projects/myapp                      │
│    • Initialize git repository                              │
│                                                             │
│    Continue? [y/n] y                                        │
└─────────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────┐
│ 6. Apply Changes                                            │
│    ✓ Installed git (2.42.0)                                 │
│    ✓ Created ~/projects/myapp                               │
│    ✓ Initialized git repository                             │
└─────────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────┐
│ 7. Verify State                                             │
│    ✓ git is installed and available                         │
│    ✓ ~/projects/myapp/.git exists                           │
│                                                             │
│    Success! 2 resources applied, 0 errors                   │
└─────────────────────────────────────────────────────────────┘
```

## Smart Dependency Installation

What happens when prerequisites are missing:

```
┌──────────────────────────────────┐
│ User: automata apply config.yaml │
└──────────────────────────────────┘
           ↓
┌──────────────────────────────────┐
│ Need to install 'git' package    │
└──────────────────────────────────┘
           ↓
┌──────────────────────────────────┐
│ Check: Is Homebrew available?    │
│ Result: NO                        │
└──────────────────────────────────┘
           ↓
┌──────────────────────────────────────────────────────┐
│ Automata: Git package requires Homebrew.             │
│ Homebrew is not installed.                           │
│                                                       │
│ Would you like to install Homebrew first?            │
│ [y] Yes, install Homebrew                            │
│ [n] No, cancel operation                             │
│ [m] Manual - I'll install it myself                  │
│                                                       │
│ Choice: y                                            │
└──────────────────────────────────────────────────────┘
           ↓
┌──────────────────────────────────┐
│ Installing Homebrew...            │
│ [=========================>] 95%  │
└──────────────────────────────────┘
           ↓
┌──────────────────────────────────┐
│ ✓ Homebrew installed              │
└──────────────────────────────────┘
           ↓
┌──────────────────────────────────┐
│ Installing git via Homebrew...    │
│ [========================>] 90%   │
└──────────────────────────────────┘
           ↓
┌──────────────────────────────────┐
│ ✓ git installed                   │
└──────────────────────────────────┘
           ↓
┌──────────────────────────────────┐
│ ✓ git_repo created                │
│                                   │
│ All done! ✨                      │
└──────────────────────────────────┘
```

## Interactive Secret Management

Secret resource with human-in-the-loop:

```
┌─────────────────────────────────────────────────────────┐
│ Config contains:                                        │
│                                                         │
│   - type: secret                                        │
│     name: GITHUB_TOKEN                                  │
│     store: env                                          │
│     url: https://github.com/settings/tokens             │
└─────────────────────────────────────────────────────────┘
                         ↓
┌─────────────────────────────────────────────────────────┐
│ Check: Does GITHUB_TOKEN exist in environment?         │
│ Result: NO                                              │
└─────────────────────────────────────────────────────────┘
                         ↓
┌─────────────────────────────────────────────────────────┐
│ Automata: GITHUB_TOKEN not found                       │
│                                                         │
│ You'll need to create a GitHub Personal Access Token   │
│                                                         │
│ Steps:                                                  │
│ 1. Visit: https://github.com/settings/tokens           │
│ 2. Click "Generate new token"                          │
│ 3. Select scopes: repo, workflow                       │
│ 4. Generate and copy the token                         │
│                                                         │
│ Press Enter when ready...                              │
└─────────────────────────────────────────────────────────┘
                         ↓
                    [User browses to GitHub]
                    [User creates token]
                    [User presses Enter]
                         ↓
┌─────────────────────────────────────────────────────────┐
│ Enter your GITHUB_TOKEN:                                │
│ > ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx              │
└─────────────────────────────────────────────────────────┘
                         ↓
┌─────────────────────────────────────────────────────────┐
│ Validating token format...                              │
│ ✓ Token format is valid                                 │
└─────────────────────────────────────────────────────────┘
                         ↓
┌─────────────────────────────────────────────────────────┐
│ Where should I store GITHUB_TOKEN?                      │
│                                                         │
│ Config says: ~/.zshrc                                   │
│                                                         │
│ Will add: export GITHUB_TOKEN="ghp_xxx..."             │
│                                                         │
│ Confirm? [y/n] y                                        │
└─────────────────────────────────────────────────────────┘
                         ↓
┌─────────────────────────────────────────────────────────┐
│ ✓ Added GITHUB_TOKEN to ~/.zshrc                        │
│                                                         │
│ Note: Run 'source ~/.zshrc' or restart your shell      │
│       to use the token in current session              │
└─────────────────────────────────────────────────────────┘
```

## Idempotent Behavior

Running the same config multiple times:

```
First Run:
──────────────────────────────────────────
automata apply config.yaml

Checking: package:git
  → Not installed
  → Installing... ✓

Checking: git_repo:myapp
  → Directory not exists
  → Creating... ✓
  → Initializing git... ✓

Result: 2 changes made


Second Run (immediately after):
──────────────────────────────────────────
automata apply config.yaml

Checking: package:git
  → Already installed ✓
  → No action needed

Checking: git_repo:myapp
  → Directory exists ✓
  → Is git repository ✓
  → No action needed

Result: 0 changes made (already in desired state)
```

## Plan Mode (Dry Run)

Preview changes without applying:

```
┌─────────────────────────────────────────────────────────┐
│ automata plan config.yaml                               │
└─────────────────────────────────────────────────────────┘
                         ↓
┌─────────────────────────────────────────────────────────┐
│ Execution Plan:                                         │
│                                                         │
│ [+] package:git                                         │
│     Will install git via homebrew                       │
│     Estimated time: 30s                                 │
│                                                         │
│ [+] git_repo:myapp                                      │
│     Will create directory: ~/projects/myapp             │
│     Will initialize git repository                      │
│     Estimated time: 1s                                  │
│                                                         │
│ [?] secret:GITHUB_TOKEN                                 │
│     Will prompt for interactive input                   │
│     Will store in: ~/.zshrc                             │
│                                                         │
│ Summary:                                                │
│   2 resources to create                                 │
│   0 resources to update                                 │
│   0 resources to delete                                 │
│   1 resource requires interaction                       │
│                                                         │
│ No changes made (plan mode)                             │
│ Run 'automata apply config.yaml' to execute            │
└─────────────────────────────────────────────────────────┘
```

## Error Handling

What happens when something fails:

```
┌─────────────────────────────────────────────────────────┐
│ automata apply config.yaml                              │
└─────────────────────────────────────────────────────────┘
                         ↓
┌─────────────────────────────────────────────────────────┐
│ [1/3] package:git                                       │
│       Installing git via homebrew...                    │
│       ✓ Success                                         │
└─────────────────────────────────────────────────────────┘
                         ↓
┌─────────────────────────────────────────────────────────┐
│ [2/3] git_repo:myapp                                    │
│       Creating ~/projects/myapp...                      │
│       ✗ ERROR: Permission denied                        │
│                                                         │
│       Details:                                          │
│       Cannot create directory ~/projects/myapp          │
│       Parent directory ~/projects is not writable       │
│                                                         │
│       Suggestions:                                      │
│       • Check directory permissions                     │
│       • Create ~/projects manually with:                │
│         mkdir ~/projects                                │
│       • Run with sudo (not recommended)                 │
└─────────────────────────────────────────────────────────┘
                         ↓
┌─────────────────────────────────────────────────────────┐
│ Execution stopped due to error                          │
│                                                         │
│ Summary:                                                │
│   ✓ 1 resource succeeded                                │
│   ✗ 1 resource failed                                   │
│   ⊘ 1 resource skipped                                  │
│                                                         │
│ Exit code: 1                                            │
└─────────────────────────────────────────────────────────┘
```

## Parallel Execution

Resources without dependencies run in parallel:

```
Resource Graph:
───────────────

    [homebrew]
         ↓
    [package:git]
         ↓
    ┌────┴────┐
    ↓         ↓
[repo:app] [repo:notes]
    ↓         ↓
[secret:A] [secret:B]


Execution Timeline:
───────────────────

t=0s   [homebrew] ══════════════════════> ✓
                     
t=10s  [package:git] ════════> ✓
                     
t=15s  [repo:app] ═════> ✓    [repo:notes] ═════> ✓
       (parallel)              (parallel)
                     
t=20s  [secret:A] ⋯⋯⋯> ✓     [secret:B] ⋯⋯⋯> ✓
       (user input)            (user input)
```

## Configuration Composition

Breaking config into multiple files:

```
machine-setup.yaml:
──────────────────
version: 1
includes:
  - packages.yaml
  - projects.yaml
  - secrets.yaml


packages.yaml:
──────────────
resources:
  - type: package
    name: git
  - type: package
    name: jq


projects.yaml:
──────────────
resources:
  - type: git_repo
    path: ~/projects/app1
  - type: git_repo
    path: ~/projects/app2


Result when applying machine-setup.yaml:
─────────────────────────────────────────
All resources from all files are merged
and executed in dependency order.
```

## Use Case: New Machine Setup

Complete workflow for setting up a new development machine:

```
Day 1 - New Laptop Arrives
═══════════════════════════

1. Install Go
   curl https://go.dev/dl/... | sh

2. Install automata
   go install github.com/neongreen/mono/automata@latest

3. Clone dotfiles
   git clone https://github.com/user/dotfiles ~/dotfiles

4. Run automata
   cd ~/dotfiles
   automata apply machine.yaml

   [Automata runs through all resources]
   [Prompts for secrets as needed]
   [Installs all tools]
   [Sets up all projects]

5. 30 minutes later...
   ✓ All development tools installed
   ✓ All projects cloned
   ✓ All secrets configured
   ✓ Ready to code! 🚀


Future Updates (Week 2, Month 2, etc.)
═══════════════════════════════════════

1. Pull dotfiles
   cd ~/dotfiles
   git pull

2. Re-run automata
   automata apply machine.yaml

   [Only applies changes]
   [New packages installed]
   [New projects cloned]
   [Existing resources unchanged]

3. Always up to date! ✨
```

## Comparison with Manual Approach

### Manual Setup Script
```bash
#!/bin/bash
set -e

# Install Homebrew
if ! command -v brew &> /dev/null; then
    /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
fi

# Install git
if ! command -v git &> /dev/null; then
    brew install git
fi

# Setup repo
if [ ! -d ~/projects/myapp ]; then
    mkdir -p ~/projects/myapp
    cd ~/projects/myapp
    git init
fi

# Handle secret
if [ -z "$GITHUB_TOKEN" ]; then
    echo "Please set GITHUB_TOKEN"
    exit 1
fi
```

Problems:
- ❌ Not declarative
- ❌ Hard to read
- ❌ Error handling is manual
- ❌ No dry-run mode
- ❌ Secret handling is awkward

### With Automata
```yaml
resources:
  - type: package
    name: git
  - type: git_repo
    path: ~/projects/myapp
  - type: secret
    name: GITHUB_TOKEN
    url: https://github.com/settings/tokens
```

Benefits:
- ✅ Clear and readable
- ✅ Automatic error handling
- ✅ Built-in dry-run
- ✅ Interactive secret prompts
- ✅ Idempotent by design

---

These workflows demonstrate how automata makes system automation simple, safe, and pleasant!
