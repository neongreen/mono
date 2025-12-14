#!/bin/bash
# k8s-kubeconfig.sh - Fetch and configure kubeconfig for the mono cluster
#
# Usage:
#   ./cloud/scripts/k8s-kubeconfig.sh           # Fetch kubeconfig
#   eval $(./cloud/scripts/k8s-kubeconfig.sh)   # Fetch and export KUBECONFIG
#
# Output:
#   Prints export command for KUBECONFIG

set -euo pipefail

# Get server IP from DNS
get_server_ip() {
    host mono.cloud.artyom.me 2>/dev/null | grep -oE '[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+' | head -1 || {
        echo "Error: Could not resolve mono.cloud.artyom.me" >&2
        exit 1
    }
}

server_ip="${MONO_SERVER_IP:-$(get_server_ip)}"
kubeconfig_dir="$HOME/.kube"
kubeconfig_file="$kubeconfig_dir/mono-config"

mkdir -p "$kubeconfig_dir"

ssh -o StrictHostKeyChecking=accept-new "emily@$server_ip" \
    'sudo cat /etc/rancher/k3s/k3s.yaml' | \
    sed "s/127.0.0.1/$server_ip/" > "$kubeconfig_file"

chmod 600 "$kubeconfig_file"

echo "export KUBECONFIG=$kubeconfig_file"
