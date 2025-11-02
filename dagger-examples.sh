#!/bin/bash
# Example Dagger commands to run from Claude Code web
# Assumes you've set _EXPERIMENTAL_DAGGER_RUNNER_HOST

set -e

# Ensure Dagger is in PATH
export PATH="$HOME/.local/bin:/.local/bin:$PATH"

# Check if remote runner is configured
if [ -z "$_EXPERIMENTAL_DAGGER_RUNNER_HOST" ]; then
    echo "❌ Please set _EXPERIMENTAL_DAGGER_RUNNER_HOST first!"
    echo "   Example: export _EXPERIMENTAL_DAGGER_RUNNER_HOST=tcp://your-server.example.com:1234"
    exit 1
fi

echo "🔗 Using remote runner: $_EXPERIMENTAL_DAGGER_RUNNER_HOST"
echo ""

# Example 1: Simple container execution
echo "📦 Example 1: Run a simple command in Alpine"
dagger query <<'EOF'
{
  container {
    from(address: "alpine:latest") {
      withExec(args: ["echo", "Hello from Dagger running in Claude Code web!"]) {
        stdout
      }
    }
  }
}
EOF

echo ""
echo "---"
echo ""

# Example 2: Multi-step build
echo "🛠️  Example 2: Multi-step build process"
dagger query <<'EOF'
{
  container {
    from(address: "golang:1.21-alpine") {
      withExec(args: ["go", "version"]) {
        stdout
      }
    }
  }
}
EOF

echo ""
echo "---"
echo ""

# Example 3: Using Dagger Functions (if you have a dagger.json in your project)
if [ -f "dagger.json" ]; then
    echo "🎯 Example 3: Using project Dagger functions"
    dagger functions
    echo ""
    echo "To call a function:"
    echo "   dagger call <function-name> --help"
else
    echo "💡 Tip: Initialize a Dagger module in your project with:"
    echo "   dagger init"
    echo ""
    echo "   Then you can define custom functions and call them with:"
    echo "   dagger call <function-name>"
fi

echo ""
echo "✅ All examples completed!"
echo ""
echo "📚 More resources:"
echo "   - Dagger quickstart: https://docs.dagger.io/quickstart"
echo "   - GraphQL API: https://docs.dagger.io/api/reference"
echo "   - SDK docs: https://docs.dagger.io/sdk"
