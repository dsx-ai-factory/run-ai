-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE policies DROP COLUMN IF EXISTS validation_issues;
ALTER TABLE policies DROP COLUMN IF EXISTS gvk_workload;
