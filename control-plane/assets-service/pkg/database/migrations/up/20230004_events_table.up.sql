-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

CREATE TABLE IF NOT EXISTS events (
    "created_at" timestamptz,
    "id" uuid NOT NULL,
    "asset_id" uuid NOT NULL,
    "tenant_id" int8 NOT NULL,
    "cluster_id" uuid NOT NULL,
    "scope_type" text NOT NULL, 
    "scope_id" text NOT NULL,
    "asset_kind" text NOT NULL,
    "object_name" text,
    "object_namespace" text,
    "object_kind" text,
    "object_uid" text,
    "type" text,
    "phase" text,
    "message" text,
    "reason" text,
    "source" text,
    PRIMARY KEY ("id")
);

CREATE INDEX IF NOT EXISTS events_by_asset ON events(object_kind, asset_id, created_at DESC);

CREATE TABLE IF NOT EXISTS assets_resources_ids (
    "tenant_id" int8 NOT NULL,
    "asset_id" uuid NOT NULL,
    "resource_id" text NOT NULL,
    "last_synced_at" timestamptz NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS assets_resources_ids_unique ON assets_resources_ids(tenant_id, asset_id, resource_id);