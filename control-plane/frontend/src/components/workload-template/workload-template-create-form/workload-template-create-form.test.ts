/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { workspaceAssetsEditMockedData } from "@/views/workspace/creation/workspace-assets-edit.mock";
import { createTestingPinia } from "@pinia/testing";
import type { VueWrapper } from "@vue/test-utils";
import { flushPromises, mount } from "@vue/test-utils";
import { setActivePinia } from "pinia";
import { Quasar } from "quasar";
import { beforeEach, describe, expect, test, vi, type MockedFunction } from "vitest";
import type { DefineComponent } from "vue";
import { WorkloadTemplateCreateForm } from "./";
import { AssetKind, Scope, type AssetIdAndKind } from "@/swagger-models/assets-service-client";
import { useDataSourceStore } from "@/stores/data-source.store";
import { useAppStore } from "@/stores/app.store";
import { useAuthStore } from "@/stores/auth.store";

import { assetService } from "@/services/control-plane/asset.service/asset.service";
import { WORKLOAD_TEMPLATE_ROUTE_NAMES } from "@/router/workload-template.routes/workload-template.routes.names";
import { useEnvironmentStore } from "@/stores/environment.store";
import { dataSourceService } from "@/services/control-plane/data-source.service/data-source.service";
import { orgTreeService } from "@/services/control-plane/rbac/org-tree.service/org-tree.service";
import { credentialService } from "@/services/control-plane/credential.service/credential.service";
import { useClusterStore } from "@/stores/cluster.store";
import type { IUIWorkloadTemplateCreation } from "@/models/workload-template.model";
import { orgTreeServiceMock } from "@/services/control-plane/rbac/org-tree.service/org-tree.service.mock";
import { DATA_SOURCE_ROUTE_NAMES } from "@/router/data-source.routes/data-source.routes.names";
import { COMPUTE_RESOURCE_ROUTE_NAMES } from "@/router/compute-resource.routes/compute-resource.routes.names";
import { ENVIRONMENT_ROUTE_NAMES } from "@/router/environment.routes/environment.routes.names";
import { setQuasarMock } from "@/utils/unit-test.util/unit-test.util";
import { FAKE_ACCESS_TOKEN_B } from "@/tests/fixtures/auth.fixtures";

vi.mock("@/services/control-plane/asset.service/asset.service");
vi.mock("@/services/control-plane/data-source.service/data-source.service");
vi.mock("@/services/control-plane/credential.service/credential.service");

describe("workload-template-create-form.vue", () => {
  let workloadTemplateCreateForm: VueWrapper<InstanceType<typeof WorkloadTemplateCreateForm>>;

  const mockIsExist = assetService.isExist as MockedFunction<typeof assetService.isExist>;
  mockIsExist.mockResolvedValue(false);

  const mockListStorageClass = dataSourceService.listStorageClass as MockedFunction<
    typeof dataSourceService.listStorageClass
  >;
  mockListStorageClass.mockImplementation(() => {
    return Promise.resolve([]);
  });

  const mockListCredentials = credentialService.listCredentials as MockedFunction<
    typeof credentialService.listCredentials
  >;
  mockListCredentials.mockResolvedValue([]);

  const mockListConfigMaps = dataSourceService.listExistingConfigMaps as MockedFunction<
    typeof dataSourceService.listExistingConfigMaps
  >;
  mockListConfigMaps.mockResolvedValue([]);

  const token = FAKE_ACCESS_TOKEN_B;

  const pinia = createTestingPinia({ createSpy: vi.fn });
  setActivePinia(pinia);

  let dataSourceStore = useDataSourceStore(pinia);
  let authStore = useAuthStore(pinia);
  let appStore = useAppStore(pinia);
  let environmentStore = useEnvironmentStore(pinia);
  let clusterStore = useClusterStore(pinia);

  const spyGetOrgTree = vi.spyOn(orgTreeService, "getOrgTree");

  beforeEach(() => {
    spyGetOrgTree.mockReturnValue(orgTreeServiceMock.orgTree);

    clusterStore = useClusterStore(pinia);
    dataSourceStore = useDataSourceStore(pinia);
    environmentStore = useEnvironmentStore(pinia);
    appStore = useAppStore(pinia);
    authStore = useAuthStore(pinia);

    authStore.$patch({ accessToken: token });
    clusterStore.$patch({ currentCluster: { uuid: "34234-322-sdfs323" } });
    environmentStore.$patch({ environments: workspaceAssetsEditMockedData.loadedEnvironments });
    dataSourceStore.$patch({ dataSources: workspaceAssetsEditMockedData.loadedDataSources });
    appStore.$patch({ pageLoading: false });

    clusterStore.isClusterVersionSufficient = vi.fn().mockReturnValue(true);
    clusterStore.isAllClustersSupportSpecificVersion = vi.fn().mockReturnValue(true);

    workloadTemplateCreateForm = mount(WorkloadTemplateCreateForm, {
      global: {
        plugins: [Quasar],
        directives: {
          clickOutside: vi.fn(),
        },
        mocks: {
          $route: {
            query: "",
          },
          $router: [],
        },
      },
      props: {
        initialWorkloadTemplate: {
          meta: {
            name: "",
            scope: Scope.Project,
            clusterId: "",
            projectId: 1,
          },
          spec: {
            assets: {
              environment: "",
              compute: null,
              datasources: [],
            },
            specificEnv: {
              allowOverQuota: false,
              args: "",
              command: "",
              environmentVariables: [],
              connections: [],
              runAsGid: null,
              runAsUid: null,
            },
          },
        },
        isScopeSelectable: false,
        submitting: false,
      },
      provide: {
        policy: {},
      },
      data() {
        return {
          workloadTemplateForm: null as HTMLFormElement | null,
        };
      },
    });

    setQuasarMock(workloadTemplateCreateForm);
  });

  test("component is loading properly", () => {
    expect(workloadTemplateCreateForm.html()).toContain("workload-template-create-form");
  });

  describe("loading data", () => {
    test("component will load credentials upon creation if the template model contains a scope", () => {
      expect(mockListCredentials).toHaveBeenCalled();
    });
  });

  describe("emits from child components", () => {
    test("updates the template name when a new name is emitted by workload-template-name-section component", async () => {
      const newName = "new-template-name";
      const nameSectionCmp = workloadTemplateCreateForm.findComponent<DefineComponent>("[aid=name-section]");
      nameSectionCmp.vm.$emit("update:modelValue", newName);
      await workloadTemplateCreateForm.vm.$nextTick();
      expect(workloadTemplateCreateForm.vm.workloadTemplateModel.name).toBe(newName);
    });

    test("updates the template environment properties correctly when a new environment data is emitted by the environment section", async () => {
      const selectedEnv = workspaceAssetsEditMockedData.loadedEnvironments[0];
      const emittedEnvironmentData = {
        environmentId: selectedEnv.meta.id as string,
        specificEnv: {
          command: selectedEnv.spec.command || "",
          args: selectedEnv.spec.args || "",
          environmentVariables: selectedEnv.spec.environmentVariables || [],
          allowOverQuota: false,
          connections: selectedEnv.spec.connections || [],
          runAsGid: selectedEnv.spec.runAsGid || null,
          runAsUid: selectedEnv.spec.runAsUid || null,
          supplementalGroups: selectedEnv.spec.supplementalGroups || null,
          nodePools: null,
          nodeType: null,
          autoDeletionTimeAfterCompletionSeconds: null,
          backoffLimit: 6,
        },
      };
      const environmentSectionCmp =
        workloadTemplateCreateForm.findComponent<DefineComponent>("[aid=environment-section]");
      environmentSectionCmp.vm.$emit("environment-changed", emittedEnvironmentData);
      expect(workloadTemplateCreateForm.vm.workloadTemplateModel.assets.environment).toBe(
        emittedEnvironmentData.environmentId,
      );
      expect(workloadTemplateCreateForm.vm.workloadTemplateModel.specificEnv).toEqual(
        emittedEnvironmentData.specificEnv,
      );
    });

    test("updates the template compute asset correctly when a new computeId is emitted by the compute resource section", async () => {
      const computeSectionCmp = workloadTemplateCreateForm.findComponent<DefineComponent>(
        "[aid=compute-resource-section]",
      );
      const computeId: string = workspaceAssetsEditMockedData.loadedComputeResources[0].meta.id;
      const event = { computeResourceId: computeId, nodeType: null };
      computeSectionCmp.vm.$emit("compute-resource-data-changed", event);
      expect(workloadTemplateCreateForm.vm.workloadTemplateModel.assets.compute).toBe(computeId);
    });

    test("updates the template data sources correctly when a new array of datasources is emitted by the datasources section", async () => {
      const computeSectionCmp = workloadTemplateCreateForm.findComponent<DefineComponent>("[aid=data-source-section]");
      const dataSourceId: AssetIdAndKind[] = [
        {
          id: workspaceAssetsEditMockedData.loadedDataSources[0].meta.id,
          kind: workspaceAssetsEditMockedData.loadedDataSources[0].meta.kind,
        },
        {
          id: workspaceAssetsEditMockedData.loadedDataSources[1].meta.id,
          kind: workspaceAssetsEditMockedData.loadedDataSources[1].meta.kind,
        },
      ];
      computeSectionCmp.vm.$emit("datasource-changed", dataSourceId);
      expect(workloadTemplateCreateForm.vm.workloadTemplateModel.assets.datasources).toEqual(dataSourceId);
    });
  });

  describe("Redirecting to the assets creation pages", () => {
    test("component will redirect the user to environment creation page when 'create-new' emitted by the environment section", async () => {
      const envSectionCmp = workloadTemplateCreateForm.findComponent<DefineComponent>("[aid=environment-section]");
      envSectionCmp.vm.$emit("create-new");

      const expectedPush = {
        name: ENVIRONMENT_ROUTE_NAMES.ENVIRONMENT_NEW,
        query: {
          clusterId: null,
          fromPage: WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_NEW,
          projectId: workloadTemplateCreateForm.vm.workloadTemplateModel.projectId,
          scope: Scope.Project,
          departmentId: null,
        },
      };

      expect(workloadTemplateCreateForm.vm.$router).toContainEqual(expectedPush);
    });

    test("component will redirect the user to compute resource creation page when 'create-new' emitted by the compute resource section", async () => {
      const computeSectionCmp = workloadTemplateCreateForm.findComponent<DefineComponent>(
        "[aid=compute-resource-section]",
      );
      computeSectionCmp.vm.$emit("create-new");

      const expectedPush = {
        name: COMPUTE_RESOURCE_ROUTE_NAMES.COMPUTE_RESOURCE_NEW,
        query: {
          clusterId: null,
          fromPage: WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_NEW,
          projectId: workloadTemplateCreateForm.vm.workloadTemplateModel.projectId,
          scope: Scope.Project,
          departmentId: null,
        },
      };

      expect(workloadTemplateCreateForm.vm.$router).toContainEqual(expectedPush);
    });

    test("component will redirect the user to data source creation page when 'create-new' emitted by the data source section", async () => {
      const envSectionCmp = workloadTemplateCreateForm.findComponent<DefineComponent>("[aid=data-source-section]");
      envSectionCmp.vm.$emit("create-new", AssetKind.Pvc);

      const expectedPush = {
        name: DATA_SOURCE_ROUTE_NAMES.DATA_SOURCE_PVC_NEW,
        query: {
          clusterId: null,
          fromPage: WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_NEW,
          projectId: workloadTemplateCreateForm.vm.workloadTemplateModel.projectId,
          scope: Scope.Project,
          departmentId: null,
        },
      };

      expect(workloadTemplateCreateForm.vm.$router).toContainEqual(expectedPush);
    });
  });

  describe("Submitting the form", () => {
    const validModel: IUIWorkloadTemplateCreation = {
      name: "test-ds21",
      scope: Scope.Project,
      namespace: "",
      clusterId: "",
      projectId: 25,
      assets: {
        environment: workspaceAssetsEditMockedData.loadedEnvironments[0].meta.id,
        compute: workspaceAssetsEditMockedData.loadedComputeResources[0].meta.id,
        datasources: [],
        uiVolumes: [],
      },
      specificEnv: {
        args: null,
        environmentVariables: [
          {
            name: "var1",
            value: "var-value",
          },
        ],
        connections: [],
        allowOverQuota: false,
        command: null,
      },
    };

    test("successfuly submitting the form when submit button is clicked and the template model is valid", async () => {
      await workloadTemplateCreateForm.setData({ workloadTemplateModel: validModel });

      const validationSuccess = await workloadTemplateCreateForm.vm.validate();
      expect(validationSuccess).toBe(true);

      await workloadTemplateCreateForm.find("[aid=create-template-btn]").trigger("click");
      await flushPromises();
      expect(workloadTemplateCreateForm.emitted("submit")).toBeTruthy();
    });

    test("preventing form submission and failing validation when submit button is clicked and the template name in the model is NOT valid", async () => {
      const invalidNameModel = { ...validModel };
      invalidNameModel.name = "";
      await workloadTemplateCreateForm.setData({ workloadTemplateModel: invalidNameModel });
      let validationSuccess = await workloadTemplateCreateForm.vm.validate();
      expect(validationSuccess).toBe(false);
      invalidNameModel.name = "A-2";
      await workloadTemplateCreateForm.setData({ workloadTemplateModel: invalidNameModel });
      validationSuccess = await workloadTemplateCreateForm.vm.validate();
      expect(validationSuccess).toBe(false);
      invalidNameModel.name = "a-";
      await workloadTemplateCreateForm.setData({ workloadTemplateModel: invalidNameModel });
      validationSuccess = await workloadTemplateCreateForm.vm.validate();
      expect(validationSuccess).toBe(false);
      invalidNameModel.name = "3";
      await workloadTemplateCreateForm.setData({ workloadTemplateModel: invalidNameModel });
      validationSuccess = await workloadTemplateCreateForm.vm.validate();
      expect(validationSuccess).toBe(false);
      workloadTemplateCreateForm.find("[aid=create-template-btn]").trigger("click");
      await flushPromises();
      expect(workloadTemplateCreateForm.emitted("submit")).toBeFalsy();
    });

    test("preventing form submission and failing validation when submit button is clicked and the template environment in the model is NOT valid", async () => {
      const invalidEnvironmentModel = { ...validModel, assets: { ...validModel.assets, environment: null } };
      await workloadTemplateCreateForm.setData({ workloadTemplateModel: invalidEnvironmentModel });
      const validationSuccess = await workloadTemplateCreateForm.vm.validate();
      expect(validationSuccess).toBe(false);
      workloadTemplateCreateForm.find("[aid=create-template-btn]").trigger("click");
      await flushPromises();
      expect(workloadTemplateCreateForm.emitted("submit")).toBeFalsy();
    });
  });
});
