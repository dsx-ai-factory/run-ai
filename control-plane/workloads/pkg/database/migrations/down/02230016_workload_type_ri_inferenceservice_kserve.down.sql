-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

DELETE FROM workload_type_resource_interfaces
WHERE workload_type_id IN (
    SELECT id FROM workload_types
    WHERE (name, "group") IN (
            ('InferenceService', 'serving.kserve.io')
        )
);

DELETE FROM external_workloads_type_config
WHERE workload_type_id IN (
    SELECT id FROM workload_types
    WHERE (name, "group") IN (
            ('InferenceService', 'serving.kserve.io')
        )
);

DELETE FROM workload_types
WHERE (name, "group") IN (
        ('InferenceService', 'serving.kserve.io')
    );
