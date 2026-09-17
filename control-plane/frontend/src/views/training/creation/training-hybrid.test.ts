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
import { useRoute, type LocationQuery } from "vue-router";
import { AssetKind } from "@/swagger-models/assets-service-client";

import TrainingHybrid from "./training-hybrid.vue";
import { MIN_CLUSTER_VERSION_FOR_PRIORITY_CLASS } from "@/common/version.constant";
import { useAppStore } from "@/stores/app.store";
import { useAuthStore } from "@/stores/auth.store";
import { useClusterStore } from "@/stores/cluster.store";
import { useDataSourceStore } from "@/stores/data-source.store";
import { useTrainingStore } from "@/stores/training.store";
import { useWorkloadStore } from "@/stores/workload.store";
import { credentialService } from "@/services/control-plane/credential.service/credential.service";
import { dataSourceService } from "@/services/control-plane/data-source.service/data-source.service";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
import { trainingService } from "@/services/control-plane/training.service/training.service";
import { workloadService } from "@/services/cluster/workload.service/workload.service";
import { mockQuasar } from "@/utils/unit-test.util/unit-test.util";
import { workloadHybridUtil } from "@/utils/workload-hybrid.util/workload-hybrid.util";
import { deepCopy } from "@/utils/common.util";
import { EPolicyGroup } from "@/utils/policy.util/policy-view.util";
import type { IUISuperWorkloadCreationRequest } from "@/models/workload.model";
import type { Project } from "@/swagger-models/org-unit-service-client";
import { UidGidSource } from "@/swagger-models/workloads-client";
import { WorkloadCreationDrawerEntities } from "@/models/workload-creation-drawer-model";

vi.mock("@/services/control-plane/credential.service/credential.service");
vi.mock("@/services/control-plane/data-source.service/data-source.service");
vi.mock("@/services/control-plane/org-unit.service/org-unit.service");
vi.mock("@/services/control-plane/training.service/training.service");
vi.mock("@/services/cluster/workload.service/workload.service");
vi.mock("@/services/control-plane/policy.service/policy.service");

vi.mock("vue-router", () => ({
  useRoute: vi.fn(),
  useRouter: vi.fn(() => ({
    push: vi.fn(),
  })),
}));

describe("training-hybrid.vue", () => {
  let wrapper: VueWrapper<InstanceType<typeof TrainingHybrid>>;
  let pinia = createTestingPinia({ createSpy: vi.fn, stubActions: false });

  const copiedTrainingModel: IUISuperWorkloadCreationRequest = {
    name: "copy-training",
    projectId: "123",
    clusterId: "priority-only-cluster",
    spec: {
      image: "alpine",
      command: "sleep infinity",
      nodePools: ["default"],
      connections: [],
      priorityClass: "low",
      preemptibility: "preemptible",
    },
  };

  function seedFormState(): void {
    const trainingStore = useTrainingStore(pinia);
    const workloadModel = workloadHybridUtil.setTrainingUIDefaults(deepCopy(copiedTrainingModel), false, null, "2.23");

    trainingStore.setFlatFormState({
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

  function factory(
    query: LocationQuery = { clusterId: "priority-only-cluster", fromCopyId: "source-training-id" },
  ): void {
    const route = { query, params: {} };
    vi.mocked(useRoute).mockReturnValue(route as ReturnType<typeof useRoute>);
    wrapper = shallowMount(TrainingHybrid, {
      global: {
        plugins: [Quasar, pinia],
        mocks: {
          $route: route,
          $q: mockQuasar,
          $router: {
            push: vi.fn(),
          },
        },
      },
    });
  }

  async function remount(query: LocationQuery): Promise<void> {
    wrapper.unmount();
    factory(query);
    await flushPromises();
  }

  async function startNewWorkload(): Promise<void> {
    wrapper.vm.onBack();
    vi.mocked(workloadService.getUnifiedWorkload).mockResolvedValueOnce({
      spec: { worker: { image: "new-workload-image" } },
    });
    await remount({ clusterId: "priority-only-cluster", projectId: "123" });
  }

  beforeEach(() => {
    pinia = createTestingPinia({ createSpy: vi.fn, stubActions: false });
    setActivePinia(pinia);

    useAppStore(pinia);
    useAuthStore(pinia);
    useWorkloadStore(pinia);
    const clusterStore = useClusterStore(pinia);
    const dataSourceStore = useDataSourceStore(pinia);

    vi.mocked(clusterStore.isClusterVersionSufficient).mockImplementation((_, minVersion) => {
      return minVersion === MIN_CLUSTER_VERSION_FOR_PRIORITY_CLASS;
    });
    vi.mocked(dataSourceStore.getAvailableDataSources).mockResolvedValue(new Set());
    vi.mocked(dataSourceService.listExistingConfigMaps).mockResolvedValue([]);
    vi.mocked(credentialService.listCredentials).mockResolvedValue([]);
    vi.mocked(credentialService.listExistingSecrets).mockResolvedValue([]);
    vi.mocked(orgUnitService.getProject).mockResolvedValue({ name: "project-1" } as Project);
    vi.mocked(workloadService.canSubmitWorkload).mockResolvedValue(null);
    vi.mocked(workloadService.getWorkloadClusterPriorities).mockResolvedValue([{ id: "low", name: "low", value: 50 }]);
    vi.mocked(workloadService.getDefaultWorkloadPriorityByType).mockResolvedValue({ priorityName: "low" } as never);
    vi.mocked(trainingService.createTrainingFlat).mockResolvedValue({ workloadId: "created-id" } as never);

    seedFormState();
    factory();
  });

  afterEach(() => {
    wrapper.unmount();
    vi.clearAllMocks();
  });

  it("updateGeneralModel - should preserve priority and omit unsupported preemptibility", async () => {
    await flushPromises();

    wrapper.vm.updateGeneralModel({
      priorityClass: "low",
      preemptibility: "preemptible",
    });

    expect(wrapper.vm.trainingModel.spec.priorityClass).toBe("low");
    expect(wrapper.vm.trainingModel.spec.preemptibility).toBeUndefined();
  });

  it("onSubmit - should omit copied preemptibility while preserving priority on unsupported clusters", async () => {
    await flushPromises();

    await wrapper.vm.onSubmit();

    const request = vi.mocked(trainingService.createTrainingFlat).mock.calls[0][0];
    expect(request.spec?.priorityClass).toBe("low");
    expect(request.spec?.preemptibility).toBeUndefined();
  });

  it.each([
    { flow: "new", source: UidGidSource.FromIdpToken },
    { flow: "copy", source: UidGidSource.FromIdpToken },
    { flow: "new", source: UidGidSource.Custom },
    { flow: "copy", source: UidGidSource.Custom },
  ])("onSubmit - should serialize $source identity in a $flow workload", async ({ flow, source }) => {
    await flushPromises();
    if (flow === "new") await startNewWorkload();
    const originalIdentity = { runAsUid: 1001, runAsGid: 1002, supplementalGroups: "1003" };
    const policyIdentity = { runAsUid: 10001, runAsGid: 10002, supplementalGroups: "10003" };
    const security = { uidGidSource: source, ...originalIdentity };
    wrapper.vm.trainingModel.spec.security = { ...security };
    wrapper.vm.policy = { defaults: { security: policyIdentity } };

    await wrapper.vm.onSubmit();

    expect(trainingService.createTrainingFlat).toHaveBeenCalledOnce();
    const request = JSON.parse(JSON.stringify(vi.mocked(trainingService.createTrainingFlat).mock.calls[0][0]));
    expect(request.spec?.security).toEqual({
      uidGidSource: source,
      ...(source === UidGidSource.FromIdpToken ? {} : originalIdentity),
    });
    expect(wrapper.vm.trainingModel.spec.security).toEqual(security);
  });

  it.each([
    { asset: WorkloadCreationDrawerEntities.Environment, kind: AssetKind.Environment, outcome: "creating" },
    { asset: WorkloadCreationDrawerEntities.Environment, kind: AssetKind.Environment, outcome: "canceling" },
    { asset: WorkloadCreationDrawerEntities.Compute, kind: AssetKind.Compute, outcome: "creating" },
    { asset: WorkloadCreationDrawerEntities.Compute, kind: AssetKind.Compute, outcome: "canceling" },
  ])("omits copied IdP identity after $outcome an $asset", async ({ asset, kind, outcome }) => {
    await flushPromises();
    const security = {
      uidGidSource: UidGidSource.FromIdpToken,
      runAsUid: 1001,
      runAsGid: 1002,
      supplementalGroups: "1003",
    };
    wrapper.vm.trainingModel.spec.security = { ...security };
    wrapper.vm.drawerDataEntity = asset;
    wrapper.vm.onCreateNew(asset);

    await remount(outcome === "creating" ? { kind, createdEntityId: "new-asset-id" } : {});
    expect(wrapper.vm.$route.query.fromCopyId).toBeUndefined();
    expect(wrapper.vm.trainingModel.spec.security).toEqual(security);

    await wrapper.vm.onSubmit();

    expect(trainingService.createTrainingFlat).toHaveBeenCalledOnce();
    const request = JSON.parse(JSON.stringify(vi.mocked(trainingService.createTrainingFlat).mock.calls[0][0]));
    expect(request.spec.security).toEqual({ uidGidSource: UidGidSource.FromIdpToken });
    expect(wrapper.vm.trainingModel.spec.security).toEqual(security);
  });

  it("omits IdP identity after returning from asset creation in a new workload", async () => {
    await flushPromises();
    await startNewWorkload();
    const security = {
      uidGidSource: UidGidSource.FromIdpToken,
      runAsUid: 10001,
      runAsGid: 10002,
      supplementalGroups: "10003",
    };
    wrapper.vm.trainingModel.spec.security = { ...security };
    wrapper.vm.drawerDataEntity = WorkloadCreationDrawerEntities.Compute;
    wrapper.vm.onCreateNew(WorkloadCreationDrawerEntities.Compute);

    await remount({ kind: AssetKind.Compute, createdEntityId: "new-asset-id" });
    await wrapper.vm.onSubmit();

    expect(trainingService.createTrainingFlat).toHaveBeenCalledOnce();
    const request = JSON.parse(JSON.stringify(vi.mocked(trainingService.createTrainingFlat).mock.calls[0][0]));
    expect(request.spec.security).toEqual({ uidGidSource: UidGidSource.FromIdpToken });
  });
});
