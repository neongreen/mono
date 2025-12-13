# Mono Dev Template

Kubernetes-based development environment for the neongreen/mono monorepo.

## What's included

- Ubuntu-based container with code-server (VS Code in browser)
- Persistent 20GB home directory
- mise for tool management (installs Go, Rust, Node, Python, etc.)
- Auto-clones the mono repo on first start

## Pushing the template

```bash
coder login https://coder.cloud.artyom.me
coder templates push mono-dev --directory cloud/coder-templates/mono-dev
```

## Creating a workspace

```bash
coder create my-workspace --template mono-dev
```

Or use the web UI at https://coder.cloud.artyom.me
