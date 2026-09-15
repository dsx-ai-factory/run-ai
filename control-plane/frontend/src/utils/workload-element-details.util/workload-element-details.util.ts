/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { WorkloadElementResponse } from "@/swagger-models/workloads-client";
import type { IDetailsPanelItem, IDetailsPanelSectionDef, IDetailsPanelTopology } from "@/models/details-panel.model";
import type { IComputeSectionData, IGeneralSectionData, INetworkSectionData } from "@/models/workload-details.model";
import { matchesItem, objectMatchesTerm } from "@/utils/details-panel.utils";
import { getCpuCoresForUI, getMemoryUnitsForUI } from "@/utils/table-format.util";
import { DETAILS_PANEL_SECTIONS, workloadDetailsUtil } from "@/utils/workload-details.util";

export interface IBuildElementDetailsSectionsOptions {
  /** Full ordered list of topology levels (broadest → most specific) for the element's topology. */
  topologyLevels?: string[];
  /** Number of pods backing this element, sourced separately since it's not on the element itself. */
  podsCount?: number;
}

function buildElementGeneralSection(element: WorkloadElementResponse): IDetailsPanelSectionDef | undefined {
  const group: IDetailsPanelItem[] = [];
  const type = element.kubernetesRef?.gvk?.kind;
  if (type) group.push({ label: "Type", value: type });
  if (element.createdAt) {
    group.push({ label: "Creation time", value: workloadDetailsUtil.formatDate(element.createdAt) });
  }
  if (!group.length) return undefined;

  const data: IGeneralSectionData = { groups: [group] };
  return {
    ...DETAILS_PANEL_SECTIONS.GENERAL,
    customMatch: (term: string) => group.some((item) => matchesItem(item, term)),
    data,
  };
}

function buildElementNetworkSection(
  element: WorkloadElementResponse,
  topologyLevels: string[],
): IDetailsPanelSectionDef | undefined {
  const topology: IDetailsPanelTopology | undefined = element.topology?.name ? element.topology : undefined;
  if (!topology) return undefined;

  const data: INetworkSectionData = { connections: [], topology, topologyLevels };
  return {
    ...DETAILS_PANEL_SECTIONS.NETWORK,
    customMatch: (term: string) => objectMatchesTerm({ ...topology, levels: topologyLevels }, term),
    data,
  };
}

function buildElementComputeItems(element: WorkloadElementResponse, podsCount: number): IDetailsPanelItem[] {
  const items: IDetailsPanelItem[] = [];
  const requested = element.podsRequestedResources;
  const allocated = element.allocatedResources;

  const requestedConfigs = [
    { label: "GPU compute", value: requested?.gpu, suffix: "G" },
    { label: "GPU memory", value: requested?.gpuMemory, suffix: undefined },
    { label: "CPU compute", value: requested?.cpu, suffix: "cores" },
    { label: "CPU memory", value: requested?.cpuMemory, suffix: undefined },
  ];
  for (const { label, value, suffix } of requestedConfigs) {
    const formatted = workloadDetailsUtil.formatResourceRequestLimit(value, suffix);
    if (formatted) items.push({ label, value: formatted });
  }

  if (allocated?.gpu !== null && allocated?.gpu !== undefined) {
    items.push({ label: "GPU compute allocation", value: String(allocated.gpu) });
  }
  if (allocated?.gpuMemory) {
    items.push({ label: "GPU memory allocation", value: getMemoryUnitsForUI(allocated.gpuMemory) });
  }
  if (allocated?.cpu !== null && allocated?.cpu !== undefined) {
    items.push({ label: "CPU compute allocation", value: getCpuCoresForUI(allocated.cpu) });
  }
  if (allocated?.cpuMemory) {
    items.push({ label: "CPU memory allocation", value: getMemoryUnitsForUI(allocated.cpuMemory) });
  }

  items.push({ label: "Pods", value: String(podsCount) });

  return items;
}

function buildElementComputeSection(element: WorkloadElementResponse, podsCount: number): IDetailsPanelSectionDef {
  const items = buildElementComputeItems(element, podsCount);
  const data: IComputeSectionData = { items };
  return {
    ...DETAILS_PANEL_SECTIONS.COMPUTE,
    customMatch: (term: string) => items.some((item) => matchesItem(item, term)),
    data,
  };
}

/**
 * Builds the accordion sections for the workload-element Details tab. Unlike the workload-level
 * builder, this always has exactly one data source (the element response) so there's no need for
 * the workload-level orchestration machinery — Environment/Storage/CLI-command don't apply to
 * elements per the design.
 */
export function buildElementDetailsSections(
  element: WorkloadElementResponse,
  options: IBuildElementDetailsSectionsOptions = {},
): IDetailsPanelSectionDef[] {
  const { topologyLevels = [], podsCount = 0 } = options;
  const sections = [
    buildElementGeneralSection(element),
    buildElementNetworkSection(element, topologyLevels),
    buildElementComputeSection(element, podsCount),
  ];
  return sections.filter((s): s is IDetailsPanelSectionDef => s !== undefined);
}
