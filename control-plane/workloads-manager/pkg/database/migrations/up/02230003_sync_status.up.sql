-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0


CREATE TABLE IF NOT EXISTS sync_status (
    "workload_id" uuid NOT NULL,

    "tenant_id" int4 NOT NULL,
    "cluster_id" uuid NOT NULL,
    "department_id" text, -- needed only for mass deletion by department id
    "project_id" text, -- needed only for mass deletion by project id

    "deleting" boolean NOT NULL,

    "sent_to_agent" int8 NOT NULL, -- the recent update that was picked by the cluster
    "processed_by_agent" int8 NOT NULL, -- the recent update that was processed by the cluster
    "sync_message" text, -- optional message for the update that was recently processed by the cluster

    PRIMARY KEY ("workload_id")
);
