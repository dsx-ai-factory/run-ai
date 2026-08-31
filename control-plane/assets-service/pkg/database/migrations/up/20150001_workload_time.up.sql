-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE workloads ALTER COLUMN created_at SET DEFAULT NOW();
ALTER TABLE workloads ALTER COLUMN updated_at SET DEFAULT NOW();