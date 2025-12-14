# Kubernetes Manifests

Kubernetes manifests for deploying to the k3s cluster.

## Structure

```
k8s/
├── system/              # Cluster infrastructure (deployed first)
│   ├── gvisor/          # RuntimeClass for container isolation
│   ├── ingress-nginx/   # Ingress controller
│   └── cert-manager/    # TLS certificate automation
│
└── apps/                # User applications
    ├── postgres/        # Shared PostgreSQL database
    ├── dagger/          # CI/CD engine
    ├── n8n/             # Workflow automation
    ├── onyx/            # AI platform
    └── coder/           # Remote dev environments
```

## Deployment

Deploy using the deploy script:

```bash
# Deploy everything (system + secrets + apps)
./cloud/scripts/k8s-deploy.sh

# Deploy only system components
./cloud/scripts/k8s-deploy.sh system

# Deploy only secrets
./cloud/scripts/k8s-deploy.sh secrets

# Deploy only apps
./cloud/scripts/k8s-deploy.sh apps

# Deploy specific app
./cloud/scripts/k8s-deploy.sh apps/n8n
```

### Deployment Order

The deploy script handles ordering automatically:
1. System components (gvisor, ingress-nginx, cert-manager)
2. Secrets (from fnox.toml)
3. Postgres (must be first app)
4. Other apps

### Prerequisites

1. Node must have workload label:
   ```bash
   kubectl label node mono workload=onyx
   ```

2. Databases must be created:
   ```bash
   kubectl exec -n postgres deploy/postgres -- psql -U postgres -c "CREATE DATABASE n8n;"
   kubectl exec -n postgres deploy/postgres -- psql -U postgres -c "CREATE DATABASE coder;"
   kubectl exec -n postgres deploy/postgres -- psql -U postgres -c "CREATE DATABASE onyx;"
   ```

## Security

### gVisor Runtime

Container isolation via gVisor:
- Use `runtimeClassName: gvisor` in pod specs
- ~10-20% performance overhead
- Not compatible with privileged containers or docker socket

### Network Policies

Each namespace has network policies with:
- Allow ingress from nginx-ingress
- Allow DNS egress
- Allow postgres egress
- Allow external internet (via ipBlock)

## Notes

- Secrets: Age-encrypted in `cloud/fnox.toml`
- Storage: `local-path` storage class (k3s default)
- TLS: cert-manager with Let's Encrypt
- Ingress: nginx-ingress-controller (traefik disabled)
