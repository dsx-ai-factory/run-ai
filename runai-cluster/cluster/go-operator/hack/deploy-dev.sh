#!/bin/bash

# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

# need to docker login first

dockerTagName=${IMAGE_NAME:-"runai-developer"} # You can also put it in the git ignored .env
dockerRepo="gcr.io/run-ai-lab/go-operator"
DEBUG=

if [[ $* == *-d* ]]; then
  echo "Building debug image"
  DEBUG=1
fi

if [ -f .env ]; then
  export $(cat .env | xargs)
fi

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )" &&
make build
IMAGE_TAG_BASE=$dockerRepo ENVIRONMENT=lab VERSION=$dockerTagName make -C $DIR/.. DEBUG=$DEBUG docker-build docker-push &&
kubectl rollout restart deployment/runai-operator -n runai
