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
import { UidGidSource } from "@/swagger-models/workloads-client";
import { SettingKeys } from "@/models/setting.model";
import { WorkloadCreationDrawerEntities } from "@/models/workload-creation-drawer-model";

vi.mock("@/services/control-plane/credential.service/credential.service");
vi.mock("@/services/control-plane/data-source.service/data-source.service");
vi.mock("@/services/control-plane/org-unit.service/org-unit.service");
vi.mock("@/services/control-plane/workspace.service/workspace.service");
vi.mock("@/services/cluster/workload.service/workload.service");
vi.mock("@/services/control-plane/policy.service/policy.service");

vi.mock("vue-router", () => ({
  useRoute: vi.fn(),
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

  function factory(query: LocationQuery = { clusterId: "old-cluster", fromCopyId: "source-workspace-id" }): void {
    const route = { query, params: {} };
    vi.mocked(useRoute).mockReturnValue(route as ReturnType<typeof useRoute>);
    wrapper = shallowMount(WorkspaceHybrid, {
      global: {
        plugins: [Quasar, pinia],
        stubs: {
          WorkloadFormWrapper: {
            template: "<div />",
            methods: { validate: vi.fn() },
          },
        },
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
    await remount({ clusterId: "old-cluster", projectId: "123" });
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
    wrapper.unmount();
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
    wrapper.vm.workspaceModel.spec.security = { ...security };
    wrapper.vm.policy = { defaults: { security: policyIdentity } };

    await wrapper.vm.onSubmit();

    expect(workspaceService.createWorkspaceFlat).toHaveBeenCalledOnce();
    const request = JSON.parse(JSON.stringify(vi.mocked(workspaceService.createWorkspaceFlat).mock.calls[0][0]));
    expect(request.spec?.security).toEqual({
      uidGidSource: source,
      ...(source === UidGidSource.FromIdpToken ? {} : originalIdentity),
    });
    expect(wrapper.vm.workspaceModel.spec.security).toEqual(security);
  });

  it("expands a workload template and omits its IdP identity before submission", async () => {
    await flushPromises();
    wrapper.vm.onBack();
    vi.mocked(wrapper.vm.settingStore.isFeatureEnabled).mockImplementation(
      (key) => key === SettingKeys.EnableTemplatesV2,
    );
    const security = {
      uidGidSource: UidGidSource.FromIdpToken,
      runAsUid: 1001,
      runAsGid: 1002,
      supplementalGroups: "1003",
    };
    vi.mocked(workloadService.getUnifiedWorkload).mockResolvedValue({
      spec: { worker: { image: "template-image", security } },
    });

    await remount({
      clusterId: "old-cluster",
      projectId: "123",
      workspaceName: "from-template",
      templateId: "template-id",
    });

    expect(workloadService.getUnifiedWorkload).toHaveBeenCalledWith(
      expect.objectContaining({ templateId: "template-id" }),
    );
    expect(wrapper.vm.workspaceModel.spec.security).toMatchObject(security);
    await wrapper.vm.onSubmit();

    const request = JSON.parse(JSON.stringify(vi.mocked(workspaceService.createWorkspaceFlat).mock.calls[0][0]));
    expect(request.spec.image).toBe("template-image");
    expect(request.spec.security).toEqual({ uidGidSource: UidGidSource.FromIdpToken, capabilities: [] });
    expect(request).not.toHaveProperty("templateId");
    expect(request).not.toHaveProperty("assets");
  });

  it("expands an environment asset selected while copying without sending its reference", async () => {
    await flushPromises();
    const security = {
      uidGidSource: UidGidSource.FromIdpToken,
      runAsUid: 1001,
      runAsGid: 1002,
      supplementalGroups: "1003",
    };
    vi.mocked(workloadService.getUnifiedWorkload).mockResolvedValue({
      spec: { worker: { image: "asset-image", security } },
    });
    wrapper.vm.policy = {
      defaults: { security: { runAsUid: 10001, runAsGid: 10002, supplementalGroups: "10003" } },
    };
    wrapper.vm.drawerDataEntity = WorkloadCreationDrawerEntities.Environment;

    await wrapper.vm.handleSelection({ id: "environment-id", name: "environment" });

    expect(workloadService.getUnifiedWorkload).toHaveBeenCalledWith(
      expect.objectContaining({ assetId: "environment-id" }),
    );
    expect(wrapper.vm.workspaceModel.spec.security).toMatchObject(security);
    await wrapper.vm.onSubmit();

    const request = JSON.parse(JSON.stringify(vi.mocked(workspaceService.createWorkspaceFlat).mock.calls[0][0]));
    expect(request.spec.image).toBe("asset-image");
    expect(request.spec.security.uidGidSource).toBe(UidGidSource.FromIdpToken);
    expect(request.spec.security).not.toHaveProperty("runAsUid");
    expect(request.spec.security).not.toHaveProperty("runAsGid");
    expect(request.spec.security).not.toHaveProperty("supplementalGroups");
    expect(request).not.toHaveProperty("templateId");
    expect(request).not.toHaveProperty("assets");
    expect(wrapper.vm.workspaceModel.spec.security).toMatchObject(security);
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
    wrapper.vm.workspaceModel.spec.security = { ...security };
    wrapper.vm.drawerDataEntity = asset;
    wrapper.vm.onCreateNew(asset);

    await remount(outcome === "creating" ? { kind, createdEntityId: "new-asset-id" } : {});
    expect(wrapper.vm.$route.query.fromCopyId).toBeUndefined();
    expect(wrapper.vm.workspaceModel.spec.security).toEqual(security);

    await wrapper.vm.onSubmit();

    expect(workspaceService.createWorkspaceFlat).toHaveBeenCalledOnce();
    const request = JSON.parse(JSON.stringify(vi.mocked(workspaceService.createWorkspaceFlat).mock.calls[0][0]));
    expect(request.spec.security).toEqual({ uidGidSource: UidGidSource.FromIdpToken });
    expect(wrapper.vm.workspaceModel.spec.security).toEqual(security);
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
    wrapper.vm.workspaceModel.spec.security = { ...security };
    wrapper.vm.drawerDataEntity = WorkloadCreationDrawerEntities.Compute;
    wrapper.vm.onCreateNew(WorkloadCreationDrawerEntities.Compute);

    await remount({ kind: AssetKind.Compute, createdEntityId: "new-asset-id" });
    await wrapper.vm.onSubmit();

    expect(workspaceService.createWorkspaceFlat).toHaveBeenCalledOnce();
    const request = JSON.parse(JSON.stringify(vi.mocked(workspaceService.createWorkspaceFlat).mock.calls[0][0]));
    expect(request.spec.security).toEqual({ uidGidSource: UidGidSource.FromIdpToken });
  });
});
