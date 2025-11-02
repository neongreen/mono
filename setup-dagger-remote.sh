#!/bin/bash
# Setup script for using Dagger in Claude Code web with remote runner

set -e

echo "🔧 Setting up Dagger CLI in Claude Code web..."

# Install Dagger CLI if not already installed
if [ ! -f "$HOME/.local/bin/dagger" ]; then
    echo "📥 Installing Dagger CLI..."
    curl -fsSL https://dl.dagger.io/dagger/install.sh | BIN_DIR=$HOME/.local/bin sh
    chmod +x $HOME/.local/bin/dagger
else
    echo "✅ Dagger CLI already installed"
fi

# Add to PATH for current session
export PATH="$HOME/.local/bin:/.local/bin:$PATH"

# Display version
echo "📌 Dagger version:"
dagger version

echo ""
echo "🎯 Next steps:"
echo ""
echo "1. Set up a remote Dagger runner on a machine with Docker:"
echo "   docker run -d --name dagger-engine --privileged -p 1234:1234 \\"
echo "     -v dagger-engine:/var/lib/dagger \\"
echo "     registry.dagger.io/engine:v0.19.4 --listen tcp://0.0.0.0:1234"
echo ""
echo "2. Set the runner host in your session:"
echo "   export _EXPERIMENTAL_DAGGER_RUNNER_HOST=tcp://your-server.example.com:1234"
echo ""
echo "3. Test the connection:"
echo "   dagger query '{container{from(address:\"alpine:latest\"){id}}}'"
echo ""
echo "⚠️  SECURITY: Remember to secure the connection with SSH tunnel, VPN, or TLS proxy!"
