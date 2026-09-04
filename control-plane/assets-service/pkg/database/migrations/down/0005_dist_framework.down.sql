-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

UPDATE assets_pvc SET ephemeral = false WHERE ephemeral IS NULL;
ALTER TABLE assets_pvc ALTER COLUMN ephemeral SET NOT NULL;