-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

-- Unlike other spec tables, this one stores the spec as JSON keyed by asset_id,
-- so new asset kinds can be added without a specific per-asset table.
-- Also carries tenant_id, which the per-asset spec tables historically omitted.
CREATE TABLE IF NOT EXISTS assets_spec (
    asset_id  UUID  NOT NULL PRIMARY KEY,
    tenant_id INT   NOT NULL,
    spec      jsonb NOT NULL
);

CREATE INDEX IF NOT EXISTS assets_spec_by_tenant ON assets_spec USING btree (tenant_id);

