-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE node
    ADD COLUMN IF NOT EXISTS nv_link_clique_id TEXT;
ALTER TABLE node
    ADD COLUMN IF NOT EXISTS nv_link_domain_uid TEXT;