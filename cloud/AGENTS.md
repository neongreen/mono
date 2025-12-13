# AGENTS.md — cloud

Guidelines for agents working in `cloud/` (OpenTofu + Spacelift) with `fnox`.

## Architecture Overview

Two Spacelift stacks, both auto-deploy on git push:

**Stack 1: `mono/cloud` (Infrastructure)**
- What: VMs, networks, firewalls, DNS
- Where: `cloud/terraform/`
- How: OpenTofu via Spacelift

**Stack 2: `mono/k8s` (Kubernetes)**
- What: Cluster system components and applications
- Where: `cloud/k8s/`
- How: `kubectl apply` via Spacelift
- Depends on: `mono/cloud` (waits for infrastructure)

Both stacks trigger on push to main. The k8s stack waits for the cloud stack to complete.

## Directory Structure

```
cloud/
├── terraform/           # Infrastructure (Spacelift: mono/cloud)
│   ├── main.tf         # VMs, networks, firewalls
│   ├── digitalocean-dns.tf  # DNS records
│   └── ...
│
└── k8s/                 # Kubernetes (Spacelift: mono/k8s)
    ├── system/          # Cluster essentials (deployed first)
    │   ├── gvisor/      # Container isolation runtime
    │   ├── network-policies/  # Default-deny policies
    │   ├── ingress-nginx/     # Ingress controller
    │   └── cert-manager/      # TLS certificates
    │
    └── apps/            # User applications
        ├── dagger/      # CI/CD engine
        ├── n8n/         # Workflow automation
        ├── onyx/        # AI platform
        ├── coder/       # Remote dev environments
        └── sample-apps/ # Test apps
```

## Deployment Workflow

**Everything auto-deploys on push:**

1. Edit files in `cloud/terraform/` or `cloud/k8s/`
2. Commit and push:
   ```bash
   mise x -- jj commit -m "Add new service"
   mise x -- jj git push
   ```
3. Spacelift auto-deploys:
   - First: `mono/cloud` (if terraform changed)
   - Then: `mono/k8s` (if k8s manifests changed)
4. Watch progress: `mise run cloud:watch`

**Manual deployment (if needed):**
```bash
# System components
mise run cloud:kubectl apply -k k8s/system/

# Specific app
mise run cloud:kubectl apply -f k8s/apps/dagger/
```

## Adding a New Service

1. Create directory: `cloud/k8s/apps/my-service/`
2. Add manifests:
   - `namespace.yaml`
   - `deployment.yaml`
   - `service.yaml`
   - `ingress.yaml` (if externally accessible)
   - `network-policy.yaml` (allow rules)
3. Push to git
4. Spacelift deploys automatically

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

## Adding a New VM (Worker Node)

Edit `cloud/terraform/main.tf`:

```hcl
# Add new server resource
resource "hcloud_server" "worker1" {
  name        = "worker1"
  image       = data.hcloud_image.ubuntu_2404.name
  server_type = "cx42"  # or whatever size
  location    = data.hcloud_location.hel1.name
  ssh_keys    = [hcloud_ssh_key.workstation.id, hcloud_ssh_key.laptop.id]

  user_data = <<-CLOUDCFG
  #cloud-config
  hostname: worker1
  # ... (copy cloud-init from mono, change K3S_URL for agent mode)
  runcmd:
    - curl -sfL https://get.k3s.io | K3S_URL=https://${hcloud_server.mono.ipv4_address}:6443 K3S_TOKEN=<token> sh -
  CLOUDCFG
}
```

Push to git, Spacelift creates the VM, it auto-joins the k3s cluster.

## Moving Workloads Between Nodes

Use node selectors in deployments:

```yaml
spec:
  template:
    spec:
      nodeSelector:
        kubernetes.io/hostname: worker1  # or use labels
```

To move a workload:
1. Change the nodeSelector in the manifest
2. Push to git
3. Pods reschedule to the new node

## Secrets & Execution (fnox)

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

## Setting Up the Spacelift k8s Stack

The `mono/k8s` stack needs to be created in the Spacelift UI:

1. Create stack at https://neongreen.app.spacelift.io
2. Settings:
   - Name: `mono/k8s`
   - Repository: `neongreen/mono`
   - Project root: `cloud/k8s`
   - Runner image: one with kubectl installed
   - Before init: fetch kubeconfig from mono server
   - Commands: `kubectl apply -k system/ && kubectl apply -f apps/`
3. Add dependency on `mono/cloud` stack
4. Attach context with `KUBECONFIG` or SSH access to fetch it

Alternative: The stack can run `kubectl` by SSHing to the mono server where kubeconfig is already present.
