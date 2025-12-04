#!/bin/bash

# This script launches the application with the necessary environment setup.
# install kind
if [ -f /usr/local/bin/kind ]; then
    echo "kind already installed"
else
    curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.20.0/kind && chmod +x ./kind && sudo mv ./kind /usr/local/bin/
    if [ $? -ne 0 ]; then
        echo "Failed to install kind"
        exit 1
    fi
fi
# kubectl
if [ -f /usr/local/bin/kubectl ]; then
    echo "kubectl already installed"
else
    curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl" && chmod +x kubectl && sudo mv kubectl /usr/local/bin/
    if [ $? -ne 0 ]; then
        echo "Failed to install kubectl"
        exit 1
    fi
fi

# operator-sdk (example)
if [ -f /usr/local/bin/operator-sdk ]; then
    echo "operator-sdk already installed"
else
    curl -LO https://github.com/operator-framework/operator-sdk/releases/download/v1.26.0/operator-sdk_linux_amd64 && chmod +x operator-sdk_linux_amd64 && sudo mv operator-sdk_linux_amd64 /usr/local/bin/operator-sdk
    if [ $? -ne 0 ]; then
        echo "Failed to install operator-sdk"
        exit 1
    fi
fi

# create kind cluster
kind create cluster --name my-cluster
if [ $? -ne 0 ]; then
    echo "Failed to create kind cluster"
    exit 1
fi
echo "Kind cluster 'my-cluster' created successfully"
# Set KUBECONFIG environment variable
export KUBECONFIG="$(kind get kubeconfig-path --name="my-cluster")"
echo "KUBECONFIG set to $KUBECONFIG"

# initialize operator-sdk
operator-sdk init --domain=localhost 
if [ $? -ne 0 ]; then
    echo "Failed to initialize operator-sdk"
    exit 1
fi
echo "operator-sdk initialized successfully"

# Add API + Controller
operator-sdk create api --group cache --version v1alpha1 --kind Memcached --resource --controller
if [ $? -ne 0 ]; then
    echo "Failed to create API and Controller"
    exit 1
fi
echo "API and Controller created successfully"

# Install OLM
operator-sdk olm install
if [ $? -ne 0 ]; then
    echo "Failed to install OLM"
    exit 1
fi
echo "OLM installed successfully"

# Launch the application (example command)
make bundle IMG="example.com/memcached-operator:v0.0.1"
make docker-build docker-push IMG="example.com/memcached-operator:v0.0.1"

