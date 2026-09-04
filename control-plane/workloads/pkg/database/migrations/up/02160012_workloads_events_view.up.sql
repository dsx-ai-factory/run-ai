-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

drop view if exists workloads_events_view;
create view workloads_events_view AS (
     SELECT
         wc.workload_id as workload_id,
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
         workloads_children as  wc
             RIGHT JOIN  events  e
                         ON  wc.child_id = e.involved_object_uid
         );