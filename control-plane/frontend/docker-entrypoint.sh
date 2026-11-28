#!/bin/sh
# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

set -e

if [ -e /proc/net/if_inet6 ]; then
    echo "listen [::]:8080;" > /nginx-run/pid/listen-ipv6.conf
else
    : > /nginx-run/pid/listen-ipv6.conf
fi

exec nginx -g "daemon off;"
