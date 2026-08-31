-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE reconciler_versions
RENAME COLUMN access_keys TO user_applications;

ALTER TABLE reconciler_versions
RENAME COLUMN service_accounts TO applications;

ALTER TABLE reconciler_versions
RENAME COLUMN internal_service_accounts TO internal_applications;

