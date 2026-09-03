-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

INSERT INTO workload_types (id, name, "group", source, created_by)
VALUES (gen_random_uuid(), 'NIMCache', 'apps.nvidia.com', 'external', 'Run.ai'),
       (gen_random_uuid(), 'NIMService', 'apps.nvidia.com', 'external', 'Run.ai'),
       (gen_random_uuid(), 'SPOTRequest', 'egx.nvidia.io', 'external', 'Run.ai')
ON CONFLICT (name, "group") DO NOTHING;

INSERT INTO external_workloads_type_config(id, scope_type, workload_type_id, category_id, priority_id, created_by, updated_by)
SELECT gen_random_uuid(),
       'system',
       workload_types.id,
       (SELECT id FROM workloads_categories WHERE name = 'Deploy' AND scope_type = 'system' LIMIT 1),
       (SELECT id FROM priorities WHERE name = 'very-high' AND scope_type = 'system' LIMIT 1),
       'Run.ai',
       'Run.ai'
FROM workload_types
WHERE workload_types.name IN ('NIMCache', 'NIMService', 'SPOTRequest')
AND NOT EXISTS (
    SELECT 1 FROM external_workloads_type_config 
    WHERE workload_type_id = workload_types.id
);

WITH mapping AS (
    SELECT name, "group", version
    FROM (
             VALUES
                 ('NIMCache', 'apps.nvidia.com', 'v1alpha1'),
                 ('NIMService', 'apps.nvidia.com', 'v1alpha1'),
                 ('SPOTRequest', 'egx.nvidia.io', 'v1')
         ) AS u(name, "group", version)
)
INSERT INTO workload_type_resource_interfaces (id, tenant_id, scope_type, workload_type_id, version, resource_interface, created_by)
SELECT gen_random_uuid(),
       NULL,
       'system',
       workload_types.id,
       m.version,
       jsonb_build_object(
           'spec', jsonb_build_object(
                   'structureDefinition', jsonb_build_object(
                           'rootComponent', jsonb_build_object(
                                   'kind', jsonb_build_object(
                                           'group', m."group",
                                           'kind', m.name,
                                           'version', m.version
                                           )
                                   )
                           )
                   )
           ),
       'Run.ai'
FROM mapping m, workload_types
WHERE workload_types.name = m.name
AND NOT EXISTS (
    SELECT 1 FROM workload_type_resource_interfaces 
    WHERE workload_type_id = workload_types.id 
    AND version = m.version
);