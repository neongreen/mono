# Coder Deployment

Coder provides cloud development environments. This deploys Coder to the k3s cluster.

## Prerequisites

- Kubernetes cluster with Traefik ingress
- cert-manager with letsencrypt-prod cluster issuer
- Shared PostgreSQL in n8n namespace
- fnox configured with keychain secrets

## Architecture

- **Database**: Shared PostgreSQL in n8n namespace (`coder` database)
- **URL**: https://coder.cloud.artyom.me
- **Workspaces**: Run as Kubernetes pods in the cluster

## Deployment

### 1. Add secrets to keychain

```bash
# Generate a strong password
CODER_PG_PASS=$(openssl rand -base64 24)
echo "Coder PG Password: $CODER_PG_PASS"

# Add to keychain
security add-generic-password -a fnox -s fnox -D "CODER_POSTGRES_PASSWORD" -w "$CODER_PG_PASS"
```

### 2. Create namespace and network policies

```bash
kubectl apply -f namespace.yaml
kubectl apply -f network-policy.yaml
```

### 3. Update n8n namespace (for cross-namespace postgres access)

```bash
# Add namespace label and update postgres network policy
kubectl apply -f ../n8n/namespace.yaml
kubectl apply -f ../n8n/network-policy.yaml
```

### 4. Create Kubernetes secrets

```bash
cd /path/to/mono/cloud
fnox exec -- kubectl create secret generic coder-secrets -n coder \
  --from-literal=POSTGRES_PASSWORD="$CODER_POSTGRES_PASSWORD" \
  --from-literal=POSTGRES_ADMIN_PASSWORD="$N8N_POSTGRES_ADMIN_PASSWORD"
```

### 5. Initialize the database

```bash
kubectl apply -f db-init-job.yaml

# Wait for job to complete
kubectl wait --for=condition=complete job/coder-db-init -n coder --timeout=60s

# Check logs if needed
kubectl logs -n coder job/coder-db-init
```

### 6. Deploy Coder

```bash
kubectl apply -f rbac.yaml
kubectl apply -f service.yaml
kubectl apply -f deployment.yaml
kubectl apply -f ingress.yaml
```

### 7. Create first admin user

```bash
# Get the Coder pod name
POD=$(kubectl get pods -n coder -l app=coder -o jsonpath='{.items[0].metadata.name}')

# Create admin user
kubectl exec -it -n coder $POD -- coder users create --username admin --email admin@example.com
```

## Files

| File | Description |
|------|-------------|
| `namespace.yaml` | Coder namespace |
| `network-policy.yaml` | Network policies (default-deny + allow rules) |
| `rbac.yaml` | ServiceAccount and ClusterRole for workspace provisioning |
| `configmap.yaml` | Coder configuration (not currently used, env vars in deployment) |
| `deployment.yaml` | Main Coder deployment |
| `service.yaml` | ClusterIP service |
| `ingress.yaml` | Traefik ingress with TLS |
| `db-init-job.yaml` | One-time job to create coder database |
| `secret-template.yaml` | Documentation of required secrets |

## Workspace Templates

After deployment, create workspace templates in the Coder UI or CLI:

```bash
# Example: Create a Kubernetes template
coder templates create kubernetes --url https://coder.cloud.artyom.me
```

## Troubleshooting

```bash
# Check Coder logs
kubectl logs -n coder -l app=coder -f

# Check database connectivity
kubectl exec -it -n coder deploy/coder -- env | grep PG

# Verify ingress
kubectl get ingress -n coder
kubectl describe ingress coder-ingress -n coder
```
