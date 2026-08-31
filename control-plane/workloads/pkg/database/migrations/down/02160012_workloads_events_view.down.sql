-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

drop view if exists workloads_events_view;
create view workloads_events_view AS (
                                     SELECT DISTINCT ON (e.id)
                                         w.id as workload_id,
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
                                         INNER JOIN  workloads_children  wc
                                     ON      wc.workload_id = w.id
                                         INNER JOIN  events  e
                                         ON      w.tenant_id = e.tenant_id AND w.cluster_id = e.cluster_id AND (wc.child_id = e.involved_object_uid OR w.id = e.involved_object_uid)
                                         );