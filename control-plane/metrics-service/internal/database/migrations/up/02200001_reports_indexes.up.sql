-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

CREATE INDEX IF NOT EXISTS idx_reports_created_by ON reports (created_by);
CREATE INDEX IF NOT EXISTS idx_reports_created_at ON reports (created_at);
CREATE INDEX IF NOT EXISTS idx_reports_status ON reports (status);

ALTER TABLE IF EXISTS reports ADD COLUMN IF NOT EXISTS search_vector tsvector GENERATED ALWAYS AS
    (to_tsvector(
            'english',
            COALESCE(name::text, '') || ' ' ||
            COALESCE(created_by::text, '') || ' ' ||
            COALESCE(status::text, ''))) STORED;
CREATE INDEX IF NOT EXISTS idx_reports_search_vector ON reports using gin(search_vector);
