# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

ARG POSTGRES_VERSION=18
FROM postgres:${POSTGRES_VERSION}
ARG POSTGRES_VERSION_MAJOR=16
RUN apt-get update \
    && apt-get install -y --no-install-recommends postgresql-${POSTGRES_VERSION_MAJOR}-cron \
    && rm -rf /var/lib/apt/lists/*
