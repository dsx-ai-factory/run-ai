-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE workloads DROP COLUMN IF EXISTS dist_specific_env;
ALTER TABLE workloads_assets DROP COLUMN IF EXISTS dist_replica_type;