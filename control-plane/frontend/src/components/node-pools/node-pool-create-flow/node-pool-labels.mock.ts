/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Temporary suggestions until node-label discovery is available. These are not cluster inventory.
export const NODE_POOL_LABEL_SUGGESTIONS: Record<string, string[]> = {
  "runai.io/node-type": ["training", "inference"],
  "nvidia.com/gpu.product": ["T4", "T5", "H100"],
};
