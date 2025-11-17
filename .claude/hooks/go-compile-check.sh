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
BUILD_OUTPUT=$(go build ./... 2>&1)
BUILD_EXIT=$?

if [ $BUILD_EXIT -ne 0 ]; then
  # Check if it's a network error
  if echo "$BUILD_OUTPUT" | grep -q "dial tcp\|connection refused\|lookup.*on.*53"; then
    echo "WARNING: Network error during Go build (DNS/connectivity issue). Skipping compilation check." >&2
    exit 0
  fi
  echo "ERROR: Go compilation failed. Fix compilation errors before stopping." >&2
  echo "$BUILD_OUTPUT" >&2
  exit 2
fi

# Try to build all Go tests
TEST_OUTPUT=$(go test -c ./... -o /dev/null 2>&1)
TEST_EXIT=$?

if [ $TEST_EXIT -ne 0 ]; then
  # Check if it's a network error
  if echo "$TEST_OUTPUT" | grep -q "dial tcp\|connection refused\|lookup.*on.*53"; then
    echo "WARNING: Network error during test compilation (DNS/connectivity issue). Skipping compilation check." >&2
    exit 0
  fi
  echo "ERROR: Go test compilation failed. Fix test compilation errors before stopping." >&2
  echo "$TEST_OUTPUT" >&2
  exit 2
fi

echo "Go compilation check passed!" >&2
exit 0
