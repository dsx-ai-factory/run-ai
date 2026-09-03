/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { AssetIdAndKind } from "@/swagger-models/assets-service-client";

export interface INimModelStoreSection {
  selectedDataSources: Array<AssetIdAndKind>;
}
