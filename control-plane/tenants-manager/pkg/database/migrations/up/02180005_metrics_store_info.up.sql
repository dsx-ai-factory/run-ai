-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE metric_stores ADD COLUMN IF NOT EXISTS "info" jsonb;
UPDATE metric_stores SET info  = '{"type":"grafanaLab", "grafanalab_identifier": null}'::jsonb;
with tenant_name as (
    select id, name from tenants
    )
UPDATE metric_stores SET info =  jsonb_set(info, '{grafanalab_identifier}', to_jsonb(tenant_name.name))
    FROM tenant_name
WHERE metric_stores.tenant_id = tenant_name.id;