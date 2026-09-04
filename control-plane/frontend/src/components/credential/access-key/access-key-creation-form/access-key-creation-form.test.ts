/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { VueWrapper } from "@vue/test-utils";
import { flushPromises, mount } from "@vue/test-utils";
import AccessKeyCreateForm from "./access-key-creation-form.vue";
import { afterEach, beforeEach, describe, expect, test, vi, type MockedFunction } from "vitest";
import type { DefineComponent } from "vue";
import { setActivePinia } from "pinia";
import { createTestingPinia } from "@pinia/testing";
import { Quasar } from "quasar";
import { type AccessKeyCreationRequest, Scope } from "@/swagger-models/assets-service-client";
import { assetService } from "@/services/control-plane/asset.service/asset.service";
import { credentialService } from "@/services/control-plane/credential.service/credential.service";
import { orgTreeService } from "@/services/control-plane/rbac/org-tree.service/org-tree.service";
import { orgTreeServiceMock } from "@/services/control-plane/rbac/org-tree.service/org-tree.service.mock";
import type { IScopeModel } from "@/models/global.model";
import { useClusterStore } from "@/stores/cluster.store";
import type { IUIAccessKeyCreationRequest } from "@/models/credential.model";

vi.mock("@/services/control-plane/asset.service/asset.service");

const pinia = createTestingPinia({ createSpy: vi.fn });

setActivePinia(pinia);

const validAccessKeyModel: IUIAccessKeyCreationRequest = {
  meta: {
    name: "ak-1",
    scope: Scope.Project,
    projectId: 25,
  },
  spec: {
    existingSecretName: null,
    accessKeyId: "placeholder",
    secretAccessKey: "placeholder",
  },
};

describe("access-key-creation-form.vue", () => {
  let accessKeyCreateFormCmp: VueWrapper<InstanceType<typeof AccessKeyCreateForm>>;

  const mockIsExist = assetService.isExist as MockedFunction<typeof assetService.isExist>;

  const emptyAccessKeyModel = credentialService.getEmptyAccessKeyModel();

  const spyGetOrgTree = vi.spyOn(orgTreeService, "getOrgTree");
  const spyListExistingSecrets = vi.spyOn(credentialService, "listExistingSecrets");

  let clusterStore = useClusterStore(pinia);

  beforeEach(() => {
    spyGetOrgTree.mockReturnValue(orgTreeServiceMock.orgTree);
    spyListExistingSecrets.mockResolvedValue([]);

    clusterStore = useClusterStore(pinia);
    clusterStore.$patch({ currentCluster: { uuid: "34234-322-sdfs323" } });
    clusterStore.isClusterVersionSufficient = vi.fn().mockReturnValue(true);
    clusterStore.isAllClustersSupportSpecificVersion = vi.fn().mockReturnValue(true);

    accessKeyCreateFormCmp = mount(AccessKeyCreateForm, {
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
        isProjectSelectable: true,
        credentialModel: emptyAccessKeyModel,
        existingSecretNames: ["secret1", "secret2", "secret3"],
        loadingSecrets: false,
      },
    });
    accessKeyCreateFormCmp.vm.$q.notify = vi.fn(); // disabling notify to prevent it from failing tests
  });
  afterEach(() => {
    vi.clearAllMocks();
    vi.resetAllMocks();
  });

  test("component is loading properly", () => {
    expect(accessKeyCreateFormCmp.html()).toContain("access-key-create-form");
  });

  test("the form validation should fail when the name is invalid", async () => {
    const accessKeyModel = validAccessKeyModel;
    // Need to click the radio button in order to make the form valid

    accessKeyModel.meta.name = "";
    await accessKeyCreateFormCmp.setData({ accessKeyModel });
    await accessKeyCreateFormCmp.find('[aid="radio-option-New secret"]').trigger("click");
    expect(await accessKeyCreateFormCmp.vm.validate()).toBe(false);

    accessKeyModel.meta.name = "1";
    await accessKeyCreateFormCmp.setData({ accessKeyModel });
    expect(await accessKeyCreateFormCmp.vm.validate()).toBe(false);

    accessKeyModel.meta.name = "ak-1";
    await accessKeyCreateFormCmp.setData({ accessKeyModel });
    expect(await accessKeyCreateFormCmp.vm.validate()).toBe(true);
  });

  test("the form validation should fail when the secret section is invalid", async () => {
    const accessKeyModel = validAccessKeyModel;

    // New secret
    accessKeyModel.spec.secretAccessKey = "";
    await accessKeyCreateFormCmp.setData({ accessKeyModel });
    await accessKeyCreateFormCmp.find('[aid="radio-option-New secret"]').trigger("click");
    expect(await accessKeyCreateFormCmp.vm.validate()).toBe(false);

    accessKeyModel.spec.secretAccessKey = "1";
    await accessKeyCreateFormCmp.setData({ accessKeyModel });
    expect(await accessKeyCreateFormCmp.vm.validate()).toBe(true);

    accessKeyModel.spec.accessKeyId = "";
    await accessKeyCreateFormCmp.setData({ accessKeyModel });
    expect(await accessKeyCreateFormCmp.vm.validate()).toBe(false);

    accessKeyModel.spec.accessKeyId = "1";
    await accessKeyCreateFormCmp.setData({ accessKeyModel });
    expect(await accessKeyCreateFormCmp.vm.validate()).toBe(true);

    // Existing secret
    await accessKeyCreateFormCmp.find('[aid="radio-option-Existing secret"]').trigger("click");
    accessKeyModel.spec.existingSecretName = "";
    await accessKeyCreateFormCmp.setData({ accessKeyModel });
    expect(await accessKeyCreateFormCmp.vm.validate()).toBe(false);

    accessKeyModel.spec.existingSecretName = "secret1";
    await accessKeyCreateFormCmp.setData({ accessKeyModel });
    expect(await accessKeyCreateFormCmp.vm.validate()).toBe(true);
  });

  test("isExists is called when submitting the form", async () => {
    mockIsExist.mockResolvedValue(false);
    const validModel: AccessKeyCreationRequest = {
      meta: {
        name: "credential1",
        scope: Scope.Project,
        projectId: 1,
        clusterId: null,
      },
      spec: {
        existingSecretName: null,
        accessKeyId: "ac-id",
        secretAccessKey: "secret-ac",
      },
    };
    await accessKeyCreateFormCmp.setData({ accessKeyModel: validModel });

    const secretSectionCmp = accessKeyCreateFormCmp.findComponent<DefineComponent>(".access-key-secret-section");
    await secretSectionCmp.setData({ isExistingSecret: false }); // Revealing the create new section in the secret section component

    await flushPromises();

    accessKeyCreateFormCmp.find("[aid=submit-btn]").trigger("click");

    await flushPromises();

    expect(mockIsExist).toHaveBeenLastCalledWith("accessKey", validModel.meta.name, {
      scopeType: Scope.Project,
      scopeId: "1",
    });
  });

  test("updates the scope and resets the spec in the model when a new scope is emitted by the scope-section", async () => {
    accessKeyCreateFormCmp.setData({
      accessKeyModel: {
        meta: {
          name: "",
          scope: null,
          projectId: null,
          clusterId: null,
        },
        spec: {
          existingSecretName: null,
          accessKeyId: "placeholder",
          secretAccessKey: "placeholder",
        },
      },
    });
    await accessKeyCreateFormCmp.vm.$nextTick();

    const newScopeModel: IScopeModel = {
      scope: Scope.Project,
      projectId: 1,
      departmentId: null,
    };

    const scopeSectionCmp = accessKeyCreateFormCmp.findComponent<DefineComponent>("[aid=runai-scope-section]");
    scopeSectionCmp.vm.$emit("changed-scope", newScopeModel);
    expect(accessKeyCreateFormCmp.vm.accessKeyModel.meta.projectId).toBe(newScopeModel.projectId);
    expect(accessKeyCreateFormCmp.vm.accessKeyModel.spec.accessKeyId).toBe("");
    expect(accessKeyCreateFormCmp.vm.accessKeyModel.spec.secretAccessKey).toBe("");
    expect(accessKeyCreateFormCmp.vm.accessKeyModel.spec.existingSecretName).toBe(null);
  });
});
