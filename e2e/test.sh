#!/bin/bash
set -eu

cd ..

GO111MODULE=off go get sigs.k8s.io/kind
kind create cluster
kind get kubeconfig >e2e/kind-kubeconfig
export KUBECONFIG=e2e/kind-kubeconfig

operator-sdk build ckotzbauer/access-manager:latest
kind load docker-image ckotzbauer/access-manager:latest

kubectl apply -f deploy/crds/rbacdefinition_crd.yaml
kubectl apply -f deploy/

sleep 5
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
