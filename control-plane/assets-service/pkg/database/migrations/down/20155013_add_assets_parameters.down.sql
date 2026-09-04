-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

---- Redo of 0013_add_assets_parameters.down.sql in idempotent in case it didn't run
ALTER TABLE assets_environment DROP COLUMN if exists run_as_non_root;
ALTER TABLE assets_environment DROP COLUMN if exists seccomp_profile_type;