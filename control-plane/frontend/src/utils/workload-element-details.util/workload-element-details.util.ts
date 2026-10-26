/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { WorkloadElementResponse } from "@/swagger-models/workloads-client";
import type { IDetailsPanelItem, IDetailsPanelSectionDef, IDetailsPanelTopology } from "@/models/details-panel.model";
import type { IComputeSectionData, IGeneralSectionData, INetworkSectionData } from "@/models/workload-details.model";
import {
  matchAnyItem,
  filterDefined,
  topologyMatches,
  defineSection,
  buildSections,
  type SectionBuilder,
} from "@/utils/details-panel.utils";
import { DETAILS_PANEL_SECTIONS, workloadDetailsUtil } from "@/utils/workload-details.util";

export interface IBuildElementDetailsSectionsOptions {
  /** Full ordered list of topology levels (broadest → most specific) for the element's topology. */
  topologyLevels?: string[];
  /** Number of pods backing this element, sourced separately since it's not on the element itself. */
  podsCount?: number;
}

interface IBuildElementSectionsInput {
  element: WorkloadElementResponse;
  topologyLevels: string[];
  podsCount: number;
}

function buildElementGeneralSection(input: IBuildElementSectionsInput): IDetailsPanelSectionDef | undefined {
  const { element } = input;
  const group = filterDefined([
    workloadDetailsUtil.pair("Type", element.kubernetesRef?.gvk?.kind),
    workloadDetailsUtil.pair(
      "Creation time",
      element.createdAt ? workloadDetailsUtil.formatDate(element.createdAt) : undefined,
    ),
  ]);
  const data: IGeneralSectionData = { groups: [group] };
  return defineSection(DETAILS_PANEL_SECTIONS.GENERAL, {
    isEmpty: !group.length,
    data,
    customMatch: (term: string) => matchAnyItem(group, term),
  });
}

function buildElementNetworkSection(input: IBuildElementSectionsInput): IDetailsPanelSectionDef | undefined {
  const { element, topologyLevels } = input;
  const topology: IDetailsPanelTopology | undefined = element.topology?.name ? element.topology : undefined;
  const data: INetworkSectionData = { connections: [], topology, topologyLevels };
  return defineSection(DETAILS_PANEL_SECTIONS.NETWORK, {
    isEmpty: !topology,
    data,
    customMatch: (term: string) => (topology ? topologyMatches(topology, topologyLevels, term) : false),
  });
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

  items.push(...workloadDetailsUtil.buildAllocatedResourceItems(allocated));
  items.push({ label: "Pods", value: String(podsCount) });

  return items;
}

function buildElementComputeSection(input: IBuildElementSectionsInput): IDetailsPanelSectionDef | undefined {
  const { element, podsCount } = input;
  const items = buildElementComputeItems(element, podsCount);
  const data: IComputeSectionData = { items };
  // Never empty — "Pods" is always included.
  return defineSection(DETAILS_PANEL_SECTIONS.COMPUTE, {
    isEmpty: false,
    data,
    customMatch: (term: string) => matchAnyItem(items, term),
  });
}

// Order here determines the on-screen order of the accordion cards.
const ELEMENT_SECTION_BUILDERS: SectionBuilder<IBuildElementSectionsInput>[] = [
  buildElementGeneralSection,
  buildElementNetworkSection,
  buildElementComputeSection,
];

/**
 * Builds the accordion sections for the workload-element Details tab. Unlike the workload-level
 * builder, this always has exactly one data source (the element response), so Environment /
 * Storage / CLI-command builders don't apply per the design — but it drives off the same
 * registry + shared `buildSections` orchestration as the workload-level tab.
 */
function buildElementDetailsSections(
  element: WorkloadElementResponse,
  options: IBuildElementDetailsSectionsOptions = {},
): IDetailsPanelSectionDef[] {
  const { topologyLevels = [], podsCount = 0 } = options;
  return buildSections(ELEMENT_SECTION_BUILDERS, { element, topologyLevels, podsCount });
}

export const workloadElementDetailsUtil = {
  buildElementDetailsSections,
};
