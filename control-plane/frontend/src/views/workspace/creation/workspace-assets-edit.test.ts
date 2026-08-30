import type { VueWrapper } from "@vue/test-utils";
import { shallowMount, flushPromises } from "@vue/test-utils";
import { Quasar } from "quasar";
import { setActivePinia } from "pinia";
import { createTestingPinia } from "@pinia/testing";

import WorkspaceAssetsEdit from "./workspace-assets-edit.vue";

// models
import type { WorkspacePolicyV2 } from "@/swagger-models/policy-service-client";
import type { CredentialsListResponseEntry } from "@/swagger-models/assets-service-client";
import {
  type AssetIdAndKind,
  type DatasourceRef,
  AssetKind,
  type WorkspaceV2,
} from "@/swagger-models/assets-service-client";
import type { IUIWorkloadCreation } from "@/models/workload.model";

// services
import { dataSourceService } from "@/services/control-plane/data-source.service/data-source.service";
import { workspaceService } from "@/services/control-plane/workspace.service/workspace.service";
import { credentialService } from "@/services/control-plane/credential.service/credential.service";

// stores
import { useWorkspaceStore } from "@/stores/workspace.store";
import { useEnvironmentStore } from "@/stores/environment.store";
import { useComputeResourceStore } from "@/stores/compute-resource.store";
import { useDataSourceStore } from "@/stores/data-source.store";

// utils
import { deepCopy } from "@/utils/common.util";
import { workloadUtil } from "@/utils/workload.util/workload.util";

import { workspaceAssetsEditMockedData, loadedWorkspaceByIdV2 } from "./workspace-assets-edit.mock";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
import { trainingAssetsEditMockedData as mockData } from "@/views/training/creation/training-assets-edit.mock";
import type { Project } from "@/swagger-models/org-unit-service-client";
import { afterAll, beforeEach, describe, expect, it, vi } from "vitest";
import { setQuasarMock } from "@/utils/unit-test.util/unit-test.util";
import type { ConfigMapListInfo, StorageClassV2 } from "@/swagger-models/k8s-objects-tracker-client";

vi.mock("@/services/control-plane/workspace.service/workspace.service");
vi.mock("@/services/control-plane/data-source.service/data-source.service");
vi.mock("@/services/control-plane/credential.service/credential.service");
vi.mock("@/services/control-plane/org-unit.service/org-unit.service");

describe("workspace-assets-edit.vue", () => {
  const getProjectSpy = vi.spyOn(orgUnitService, "getProject").mockImplementation(async (projectId) => {
    const found = mockData.loadedProjectsV2.find((project: Project) => {
      return project.id === projectId;
    }) as Project;

    return found;
  });

  describe("load assets", () => {
    let pinia = createTestingPinia({ createSpy: vi.fn });
    let workspaceStore = useWorkspaceStore(pinia);
    let environmentStore = useEnvironmentStore(pinia);
    let computeResourceStore = useComputeResourceStore(pinia);
    let dataSourceStore = useDataSourceStore(pinia);

    beforeEach(() => {
      pinia = createTestingPinia({ createSpy: vi.fn });
      workspaceStore = useWorkspaceStore(pinia);
      environmentStore = useEnvironmentStore(pinia);
      computeResourceStore = useComputeResourceStore(pinia);
      dataSourceStore = useDataSourceStore(pinia);
      setActivePinia(pinia);

      vi.mocked(dataSourceService.listStorageClassV2).mockImplementation(async (): Promise<Array<StorageClassV2>> => {
        return [];
      });

      vi.mocked(dataSourceService.listExistingConfigMaps).mockImplementation(
        async (): Promise<Array<ConfigMapListInfo>> => {
          return [];
        },
      );

      vi.mocked(credentialService.listCredentials).mockImplementation(
        async (): Promise<Array<CredentialsListResponseEntry>> => {
          return [];
        },
      );

      vi.mocked(workspaceService.getPolicy).mockImplementation(async (): Promise<WorkspacePolicyV2> => {
        return Promise.resolve({});
      });
    });

    afterAll(() => {
      vi.clearAllMocks();
    });

    async function loadShallowMountCmp(query = {}): Promise<VueWrapper<InstanceType<typeof WorkspaceAssetsEdit>>> {
      const workspaceAssetsEditCmp: VueWrapper<InstanceType<typeof WorkspaceAssetsEdit>> = shallowMount(
        WorkspaceAssetsEdit,
        {
          global: {
            plugins: [Quasar, pinia],
            mocks: {
              $route: {
                query: query,
              },
              $router: {
                push: vi.fn(),
              },
            },
          },
        },
      );

      setQuasarMock(workspaceAssetsEditCmp);
      await flushPromises();
      return workspaceAssetsEditCmp;
    }

    describe("loadPolicy", () => {
      beforeEach(() => {
        vi.clearAllMocks();
      });

      it("should load policy when select project defined", async () => {
        workspaceStore.$patch({
          workspace: { ...workspaceStore.workspace, projectId: +mockData.loadedProjectsV2[0].id },
        });
        await loadShallowMountCmp();
        expect(workspaceService.getPolicy).toHaveBeenCalled();
      });

      it("should not load policy when selected project is undefined", async () => {
        workspaceStore.$patch({ workspace: {} });
        await loadShallowMountCmp();
        expect(workspaceService.getPolicy).not.toHaveBeenCalled();
      });
    });

    describe("loadEnvironments", () => {
      beforeEach(() => {
        vi.clearAllMocks();
      });

      it("should not load environmets when selected project is undefined", async () => {
        workspaceStore.$patch({ workspace: {} });
        const workspaceAssetsEditCmp: VueWrapper<InstanceType<typeof WorkspaceAssetsEdit>> = await loadShallowMountCmp();
        expect(environmentStore.loadEnvironments).not.toHaveBeenCalled();
        expect(workspaceAssetsEditCmp.vm.environments.length).toBe(0);
      });

      it("should load environmets when selected project exists", async () => {
        workspaceStore.$patch({
          workspace: { ...workspaceStore.workspace, projectId: +mockData.loadedProjectsV2[0].id },
        });
        environmentStore.$patch({ environments: workspaceAssetsEditMockedData.loadedEnvironments });
        const workspaceAssetsEditCmp: VueWrapper<InstanceType<typeof WorkspaceAssetsEdit>> = await loadShallowMountCmp();
        expect(environmentStore.loadEnvironments).toHaveBeenCalled();
        expect(workspaceAssetsEditCmp.vm.environments.length > 0).toBe(true);
      });

      it("should set environment as an empty when createdEntityId is undefined", async () => {
        workspaceStore.$patch({
          workspace: { ...workspaceStore.workspace, projectId: +mockData.loadedProjectsV2[0].id },
        });
        const workspaceAssetsEditCmp: VueWrapper<InstanceType<typeof WorkspaceAssetsEdit>> = await loadShallowMountCmp();
        expect(environmentStore.loadEnvironments).toHaveBeenCalled();
        expect(workspaceAssetsEditCmp.vm.workspace.assets.environment).toBe("");
      });

      it("should set environment as selected when createdEntityId and kind are defined", async () => {
        workspaceStore.$patch({
          workspace: { ...workspaceStore.workspace, projectId: +mockData.loadedProjectsV2[0].id },
        });
        const environments = workspaceAssetsEditMockedData.loadedEnvironments;
        const createdEnvironmentId = environments[0].meta.id;
        environmentStore.$patch({ environments });
        const workspaceAssetsEditCmp: VueWrapper<InstanceType<typeof WorkspaceAssetsEdit>> = await loadShallowMountCmp({
          createdEntityId: createdEnvironmentId,
          kind: AssetKind.Environment,
        });
        expect(workspaceAssetsEditCmp.vm.workspace.assets.environment).toBe(createdEnvironmentId);
      });

      it("should set environment as an empty string when createdEntityId and kind are defined but not complied", async () => {
        workspaceStore.$patch({
          workspace: { ...workspaceStore.workspace, projectId: +mockData.loadedProjectsV2[0].id },
        });
        const environments = deepCopy(workspaceAssetsEditMockedData.loadedEnvironments);
        environments[0].compliance && (environments[0].compliance.compliance = false);
        const createdEnvironmentId = environments[0].meta.id;
        environmentStore.$patch({ environments });
        const workspaceAssetsEditCmp: VueWrapper<InstanceType<typeof WorkspaceAssetsEdit>> = await loadShallowMountCmp({
          createdEntityId: createdEnvironmentId,
          kind: AssetKind.Environment,
        });
        expect(workspaceAssetsEditCmp.vm.workspace.assets.environment).toBe("");
      });
    });

    describe("loadComputeResources", () => {
      beforeEach(() => {
        vi.clearAllMocks();
      });

      it("should not load compute resources when selected project is undefined", async () => {
        workspaceStore.$patch({ workspace: {} });
        const workspaceAssetsEditCmp: VueWrapper<InstanceType<typeof WorkspaceAssetsEdit>> = await loadShallowMountCmp();
        expect(computeResourceStore.loadComputeResources).not.toHaveBeenCalled();
        expect(workspaceAssetsEditCmp.vm.dataSources.length).toBe(0);
      });

      it("should load compute resources when selected project exists", async () => {
        workspaceStore.$patch({
          workspace: { ...workspaceStore.workspace, projectId: +mockData.loadedProjectsV2[0].id },
        });
        computeResourceStore.$patch({ computeResources: workspaceAssetsEditMockedData.loadedComputeResources });
        const workspaceAssetsEditCmp: VueWrapper<InstanceType<typeof WorkspaceAssetsEdit>> = await loadShallowMountCmp();
        expect(computeResourceStore.loadComputeResources).toHaveBeenCalled();
        expect(workspaceAssetsEditCmp.vm.computeResources.length > 0).toBe(true);
      });

      it("should set compute resource as empty string when createdEntityId is undefined", async () => {
        workspaceStore.$patch({
          workspace: { ...workspaceStore.workspace, projectId: +mockData.loadedProjectsV2[0].id },
        });
        const workspaceAssetsEditCmp: VueWrapper<InstanceType<typeof WorkspaceAssetsEdit>> = await loadShallowMountCmp();
        expect(workspaceAssetsEditCmp.vm.workspace.assets.compute).toBe("");
      });

      it("should set compute resource as selected when createdEntityId and kind are defined", async () => {
        workspaceStore.$patch({
          workspace: { ...workspaceStore.workspace, projectId: +mockData.loadedProjectsV2[0].id },
        });
        const computeResources = workspaceAssetsEditMockedData.loadedComputeResources;
        const createdComputeResourceId = computeResources[0].meta.id;
        computeResourceStore.$patch({ computeResources });
        const workspaceAssetsEditCmp: VueWrapper<InstanceType<typeof WorkspaceAssetsEdit>> = await loadShallowMountCmp({
          createdEntityId: createdComputeResourceId,
          kind: AssetKind.Compute,
        });
        expect(workspaceAssetsEditCmp.vm.workspace.assets.compute).toBe(createdComputeResourceId);
      });

      it("should set compute resource as an empty string when createdEntityId and kind are defined but not complied", async () => {
        workspaceStore.$patch({
          workspace: { ...workspaceStore.workspace, projectId: +mockData.loadedProjectsV2[0].id },
        });
        const computeResources = deepCopy(workspaceAssetsEditMockedData.loadedComputeResources);
        computeResources[0].compliance && (computeResources[0].compliance.compliance = false);
        const createdComputeResourceId = computeResources[0].meta.id;
        computeResourceStore.$patch({ computeResources });
        const workspaceAssetsEditCmp: VueWrapper<InstanceType<typeof WorkspaceAssetsEdit>> = await loadShallowMountCmp({
          createdEntityId: createdComputeResourceId,
          kind: AssetKind.Compute,
        });
        expect(workspaceAssetsEditCmp.vm.workspace.assets.compute).toBe("");
      });
    });

    describe("loadDataSources", () => {
      beforeEach(() => {
        vi.clearAllMocks();
      });

      it("should not load data sources when selected project is undefined", async () => {
        workspaceStore.$patch({ workspace: {} });
        const workspaceAssetsEditCmp: VueWrapper<InstanceType<typeof WorkspaceAssetsEdit>> = await loadShallowMountCmp();
        expect(dataSourceStore.loadDataSources).not.toHaveBeenCalled();
        expect(workspaceAssetsEditCmp.vm.dataSources.length).toBe(0);
      });

      it("should load compute resources when selected project exists", async () => {
        workspaceStore.$patch({
          workspace: { ...workspaceStore.workspace, projectId: +mockData.loadedProjectsV2[0].id },
        });
        dataSourceStore.$patch({ dataSources: workspaceAssetsEditMockedData.loadedDataSources });
        const workspaceAssetsEditCmp: VueWrapper<InstanceType<typeof WorkspaceAssetsEdit>> = await loadShallowMountCmp();
        expect(dataSourceStore.loadDataSources).toHaveBeenCalled();
        expect(workspaceAssetsEditCmp.vm.dataSources.length > 0).toBe(true);
      });

      it("should keep the selected data sources undefiend when createdEntityId is undefined", async () => {
        const workspaceAssetsEditCmp: VueWrapper<InstanceType<typeof WorkspaceAssetsEdit>> = await loadShallowMountCmp();
        expect(workspaceAssetsEditCmp.vm.workspace.assets.datasources).toBeUndefined();
      });

      it("should add data source to selected data sources array when createdEntityId and kind are defined", async () => {
        workspaceStore.$patch({
          workspace: { ...workspaceStore.workspace, projectId: +mockData.loadedProjectsV2[0].id },
        });
        const dataSources = workspaceAssetsEditMockedData.loadedDataSources;
        const createdDataSourceId = dataSources[0].meta.id;
        const kind = dataSources[0].meta.kind;
        dataSourceStore.$patch({ dataSources });
        const workspaceAssetsEditCmp: VueWrapper<InstanceType<typeof WorkspaceAssetsEdit>> = await loadShallowMountCmp({
          createdEntityId: createdDataSourceId,
          kind,
        });
        expect(dataSourceStore.loadDataSources).toHaveBeenCalled();
        expect(workspaceAssetsEditCmp.vm.workspace.assets.datasources?.length).toBe(1);
        const dataSource: AssetIdAndKind | undefined = workspaceAssetsEditCmp.vm.workspace.assets.datasources
          ? workspaceAssetsEditCmp.vm.workspace.assets.datasources[0]
          : undefined;
        expect(dataSource?.id).toBe(createdDataSourceId);
      });

      it("should unset data source from the array when createdEntityId and kind are defined but not complied", async () => {
        workspaceStore.$patch({
          workspace: { ...workspaceStore.workspace, projectId: +mockData.loadedProjectsV2[0].id },
        });
        const dataSources = deepCopy(workspaceAssetsEditMockedData.loadedDataSources);
        const firstDataSource = dataSources[0];

        if (firstDataSource.compliance) {
          firstDataSource.compliance.compliance = false;
        }
        const createdDataSourceId = firstDataSource.meta.id;
        const kind = firstDataSource.meta.kind;
        dataSourceStore.$patch({ dataSources });
        const workspaceAssetsEditCmp: VueWrapper<InstanceType<typeof WorkspaceAssetsEdit>> = await loadShallowMountCmp({
          createdEntityId: createdDataSourceId,
          kind,
        });
        expect(dataSourceStore.loadDataSources).toHaveBeenCalled();
        expect(workspaceAssetsEditCmp.vm.workspace.assets.datasources).toStrictEqual([]);
      });

      it("should add data source to selected data sources array when data source is imposed", async () => {
        workspaceStore.$patch({
          workspace: { ...workspaceStore.workspace, projectId: +mockData.loadedProjectsV2[0].id },
        });
        const dataSources = deepCopy(workspaceAssetsEditMockedData.loadedDataSources);
        const firstDataSource = dataSources[0];

        if (firstDataSource.compliance) {
          firstDataSource.compliance.imposed = true;
        }
        const createdDataSourceId = firstDataSource.meta.id;
        const kind = firstDataSource.meta.kind;
        dataSourceStore.$patch({ dataSources });
        const workspaceAssetsEditCmp: VueWrapper<InstanceType<typeof WorkspaceAssetsEdit>> = await loadShallowMountCmp({
          createdEntityId: createdDataSourceId,
          kind,
        });
        expect(dataSourceStore.loadDataSources).toHaveBeenCalled();
        expect(workspaceAssetsEditCmp.vm.workspace.assets.datasources?.length).toBe(1);
        const dataSource: AssetIdAndKind | undefined = workspaceAssetsEditCmp.vm.workspace.assets.datasources
          ? workspaceAssetsEditCmp.vm.workspace.assets.datasources[0]
          : undefined;
        expect(dataSource?.id).toBe(createdDataSourceId);
      });

      it("should not add data source to selected data sources array when data source is imposed but not complied", async () => {
        workspaceStore.$patch({
          workspace: { ...workspaceStore.workspace, projectId: +mockData.loadedProjectsV2[0].id },
        });
        const dataSources = deepCopy(workspaceAssetsEditMockedData.loadedDataSources);
        const firstDataSource = dataSources[0];

        if (firstDataSource.compliance) {
          firstDataSource.compliance.compliance = false;
          firstDataSource.compliance.imposed = true;
        }
        const createdDataSourceId = firstDataSource.meta.id;
        const kind = firstDataSource.meta.kind;
        dataSourceStore.$patch({ dataSources });
        const workspaceAssetsEditCmp: VueWrapper<InstanceType<typeof WorkspaceAssetsEdit>> = await loadShallowMountCmp({
          createdEntityId: createdDataSourceId,
          kind,
        });
        expect(dataSourceStore.loadDataSources).toHaveBeenCalled();
        expect(workspaceAssetsEditCmp.vm.workspace.assets.datasources?.length).toBe(0);
      });
    });

    describe("loadProjectById", () => {
      beforeEach(() => {
        vi.clearAllMocks();
      });

      it("should load load relevant project by id", async () => {
        workspaceStore.$patch({
          workspace: { ...workspaceStore.workspace, projectId: +mockData.loadedProjectsV2[0].id },
        });
        await loadShallowMountCmp();
        expect(getProjectSpy).toHaveBeenCalledWith(mockData.loadedProjectsV2[0].id);
      });
    });
  });

  describe("ui model", () => {
    let pinia = createTestingPinia({ createSpy: vi.fn });
    let workspaceStore = useWorkspaceStore(pinia);
    let environmentStore = useEnvironmentStore(pinia);
    let computeResourceStore = useComputeResourceStore(pinia);
    let dataSourceStore = useDataSourceStore(pinia);

    beforeEach(() => {
      pinia = createTestingPinia({ createSpy: vi.fn });
      workspaceStore = useWorkspaceStore(pinia);
      environmentStore = useEnvironmentStore(pinia);
      computeResourceStore = useComputeResourceStore(pinia);
      dataSourceStore = useDataSourceStore(pinia);
      setActivePinia(pinia);

      vi.mocked(dataSourceService.listStorageClass).mockImplementation(async (): Promise<Array<string>> => {
        return [];
      });

      vi.mocked(workspaceService.getPolicy).mockImplementation(async (): Promise<WorkspacePolicyV2> => {
        return Promise.resolve({});
      });

      vi.mocked(workspaceService.getWorkspaceMergedWithPolicy).mockImplementation(async (): Promise<WorkspaceV2> => {
        return Promise.resolve(loadedWorkspaceByIdV2);
      });
    });

    afterAll(() => {
      vi.resetAllMocks();
    });

    async function loadShallowMountCmp(query = {}): Promise<VueWrapper<InstanceType<typeof WorkspaceAssetsEdit>>> {
      const workspaceAssetsEditCmp: VueWrapper<InstanceType<typeof WorkspaceAssetsEdit>> = shallowMount(
        WorkspaceAssetsEdit,
        {
          global: {
            plugins: [Quasar, pinia],
            mocks: {
              $route: {
                query: query,
              },
              $router: {
                push: vi.fn(),
              },
            },
          },
        },
      );

      setQuasarMock(workspaceAssetsEditCmp);
      await flushPromises();
      return workspaceAssetsEditCmp;
    }

    describe("getWorkloadUI", () => {
      beforeEach(() => {
        vi.mocked(credentialService.listCredentials).mockImplementation(
          async (): Promise<Array<CredentialsListResponseEntry>> => {
            return [];
          },
        );
      });

      it("should set workspace ui model from an existing workspace", async () => {
        workspaceStore.$patch({ workspace: {} });
        environmentStore.$patch({ environments: workspaceAssetsEditMockedData.loadedEnvironments });
        computeResourceStore.$patch({ computeResources: workspaceAssetsEditMockedData.loadedComputeResources });
        dataSourceStore.$patch({ dataSources: workspaceAssetsEditMockedData.loadedDataSources });

        const expectedWorkloadUI: IUIWorkloadCreation = {
          name: loadedWorkspaceByIdV2.meta.name,
          namespace: "",
          clusterId: loadedWorkspaceByIdV2.meta.clusterId,
          projectId: +loadedWorkspaceByIdV2.meta.projectId,
          assets: {
            environment: loadedWorkspaceByIdV2.assets.environment.id,
            compute: loadedWorkspaceByIdV2.assets.compute?.id || null,
            datasources:
              loadedWorkspaceByIdV2.assets.datasources?.map(
                (ds: DatasourceRef): AssetIdAndKind => ({
                  id: ds.id,
                  kind: ds.kind,
                }),
              ) || [],
            uiVolumes: undefined,
          },
          specificEnv: {
            ...workloadUtil.getUIWorkloadSpecificEnv(loadedWorkspaceByIdV2.specificRunParams || undefined),
          },
        };

        const workspaceAssetsEditCmp: VueWrapper<InstanceType<typeof WorkspaceAssetsEdit>> = await loadShallowMountCmp({
          fromCopyId: loadedWorkspaceByIdV2.meta.workloadId,
        });
        expect(workspaceService.getWorkspaceMergedWithPolicy).toHaveBeenCalled();
        expect(workspaceAssetsEditCmp.vm.workspace).toStrictEqual(expectedWorkloadUI);
      });

      it("should keep the workspace model data from the store and update only the last created asset he has in the url", async () => {
        const workspaceUI: IUIWorkloadCreation = {
          name: "some name",
          clusterId: "",
          projectId: +mockData.loadedProjectsV2[0].id,
          namespace: "",
          assets: {
            environment: "ad947a50-55c3-4498-892c-2aec5ad1cefe",
            compute: "ed644ac3-1a72-4419-9dc5-e424950d02ef",
          },
          specificEnv: {
            command: "start-notebook.sh",
            args: "--NotebookApp.base_url=/${RUNAI_PROJECT}/${RUNAI_JOB_NAME} --NotebookApp.token=''",
            connections: [{ name: "custom-tool", externalUrl: "http://domain.com" }],
            allowOverQuota: false,
            environmentVariables: [{ name: "LOG_DIR", value: "./" }],
            annotations: [{ name: "ann1", value: "ann1-value" }],
            labels: [{ name: "lbl1", value: "lbl1-value" }],
            backoffLimit: 6,
            nodePools: [],
          },
        };

        workspaceStore.$patch({ workspace: workspaceUI });
        environmentStore.$patch({ environments: workspaceAssetsEditMockedData.loadedEnvironments });
        computeResourceStore.$patch({ computeResources: workspaceAssetsEditMockedData.loadedComputeResources });
        dataSourceStore.$patch({ dataSources: workspaceAssetsEditMockedData.loadedDataSources });

        const lastCreatedDataSource = workspaceAssetsEditMockedData.loadedDataSources[0];
        const expectedWorkloadUI: IUIWorkloadCreation = {
          name: workspaceUI.name,
          namespace: "",
          clusterId: workspaceUI.clusterId,
          projectId: workspaceUI.projectId,
          assets: {
            environment: workspaceUI.assets.environment,
            compute: workspaceUI.assets.compute || null,
            datasources: [
              {
                id: lastCreatedDataSource.meta.id,
                kind: lastCreatedDataSource.meta.kind,
              },
            ],
          },
          specificEnv: workspaceUI.specificEnv,
        };

        const workspaceAssetsEditCmp: VueWrapper<InstanceType<typeof WorkspaceAssetsEdit>> = await loadShallowMountCmp({
          createdEntityId: lastCreatedDataSource.meta.id,
          kind: lastCreatedDataSource.meta.kind,
        });
        expect(workspaceAssetsEditCmp.vm.workspace).toStrictEqual(expectedWorkloadUI);
      });
    });
  });
});
