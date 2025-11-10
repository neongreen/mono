# Miniflux RSS Reader

Deployed application demonstrating complete backend + frontend + database stack with proper isolation.

## Architecture

- **Backend**: Miniflux (Go) - RSS feed reader with built-in web UI
- **Frontend**: Static files served by Miniflux backend
- **Database**: PostgreSQL 16 with persistent storage
- **Isolation**: gVisor for Miniflux pods, NetworkPolicies for network isolation

## Components

### Namespace
- `miniflux` - Isolated namespace for the application

### Database
- StatefulSet: `postgres`
- Persistent storage: 5Gi
- Credentials stored in Secret `postgres-credentials`

### Application
- Deployment: `miniflux`
- Image: `miniflux/miniflux:latest`
- Runtime: gVisor isolation enabled
- Service: ClusterIP on port 8080

### Network Security
- Default-deny policy blocks all traffic
- Explicit allow rules:
  - Miniflux → PostgreSQL (port 5432)
  - Miniflux → Internet (ports 80, 443 for RSS feeds)
  - Miniflux → DNS (kube-system namespace)
  - External → Miniflux (port 8080 for ingress)

## Deployment

```bash
# Deploy all components
mise run cloud:kubectl apply -f k8s/miniflux/

# Check status
mise run cloud:kubectl get pods -n miniflux
mise run cloud:kubectl get svc -n miniflux

# View logs
mise run cloud:kubectl logs -n miniflux -l app=miniflux

# Access via port-forward (if needed)
mise run cloud:kubectl port-forward -n miniflux svc/miniflux 8080:8080
```

## Default Credentials

- Username: `admin`
- Password: `miniflux-admin-2024`

**Note**: Change these credentials after first login in a production deployment.

## Production Considerations

For production use, consider:
1. Set up Ingress with TLS termination
2. Use stronger passwords stored in external secrets manager
3. Configure backup strategy for PostgreSQL data
4. Set up monitoring and alerting
5. Configure RSS feed refresh schedules
