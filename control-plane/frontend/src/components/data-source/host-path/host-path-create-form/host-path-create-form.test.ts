/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { VueWrapper } from "@vue/test-utils";
import { flushPromises, mount } from "@vue/test-utils";
import { HostPathCreateForm } from ".";
import { afterEach, beforeEach, describe, expect, test, vi, type MockedFunction } from "vitest";
import { setActivePinia } from "pinia";
import { createTestingPinia } from "@pinia/testing";
import { Quasar } from "quasar";
import { Scope } from "@/swagger-models/assets-service-client";
import { assetService } from "@/services/control-plane/asset.service/asset.service";
import { orgTreeService } from "@/services/control-plane/rbac/org-tree.service/org-tree.service";
import { orgTreeServiceMock } from "@/services/control-plane/rbac/org-tree.service/org-tree.service.mock";
import type { DefineComponent } from "vue";
import { useClusterStore } from "@/stores/cluster.store";
import type { IUIHostPathCreationRequest } from "@/models/data-source.model";

vi.mock("@/services/control-plane/asset.service/asset.service");

const pinia = createTestingPinia({ createSpy: vi.fn });
setActivePinia(pinia);

const validHostPathModel: IUIHostPathCreationRequest = {
  meta: {
    name: "test-name",
    scope: Scope.Project,
    clusterId: null,
    projectId: 25,
  },
  spec: {
    path: "/valid-path",
    readOnly: false,
    mountPath: "/valid-mount-path",
  },
};

describe("host-path-create-form.vue", () => {
  let hostPathCreateForm: VueWrapper<InstanceType<typeof HostPathCreateForm>>;
  const mockIsExist = assetService.isExist as MockedFunction<typeof assetService.isExist>;

  const spyGetOrgTree = vi.spyOn(orgTreeService, "getOrgTree");

  let clusterStore = useClusterStore(pinia);
  beforeEach(() => {
    spyGetOrgTree.mockReturnValue(orgTreeServiceMock.orgTree);

    clusterStore = useClusterStore(pinia);
    clusterStore.isClusterVersionSufficient = vi.fn().mockReturnValue(true);
    clusterStore.isAllClustersSupportSpecificVersion = vi.fn().mockReturnValue(true);
    clusterStore.$patch({ currentCluster: { uuid: "34234-322-sdfs323" } });
    hostPathCreateForm = mount(HostPathCreateForm, {
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
        isScopeSelectable: false,
        dataSourceModel: {
          meta: {
            name: "",
            scope: Scope.Tenant,
            clusterId: null,
            projectId: null,
          },
          spec: {
            path: "",
            readOnly: false,
            mountPath: "",
          },
        },
        allowedScopes: {},
      },
    });
    hostPathCreateForm.vm.$q.notify = vi.fn();
  });

  afterEach(() => {
    vi.clearAllMocks();
    vi.resetAllMocks();
  });

  test("component is loading properly", () => {
    expect(hostPathCreateForm.html()).toContain("host-path-create-form");
  });

  test("form validation should fail when name is invalid", async () => {
    validHostPathModel.meta.name = "";
    await hostPathCreateForm.setData({ hostPathModel: validHostPathModel });
    expect(await hostPathCreateForm.vm.validate()).toBe(false);

    validHostPathModel.meta.name = "1-3";
    await hostPathCreateForm.setData({ hostPathModel: validHostPathModel });
    expect(await hostPathCreateForm.vm.validate()).toBe(false);

    validHostPathModel.meta.name = "valid-name";
    await hostPathCreateForm.setData({ hostPathModel: validHostPathModel });
    expect(await hostPathCreateForm.vm.validate()).toBe(true);
  });

  test("form validation should fail when data mount section is invalid", async () => {
    validHostPathModel.spec.path = "";
    await hostPathCreateForm.setData({ hostPathModel: validHostPathModel });
    expect(await hostPathCreateForm.vm.validate()).toBe(false);

    validHostPathModel.spec.path = "/valid-path";
    await hostPathCreateForm.setData({ hostPathModel: validHostPathModel });
    expect(await hostPathCreateForm.vm.validate()).toBe(true);

    validHostPathModel.spec.mountPath = "";
    await hostPathCreateForm.setData({ hostPathModel: validHostPathModel });
    expect(await hostPathCreateForm.vm.validate()).toBe(false);

    validHostPathModel.spec.mountPath = "/valid-mount-path";
    await hostPathCreateForm.setData({ hostPathModel: validHostPathModel });
    expect(await hostPathCreateForm.vm.validate()).toBe(true);
  });

  test("failing to submit the form when button is clicked and the form is invalid", async () => {
    hostPathCreateForm.find("[aid=submit-btn]").trigger("click");
    await flushPromises();
    expect(hostPathCreateForm.emitted("submit")).toBeFalsy();
  });

  test("successfuly submitting the form when button is clicked and the form is valid", async () => {
    mockIsExist.mockResolvedValue(false);
    const validModel = {
      meta: {
        name: "test-name",
        scope: Scope.Project,
        clusterId: null,
        projectId: 25,
      },
      spec: {
        path: "/valid-path",
        readOnly: false,
        mountPath: "/valid-mount-path",
      },
    };
    await hostPathCreateForm.setData({ hostPathModel: validModel });
    await hostPathCreateForm.find("[aid=submit-btn]").trigger("click");
    await flushPromises();
    expect(hostPathCreateForm.emitted().submit).toBeTruthy();
  });

  test("isExists is called when submitting the form", async () => {
    mockIsExist.mockResolvedValue(false);
    const validModel = {
      meta: {
        name: "test-check-name",
        scope: Scope.Tenant,
        clusterId: null,
      },
      spec: {
        path: "/valid-path",
        readOnly: false,
        mountPath: "/valid-mount-path",
      },
    };
    await hostPathCreateForm.setData({ hostPathModel: validModel });
    const scopeSectionCmp = hostPathCreateForm.findComponent<DefineComponent>("[aid=runai-scope-section]");

    await scopeSectionCmp.setData({ scope: "project" });
    hostPathCreateForm.find("[aid=submit-btn]").trigger("click");
    await flushPromises();
    expect(mockIsExist).toHaveBeenLastCalledWith("hostPath", validModel.meta.name, {
      scopeType: Scope.Tenant,
    });
  });
});
