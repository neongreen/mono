# Quick Start Guide

> **Note:** This is a design document. Implementation hasn't started yet.

This guide will help you get started with automata once it's implemented.

## Installation (Future)

```bash
# With Go
go install github.com/neongreen/mono/automata@latest

# Or download binary
curl -L https://github.com/neongreen/mono/releases/latest/download/automata-darwin-amd64 -o automata
chmod +x automata
mv automata /usr/local/bin/
```

## Your First Config

Create `config.yaml`:

```yaml
version: 1

resources:
  # Ensure git is installed
  - type: package
    name: git
  
  # Initialize a git repository
  - type: git_repo
    path: ~/projects/my-first-project
```

## Basic Commands

```bash
# Validate configuration
automata validate config.yaml

# Preview changes (dry-run)
automata plan config.yaml

# Apply configuration
automata apply config.yaml

# Check current state
automata check config.yaml
```

## Common Use Cases

### 1. Install Development Tools

```yaml
resources:
  - type: package
    name: git
  - type: package
    name: jq
  - type: package
    name: ripgrep
  - type: package
    name: gh  # GitHub CLI
```

### 2. Setup Project Structure

```yaml
resources:
  - type: file
    path: ~/projects
    state: directory
  
  - type: git_repo
    path: ~/projects/webapp
    remote: https://github.com/user/webapp.git
  
  - type: file
    path: ~/projects/webapp/data
    state: directory
```

### 3. Manage Secrets

```yaml
resources:
  - type: secret
    name: GITHUB_TOKEN
    store: env
    prompt: "Create a personal access token"
    url: https://github.com/settings/tokens
```

### 4. Complete Machine Setup

```yaml
resources:
  # Development tools
  - type: package
    name: git
  - type: package
    name: go
  - type: package
    name: node
  
  # Dotfiles
  - type: git_repo
    path: ~/.dotfiles
    remote: https://github.com/user/dotfiles.git
  
  # Projects
  - type: git_repo
    path: ~/projects/webapp
    remote: https://github.com/user/webapp.git
  
  - type: git_repo
    path: ~/projects/cli-tool
    remote: https://github.com/user/cli-tool.git
  
  # Secrets
  - type: secret
    name: GITHUB_TOKEN
    store: env
    url: https://github.com/settings/tokens
```

## Tips

### Organize with Multiple Files

Create modular configs:

```
setup/
├── main.yaml        # Includes all others
├── packages.yaml    # Development tools
├── projects.yaml    # Git repositories  
└── secrets.yaml     # API keys
```

`main.yaml`:
```yaml
version: 1
includes:
  - packages.yaml
  - projects.yaml
  - secrets.yaml
```

### Use Environment Variables

```yaml
resources:
  - type: git_repo
    path: ${HOME}/projects/myapp
  
  - type: secret
    name: ${PROJECT_NAME}_API_KEY
    store: env
```

### Comments for Documentation

```yaml
resources:
  # Core tools needed for all projects
  - type: package
    name: git
  
  # Language-specific tools
  - type: package
    name: go
    # Note: Will install latest stable version
```

### Apply Specific Resources

```bash
# Apply only git-related resources
automata apply --target="git_*" config.yaml

# Apply everything except secrets
automata apply --exclude="secret" config.yaml
```

### Dry-Run First

Always preview changes:

```bash
# See what will change
automata plan config.yaml

# Then apply if looks good
automata apply config.yaml
```

### Check State Without Changes

```bash
# See if system matches config
automata check config.yaml

# Output:
# ✓ git: installed
# ✓ ~/projects/myapp: git repository with correct remote
# ✗ GITHUB_TOKEN: not found in environment
```

## Common Workflows

### New Machine Setup

```bash
# 1. Clone your dotfiles
git clone https://github.com/user/dotfiles ~/dotfiles

# 2. Apply the configuration
cd ~/dotfiles
automata apply machine.yaml

# 3. Everything is set up!
```

### Daily Development

```bash
# Pull latest dotfiles
cd ~/dotfiles
git pull

# Re-apply (only changes are applied)
automata apply machine.yaml
```

### New Project Bootstrap

```bash
# Create project config
cat > project.yaml << 'EOF'
resources:
  - type: git_repo
    path: ~/projects/new-project
    remote: https://github.com/user/new-project.git
  - type: file
    path: ~/projects/new-project/.env
    content: |
      DATABASE_URL=postgresql://localhost/dev
EOF

# Set it up
automata apply project.yaml
```

### Add New Secret

```bash
# Add to config
cat >> config.yaml << 'EOF'
  - type: secret
    name: NEW_API_KEY
    store: env
    url: https://service.com/api-keys
EOF

# Apply (will only prompt for the new secret)
automata apply config.yaml
```

## Troubleshooting

### Config Validation Errors

```bash
# Check syntax and structure
automata validate config.yaml

# Will show specific errors:
# Error: line 5: unknown field 'typo'
# Error: resource 'git_repo' missing required field 'path'
```

### Permission Issues

```bash
# Check what automata wants to do
automata plan config.yaml

# Run with verbose output
automata apply --verbose config.yaml

# Check logs
automata apply --log-level=debug config.yaml
```

### Dependency Problems

```bash
# See the dependency graph
automata graph config.yaml

# Output shows resource order and dependencies
```

### Reset to Clean State

```bash
# Remove resources (careful!)
automata destroy config.yaml

# Will prompt before each destructive action
```

## Example Output

### Successful Apply

```
$ automata apply config.yaml

Parsing configuration... ✓
Validating resources... ✓
Resolving dependencies... ✓

Execution plan:
  • Install package 'git'
  • Create git repository at ~/projects/myapp
  • Configure secret 'GITHUB_TOKEN'

Continue? [y/n] y

[1/3] Installing git via homebrew...
      ✓ git 2.42.0 installed (5.2s)

[2/3] Creating git repository...
      ✓ Created ~/projects/myapp
      ✓ Initialized git repository (0.1s)

[3/3] Configuring GITHUB_TOKEN...
      Visit: https://github.com/settings/tokens
      Create a token with scopes: repo, workflow
      
      Enter token: ********************
      ✓ Stored in ~/.zshrc (0.5s)

Summary:
  ✓ 3 resources succeeded
  ✗ 0 resources failed
  
Total time: 5.8s
```

### Idempotent Run

```
$ automata apply config.yaml

Parsing configuration... ✓
Validating resources... ✓
Resolving dependencies... ✓

Checking current state:
  ✓ git: already installed
  ✓ ~/projects/myapp: already configured
  ✓ GITHUB_TOKEN: already set

No changes needed. System is in desired state.
```

## Next Steps

1. **Read the full design** - See [DESIGN.md](DESIGN.md)
2. **Review examples** - Check [examples/](examples/)
3. **Provide feedback** - Comment on [DISCUSSION.md](DISCUSSION.md)
4. **Wait for implementation** - Coming soon!

## Help and Support

- **Documentation:** See other .md files in this directory
- **Examples:** Browse [examples/](examples/) directory
- **Issues:** (GitHub issues once implemented)
- **Discussions:** (GitHub discussions once implemented)

---

**Status:** Design phase - implementation coming soon!
