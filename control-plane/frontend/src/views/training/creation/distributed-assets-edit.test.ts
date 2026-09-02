/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { VueWrapper } from "@vue/test-utils";
import { shallowMount, flushPromises } from "@vue/test-utils";
import { Quasar } from "quasar";
import { setActivePinia } from "pinia";
import { createTestingPinia } from "@pinia/testing";

import DistributedAssetsEdit from "./distributed-assets-edit.vue";

// models
import type { TrainingPolicyV2 } from "@/swagger-models/policy-service-client";
import type { CredentialsListResponseEntry } from "@/swagger-models/assets-service-client";

// services
import { trainingService } from "@/services/control-plane/training.service/training.service";
import { dataSourceService } from "@/services/control-plane/data-source.service/data-source.service";
import { credentialService } from "@/services/control-plane/credential.service/credential.service";

// stores
import { useTrainingStore } from "@/stores/training.store";
import { useEnvironmentStore } from "@/stores/environment.store";
import { useComputeResourceStore } from "@/stores/compute-resource.store";
import { useDataSourceStore } from "@/stores/data-source.store";

import { trainingAssetsEditMockedData } from "./training-assets-edit.mock";
import { distributedAssetsEditMockedData } from "./distributed-assets-edit.mock";
import { afterAll, beforeEach, describe, expect, it, vi } from "vitest";
import { setQuasarMock } from "@/utils/unit-test.util/unit-test.util";
import type { ConfigMapListInfo, StorageClassV2 } from "@/swagger-models/k8s-objects-tracker-client";

vi.mock("@/services/control-plane/training.service/training.service");
vi.mock("@/services/control-plane/data-source.service/data-source.service");
vi.mock("@/services/control-plane/credential.service/credential.service");
vi.mock("@/services/control-plane/org-unit.service/org-unit.service");

describe("distributed-assets-edit.vue", () => {
  let pinia = createTestingPinia({ createSpy: vi.fn });
  let trainingStore = useTrainingStore(pinia);

  let environmentStore = useEnvironmentStore(pinia);
  let computeResourceStore = useComputeResourceStore(pinia);
  let dataSourceStore = useDataSourceStore(pinia);

  async function loadShallowMountCmp(query = {}): Promise<VueWrapper<InstanceType<typeof DistributedAssetsEdit>>> {
    const distributedAssetsEditCmp: VueWrapper<InstanceType<typeof DistributedAssetsEdit>> = shallowMount(
      DistributedAssetsEdit,
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

    setQuasarMock(distributedAssetsEditCmp);
    await flushPromises();
    return distributedAssetsEditCmp;
  }

  describe("form behaviour", () => {
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

    it("should set training.enableEditingMaster to TRUE when master policy is defined", async () => {
      trainingStore.$patch({ training: distributedAssetsEditMockedData.trainingUIMock });
      trainingStore.$patch({
        loadedTrainingModelWithDefaultsAndPolicy:
          distributedAssetsEditMockedData.loadedTrainingModelWithDefaultsAndPolicyMock,
      });

      const distributedAssetsEditCmp: VueWrapper<InstanceType<typeof DistributedAssetsEdit>> =
        await loadShallowMountCmp();

      distributedAssetsEditCmp.vm.onContinueStep();
      await distributedAssetsEditCmp.vm.$nextTick();

      expect(distributedAssetsEditCmp.vm.isMasterPolicyDefined).toBe(true);
      expect(distributedAssetsEditCmp.vm.training.enableEditingMaster).toBe(true);
    });
  });
});
