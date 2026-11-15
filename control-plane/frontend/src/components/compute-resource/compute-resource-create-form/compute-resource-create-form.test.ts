/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { VueWrapper } from "@vue/test-utils";
import { flushPromises, mount } from "@vue/test-utils";
import ComputeCreateForm from "./compute-resource-create-form.vue";
import { afterEach, beforeEach, describe, expect, test, vi, type MockedFunction } from "vitest";
import type { DefineComponent } from "vue";
import { setActivePinia } from "pinia";
import { createTestingPinia } from "@pinia/testing";
import { Quasar } from "quasar";
import { GpuComputeMode, GpuRequestType, Scope } from "@/swagger-models/assets-service-client";
import type { IProjectSectionModel } from "@/models/project.model";
import { assetService } from "@/services/control-plane/asset.service/asset.service";
import type { IUIComputeResourceCreation } from "@/models/compute-resource.model";
import { computeResourceService } from "@/services/control-plane/compute-resource.service/compute-resource.service";
import { orgTreeService } from "@/services/control-plane/rbac/org-tree.service/org-tree.service";
import { orgTreeServiceMock } from "@/services/control-plane/rbac/org-tree.service/org-tree.service.mock";
import { useClusterStore } from "@/stores/cluster.store";

vi.mock("@/services/control-plane/asset.service/asset.service");

const pinia = createTestingPinia({ createSpy: vi.fn });

setActivePinia(pinia);

const validComputeModel: IUIComputeResourceCreation = {
  meta: {
    name: "test-101-name",
    scope: Scope.Project,
    projectId: 25,
  },
  spec: {
    gpuDevicesRequest: 1,
    gpuRequestType: GpuRequestType.Portion,
    gpuPortionRequest: 0.8,
    gpuPortionLimit: 0.8,
    cpuCoreRequest: 0.08,
    cpuMemoryRequest: "80M",
    largeShmRequest: false,
  },
};
describe("compute-resource-create-form.vue", () => {
  let computeCreateFormCmp: VueWrapper<InstanceType<typeof ComputeCreateForm>>;

  const mockIsExist = assetService.isExist as MockedFunction<typeof assetService.isExist>;

  const emptyComputeModel: IUIComputeResourceCreation = computeResourceService.getEmptyComputeResourceModel();

  const spyGetOrgTree = vi.spyOn(orgTreeService, "getOrgTree");

  let clusterStore = useClusterStore(pinia);
  beforeEach(() => {
    spyGetOrgTree.mockReturnValue(orgTreeServiceMock.orgTree);
    clusterStore = useClusterStore(pinia);
    clusterStore.$patch({ currentCluster: { uuid: "34234-322-sdfs323" } });
    clusterStore.isClusterVersionSufficient = vi.fn().mockReturnValue(true);
    clusterStore.isAllClustersSupportSpecificVersion = vi.fn().mockReturnValue(true);

    computeCreateFormCmp = mount(ComputeCreateForm, {
      global: {
        plugins: [Quasar, pinia],
        mocks: {
          $route: {
            query: "",
          },
        },
      },
      props: {
        submitting: false,
        initialModel: emptyComputeModel as IUIComputeResourceCreation,
        isScopeSelectable: true,
      },
    });
    computeCreateFormCmp.vm.$q.notify = vi.fn(); // disabling notify to prevent it from failing tests
  });
  afterEach(() => {
    vi.clearAllMocks();
    vi.resetAllMocks();
  });

  test("component is loading properly", () => {
    expect(computeCreateFormCmp.html()).toContain("Compute resource");
    expect(computeCreateFormCmp.classes()).toContain("compute-resource-create-form");
  });

  test("form validation should fail when name is invalid", async () => {
    validComputeModel.meta.name = "";
    await computeCreateFormCmp.setData({ computeModel: validComputeModel });
    expect(await computeCreateFormCmp.vm.validate()).toBe(false);

    validComputeModel.meta.name = "1-3";
    await computeCreateFormCmp.setData({ computeModel: validComputeModel });
    expect(await computeCreateFormCmp.vm.validate()).toBe(false);

    validComputeModel.meta.name = "test-name-101";
    await computeCreateFormCmp.setData({ computeModel: validComputeModel });
    expect(await computeCreateFormCmp.vm.validate()).toBe(true);
  });

  test("updates the projectId & scope in the model when a new projectId & scope are emitted by project-section", async () => {
    const newScope: IProjectSectionModel = {
      scope: Scope.Project,
      projectId: 12,
    };
    const scopeSectionCmp = computeCreateFormCmp.findComponent<DefineComponent>("[aid='runai-scope-section']");
    scopeSectionCmp.vm.$emit("changed-scope", newScope);
    await computeCreateFormCmp.vm.$nextTick();
    expect(computeCreateFormCmp.vm.computeModel.meta.projectId).toBe(newScope.projectId);
    expect(computeCreateFormCmp.vm.computeModel.meta.scope).toBe(newScope.scope);
  });

  test("failing to submit the form when button is clicked and the form is invalid", async () => {
    // selecting scope so the rest of the form apears
    await computeCreateFormCmp.setData({
      computeModel: {
        meta: { ...computeCreateFormCmp.vm.computeModel.meta, scope: Scope.Project, projectId: 1 },
        spec: computeCreateFormCmp.vm.computeModel.spec,
      },
    });
    await computeCreateFormCmp.vm.$nextTick();

    // trigger submit
    computeCreateFormCmp.find("[aid=form-create-compute-btn]").trigger("click");

    expect(computeCreateFormCmp.emitted("submit")).toBeFalsy();
  });

  test("successfuly submitting the form when button is clicked and the form is valid", async () => {
    mockIsExist.mockResolvedValue(false);
    const validModel: IUIComputeResourceCreation = {
      meta: {
        name: "test-101-name",
        scope: Scope.Project,
        projectId: 25,
      },
      spec: {
        gpuDevicesRequest: 1,
        gpuRequestType: GpuRequestType.Portion,
        gpuPortionRequest: 0.8,
        gpuPortionLimit: 0.8,
        cpuCoreRequest: 0.08,
        cpuMemoryRequest: "80M",
        largeShmRequest: false,
      },
    };
    await computeCreateFormCmp.setData({ computeModel: validModel });
    await flushPromises();
    computeCreateFormCmp.find("[aid=form-create-compute-btn]").trigger("click");
    await flushPromises();
    await computeCreateFormCmp.vm.$nextTick();
    expect(computeCreateFormCmp.emitted().submit).toBeTruthy();
  });

  test.each([
    { name: "edited", devices: 1, request: 0.25, limit: 0.75, requestType: GpuRequestType.Portion },
    { name: "cleared", devices: 2, request: undefined, limit: undefined, requestType: undefined },
  ])(
    "submits $name memory portions without stale response aliases",
    async ({ devices, request, limit, requestType }) => {
      vi.spyOn(computeCreateFormCmp.vm, "validate").mockResolvedValue(true);
      const model = {
        meta: { name: "compute-alias-test", scope: Scope.Project, projectId: 25 },
        spec: {
          gpuDevicesRequest: 1,
          gpuRequestType: GpuRequestType.Portion,
          gpuPortionRequest: 0.5,
          gpuPortionLimit: 0.5,
          gpuMemoryPortionRequest: 0.5,
          gpuMemoryPortionLimit: 0.5,
          gpuComputeMode: GpuComputeMode.SmSharing,
          cpuCoreRequest: 2,
        },
      };
      await computeCreateFormCmp.setData({ computeModel: model });
      computeCreateFormCmp.vm.updateResourceSectionModel({
        gpuDevicesRequest: devices,
        gpuRequestType: requestType,
        gpuPortionRequest: request,
        gpuPortionLimit: limit,
        cpuCore: { request: 2 },
        cpuMemory: {},
      });
      await computeCreateFormCmp.vm.save();

      const submitted = computeCreateFormCmp.emitted<[IUIComputeResourceCreation]>("submit")?.[0][0];
      expect(submitted?.spec).toMatchObject({
        gpuDevicesRequest: devices,
        gpuPortionRequest: request,
        gpuPortionLimit: limit,
        gpuComputeMode: GpuComputeMode.SmSharing,
        cpuCoreRequest: 2,
      });
      expect(submitted?.spec).not.toHaveProperty("gpuMemoryPortionRequest");
      expect(submitted?.spec).not.toHaveProperty("gpuMemoryPortionLimit");
    },
  );

  test("isExists is called when submitting the form", async () => {
    mockIsExist.mockResolvedValue(false);
    const validModel: IUIComputeResourceCreation = {
      meta: {
        name: "test-101-name",
        scope: Scope.Project,
        projectId: 25,
        clusterId: null,
      },
      spec: {
        gpuDevicesRequest: 1,
        gpuRequestType: GpuRequestType.Portion,
        gpuPortionRequest: 0.8,
        gpuPortionLimit: 0.8,
        cpuCoreRequest: 0.08,
        cpuMemoryRequest: "80M",
        largeShmRequest: false,
      },
    };
    await computeCreateFormCmp.setData({ computeModel: validModel });
    computeCreateFormCmp.find("[aid=form-create-compute-btn]").trigger("click");
    await flushPromises();
    expect(mockIsExist).toHaveBeenLastCalledWith("compute", validModel.meta.name, {
      scopeType: Scope.Project,
      scopeId: "25",
    });
  });
});
