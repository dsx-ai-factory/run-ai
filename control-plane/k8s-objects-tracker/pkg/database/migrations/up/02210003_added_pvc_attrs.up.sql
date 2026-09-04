-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0


ALTER TABLE pvcs ADD COLUMN IF NOT EXISTS storage_class_name text;

ALTER TABLE pvcs ADD COLUMN IF NOT EXISTS phase text;

ALTER TABLE pvcs ADD COLUMN IF NOT EXISTS requested_capacity text;
ALTER TABLE pvcs ADD COLUMN IF NOT EXISTS actual_capacity text;
