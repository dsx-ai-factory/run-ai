-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

CREATE TABLE IF NOT EXISTS mapping_locking (
    tenant_id   integer      NOT NULL PRIMARY KEY,
    locked_at   timestamptz  NOT NULL
);
