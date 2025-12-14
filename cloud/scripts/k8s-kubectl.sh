#!/bin/bash
# k8s-kubectl.sh - Run kubectl against the mono cluster
#
# Usage:
#   ./cloud/scripts/k8s-kubectl.sh get pods -A
#   ./cloud/scripts/k8s-kubectl.sh logs -n n8n deployment/n8n
#
# This script auto-fetches kubeconfig if needed, then runs kubectl.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
KUBECONFIG_FILE="$HOME/.kube/mono-config"

# Fetch kubeconfig if missing or older than 1 hour
if [[ ! -f "$KUBECONFIG_FILE" ]] || [[ $(find "$KUBECONFIG_FILE" -mmin +60 2>/dev/null) ]]; then
    "$SCRIPT_DIR/k8s-kubeconfig.sh" >/dev/null
fi

export KUBECONFIG="$KUBECONFIG_FILE"
exec kubectl "$@"
