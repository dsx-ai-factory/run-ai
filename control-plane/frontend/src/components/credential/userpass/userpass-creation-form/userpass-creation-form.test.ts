/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { VueWrapper } from "@vue/test-utils";
import { flushPromises, mount } from "@vue/test-utils";
import UserpassCreateForm from "./userpass-creation-form.vue";
import { afterEach, beforeEach, describe, expect, test, vi, type MockedFunction } from "vitest";
import type { DefineComponent } from "vue";
import { setActivePinia } from "pinia";
import { createTestingPinia } from "@pinia/testing";
import { Quasar } from "quasar";
import { type PasswordCreationRequest, Scope } from "@/swagger-models/assets-service-client";
import { assetService } from "@/services/control-plane/asset.service/asset.service";
import { credentialService } from "@/services/control-plane/credential.service/credential.service";
import { orgTreeService } from "@/services/control-plane/rbac/org-tree.service/org-tree.service";
import { orgTreeServiceMock } from "@/services/control-plane/rbac/org-tree.service/org-tree.service.mock";
import type { IScopeModel } from "@/models/global.model";
import { useClusterStore } from "@/stores/cluster.store";
import type { IUIPasswordCreationRequest } from "@/models/credential.model";

vi.mock("@/services/control-plane/asset.service/asset.service");

const pinia = createTestingPinia({ createSpy: vi.fn });

setActivePinia(pinia);

const validUserPassModel: IUIPasswordCreationRequest = {
  meta: {
    name: "credential1",
    scope: Scope.Project,
    projectId: 25,
    clusterId: null,
  },
  spec: {
    existingSecretName: null,
    password: "password",
    user: "user",
  },
};

describe("userpass-creation-form.vue", () => {
  let UserpassCreateFormCmp: VueWrapper<InstanceType<typeof UserpassCreateForm>>;

  const emptyUserpassModel = credentialService.getEmptyUserpassModel();

  const mockIsExist = assetService.isExist as MockedFunction<typeof assetService.isExist>;
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

    UserpassCreateFormCmp = mount(UserpassCreateForm, {
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
        credentialModel: emptyUserpassModel,
        existingSecretNames: ["secret1", "secret2", "secret3"],
        loadingSecrets: false,
      },
    });
    UserpassCreateFormCmp.vm.$q.notify = vi.fn(); // disabling notify to prevent it from failing tests
  });
  afterEach(() => {
    vi.clearAllMocks();
    vi.resetAllMocks();
  });

  test("component is loading properly", () => {
    expect(UserpassCreateFormCmp.html()).toContain("userpass-create-form");
  });

  test("form validation should fail if name is invalid", async () => {
    validUserPassModel.meta.name = "";
    await UserpassCreateFormCmp.setData({ userpassModel: validUserPassModel });
    await UserpassCreateFormCmp.find('[aid="radio-option-New secret"]').trigger("click");
    expect(await UserpassCreateFormCmp.vm.validate()).toBe(false);

    validUserPassModel.meta.name = "1-3";
    await UserpassCreateFormCmp.setData({ userpassModel: validUserPassModel });
    expect(await UserpassCreateFormCmp.vm.validate()).toBe(false);

    validUserPassModel.meta.name = "valid-name";
    await UserpassCreateFormCmp.setData({ userpassModel: validUserPassModel });
    expect(await UserpassCreateFormCmp.vm.validate()).toBe(true);
  });

  test("form validation should fail if secret section is invalid", async () => {
    validUserPassModel.spec.password = "";
    // New secret
    await UserpassCreateFormCmp.setData({ userpassModel: validUserPassModel });
    await UserpassCreateFormCmp.find('[aid="radio-option-New secret"]').trigger("click");
    expect(await UserpassCreateFormCmp.vm.validate()).toBe(false);

    validUserPassModel.spec.password = "validpassword";
    await UserpassCreateFormCmp.setData({ userpassModel: validUserPassModel });
    expect(await UserpassCreateFormCmp.vm.validate()).toBe(true);

    validUserPassModel.spec.user = "";
    await UserpassCreateFormCmp.setData({ userpassModel: validUserPassModel });
    expect(await UserpassCreateFormCmp.vm.validate()).toBe(false);

    validUserPassModel.spec.user = "validuser";
    await UserpassCreateFormCmp.setData({ userpassModel: validUserPassModel });
    expect(await UserpassCreateFormCmp.vm.validate()).toBe(true);

    // Existing secret
    await UserpassCreateFormCmp.find('[aid="radio-option-Existing secret"]').trigger("click");
    await UserpassCreateFormCmp.setData({ userpassModel: validUserPassModel });
    expect(await UserpassCreateFormCmp.vm.validate()).toBe(false);

    validUserPassModel.spec.existingSecretName = "validsecret";
    await UserpassCreateFormCmp.setData({ userpassModel: validUserPassModel });
    expect(await UserpassCreateFormCmp.vm.validate()).toBe(true);
  });

  test("updates the scope and resets the spec in the model when a new scope is emitted by the scope-section", async () => {
    UserpassCreateFormCmp.setData({
      userpassModel: {
        meta: {
          name: "",
          scope: null,
          projectId: null,
          clusterId: null,
        },
        spec: {
          existingSecretName: null,
          password: "placeholder",
          user: "placeholder",
        },
      },
    });
    await UserpassCreateFormCmp.vm.$nextTick();

    const newScopeModel: IScopeModel = {
      scope: Scope.Project,
      projectId: 1,
      departmentId: null,
    };

    const scopeSectionCmp = UserpassCreateFormCmp.findComponent<DefineComponent>("[aid=runai-scope-section]");
    scopeSectionCmp.vm.$emit("changed-scope", newScopeModel);
    await UserpassCreateFormCmp.vm.$nextTick();
    expect(UserpassCreateFormCmp.vm.userpassModel.meta.projectId).toBe(newScopeModel.projectId);
    expect(UserpassCreateFormCmp.vm.userpassModel.spec.user).toBe("");
    expect(UserpassCreateFormCmp.vm.userpassModel.spec.password).toBe("");
    expect(UserpassCreateFormCmp.vm.userpassModel.spec.existingSecretName).toBe(null);
  });

  test("failing to submit the form when button is clicked and the form is invalid", async () => {
    // Setting scope so the submit button will appear
    await UserpassCreateFormCmp.setData({
      userpassModel: {
        meta: { scope: "cluster" },
      },
    });

    UserpassCreateFormCmp.find("[aid=submit-btn]").trigger("click");
    await flushPromises();
    expect(UserpassCreateFormCmp.emitted("submit")).toBeFalsy();
  });

  test("successfuly submitting the form when button is clicked and the form is valid", async () => {
    mockIsExist.mockResolvedValue(false);
    const validModel: PasswordCreationRequest = {
      meta: {
        name: "credential1",
        scope: Scope.Project,
        projectId: 25,
        clusterId: null,
      },
      spec: {
        existingSecretName: null,
        password: "password",
        user: "user",
      },
    };
    await UserpassCreateFormCmp.setData({ userpassModel: validModel });
    await flushPromises();

    const secretSectionCmp = UserpassCreateFormCmp.findComponent<DefineComponent>(".userpass-secret-section");
    UserpassCreateFormCmp.vm.userpassModel.spec = validModel.spec; // When updating the project it resets the spec, therefore I update the spec directly to avoid it
    await secretSectionCmp.setData({ isExistingSecret: false }); // Revealing the create new section in the secret section component
    UserpassCreateFormCmp.find("[aid=submit-btn]").trigger("click");
    await flushPromises();
    expect(UserpassCreateFormCmp.emitted().submit).toBeTruthy();
  });

  test("isExists is called when submitting the form", async () => {
    mockIsExist.mockResolvedValue(false);
    const validModel: PasswordCreationRequest = {
      meta: {
        name: "credential1",
        scope: Scope.Project,
        projectId: 1,
        clusterId: null,
      },
      spec: {
        existingSecretName: null,
        password: "password",
        user: "user",
      },
    };
    await UserpassCreateFormCmp.setData({ userpassModel: validModel });
    const secretSectionCmp = UserpassCreateFormCmp.findComponent<DefineComponent>(".userpass-secret-section");
    await secretSectionCmp.setData({ isExistingSecret: false }); // Revealing the create new section in the secret section component

    await flushPromises();

    UserpassCreateFormCmp.find("[aid=submit-btn]").trigger("click");

    await flushPromises();
    expect(mockIsExist).toHaveBeenLastCalledWith("password", validModel.meta.name, {
      scopeType: Scope.Project,
      scopeId: "1",
    });
  });
});
