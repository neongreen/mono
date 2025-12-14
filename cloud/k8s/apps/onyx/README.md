# Onyx Deployment

Self-hosted Onyx AI platform running on k3s.

## Architecture

- **Infrastructure**: Single mono server (CX53: 4 vCPU, 8GB RAM)
- **Node labeling**: All pods require `workload=onyx` label on node
- **Isolation**: gVisor runtime for most services (except code-interpreter which is disabled)
- **Storage**: 20Gi for Vespa, 20Gi for MinIO, 10Gi for model cache
- **Database**: Shared postgres in `postgres` namespace

## Services

### Infrastructure
- **redis**: Cache (ephemeral)
- **vespa**: Vector database (persistent)
- **minio**: S3-compatible object storage (persistent)

### Application
- **api-server**: Main backend API
- **background**: Background worker
- **web-server**: Next.js frontend
- **mcp-server**: Model Context Protocol server

### What's NOT deployed
- code-interpreter (requires Docker socket, not available on k3s/containerd)
- inference_model_server (disabled via DISABLE_MODEL_SERVER=true)
- indexing_model_server (disabled via DISABLE_MODEL_SERVER=true)

## Deployment

Secrets are managed via age encryption in `cloud/fnox.toml`. Deploy using:

```bash
# Full deployment (system + secrets + apps)
./cloud/scripts/k8s-deploy.sh

# Or deploy just onyx
./cloud/scripts/k8s-deploy.sh apps/onyx
```

### Node labeling requirement

The node must have the `workload=onyx` label:
```bash
kubectl label node mono workload=onyx
```

### Create database

The postgres init script should create the onyx database, but if needed:
```bash
kubectl exec -n postgres deploy/postgres -- psql -U postgres -c "CREATE DATABASE onyx;"
```

## Access

- **URL**: https://onyx.cloud.artyom.me
- **Auth**: Basic (username/password)
- **First time**: Create admin account via web interface

## Configuration

Key settings in `configmap.yaml`:
- `AUTH_TYPE=basic`: Username/password authentication
- `DISABLE_MODEL_SERVER=true`: Use external AI APIs instead of local models
- `MCP_SERVER_ENABLED=true`: Enable MCP server

## Troubleshooting

### Check pod logs
```bash
kubectl logs -n onyx deployment/api-server
kubectl logs -n onyx deployment/background
kubectl logs -n onyx deployment/vespa
```

### Check resource usage
```bash
kubectl top pods -n onyx
kubectl top node mono
```

### Restart services
```bash
kubectl rollout restart deployment/api-server -n onyx
kubectl rollout restart deployment/background -n onyx
```

### Database access
```bash
kubectl exec -n postgres -it deploy/postgres -- psql -U postgres -d onyx
```
