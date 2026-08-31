-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE nodepools ADD COLUMN IF NOT EXISTS gpu_network_acceleration_detection TEXT DEFAULT 'Auto';
ALTER TABLE nodepools ADD COLUMN IF NOT EXISTS gpu_network_acceleration_label_key TEXT;
