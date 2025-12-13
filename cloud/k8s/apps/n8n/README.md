# n8n Deployment

n8n is a workflow automation platform deployed on the k3s cluster with PostgreSQL as the database backend.

## Architecture

- **n8n**: Workflow automation platform (runs with gVisor isolation)
- **PostgreSQL 16**: Database backend (runs without gVisor for performance)
- **Storage**: Uses local-path storage class (10Gi for postgres, 2Gi for n8n)
- **Network**: ClusterIP service, exposed only via Traefik Ingress with automatic TLS (Let's Encrypt)

## Security Features

1. **Container Isolation**: n8n runs with gVisor runtime for sandboxed execution (disabled temporarily)
2. **Network Policies**: Strict ingress/egress rules between components
3. **Secrets**: Generated secure passwords for database access
4. **TLS Encryption**: Automatic HTTPS with Let's Encrypt certificates via cert-manager
5. **Ingress Only**: ClusterIP service means n8n is only accessible via HTTPS Ingress, not directly exposed
6. **HTTP Redirect**: Traefik ingress with automatic HTTP → HTTPS redirect

## Deployment

### Secrets Management

**Single-user setup (current):** Database passwords are stored in fnox (local keychain) and injected at deployment time. This follows the same pattern as other cloud infrastructure secrets (Hetzner API, Spacelift).

The passwords are set up in fnox:
```bash
# First-time setup or password rotation
fnox set N8N_POSTGRES_ADMIN_PASSWORD "$(openssl rand -base64 32)"
fnox set N8N_POSTGRES_N8N_PASSWORD "$(openssl rand -base64 32)"
```

**Multi-user / CI/CD setup:** If you need to share deployment access:
1. Use SOPS with age/pgp keys for encrypted secrets in git
2. Use Kubernetes External Secrets Operator with a shared vault
3. Create a gitignored `postgres-secret.yaml` that each user maintains locally

For now, fnox works since only you deploy to this cluster.

### Deploy

Deploy all components (passwords injected from fnox):
```bash
mise run cloud:kubectl apply -f k8s/n8n/
```

Or deploy step-by-step:
```bash
# 1. Create namespace and secrets
mise run cloud:kubectl apply -f k8s/n8n/namespace.yaml
mise run cloud:kubectl apply -f k8s/n8n/postgres-secret.yaml
mise run cloud:kubectl apply -f k8s/n8n/postgres-configmap.yaml

# 2. Create persistent volumes
mise run cloud:kubectl apply -f k8s/n8n/postgres-pvc.yaml
mise run cloud:kubectl apply -f k8s/n8n/n8n-pvc.yaml

# 3. Deploy PostgreSQL
mise run cloud:kubectl apply -f k8s/n8n/postgres-deployment.yaml
mise run cloud:kubectl apply -f k8s/n8n/postgres-service.yaml

# 4. Wait for postgres to be ready
mise run cloud:kubectl wait --for=condition=ready pod -l service=postgres-n8n -n n8n --timeout=120s

# 5. Deploy n8n
mise run cloud:kubectl apply -f k8s/n8n/n8n-deployment.yaml
mise run cloud:kubectl apply -f k8s/n8n/n8n-service.yaml

# 6. Apply network policies
mise run cloud:kubectl apply -f k8s/n8n/network-policy.yaml

# 7. Deploy Ingress (requires DNS to be set up first)
mise run cloud:kubectl apply -f k8s/n8n/n8n-ingress.yaml
```

## Accessing n8n

### Primary Access (HTTPS with TLS)

Once DNS is configured (via Terraform in `cloud/terraform/02-digitalocean-dns`):

```
https://n8n.cloud.artyom.me
```

- Automatic HTTPS with Let's Encrypt certificate
- HTTP automatically redirects to HTTPS
- Webhooks configured for this domain

### Alternative Access (Port Forward)

For local development or before DNS is set up:

```bash
mise run cloud:kubectl port-forward -n n8n svc/n8n 5678:5678
```

Then open http://localhost:5678 in your browser.

## Management

### Check Status

```bash
# View all n8n resources
mise run cloud:kubectl get all -n n8n

# Check pod status
mise run cloud:kubectl get pods -n n8n

# View n8n logs
mise run cloud:kubectl logs -n n8n -l service=n8n --tail=100 -f

# View postgres logs
mise run cloud:kubectl logs -n n8n -l service=postgres-n8n --tail=100 -f
```

### Verify gVisor Isolation

Check that n8n is running with gVisor:
```bash
mise run cloud:kubectl get pod -n n8n -l service=n8n -o jsonpath='{.items[0].spec.runtimeClassName}'
# Should output: gvisor
```

### Database Access

Connect to postgres for debugging:
```bash
mise run cloud:kubectl exec -it -n n8n deployment/postgres -- psql -U n8n -d n8n
```

### Storage

Check persistent volume claims:
```bash
mise run cloud:kubectl get pvc -n n8n
```

## Scaling

n8n is deployed with `replicas: 1`. For high availability:
1. Set `replicas: 2` or more in n8n-deployment.yaml
2. Consider using a managed PostgreSQL database
3. Ensure shared storage if needed for workflows

## Backup

### Database Backup

```bash
# Create a database backup
mise run cloud:kubectl exec -n n8n deployment/postgres -- pg_dump -U postgres n8n > n8n-backup-$(date +%Y%m%d).sql
```

### Restore from Backup

```bash
# Restore database
cat n8n-backup-20240101.sql | mise run cloud:kubectl exec -i -n n8n deployment/postgres -- psql -U postgres -d n8n
```

## Troubleshooting

### n8n won't start

Check logs:
```bash
mise run cloud:kubectl logs -n n8n -l service=n8n
```

Common issues:
- Database not ready: Wait for postgres pod to be running
- Permission issues: Check volume-permissions init container logs

### Database connection issues

Verify postgres is running:
```bash
mise run cloud:kubectl get pods -n n8n -l service=postgres-n8n
```

Test connection from n8n pod:
```bash
mise run cloud:kubectl exec -n n8n deployment/n8n -- nc -zv postgres-service 5432
```

### Network policy issues

Temporarily disable network policies for debugging:
```bash
mise run cloud:kubectl delete networkpolicies -n n8n --all
```

Re-apply when done:
```bash
mise run cloud:kubectl apply -f k8s/n8n/network-policy.yaml
```

## Cleanup

Remove the entire n8n deployment:
```bash
mise run cloud:kubectl delete namespace n8n
```

This will delete all resources (deployments, services, PVCs, secrets) in the n8n namespace.

## Production Considerations

Before using n8n in production:

1. **External Access**: Set up Ingress with TLS instead of port-forward
2. **Domain**: Configure N8N_HOST environment variable
3. **Authentication**: Enable user authentication in n8n
4. **Secrets**: Store secrets in a secure secrets manager (not in git)
5. **Backups**: Set up automated database backups
6. **Monitoring**: Add monitoring and alerting for n8n and postgres
7. **Updates**: Plan for version updates and migrations
