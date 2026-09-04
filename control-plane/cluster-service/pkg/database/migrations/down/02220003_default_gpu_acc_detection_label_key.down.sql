-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

UPDATE nodepools
SET gpu_network_acceleration_label_key = ''
WHERE name = 'default'
  AND gpu_network_acceleration_label_key = 'nvidia.com/gpu.clique'
  AND deleted_at IS NULL;
