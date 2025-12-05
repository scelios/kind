# Kubernetes Operator with Kind

A learning project to practice Kubernetes Operator SDK and Kind (Kubernetes in Docker). This operator deploys and manages my [expert-system](https://github.com/znichola/expert-system) school project in a local Kubernetes cluster.

## Overview

This project demonstrates:
- Creating a custom Kubernetes operator using Operator SDK
- Managing custom resources
- Deploying applications through Kubernetes controllers
- Running a local Kubernetes cluster with Kind


## Prerequisites

The following tools must be installed on your system:
- `kind` - Kubernetes in Docker
- `kubectl` - Kubernetes CLI
- `operator-sdk` - Operator SDK CLI
- `docker`
- `go`

The `launch.sh` script will check for these prerequisites before running.

## Quick Start

### Launch the Application

```bash
./launch.sh
```

This script will:

    1.Create a Kind cluster (if not already exists)
    2.Initialize the Operator SDK project
    3.Generate the HelloWorld API and controller
    4.Deploy the operator to the cluster
    5.Deploy the expert-system application as pods
    6.Set up port-forwarding to access the app

Access the Application
Once launched, open your browser to:
http://localhost:8080

For more information about the expert-system application itself, see the expert-system README.

Stop the Application
Press Ctrl+C in the terminal where launch.sh is running to stop all processes.

Clean Up
To completely remove the cluster, containers, and build artifacts:

```bash
./cleanup.sh
```

This will:

    1.Delete the Kind cluster
    2.Stop all running processes (operator, port-forward)
    3.Remove Docker containers
    4.Clean up temporary files
    5.Optionally remove the src directory (for a fresh start)

What's Happening Under the Hood
Kind Cluster: A local Kubernetes cluster runs in Docker containers
Custom Resource Definition (CRD): Defines the HelloWorld resource type
Operator Controller: Watches for HelloWorld resources and reconciles state
Deployment: Creates pods running the expert-system Docker image (scelios/expert-system:latest)
Service: Exposes the pods on port 7711
Port-Forward: Makes the service accessible on localhost:8080


