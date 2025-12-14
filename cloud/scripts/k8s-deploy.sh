#!/bin/bash
# k8s-deploy.sh - Deploy Kubernetes manifests to the mono cluster
#
# Usage:
#   ./cloud/scripts/k8s-deploy.sh              # Deploy everything (system + apps + secrets)
#   ./cloud/scripts/k8s-deploy.sh system       # Deploy only system components
#   ./cloud/scripts/k8s-deploy.sh apps         # Deploy only apps (with secrets)
#   ./cloud/scripts/k8s-deploy.sh apps/dagger  # Deploy specific app
#   ./cloud/scripts/k8s-deploy.sh secrets      # Create/update secrets only
#
# This script:
# 1. Fetches kubeconfig from the mono server via SSH
# 2. Creates k8s secrets from fnox (age-encrypted values)
# 3. Applies kustomize manifests in the correct order
# 4. Reports deployment status

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CLOUD_DIR="$(dirname "$SCRIPT_DIR")"
K8S_DIR="$CLOUD_DIR/k8s"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log() { echo -e "${GREEN}[deploy]${NC} $*"; }
warn() { echo -e "${YELLOW}[deploy]${NC} $*"; }
error() { echo -e "${RED}[deploy]${NC} $*" >&2; }

# Get server IP from terraform output or use cached value
get_server_ip() {
    # Try terraform output first
    if command -v tofu &>/dev/null && [[ -f "$CLOUD_DIR/terraform/terraform.tfstate" ]]; then
        ip=$(cd "$CLOUD_DIR/terraform" && tofu output -raw mono_ip 2>/dev/null || true)
        if [[ -n "$ip" ]]; then
            echo "$ip"
            return
        fi
    fi

    # Fall back to DNS
    host mono.cloud.artyom.me 2>/dev/null | grep -oE '[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+' | head -1 || {
        error "Could not determine server IP. Set MONO_SERVER_IP or check DNS."
        exit 1
    }
}

# Fetch kubeconfig from server
setup_kubeconfig() {
    local server_ip="${MONO_SERVER_IP:-$(get_server_ip)}"
    local kubeconfig_dir="$HOME/.kube"
    local kubeconfig_file="$kubeconfig_dir/mono-config"

    log "Fetching kubeconfig from $server_ip..."

    mkdir -p "$kubeconfig_dir"

    ssh -o StrictHostKeyChecking=accept-new "emily@$server_ip" \
        'sudo cat /etc/rancher/k3s/k3s.yaml' | \
        sed "s/127.0.0.1/$server_ip/" > "$kubeconfig_file"

    chmod 600 "$kubeconfig_file"
    export KUBECONFIG="$kubeconfig_file"

    log "Kubeconfig saved to $kubeconfig_file"
}

# Deploy system components (order matters)
deploy_system() {
    log "Deploying system components..."

    # Order: gvisor -> network-policies -> ingress-nginx -> cert-manager
    local components=(
        "gvisor"
        "network-policies"
        "ingress-nginx"
        "cert-manager"
    )

    for component in "${components[@]}"; do
        local dir="$K8S_DIR/system/$component"
        if [[ -d "$dir" ]]; then
            log "  Applying $component..."
            kubectl apply -k "$dir" || {
                warn "  Failed to apply $component (may need to retry)"
            }
        fi
    done

    log "System components deployed."
}

# Create secrets from fnox
deploy_secrets() {
    log "Creating secrets from fnox..."
    "$SCRIPT_DIR/k8s-secrets.sh" all
}

# Deploy all apps or a specific one
deploy_apps() {
    local target="${1:-}"

    if [[ -n "$target" ]]; then
        # Deploy specific app
        local dir="$K8S_DIR/apps/$target"
        if [[ -d "$dir" ]]; then
            log "Deploying app: $target..."
            kubectl apply -k "$dir" 2>/dev/null || kubectl apply -f "$dir"
        else
            error "App not found: $target"
            exit 1
        fi
    else
        # Deploy all apps in order (postgres first, then others)
        log "Deploying all apps..."

        # Postgres must be first - other apps depend on it
        local postgres_dir="$K8S_DIR/apps/postgres"
        if [[ -d "$postgres_dir" ]]; then
            log "  Applying postgres (shared database)..."
            kubectl apply -f "$postgres_dir" || {
                warn "  Failed to apply postgres"
            }
            log "  Waiting for postgres to be ready..."
            kubectl wait --for=condition=ready pod -l app=postgres -n postgres --timeout=120s || {
                warn "  Postgres not ready yet, continuing..."
            }
        fi

        # Then deploy other apps
        for app_dir in "$K8S_DIR/apps"/*/; do
            if [[ -d "$app_dir" ]]; then
                local app_name=$(basename "$app_dir")
                # Skip postgres, already deployed
                [[ "$app_name" == "postgres" ]] && continue
                log "  Applying $app_name..."
                kubectl apply -k "$app_dir" 2>/dev/null || kubectl apply -f "$app_dir" || {
                    warn "  Failed to apply $app_name"
                }
            fi
        done
    fi

    log "Apps deployed."
}

# Main
main() {
    local target="${1:-all}"

    setup_kubeconfig

    case "$target" in
        all)
            deploy_system
            deploy_secrets
            deploy_apps
            ;;
        system)
            deploy_system
            ;;
        secrets)
            deploy_secrets
            ;;
        apps)
            deploy_secrets
            deploy_apps
            ;;
        apps/*)
            deploy_apps "${target#apps/}"
            ;;
        *)
            error "Unknown target: $target"
            echo "Usage: $0 [all|system|secrets|apps|apps/<name>]"
            exit 1
            ;;
    esac

    log "Deployment complete!"

    # Show cluster status
    echo ""
    log "Cluster status:"
    kubectl get nodes
    echo ""
    kubectl get pods -A --field-selector=status.phase!=Running,status.phase!=Succeeded 2>/dev/null || true
}

main "$@"
