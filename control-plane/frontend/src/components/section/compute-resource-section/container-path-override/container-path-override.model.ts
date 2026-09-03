/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { type AssetIdAndKind } from "@/swagger-models/assets-service-client";

export interface ISelectedDatasourcesContainerPathData extends AssetIdAndKind {
  name: string;
  origin: string;
}

export interface IContainerPathOverride {
  dataSourceId: string;
  containerPathOverride: string;
}

export interface ITabProps {
  key: string;
  id: string;
  label: string;
  isInvalid?: boolean;
  icon?: string;
  removeable?: boolean;
}
