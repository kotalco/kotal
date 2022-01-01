#!/bin/sh

set -e

if ! docker info > /dev/null 2>&1; then
  echo "Docker isn't running"
  echo "Start docker, then try again!"
  exit 1
fi

if [ "$KOTAL_VERSION" == "" ]
then
  # build manager image once
  echo "Building docker image"
  make docker-build
fi

for VERSION in '1.19.11' '1.20.7' '1.21.1' '1.22.4' '1.23.0'
do
  echo "Testing Kotal operator in kubernetes v$VERSION"
    # start Kubernetes in Docker with this kubernetes version
    echo "Creating cluster"
	  kind create cluster --image=kindest/node:v${VERSION}
    # install cert-manager
    echo "Installing cert manager"
    kubectl apply -f https://github.com/jetstack/cert-manager/releases/download/v1.5.3/cert-manager.yaml
    # load image and deploy manifests
    echo "⏳ Waiting for cert manager to be up and running"
    sleep 5
    kubectl wait -n cert-manager --for=condition=available deployments/cert-manager --timeout=600s
    kubectl wait -n cert-manager --for=condition=available deployments/cert-manager-cainjector --timeout=600s
    kubectl wait -n cert-manager --for=condition=available deployments/cert-manager-webhook --timeout=600s
    echo "🚀 Cert manager is up and running"

    if [ "$KOTAL_VERSION" == "" ]
    then
      echo "Installing Kotal custom resources"
      echo "Deploying Kotal controller manager"
      make kind
    else
      kubectl apply -f https://github.com/kotalco/kotal/releases/download/$KOTAL_VERSION/kotal.yaml
    fi

    echo "⏳ Waiting for kotal controller manager to be up and running"
    kubectl wait -n kotal --for=condition=available deployments/controller-manager --timeout=600s
    echo "🚀 Kotal is up and running"

    echo "🔥 Running tests"
    # test against image
    USE_EXISTING_CLUSTER=true make test
    # delete cluster
    echo "🎉 All tests has been passed"

    echo "🔥 Deleting kubernetes cluster v$VERSION"
    kind delete cluster
done