-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE node ADD COLUMN IF NOT EXISTS gpu_pci_device_id TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS gpu_pci_device_id_observed_at TIMESTAMPTZ;
