# Design Discussion: Automata Tool

This document guides the discussion about the automata tool design. Please review the design documents and provide feedback on the questions below.

## Documents to Review

1. **[README.md](README.md)** - High-level overview and use cases
2. **[DESIGN.md](DESIGN.md)** - Detailed design including resource types and execution model
3. **[ARCHITECTURE.md](ARCHITECTURE.md)** - Technical architecture and implementation details
4. **[examples/](examples/)** - Sample configuration files

## Key Design Decisions to Discuss

### 1. Scope for MVP

**Question:** What should be included in the first working version?

**Proposed MVP:**
- YAML configuration parsing
- Basic executor (sequential, no dependency resolution)
- Three resource types:
  - `file` - Create/manage files and directories
  - `git_repo` - Initialize git repositories
  - `package` - Install packages via Homebrew (macOS only)
- CLI: `apply` and `validate` commands

**Alternative:** Start even simpler with just `file` resource?

### 2. Resource Types Priority

**Question:** Which resource types are most important to you?

**Proposed order:**
1. `file` (simplest, good for testing)
2. `git_repo` (matches your example)
3. `package` (enables tool installation)
4. `secret` (interactive, more complex)

**Your priorities:** Which would you use most?

### 3. Package Providers

**Question:** Which package managers should we support initially?

**Options:**
- Start with Homebrew only (macOS focus)
- Add apt immediately (Linux support)
- Add go toolchain (for Go tools)
- Support all major providers from start

**Your environment:** What do you primarily use?

### 4. Configuration Format

**Question:** Is YAML the right choice?

**Current proposal:**
```yaml
version: 1
resources:
  - type: git_repo
    path: ~/projects/myapp
```

**Alternatives:**
- JSON (more verbose, machine-friendly)
- TOML (less indentation-sensitive)
- HCL (Terraform-style)
- Custom DSL

**Your preference:** Any strong opinions?

### 5. Dependency Management

**Question:** How important is automatic dependency resolution in the MVP?

**Option A - Manual ordering (simpler):**
```yaml
resources:
  - type: package
    name: git
  - type: git_repo  # Assumes git already installed
    path: ~/projects/myapp
```
User must order resources correctly.

**Option B - Automatic resolution (complex):**
```yaml
resources:
  - type: git_repo
    path: ~/projects/myapp
  - type: package
    name: git
```
Tool figures out git must be installed first.

**Your need:** Is automatic ordering important from day 1?

### 6. State Management

**Question:** Should automata track state?

**Stateless (proposed):**
- Always check actual system state
- No state file to maintain
- Simpler, less prone to issues
- Slightly slower

**Stateful:**
- Maintain `.automata/state.json`
- Track what automata created
- Faster checks
- Can drift from reality

**Your preference:** Which approach fits better?

### 7. Interactive Features

**Question:** How important is the interactive secret management?

**High priority:** Implement in MVP
- Essential for your use case
- Include prompt system
- Add secret resource type early

**Lower priority:** Defer to v0.2
- Focus on non-interactive resources first
- Add interactivity later
- Simpler MVP

**Your use case:** How often would you use interactive prompts?

### 8. Error Handling

**Question:** What should happen when a resource fails?

**Stop on error (proposed):**
- Safer default
- Clear what failed
- Easy to debug

**Continue on error:**
- More forgiving
- Might leave system in inconsistent state
- Could hide problems

**Your preference:** Which behavior is more useful?

### 9. Platform Support

**Question:** Which platforms should we target?

**Options:**
- macOS only (simpler, matches Homebrew focus)
- macOS + Linux (broader appeal)
- Windows too (significant effort)

**Your environment:** What do you use?

### 10. Testing Approach

**Question:** How should we test system-modifying operations?

**Options:**
- Mock all system operations (unit tests only)
- Docker containers (real operations, isolated)
- Both (comprehensive but complex)
- Manual testing only (faster development)

**Your preference:** Test rigor vs development speed?

## Naming Discussion

**Current name:** `automata`

**Considerations:**
- Easy to type
- Related to automation
- Not already taken (mostly)

**Alternatives:**
- `declara` (declarative)
- `stateful` (state management focus)
- `machina` (machine setup focus)
- `toolkit` (generic tools)
- Something else?

**Your opinion:** Like `automata` or prefer alternatives?

## Example Use Cases

**Question:** Which examples resonate most with your needs?

Please review [examples/](examples/) and indicate:
1. Which ones match your use cases?
2. What's missing?
3. Any examples that seem unnecessary?

## Implementation Timeline

**Question:** What's your timeline/urgency?

**Options:**
- Quick MVP (2 weeks, basic features)
- Full featured (6-8 weeks, polished)
- Gradual (deliver incrementally)

**Your need:** When would you want to start using this?

## Related to Your Original Request

Based on your problem statement, here's how the design addresses each point:

### ✅ "Folder becomes a git repository"
Addressed by `git_repo` resource:
```yaml
- type: git_repo
  path: ~/projects/myapp
```

### ✅ "Declarative"
Entire design is declarative - describe desired state.

### ✅ "Idempotent"
All resources check state before modifying, safe to run multiple times.

### ✅ "Propose to install git"
Handled by dependency system - if git not available, offer to install.

### ✅ "Install homebrew if not available"
Package provider system can install prerequisite package managers.

### ✅ "Human in the loop for API keys"
Secret resource with interactive prompts:
```yaml
- type: secret
  name: GITHUB_API_KEY
  prompt: "Get your key from GitHub settings"
  url: https://github.com/settings/tokens
```

**Missing anything:** Does this cover your requirements?

## Open Questions

Additional topics to consider:

1. **Logging:** How verbose should output be? Multiple levels?

2. **Configuration validation:** How strict? Fail on unknown fields?

3. **Dry-run mode:** Essential for MVP or can wait?

4. **Undo/rollback:** Important or too complex for first version?

5. **Multi-machine sync:** Out of scope or design for it?

6. **Secret storage:** Which backends? (env, file, keychain, 1Password, etc.)

7. **Version constraints:** For packages, how to specify versions?

8. **Remote configs:** Support fetching config from URLs?

9. **Includes/imports:** Split config across multiple files?

10. **Templating:** Variables in configs? Environment variable substitution?

## Next Steps

After reviewing the design documents, please provide feedback on:

1. ✅ **Approval:** Is the overall approach sound?
2. 📋 **Priorities:** Which features are most important?
3. 🔧 **Changes:** What needs to be modified?
4. ❓ **Questions:** What's unclear?
5. 💡 **Ideas:** Any features missing?

Once we align on the design, we can:
1. Finalize the MVP scope
2. Create detailed implementation plan
3. Set up project structure
4. Begin implementation

---

## How to Provide Feedback

Please comment on:
- Specific numbered questions above
- General design direction
- Any concerns or suggestions
- Priority of features for your use case

Looking forward to your thoughts!
