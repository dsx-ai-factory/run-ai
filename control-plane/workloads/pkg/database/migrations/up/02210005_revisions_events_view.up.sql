-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

DROP VIEW IF EXISTS revisions_events_view;

CREATE VIEW revisions_events_view AS (
     SELECT
         rc.revision_id as revision_id,
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
         revisions_children AS rc
             RIGHT JOIN events e
                 ON rc.child_id = e.involved_object_uid
);