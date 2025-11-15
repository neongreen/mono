# Backstage Deployment Manifests

These files support task `cloud-24` (author Kubernetes manifests) and serve as the base for `cloud-25` (secrets apply flow) and `cloud-27` (deploy + verify).

## Structure

```
cloud/k8s/backstage/
  namespace.yaml           # Namespace definition
  rbac.yaml                # ServiceAccount + read-only ClusterRole
  configmap.yaml           # Production app-config (mounted into the pod)
  secrets.tmpl.yaml        # Template rendered via fnox/envsubst
  deployment.yaml          # Backstage Deployment
  service.yaml             # ClusterIP service for the pod
  ingress.yaml             # Optional Traefik ingress + TLS
  postgres/
    values.yaml            # Helm overrides for Bitnami PostgreSQL
    secret.tmpl.yaml       # Template for DB credentials
```

## Apply order

1. `kubectl apply -f cloud/k8s/backstage/namespace.yaml`
2. `kubectl apply -f cloud/k8s/backstage/rbac.yaml`
3. Render secrets (requires fnox secrets to be set):
   ```bash
   mise run cloud:backstage:secrets
   ```
4. Install PostgreSQL:
   ```bash
   helm upgrade --install backstage-postgres oci://registry-1.docker.io/bitnamicharts/postgresql \
     -n backstage \
     -f cloud/k8s/backstage/postgres/values.yaml
   ```
5. `kubectl apply -f cloud/k8s/backstage/configmap.yaml`
6. `kubectl apply -f cloud/k8s/backstage/deployment.yaml`
7. `kubectl apply -f cloud/k8s/backstage/service.yaml`
8. (Optional) `kubectl apply -f cloud/k8s/backstage/ingress.yaml`

## Notes

- `deployment.yaml` mounts `app-config.production.yaml` from the ConfigMap. Keep that file synchronized with `backstage/app-config.production.yaml`.
- Secrets use `${VAR}` placeholders for `fnox exec -- envsubst`. See `cloud/backstage/secret-plan.md`.
- Ingress references `backstage.artyom.me` and cert-manager issuer `letsencrypt-prod`. Adjust if DNS or issuer changes.
- The Deployment sets `APP_CONFIG_backend_database_connection_string` from the `DATABASE_URL` key inside `backstage-secrets`. Populate that key when rendering the secret (e.g., `postgres://backstage:${BACKSTAGE_DB_PASSWORD}@backstage-postgres:5432/backstage`).
