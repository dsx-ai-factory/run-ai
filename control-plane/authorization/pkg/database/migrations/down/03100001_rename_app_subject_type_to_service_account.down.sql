-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

-- Intentionally a no-op. 'app' and 'service-account' rows are indistinguishable after the up migration, and previous
-- releases read, filter and delete both values as equivalent, so rolling back needs no data change. Converting every
-- service-account row back to 'app' would also rewrite rows that were created natively as service-account.
SELECT 1;
