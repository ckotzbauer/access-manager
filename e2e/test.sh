#!/bin/bash
set -eu

KIND=$1
K8S_VERSION=$2

cd ..

${KIND} create cluster --image kindest/node:v${K8S_VERSION}
${KIND} get kubeconfig >e2e/kind-kubeconfig
export KUBECONFIG=e2e/kind-kubeconfig

docker build -t ckotzbauer/access-manager:latest .
${KIND} load docker-image ckotzbauer/access-manager:latest

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

kubectl logs $OPERATOR_POD

${KIND} delete cluster
