-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

DROP TABLE IF EXISTS metric_stores;

DROP INDEX IF EXISTS idx_metric_stores_tenant_id;

DROP INDEX IF EXISTS uniq_metric_stores_tenant_active;