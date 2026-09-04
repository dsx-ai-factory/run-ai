-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE resources
    ALTER COLUMN gpu_deserved SET NOT NULL,
    ALTER COLUMN gpu_deserved SET DEFAULT 0;
