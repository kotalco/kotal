#!/bin/sh

set -e

K8S_PROVIDER="${K8S_PROVIDER:-kind}"


function cleanup {
  echo "🧽 Cleaning up"
  if [ "$K8S_PROVIDER" == "minikube" ]
    then
      minikube delete --all
    else
      kind delete clusters --all
  fi
}

trap cleanup EXIT

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


if [ "$K8S_PROVIDER" == "minikube" ]
then
# minikube cluster versions
VERSIONS=("1.19.0" "1.20.0" "1.21.0" "1.22.0" "1.23.0" "1.24.0" "1.25.0" "1.26.0")
echo "🗑 Deleting all Minikube clusters"
minikube delete --all
else
# kind cluster versions
# https://hub.docker.com/r/kindest/node/tags
VERSIONS=("1.19.16" "1.20.15" "1.21.14" "1.22.15" "1.23.13" "1.24.7" "1.25.3" "1.26.0")
echo "🗑 Deleting all Kind clusters"
kind delete clusters --all
fi

for VERSION in "${VERSIONS[@]}"
do
  echo "Testing Kotal operator in kubernetes v$VERSION"
    # start Kubernetes in Docker with this kubernetes version
    echo "Creating cluster"
    if [ "$K8S_PROVIDER" == "minikube" ]
    then
      minikube start --kubernetes-version=v${VERSION}
    else
	    kind create cluster --image=kindest/node:v${VERSION}
    fi
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
      if [ "$K8S_PROVIDER" == "minikube" ]
      then
        make minikube
      else
        make kind
      fi
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
    if [ "$K8S_PROVIDER" == "minikube" ]
    then
      minikube delete
    else
      kind delete cluster
    fi
done