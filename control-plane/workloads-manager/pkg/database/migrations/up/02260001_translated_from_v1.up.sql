-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0


ALTER TABLE workloads_queue
    ADD COLUMN IF NOT EXISTS translated_from_v1 BOOLEAN NOT NULL DEFAULT FALSE;


