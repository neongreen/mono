# Kubernetes Manifests

This directory contains Kubernetes manifests for deploying applications to the k3s cluster.

## Deployment Philosophy

**Infrastructure vs Applications:**
- **Infrastructure** (VMs, networks, k3s): Managed by OpenTofu/Spacelift, auto-deploys on git push
- **Applications** (Kubernetes resources): Manual deployment via `kubectl apply`

This separation keeps infrastructure automated while giving explicit control over application deployments.

## Structure

- `gvisor/` - gVisor runtime configuration
  - `runtimeclass.yaml` - RuntimeClass for gVisor-isolated pods
- `dagger/` - Dagger Engine deployment
  - `deployment.yaml` - Kubernetes manifests (Deployment, Service, PVC, Namespace)

## Security: gVisor Runtime

The cluster is configured with gVisor for enhanced container isolation:
- Syscalls intercepted in userspace (not passed to kernel)
- Suitable for running untrusted workloads
- ~10-20% performance overhead vs regular containers
- Deploy RuntimeClass: `mise run cloud:gvisor:deploy`

Pods using gVisor set `runtimeClassName: gvisor` in their spec.

## Deployment

### Dagger Engine

Deploy Dagger Engine (with gVisor):
```bash
mise run cloud:dagger:deploy
```

Check status:
```bash
mise run cloud:dagger:status
```

View logs:
```bash
mise run cloud:dagger:logs
```

Delete deployment:
```bash
mise run cloud:kubectl delete -f k8s/dagger/deployment.yaml
```

## Workflow

1. **Modify manifests** in `cloud/k8s/`
2. **Commit and push** to git
3. **Wait for Spacelift** to apply infrastructure changes (if any)
4. **Manually deploy** Kubernetes resources:
   ```bash
   mise run cloud:dagger:deploy
   # or
   mise run cloud:kubectl apply -f k8s/dagger/deployment.yaml
   ```

## Why Manual Deployment?

Kubernetes resources are deployed manually (not via Spacelift) because:
- **Explicit control**: You decide when applications are deployed
- **Simpler setup**: No need for GitOps tools (ArgoCD/Flux) or Terraform Kubernetes provider
- **Clear separation**: Infrastructure changes (VM recreation) are separate from app deployments

To make deployments automatic, you would need:
- **Option A**: Terraform Kubernetes provider in Spacelift (couples infrastructure and apps)
- **Option B**: GitOps (ArgoCD/Flux) - watches git, auto-applies manifests
- **Current**: Manual via `mise run cloud:kubectl` tasks (simple, explicit)

## Notes

- All deployments use the kubeconfig from `cloud/kubeconfig` (automatically set via `cloud:kubectl` task)
- Manifests are version-controlled and declarative
- Namespaces are created automatically via manifests
- Storage uses `local-path` storage class (comes with k3s)
- gVisor runtime is installed via cloud-init and configured in containerd

