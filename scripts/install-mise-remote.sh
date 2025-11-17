#!/bin/bash
set -e

# Only run in remote (Claude Code on the web) environments
if [ "$CLAUDE_CODE_REMOTE" != "true" ]; then
  echo "Skipping mise installation (not in remote environment)"
  exit 0
fi

echo "Installing mise for Claude Code on the web..."

# Check if mise is already installed
if command -v mise &> /dev/null; then
  echo "mise is already installed: $(mise --version)"
  mise trust
  exit 0
fi

# Download and install mise
curl https://mise.jdx.dev/install.sh | sh

# Fix Go module downloads by removing no_proxy restrictions
# The environment has an HTTP proxy that handles DNS, but *.googleapis.com in no_proxy
# prevents Go from using it, causing DNS resolution failures.
if [ -n "$CLAUDE_ENV_FILE" ]; then
  echo 'unset no_proxy' >> "$CLAUDE_ENV_FILE"
  echo 'unset NO_PROXY' >> "$CLAUDE_ENV_FILE"
fi
unset no_proxy
unset NO_PROXY

# Add mise to PATH for this session via CLAUDE_ENV_FILE
if [ -n "$CLAUDE_ENV_FILE" ]; then
  echo 'export PATH="$HOME/.local/bin:$PATH"' >> "$CLAUDE_ENV_FILE"
fi

# Also add to .bashrc for future shells
if [ -f "$HOME/.bashrc" ]; then
  if ! grep -q '.local/bin' "$HOME/.bashrc"; then
    echo 'export PATH="$HOME/.local/bin:$PATH"' >> "$HOME/.bashrc"
  fi
fi

# Source the path for immediate use
export PATH="$HOME/.local/bin:$PATH"

# Verify installation
if command -v mise &> /dev/null; then
  echo "mise successfully installed: $(mise --version)"
  
  # Trust the repository's mise configuration
  cd "$CLAUDE_PROJECT_DIR"
  mise trust
  
  echo "mise installation complete and repository trusted"
else
  echo "Error: mise installation failed"
  exit 1
fi

exit 0

