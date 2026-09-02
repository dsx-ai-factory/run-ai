/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { VueWrapper } from "@vue/test-utils";
import { flushPromises, shallowMount } from "@vue/test-utils";
import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { Quasar } from "quasar";
import { setActivePinia } from "pinia";
import { createTestingPinia } from "@pinia/testing";

import WorkspaceHybrid from "./workspace-hybrid.vue";
import { useAppStore } from "@/stores/app.store";
import { useAuthStore } from "@/stores/auth.store";
import { useClusterStore } from "@/stores/cluster.store";
import { useDataSourceStore } from "@/stores/data-source.store";
import { useWorkspaceStore } from "@/stores/workspace.store";
import { useWorkloadStore } from "@/stores/workload.store";
import { credentialService } from "@/services/control-plane/credential.service/credential.service";
import { dataSourceService } from "@/services/control-plane/data-source.service/data-source.service";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
import { workspaceService } from "@/services/control-plane/workspace.service/workspace.service";
import { workloadService } from "@/services/cluster/workload.service/workload.service";
import { mockQuasar } from "@/utils/unit-test.util/unit-test.util";
import { workloadHybridUtil } from "@/utils/workload-hybrid.util/workload-hybrid.util";
import { deepCopy } from "@/utils/common.util";
import { EPolicyGroup } from "@/utils/policy.util/policy-view.util";
import type { IUISuperWorkloadCreationRequest } from "@/models/workload.model";
import type { Project } from "@/swagger-models/org-unit-service-client";

vi.mock("@/services/control-plane/credential.service/credential.service");
vi.mock("@/services/control-plane/data-source.service/data-source.service");
vi.mock("@/services/control-plane/org-unit.service/org-unit.service");
vi.mock("@/services/control-plane/workspace.service/workspace.service");
vi.mock("@/services/cluster/workload.service/workload.service");
vi.mock("@/services/control-plane/policy.service/policy.service");

vi.mock("vue-router", () => ({
  useRoute: vi.fn(() => ({
    query: {
      clusterId: "old-cluster",
      fromCopyId: "source-workspace-id",
    },
  })),
  useRouter: vi.fn(() => ({
    push: vi.fn(),
  })),
}));

describe("workspace-hybrid.vue", () => {
  let wrapper: VueWrapper<InstanceType<typeof WorkspaceHybrid>>;
  let pinia = createTestingPinia({ createSpy: vi.fn, stubActions: false });

  const copiedWorkspaceModel: IUISuperWorkloadCreationRequest = {
    name: "copy-workspace",
    projectId: "123",
    clusterId: "old-cluster",
    spec: {
      image: "alpine",
      command: "sleep infinity",
      nodePools: ["default"],
      connections: [],
      priorityClass: "interactive-preemptible",
      preemptibility: "non-preemptible",
    },
  };

  function seedFormState(): void {
    const workspaceStore = useWorkspaceStore(pinia);
    const workloadModel = workloadHybridUtil.setWorkspaceUIDefaults(deepCopy(copiedWorkspaceModel), false, null, "2.23");

    workspaceStore.setFlatFormState({
      projectInfo: { name: "project-1" },
      projectNodePoolsResources: [],
      selectedTemplates: workloadHybridUtil.getWorkloadHybridTemplatesSetup(),
      policy: null,
      policyViewDetails: {
        [EPolicyGroup.Environment]: {},
        [EPolicyGroup.Compute]: {},
        [EPolicyGroup.Storage]: {},
        [EPolicyGroup.General]: {},
      },
      selectedProjectNodeTypes: [],
      workloadModel,
    });
  }

  function factory(): void {
    wrapper = shallowMount(WorkspaceHybrid, {
      global: {
        plugins: [Quasar, pinia],
        mocks: {
          $route: {
            query: {
              clusterId: "old-cluster",
              fromCopyId: "source-workspace-id",
            },
          },
          $q: mockQuasar,
          $router: {
            push: vi.fn(),
          },
        },
      },
    });
  }

  beforeEach(() => {
    pinia = createTestingPinia({ createSpy: vi.fn, stubActions: false });
    setActivePinia(pinia);

    useAppStore(pinia);
    useAuthStore(pinia);
    useWorkloadStore(pinia);
    const clusterStore = useClusterStore(pinia);
    const dataSourceStore = useDataSourceStore(pinia);

    vi.mocked(clusterStore.isClusterVersionSufficient).mockReturnValue(false);
    vi.mocked(dataSourceStore.getAvailableDataSources).mockResolvedValue(new Set());
    vi.mocked(dataSourceService.listExistingConfigMaps).mockResolvedValue([]);
    vi.mocked(credentialService.listCredentials).mockResolvedValue([]);
    vi.mocked(credentialService.listExistingSecrets).mockResolvedValue([]);
    vi.mocked(orgUnitService.getProject).mockResolvedValue({ name: "project-1" } as Project);
    vi.mocked(workloadService.canSubmitWorkload).mockResolvedValue(null);
    vi.mocked(workspaceService.createWorkspaceFlat).mockResolvedValue({ workloadId: "created-id" } as never);

    seedFormState();
    factory();
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it("updateGeneralModel - should preserve legacy allow over quota priority and omit unsupported preemptibility", async () => {
    await flushPromises();

    wrapper.vm.updateGeneralModel({
      allowOverQuota: true,
      priorityClass: "ignored-priority",
    });

    expect(wrapper.vm.workspaceModel.spec.priorityClass).toBe("interactive-preemptible");
    expect(wrapper.vm.workspaceModel.spec.preemptibility).toBeUndefined();
  });

  it("onSubmit - should omit copied preemptibility while preserving legacy priority on unsupported clusters", async () => {
    await flushPromises();

    await wrapper.vm.onSubmit();

    const request = vi.mocked(workspaceService.createWorkspaceFlat).mock.calls[0][0];
    expect(request.spec?.priorityClass).toBe("interactive-preemptible");
    expect(request.spec?.preemptibility).toBeUndefined();
  });
});
