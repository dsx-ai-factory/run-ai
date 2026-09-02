/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { VueWrapper } from "@vue/test-utils";
import { flushPromises, mount } from "@vue/test-utils";
import NgcApiKeyCreateForm from "./ngc-api-key-create-form.vue";
import { afterEach, beforeEach, describe, expect, test, vi, type MockedFunction } from "vitest";
import { setActivePinia } from "pinia";
import { createTestingPinia } from "@pinia/testing";
import { Quasar } from "quasar";
import { type NgcApiKeyCreationRequest, Scope } from "@/swagger-models/assets-service-client";
import { assetService } from "@/services/control-plane/asset.service/asset.service";
import { credentialService } from "@/services/control-plane/credential.service/credential.service";
import { orgTreeService } from "@/services/control-plane/rbac/org-tree.service/org-tree.service";
import { orgTreeServiceMock } from "@/services/control-plane/rbac/org-tree.service/org-tree.service.mock";
import { useClusterStore } from "@/stores/cluster.store";
import type { IUINgcApiKeyCreationRequest } from "@/models/credential.model";
import type { PermittedScopes } from "@/swagger-models/authorization-client";

vi.mock("@/services/control-plane/asset.service/asset.service");

const pinia = createTestingPinia({ createSpy: vi.fn });

setActivePinia(pinia);

const validNgcApiKeyModel: IUINgcApiKeyCreationRequest = {
  meta: {
    name: "ngc-api-key-1",
    scope: Scope.Project,
    projectId: 25,
  },
  spec: {
    ngc_api_key: "NgcApiKeyPlaceholder",
  },
};

describe("ngc-api-key-create-form.vue", () => {
  let ngcApiKeyCreateFormCmp: VueWrapper<InstanceType<typeof NgcApiKeyCreateForm>>;

  const mockIsExist = assetService.isExist as MockedFunction<typeof assetService.isExist>;

  const emptyNgcApiKeyModel = credentialService.getEmptyNgcApiKeyModel();

  const spyGetOrgTree = vi.spyOn(orgTreeService, "getOrgTree");

  let clusterStore = useClusterStore(pinia);

  beforeEach(() => {
    spyGetOrgTree.mockReturnValue(orgTreeServiceMock.orgTree);

    clusterStore = useClusterStore(pinia);
    clusterStore.$patch({ currentCluster: { uuid: "34234-322-sdfs323" } });
    clusterStore.isClusterVersionSufficient = vi.fn().mockReturnValue(true);
    clusterStore.isAllClustersSupportSpecificVersion = vi.fn().mockReturnValue(true);

    ngcApiKeyCreateFormCmp = mount(NgcApiKeyCreateForm, {
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
        credentialModel: emptyNgcApiKeyModel,
        allowedScopes: {
          clusters: ["cluster-1"],
          departments: ["dept-1"],
          projects: ["project-1"],
        } as PermittedScopes,
      },
    });
    ngcApiKeyCreateFormCmp.vm.$q.notify = vi.fn(); // disabling notify to prevent it from failing tests
  });

  afterEach(() => {
    vi.clearAllMocks();
    vi.resetAllMocks();
  });

  test("component is loading properly", () => {
    expect(ngcApiKeyCreateFormCmp.html()).toContain("ngc-api-key-create-form");
  });

  test("the form validation should fail when the name is invalid", async () => {
    const ngcApiKeyModel = validNgcApiKeyModel;

    ngcApiKeyModel.meta.name = "";
    await ngcApiKeyCreateFormCmp.setData({ ngcApiKeyModel });
    expect(await ngcApiKeyCreateFormCmp.vm.validate()).toBe(false);

    ngcApiKeyModel.meta.name = "1";
    await ngcApiKeyCreateFormCmp.setData({ ngcApiKeyModel });
    expect(await ngcApiKeyCreateFormCmp.vm.validate()).toBe(false);

    ngcApiKeyModel.meta.name = "ngc-api-key-1";
    await ngcApiKeyCreateFormCmp.setData({ ngcApiKeyModel });
    expect(await ngcApiKeyCreateFormCmp.vm.validate()).toBe(true);
  });

  test("the form validation should fail when the NGC API key is invalid", async () => {
    const ngcApiKeyModel = validNgcApiKeyModel;

    ngcApiKeyModel.spec.ngc_api_key = "";
    await ngcApiKeyCreateFormCmp.setData({ ngcApiKeyModel });
    expect(await ngcApiKeyCreateFormCmp.vm.validate()).toBe(false);

    ngcApiKeyModel.spec.ngc_api_key = "validNgcApiKey";
    await ngcApiKeyCreateFormCmp.setData({ ngcApiKeyModel });
    expect(await ngcApiKeyCreateFormCmp.vm.validate()).toBe(true);
  });

  test("updates description when description is changed", async () => {
    const newDescription = "New description for NGC API key";
    ngcApiKeyCreateFormCmp.vm.updateDescription(newDescription);
    expect(ngcApiKeyCreateFormCmp.vm.ngcApiKeyModel.meta.description).toBe(newDescription);
  });

  test("emits cancel event when cancel button is clicked", async () => {
    ngcApiKeyCreateFormCmp.find("[aid=cancel-btn]").trigger("click");
    expect(ngcApiKeyCreateFormCmp.emitted("cancel")).toBeTruthy();
  });

  test("failing to submit the form when button is clicked and the form is invalid", async () => {
    // Setting scope so the submit button will appear
    await ngcApiKeyCreateFormCmp.setData({
      ngcApiKeyModel: {
        meta: { scope: "cluster" },
      },
    });

    ngcApiKeyCreateFormCmp.find("[aid=submit-btn]").trigger("click");
    await flushPromises();
    expect(ngcApiKeyCreateFormCmp.emitted("submit")).toBeFalsy();
  });

  test("successfully submitting the form when button is clicked and the form is valid", async () => {
    mockIsExist.mockResolvedValue(false);
    const validModel: NgcApiKeyCreationRequest = {
      meta: {
        name: "ngc-api-key-1",
        scope: Scope.Project,
        projectId: 25,
        clusterId: null,
      },
      spec: {
        ngc_api_key: "validNgcApiKey",
      },
    };
    await ngcApiKeyCreateFormCmp.setData({ ngcApiKeyModel: validModel });
    await flushPromises();

    ngcApiKeyCreateFormCmp.find("[aid=submit-btn]").trigger("click");
    await flushPromises();
    expect(ngcApiKeyCreateFormCmp.emitted().submit).toBeTruthy();
  });

  test("isExists is called when submitting the form", async () => {
    mockIsExist.mockResolvedValue(false);
    const validModel: NgcApiKeyCreationRequest = {
      meta: {
        name: "ngc-api-key-1",
        scope: Scope.Project,
        projectId: 1,
        clusterId: null,
      },
      spec: {
        ngc_api_key: "valid-ngc-api-key",
      },
    };
    await ngcApiKeyCreateFormCmp.setData({ ngcApiKeyModel: validModel });

    await flushPromises();

    ngcApiKeyCreateFormCmp.find("[aid=submit-btn]").trigger("click");

    await flushPromises();
    expect(mockIsExist).toHaveBeenLastCalledWith("ngcApiKey", validModel.meta.name, {
      scopeType: Scope.Project,
      scopeId: "1",
    });
  });

  test("shows error notification when save fails", async () => {
    const consoleErrorSpy = vi.spyOn(console, "error").mockImplementation(() => {}); // eslint-disable-line @typescript-eslint/no-empty-function
    const notifySpy = vi.fn();
    ngcApiKeyCreateFormCmp.vm.$q.notify = notifySpy;

    // Mock validate to throw an error
    vi.spyOn(ngcApiKeyCreateFormCmp.vm, "validate").mockRejectedValue(new Error("Validation failed"));

    await ngcApiKeyCreateFormCmp.setData({
      ngcApiKeyModel: {
        meta: { scope: "cluster" },
      },
    });

    ngcApiKeyCreateFormCmp.find("[aid=submit-btn]").trigger("click");
    await flushPromises();

    expect(notifySpy).toHaveBeenCalledWith(
      expect.objectContaining({
        message: "Failed to save NGC API key",
      }),
    );

    consoleErrorSpy.mockRestore();
  });

  test("computed properties work correctly", () => {
    // Test isScopeSelected
    ngcApiKeyCreateFormCmp.setData({
      ngcApiKeyModel: {
        meta: { scope: null },
      },
    });
    expect(ngcApiKeyCreateFormCmp.vm.isScopeSelected).toBe(false);

    ngcApiKeyCreateFormCmp.setData({
      ngcApiKeyModel: {
        meta: { scope: "project" },
      },
    });
    expect(ngcApiKeyCreateFormCmp.vm.isScopeSelected).toBe(true);

    // Test scopeModel
    const expectedScopeModel = {
      scope: "project",
      projectId: 25,
      departmentId: null,
      clusterId: null,
    };
    ngcApiKeyCreateFormCmp.setData({
      ngcApiKeyModel: {
        meta: expectedScopeModel,
      },
    });
    expect(ngcApiKeyCreateFormCmp.vm.scopeModel).toEqual(expectedScopeModel);

    // Test secretSummary
    ngcApiKeyCreateFormCmp.setData({
      ngcApiKeyModel: {
        spec: { ngc_api_key: "testKey" },
      },
    });
    expect(ngcApiKeyCreateFormCmp.vm.secretSummary).toBe("testKey");

    ngcApiKeyCreateFormCmp.setData({
      ngcApiKeyModel: {
        spec: { ngc_api_key: "" },
      },
    });
    expect(ngcApiKeyCreateFormCmp.vm.secretSummary).toBe("None");
  });

  test("shows form hint when validation fails", async () => {
    await ngcApiKeyCreateFormCmp.setData({
      ngcApiKeyModel: {
        meta: { scope: "cluster" },
      },
    });

    // Mock validate to return false
    vi.spyOn(ngcApiKeyCreateFormCmp.vm, "validate").mockResolvedValue(false);

    ngcApiKeyCreateFormCmp.find("[aid=submit-btn]").trigger("click");
    await flushPromises();

    expect(ngcApiKeyCreateFormCmp.vm.displayFormHint).toBe(true);
  });

  test("clears timeout on unmount", () => {
    const clearTimeoutSpy = vi.spyOn(window, "clearTimeout");
    ngcApiKeyCreateFormCmp.vm.timeoutId = 123;

    ngcApiKeyCreateFormCmp.unmount();

    expect(clearTimeoutSpy).toHaveBeenCalledWith(123);
  });

  test("secret section is only visible when scope is selected", async () => {
    // Initially no scope selected
    await ngcApiKeyCreateFormCmp.setData({
      ngcApiKeyModel: {
        meta: { scope: null },
      },
    });
    expect(ngcApiKeyCreateFormCmp.find(".ngc-api-key-secret-section").exists()).toBe(false);

    // Scope selected
    await ngcApiKeyCreateFormCmp.setData({
      ngcApiKeyModel: {
        meta: { scope: "project" },
      },
    });
    expect(ngcApiKeyCreateFormCmp.find(".ngc-api-key-secret-section").exists()).toBe(true);
  });
});
