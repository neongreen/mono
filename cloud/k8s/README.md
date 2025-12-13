# Kubernetes Manifests

This directory contains Kubernetes manifests for deploying to the k3s cluster.

## Structure

```
k8s/
├── system/              # Cluster infrastructure (deployed first)
│   ├── gvisor/          # RuntimeClass for container isolation
│   ├── network-policies/ # Default-deny policies
│   ├── ingress-nginx/   # Ingress controller
│   └── cert-manager/    # TLS certificate automation
│
└── apps/                # User applications
    ├── dagger/          # CI/CD engine
    ├── n8n/             # Workflow automation
    ├── onyx/            # AI platform
    ├── coder/           # Remote dev environments
    └── sample-apps/     # Test apps
```

## Deployment

**Kubernetes manifests auto-deploy via Spacelift** when you push to main.

Spacelift has two stacks:
1. `mono/cloud` - Terraform (VMs, networks, DNS)
2. `mono/k8s` - Kubernetes manifests (this directory)

The k8s stack depends on the cloud stack, so infrastructure is always ready before apps deploy.

### Deployment Order

Spacelift applies manifests in this order:
1. `system/` - Cluster essentials (gvisor, ingress, cert-manager)
2. `apps/` - Applications

### Manual Deployment (if needed)

```bash
# Deploy everything
mise run cloud:k8s:deploy

# Deploy specific app
mise run cloud:kubectl apply -f k8s/apps/dagger/
```

## Security

### gVisor Runtime

Container isolation via gVisor:
- Syscalls intercepted in userspace
- Use `runtimeClassName: gvisor` in pod specs
- ~10-20% performance overhead

### Network Policies

Default-deny policies in `system/network-policies/`. Each app namespace should have explicit allow rules.

## Notes

- Kubeconfig: `cloud/kubeconfig` (fetched via `mise run cloud:kubeconfig`)
- Storage: `local-path` storage class (k3s default)
- TLS: cert-manager with Let's Encrypt
- Ingress: nginx-ingress-controller
