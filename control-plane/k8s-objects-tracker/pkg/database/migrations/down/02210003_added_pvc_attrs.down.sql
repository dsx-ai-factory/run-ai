-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0


ALTER TABLE pvcs DROP COLUMN IF EXISTS storage_class_name;

ALTER TABLE pvcs DROP COLUMN IF EXISTS phase;

ALTER TABLE pvcs DROP COLUMN IF EXISTS requested_capacity;
ALTER TABLE pvcs DROP COLUMN IF EXISTS actual_capacity;
