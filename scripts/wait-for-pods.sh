#!/bin/bash

# wait-for-pods.sh - Wait for Kubernetes pods to be ready
# Usage: ./wait-for-pods.sh <namespace> <label-selector> <timeout-seconds>

set -e

NAMESPACE=${1:-day5}
LABEL_SELECTOR=${2:-""}
TIMEOUT=${3:-120}
POLL_INTERVAL=5

if [ -z "$LABEL_SELECTOR" ]; then
    echo "❌ Error: Label selector is required"
    echo "Usage: $0 <namespace> <label-selector> <timeout-seconds>"
    echo "Example: $0 day5 'app=day5' 120"
    exit 1
fi

echo "🔍 Waiting for pods with label '$LABEL_SELECTOR' in namespace '$NAMESPACE' to be ready..."
echo "⏰ Timeout: ${TIMEOUT}s"

start_time=$(date +%s)

while true; do
    current_time=$(date +%s)
    elapsed=$((current_time - start_time))
    
    if [ $elapsed -ge $TIMEOUT ]; then
        echo "❌ Timeout reached (${TIMEOUT}s). Pods are not ready."
        echo ""
        echo "📊 Current pod status:"
        kubectl get pods -n "$NAMESPACE" -l "$LABEL_SELECTOR" --no-headers 2>/dev/null || echo "No pods found"
        echo ""
        echo "🔍 Pod details:"
        kubectl describe pods -n "$NAMESPACE" -l "$LABEL_SELECTOR" 2>/dev/null || echo "No pods found"
        exit 1
    fi
    
    # Get pod information
    pod_info=$(kubectl get pods -n "$NAMESPACE" -l "$LABEL_SELECTOR" --no-headers 2>/dev/null || echo "")
    
    if [ -z "$pod_info" ]; then
        echo "⏳ No pods found yet (${elapsed}s elapsed)..."
        sleep $POLL_INTERVAL
        continue
    fi
    
    # Count total pods and ready pods
    total_pods=$(echo "$pod_info" | wc -l | tr -d ' ')
    ready_pods=0
    
    # Check each pod's readiness
    while IFS= read -r line; do
        if [ -n "$line" ]; then
            # Extract the READY column (2nd column) - format is "ready/total"
            ready_status=$(echo "$line" | awk '{print $2}')
            ready_count=$(echo "$ready_status" | cut -d'/' -f1)
            total_count=$(echo "$ready_status" | cut -d'/' -f2)
            
            if [ "$ready_count" = "$total_count" ] && [ "$ready_count" != "0" ]; then
                ready_pods=$((ready_pods + 1))
            fi
        fi
    done <<< "$pod_info"
    
    echo "📊 Pod status: $ready_pods/$total_pods ready (${elapsed}s elapsed)"
    
    # Display current pod status
    echo "$pod_info" | while IFS= read -r line; do
        if [ -n "$line" ]; then
            pod_name=$(echo "$line" | awk '{print $1}')
            ready_status=$(echo "$line" | awk '{print $2}')
            status=$(echo "$line" | awk '{print $3}')
            echo "   • $pod_name: $ready_status $status"
        fi
    done
    
    # Check if all pods are ready
    if [ "$ready_pods" -eq "$total_pods" ] && [ "$total_pods" -gt 0 ]; then
        echo "✅ All $total_pods pods are ready!"
        echo ""
        echo "📊 Final pod status:"
        kubectl get pods -n "$NAMESPACE" -l "$LABEL_SELECTOR"
        exit 0
    fi
    
    sleep $POLL_INTERVAL
done
