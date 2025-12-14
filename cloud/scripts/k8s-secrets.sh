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

log() { echo "[secrets] $*"; }
error() { echo "[secrets] ERROR: $*" >&2; }

# Create secret if it doesn't exist, or update if --force
create_secret() {
    local namespace="$1"
    local name="$2"
    shift 2
    local args=("$@")

    # Ensure namespace exists
    kubectl create namespace "$namespace" 2>/dev/null || true

    # Check if secret exists
    if kubectl get secret "$name" -n "$namespace" &>/dev/null; then
        log "Secret $namespace/$name already exists, updating..."
        kubectl delete secret "$name" -n "$namespace"
    fi

    log "Creating secret $namespace/$name..."
    kubectl create secret generic "$name" -n "$namespace" "${args[@]}"
}

create_n8n_secrets() {
    log "Creating n8n secrets..."

    mise x -- fnox exec -- bash -c '
        kubectl create namespace n8n 2>/dev/null || true
        kubectl delete secret postgres-secret -n n8n 2>/dev/null || true
        kubectl create secret generic postgres-secret -n n8n \
            --from-literal=POSTGRES_USER=postgres \
            --from-literal=POSTGRES_PASSWORD="$N8N_POSTGRES_ADMIN_PASSWORD" \
            --from-literal=POSTGRES_NON_ROOT_USER=n8n \
            --from-literal=POSTGRES_NON_ROOT_PASSWORD="$N8N_POSTGRES_N8N_PASSWORD"
    '
    log "n8n secrets created."
}

create_onyx_secrets() {
    log "Creating onyx secrets..."

    mise x -- fnox exec -- bash -c '
        kubectl create namespace onyx 2>/dev/null || true
        kubectl delete secret onyx-secrets -n onyx 2>/dev/null || true
        kubectl create secret generic onyx-secrets -n onyx \
            --from-literal=POSTGRES_PASSWORD="$ONYX_POSTGRES_PASSWORD" \
            --from-literal=SECRET="$ONYX_SECRET" \
            --from-literal=S3_AWS_SECRET_ACCESS_KEY="$ONYX_MINIO_ROOT_PASSWORD"
    '
    log "onyx secrets created."
}

create_coder_secrets() {
    log "Creating coder secrets..."

    mise x -- fnox exec -- bash -c '
        kubectl create namespace coder 2>/dev/null || true
        kubectl delete secret coder-secrets -n coder 2>/dev/null || true
        kubectl create secret generic coder-secrets -n coder \
            --from-literal=POSTGRES_PASSWORD="$CODER_POSTGRES_PASSWORD" \
            --from-literal=POSTGRES_ADMIN_PASSWORD="$N8N_POSTGRES_ADMIN_PASSWORD"
    '
    log "coder secrets created."
}

main() {
    local target="${1:-all}"

    # Ensure kubeconfig is set up
    if [[ -z "${KUBECONFIG:-}" ]]; then
        eval "$("$SCRIPT_DIR/k8s-kubeconfig.sh")"
    fi

    case "$target" in
        all)
            create_n8n_secrets
            create_onyx_secrets
            create_coder_secrets
            ;;
        n8n)
            create_n8n_secrets
            ;;
        onyx)
            create_onyx_secrets
            ;;
        coder)
            create_coder_secrets
            ;;
        *)
            error "Unknown target: $target"
            echo "Usage: $0 [all|n8n|onyx|coder]"
            exit 1
            ;;
    esac

    log "All secrets created successfully."
}

main "$@"
