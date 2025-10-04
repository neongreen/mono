# Automata Architecture

This document provides a detailed technical architecture for the automata tool.

## Core Principles

1. **Resource-Based Design:** Everything is a resource with a desired state
2. **Idempotency:** Operations produce the same result when run multiple times
3. **Dependency Awareness:** Automatic resolution of prerequisites
4. **Fail-Safe:** Never destructive without confirmation
5. **Testability:** Pure functions, interfaces for mocking

## Component Diagram

```
┌─────────────────────────────────────────────────────────┐
│                      CLI (main.go)                      │
│  Commands: apply, plan, check, validate, init          │
└────────────┬────────────────────────────────────────────┘
             │
             ▼
┌─────────────────────────────────────────────────────────┐
│                   Config Parser                         │
│  - Read YAML/JSON                                       │
│  - Validate schema                                      │
│  - Build resource graph                                 │
└────────────┬────────────────────────────────────────────┘
             │
             ▼
┌─────────────────────────────────────────────────────────┐
│                   Executor/Planner                      │
│  - Topological sort (dependency order)                  │
│  - Generate execution plan                              │
│  - Dry-run simulation                                   │
└────────────┬────────────────────────────────────────────┘
             │
             ▼
┌─────────────────────────────────────────────────────────┐
│                  Resource Registry                      │
│  - Map resource types to implementations                │
│  - Lifecycle hooks (validate, plan, apply, check)       │
└────────────┬────────────────────────────────────────────┘
             │
             ▼
┌──────────────────────────────────────────────────────────┐
│                Resource Implementations                  │
│                                                          │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌────────┐  │
│  │ Git Repo │  │ Package  │  │  Secret  │  │  File  │  │
│  └──────────┘  └──────────┘  └──────────┘  └────────┘  │
│                                                          │
└──────────────────────────────────────────────────────────┘
             │
             ▼
┌──────────────────────────────────────────────────────────┐
│                   Provider System                        │
│                                                          │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌────────┐  │
│  │ Homebrew │  │   Apt    │  │    Go    │  │  Npm   │  │
│  └──────────┘  └──────────┘  └──────────┘  └────────┘  │
│                                                          │
└──────────────────────────────────────────────────────────┘
             │
             ▼
┌──────────────────────────────────────────────────────────┐
│                    System Interface                      │
│  - File operations                                       │
│  - Process execution                                     │
│  - User interaction                                      │
└──────────────────────────────────────────────────────────┘
```

## Resource Interface

All resources implement a common interface:

```go
type Resource interface {
    // Type returns the resource type name (e.g., "git_repo", "package")
    Type() string
    
    // Validate checks if the resource configuration is valid
    Validate() error
    
    // Dependencies returns list of other resources this depends on
    Dependencies() []string
    
    // Check examines current state and returns whether it matches desired state
    Check(ctx context.Context) (CheckResult, error)
    
    // Plan describes what changes would be made
    Plan(ctx context.Context) (Plan, error)
    
    // Apply makes changes to reach desired state
    Apply(ctx context.Context) error
}

type CheckResult struct {
    InDesiredState bool
    CurrentState   map[string]interface{}
    DesiredState   map[string]interface{}
}

type Plan struct {
    Changes []Change
}

type Change struct {
    Field     string
    Current   interface{}
    Desired   interface{}
    Action    string  // "create", "update", "delete", "none"
}
```

## Resource Lifecycle

Each resource goes through this lifecycle:

```
1. Parse Config
   ↓
2. Validate
   ↓
3. Resolve Dependencies
   ↓
4. Check Current State ────→ [Already in desired state] → Done
   ↓
   [Needs changes]
   ↓
5. Generate Plan
   ↓
6. User Approval (if needed)
   ↓
7. Apply Changes
   ↓
8. Verify State
```

## Dependency Resolution

### Implicit Dependencies

Some dependencies are implicit based on resource types:
- `package` resources depend on their provider being available
- `git_repo` resources depend on `git` package
- `secret` with file store depends on parent directory existing

### Explicit Dependencies

Users can declare explicit dependencies:

```yaml
resources:
  - type: git_repo
    name: my-repo
    path: ~/projects/myapp
    
  - type: file
    name: config-file
    path: ~/projects/myapp/config.yaml
    depends_on:
      - my-repo  # Only create after repo exists
```

### Resolution Algorithm

1. Build dependency graph (DAG)
2. Detect cycles → Error if found
3. Topological sort
4. Execute in sorted order

## Provider System

Providers handle different package managers:

```go
type PackageProvider interface {
    // Name returns provider name (e.g., "homebrew", "apt")
    Name() string
    
    // Available checks if this provider is available on the system
    Available(ctx context.Context) bool
    
    // Install installs the provider itself (e.g., install Homebrew)
    Install(ctx context.Context) error
    
    // IsInstalled checks if a package is installed
    IsInstalled(ctx context.Context, pkg string) (bool, error)
    
    // InstallPackage installs a package
    InstallPackage(ctx context.Context, pkg string, version string) error
    
    // UpdatePackage updates a package
    UpdatePackage(ctx context.Context, pkg string, version string) error
}
```

### Provider Selection

1. If `provider` specified → Use that provider
2. Auto-detect based on:
   - OS (macOS → homebrew, Ubuntu → apt)
   - Package name (e.g., github.com/* → go)
3. Check if provider is available
4. If not, offer to install provider

## State Management

### Stateless Approach (Recommended)

- No state file maintained
- Always check actual system state
- Idempotent by design
- Simpler, less prone to drift

**Pros:**
- No state file corruption issues
- Works across machines
- True reflection of system state
- No sync problems

**Cons:**
- Slightly slower (must check state each time)
- Can't track what automata did vs manual changes

### Stateful Approach (Alternative)

- Maintain `.automata/state.json`
- Track what automata created/modified
- Faster checks (compare to state file)

**Pros:**
- Faster execution
- Can show what automata manages vs external changes
- Enable "undo" functionality

**Cons:**
- State file can drift from reality
- Must handle state file corruption
- Sync issues in teams

**Decision:** Start with **stateless** for simplicity. Add state tracking later if needed.

## Error Handling Strategy

### Levels of Errors

1. **Configuration Errors** (fail fast)
   - Invalid YAML syntax
   - Unknown resource types
   - Missing required fields
   - Circular dependencies

2. **Pre-flight Errors** (fail before changes)
   - Missing prerequisites (with auto-fix offer)
   - Permission issues
   - Validation failures

3. **Runtime Errors** (during apply)
   - Network failures
   - Disk full
   - Process crashes

### Error Recovery

```go
type ExecutionResult struct {
    Succeeded []ResourceResult
    Failed    []ResourceResult
    Skipped   []ResourceResult
}

type ResourceResult struct {
    Resource Resource
    Error    error
    Changes  []Change
}
```

**Strategy:**
- Stop on first error by default
- `--continue-on-error` flag to proceed
- Rollback not implemented initially (complex, error-prone)
- Clear error messages with suggestions

## User Interaction

### Interactive Prompts

For secrets and confirmations:

```go
type Prompter interface {
    // Confirm asks yes/no question
    Confirm(message string, defaultYes bool) (bool, error)
    
    // Input asks for text input
    Input(prompt string, defaultValue string) (string, error)
    
    // Secret asks for sensitive input (hidden)
    Secret(prompt string) (string, error)
    
    // Select presents multiple choices
    Select(prompt string, options []string) (int, error)
}
```

### Output Formatting

Multiple output modes:
- **Human:** Colored, formatted for reading
- **JSON:** Machine-readable
- **Quiet:** Minimal output

```bash
automata apply config.yaml                    # Human-friendly
automata apply config.yaml --output=json      # JSON output
automata apply config.yaml --quiet            # Minimal output
```

## Testing Strategy

### Unit Tests

- Pure functions tested in isolation
- Mock interfaces (Prompter, PackageProvider, etc.)
- Table-driven tests

### Integration Tests

- Real operations in Docker containers
- Test actual package installation
- Verify git operations

### Example:

```go
func TestGitRepoResource_Apply(t *testing.T) {
    tests := []struct {
        name    string
        repo    *GitRepo
        setup   func(dir string)
        verify  func(dir string) error
        wantErr bool
    }{
        {
            name: "init new repo",
            repo: &GitRepo{Path: "test-repo"},
            verify: func(dir string) error {
                // Check .git exists
                if _, err := os.Stat(filepath.Join(dir, ".git")); err != nil {
                    return err
                }
                return nil
            },
        },
        // More test cases...
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Run test
        })
    }
}
```

## Performance Considerations

### Parallelization

Resources with no dependencies can execute in parallel:

```
Level 0: [homebrew]
         ↓
Level 1: [git, jq, ripgrep]  ← Can run in parallel
         ↓
Level 2: [repo-1, repo-2]    ← Can run in parallel
```

### Caching

Cache expensive checks:
- Package installation status
- Git remote existence
- Provider availability

Cache invalidation: per-execution (no persistent cache initially)

## Security Considerations

### Secrets Handling

1. **Never log secrets**
   - Sanitize logs and error messages
   - Mark fields as sensitive

2. **Secure storage**
   - Use OS keychain when possible
   - File permissions: 0600 for secret files
   - Environment variables in secure shell profiles

3. **Input validation**
   - Validate secret format before storage
   - Prevent injection attacks

### Command Execution

1. **No shell interpretation** by default
   - Use `exec.Command()` not `sh -c`
   - Prevent command injection

2. **Privilege escalation**
   - Prompt before `sudo`
   - Never store sudo passwords

3. **Safe defaults**
   - Don't follow symlinks outside config directory
   - Validate file paths

## Extensibility

### Plugin System (Future)

```go
// Plugin interface for custom resources
type ResourcePlugin interface {
    Type() string
    Create(config map[string]interface{}) (Resource, error)
}

// Register plugin
automata.RegisterPlugin(&MyCustomResource{})
```

### Custom Providers (Future)

```go
// Add custom package provider
automata.RegisterProvider(&MyPackageManager{})
```

## File Structure

```
automata/
├── main.go                    # CLI entry point
├── go.mod
├── go.sum
├── cmd/
│   ├── apply.go              # Apply command
│   ├── plan.go               # Plan command
│   ├── check.go              # Check command
│   └── validate.go           # Validate command
├── config/
│   ├── parser.go             # YAML/JSON parsing
│   ├── validator.go          # Config validation
│   └── types.go              # Config structs
├── resource/
│   ├── interface.go          # Resource interface
│   ├── registry.go           # Resource registry
│   ├── base.go               # Base implementation
│   └── types/
│       ├── git.go            # Git resource
│       ├── package.go        # Package resource
│       ├── secret.go         # Secret resource
│       └── file.go           # File resource
├── provider/
│   ├── interface.go          # Provider interface
│   ├── homebrew.go
│   ├── apt.go
│   ├── go.go
│   └── npm.go
├── executor/
│   ├── planner.go            # Dependency resolution
│   ├── runner.go             # Execution engine
│   └── graph.go              # Dependency graph
├── ui/
│   ├── prompt.go             # User prompts
│   ├── output.go             # Output formatting
│   └── progress.go           # Progress reporting
└── internal/
    ├── system/
    │   ├── filesystem.go     # File operations
    │   └── process.go        # Process execution
    └── util/
        ├── validation.go     # Common validators
        └── errors.go         # Error types
```

## Development Roadmap

### Phase 1: Foundation (Week 1-2)
- [ ] Project structure
- [ ] Resource interface
- [ ] Config parser (YAML)
- [ ] Basic executor (no dependencies)
- [ ] File resource (simplest)

### Phase 2: Core Resources (Week 3-4)
- [ ] Git resource
- [ ] Package resource
- [ ] Homebrew provider
- [ ] Dependency resolution

### Phase 3: User Interaction (Week 5-6)
- [ ] Secret resource
- [ ] Interactive prompts
- [ ] Progress reporting
- [ ] Better error messages

### Phase 4: Robustness (Week 7-8)
- [ ] Comprehensive tests
- [ ] Integration tests
- [ ] Error recovery
- [ ] Documentation

### Phase 5: Enhancement (Week 9-10)
- [ ] Additional providers (apt, go, npm)
- [ ] Parallel execution
- [ ] Caching
- [ ] Performance optimization

## Open Design Questions

### 1. Configuration Location

Where should automata look for config files?

**Options:**
- `./automata.yaml` (current directory)
- `~/.config/automata/config.yaml` (user config)
- `--config` flag
- All of the above with precedence?

### 2. Breaking Changes

What if changes would be destructive?

**Options:**
- Always prompt for confirmation
- `--force` flag to skip prompts
- `--dry-run` mode to preview
- Different confirmation levels (low, medium, high risk)

### 3. Partial Application

What if only some resources need updates?

**Options:**
- Apply all resources (slower but comprehensive)
- Apply only changed resources (faster)
- `--target` flag to apply specific resources

### 4. State Persistence

If we add state tracking later, where to store it?

**Options:**
- `.automata/state.json` in config directory
- `~/.local/share/automata/state.json`
- Alongside config file
- Remote state (S3, etc.) for teams

---

## Feedback Requested

Please provide feedback on:

1. **Architecture:** Any concerns with the proposed design?
2. **Resource Interface:** Too simple or too complex?
3. **Provider System:** Should we start with just Homebrew?
4. **State Management:** Stateless or stateful from the start?
5. **Testing:** Is the testing strategy sufficient?
6. **File Structure:** Any better organization?
7. **Open Questions:** Preferences for the questions above?
