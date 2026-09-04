/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

export interface IHybridDataSourceSectionDisableFields {
  pvc?: string;
  git?: string;
  hostPath?: string;
  nfs?: string;
  volume?: string;
  configMap?: string;
  emptyDir?: string;
}
