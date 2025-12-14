# n8n Deployment

n8n is a workflow automation platform deployed on the k3s cluster.

## Architecture

- **n8n**: Workflow automation platform
- **Database**: Shared postgres in `postgres` namespace
- **Storage**: 2Gi for n8n data (local-path storage class)
- **Ingress**: nginx-ingress with automatic TLS (Let's Encrypt)

## Security Features

1. **Network Policies**: Strict ingress/egress rules
2. **Secrets**: Age-encrypted in cloud/fnox.toml
3. **TLS Encryption**: Automatic HTTPS with Let's Encrypt via cert-manager
4. **Ingress Only**: ClusterIP service, only accessible via HTTPS

## Deployment

Secrets are managed via age encryption in `cloud/fnox.toml`. Deploy using:

```bash
# Full deployment (system + secrets + apps)
./cloud/scripts/k8s-deploy.sh

# Or deploy just n8n
./cloud/scripts/k8s-deploy.sh apps/n8n
```

### Create database

The postgres init script should create the n8n database, but if needed:
```bash
kubectl exec -n postgres deploy/postgres -- psql -U postgres -c "CREATE DATABASE n8n;"
```

## Accessing n8n

### Primary Access (HTTPS)

```
https://n8n.cloud.artyom.me
```

### Alternative Access (Port Forward)

For local development:
```bash
kubectl port-forward -n n8n svc/n8n 5678:5678
```

Then open http://localhost:5678

## Management

### Check Status

```bash
kubectl get all -n n8n
kubectl logs -n n8n -l service=n8n --tail=100 -f
```

### Database Access

```bash
kubectl exec -n postgres -it deploy/postgres -- psql -U postgres -d n8n
```

### Restart

```bash
kubectl rollout restart deployment/n8n -n n8n
```

## Cleanup

Remove the entire n8n deployment:
```bash
kubectl delete namespace n8n
```
