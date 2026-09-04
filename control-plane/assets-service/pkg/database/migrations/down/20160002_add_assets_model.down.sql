-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

-- drop table assets_model if exists
DROP TABLE IF EXISTS assets_model;


-- drop new field inference from assets_supported_workload_types
alter table assets_supported_workload_types
    drop column IF EXISTS inference;
