-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

-- Intentionally a no-op. Rows renamed from 'apps' cannot be told apart from native 'service-account' rows, and the
-- 'apps' rows that were dropped were duplicates of an existing 'service-account' row.
SELECT 1;
