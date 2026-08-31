/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect, vi, beforeEach } from "vitest";
import { shallowMount } from "@vue/test-utils";
import { createTestingPinia } from "@pinia/testing";
import { createRouter, createMemoryHistory } from "vue-router";
import WorkloadYamlNew from "./workload-yaml-new.vue";
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
import type { Project } from "@/swagger-models/org-unit-service-client";
import type { NodepoolMinimal } from "@/swagger-models/cluster-service-client";

const router = createRouter({
  history: createMemoryHistory(),
  routes: [{ path: "/", component: { template: "<div />" } }],
});

vi.mock("@/services/control-plane/cluster.service/cluster.service", () => ({
  clusterService: {
    getNodepoolsMinimal: vi.fn(),
  },
}));

vi.mock("@/services/control-plane/org-unit.service/org-unit.service", () => ({
  orgUnitService: {
    getProject: vi.fn(),
    getProjects: vi.fn().mockResolvedValue([]),
  },
}));

vi.mock("@/services/authorization/permission.service/permission.service", () => ({
  permissionService: {
    getPermittedScopesForAction: vi.fn().mockResolvedValue({}),
  },
}));

vi.mock("@/services/cluster/workload.service/workload.service", () => ({
  workloadService: {
    canSubmitWorkload: vi.fn().mockReturnValue(null),
  },
}));

vi.mock("@/utils/filter.util/filter.util", () => ({
  filterUtil: {
    getEqualsFilterString: vi.fn().mockReturnValue("clusterId==test-cluster"),
  },
}));

const makeProject = (nodePoolIds: Array<string | undefined> = []): Project =>
  ({
    resources: nodePoolIds.map((id) => ({ nodePool: id ? { id } : undefined })),
    effective: { defaultNodePools: [] },
  }) as unknown as Project;

const makeNodepool = (id: string, detected: boolean): NodepoolMinimal =>
  ({ id, gpuNetworkAccelerationDetected: detected }) as NodepoolMinimal;

const mountComponent = () => {
  return shallowMount(WorkloadYamlNew, {
    global: {
      plugins: [router, createTestingPinia({ createSpy: vi.fn })],
      stubs: {
        WorkloadFormWrapper: { template: "<div><slot /></div>" },
        RunaiExpansionWrapper: { template: "<div><slot /></div>" },
        RunaiClusterSection: true,
        YamlWorkloadSection: true,
        GeneralSection: true,
        Transition: { template: "<div><slot /></div>" },
      },
    },
  });
};

describe("WorkloadYamlNew - checkProjectSupportsMnnvl", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("returns false when project has no node pools", async () => {
    const wrapper = mountComponent();
    const project = makeProject([]);

    const result = await wrapper.vm.checkProjectSupportsMnnvl(project);

    expect(result).toBe(false);
    expect(clusterService.getNodepoolsMinimal).not.toHaveBeenCalled();
  });

  it("returns true when a project nodepool has gpuNetworkAccelerationDetected=true", async () => {
    vi.mocked(clusterService.getNodepoolsMinimal).mockResolvedValue([
      makeNodepool("np-1", true),
      makeNodepool("np-2", false),
    ]);

    const wrapper = mountComponent();
    const project = makeProject(["np-1"]);

    const result = await wrapper.vm.checkProjectSupportsMnnvl(project);

    expect(result).toBe(true);
  });

  it("returns false when project nodepool exists but detected=false", async () => {
    vi.mocked(clusterService.getNodepoolsMinimal).mockResolvedValue([makeNodepool("np-1", false)]);

    const wrapper = mountComponent();
    const project = makeProject(["np-1"]);

    const result = await wrapper.vm.checkProjectSupportsMnnvl(project);

    expect(result).toBe(false);
  });

  it("returns false when the detected nodepool is not in the project's nodepools", async () => {
    vi.mocked(clusterService.getNodepoolsMinimal).mockResolvedValue([makeNodepool("np-other", true)]);

    const wrapper = mountComponent();
    const project = makeProject(["np-1"]);

    const result = await wrapper.vm.checkProjectSupportsMnnvl(project);

    expect(result).toBe(false);
  });

  it("returns false and does not throw when getNodepoolsMinimal rejects", async () => {
    vi.mocked(clusterService.getNodepoolsMinimal).mockRejectedValue(new Error("network error"));

    const wrapper = mountComponent();
    const project = makeProject(["np-1"]);

    const result = await wrapper.vm.checkProjectSupportsMnnvl(project);

    expect(result).toBe(false);
  });

  it("returns true when at least one of multiple project nodepools is detected", async () => {
    vi.mocked(clusterService.getNodepoolsMinimal).mockResolvedValue([
      makeNodepool("np-1", false),
      makeNodepool("np-2", true),
      makeNodepool("np-3", false),
    ]);

    const wrapper = mountComponent();
    const project = makeProject(["np-1", "np-2", "np-3"]);

    const result = await wrapper.vm.checkProjectSupportsMnnvl(project);

    expect(result).toBe(true);
  });

  it("ignores undefined nodePool ids in project resources", async () => {
    vi.mocked(clusterService.getNodepoolsMinimal).mockResolvedValue([makeNodepool("np-1", true)]);

    const wrapper = mountComponent();
    // one resource has no nodePool
    const project = makeProject([undefined, "np-1"]);

    const result = await wrapper.vm.checkProjectSupportsMnnvl(project);

    expect(result).toBe(true);
  });
});

describe("WorkloadYamlNew - supportMnnvl computed", () => {
  it("returns the value of selectedProjectSupportsMnnvl", async () => {
    const wrapper = mountComponent();

    expect(wrapper.vm.supportMnnvl).toBe(false);

    wrapper.vm.selectedProjectSupportsMnnvl = true;
    await wrapper.vm.$nextTick();

    expect(wrapper.vm.supportMnnvl).toBe(true);
  });
});
