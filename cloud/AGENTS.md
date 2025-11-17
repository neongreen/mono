# AGENTS.md — cloud

Guidelines for agents working in `cloud/` (OpenTofu + Spacelift) with `fnox`.

## Architecture Overview

Two layers with different deployment approaches:

**Infrastructure Layer (Automated):**
- What: VMs, networks, firewalls, k3s cluster
- Where: `cloud/terraform/`
- How: OpenTofu + Spacelift, auto-deploys on git push
- When: Push to main triggers Spacelift run

**Application Layer (Manual):**
- What: Kubernetes resources (pods, services, etc.)
- Where: `cloud/k8s/`
- How: kubectl via `mise run cloud:kubectl`
- When: Manual deployment after infrastructure is ready

## Secrets & Execution (fnox)

- Always run commands that require secrets via `fnox exec -- …`.
- Do not export secrets into your shell.
- Execute tools with secrets injected:

```bash
# Spacelift CLI
fnox exec -- spacectl whoami
fnox exec -- spacectl stack list

# Local tofu (validation only; no remote backend contact)
cd cloud/terraform/01-hcloud
fnox exec -- tofu init -backend=false -input=false
fnox exec -- tofu validate
```

## Deployment Workflow

**For infrastructure changes (Terraform):**
1. Edit files in `cloud/terraform/`
2. Validate locally: `mise run cloud:validate`
3. Commit and push to git
4. Spacelift auto-deploys (watch with `mise run cloud:watch`)

**For application changes (Kubernetes):**
1. Edit manifests in `cloud/k8s/`
2. Commit and push to git (version control)
3. Wait for infrastructure to be ready (if Terraform changed)
4. Manually deploy: `mise run cloud:dagger:deploy`

This separation keeps infrastructure automated while giving explicit control over when applications are deployed.

## Hetzner Cloud Best Practices

- Provider: `hetznercloud/hcloud` (token from `HCLOUD_TOKEN` environment variable)
- Subnet resource is `hcloud_network_subnet` (not `hcloud_subnetwork`)
- **Use `.name` not `.id`** for immutable server attributes (location, server_type, image)
  - Using `.id` returns numeric IDs which can cause Terraform to see "nbg1" → "2" as a change
  - This forces unnecessary server replacements
  - Use `.name` for stable, human-readable values
- Typical resources:
  - Network: `hcloud_network`, `hcloud_network_subnet`
  - Security: `hcloud_firewall`, `hcloud_firewall_attachment`
  - SSH: `hcloud_ssh_key` (read from `ssh_key.pub` in module directory)
  - Compute: `hcloud_server`, `hcloud_server_network`
