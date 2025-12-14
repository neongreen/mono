# AGENTS.md — cloud

Guidelines for agents working in `cloud/` (OpenTofu + Kubernetes).

## Architecture Overview

**Infrastructure** (Spacelift `mono/cloud`):
- VMs, networks, firewalls, DNS
- Location: `cloud/terraform/`
- Auto-deploys on push to main via Spacelift

**Kubernetes** (Manual Deploy):
- Cluster system components and applications
- Location: `cloud/k8s/`
- Deploy via `./cloud/scripts/k8s-deploy.sh`

## Directory Structure

```
cloud/
├── terraform/           # Infrastructure (Spacelift: mono/cloud)
│   ├── main.tf         # VMs, networks, firewalls
│   ├── digitalocean-dns.tf  # DNS records
│   └── ...
│
├── k8s/                 # Kubernetes (manual deploy)
│   ├── system/          # Cluster essentials (deployed first)
│   │   ├── gvisor/      # Container isolation runtime
│   │   ├── ingress-nginx/     # Ingress controller
│   │   └── cert-manager/      # TLS certificates
│   │
│   └── apps/            # User applications
│       ├── postgres/    # Shared PostgreSQL database
│       ├── dagger/      # CI/CD engine
│       ├── n8n/         # Workflow automation
│       ├── onyx/        # AI platform
│       └── coder/       # Remote dev environments
│
├── scripts/             # Deployment scripts
│   ├── k8s-deploy.sh    # Main deploy script
│   ├── k8s-secrets.sh   # Create k8s secrets from fnox
│   ├── k8s-kubeconfig.sh # Fetch kubeconfig
│   └── k8s-kubectl.sh   # Kubectl wrapper
│
└── fnox.toml            # Secrets config (age-encrypted)
```

## Deployment Scripts

### Deploy Kubernetes Manifests

```bash
# Deploy everything (system + secrets + apps)
./cloud/scripts/k8s-deploy.sh

# Deploy only system components
./cloud/scripts/k8s-deploy.sh system

# Deploy only secrets
./cloud/scripts/k8s-deploy.sh secrets

# Deploy only apps
./cloud/scripts/k8s-deploy.sh apps

# Deploy specific app
./cloud/scripts/k8s-deploy.sh apps/dagger
```

### Node Labeling

The mono node must have the `workload=onyx` label for pods to schedule:
```bash
kubectl label node mono workload=onyx
```

### Fetch Kubeconfig

```bash
# Fetch and export kubeconfig
eval $(./cloud/scripts/k8s-kubeconfig.sh)
```

### Run Kubectl Commands

```bash
# Kubectl wrapper (auto-fetches kubeconfig)
./cloud/scripts/k8s-kubectl.sh get pods -A
./cloud/scripts/k8s-kubectl.sh get nodes
```

## Deployment Workflow

### Infrastructure Changes

1. Edit files in `cloud/terraform/`
2. Commit and push:
   ```bash
   mise x -- jj commit -m "Description of change"
   mise x -- jj bookmark set main -r @-
   mise x -- jj git push
   ```
3. Spacelift auto-deploys `mono/cloud` stack
4. Monitor: `mise x -- fnox exec -- spacectl stack run list --id mono-cloud-01`

### Kubernetes Changes

1. Edit manifests in `cloud/k8s/`
2. Deploy:
   ```bash
   ./cloud/scripts/k8s-deploy.sh
   ```
3. Verify:
   ```bash
   ./cloud/scripts/k8s-kubectl.sh get pods -A
   ```

## Adding a New Service

1. Create directory: `cloud/k8s/apps/my-service/`
2. Add manifests:
   - `namespace.yaml`
   - `deployment.yaml`
   - `service.yaml`
   - `ingress.yaml` (if externally accessible)
   - `network-policy.yaml` (allow rules)
3. Deploy: `./cloud/scripts/k8s-deploy.sh apps/my-service`

Example ingress for `my-service.cloud.artyom.me`:
```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: my-service
  namespace: my-service
  annotations:
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
spec:
  ingressClassName: nginx
  tls:
    - hosts: [my-service.cloud.artyom.me]
      secretName: my-service-tls
  rules:
    - host: my-service.cloud.artyom.me
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: my-service
                port:
                  number: 8080
```

## Secrets Management

Secrets are age-encrypted in `cloud/fnox.toml` and created in k8s via script:

```bash
# Deploy secrets to cluster
./cloud/scripts/k8s-secrets.sh
```

The script creates secrets in each namespace:
- `postgres/postgres-secret` - Shared postgres password
- `n8n/postgres-secret` - Copy for n8n access
- `coder/coder-secrets` - Postgres password
- `onyx/onyx-secrets` - Postgres password, session secret, minio password

### Adding New Secrets

1. Generate a random password:
   ```bash
   openssl rand -base64 32
   ```

2. Encrypt with age and add to `cloud/fnox.toml`:
   ```bash
   echo "your-secret" | age -r age1q8f03l7hmyrk3wazqddc40ucc7rtvugl527pp9lxahvl8rar5p2q2plg56 | base64
   ```

3. Add to fnox.toml:
   ```toml
   MY_SECRET = { provider = "age", value = "YWdlLWVu..." }
   ```

4. Update `k8s-secrets.sh` to create the k8s secret

## Tools & Execution (fnox)

All tools run via mise, secrets via fnox:
```bash
# Spacelift CLI
mise x -- fnox exec -- spacectl whoami
mise x -- fnox exec -- spacectl stack list

# Local tofu (validation only)
cd cloud/terraform
mise x -- fnox exec -- tofu init -backend=false -input=false
mise x -- fnox exec -- tofu validate
```

## Hetzner Cloud Best Practices

- Provider: `hetznercloud/hcloud` (token from `HCLOUD_TOKEN`)
- **Use `.name` not `.id`** for immutable attributes (location, server_type, image)
  - `.id` returns numeric IDs causing spurious replacements
- Resources:
  - Network: `hcloud_network`, `hcloud_network_subnet`
  - Security: `hcloud_firewall`, `hcloud_firewall_attachment`
  - SSH: `hcloud_ssh_key`
  - Compute: `hcloud_server`, `hcloud_server_network`

## DANGER ZONES

### SSH Keys in Terraform

**⚠️ DO NOT MODIFY `ssh_keys` on `hcloud_server` ⚠️**

Changing the `ssh_keys` attribute on a Hetzner server **DESTROYS AND RECREATES THE ENTIRE SERVER**. This will:
- Delete all data on the server
- Destroy the k3s cluster
- Require full re-provisioning

To add SSH keys to an existing server:
1. SSH into the server: `ssh emily@mono.cloud.artyom.me`
2. Edit `~/.ssh/authorized_keys`
3. Add the new public key

The `ssh_keys` in terraform are ONLY used for initial server provisioning.

### Manual Terraform

**Never run `tofu apply` locally** - always let Spacelift handle it. Local applies can cause state drift and conflicts.

## Spacelift Operations

```bash
# List stacks
mise x -- fnox exec -- spacectl stack list

# List runs
mise x -- fnox exec -- spacectl stack run list --id mono-cloud-01

# View logs
mise x -- fnox exec -- spacectl stack logs --id mono-cloud-01 --run <run-id>

# Check environment variables
mise x -- fnox exec -- spacectl stack environment list --id mono-cloud-01
```
