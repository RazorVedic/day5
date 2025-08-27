#!/bin/bash

# Deployment script for the Product API

set -e

# Variables
NAMESPACE=${NAMESPACE:-"day5"}
RELEASE_NAME=${RELEASE_NAME:-"day5"}
HELM_CHART_PATH="./deployments/helm/day5"

echo "Deploying Product API..."

# Check if namespace exists, create if not
if ! kubectl get namespace $NAMESPACE >/dev/null 2>&1; then
    echo "Creating namespace: $NAMESPACE"
    kubectl create namespace $NAMESPACE
fi

# Deploy using Helm
echo "Deploying with Helm..."
helm upgrade --install $RELEASE_NAME $HELM_CHART_PATH \
    --namespace $NAMESPACE \
    --wait \
    --timeout 300s

echo "Deployment completed successfully!"

# Show deployment status
echo "Checking deployment status..."
kubectl get pods -n $NAMESPACE
kubectl get svc -n $NAMESPACE

echo "To access the API, run:"
echo "kubectl port-forward -n $NAMESPACE svc/$RELEASE_NAME 8080:80"
