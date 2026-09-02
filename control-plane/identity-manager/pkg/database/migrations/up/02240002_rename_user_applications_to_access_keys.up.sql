-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE reconciler_versions
RENAME COLUMN user_applications TO access_keys;

ALTER TABLE reconciler_versions
RENAME COLUMN applications TO service_accounts;

ALTER TABLE reconciler_versions
RENAME COLUMN internal_applications TO internal_service_accounts;

