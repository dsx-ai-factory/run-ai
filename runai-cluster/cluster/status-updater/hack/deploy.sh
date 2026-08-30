#!/usr/bin/env bash
set -x

DOCKER_REPO="gcr.io/run-ai-lab/"
IMAGE_NAME="status-updater"
TAG=${TAG:-$USER} # You can also put it in the git ignored .env
IMAGE_FULL_NAME="${DOCKER_REPO}${IMAGE_NAME}:${TAG}"
RELEASE_VER=`git rev-parse --short HEAD`

DEBUG=
while getopts d flag
do
    case "${flag}" in
        d) DEBUG=1;;
        *) DEBUG=0;;
    esac
done

if [ -f .env ]; then
  export "$(cat .env | xargs)"
fi

make image DEBUG=$DEBUG &&
docker tag runai/status-updater:"$RELEASE_VER" "$IMAGE_FULL_NAME" &&
docker push "$IMAGE_FULL_NAME" &&
kubectl rollout restart deployment status-updater -n runai
