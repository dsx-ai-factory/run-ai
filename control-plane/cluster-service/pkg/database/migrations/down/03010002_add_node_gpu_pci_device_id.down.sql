-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE node DROP COLUMN IF EXISTS gpu_pci_device_id_observed_at,
    DROP COLUMN IF EXISTS gpu_pci_device_id;
