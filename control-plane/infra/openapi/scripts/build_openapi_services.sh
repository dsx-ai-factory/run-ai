#!/bin/bash

# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

echo "Building OpenAPI services after updating the OpenAPI template"
find . -name swagger | grep -v -E "./infra/|./redoc|./frontend|./backend/js/" | xargs dirname | xargs -I {} bash -c 'cd {} && make openapi-build'

echo
echo "OpenAPI services built successfully"
