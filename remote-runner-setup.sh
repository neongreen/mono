#!/bin/bash
# Run this script on your remote machine (with Docker installed)
# This sets up a Dagger Engine that Claude Code web can connect to

set -e

echo "🚀 Setting up Dagger Engine as remote runner..."
echo ""

# Check if Docker is available
if ! command -v docker &> /dev/null; then
    echo "❌ Docker not found. Please install Docker first."
    exit 1
fi

echo "✅ Docker is installed"

# Configuration
RUNNER_PORT=${DAGGER_RUNNER_PORT:-1234}
RUNNER_IMAGE="registry.dagger.io/engine:v0.19.4"

echo "📥 Pulling Dagger engine image..."
docker pull "$RUNNER_IMAGE"

# Check if container already exists
if docker ps -a --format '{{.Names}}' | grep -q '^dagger-engine$'; then
    echo "⚠️  Container 'dagger-engine' already exists. Removing..."
    docker rm -f dagger-engine
fi

echo "🏃 Starting Dagger engine on port $RUNNER_PORT..."
docker run -d \
  --name dagger-engine \
  --restart unless-stopped \
  --privileged \
  -p "${RUNNER_PORT}:${RUNNER_PORT}" \
  -v dagger-engine:/var/lib/dagger \
  "$RUNNER_IMAGE" \
  --listen "tcp://0.0.0.0:${RUNNER_PORT}"

echo ""
echo "✅ Dagger engine is running!"
echo ""
echo "📊 Container status:"
docker ps --filter name=dagger-engine --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
echo ""
echo "🔍 To check logs:"
echo "   docker logs -f dagger-engine"
echo ""
echo "🌐 Connection string for Claude Code web:"
echo "   export _EXPERIMENTAL_DAGGER_RUNNER_HOST=tcp://$(hostname -I | awk '{print $1}'):${RUNNER_PORT}"
echo ""
echo "⚠️  SECURITY REMINDERS:"
echo "   - This setup does NOT use encryption"
echo "   - Consider using SSH tunnel: ssh -L 8080:localhost:${RUNNER_PORT} $(whoami)@$(hostname)"
echo "   - Or set up a VPN/Tailscale for secure access"
echo "   - Don't expose port ${RUNNER_PORT} directly to the internet!"
echo ""
echo "🔒 For SSH tunnel from Claude Code web (if you have SSH access):"
echo "   1. In Claude Code web terminal: ssh -L 8080:localhost:${RUNNER_PORT} your-user@$(hostname -I | awk '{print $1}') -N -f"
echo "   2. Then use: export _EXPERIMENTAL_DAGGER_RUNNER_HOST=tcp://localhost:8080"
