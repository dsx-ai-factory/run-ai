-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE assets_environment DROP COLUMN if exists run_as_non_root;
ALTER TABLE assets_environment DROP COLUMN if exists seccomp_profile_type;