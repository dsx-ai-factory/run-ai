-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

UPDATE permission_sets
SET name = 'Integration metrics read'
WHERE name = 'NetQ metrics read';

UPDATE roles
SET name = 'Integration metrics collector'
WHERE name = 'Metrics scraper';

UPDATE role_permissions
SET resource_type = 'workload-integration-metrics'
WHERE resource_type = 'scrapeable-metrics';
