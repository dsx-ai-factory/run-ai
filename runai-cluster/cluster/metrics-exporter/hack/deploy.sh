#!/usr/bin/env bash

# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

set -x

DOCKER_REPO="gcr.io/run-ai-lab/"
IMAGE_NAME="metrics-exporter"
TAG=${TAG:-$USER} # You can also put it in the git ignored .env
IMAGE_FULL_NAME="${DOCKER_REPO}${IMAGE_NAME}:${TAG}"
RELEASE_VER=`git rev-parse --short HEAD`

DEBUG=
while getopts d flag
do
    case "${flag}" in
        d) DEBUG=1;;
    esac
done

if [ -f .env ]; then
  export $(cat .env | xargs)
fi

make image DEBUG=$DEBUG &&
docker tag runai/metrics-exporter:"$RELEASE_VER" "$IMAGE_FULL_NAME" &&
docker push "$IMAGE_FULL_NAME" &&
kubectl rollout restart deployment metrics-exporter -n runai
