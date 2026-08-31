-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE workloads_queue ADD COLUMN IF NOT EXISTS input_request_path TEXT;
ALTER TABLE workloads_queue ADD COLUMN IF NOT EXISTS input_request_body JSONB;
