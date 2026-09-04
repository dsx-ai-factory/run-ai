#!/usr/bin/env bash

# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0


set -x


TAG=${TAG:-"your_name_here"} # You can also put it in the git ignored .env
DOCKERFILE=${DOCKERFILE:-"Dockerfile"}
BUILD_TARGET=build

while getopts d flag
do
    case "${flag}" in
        d) DOCKERFILE=`dirname "$0"`/../debug/Dockerfile.debug
           BUILD_TARGET=build-debug
            ;;
    esac
done

make $BUILD_TARGET

# export TAG if exists in .env
if [ -f .env ]; then
  export $(cat .env | xargs)
fi

docker build --platform linux/amd64 -t gcr.io/run-ai-lab/project-controller:$TAG `dirname "$0"`/.. -f $DOCKERFILE
docker push gcr.io/run-ai-lab/project-controller:$TAG
kubectl rollout restart deployment/runai-project-controller -n runai
