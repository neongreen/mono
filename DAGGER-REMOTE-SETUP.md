# Running Dagger in Claude Code Web with Remote Runner

## TL;DR

**YES, Dagger works in Claude Code web!** You just need a remote Dagger runner.

The Dagger CLI is a thin client that doesn't need Docker locally - it connects to a Dagger Engine running elsewhere.

## Architecture

```
┌─────────────────────────┐         ┌──────────────────────────┐
│  Claude Code Web        │         │  Your Server/Machine     │
│                         │         │                          │
│  ┌─────────────────┐    │  TCP    │  ┌──────────────────┐   │
│  │  Dagger CLI     │────┼────────▶│  │  Dagger Engine   │   │
│  │  (no Docker)    │    │         │  │  (needs Docker)  │   │
│  └─────────────────┘    │         │  └──────────────────┘   │
└─────────────────────────┘         └──────────────────────────┘
```

## Quick Start

### 1. Set Up Remote Runner (on your machine with Docker)

```bash
# Run this on your local machine or a server with Docker
./remote-runner-setup.sh

# Or manually:
docker run -d \
  --name dagger-engine \
  --privileged \
  -p 1234:1234 \
  -v dagger-engine:/var/lib/dagger \
  registry.dagger.io/engine:v0.19.4 \
  --listen tcp://0.0.0.0:1234
```

### 2. Set Up Dagger CLI (in Claude Code web)

```bash
# Run this in your Claude Code web terminal
./setup-dagger-remote.sh

# Or manually:
curl -fsSL https://dl.dagger.io/dagger/install.sh | BIN_DIR=$HOME/.local/bin sh
export PATH="$HOME/.local/bin:/.local/bin:$PATH"
```

### 3. Connect to Remote Runner

```bash
# Set the remote runner address
export _EXPERIMENTAL_DAGGER_RUNNER_HOST=tcp://your-server.example.com:1234

# Test it!
dagger query '{container{from(address:"alpine:latest"){withExec(args:["echo","hello from remote!"]){stdout}}}}'
```

### 4. Run Examples

```bash
./dagger-examples.sh
```

## Security Setup (IMPORTANT!)

⚠️ **Dagger does NOT encrypt traffic by default!** Choose one of these security options:

### Option A: SSH Tunnel (Recommended for most users)

```bash
# On Claude Code web, create an SSH tunnel
ssh -L 8080:localhost:1234 your-user@your-server.example.com -N -f

# Use localhost
export _EXPERIMENTAL_DAGGER_RUNNER_HOST=tcp://localhost:8080
```

### Option B: VPN/Tailscale

```bash
# Install Tailscale on both machines
# Connect them to the same tailnet
# Use the Tailscale IP
export _EXPERIMENTAL_DAGGER_RUNNER_HOST=tcp://100.x.x.x:1234
```

### Option C: TLS Proxy (Production)

Use nginx or Caddy with TLS in front of the Dagger engine.

## Use Cases

### Use Case 1: CI/CD Testing Before Push

```bash
# Test your CI/CD pipeline locally before pushing
dagger call build
dagger call test
dagger call deploy --dry-run
```

### Use Case 2: Consistent Build Environments

```bash
# Same build that runs in CI, but from Claude Code web
export _EXPERIMENTAL_DAGGER_RUNNER_HOST=tcp://your-runner:1234
dagger call build-production
```

### Use Case 3: Multi-Project Workflows

```bash
# Use Dagger to orchestrate multiple services
dagger call deploy-stack \
  --frontend=./frontend \
  --backend=./backend \
  --database=postgres:15
```

## Working with Your Project

If your project has a `dagger.json` (Dagger module):

```bash
# List available functions
dagger functions

# Get help for a specific function
dagger call build --help

# Call a function
dagger call build --args="--verbose"

# Chain functions
dagger call build | dagger call test | dagger call deploy
```

If your project doesn't have Dagger yet:

```bash
# Initialize a new Dagger module
dagger init

# Choose your SDK (TypeScript, Python, Go)
dagger develop
```

## Troubleshooting

### "Connection refused"

- Check if the remote runner is running: `docker ps | grep dagger`
- Check if the port is open: `curl -v telnet://your-server:1234`
- Verify firewall rules on the remote server

### "Driver for scheme 'image' was not available"

- You forgot to set `_EXPERIMENTAL_DAGGER_RUNNER_HOST`
- Dagger is trying to use local Docker (which doesn't exist in web environment)

### "Unauthorized" or authentication errors

- The Dagger protocol doesn't have built-in auth
- You must secure the connection at the network level (SSH, VPN, etc.)

### Network timeout from Claude Code web

- Claude Code web has egress filtering
- SSH tunnel through an approved domain should work
- Or use a VPN/Tailscale connection

## Comparison with Other Approaches

| Approach | Works in Web? | Complexity | Security | Performance |
|----------|---------------|------------|----------|-------------|
| **Remote Runner** ✅ | **Yes** | Medium | DIY | Good |
| Local Desktop + Docker | No | Low | Local | Best |
| DevContainer | No | Medium | Local | Good |
| MCP Server (custom) | Maybe | High | DIY | Good |

## Advantages of This Approach

1. ✅ **True "everywhere" Dagger** - Same workflows in web, local, and CI
2. ✅ **No local Docker needed** - Works in restricted environments
3. ✅ **Shared compute** - Team can share a powerful runner
4. ✅ **Centralized caching** - Dagger cache shared across team
5. ✅ **Easy to set up** - Just two scripts

## Limitations

1. ⚠️ **Network dependency** - Need connectivity to runner
2. ⚠️ **Security setup required** - Must secure the connection yourself
3. ⚠️ **Experimental flag** - `_EXPERIMENTAL_DAGGER_RUNNER_HOST` may change
4. ⚠️ **Session-based** - Need to set env var each Claude Code session

## Next Steps

1. **Try it now**: Run `./setup-dagger-remote.sh` and connect to your runner
2. **Initialize Dagger in your project**: `dagger init`
3. **Create Dagger functions**: See https://docs.dagger.io/quickstart
4. **Set up secure access**: Use SSH tunnel or VPN
5. **Share with team**: Document the runner address for others

## Resources

- [Dagger Documentation](https://docs.dagger.io)
- [Custom Runner Guide](https://docs.dagger.io/manuals/administrator/custom-runner/)
- [Dagger Cloud](https://dagger.io/cloud) - Adds visualization and distributed caching
- [Dagger Modules](https://daggerverse.dev) - Pre-built Dagger functions

## Files in This Setup

- `setup-dagger-remote.sh` - Run in Claude Code web to install Dagger CLI
- `remote-runner-setup.sh` - Run on your server to start Dagger Engine
- `dagger-examples.sh` - Example commands to test the setup
- `DAGGER-REMOTE-SETUP.md` - This documentation

---

**You were right!** Dagger absolutely can work in Claude Code web with a remote runner. No Docker needed locally, no MCP servers needed, just a simple TCP connection to a remote Dagger Engine. 🎉
