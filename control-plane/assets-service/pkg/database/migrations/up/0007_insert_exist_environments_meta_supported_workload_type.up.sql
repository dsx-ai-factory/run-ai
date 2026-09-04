-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

INSERT INTO assets_supported_workload_types ("asset_id", "workspace", "training", "distributed")
SELECT
    am. "id" AS "asset_id",
    TRUE AS "workspace",
    TRUE AS "training",
    FALSE AS "distributed"
FROM
    assets_metadata AS am
WHERE
        am. "kind" = 'environment'
ON CONFLICT ("asset_id") DO NOTHING;