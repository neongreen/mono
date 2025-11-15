#!/bin/bash
# Hook to compile all Go code and tests on Stop
# Only runs in Claude Code web environment

# Only run in Claude Code web (remote environment)
# CLAUDE_CODE_REMOTE=true means we're in the web version
if [ "$CLAUDE_CODE_REMOTE" != "true" ]; then
  # Running in local Claude Code CLI - skip this check
  exit 0
fi

echo "Checking Go compilation..." >&2

# Find all Go modules in the repository
cd "$CLAUDE_PROJECT_DIR" || exit 1

# Try to build all Go packages
if ! go build ./...; then
  echo "ERROR: Go compilation failed. Fix compilation errors before stopping." >&2
  exit 2
fi

# Try to build all Go tests
if ! go test -c ./... -o /dev/null; then
  echo "ERROR: Go test compilation failed. Fix test compilation errors before stopping." >&2
  exit 2
fi

echo "Go compilation check passed!" >&2
exit 0
