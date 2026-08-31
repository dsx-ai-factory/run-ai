-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE workloads_queue DROP COLUMN IF EXISTS input_request_path;
ALTER TABLE workloads_queue DROP COLUMN IF EXISTS input_request_body;
