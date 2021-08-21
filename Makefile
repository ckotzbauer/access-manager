# Current Operator version
ifeq (,${VERSION})
BIN_VERSION=latest
else
BIN_VERSION=${VERSION}
endif

# Image URL to use all building/pushing image targets
IMG ?= ckotzbauer/access-manager
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"

# default k8s version for e2e tests
K8S_VERSION ?= 1.22.0

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

all: manager

# Run unit-tests
test: generate fmt vet manifests
	go test access-manager/pkg/reconciler -coverprofile cover.out

# Run e2e-tests
e2e-test: kind
	cd e2e && \
	bash test.sh $(KIND) $(K8S_VERSION)

# Build manager binary
manager: generate fmt vet
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags '-w -X=main.Version=$(BIN_VERSION)' -o bin/manager main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet manifests
	go run ./main.go

# Install CRDs into a cluster
install: manifests
	kubectl apply -f config/crd

# Uninstall CRDs from a cluster
uninstall: manifests
	kubectl delete -f config/rbac
	kubectl delete -f config/manager
	kubectl delete -f config/crd

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests
	kubectl apply -f config/rbac
	kubectl apply -f config/manager

# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

# Generate code
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile= paths="./..."

# find or download controller-gen
# download controller-gen if necessary
controller-gen:
ifeq (, $(shell which controller-gen))
	@{ \
	set -e ;\
	CONTROLLER_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$CONTROLLER_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.6.0 ;\
	rm -rf $$CONTROLLER_GEN_TMP_DIR ;\
	}
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif

# find or download kind
# download kind if necessary
kind:
ifeq (, $(shell which kind))
	@{ \
	set -e ;\
	KIND_TMP_DIR=$$(mktemp -d) ;\
	cd $$KIND_TMP_DIR ;\
	go mod init tmp ;\
	go get sigs.k8s.io/kind@v0.11.1 ;\
	rm -rf $$KIND_TMP_DIR ;\
	}
KIND=$(GOBIN)/kind
else
KIND=$(shell which kind)
endif

