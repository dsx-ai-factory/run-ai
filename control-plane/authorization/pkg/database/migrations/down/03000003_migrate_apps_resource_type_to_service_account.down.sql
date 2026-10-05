-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

-- Intentionally a no-op. Rows renamed from 'apps' cannot be told apart from native 'service-account' rows, and the
-- 'apps' rows that were dropped were exact duplicates seeded next to 'service-account'. A previous release
-- re-creates the static role/permission-set rows on startup and never required 'apps' rows to exist.
SELECT 1;
