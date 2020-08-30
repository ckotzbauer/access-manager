#!/bin/bash
set -eu

K8S_VERSION=$1

cd ..

GO111MODULE=off go get sigs.k8s.io/kind
kind create cluster --image kindest/node:v${K8S_VERSION}
kind get kubeconfig >e2e/kind-kubeconfig
export KUBECONFIG=e2e/kind-kubeconfig

make docker-build -e VERSION=latest
kind load docker-image ckotzbauer/access-manager:latest

make install deploy

sleep 10
OPERATOR_POD=$(kubectl get pod -l name=access-manager -o jsonpath='{.items[*].metadata.name}')
kubectl wait --for=condition=Ready pod/$OPERATOR_POD

kubectl create ns namespace1
kubectl create ns namespace2
kubectl create ns namespace3
kubectl create ns namespace4

kubectl label ns namespace1 ci=true

cd e2e
export KUBECONFIG=kind-kubeconfig
go test

kind delete cluster
