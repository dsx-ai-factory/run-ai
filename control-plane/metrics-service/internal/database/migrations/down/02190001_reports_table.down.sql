-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

DROP TRIGGER IF EXISTS set_report_status_update_time on reports;
DROP FUNCTION IF EXISTS update_report_status_update_time();
DROP TABLE IF EXISTS report_metrics;
DROP INDEX IF EXISTS idx_reports_tenant_name;
DROP TABLE IF EXISTS reports;
