#!/bin/bash
# k8s-secrets.sh - Create k8s secrets from fnox values
#
# Usage:
#   ./cloud/scripts/k8s-secrets.sh        # Create all secrets
#   ./cloud/scripts/k8s-secrets.sh n8n    # Create only n8n secrets
#
# Requires: fnox with age key available

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CLOUD_DIR="$(dirname "$SCRIPT_DIR")"

log() { echo "[secrets] $*"; }
error() { echo "[secrets] ERROR: $*" >&2; }

# Create or update a secret
create_secret() {
    local namespace="$1"
    local name="$2"
    shift 2
    local args=("$@")

    kubectl create namespace "$namespace" 2>/dev/null || true

    if kubectl get secret "$name" -n "$namespace" &>/dev/null; then
        log "Secret $namespace/$name already exists, updating..."
        kubectl delete secret "$name" -n "$namespace"
    fi

    log "Creating secret $namespace/$name..."
    kubectl create secret generic "$name" -n "$namespace" "${args[@]}"
}

# Shared postgres secret - used by all apps
create_postgres_secrets() {
    log "Creating shared postgres secrets..."

    mise x -- fnox -c "$CLOUD_DIR/fnox.toml" exec -- bash -c '
        # Create secret in postgres namespace (for the postgres deployment itself)
        kubectl create namespace postgres 2>/dev/null || true
        kubectl delete secret postgres-secret -n postgres 2>/dev/null || true
        kubectl create secret generic postgres-secret -n postgres \
            --from-literal=POSTGRES_PASSWORD="$POSTGRES_PASSWORD"

        # Create secret in n8n namespace (for n8n to connect)
        kubectl create namespace n8n 2>/dev/null || true
        kubectl delete secret postgres-secret -n n8n 2>/dev/null || true
        kubectl create secret generic postgres-secret -n n8n \
            --from-literal=POSTGRES_PASSWORD="$POSTGRES_PASSWORD"

        # Create secret in coder namespace
        kubectl create namespace coder 2>/dev/null || true
        kubectl delete secret coder-secrets -n coder 2>/dev/null || true
        kubectl create secret generic coder-secrets -n coder \
            --from-literal=POSTGRES_PASSWORD="$POSTGRES_PASSWORD"

        # Create secret in onyx namespace
        kubectl create namespace onyx 2>/dev/null || true
        kubectl delete secret onyx-secrets -n onyx 2>/dev/null || true
        kubectl create secret generic onyx-secrets -n onyx \
            --from-literal=POSTGRES_PASSWORD="$POSTGRES_PASSWORD" \
            --from-literal=SECRET="$ONYX_SECRET" \
            --from-literal=S3_AWS_SECRET_ACCESS_KEY="$ONYX_MINIO_ROOT_PASSWORD"
    '
    log "All postgres secrets created."
}

main() {
    local target="${1:-all}"

    # Ensure kubeconfig is set up
    if [[ -z "${KUBECONFIG:-}" ]]; then
        eval "$("$SCRIPT_DIR/k8s-kubeconfig.sh")"
    fi

    case "$target" in
        all|postgres)
            create_postgres_secrets
            ;;
        *)
            error "Unknown target: $target"
            echo "Usage: $0 [all|postgres]"
            exit 1
            ;;
    esac

    log "All secrets created successfully."
}

main "$@"
