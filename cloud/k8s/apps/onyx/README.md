# Onyx Deployment

Self-hosted Onyx AI platform running on k3s.

## Architecture

- **Infrastructure**: Dedicated cpx51 VM (vm2-onyx) with 16 vCPU, 32GB RAM
- **Node labeling**: All pods pinned to `workload=onyx` node
- **Isolation**: gVisor runtime for most services (except code-interpreter)
- **Storage**: 50Gi for Vespa, 50Gi for MinIO, 10Gi for model cache
- **Database**: Shared postgres from n8n namespace

## Services

### Infrastructure
- **redis**: Cache (ephemeral)
- **vespa**: Vector database (persistent)
- **minio**: S3-compatible object storage (persistent)

### Application
- **api-server**: Main backend API
- **background**: Background worker (heavy - 8Gi RAM)
- **web-server**: Next.js frontend
- **mcp-server**: Model Context Protocol server
- **code-interpreter**: Code execution environment

### What's NOT deployed
- ❌ inference_model_server (disabled via DISABLE_MODEL_SERVER=true)
- ❌ indexing_model_server (disabled via DISABLE_MODEL_SERVER=true)

## Deployment

### Prerequisites

1. **Secrets**: Add to `cloud/fnox.toml`:
   ```toml
   # Onyx secrets
   ONYX_POSTGRES_PASSWORD = { provider = "keychain", value = "ONYX_POSTGRES_PASSWORD" }
   ONYX_SECRET = { provider = "keychain", value = "ONYX_SECRET" }
   ONYX_MINIO_ROOT_PASSWORD = { provider = "keychain", value = "ONYX_MINIO_ROOT_PASSWORD" }
   ```

2. **Store secrets in keychain**:
   ```bash
   security add-generic-password -a "$USER" -s fnox -w "your-postgres-password" -l "ONYX_POSTGRES_PASSWORD"
   security add-generic-password -a "$USER" -s fnox -w "$(openssl rand -hex 32)" -l "ONYX_SECRET"
   security add-generic-password -a "$USER" -s fnox -w "your-minio-password" -l "ONYX_MINIO_ROOT_PASSWORD"
   ```

3. **Create k8s secret**:
   ```bash
   fnox exec -- kubectl create secret generic onyx-secrets -n onyx \
     --from-literal=POSTGRES_PASSWORD="$ONYX_POSTGRES_PASSWORD" \
     --from-literal=SECRET="$ONYX_SECRET" \
     --from-literal=S3_AWS_SECRET_ACCESS_KEY="$ONYX_MINIO_ROOT_PASSWORD"
   ```

4. **Create Postgres database** (in n8n's postgres pod):
   ```bash
   kubectl exec -n n8n postgres-<pod-id> -- psql -U postgres -c "CREATE DATABASE onyx;"
   ```

### Deploy

```bash
# Apply all manifests
kubectl apply -f cloud/k8s/onyx/

# Check deployment
kubectl get pods -n onyx
kubectl get svc -n onyx
kubectl get ingress -n onyx
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
- `CODE_INTERPRETER_BETA_ENABLED=true`: Enable code execution

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
kubectl top node vm2-onyx
```

### Restart services
```bash
kubectl rollout restart deployment/api-server -n onyx
kubectl rollout restart deployment/background -n onyx
```

### Database access
```bash
# Connect to postgres
kubectl exec -n n8n -it postgres-<pod-id> -- psql -U postgres -d onyx
```

## Resource Usage

Expected resource consumption:
- **Total**: ~9 vCPU, ~20Gi RAM
- **Heaviest**: background (8Gi), vespa (8Gi)
- **Storage**: ~110Gi total (vespa 50Gi + minio 50Gi + cache 10Gi)
