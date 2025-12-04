# Minimal Makefile for operator-sdk-style workflow (install CRDs, run locally, build image, deploy to kind)

IMG ?= helloworld-operator:0.1.0
KIND_CLUSTER ?= kind
CRD_DIR := config/crd/bases

GEN := $(shell command -v controller-gen 2>/dev/null)

.PHONY: all install uninstall build run docker-build kind-load deploy deploy-image clean
.PHONY: generate manifests

all: build

install:
	kubectl apply -f $(CRD_DIR)

uninstall:
	kubectl delete -f $(CRD_DIR) --ignore-not-found

build:
	go build ./...

# run controller locally (chooses common main locations)
run:
	@if [ -f ./cmd/manager/main.go ]; then \
		go run ./cmd/manager/main.go; \
	elif [ -f ./main.go ]; then \
		go run ./main.go; \
	else \
		go run ./...; \
	fi

docker-build:
	docker build -t $(IMG) .

# load built image into the kind cluster
kind-load: docker-build
	kind load docker-image $(IMG) --name $(KIND_CLUSTER)

# deploy using kustomize (config/default)
deploy:
	kustomize build config/default | kubectl apply -f -

# set image (optional) then deploy to cluster (useful for kind)
deploy-image: kind-load
	# try to update kustomize image (will not fail the make if it does)
	-kustomize edit set image controller=$(IMG)
	kustomize build config/default | kubectl apply -f -

clean:
	go clean

generate:
	@if [ -z "$(GEN)" ]; then \
		echo "controller-gen not found in PATH; install with: go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.12.0"; \
		exit 1; \
	fi
	controller-gen object:headerFile="hack/boilerplate.go.txt" paths="./api/..."

manifests:
	@if [ -z "$(GEN)" ]; then \
		echo "controller-gen not found in PATH; install with: go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.12.0"; \
		exit 1; \
	fi
	controller-gen crd:trivialVersions=true paths="./api/..." output:crd:dir=config/crd/bases