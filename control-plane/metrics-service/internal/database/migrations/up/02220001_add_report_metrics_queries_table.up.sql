-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

CREATE TABLE IF NOT EXISTS report_metrics_queries (
                                      report_id UUID,
                                      FOREIGN KEY (report_id) REFERENCES reports(id) ON DELETE CASCADE,
                                      display_name TEXT NOT NULL,
                                      query TEXT NOT NULL
);