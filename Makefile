
# Image URL to use all building/pushing image targets
IMG ?= kotalco/kotal:v0.1-alpha.6

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

all: manager

# Run tests
# remove zz_generated files from test coverage report
test: generate fmt vet manifests
	KUBEBUILDER_CONTROLPLANE_START_TIMEOUT=100s ACK_GINKGO_DEPRECATIONS=1.16.4 go test -v -coverprofile cover.out.tmp ./...
	cat cover.out.tmp | grep -v zz_generated > cover.out

# test operator on multiple k8s cluster versions
# KOTAL_VERSION is released kotal image tag
# K8S_PROVIDER is k8s cluster provider: kind or minikube
# KOTAL_VERSION=$IMG K8S_PROVIDER=minikube make test-multi
# KOTAL_VERSION=$IMG K8S_PROVIDER=kind make test-multi
# make test-multi
.SILENT: test-multi
test-multi:
	chmod +x multi.sh
	./multi.sh

cover:
	go tool cover -html=cover.out

# Build manager binary
manager: generate fmt vet
	go build -o bin/manager main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet manifests
	ENABLE_WEBHOOKS=false go run ./main.go

# Install CRDs into a cluster
install: manifests
	kustomize build config/crd | kubectl apply -f -

# Uninstall CRDs from a cluster
uninstall: manifests
	kustomize build config/crd | kubectl delete -f -

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests
	cd config/manager && kustomize edit set image controller=${IMG}
	kustomize build config/default | kubectl apply -f -

# output manifest files for the release
release: manifests
	cd config/manager && kustomize edit set image controller=${IMG}
	kustomize build config/default > kotal.yaml

# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

# Generate code
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

# Build the docker image
docker-build: test
	docker build . -t ${IMG}

# load image into kind
kind-load:
	kind load docker-image ${IMG}

# Build the docker image
kind: kind-load deploy

# load image into minikube registry
minikube-load:
	minikube image load ${IMG}
	minikube cache reload

# Build the docker image
minikube: minikube-load deploy


# Push the docker image
docker-push:
	docker push ${IMG}

# find or download controller-gen
# download controller-gen if necessary
controller-gen:
ifeq (, $(shell which controller-gen))
	@{ \
	set -e ;\
	CONTROLLER_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$CONTROLLER_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.5.0 ;\
	rm -rf $$CONTROLLER_GEN_TMP_DIR ;\
	}
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif
