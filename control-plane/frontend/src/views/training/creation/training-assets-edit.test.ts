import type { VueWrapper } from "@vue/test-utils";
import { shallowMount, flushPromises } from "@vue/test-utils";
import { Quasar } from "quasar";
import { setActivePinia } from "pinia";
import { createTestingPinia } from "@pinia/testing";

import TrainingAssetsEdit from "./training-assets-edit.vue";

// models
import type { TrainingPolicyV2 } from "@/swagger-models/policy-service-client";
import type { CredentialsListResponseEntry } from "@/swagger-models/assets-service-client";
import {
  type AssetIdAndKind,
  type DatasourceRef,
  AssetKind,
  type TrainingV2,
} from "@/swagger-models/assets-service-client";
import type { IUIWorkloadCreation } from "@/models/workload.model";

// services
import { trainingService } from "@/services/control-plane/training.service/training.service";
import { dataSourceService } from "@/services/control-plane/data-source.service/data-source.service";
import { credentialService } from "@/services/control-plane/credential.service/credential.service";

// stores
import { useTrainingStore } from "@/stores/training.store";
import { useEnvironmentStore } from "@/stores/environment.store";
import { useComputeResourceStore } from "@/stores/compute-resource.store";
import { useDataSourceStore } from "@/stores/data-source.store";

// utils
import { deepCopy } from "@/utils/common.util";
import { workloadUtil } from "@/utils/workload.util/workload.util";

import { trainingAssetsEditMockedData, loadedTrainingByIdV2 } from "./training-assets-edit.mock";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
import type { Project } from "@/swagger-models/org-unit-service-client";
import { afterAll, beforeEach, describe, expect, it, vi } from "vitest";
import { setQuasarMock } from "@/utils/unit-test.util/unit-test.util";
import type { ConfigMapListInfo, StorageClassV2 } from "@/swagger-models/k8s-objects-tracker-client";

vi.mock("@/services/control-plane/training.service/training.service");
vi.mock("@/services/control-plane/data-source.service/data-source.service");
vi.mock("@/services/control-plane/credential.service/credential.service");
vi.mock("@/services/control-plane/org-unit.service/org-unit.service");

describe("training-assets-edit.vue", () => {
  const getProjectSpy = vi.spyOn(orgUnitService, "getProject").mockImplementation(async (projectId) => {
    const found = trainingAssetsEditMockedData.loadedProjectsV2.find((project: Project) => {
      return project.id === projectId;
    }) as Project;

    return found;
  });

  describe("load assets", () => {
    let pinia = createTestingPinia({ createSpy: vi.fn });
    let trainingStore = useTrainingStore(pinia);

    let environmentStore = useEnvironmentStore(pinia);
    let computeResourceStore = useComputeResourceStore(pinia);
    let dataSourceStore = useDataSourceStore(pinia);

    beforeEach(() => {
      pinia = createTestingPinia({ createSpy: vi.fn });
      trainingStore = useTrainingStore(pinia);

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
      vi.mocked(environmentStore.loadEnvironments).mockImplementation(async () => {
        return trainingAssetsEditMockedData.loadedEnvironments;
      });
      vi.mocked(computeResourceStore.loadComputeResources).mockImplementation(async () => {
        return trainingAssetsEditMockedData.loadedComputeResources;
      });
      vi.mocked(dataSourceStore.loadDataSources).mockImplementation(async () => {
        return trainingAssetsEditMockedData.loadedDataSources;
      });

      vi.mocked(trainingService.getPolicy).mockImplementation(async (): Promise<TrainingPolicyV2> => {
        return Promise.resolve({});
      });
    });

    afterAll(() => {
      vi.clearAllMocks();
    });

    async function loadShallowMountCmp(query = {}): Promise<VueWrapper<InstanceType<typeof TrainingAssetsEdit>>> {
      const trainingAssetsEditCmp: VueWrapper<InstanceType<typeof TrainingAssetsEdit>> = shallowMount(
        TrainingAssetsEdit,
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

      setQuasarMock(trainingAssetsEditCmp);
      await flushPromises();
      return trainingAssetsEditCmp;
    }

    describe("loadPolicy", () => {
      beforeEach(() => {
        vi.clearAllMocks();
      });

      it("should load policy when select project defined", async () => {
        trainingStore.$patch({
          training: { ...trainingStore.training, projectId: +trainingAssetsEditMockedData.loadedProjectsV2[0].id },
        });
        await loadShallowMountCmp();

        expect(trainingService.getPolicy).toHaveBeenCalled();
      });

      it("should not load policy when selected project is undefined", async () => {
        trainingStore.$patch({ training: {} });
        await loadShallowMountCmp();
        expect(trainingService.getPolicy).not.toHaveBeenCalled();
      });
    });

    describe("loadEnvironments", () => {
      beforeEach(() => {
        vi.clearAllMocks();
      });

      it("should not load environmets when selected project is undefined", async () => {
        trainingStore.$patch({ training: {} });
        const trainingAssetsEditCmp: VueWrapper<InstanceType<typeof TrainingAssetsEdit>> = await loadShallowMountCmp();
        expect(environmentStore.loadEnvironments).not.toHaveBeenCalled();
        expect(trainingAssetsEditCmp.vm.assetsLists.environments.length).toBe(0);
      });

      it("should load environmets when selected project exists", async () => {
        trainingStore.$patch({
          training: { ...trainingStore.training, projectId: +trainingAssetsEditMockedData.loadedProjectsV2[0].id },
        });
        const trainingAssetsEditCmp: VueWrapper<InstanceType<typeof TrainingAssetsEdit>> = await loadShallowMountCmp();
        expect(environmentStore.loadEnvironments).toHaveBeenCalled();
        expect(trainingAssetsEditCmp.vm.assetsLists.environments.length > 0).toBe(true);
      });

      it("should set environment as an empty when createdEntityId is undefined", async () => {
        trainingStore.$patch({
          training: { ...trainingStore.training, projectId: +trainingAssetsEditMockedData.loadedProjectsV2[0].id },
        });
        const trainingAssetsEditCmp: VueWrapper<InstanceType<typeof TrainingAssetsEdit>> = await loadShallowMountCmp();
        expect(environmentStore.loadEnvironments).toHaveBeenCalled();
        expect(trainingAssetsEditCmp.vm.training.assets.environment).toBe("");
      });

      it("should set environment as selected when createdEntityId and kind are defined", async () => {
        trainingStore.$patch({
          training: { ...trainingStore.training, projectId: +trainingAssetsEditMockedData.loadedProjectsV2[0].id },
        });
        const environments = trainingAssetsEditMockedData.loadedEnvironments;
        const createdEnvironmentId = environments[0].meta.id;

        const trainingAssetsEditCmp: VueWrapper<InstanceType<typeof TrainingAssetsEdit>> = await loadShallowMountCmp({
          createdEntityId: createdEnvironmentId,
          kind: AssetKind.Environment,
        });
        expect(trainingAssetsEditCmp.vm.training.assets.environment).toBe(createdEnvironmentId);
      });

      it("should set environment as an empty string when createdEntityId is defined but not complied", async () => {
        trainingStore.$patch({
          training: { ...trainingStore.training, projectId: +trainingAssetsEditMockedData.loadedProjectsV2[0].id },
        });
        const environments = deepCopy(trainingAssetsEditMockedData.loadedEnvironments);
        environments[0].compliance && (environments[0].compliance.compliance = false);
        const createdEnvironmentId = environments[0].meta.id;

        vi.mocked(environmentStore.loadEnvironments).mockImplementation(async () => {
          return environments;
        });

        const trainingAssetsEditCmp: VueWrapper<InstanceType<typeof TrainingAssetsEdit>> = await loadShallowMountCmp({
          createdEntityId: createdEnvironmentId,
          kind: AssetKind.Environment,
        });
        expect(trainingAssetsEditCmp.vm.training.assets.environment).toBe("");
      });
    });

    describe("loadComputeResources", () => {
      beforeEach(() => {
        vi.clearAllMocks();
      });

      it("should not load compute resources when selected project is undefined", async () => {
        trainingStore.$patch({ training: {} });
        const trainingAssetsEditCmp: VueWrapper<InstanceType<typeof TrainingAssetsEdit>> = await loadShallowMountCmp();
        expect(computeResourceStore.loadComputeResources).not.toHaveBeenCalled();
        expect(trainingAssetsEditCmp.vm.assetsLists.datasources.length).toBe(0);
      });

      it("should load compute resources when selected project exists", async () => {
        trainingStore.$patch({
          training: { ...trainingStore.training, projectId: +trainingAssetsEditMockedData.loadedProjectsV2[0].id },
        });

        const trainingAssetsEditCmp: VueWrapper<InstanceType<typeof TrainingAssetsEdit>> = await loadShallowMountCmp();
        expect(computeResourceStore.loadComputeResources).toHaveBeenCalled();
        expect(trainingAssetsEditCmp.vm.assetsLists.computes.length > 0).toBe(true);
      });

      it("should set compute resource as empty string when createdEntityId is undefined", async () => {
        trainingStore.$patch({
          training: { ...trainingStore.training, projectId: +trainingAssetsEditMockedData.loadedProjectsV2[0].id },
        });
        const trainingAssetsEditCmp: VueWrapper<InstanceType<typeof TrainingAssetsEdit>> = await loadShallowMountCmp();
        expect(trainingAssetsEditCmp.vm.training.assets.compute).toBe("");
      });

      it("should set compute resource as selected when createdEntityId and kind are defined", async () => {
        trainingStore.$patch({
          training: { ...trainingStore.training, projectId: +trainingAssetsEditMockedData.loadedProjectsV2[0].id },
        });
        const computeResources = trainingAssetsEditMockedData.loadedComputeResources;
        const createdComputeResourceId = computeResources[0].meta.id;

        const trainingAssetsEditCmp: VueWrapper<InstanceType<typeof TrainingAssetsEdit>> = await loadShallowMountCmp({
          createdEntityId: createdComputeResourceId,
          kind: AssetKind.Compute,
        });
        expect(trainingAssetsEditCmp.vm.training.assets.compute).toBe(createdComputeResourceId);
      });

      it("should set compute resource as an empty string when createdEntityId and kind are defined but not complied", async () => {
        trainingStore.$patch({
          training: { ...trainingStore.training, projectId: +trainingAssetsEditMockedData.loadedProjectsV2[0].id },
        });
        const computeResources = deepCopy(trainingAssetsEditMockedData.loadedComputeResources);
        computeResources[0].compliance && (computeResources[0].compliance.compliance = false);
        const createdComputeResourceId = computeResources[0].meta.id;
        vi.mocked(computeResourceStore.loadComputeResources).mockImplementation(async () => {
          return computeResources;
        });
        const trainingAssetsEditCmp: VueWrapper<InstanceType<typeof TrainingAssetsEdit>> = await loadShallowMountCmp({
          createdEntityId: createdComputeResourceId,
          kind: AssetKind.Compute,
        });
        expect(trainingAssetsEditCmp.vm.training.assets.compute).toBe("");
      });
    });

    describe("loadDataSources", () => {
      beforeEach(() => {
        vi.clearAllMocks();
      });

      it("should not load data sources when selected project is undefined", async () => {
        trainingStore.$patch({ training: {} });
        const trainingAssetsEditCmp: VueWrapper<InstanceType<typeof TrainingAssetsEdit>> = await loadShallowMountCmp();
        expect(dataSourceStore.loadDataSources).not.toHaveBeenCalled();
        expect(trainingAssetsEditCmp.vm.assetsLists.datasources.length).toBe(0);
      });

      it("should load compute resources when selected project exists", async () => {
        trainingStore.$patch({
          training: { ...trainingStore.training, projectId: +trainingAssetsEditMockedData.loadedProjectsV2[0].id },
        });

        const trainingAssetsEditCmp: VueWrapper<InstanceType<typeof TrainingAssetsEdit>> = await loadShallowMountCmp();
        expect(dataSourceStore.loadDataSources).toHaveBeenCalled();
        expect(trainingAssetsEditCmp.vm.assetsLists.datasources.length > 0).toBe(true);
      });

      it("should keep the selected data sources undefiend when createdEntityId is undefined", async () => {
        const trainingAssetsEditCmp: VueWrapper<InstanceType<typeof TrainingAssetsEdit>> = await loadShallowMountCmp();
        expect(trainingAssetsEditCmp.vm.training.assets.datasources).toBeUndefined();
      });

      it("should add data source to selected data sources array when createdEntityId and kind are defined", async () => {
        trainingStore.$patch({
          training: { ...trainingStore.training, projectId: +trainingAssetsEditMockedData.loadedProjectsV2[0].id },
        });
        const dataSources = trainingAssetsEditMockedData.loadedDataSources;
        const createdDataSourceId = dataSources[0].meta.id;
        const kind = dataSources[0].meta.kind;

        const trainingAssetsEditCmp: VueWrapper<InstanceType<typeof TrainingAssetsEdit>> = await loadShallowMountCmp({
          createdEntityId: createdDataSourceId,
          kind,
        });
        expect(dataSourceStore.loadDataSources).toHaveBeenCalled();
        expect(trainingAssetsEditCmp.vm.training.assets.datasources?.length).toBe(1);
        const dataSource: AssetIdAndKind | undefined = trainingAssetsEditCmp.vm.training.assets.datasources
          ? trainingAssetsEditCmp.vm.training.assets.datasources[0]
          : undefined;
        expect(dataSource?.id).toBe(createdDataSourceId);
      });

      it("should unset data source from the array when createdEntityId and kind are defined but not complied", async () => {
        trainingStore.$patch({
          training: { ...trainingStore.training, projectId: +trainingAssetsEditMockedData.loadedProjectsV2[0].id },
        });
        const dataSources = deepCopy(trainingAssetsEditMockedData.loadedDataSources);
        const firstDataSource = dataSources[0];

        if (firstDataSource.compliance) {
          firstDataSource.compliance.compliance = false;
        }
        const createdDataSourceId = firstDataSource.meta.id;
        const kind = firstDataSource.meta.kind;

        vi.mocked(dataSourceStore.loadDataSources).mockImplementation(async () => {
          return dataSources;
        });

        const trainingAssetsEditCmp: VueWrapper<InstanceType<typeof TrainingAssetsEdit>> = await loadShallowMountCmp({
          createdEntityId: createdDataSourceId,
          kind,
        });
        expect(dataSourceStore.loadDataSources).toHaveBeenCalled();
        expect(trainingAssetsEditCmp.vm.training.assets.datasources).toStrictEqual([]);
      });

      it("should add data source to selected data sources array when data source is imposed", async () => {
        trainingStore.$patch({
          training: { ...trainingStore.training, projectId: +trainingAssetsEditMockedData.loadedProjectsV2[0].id },
        });
        const dataSources = deepCopy(trainingAssetsEditMockedData.loadedDataSources);
        const firstDataSource = dataSources[0];

        if (firstDataSource.compliance) {
          firstDataSource.compliance.imposed = true;
        }
        const createdDataSourceId = firstDataSource.meta.id;
        const kind = firstDataSource.meta.kind;

        const trainingAssetsEditCmp: VueWrapper<InstanceType<typeof TrainingAssetsEdit>> = await loadShallowMountCmp({
          createdEntityId: createdDataSourceId,
          kind,
        });
        expect(dataSourceStore.loadDataSources).toHaveBeenCalled();
        expect(trainingAssetsEditCmp.vm.training.assets.datasources?.length).toBe(1);
        const dataSource: AssetIdAndKind | undefined = trainingAssetsEditCmp.vm.training.assets.datasources
          ? trainingAssetsEditCmp.vm.training.assets.datasources[0]
          : undefined;
        expect(dataSource?.id).toBe(createdDataSourceId);
      });

      it("should not add data source to selected data sources array when data source is imposed but not complied", async () => {
        trainingStore.$patch({
          training: { ...trainingStore.training, projectId: +trainingAssetsEditMockedData.loadedProjectsV2[0].id },
        });
        const dataSources = deepCopy(trainingAssetsEditMockedData.loadedDataSources);
        const firstDataSource = dataSources[0];

        if (firstDataSource.compliance) {
          firstDataSource.compliance.compliance = false;
          firstDataSource.compliance.imposed = true;
        }
        const createdDataSourceId = firstDataSource.meta.id;
        const kind = firstDataSource.meta.kind;

        vi.mocked(dataSourceStore.loadDataSources).mockImplementation(async () => {
          return dataSources;
        });

        const trainingAssetsEditCmp: VueWrapper<InstanceType<typeof TrainingAssetsEdit>> = await loadShallowMountCmp({
          createdEntityId: createdDataSourceId,
          kind,
        });
        expect(dataSourceStore.loadDataSources).toHaveBeenCalled();
        expect(trainingAssetsEditCmp.vm.training.assets.datasources?.length).toBe(0);
      });
    });

    describe("loadProjectById", () => {
      beforeEach(() => {
        vi.clearAllMocks();
      });

      it("should load load relevant project by id", async () => {
        trainingStore.$patch({
          training: { ...trainingStore.training, projectId: +trainingAssetsEditMockedData.loadedProjectsV2[0].id },
        });
        await loadShallowMountCmp();
        expect(getProjectSpy).toHaveBeenCalledWith(trainingAssetsEditMockedData.loadedProjectsV2[0].id);
      });
    });
  });

  describe("ui model", () => {
    let pinia = createTestingPinia({ createSpy: vi.fn });
    let trainingStore = useTrainingStore(pinia);
    let environmentStore = useEnvironmentStore(pinia);
    let computeResourceStore = useComputeResourceStore(pinia);
    let dataSourceStore = useDataSourceStore(pinia);

    beforeEach(() => {
      pinia = createTestingPinia({ createSpy: vi.fn });
      trainingStore = useTrainingStore(pinia);
      environmentStore = useEnvironmentStore(pinia);
      computeResourceStore = useComputeResourceStore(pinia);
      dataSourceStore = useDataSourceStore(pinia);
      setActivePinia(pinia);

      vi.mocked(dataSourceService.listStorageClass).mockImplementation(async (): Promise<Array<string>> => {
        return [];
      });

      vi.mocked(credentialService.listCredentials).mockImplementation(
        async (): Promise<Array<CredentialsListResponseEntry>> => {
          return [];
        },
      );

      vi.mocked(trainingService.getPolicy).mockImplementation(async (): Promise<TrainingPolicyV2> => {
        return Promise.resolve({});
      });

      vi.mocked(trainingService.getTrainingMergedWithPolicy).mockImplementation(async (): Promise<TrainingV2> => {
        return Promise.resolve(loadedTrainingByIdV2);
      });
      vi.mocked(environmentStore.loadEnvironments).mockImplementation(async () => {
        return trainingAssetsEditMockedData.loadedEnvironments;
      });
      vi.mocked(computeResourceStore.loadComputeResources).mockImplementation(async () => {
        return trainingAssetsEditMockedData.loadedComputeResources;
      });
      vi.mocked(dataSourceStore.loadDataSources).mockImplementation(async () => {
        return trainingAssetsEditMockedData.loadedDataSources;
      });
    });

    afterAll(() => {
      vi.clearAllMocks();
    });

    async function loadShallowMountCmp(query = {}): Promise<VueWrapper<InstanceType<typeof TrainingAssetsEdit>>> {
      const trainingAssetsEditCmp: VueWrapper<InstanceType<typeof TrainingAssetsEdit>> = shallowMount(
        TrainingAssetsEdit,
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

      setQuasarMock(trainingAssetsEditCmp);
      await flushPromises();
      return trainingAssetsEditCmp;
    }

    describe("getWorkloadUI", () => {
      // beforeEach(() => {
      //   vi.resetAllMocks();
      // });

      it("should set training ui model from an existing training", async () => {
        trainingStore.$patch({ training: {} });

        const expectedWorkloadUI: IUIWorkloadCreation = {
          name: loadedTrainingByIdV2.meta.name,
          namespace: "",
          clusterId: loadedTrainingByIdV2.meta.clusterId,
          projectId: +loadedTrainingByIdV2.meta.projectId,
          assets: {
            environment: loadedTrainingByIdV2.assets.environment.id,
            compute: loadedTrainingByIdV2.assets.compute?.id || null,
            datasources:
              loadedTrainingByIdV2.assets.datasources?.map(
                (ds: DatasourceRef): AssetIdAndKind => ({
                  id: ds.id,
                  kind: ds.kind,
                }),
              ) || [],
            uiVolumes: undefined,
          },
          specificEnv: {
            ...workloadUtil.getUIWorkloadSpecificEnv(loadedTrainingByIdV2.specificRunParams || undefined),
          },
        };

        const trainingAssetsEditCmp: VueWrapper<InstanceType<typeof TrainingAssetsEdit>> = await loadShallowMountCmp({
          fromCopyId: loadedTrainingByIdV2.meta.workloadId,
        });
        expect(trainingService.getTrainingMergedWithPolicy).toHaveBeenCalled();

        expect(trainingAssetsEditCmp.vm.training).toStrictEqual(expectedWorkloadUI);
      });

      it("should keep the training model data from the store and update only the last created asset he has in the url", async () => {
        const trainingUI: IUIWorkloadCreation = {
          name: "some name",
          clusterId: "",
          projectId: +loadedTrainingByIdV2.meta.projectId,
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

        trainingStore.$patch({ training: trainingUI });

        const lastCreatedDataSource = trainingAssetsEditMockedData.loadedDataSources[0];
        const expectedWorkloadUI: IUIWorkloadCreation = {
          name: trainingUI.name,
          namespace: "",
          clusterId: trainingUI.clusterId,
          projectId: trainingUI.projectId,
          assets: {
            environment: trainingUI.assets.environment,
            compute: trainingUI.assets.compute || null,
            datasources: [
              {
                id: lastCreatedDataSource.meta.id,
                kind: lastCreatedDataSource.meta.kind,
              },
            ],
          },
          specificEnv: trainingUI.specificEnv,
        };

        const trainingAssetsEditCmp: VueWrapper<InstanceType<typeof TrainingAssetsEdit>> = await loadShallowMountCmp({
          createdEntityId: lastCreatedDataSource.meta.id,
          kind: lastCreatedDataSource.meta.kind,
        });
        expect(trainingAssetsEditCmp.vm.training).toStrictEqual(expectedWorkloadUI);
      });
    });
  });
});
