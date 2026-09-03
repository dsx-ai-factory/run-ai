-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE assets_environment ADD COLUMN if not exists stdin boolean;
ALTER TABLE assets_environment ADD COLUMN if not exists tty boolean;