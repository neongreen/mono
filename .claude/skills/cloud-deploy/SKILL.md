---
name: cloud-deploy
description: Deploy Kubernetes manifests to the mono cluster, manage Spacelift stacks, and run cloud infrastructure operations. Use when asked to deploy, check cluster status, or manage cloud resources.
---

# Cloud Deploy Skill

## Purpose

Provides safe, consistent methods for deploying to the mono Kubernetes cluster and managing cloud infrastructure via Spacelift.

## When to Use

Use this skill when the user requests:
- "Deploy to the cluster"
- "Deploy the k8s manifests"
- "Check cluster status"
- "Apply changes to k8s"
- "Check Spacelift status"
- "What's running in the cluster?"

## Architecture Overview

**Infrastructure** (`cloud/terraform/`):
- Managed by Spacelift `mono/cloud` stack
- Auto-deploys on push to main
- DO NOT manually run tofu apply

**Kubernetes** (`cloud/k8s/`):
- Deployed manually via scripts
- System components in `cloud/k8s/system/`
- Applications in `cloud/k8s/apps/`

## Scripts

All scripts are in `cloud/scripts/`:

### k8s-deploy.sh - Main Deploy Script

```bash
# Deploy everything (system + apps)
./cloud/scripts/k8s-deploy.sh

# Deploy only system components
./cloud/scripts/k8s-deploy.sh system

# Deploy only apps
./cloud/scripts/k8s-deploy.sh apps

# Deploy specific app
./cloud/scripts/k8s-deploy.sh apps/dagger
```

### k8s-kubeconfig.sh - Fetch Kubeconfig

```bash
# Fetch and export kubeconfig
eval $(./cloud/scripts/k8s-kubeconfig.sh)
```

### k8s-kubectl.sh - Kubectl Wrapper

```bash
# Run kubectl commands against the cluster
./cloud/scripts/k8s-kubectl.sh get pods -A
./cloud/scripts/k8s-kubectl.sh logs -n n8n deployment/n8n
```

## Deployment Order

System components MUST be deployed in this order:
1. gvisor (container runtime)
2. network-policies (security)
3. ingress-nginx (ingress controller)
4. cert-manager (TLS certificates)

Apps can be deployed in any order after system components.

## Common Operations

### Check Cluster Status

```bash
./cloud/scripts/k8s-kubectl.sh get nodes
./cloud/scripts/k8s-kubectl.sh get pods -A
./cloud/scripts/k8s-kubectl.sh get ingress -A
```

### Deploy New App

1. Create manifests in `cloud/k8s/apps/<app-name>/`
2. Run `./cloud/scripts/k8s-deploy.sh apps/<app-name>`

### Check Spacelift Status

```bash
mise x -- fnox exec -- spacectl stack list
mise x -- fnox exec -- spacectl stack run list --id mono-cloud-01
```

### View Spacelift Logs

```bash
mise x -- fnox exec -- spacectl stack logs --id mono-cloud-01 --run <run-id>
```

## Safety Rules

### NEVER DO:
- Modify `ssh_keys` in terraform (destroys server!)
- Run `tofu apply` manually (use Spacelift)
- Commit secrets to the repo
- Delete namespaces without confirmation

### ALWAYS DO:
- Use the deploy scripts (not raw kubectl)
- Check cluster status after deploys
- Commit manifest changes before deploying
- Wait for Spacelift runs to complete before making more infra changes

## Troubleshooting

### Kubeconfig Issues

```bash
# Re-fetch kubeconfig
rm ~/.kube/mono-config
./cloud/scripts/k8s-kubeconfig.sh
```

### Pod Not Starting

```bash
./cloud/scripts/k8s-kubectl.sh describe pod <pod-name> -n <namespace>
./cloud/scripts/k8s-kubectl.sh logs <pod-name> -n <namespace>
```

### Ingress Not Working

```bash
# Check ingress-nginx
./cloud/scripts/k8s-kubectl.sh get pods -n ingress-nginx
./cloud/scripts/k8s-kubectl.sh logs -n ingress-nginx deployment/ingress-nginx-controller
```

### Cert-Manager Issues

```bash
# Check certificate status
./cloud/scripts/k8s-kubectl.sh get certificates -A
./cloud/scripts/k8s-kubectl.sh describe certificate <name> -n <namespace>
```

## Server Access

SSH to the mono server:
```bash
ssh emily@mono.cloud.artyom.me
```

K3s kubeconfig is at `/etc/rancher/k3s/k3s.yaml` on the server.
