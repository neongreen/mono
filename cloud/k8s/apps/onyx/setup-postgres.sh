#!/usr/bin/env bash
# Setup PostgreSQL database for Onyx in the existing n8n postgres instance

set -e

echo "Setting up Onyx database in n8n's PostgreSQL..."

# Get the postgres pod name
POSTGRES_POD=$(kubectl get pods -n n8n -l service=postgres -o jsonpath='{.items[0].metadata.name}')

if [ -z "$POSTGRES_POD" ]; then
  echo "Error: Could not find postgres pod in n8n namespace"
  exit 1
fi

echo "Found postgres pod: $POSTGRES_POD"

# Create the onyx database
echo "Creating 'onyx' database..."
kubectl exec -n n8n "$POSTGRES_POD" -- psql -U postgres -c "
  SELECT 'CREATE DATABASE onyx'
  WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'onyx')\gexec
"

echo "Database 'onyx' is ready!"
echo ""
echo "Next steps:"
echo "1. Add secrets to keychain (see README.md)"
echo "2. Create k8s secret: fnox exec -- kubectl create secret ..."
echo "3. Deploy Onyx: kubectl apply -f cloud/k8s/onyx/"
