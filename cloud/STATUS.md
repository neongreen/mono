# Cloud Infrastructure Status

**Last Updated:** 2025-11-10

## ✅ Working Infrastructure

### Infrastructure Layer (Auto-deployed via Spacelift)
- **Hetzner Cloud VM**: cpx22 (2 vCPU, 4GB RAM) in nbg1
- **Operating System**: Ubuntu 24.04 LTS
- **Kubernetes**: k3s v1.33.5+k3s1
- **Container Isolation**: gVisor (runsc) properly configured
- **Networking**: Private network (10.0.0.0/16), firewall configured
- **Storage**: 80GB SSD, local-path storage class

### Deployed Applications

#### Sample Apps (sample-apps namespace)
1. **nginx-gvisor**: Simple nginx web server
   - Running with gVisor isolation ✓
   - Status: Running

2. **httpbin-gvisor**: HTTP testing service
   - Running with gVisor isolation ✓
   - Status: Running

#### System Services
1. **Dagger Engine** (dagger namespace)
   - Version: v0.19.6
   - Runs WITHOUT gVisor (requires privileged access for CNI)
   - Workloads executed by Dagger can use gVisor
   - Status: Running
   - Logs task: Fixed and working

2. **n8n** (n8n namespace) - Ready to deploy
   - Workflow automation platform
   - n8n: Runs WITH gVisor isolation ✓
   - PostgreSQL 16: Runs without gVisor (database performance)
   - Storage: 10Gi (postgres), 2Gi (n8n)
   - Access: ClusterIP + port-forward
   - Deploy: `mise run cloud:n8n:deploy`

3. **Onyx** (onyx namespace) - Ready to deploy
   - Self-hosted AI platform for work
   - VM: Dedicated vm2-onyx (cpx51: 16 vCPU, 32GB RAM)
   - Services: api-server, background, web-server, mcp-server, code-interpreter, redis, vespa, minio
   - Runs WITH gVisor isolation ✓ (except code-interpreter)
   - Storage: 50Gi (vespa), 50Gi (minio), 10Gi (model-cache)
   - Database: Shared postgres from n8n namespace
   - Access: https://onyx.cloud.artyom.me (public ingress)
   - Features: Basic auth, MCP server, code interpreter, NO local AI models
   - Deploy: See `cloud/k8s/onyx/README.md`

### Security & Isolation

#### Container Isolation (gVisor)
- ✅ gVisor RuntimeClass deployed and working
- ✅ Sample apps verified to run with gVisor (`dmesg` shows "Starting gVisor...")
- ✅ Provides sandboxed execution for untrusted workloads

#### Network Isolation
- ✅ Default-deny NetworkPolicy deployed to sample-apps namespace
- ✅ Verified: Pods cannot communicate without explicit allow rules
- ✅ DNS resolution blocked between isolated pods

## 📝 Completed Tasks

- ✅ cloud-1: Fix Dagger Engine version mismatch (v0.13.3 → v0.19.6)
- ✅ cloud-2: Enable gVisor for workloads (documented Dagger Engine limitation)
- ✅ cloud-3: Fix cloud:dagger:logs task label selector
- ✅ cloud-5: Set up network isolation with NetworkPolicies
- ✅ cloud-8: Experiment: Deploy a test app with gVisor
- ✅ cloud-10: Fix gVisor installation in terraform cloud-init
- ✅ cloud-126: Analyze Onyx docker-compose configuration
- ✅ cloud-127: Design k8s architecture for Onyx
- ✅ cloud-128: Add vm2-onyx (cpx51) to Terraform
- ✅ cloud-129: Configure DNS for onyx.cloud.artyom.me
- ✅ cloud-130: Create k8s manifests for all Onyx services
- ✅ cloud-131: Configure fnox secrets for Onyx
- ✅ cloud-132: Create postgres setup script
- ✅ cloud-133: Configure Ingress for public access

## 🚀 How to Deploy New Services

### Deploying with gVisor Isolation

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-service
  namespace: my-namespace
spec:
  template:
    spec:
      runtimeClassName: gvisor  # Add this line
      containers:
        - name: app
          image: my-image
```

### Adding Network Isolation

```yaml
# Apply default-deny to namespace first
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: default-deny-all
  namespace: my-namespace
spec:
  podSelector: {}
  policyTypes:
    - Ingress
    - Egress
```

Then add specific allow rules as needed.

## 🛠️ Useful Commands

```bash
# Check infrastructure status
mise run cloud:outputs:hcloud

# Get cluster access
mise run cloud:kubeconfig

# Deploy/update applications
mise run cloud:kubectl apply -f k8s/my-app/

# Check Dagger status
mise run cloud:dagger:status
mise run cloud:dagger:logs

# Watch Spacelift deployments
mise run cloud:watch
```

## 📋 Remaining Tasks

- cloud-4: Investigate Dagger GC error (low priority, non-critical)
- cloud-6: Document pattern for deploying isolated services
- cloud-7: Set up CI runner infrastructure

## 🎯 Architecture Decisions

### gVisor Usage
- **Workload Isolation**: Use `runtimeClassName: gvisor` for untrusted code
- **System Services**: Services needing privileged access (like Dagger Engine) run with standard runtime
- **Performance**: gVisor adds ~10-30% overhead but provides strong isolation

### Network Security
- **Default-Deny**: All namespaces should have default-deny NetworkPolicy
- **Explicit Allow**: Only explicitly allow required connections
- **Namespace Isolation**: Each service gets its own namespace

### Deployment Model
- **Infrastructure**: Declarative (OpenTofu) → Auto-deploy via Spacelift
- **Applications**: Declarative (k8s manifests) → Manual deploy via `mise run cloud:kubectl`
- **GitOps**: All config in git, reproducible deployments
