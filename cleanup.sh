#!/bin/bash
set -euo pipefail

echo "🧹 Cleaning up Kind cluster and operator resources..."
echo ""

# Kill any port-forward processes
echo "Stopping port-forward processes..."
pkill -f "kubectl port-forward" || true
sleep 1

# Kill operator process if running
echo "Stopping operator..."
pkill -f "make run" || true
sleep 1

# Delete kind cluster
CLUSTER_NAME="my-cluster"
if kind get clusters | grep -q "^${CLUSTER_NAME}$"; then
    echo "Deleting kind cluster '${CLUSTER_NAME}'..."
    kind delete cluster --name "${CLUSTER_NAME}"
    echo "Kind cluster '${CLUSTER_NAME}' deleted successfully"
else
    echo "Kind cluster '${CLUSTER_NAME}' not found (already deleted)"
fi

# Clean up temporary kubeconfig files
echo "Cleaning up temporary kubeconfig files..."
rm -f /tmp/kind-* 2>/dev/null || true
rm -f /tmp/kubeconfig-* 2>/dev/null || true

# Clean up src directory (optional - uncomment if you want to rebuild from scratch)
# echo "Cleaning up src directory..."
# rm -rf src/

# Clean up temporary files
echo "Cleaning up temporary files..."
rm -f /tmp/operator.log
rm -f /tmp/port-forward.log

# Clean up Docker images (optional - uncomment to remove Docker images)
echo "Removing Docker images related to this operator..."
docker rmi znichola/expert-system:latest 2>/dev/null || true
docker rmi -f $(docker images --format "{{.Repository}}:{{.Tag}}" | grep -E "(memcached|helloworld|operator)" | head -20) 2>/dev/null || true

echo ""
echo "=========================================="
echo "✓ Cleanup completed!"
echo "=========================================="
echo ""
echo "What was cleaned:"
echo "  ✓ Kind cluster deleted"
echo "  ✓ Operator and port-forward processes stopped"
echo "  ✓ Temporary files removed"
echo ""
echo "Optional: To also remove Docker images, uncomment the Docker section in this script"
echo ""