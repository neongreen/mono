#!/bin/bash

# Only block TodoWrite when running in Claude Code CLI, not Claude web
# Check CLAUDE_CODE_REMOTE environment variable
# When CLAUDE_CODE_REMOTE=true, we're in Claude web (remote environment)
if [ "$CLAUDE_CODE_REMOTE" = "true" ]; then
  # Running in Claude web - allow TodoWrite
  exit 0
fi

# Running in Claude Code CLI - block TodoWrite and redirect to tk
echo "Usage of TodoWrite is forbidden. All tasks must be managed via the 'tk' command-line tool as specified in CLAUDE.md." >&2
exit 2
