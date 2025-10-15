#!/bin/bash
set -e

echo "Setting up development environment..."

# Install mise
curl https://mise.run | sh
echo 'eval "$(~/.local/bin/mise activate bash)"' >> ~/.bashrc

# Activate mise for this session
export PATH="$HOME/.local/bin:$PATH"
eval "$(~/.local/bin/mise activate bash)"

# Install tools from mise.toml
mise install

# Install Go dependencies for all Go projects
echo "Installing Go dependencies..."
for project in dissect markdown-format prrun printpdf claude-trace want; do
  if [ -d "$project" ] && [ -f "$project/go.mod" ]; then
    echo "  - $project"
    cd "$project" && go mod download && cd ..
  fi
done

# Install lib/ghrelease dependencies
if [ -d "lib/ghrelease" ] && [ -f "lib/ghrelease/go.mod" ]; then
  echo "  - lib/ghrelease"
  cd lib/ghrelease && go mod download && cd ../..
fi

# Install Node.js dependencies for diagram-dsl
if [ -d "diagram-dsl" ] && [ -f "diagram-dsl/package.json" ]; then
  echo "Installing Node.js dependencies for diagram-dsl..."
  cd diagram-dsl && npm install && cd ..
fi

# Build Rust project (mdbook-comments)
if [ -d "mdbook-comments" ] && [ -f "mdbook-comments/Cargo.toml" ]; then
  echo "Building mdbook-comments..."
  cd mdbook-comments && cargo build && cd ..
fi

echo "Development environment setup complete!"
