#!/usr/bin/env bash

# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

set -x

if [ -f .env ]; then
  export $(cat .env | xargs)
fi

DOCKER_REPO="gcr.io/run-ai-lab/"
IMAGE_NAME="cluster-sync"
TAG=${TAG:-"latest"}
IMAGE_FULL_NAME="${DOCKER_REPO}${IMAGE_NAME}:${TAG}"

make build-only
docker build --platform linux/amd64 -t $IMAGE_FULL_NAME `dirname "$0"`/..
docker push $IMAGE_FULL_NAME &&
kubectl rollout restart deployment cluster-sync -n runai
