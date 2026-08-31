#!/bin/sh

# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

set -e
cp -r /extensions/providers/. /providers/
cp -r /extensions/themes/. /themes/
cp -r /extensions/agent/. /agent/
