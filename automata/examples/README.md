# Automata Configuration Examples

This directory contains example configuration files demonstrating various use cases.

## Examples

### [personal-machine.yaml](personal-machine.yaml)
Complete setup for a personal development machine including:
- Core development tools (git, gh, jq, ripgrep)
- Programming languages (Go, Node, Python)
- Personal projects and dotfiles
- API keys and secrets
- Shell configuration

### [project-setup.yaml](project-setup.yaml)
Bootstrap a specific project with:
- Git repository setup
- Project dependencies (Node, PostgreSQL, Redis)
- Development tools (gopls, golangci-lint)
- Environment configuration
- Project-specific secrets

### [simple-git.yaml](simple-git.yaml)
Minimal example focusing on git repository management:
- Install git (with automatic Homebrew installation if needed)
- Initialize git repositories
- Configure remotes

### [secrets-management.yaml](secrets-management.yaml)
Different approaches to secret management:
- Environment variables
- File-based secrets
- Interactive prompts with guidance
- Secret validation

## Usage

Apply a configuration:
```bash
automata apply personal-machine.yaml
```

Preview changes without applying:
```bash
automata plan personal-machine.yaml
```

Validate configuration syntax:
```bash
automata validate personal-machine.yaml
```

## Configuration Format

All examples use YAML format with this structure:

```yaml
version: 1
resources:
  - type: <resource_type>
    # resource-specific attributes
```

### Resource Types

#### git_repo
Ensure a directory is a git repository.

**Attributes:**
- `path` (required): Directory path
- `remote` (optional): Git remote URL
- `branch` (optional): Default branch

#### package
Ensure a package is installed.

**Attributes:**
- `name` (required): Package name
- `provider` (optional): homebrew, apt, go, npm, pip
- `version` (optional): Version constraint

#### secret
Manage secrets with user interaction.

**Attributes:**
- `name` (required): Secret identifier
- `store` (required): env, file, keychain
- `path` (optional): File path (for file store)
- `prompt` (optional): User guidance
- `url` (optional): Where to obtain the secret
- `validate` (optional): Regex pattern for validation

#### file
Ensure files or directories exist.

**Attributes:**
- `path` (required): File/directory path
- `state` (required): present, absent, directory
- `content` (optional): File content
- `source` (optional): Source file to copy
- `mode` (optional): File permissions

## Customization

These examples are templates. Customize them for your needs:

1. Update package names for your preferred tools
2. Change repository URLs to your own
3. Adjust secret storage locations
4. Add or remove resources as needed

## Tips

- Start with `simple-git.yaml` to understand basics
- Use `--dry-run` to preview changes safely
- Resources are processed in dependency order automatically
- Running the same config multiple times is safe (idempotent)
- Automata will prompt before destructive operations

## Future Examples

Planned examples for future versions:
- CI/CD agent setup
- Docker development environment
- Multi-language polyglot project
- Team onboarding configuration
- Server provisioning basics
