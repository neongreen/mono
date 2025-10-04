# Automata - Declarative System Automation Tool

## Overview

Automata is a declarative automation tool for managing system state, similar to Ansible but designed for personal machine administration and programming tasks. It allows users to specify desired states and automatically brings the system to that state, handling dependencies and prerequisites automatically.

## Core Concepts

### 1. Declarative State Management

Users define **what they want** rather than **how to achieve it**:

```yaml
# Example configuration
resources:
  - type: git_repo
    path: /home/user/projects/myapp
    remote: https://github.com/user/myapp.git
    
  - type: package
    name: git
    provider: homebrew
    
  - type: secret
    name: GITHUB_API_KEY
    store: env
    prompt: "Get your API key from https://github.com/settings/tokens"
```

### 2. Idempotent Operations

- Check current state before taking action
- Only modify if needed to reach desired state
- Safe to run multiple times

### 3. Dependency Resolution

- Automatic prerequisite installation (e.g., Homebrew before Git)
- Topological sorting of resources
- Clear error messages when dependencies can't be satisfied

## Architecture

### Core Components

```
automata/
├── main.go              # CLI entry point
├── config/              # Configuration parsing
│   ├── parser.go
│   └── validator.go
├── resource/            # Resource interface and registry
│   ├── interface.go
│   ├── registry.go
│   └── types/           # Specific resource implementations
│       ├── git.go
│       ├── package.go
│       ├── secret.go
│       └── file.go
├── provider/            # Package manager providers
│   ├── interface.go
│   ├── homebrew.go
│   ├── apt.go
│   └── go.go
├── executor/            # Execution engine
│   ├── planner.go       # Dependency resolution
│   └── runner.go        # State changes
└── ui/                  # User interaction
    ├── prompt.go
    └── output.go
```

## Resource Types

### 1. Git Repository (`git_repo`)

**Purpose:** Ensure a directory is a git repository

**Attributes:**
- `path` (required): Directory path
- `remote` (optional): Remote URL to set
- `branch` (optional): Default branch

**States:**
- Not exists → Create directory + git init
- Exists but not git → Error (safety) or git init (with force flag)
- Already git repo → Check remote/branch, update if needed

**Example:**
```yaml
- type: git_repo
  path: /home/user/projects/myapp
  remote: https://github.com/user/myapp.git
  branch: main
```

### 2. Package (`package`)

**Purpose:** Ensure a package/tool is installed

**Attributes:**
- `name` (required): Package name
- `provider` (optional): homebrew, apt, go, npm, pip
- `version` (optional): Specific version

**Auto-detection:**
- Detect OS and available package managers
- Install package manager if needed (e.g., Homebrew on macOS)
- Choose appropriate provider if not specified

**Example:**
```yaml
- type: package
  name: git
  provider: homebrew

- type: package
  name: github.com/neongreen/mono/markdown-format
  provider: go
```

### 3. Secret (`secret`)

**Purpose:** Manage secrets/API keys with user interaction

**Attributes:**
- `name` (required): Secret identifier
- `store` (required): env, file, keychain
- `prompt` (optional): User guidance message
- `url` (optional): Where to obtain the secret

**Flow:**
1. Check if secret exists
2. If not, show prompt/URL to user
3. Interactively collect secret from user
4. Store in specified location
5. Verify storage

**Example:**
```yaml
- type: secret
  name: GITHUB_API_KEY
  store: env
  prompt: "Create a personal access token"
  url: https://github.com/settings/tokens
```

### 4. File/Directory (`file`)

**Purpose:** Ensure files/directories exist with specific content

**Attributes:**
- `path` (required): File/directory path
- `state` (required): present, absent, directory
- `content` (optional): File content
- `source` (optional): Source file to copy
- `mode` (optional): File permissions

## Configuration Format

### YAML-based (Primary)

```yaml
version: 1
resources:
  - type: git_repo
    name: myapp-repo
    path: ~/projects/myapp
    
  - type: package
    name: git
    
  - type: secret
    name: API_KEY
    store: env
```

### Programmatic Go API (Advanced)

```go
config := automata.Config{
    Resources: []automata.Resource{
        &resource.GitRepo{
            Path: "~/projects/myapp",
        },
    },
}
executor.Run(config)
```

## Execution Model

### 1. Parse & Validate
- Read configuration file
- Validate syntax and resource definitions
- Check for circular dependencies

### 2. Plan
- Determine current state for each resource
- Calculate required changes
- Resolve dependencies and order operations
- Show plan to user (dry-run mode)

### 3. Execute
- Apply changes in dependency order
- Handle failures gracefully
- Provide progress feedback
- Generate summary report

### 4. Verify
- Confirm desired state achieved
- Report any discrepancies

## CLI Interface

```bash
# Apply configuration
automata apply config.yaml

# Dry-run (show what would change)
automata plan config.yaml

# Check current state
automata check config.yaml

# Interactive mode
automata init

# Validate configuration
automata validate config.yaml
```

## Key Design Decisions

### 1. Human-in-the-Loop for Secrets

Unlike fully automated systems, automata embraces human interaction for secrets:
- Can't automatically obtain API keys
- Shows user where to go (URL)
- Prompts for input
- Validates and stores securely

### 2. Safety First

- Never destructive by default
- Ask confirmation for dangerous operations
- Preserve existing data
- Clear error messages

### 3. Extensible Architecture

- Plugin-based resource types
- Easy to add new providers
- Clean interfaces for testing

### 4. Smart Dependency Management

- Auto-install prerequisites (with permission)
- Example flow for `git` package:
  1. Check if git installed → No
  2. Check if Homebrew installed → No
  3. Prompt: "Need to install Homebrew to install git. Continue?"
  4. Install Homebrew
  5. Install git

## Implementation Phases

### Phase 1: Core Framework (MVP)
- [ ] Configuration parser (YAML)
- [ ] Resource interface and registry
- [ ] Basic executor (no dependency resolution)
- [ ] CLI with apply/plan commands
- [ ] One resource type: `file` (simplest)

### Phase 2: Essential Resources
- [ ] Git repository resource
- [ ] Package resource with Homebrew provider
- [ ] Dependency resolution
- [ ] Better error handling

### Phase 3: Secrets & Interaction
- [ ] Secret resource with user prompts
- [ ] Interactive UI improvements
- [ ] Progress reporting

### Phase 4: Enhanced Providers
- [ ] Additional package providers (apt, go, npm, pip)
- [ ] Keychain integration for secrets
- [ ] Advanced git operations (clone, pull)

### Phase 5: Polish
- [ ] Comprehensive testing
- [ ] Documentation
- [ ] Example configurations
- [ ] Integration with mise/other tools

## Open Questions for Discussion

1. **Configuration Format:**
   - YAML only, or support JSON/TOML too?
   - Should we support programmatic Go API initially?

2. **State Storage:**
   - Should automata track what it has done (statefile)?
   - Or always check actual system state (stateless)?

3. **Privilege Escalation:**
   - How to handle operations requiring sudo?
   - Prompt each time or request upfront?

4. **Error Recovery:**
   - Rollback on failure?
   - Continue on error with flag?

5. **Scope:**
   - Focus on personal machine setup first?
   - Or design for multi-machine from start?

6. **Package Providers:**
   - Which providers are priority?
   - Should we support all major OS package managers?

7. **Secret Storage:**
   - Start with env vars only?
   - Which secret stores to support? (keychain, 1Password, etc.)

8. **Testing Strategy:**
   - How to test system-modifying operations?
   - Mock providers? Docker containers?

## Similar Tools Comparison

### vs Ansible
- **Lighter:** Single binary, no Python dependencies
- **Simpler:** Focused on personal use, not datacenter scale
- **Interactive:** Embraces human input for secrets
- **Go-based:** Easy to distribute and install

### vs Homebrew Bundle
- **Broader:** Not just packages, includes git repos, secrets, etc.
- **Cross-platform:** Can work on Linux too
- **Declarative:** Describes complete system state

### vs Nix/NixOS
- **Pragmatic:** Works on existing systems, no special OS
- **Gradual:** Can adopt incrementally
- **Familiar:** YAML config, not new language

## Success Criteria

A successful implementation should:

1. Make it trivial to set up a new machine
2. Be safe to run repeatedly (idempotent)
3. Handle common tasks (git, packages, secrets)
4. Guide users through manual steps when needed
5. Have clear, helpful error messages
6. Be well-tested and reliable

## Next Steps

Before implementation, we should:

1. **Validate requirements** - Confirm this matches your needs
2. **Prioritize features** - Which resource types are most important?
3. **Choose scope** - MVP vs full feature set for initial release
4. **Decide on format** - YAML structure and resource schemas
5. **Review architecture** - Any concerns with proposed design?

---

**Feedback Welcome!**

Please review and provide feedback on:
- Overall approach and architecture
- Resource types and their attributes
- CLI commands and workflow
- Priority of features for MVP
- Any missing requirements or use cases
