-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE resources ALTER COLUMN gpu_over_quota_weight TYPE SMALLINT;
ALTER TABLE resources ALTER COLUMN cpu_deserved TYPE INT;
ALTER TABLE resources ALTER COLUMN cpu_limit TYPE INT;
ALTER TABLE resources ALTER COLUMN cpu_over_quota_weight TYPE SMALLINT;
ALTER TABLE resources ALTER COLUMN memory_deserved TYPE INT;
ALTER TABLE resources ALTER COLUMN memory_limit TYPE INT;
ALTER TABLE resources ALTER COLUMN memory_over_quota_weight TYPE SMALLINT;