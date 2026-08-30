#!/usr/bin/env bash
IMAGE_VERSION="$(cat VERSION)"
make reset
make docker 
kind load docker-image runai/runai-admission-controller:${IMAGE_VERSION} --name k8s-local
sh ./deploy/ssl.sh
export DEPLOY_VERSION=$(uuidgen)
sed -e "s|IMAGE_VERSION|${IMAGE_VERSION}|g" ./deploy/deployment-template.yaml > ./deploy/deployment.yaml
make deploy
kubectl delete -n runai ValidatingWebhookConfiguration -l app=runai,deploy-version!=${DEPLOY_VERSION}
kubectl delete -n runai MutatingWebhookConfiguration -l app=runai,deploy-version!=${DEPLOY_VERSION}
