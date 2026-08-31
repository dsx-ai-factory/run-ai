-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE IF EXISTS pods ADD COLUMN IF NOT EXISTS search_vector tsvector GENERATED ALWAYS AS
    (to_tsvector(
            'english',
            COALESCE(name::text, '') || ' ' ||
            COALESCE(phase::text, '') || ' ' ||
            COALESCE(node_name::text, ''))) STORED;
CREATE INDEX IF NOT EXISTS pods_search_vector_idx ON pods using gin(search_vector);