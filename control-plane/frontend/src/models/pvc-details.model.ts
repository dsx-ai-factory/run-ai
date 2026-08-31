/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { IDetailsPanelItem, TDetailsPanelItemGroup } from "@/models/details-panel.model";

export interface IPvcGeneralSectionData {
  groups: TDetailsPanelItemGroup[];
}

export interface IPvcRequestSectionData {
  items: IDetailsPanelItem[];
}

export interface IPvcStorageClassSectionData {
  items: IDetailsPanelItem[];
}
