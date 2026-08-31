-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

-- The V2 asset-sync API has been removed. Its bookkeeping table is obsolete;
-- the current agent sync flow uses the separate sync_status table.
DROP TABLE IF EXISTS assets_sync_status;
