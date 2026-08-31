/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { VueWrapper } from "@vue/test-utils";
import { flushPromises, mount } from "@vue/test-utils";
import NfsCreateForm from "./nfs-create-form.vue";
import { afterEach, beforeEach, describe, expect, test, vi, type MockedFunction } from "vitest";
import type { DefineComponent } from "vue";
import { setActivePinia } from "pinia";
import { createTestingPinia } from "@pinia/testing";
import { Quasar } from "quasar";
import { Scope, type NFSCreationRequest } from "@/swagger-models/assets-service-client";
import type { IProjectSectionModel } from "@/models/project.model";
import { assetService } from "@/services/control-plane/asset.service/asset.service";
import { dataSourceService } from "@/services/control-plane/data-source.service/data-source.service";
import { orgTreeService } from "@/services/control-plane/rbac/org-tree.service/org-tree.service";
import { orgTreeServiceMock } from "@/services/control-plane/rbac/org-tree.service/org-tree.service.mock";
import { useClusterStore } from "@/stores/cluster.store";
import type { IUINFSCreationRequest } from "@/models/data-source.model";

vi.mock("@/services/control-plane/asset.service/asset.service");

const pinia = createTestingPinia({ createSpy: vi.fn });

const validNfsModel: IUINFSCreationRequest = {
  meta: {
    name: "test101",
    scope: Scope.Project,
    clusterId: null,
    projectId: 25,
  },
  spec: {
    path: "/test",
    readOnly: false,
    server: "test-server",
    mountPath: "/test-mount-path",
  },
};

setActivePinia(pinia);
describe("nfs-create-form.vue", () => {
  let nfsCreateFormCmp: VueWrapper<InstanceType<typeof NfsCreateForm>>;
  const emptyNfsModel = dataSourceService.getEmptyNfsModel();

  const mockIsExist = assetService.isExist as MockedFunction<typeof assetService.isExist>;

  let clusterStore = useClusterStore(pinia);

  const spyGetOrgTree = vi.spyOn(orgTreeService, "getOrgTree");

  beforeEach(() => {
    spyGetOrgTree.mockReturnValue(orgTreeServiceMock.orgTree);

    clusterStore = useClusterStore(pinia);
    clusterStore.$patch({ currentCluster: { uuid: "34234-322-sdfs323" } });
    clusterStore.isClusterVersionSufficient = vi.fn().mockReturnValue(true);
    clusterStore.isAllClustersSupportSpecificVersion = vi.fn().mockReturnValue(true);

    nfsCreateFormCmp = mount(NfsCreateForm, {
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
        dataSourceModel: emptyNfsModel,
        isScopeSelectable: false,
      },
    });
    nfsCreateFormCmp.vm.$q.notify = vi.fn(); // disabling notify to prevent it from failing tests
  });
  afterEach(() => {
    vi.clearAllMocks();
    vi.resetAllMocks();
  });

  test("component is loading properly", () => {
    expect(nfsCreateFormCmp.html()).toContain("nfs-create-form");
  });

  test("form validation should fail when the name is invalid", async () => {
    validNfsModel.meta.name = "";
    await nfsCreateFormCmp.setData({ nfsModel: validNfsModel });
    expect(await nfsCreateFormCmp.vm.validate()).toBe(false);

    validNfsModel.meta.name = "101";
    await nfsCreateFormCmp.setData({ nfsModel: validNfsModel });
    expect(await nfsCreateFormCmp.vm.validate()).toBe(false);

    validNfsModel.meta.name = "test101";
    await nfsCreateFormCmp.setData({ nfsModel: validNfsModel });
    expect(await nfsCreateFormCmp.vm.validate()).toBe(true);
  });

  test("form validation should fail when the data mount section is invalid", async () => {
    validNfsModel.spec.path = "";
    await nfsCreateFormCmp.setData({ nfsModel: validNfsModel });
    expect(await nfsCreateFormCmp.vm.validate()).toBe(false);

    validNfsModel.spec.path = "/test";
    await nfsCreateFormCmp.setData({ nfsModel: validNfsModel });
    expect(await nfsCreateFormCmp.vm.validate()).toBe(true);

    validNfsModel.spec.server = "";
    await nfsCreateFormCmp.setData({ nfsModel: validNfsModel });
    expect(await nfsCreateFormCmp.vm.validate()).toBe(false);

    validNfsModel.spec.server = "test-server";
    await nfsCreateFormCmp.setData({ nfsModel: validNfsModel });
    expect(await nfsCreateFormCmp.vm.validate()).toBe(true);

    validNfsModel.spec.mountPath = "";
    await nfsCreateFormCmp.setData({ nfsModel: validNfsModel });
    expect(await nfsCreateFormCmp.vm.validate()).toBe(false);

    validNfsModel.spec.mountPath = "/test-mount-path";
    await nfsCreateFormCmp.setData({ nfsModel: validNfsModel });
    expect(await nfsCreateFormCmp.vm.validate()).toBe(true);
  });

  test("updates the projectId & scope in the model when a new projectId & scope are emitted by project-section", async () => {
    const newScope: IProjectSectionModel = {
      scope: Scope.Project,
      projectId: 12,
    };
    const scopeSectionCmp = nfsCreateFormCmp.findComponent<DefineComponent>("[aid='runai-scope-section']");
    scopeSectionCmp.vm.$emit("changed-scope", newScope);
    await nfsCreateFormCmp.vm.$nextTick();
    expect(nfsCreateFormCmp.vm.nfsModel.meta.projectId).toBe(newScope.projectId);
    expect(nfsCreateFormCmp.vm.nfsModel.meta.scope).toBe(newScope.scope);
  });

  test("failing to submit the form when button is clicked and the form is invalid", async () => {
    // selecting scope so the rest of the form apears
    await nfsCreateFormCmp.setData({
      nfsModel: {
        meta: { ...nfsCreateFormCmp.vm.nfsModel.meta, scope: Scope.Project, projectId: 1 },
        spec: nfsCreateFormCmp.vm.nfsModel.spec,
      },
    });
    await nfsCreateFormCmp.vm.$nextTick();

    nfsCreateFormCmp.find("[aid=submit-btn]").trigger("click");
    await flushPromises();
    expect(nfsCreateFormCmp.emitted("submit")).toBeFalsy();
  });

  test("successfuly submitting the form when button is clicked and the form is valid", async () => {
    const validModel: NFSCreationRequest = {
      meta: {
        name: "test101",
        scope: Scope.Project,
        clusterId: null,
        projectId: 25,
      },
      spec: {
        path: "/test",
        readOnly: false,
        server: "test-server",
        mountPath: "/test-mount-path",
      },
    };
    await nfsCreateFormCmp.setData({ nfsModel: validModel });
    nfsCreateFormCmp.find("[aid=submit-btn]").trigger("click");
    await flushPromises();
    expect(nfsCreateFormCmp.emitted().submit).toBeTruthy();
  });

  test("isExists is called when submitting the form", async () => {
    mockIsExist.mockResolvedValue(false);
    const validModel: NFSCreationRequest = {
      meta: {
        name: "test101",
        scope: Scope.Project,
        clusterId: null,
        projectId: 25,
      },
      spec: {
        path: "/test",
        readOnly: false,
        server: "test-server",
        mountPath: "/test-mount-path",
      },
    };
    await nfsCreateFormCmp.setData({ nfsModel: validModel });
    nfsCreateFormCmp.find("[aid=submit-btn]").trigger("click");
    await flushPromises();
    expect(mockIsExist).toHaveBeenLastCalledWith("nfs", validModel.meta.name, {
      scopeType: Scope.Project,
      scopeId: "25",
    });
  });
});
