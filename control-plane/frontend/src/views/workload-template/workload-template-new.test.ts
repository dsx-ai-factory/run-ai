/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import WorkloadTemplateNew from "./workload-template-new.vue";
import { useWorkloadTemplateStore } from "@/stores/workload-template.store";
import { createTestingPinia } from "@pinia/testing";
import type { VueWrapper } from "@vue/test-utils";
import { flushPromises, mount } from "@vue/test-utils";
import { setActivePinia } from "pinia";
import { Quasar } from "quasar";
import { describe, type MockedFunction } from "vitest";
import {
  workspaceAssetsEditMockedData,
  loadedWorkloadTemplateById,
} from "../workspace/creation/workspace-assets-edit.mock";
import { dataSourceService } from "@/services/control-plane/data-source.service/data-source.service";
import { AssetKind, type PVCAsset } from "@/swagger-models/assets-service-client";
import { useClusterStore } from "@/stores/cluster.store";
import { useAuthStore } from "@/stores/auth.store";
import { orgTreeService } from "@/services/control-plane/rbac/org-tree.service/org-tree.service";
import { orgTreeServiceMock } from "@/services/control-plane/rbac/org-tree.service/org-tree.service.mock";
import { useDataSourceStore } from "@/stores/data-source.store";
import { token } from "./workload-template-new.mock";

vi.mock("@/services/control-plane/data-source.service/data-source.service");
vi.mock("@/services/control-plane/rbac/org-tree.service/org-tree.service");

describe("workload-template-new.vue", () => {
  let workloadTemplateNew: VueWrapper<InstanceType<typeof WorkloadTemplateNew>>;
  const pinia = createTestingPinia({ createSpy: vi.fn });
  setActivePinia(pinia);

  const mockGetPvcById = dataSourceService.getPvcById as MockedFunction<typeof dataSourceService.getPvcById>;
  mockGetPvcById.mockImplementation(() => {
    return Promise.resolve({} as PVCAsset);
  });

  const mockListStorageClasses = dataSourceService.listStorageClass as MockedFunction<
    typeof dataSourceService.listStorageClass
  >;
  mockListStorageClasses.mockResolvedValue([]);

  const mockListDataSources = dataSourceService.list as MockedFunction<typeof dataSourceService.list>;
  mockListDataSources.mockResolvedValue(workspaceAssetsEditMockedData.loadedDataSources);

  const mockGetOrgTree = orgTreeService.getOrgTree as MockedFunction<typeof orgTreeService.getOrgTree>;
  mockGetOrgTree.mockResolvedValue(orgTreeServiceMock.orgTree);

  let authStore = useAuthStore(pinia);
  let workloadTemplateStore = useWorkloadTemplateStore(pinia);
  let clusterStore = useClusterStore(pinia);
  let dataSourceStore = useDataSourceStore(pinia);

  beforeEach(() => {
    authStore = useAuthStore(pinia);
    clusterStore = useClusterStore(pinia);
    workloadTemplateStore = useWorkloadTemplateStore(pinia);
    dataSourceStore = useDataSourceStore(pinia);

    authStore.$patch({ accessToken: token });
    clusterStore.$patch({ currentCluster: { uuid: "34234-322-sdfs323" } });
    dataSourceStore.$patch({
      dataSources: workspaceAssetsEditMockedData.loadedDataSources,
    });

    workloadTemplateNew = mount(WorkloadTemplateNew, {
      global: {
        plugins: [Quasar],
        mocks: {
          $route: {
            params: {},
            query: {},
          },
        },
      },
    });
    workloadTemplateNew.vm.$q.notify = vi.fn(); // disabling notify to prevent it from failing tests
  });

  test("component is loading properly", () => {
    expect(workloadTemplateNew.html()).toContain("workload-template-new");
  });

  describe("workload-template-new.vue testing component behavior with $route.query data", () => {
    test("component will set the template's environment correctly if the component is loaded with createdEntityId and kind in $route.query", async () => {
      workloadTemplateNew = mount(WorkloadTemplateNew, {
        global: {
          plugins: [Quasar, pinia],
          mocks: {
            $route: {
              query: {
                createdEntityId: workspaceAssetsEditMockedData.loadedEnvironments[0].meta.id,
                kind: AssetKind.Environment,
              },
              params: {},
            },
            $router: [],
          },
        },
      });
      await flushPromises();
      expect(workloadTemplateNew.vm.workloadTemplate?.spec.assets.environment).toBe(
        workspaceAssetsEditMockedData.loadedEnvironments[0].meta.id,
      );
    });

    test("component will set the template's compute correctly if the component is loaded with createdEntityId and kind in $route.query", async () => {
      workloadTemplateNew = mount(WorkloadTemplateNew, {
        global: {
          plugins: [Quasar, pinia],
          mocks: {
            $route: {
              query: {
                createdEntityId: workspaceAssetsEditMockedData.loadedComputeResources[0].meta.id,
                kind: AssetKind.Compute,
              },
              params: {},
            },
            $router: [],
          },
        },
      });
      await flushPromises();
      expect(workloadTemplateNew.vm.workloadTemplate?.spec.assets.compute).toBe(
        workspaceAssetsEditMockedData.loadedComputeResources[0].meta.id,
      );
    });
    test("component will set the template's datasources correctly if the component is loaded with createdEntityId and kind in $route.query", async () => {
      const dataSourceId: string = workspaceAssetsEditMockedData.loadedDataSources[0].meta.id;
      const dataSourceKind: string = workspaceAssetsEditMockedData.loadedDataSources[0].meta.kind;
      workloadTemplateNew = mount(WorkloadTemplateNew, {
        global: {
          plugins: [Quasar, pinia],
          mocks: {
            $route: {
              query: {
                createdEntityId: dataSourceId,
                kind: dataSourceKind,
              },
              params: {},
            },
            $router: [],
          },
        },
      });
      await flushPromises();
      expect(workloadTemplateNew.vm.workloadTemplate?.spec.assets.datasources).toContainEqual({
        id: dataSourceId,
        kind: dataSourceKind,
      });
    });
  });

  test("component will set the template's properties correctly if the component is loaded with fromCopyId in $route.query", async () => {
    vi.mocked(workloadTemplateStore.loadById).mockImplementation(() => {
      workloadTemplateStore.selectedWorkloadTemplate = loadedWorkloadTemplateById;
      return Promise.resolve(loadedWorkloadTemplateById);
    });

    workloadTemplateNew = mount(WorkloadTemplateNew, {
      global: {
        plugins: [Quasar, pinia],
        mocks: {
          $route: {
            query: {
              fromCopyId: loadedWorkloadTemplateById.meta.id,
            },
            params: {},
          },
          $router: [],
        },
      },
    });
    await flushPromises();

    expect(workloadTemplateNew.vm.workloadTemplate?.spec.assets.compute).toEqual(
      loadedWorkloadTemplateById.spec.assets.compute?.id,
    );
    expect(workloadTemplateNew.vm.workloadTemplate?.spec.assets.environment).toEqual(
      loadedWorkloadTemplateById.spec.assets.environment.id,
    );
    expect(workloadTemplateNew.vm.workloadTemplate?.spec.specificEnv).toEqual(
      loadedWorkloadTemplateById.spec.specificEnv,
    );
    expect(workloadTemplateNew.vm.workloadTemplate?.meta.name).toBe(loadedWorkloadTemplateById.meta.name);
  });
});
