-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

DROP VIEW IF EXISTS external_workloads_type_config_view;
CREATE VIEW external_workloads_type_config_view
    AS (
    SELECT
        t.*,
        c.name as category_name
    FROM
        external_workloads_type_config as t
    JOIN workloads_categories as c ON c.id = t.category_id
);