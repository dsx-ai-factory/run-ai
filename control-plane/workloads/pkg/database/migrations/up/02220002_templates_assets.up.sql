-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

CREATE TABLE IF NOT EXISTS crud_templates_assets (
                                       "id" uuid NOT NULL,
                                       "tenant_id" int4 NOT NULL,
                                       "template_id" uuid NOT NULL,
                                       "asset_id" uuid NOT NULL,
                                       "replica_type" text NOT NULL,  -- worker or master, worker for non-distributed
                                       PRIMARY KEY ("id")
);

-- asset_id first, because we will query for list of templates by asset_id
CREATE UNIQUE INDEX IF NOT EXISTS crud_templates_assets_unique ON crud_templates_assets
    USING btree (tenant_id, asset_id, template_id, replica_type);

-- for joining templates <-> templates_assets
CREATE INDEX IF NOT EXISTS crud_templates_assets_template_id ON crud_templates_assets
    USING btree (tenant_id, template_id);
