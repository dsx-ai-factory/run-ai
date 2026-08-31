#!/usr/bin/env bash

# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

set -x
TAG=${TAG:-"yourname"}

# export TAG if exists in .env
if [ -f .env ]; then
  export $(cat .env | xargs)
fi

go get -v -t ./...
make deploy TAG=$TAG

make clean
