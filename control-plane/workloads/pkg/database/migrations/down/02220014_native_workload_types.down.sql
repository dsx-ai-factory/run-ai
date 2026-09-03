-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

DROP INDEX IF EXISTS external_workload_type_config_unique_per_tenant;

DELETE
FROM workload_types
WHERE name IN (
               'TrainingWorkload',
               'DistributedWorkload',
               'InteractiveWorkload',
               'DistributedInferenceWorkload',
               'InferenceWorkload'
    );

ALTER TABLE workload_types
    DROP COLUMN IF EXISTS type;
