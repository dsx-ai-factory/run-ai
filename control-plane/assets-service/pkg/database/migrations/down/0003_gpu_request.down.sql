-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0


ALTER TABLE assets_compute DROP COLUMN if exists gpu_request_type;
ALTER TABLE assets_compute DROP COLUMN if exists gpu_portion;
ALTER TABLE assets_compute DROP COLUMN if exists gpu_memory;
ALTER TABLE assets_compute DROP COLUMN if exists mig_profile;
