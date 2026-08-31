-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE resources
    ALTER COLUMN gpu_deserved DROP NOT NULL,
    ALTER COLUMN gpu_deserved DROP DEFAULT;
