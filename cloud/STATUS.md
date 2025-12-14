# Cloud Infrastructure Status

**Last Updated:** 2025-12-14

## Infrastructure

### Server
- **mono**: CX53 (4 vCPU, 8GB RAM) in Helsinki (hel1)
- **OS**: Ubuntu 24.04 LTS
- **Kubernetes**: k3s v1.33.6+k3s1
- **Container Isolation**: gVisor (runsc)
- **Ingress**: nginx-ingress (traefik disabled)
- **Certificates**: cert-manager with Let's Encrypt

### Networking
- Private network: 10.0.0.0/16 (mono-net)
- Firewall: SSH (22), HTTP (80), HTTPS (443), k3s API (6443), kubelet (10250)
- DNS: *.cloud.artyom.me → mono server IP

## Deployed Applications

| Namespace | App | URL | Status |
|-----------|-----|-----|--------|
| postgres | PostgreSQL 16 | - | Shared database |
| n8n | n8n | https://n8n.cloud.artyom.me | Workflow automation |
| coder | Coder | https://coder.cloud.artyom.me | Remote dev environments |
| onyx | Onyx AI | https://onyx.cloud.artyom.me | AI platform |
| dagger | Dagger Engine | - | CI/CD engine (internal) |

### System Services
- **ingress-nginx**: Ingress controller (hostNetwork mode)
- **cert-manager**: TLS certificate management
- **gvisor**: Container isolation RuntimeClass

## Deployment

### Infrastructure (Terraform)
Auto-deploys via Spacelift on push to main.

### Kubernetes (Manual)
```bash
./cloud/scripts/k8s-deploy.sh           # Deploy everything
./cloud/scripts/k8s-deploy.sh system    # System components only
./cloud/scripts/k8s-deploy.sh apps      # Apps only
./cloud/scripts/k8s-deploy.sh secrets   # Secrets only
```

## Node Labels

Required labels on the mono node:
```bash
kubectl label node mono workload=onyx
```

## Secrets

Secrets are age-encrypted in `cloud/fnox.toml`. Deploy with:
```bash
./cloud/scripts/k8s-secrets.sh
```

## Architecture Decisions

### gVisor Usage
- **Enabled**: n8n, onyx services (api-server, background, web-server, mcp-server)
- **Disabled**: PostgreSQL (performance), Dagger Engine (privileged access required)

### Network Security
- Default-deny policies in each namespace
- Explicit allow rules for required connections
- External internet access via ipBlock (not namespaceSelector)

### Deployment Model
- **Infrastructure**: Declarative (OpenTofu) → Auto-deploy via Spacelift
- **Kubernetes**: Declarative (manifests) → Manual deploy via scripts
- **Secrets**: Age-encrypted in git → Created via k8s-secrets.sh
