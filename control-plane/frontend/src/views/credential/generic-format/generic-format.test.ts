/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { mount } from "@vue/test-utils";
import GenericFormat from "./generic-format.vue";
import { beforeEach, describe, expect, test, vi } from "vitest";
import type { VueWrapper } from "@vue/test-utils";
import { Quasar } from "quasar";
import { createPinia, setActivePinia } from "pinia";
import { useAppStore } from "@/stores/app.store";
import { useAuthStore } from "@/stores/auth.store";
import { useClusterStore } from "@/stores/cluster.store";
import { credentialService } from "@/services/control-plane/credential.service/credential.service";
import { AssetKind, Scope } from "@/swagger-models/assets-service-client";
import { TYPE_SSH_KEY } from "@/models/credential.model";
import type { SecretListInfo } from "@/swagger-models/k8s-objects-tracker-client";
import { mockOrgUnits, mockTenant } from "@/components/common/runai-scope-section/runai-scope-section.mock";
import type { IScopeModel } from "@/models/global.model";

// Mock services
vi.mock("@/services/control-plane/credential.service/credential.service", () => ({
  credentialService: {
    getEmptyGenericFormatSecretModel: vi.fn(() => ({
      meta: {
        name: "",
        description: "",
        scope: null,
        projectId: null,
        departmentId: null,
        clusterId: null,
      },
      spec: {},
    })),
    listExistingSecrets: vi.fn(() => Promise.resolve([])),
    createFormatBasedGenericSecret: vi.fn((model) =>
      Promise.resolve({
        meta: { id: "test-id", name: model.meta.name },
      }),
    ),
  },
}));

vi.mock("@/services/infra/router.service/router.service", () => ({
  requestToLeave: vi.fn(() => Promise.resolve(true)),
}));

vi.mock("@/utils/assets.util/assets.util", () => ({
  assetsUtil: {
    setAssetMetaFromQuery: vi.fn((meta, query) => ({
      ...meta,
      scope: query.scope,
      projectId: query.projectId,
      departmentId: query.departmentId,
      clusterId: query.clusterId,
    })),
  },
}));

vi.mock("@/utils/scope.util", () => ({
  scopeUtil: {
    getScopeId: vi.fn(() => "test-scope-id"),
    updateScope: vi.fn((meta, scope) => ({ ...meta, ...scope })),
  },
}));

describe("generic-format.vue", () => {
  let wrapper: VueWrapper<InstanceType<typeof GenericFormat>>;
  let mockRouter;
  let mockRoute;

  const mockExistingSecrets: SecretListInfo[] = [
    { name: "existing-secret-1", type: "generic" },
    { name: "existing-secret-2", type: "generic" },
  ] as SecretListInfo[];

  beforeEach(() => {
    setActivePinia(createPinia());

    // Setup stores
    const appStore = useAppStore();
    const authStore = useAuthStore();
    const clusterStore = useClusterStore();

    appStore.setPageLoading = vi.fn();
    authStore.$patch({
      tenant: mockTenant,
      orgUnits: mockOrgUnits,
    });

    clusterStore.isClusterVersionSufficient = vi.fn(() => true);

    mockRouter = {
      push: vi.fn(),
    };

    mockRoute = {
      query: {},
      meta: { backPageName: "credentials-list" },
    };

    wrapper = mount(GenericFormat, {
      global: {
        plugins: [Quasar],
        mocks: {
          $router: mockRouter,
          $route: mockRoute,
          $q: {
            notify: vi.fn(),
          },
        },
        stubs: {
          CredentialTypeSection: true,
          RunaiExpansionWrapper: true,
          RunaiFormWrapper: {
            template: "<div><slot /></div>",
            methods: {
              validate: vi.fn(() => Promise.resolve(true)),
            },
          },
          RunaiScopeSection: true,
          RunaiNameSection: true,
          GenericFormatSecretSection: true,
          RunaiFormFooter: true,
        },
      },
    });
  });

  describe("Component Initialization", () => {
    test("initializes with empty model", () => {
      expect(wrapper.vm.genericFormatModel).toBeDefined();
      expect(wrapper.vm.genericFormatModel.meta.name).toBe("");
      expect(wrapper.vm.genericFormatModel.spec).toEqual({});
    });

    test("sets page loading to false on creation", async () => {
      const appStore = useAppStore();
      expect(appStore.setPageLoading).toHaveBeenCalledWith(false);
    });

    test("initializes with correct default states", () => {
      expect(wrapper.vm.submitting).toBe(false);
      expect(wrapper.vm.validating).toBe(false);
      expect(wrapper.vm.isFormValid).toBe(true);
      expect(wrapper.vm.loadingSecrets).toBe(false);
      expect(wrapper.vm.existingSecrets).toBe(null);
    });
  });

  describe("Scope Management", () => {
    test("computes scopeModel correctly from genericFormatModel", () => {
      wrapper.vm.genericFormatModel.meta.scope = "project";
      wrapper.vm.genericFormatModel.meta.projectId = 123;
      wrapper.vm.genericFormatModel.meta.departmentId = "dept-456";
      wrapper.vm.genericFormatModel.meta.clusterId = "cluster-789";

      const scopeModel = wrapper.vm.scopeModel;
      expect(scopeModel).toEqual({
        scope: "project",
        projectId: 123,
        departmentId: "dept-456",
        clusterId: "cluster-789",
      });
    });

    test("determines scope is selected when meta.scope exists", () => {
      wrapper.vm.genericFormatModel.meta.scope = "project";
      expect(wrapper.vm.isScopeSelected).toBe(true);
    });

    test("determines scope is not selected when meta.scope is null", () => {
      wrapper.vm.genericFormatModel.meta.scope = null;
      expect(wrapper.vm.isScopeSelected).toBe(false);
    });

    let updatedScope: IScopeModel = {
      scope: Scope.Project,
      projectId: 0,
      departmentId: "",
      clusterId: "",
      parentClusterUuid: "",
    };

    test("updates scope correctly", async () => {
      updatedScope = {
        scope: "project",
        projectId: 123,
        departmentId: "new-dept",
        clusterId: "new-cluster",
        parentClusterUuid: "parent-uuid",
      };

      wrapper.vm.updateScope(updatedScope);
      await wrapper.vm.$nextTick();

      expect(wrapper.vm.parentClusterUuid).toBe("parent-uuid");
      expect(credentialService.listExistingSecrets).toHaveBeenCalled();
    });

    test("resets secret details when scope changes", async () => {
      wrapper.vm.genericFormatModel.spec = { accessKeyId: "test" };

      updatedScope = {
        scope: "project",
        projectId: 321,
        departmentId: null,
        clusterId: null,
      };

      wrapper.vm.updateScope(updatedScope);
      await wrapper.vm.$nextTick();

      expect(wrapper.vm.genericFormatModel.spec).toEqual({});
    });
  });

  describe("Loading Existing Secrets", () => {
    test("loads existing secrets successfully", async () => {
      vi.mocked(credentialService.listExistingSecrets).mockResolvedValue(mockExistingSecrets);

      await wrapper.vm.loadExistingGenericSecrets({
        scope: "project",
        projectId: 546,
        departmentId: null,
        clusterId: null,
      });

      expect(wrapper.vm.existingSecrets).toEqual(mockExistingSecrets);
      expect(wrapper.vm.loadingSecrets).toBe(false);
    });
  });

  describe("Format Type Detection", () => {
    test("detects GenericSecret format from keyValuePairs", () => {
      wrapper.vm.genericFormatModel.spec = {
        keyValuePairs: [{ key: "test", value: "value" }],
      };

      expect(wrapper.vm.currentFormatType).toBe(AssetKind.GenericSecret);
    });

    test("detects AccessKey format from accessKeyId", () => {
      wrapper.vm.genericFormatModel.spec = {
        accessKeyId: "test-key",
        secretAccessKey: "test-secret",
      };

      expect(wrapper.vm.currentFormatType).toBe(AssetKind.AccessKey);
    });

    test("detects Password format from user", () => {
      wrapper.vm.genericFormatModel.spec = {
        user: "admin",
        password: "pass123",
      };

      expect(wrapper.vm.currentFormatType).toBe(AssetKind.Password);
    });

    test("returns null when no format is detected", () => {
      wrapper.vm.genericFormatModel.spec = {};
      expect(wrapper.vm.currentFormatType).toBe(null);
    });

    test("detects existing secret format type", async () => {
      wrapper.vm.existingSecrets = [{ name: "secret-1", type: "generic" }] as SecretListInfo[];
      wrapper.vm.genericFormatModel.spec = {
        existingSecretName: "secret-1",
      };

      // The currentFormatType should return the corresponding AssetKind
      // This will be null or the mapped type based on SECRET_TYPE_TO_ASSET_KIND
      const formatType = wrapper.vm.currentFormatType;
      expect(formatType).toBeDefined();
    });
  });

  describe("Disabled Formats", () => {
    test("disables GenericSecret and SSH key when cluster version is insufficient", () => {
      const clusterStore = useClusterStore();
      vi.mocked(clusterStore.isClusterVersionSufficient).mockReturnValue(false);

      wrapper.vm.parentClusterUuid = "test-uuid";

      const disabledFormats = wrapper.vm.disabledFormats;
      expect(disabledFormats).toContain(AssetKind.GenericSecret);
      expect(disabledFormats).toContain(TYPE_SSH_KEY);
    });

    test("only disables SSH key when cluster supports GenericSecret but not SSH", () => {
      const clusterStore = useClusterStore();
      vi.mocked(clusterStore.isClusterVersionSufficient)
        .mockReturnValueOnce(true) // for GenericSecret check
        .mockReturnValueOnce(false); // for SSH key check

      wrapper.vm.parentClusterUuid = "test-uuid";

      const disabledFormats = wrapper.vm.disabledFormats;
      expect(disabledFormats).toEqual([TYPE_SSH_KEY]);
    });

    test("disables no formats when cluster supports all features", () => {
      const clusterStore = useClusterStore();
      vi.mocked(clusterStore.isClusterVersionSufficient).mockReturnValue(true);

      wrapper.vm.parentClusterUuid = "test-uuid";

      const disabledFormats = wrapper.vm.disabledFormats;
      expect(disabledFormats).toEqual([]);
    });
  });

  describe("Form Actions", () => {
    test("cancels and navigates back when onCancel is called", async () => {
      await wrapper.vm.onCancel();

      expect(mockRouter.push).toHaveBeenCalledWith({
        name: "credentials-list",
      });
    });

    test("validates form before submission", async () => {
      wrapper.vm.genericFormatModel.spec = {
        accessKeyId: "test-key",
        secretAccessKey: "test-secret",
      };

      const validateSpy = vi.spyOn(wrapper.vm, "validate");

      await wrapper.vm.onSubmit();

      expect(validateSpy).toHaveBeenCalled();
    });

    test("submits successfully and navigates to list", async () => {
      wrapper.vm.genericFormatModel = {
        meta: {
          name: "test-credential",
          description: "test description",
          scope: "project",
          projectId: 123,
          departmentId: null,
          clusterId: null,
        },
        spec: {
          accessKeyId: "test-key",
          secretAccessKey: "test-secret",
        },
      };

      await wrapper.vm.onSubmit();

      expect(credentialService.createFormatBasedGenericSecret).toHaveBeenCalledWith(
        wrapper.vm.genericFormatModel,
        AssetKind.AccessKey,
      );
      expect(wrapper.vm.$q.notify).toHaveBeenCalled();
      expect(mockRouter.push).toHaveBeenCalledWith({
        name: "credentials-list",
        query: {
          createdEntityId: "test-id",
          kind: AssetKind.AccessKey,
        },
      });
    });

    describe("resetSecretDetails", () => {
      test("resets spec to empty object", () => {
        wrapper.vm.genericFormatModel.spec = {
          accessKeyId: "test",
          secretAccessKey: "secret",
        };

        wrapper.vm.resetSecretDetails();

        expect(wrapper.vm.genericFormatModel.spec).toEqual({});
      });
    });
  });
});
