-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

CREATE TABLE IF NOT EXISTS sync_status (
    "id" uuid NOT NULL,

    "tenant_id" int4 NOT NULL,
    "cluster_id" uuid NOT NULL,

    "asset_id" uuid NOT NULL,
    "prev_name" text,
    "spec" jsonb NOT NULL,

    "sent_to_agent_apply" int8,
    "sent_to_agent_delete" int8,
    "processed_by_agent" int8,

    "message" text,

    PRIMARY KEY ("id")
);

CREATE UNIQUE INDEX IF NOT EXISTS sync_status_by_asset   ON sync_status USING btree (tenant_id, asset_id, cluster_id);
CREATE        INDEX IF NOT EXISTS sync_status_by_apply   ON sync_status USING btree (cluster_id, sent_to_agent_apply);
CREATE        INDEX IF NOT EXISTS sync_status_by_delete  ON sync_status USING btree (cluster_id, sent_to_agent_delete);
