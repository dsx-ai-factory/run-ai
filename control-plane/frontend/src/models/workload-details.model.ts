/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { PodNodeAffinity } from "@/swagger-models/workloads-client";
import type {
  IDetailsPanelItem,
  IDetailsPanelNetworkConnection,
  IDetailsPanelStorageVolume,
  IDetailsPanelTopology,
  TDetailsPanelItemGroup,
} from "@/models/details-panel.model";

export interface IGeneralSectionData {
  groups: TDetailsPanelItemGroup[];
}

export interface IEnvironmentSectionData {
  items: IDetailsPanelItem[];
  environmentVariables: string[];
  applicationParameters: IDetailsPanelItem[];
  labels: string[];
  annotations: string[];
}

export interface INetworkSectionData {
  connections: IDetailsPanelNetworkConnection[];
  topology?: IDetailsPanelTopology;
  topologyLevels: string[];
}

export interface IComputeSectionData {
  items: IDetailsPanelItem[];
  nodeAffinity?: PodNodeAffinity;
}

export interface IStorageSectionData {
  volumes: IDetailsPanelStorageVolume[];
}

export interface ICliCommandSectionData {
  command: string;
}
