/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect } from "vitest";
import type { WorkloadElementResponse } from "@/swagger-models/workloads-client";
import { buildElementDetailsSections } from "./workload-element-details.util";
import type { IGeneralSectionData, INetworkSectionData, IComputeSectionData } from "@/models/workload-details.model";

function buildElement(overrides: Partial<WorkloadElementResponse> = {}): WorkloadElementResponse {
  return {
    id: "el-1",
    name: "prefill-worker",
    hierarchyLevel: 1,
    createdAt: "2026-01-13T10:32:00Z",
    kubernetesRef: { name: "prefill-worker", gvk: { group: "apps", version: "v1", kind: "Deployment" } },
    podsRequestedResources: null,
    allocatedResources: null,
    ...overrides,
  };
}

describe("buildElementDetailsSections", () => {
  it("builds the general section from kubernetesRef kind and createdAt", () => {
    const sections = buildElementDetailsSections(buildElement());
    const general = sections.find((s) => s.id === "general");
    expect(general).toBeDefined();
    const data = general?.data as IGeneralSectionData;
    expect(data.groups[0]).toEqual(
      expect.arrayContaining([
        { label: "Type", value: "Deployment" },
        expect.objectContaining({ label: "Creation time" }),
      ]),
    );
  });

  it("omits the general section when there is no type or creation time", () => {
    const sections = buildElementDetailsSections(buildElement({ kubernetesRef: null, createdAt: undefined }));
    expect(sections.find((s) => s.id === "general")).toBeUndefined();
  });

  it("omits the network section when the element has no topology", () => {
    const sections = buildElementDetailsSections(buildElement({ topology: null }));
    expect(sections.find((s) => s.id === "network")).toBeUndefined();
  });

  it("builds the network section with the element's topology and given levels", () => {
    const sections = buildElementDetailsSections(
      buildElement({
        topology: {
          name: "my-gb200-topology",
          requiredPlacement: "network.topology.nvidia.com/region",
          actualPlacement: { level: "network.topology.nvidia.com/zone", value: "us-central1" },
        },
      }),
      { topologyLevels: ["network.topology.nvidia.com/region", "network.topology.nvidia.com/zone"] },
    );
    const network = sections.find((s) => s.id === "network");
    expect(network).toBeDefined();
    const data = network?.data as INetworkSectionData;
    expect(data.topology?.name).toBe("my-gb200-topology");
    expect(data.topologyLevels).toEqual(["network.topology.nvidia.com/region", "network.topology.nvidia.com/zone"]);
    expect(data.connections).toEqual([]);
  });

  it("builds compute items from requested + allocated resources and the given pod count", () => {
    const sections = buildElementDetailsSections(
      buildElement({
        podsRequestedResources: {
          gpu: { request: 4, limit: 4 },
          gpuMemory: { request: "0.5G", limit: "0.5G" },
          cpu: { request: 1, limit: 1 },
          cpuMemory: { request: "100M", limit: "200M" },
        },
        allocatedResources: { gpu: 4, gpuMemory: "0.5G", cpu: 1, cpuMemory: "100M" },
      }),
      { podsCount: 4 },
    );
    const compute = sections.find((s) => s.id === "compute");
    expect(compute).toBeDefined();
    const items = (compute?.data as IComputeSectionData).items;
    const labels = items.map((i) => i.label);
    expect(labels).toEqual([
      "GPU compute",
      "GPU memory",
      "CPU compute",
      "CPU memory",
      "GPU compute allocation",
      "GPU memory allocation",
      "CPU compute allocation",
      "CPU memory allocation",
      "Pods",
    ]);
    expect(items.find((i) => i.label === "GPU compute allocation")?.value).toBe("4");
    expect(items.find((i) => i.label === "Pods")?.value).toBe("4");
  });

  it("always includes the compute section with at least the pods count", () => {
    const sections = buildElementDetailsSections(buildElement(), { podsCount: 0 });
    const compute = sections.find((s) => s.id === "compute");
    expect(compute).toBeDefined();
    expect((compute?.data as IComputeSectionData).items).toEqual([{ label: "Pods", value: "0" }]);
  });
});
