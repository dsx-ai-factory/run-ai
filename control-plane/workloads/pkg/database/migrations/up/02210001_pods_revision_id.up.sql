-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE IF EXISTS pods ADD COLUMN IF NOT EXISTS revision_id uuid;
CREATE INDEX IF NOT EXISTS pods_revision_idx ON pods (revision_id) where revision_id is not null;