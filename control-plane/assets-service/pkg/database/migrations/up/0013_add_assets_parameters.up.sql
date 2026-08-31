-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE assets_environment ADD COLUMN if not exists run_as_non_root boolean;
ALTER TABLE assets_environment ADD COLUMN if not exists seccomp_profile_type text;