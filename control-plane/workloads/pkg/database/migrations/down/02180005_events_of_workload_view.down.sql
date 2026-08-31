-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

drop view if exists timeout_events_of_workload_view;
drop view if exists events_of_workload_view;
create view events_of_workload_view AS (
     SELECT
         w.id as workload_id,
         w.project_name,
         w.type as workload_type,
         w.submitted_by,
         e.id as event_id,
         e.created_at,
         e.cluster_id,
         e.tenant_id,
         e.type,
         e.message,
         e.reason,
         e.source,
         e.involved_object_uid,
         e.involved_object_kind,
         e.involved_object_name,
         e.involved_object_namespace
     FROM
         workloads as w
             RIGHT JOIN  events  e
                ON  w.id = e.involved_object_uid
         );