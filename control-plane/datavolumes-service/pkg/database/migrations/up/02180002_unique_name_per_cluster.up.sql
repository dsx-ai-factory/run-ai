-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE datavolumes DROP constraint if exists datavolume_cluster_unique_constraint, ADD constraint datavolume_cluster_unique_constraint UNIQUE (cluster_id, name);
ALTER TABLE datavolumes DROP constraint if exists datavolume_tenant_unique_constraint;

drop view if exists datavolumes_view;
create view datavolumes_view AS (SELECT DISTINCT ON (d.id, s.scope_id, s.scope_type)
    d.id,
    d.created_at,
    d.updated_at,
    d.created_by,
    d.updated_by,
    d.name,
    d.description,
    d.project_id,
    d.origin_pvc_name,
    d.project_name,
    d.department_id,
    d.tenant_id,
    d.cluster_id,
    d.phase,
    d.phase_message,
    d.conditions,
    d.datavolume_pvc_name,
    d.datavolume_pv_name,
    d.should_delete_original_volume,
    s.scope_id,
    s.scope_type,
    s.scope_name,
    s.department_id as shared_scope_department_id,
    s.project_id    as shared_scope_project_id,
    s.created_at    as shared_scope_created_at,
    s.created_by    as shared_scope_created_by

    FROM datavolumes as d
        LEFT JOIN shared_scopes as s ON d.id = s.datavolume_id
);

ALTER TABLE shared_scopes DROP COLUMN if exists project_name;