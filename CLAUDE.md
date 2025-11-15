# Claude rules

## Task tracking

Use `tk` for task tracking.

- Create tasks for all work you do
- Always keep status up to date
- Break big tasks into subtasks
- Search for related tasks and mark them as related
- Add notes to tasks as you go
- In commit messages mention which tasks they are related to

There are two versions:
- `tk` is the globally installed binary
- `tk-dev` is an alias that automatically builds and runs tk from the local checkout

**Important:** `tk-dev` Just Works™ - you don't need to run `go build` or anything. Just use `tk-dev` directly as a command, and it will build from source if needed.

When working on `tk` itself, use `tk-dev` to test your changes. Use `tk` for normal task tracking.

## Remote environment (Claude Code on the web)

### Go builds and dependencies

The remote environment has internet access through an HTTP proxy. However, there's a networking quirk that affects Go module downloads:

**The issue:** `storage.googleapis.com` (where Go stores module zip files) is in the `GLOBAL_AGENT_NO_PROXY` environment variable. This prevents Go from using the HTTP proxy for DNS resolution, causing failures like:
```
dial tcp: lookup storage.googleapis.com on [::1]:53: read udp [::1]:xxxxx->[::1]:53: read: connection refused
```

**The fix:** The SessionStart hook (`scripts/install-mise-remote.sh`) automatically unsets `no_proxy`, `NO_PROXY`, and `GLOBAL_AGENT_NO_PROXY` environment variables. This allows Go to use the HTTP proxy for all connections, including DNS resolution.

**What this means:**
- `go build`, `go mod download`, etc. work without any special configuration
- All Go commands automatically use the proxy
- Direct DNS queries (UDP port 53) don't work, but proxy-based resolution does
- This fix is applied automatically when you start a new Claude Code session

If you encounter Go build issues in the remote environment, verify that no_proxy variables are unset:
```bash
env | grep -i no_proxy  # Should return nothing
```
