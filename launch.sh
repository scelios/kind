#!/bin/bash
mkdir -p src
cd src/
set -euo pipefail

DOMAIN="localhost"
REPO="github.com/scelios/kind"
CLUSTER_NAME="my-cluster"
IMG="znichola/expert-system:latest"

# This script launches the application with the necessary environment setup.

# create kind cluster
if kind get clusters | grep -q "^${CLUSTER_NAME}$"; then
    echo "Kind cluster '${CLUSTER_NAME}' already exists"
else
    echo "Creating kind cluster '${CLUSTER_NAME}'"
    kind create cluster --name "${CLUSTER_NAME}"
    echo "Kind cluster '${CLUSTER_NAME}' created successfully"
fi

# Export kubeconfig to a temporary file and set KUBECONFIG
KUBECONFIG_FILE="$(mktemp)"
if [ -f "${KUBECONFIG_FILE}" ]; then
    echo "kubeconfig already exists"
else
    echo "Creating kubeconfig file at ${KUBECONFIG_FILE}"
    kind get kubeconfig --name "${CLUSTER_NAME}" > "${KUBECONFIG_FILE}"
    export KUBECONFIG="${KUBECONFIG_FILE}"
    echo "KUBECONFIG set to ${KUBECONFIG}"
fi

# initialize operator-sdk (ensure REPO is set correctly)
if [ -f "PROJECT" ]; then
    echo "Already initialized"
else
    echo "Initializing operator-sdk"
    operator-sdk init --domain="${DOMAIN}" --repo="${REPO}"
    if [ $? -ne 0 ]; then
        echo "Failed to initialize operator-sdk"
        exit 1
    fi
    echo "operator-sdk initialized successfully"
fi

# Add API + Controller
if [ -d "api/v1alpha1" ]; then
    echo "API and Controller already exist"
else
    echo "Creating API and Controller"
    operator-sdk create api --group cache --version v1alpha1 --kind Memcached --resource --controller
    if [ $? -ne 0 ]; then
        echo "Failed to create API and Controller"
        exit 1
    fi
    echo "API and Controller created successfully"
fi

# Copy HelloWorld API and Controller from root directory (force overwrite)
echo "Copying HelloWorld types..."
cp ../helloworld_types.go api/v1alpha1/helloworld_types.go

echo "Copying HelloWorld controller..."
cp ../helloworld_controller.go internal/controller/helloworld_controller.go

# Update main.go with HelloWorld controller registration
if ! grep -q "HelloWorldReconciler" cmd/main.go; then
    echo "Updating main.go with HelloWorld controller..."
    cp ../main.go cmd/main.go
fi

# Generate deepcopy methods for HelloWorld
echo "Generating HelloWorld deepcopy methods..."
make generate
if [ $? -ne 0 ]; then
    echo "Failed to generate HelloWorld deepcopy methods"
    exit 1
fi

# Install OLM
# operator-sdk olm install
# if [ $? -ne 0 ]; then
#     echo "Failed to install OLM"
#     exit 1
# fi
# echo "OLM installed successfully"

# Launch the application (example command)
# make bundle IMG="${DOMAIN}/memcached-operator:v0.0.1" 
# make docker-build docker-push IMG="${DOMAIN}/memcached-operator:v0.0.1"

# load image into kind
# kind load docker-image memcached-operator:v0.0.1 --name my-cluster
# if [ $? -ne 0 ]; then
#     echo "Failed to load image into kind"
#     exit 1
# fi
# echo "Image loaded into kind successfully"

# # Run the operator from the bundle
# operator-sdk run bundle ${DOMAIN}/memcached-operator-bundle:v0.0.1
# if [ $? -ne 0 ]; then
#     echo "Failed to run the operator"
#     exit 1
# fi
# echo "Operator is running successfully"

# copy sample(s)
cp ../helloWorld.yaml config/samples/helloWorld.yaml

# Apply sample file (CRD may be created here). If this returns a non-zero
# exit code we preserve script failure semantics but continue to wait for CRD.
set +e
kubectl apply -f config/samples/helloWorld.yaml
_apply_rc=$?
set -e
if [ $_apply_rc -ne 0 ]; then
  echo "Initial apply returned $_apply_rc; continuing to wait for CRD to settle"
fi

# Wait for CRD to be fully registered by API server
kubectl wait --for=condition=Established crd/helloworlds.cache.localhost --timeout=120s

# Re-apply now that CRD is available (this will actually create the CR)
kubectl apply -f config/samples/helloWorld.yaml
if [ $? -ne 0 ]; then
    echo "Failed to apply helloWorld.yaml"
    exit 1
fi
echo "helloWorld.yaml applied successfully"
# Note: Replace the example commands and paths with actual application-specific commands as needed.

# make deploy IMG="${IMG}"

# if [ $? -ne 0 ]; then
#     echo "Failed to deploy the operator"
#     exit 1
# fi
# echo "Operator deployed successfully"

make install
if [ $? -ne 0 ]; then
    echo "Failed to install the operator"
    exit 1
fi
echo "Operator installed successfully"

# Run the operator in the background
echo "Starting operator..."
make run IMG="${DOMAIN}/memcached-operator:v0.0.1" > /tmp/operator.log 2>&1 &
OPERATOR_PID=$!
echo "Operator started (PID: $OPERATOR_PID)"

# Wait for operator to be ready
sleep 5

# Wait for service to be ready
echo "Waiting for HelloWorld service to be ready..."
kubectl wait --for=condition=Ready pod -l app=example-helloworld --timeout=30s -n default 2>/dev/null || true

# Start port-forward in background
echo "Setting up port-forward to HelloWorld service..."
kubectl port-forward svc/example-helloworld 8080:80 -n default > /tmp/port-forward.log 2>&1 &
PORT_FORWARD_PID=$!
echo "Port-forward started (PID: $PORT_FORWARD_PID)"

# Give it a moment to start
sleep 2

echo ""
echo "=========================================="
echo "✓ Application is ready!"
echo "✓ Open your browser to:"
echo "  → http://localhost:8080"
echo "=========================================="
echo ""
echo "Logs:"
echo "  Operator:     tail -f /tmp/operator.log"
echo "  Port-forward: tail -f /tmp/port-forward.log"
echo ""
echo "To stop everything, press Ctrl+C"
echo ""

# Keep script running and handle cleanup
trap "kill $OPERATOR_PID $PORT_FORWARD_PID 2>/dev/null; exit 0" SIGINT SIGTERM
wait

