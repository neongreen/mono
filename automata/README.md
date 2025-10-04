# Automata

> A declarative system automation tool for personal machine setup and programming tasks

**Status:** 🚧 Design Phase - See [DESIGN.md](DESIGN.md) for detailed design discussion

## What is Automata?

Automata is like Ansible for your personal machine. Define what you want, and it makes it happen:

- **Declarative:** Describe desired state, not steps
- **Idempotent:** Safe to run multiple times
- **Smart:** Auto-installs prerequisites (e.g., Homebrew before Git)
- **Interactive:** Guides you through manual steps (like obtaining API keys)

## Example

```yaml
# machine-setup.yaml
resources:
  # Ensure directory is a git repository
  - type: git_repo
    path: ~/projects/myapp
    remote: https://github.com/user/myapp.git
  
  # Ensure git is installed (will install Homebrew first if needed)
  - type: package
    name: git
    
  # Manage API keys with prompts
  - type: secret
    name: GITHUB_API_KEY
    store: env
    url: https://github.com/settings/tokens
```

```bash
automata apply machine-setup.yaml
```

## Use Cases

- **New Machine Setup:** One command to configure a fresh installation
- **Dotfiles Management:** Ensure configs are always in place
- **Development Environment:** Set up projects with all dependencies
- **Secret Management:** Interactive secret provisioning with guidance
- **CI/CD Bootstrap:** Prepare build agents declaratively

## Why Another Tool?

**vs Ansible:**
- Lighter (single Go binary)
- Designed for personal use, not datacenter scale
- Embraces human interaction for secrets

**vs Homebrew Bundle:**
- Manages more than packages (git repos, secrets, files)
- Cross-platform
- Describes complete system state

**vs Shell Scripts:**
- Declarative, not imperative
- Idempotent by design
- Clear state representation

## Current Status

This tool is in the **design phase**. We're discussing:

1. Resource types and their APIs
2. Configuration format
3. Dependency resolution approach
4. User interaction model for secrets
5. Testing strategy

See [DESIGN.md](DESIGN.md) for the complete design document.

## Example Configurations

See [examples/](examples/) directory for sample configurations:
- `personal-machine.yaml` - Full machine setup
- `project-setup.yaml` - Development project bootstrap
- `secrets.yaml` - Secret management examples

## Contributing

Currently seeking feedback on the design! Please review [DESIGN.md](DESIGN.md) and provide your thoughts on:

- Resource types we should support
- CLI interface and commands
- Configuration format
- Features that would make this useful for you

## Roadmap

- [x] Initial design document
- [ ] Design review and refinement
- [ ] Core framework implementation
- [ ] Basic resource types (file, git, package)
- [ ] Secret management with prompts
- [ ] Testing infrastructure
- [ ] Documentation and examples
- [ ] v0.1.0 release

## License

MIT License - See LICENSE file in the repository root.
