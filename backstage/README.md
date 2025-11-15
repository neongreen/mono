# [Backstage](https://backstage.io)

This app is managed via the repo-wide `mise` tasks so it stays consistent with other projects.

## Development

```bash
# Install deps + start dev server
mise run backstage

# Run tests
mise run backstage:test

# Build frontend + backend bundles
mise run backstage:build
```

## Container image

Build and optionally push to GitHub Container Registry:

```bash
# Build locally (no push)
PUSH=0 mise run backstage:image IMAGE_TAG=dev

# Build + push (requires GHCR_TOKEN in fnox)
mise run backstage:image IMAGE_TAG=main.$(date +%Y%m%d%H%M)
```

The mise task logs into GHCR using `GHCR_TOKEN` from fnox, builds with BuildKit, and pushes when `PUSH=1` (default).
