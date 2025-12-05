#!/bin/bash
mkdir -p src
cd src/
set -euo pipefail

DOMAIN="localhost"
REPO="github.com/scelios/kind"
CLUSTER_NAME="my-cluster"
IMG="scelios/expert-system:latest"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Check prerequisites
echo "Checking prerequisites..."
for cmd in kind kubectl operator-sdk docker; do
    if ! command_exists "$cmd"; then
        echo -e "${RED}✗ $cmd is not installed${NC}"
        exit 1
    fi
done
echo -e "${GREEN}✓ All prerequisites installed${NC}"

# Create kind cluster
if kind get clusters | grep -q "^${CLUSTER_NAME}$"; then
    echo "Kind cluster '${CLUSTER_NAME}' already exists"
else
    echo "Creating kind cluster '${CLUSTER_NAME}'..."
    kind create cluster --name "${CLUSTER_NAME}"
    echo -e "${GREEN}✓ Kind cluster '${CLUSTER_NAME}' created${NC}"
fi

# Set up kubeconfig
KUBECONFIG_FILE="$(mktemp)"
kind get kubeconfig --name "${CLUSTER_NAME}" > "${KUBECONFIG_FILE}"
export KUBECONFIG="${KUBECONFIG_FILE}"
echo -e "${GREEN}✓ KUBECONFIG set${NC}"

# Initialize operator-sdk
if [ ! -f "PROJECT" ]; then
    echo "Initializing operator-sdk..."
    operator-sdk init --domain="${DOMAIN}" --repo="${REPO}"
    echo -e "${GREEN}✓ Operator-sdk initialized${NC}"
fi

# Create API and Controller
if [ ! -d "api/v1alpha1" ]; then
    echo "Creating API and Controller..."
    operator-sdk create api --group cache --version v1alpha1 --kind Memcached --resource --controller
    echo -e "${GREEN}✓ API and Controller created${NC}"
fi

# Copy HelloWorld files
echo "Copying HelloWorld files..."
cp ../helloworld_types.go api/v1alpha1/helloworld_types.go
cp ../helloworld_controller.go internal/controller/helloworld_controller.go

if ! grep -q "HelloWorldReconciler" cmd/main.go; then
    cp ../main.go cmd/main.go
fi

# Generate code
echo "Generating code..."
make generate
echo -e "${GREEN}✓ Code generated${NC}"

# Copy sample
cp ../helloWorld.yaml config/samples/helloWorld.yaml

# Apply CRD and sample
echo "Applying HelloWorld CRD and sample..."
set +e
kubectl apply -f config/samples/helloWorld.yaml
_apply_rc=$?
set -e

if [ $_apply_rc -ne 0 ]; then
    echo "Initial apply returned $_apply_rc; waiting for CRD..."
fi

kubectl wait --for=condition=Established crd/helloworlds.cache.localhost --timeout=120s
kubectl apply -f config/samples/helloWorld.yaml
echo -e "${GREEN}✓ HelloWorld sample applied${NC}"

# Build and load Docker image
echo "Building Docker image..."
make docker-build IMG="${IMG}"

echo "Loading image into kind..."
kind load docker-image "${IMG}" --name "${CLUSTER_NAME}"
echo -e "${GREEN}✓ Image loaded${NC}"

# Deploy operator
echo "Deploying operator..."
make install
make deploy IMG="${IMG}"
echo -e "${GREEN}✓ Operator deployed${NC}"

# Run operator in background
echo "Starting operator..."
make run IMG="${IMG}" > /tmp/operator.log 2>&1 &
OPERATOR_PID=$!
echo -e "${GREEN}✓ Operator started (PID: $OPERATOR_PID)${NC}"

sleep 5

# Wait for pods
echo "Waiting for HelloWorld pods..."
kubectl wait --for=condition=Ready pod -l app=example-helloworld --timeout=30s -n default 2>/dev/null || true

# Start port-forward (change 8080:7711 to forward port 8080 to the app's 7711)
echo "Setting up port-forward..."
kubectl port-forward svc/example-helloworld 8080:7711 -n default > /tmp/port-forward.log 2>&1 &
PORT_FORWARD_PID=$!
sleep 2

echo ""
echo "=========================================="
echo -e "${GREEN}✓ Application is ready!${NC}"
echo "=========================================="
echo "Open your browser to:"
echo -e "  ${YELLOW}→ http://localhost:8080${NC}"
echo ""
echo "Logs:"
echo "  tail -f /tmp/operator.log"
echo "  tail -f /tmp/port-forward.log"
echo ""
echo "To stop: press Ctrl+C"
echo "=========================================="
echo ""

trap "kill $OPERATOR_PID $PORT_FORWARD_PID 2>/dev/null; exit 0" SIGINT SIGTERM
wait

