-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

CREATE TABLE IF NOT EXISTS assets_sync_status(
                                          id UUID NOT NULL PRIMARY KEY,

                                          tenant_id INT NOT NULL,
                                          cluster_id UUID NOT NULL,
                                          scope_type TEXT NOT NULL,
                                          scope_id TEXT NOT NULL,
                                          asset_id UUID NOT NULL,
                                          resource_name TEXT NOT NULL,

                                          sent_to_sync_add timestamptz NOT NULL,
                                          sent_to_sync_update timestamptz,
                                          sent_to_sync_delete timestamptz,
                                          processed_by_cluster timestamptz,

                                          phase TEXT,
                                          message TEXT
);

CREATE UNIQUE INDEX IF NOT EXISTS sync_status_by_asset_cluster ON assets_sync_status(tenant_id, cluster_id, scope_type, scope_id, resource_name);

CREATE INDEX IF NOT EXISTS sync_status_by_asset_id ON assets_sync_status(tenant_id, cluster_id, asset_id)