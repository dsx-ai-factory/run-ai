#!/bin/bash

# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0


RED='\033[0;31m'
NC='\033[0m' # No Color

servicePublicRouts=$(yq e '.paths | keys | .[]' $OPENAPI_TMP_OUT_PATH | sort)
allPublicRouts=$(yq e '.paths | keys | .[]' $OPENAPI_PUBLIC_SPEC_PATH | sort)

exportedRouts=$(comm -23 <(echo $servicePublicRouts | tr ' ' '\n' | sort) <(echo $allPublicRouts | tr ' ' '\n' | sort))

if [ -n "$exportedRouts" ]; then
    echo -e "${RED}routs missing from public spec:${NC}" $exportedRouts
    exit 0
else
    exit 0
fi