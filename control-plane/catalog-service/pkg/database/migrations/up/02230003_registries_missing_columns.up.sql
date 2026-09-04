-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE registries ADD COLUMN if NOT EXISTS description text;
ALTER TABLE registries ADD COLUMN if NOT EXISTS version text;