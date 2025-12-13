# Claude rules

## Tooling

All tools are managed via **mise** and should be run through it:
```bash
mise x -- <tool> <args>
```

This ensures consistent tool versions across environments.

## Version Control

We use **jj** (Jujutsu), not git. Run via mise:
```bash
mise x -- jj status              # see working copy status
mise x -- jj log                 # view history
mise x -- jj commit -m "msg"     # create commit from working copy (always use this for new commits)
mise x -- jj describe -m "msg"   # amend description of existing commit (not for new commits)
mise x -- jj squash              # squash into the parent change
```

**Important**: Always use `jj commit` to create new commits from the working copy. Don't use `jj edit`.

## Secrets Management

**Legacy: fnox** - Some secrets are stored in fnox. These may not be available on all machines (especially remote environments). Use `mise x -- fnox exec -- <command>` to run commands with secrets injected.

**New: age encryption** - New secrets are stored encrypted with age and can be committed to the repo. See `cloud/fnox.toml` for the pattern:
```toml
[providers.age]
type = "age"
recipients = ["age1..."]

[secrets]
MY_SECRET = { provider = "age", value = "YWdlLWVu..." }
```

Age-encrypted values are safe to commit since only the private key holder can decrypt them.

## Public Repository

This is a **public repo** - never commit plaintext secrets. Use age encryption for any secrets that need to be in the repo.

## Infrastructure (`cloud/`)

All infrastructure is **fully declarative** - never make manual changes to cloud resources. Everything is defined in code and deployed via automation.

See `cloud/AGENTS.md` for detailed guidelines. Key points:
- **OpenTofu + Spacelift** for infrastructure (auto-deploys on git push)
- **kubectl** for Kubernetes apps (manual deployment via `mise run cloud:kubectl`)
- Use `fnox exec -- spacectl ...` to interact with Spacelift stacks

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

**The issue:** `storage.googleapis.com` (where Go stores module zip files) is in the `no_proxy` environment variable. This prevents Go from using the HTTP proxy for DNS resolution, causing failures like:
```
dial tcp: lookup storage.googleapis.com on 8.8.8.8:53: i/o timeout
```

**The fix:** The SessionStart hook (`scripts/install-mise-remote.sh`) automatically unsets `no_proxy` and `NO_PROXY` environment variables. This allows Go to use the HTTP proxy for all connections, including DNS resolution.

**What this means:**
- `go build`, `go mod download`, etc. work without any special configuration
- All Go commands automatically use the proxy
- Direct DNS queries (UDP port 53) don't work, but proxy-based resolution does
- This fix is applied automatically when you start a new Claude Code session

If you encounter Go build issues in the remote environment, verify that `no_proxy` is unset:
```bash
env | grep -i no_proxy  # Should return nothing
```

## Cloud Infrastructure Reference

### Directory Structure

```
cloud/
├── terraform/           # Infrastructure (Spacelift: mono/cloud)
│   ├── main.tf         # Hetzner Cloud resources (VMs, networks, firewalls)
│   ├── providers.tf    # Provider configs (hcloud, digitalocean)
│   ├── digitalocean-dns.tf  # DNS records
│   └── ssh_key*.pub    # SSH keys
│
├── k8s/                 # Kubernetes (Spacelift: mono/k8s)
│   ├── system/         # Cluster essentials (deployed first)
│   │   ├── gvisor/     # Container isolation runtime
│   │   ├── network-policies/  # Default-deny policies
│   │   ├── ingress-nginx/     # Ingress controller
│   │   └── cert-manager/      # TLS certificates
│   │
│   └── apps/           # User applications
│       ├── dagger/     # CI/CD engine
│       ├── n8n/        # Workflow automation
│       ├── onyx/       # AI platform
│       ├── coder/      # Remote dev environments
│       └── sample-apps/
│
├── coder-templates/     # Coder workspace templates
├── fnox.toml           # Age-encrypted secrets
├── AGENTS.md           # Agent guidelines for cloud work
└── STATUS.md           # Current infrastructure status
```

### Providers

- **Hetzner Cloud**: VMs, networks, firewalls (token via `HCLOUD_TOKEN`)
- **DigitalOcean**: DNS management (token via `DIGITALOCEAN_TOKEN`)
- **Spacelift**: Infrastructure deployment automation (`spacectl` CLI)

### Current Infrastructure

**Single server architecture** (consolidated from multiple VMs):
- **mono**: CX53 (4 vCPU, 8GB RAM) in Helsinki (hel1)
  - Ubuntu 24.04 LTS
  - k3s Kubernetes cluster
  - gVisor for container isolation

**Networking**:
- Private network: `10.0.0.0/16` (mono-net)
- Firewall: SSH (22), HTTP (80), HTTPS (443), k3s API (6443), kubelet (10250)
- DNS: `*.cloud.artyom.me` → mono server IP (wildcard A record)

### Kubernetes Applications

| Namespace | App | Purpose | gVisor |
|-----------|-----|---------|--------|
| dagger | Dagger Engine | CI/CD pipelines | No (needs privileged) |
| n8n | n8n + PostgreSQL | Workflow automation | n8n: Yes, Postgres: No |
| onyx | Onyx AI platform | Self-hosted AI | Most services: Yes |
| coder | Coder | Remote dev environments | Workspaces: Yes |
| sample-apps | nginx, httpbin | Testing | Yes |

### Deployment

**Both stacks auto-deploy on push to main.**

```bash
# Commit and push (triggers Spacelift)
mise x -- jj commit -m "Add new service"
mise x -- jj git push

# Watch Spacelift deployments
mise run cloud:watch

# Check Spacelift stacks
mise x -- fnox exec -- spacectl stack list
```

**Manual deployment (if needed):**
```bash
# Get kubeconfig
mise run cloud:kubeconfig

# Deploy system components
mise run cloud:kubectl apply -k k8s/system/

# Deploy specific app
mise run cloud:kubectl apply -f k8s/apps/dagger/

# Validate terraform locally
mise run cloud:validate
```
