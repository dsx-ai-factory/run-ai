/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { VueWrapper } from "@vue/test-utils";
import { flushPromises, shallowMount } from "@vue/test-utils";
import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import DistributedHybrid from "@/views/training/distributed-hybrid/distributed-hybrid.vue";
import { useAppStore } from "@/stores/app.store";
import { useAuthStore } from "@/stores/auth.store";
import { useTrainingStore } from "@/stores/training.store";
import { useClusterStore } from "@/stores/cluster.store";
import { workloadService } from "@/services/cluster/workload.service/workload.service";
import { credentialService } from "@/services/control-plane/credential.service/credential.service";
import { createTestingPinia } from "@pinia/testing";
import { mockQuasar } from "@/utils/unit-test.util/unit-test.util";
import { Quasar } from "quasar";
import { dataSourceService } from "@/services/control-plane/data-source.service/data-source.service";
import type { ConfigMapListInfo, SecretListInfo, StorageClassV2 } from "@/swagger-models/k8s-objects-tracker-client";
import type { IUISuperWorkloadCreationRequest } from "@/models/workload.model";
import { setActivePinia } from "pinia";
import type { UnifiedWorkload, WorkloadTypeConfig } from "@/swagger-models/workloads-client";
import type { CredentialsListResponseEntry, DistributedFramework } from "@/swagger-models/assets-service-client";
import { workloadHybridUtil } from "@/utils/workload-hybrid.util/workload-hybrid.util";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
import type { Project } from "@/swagger-models/org-unit-service-client";
import { trainingService } from "@/services/control-plane/training.service/training.service";

vi.mock("@/services/control-plane/data-source.service/data-source.service");
vi.mock("@/services/control-plane/asset.service/asset.service");
vi.mock("@/services/control-plane/training.service/training.service");
vi.mock("@/services/control-plane/credential.service/credential.service");
vi.mock("@/services/cluster/workload.service/workload.service");
vi.mock("@/services/control-plane/policy.service/policy.service");
vi.mock("@/services/control-plane/org-unit.service/org-unit.service");

vi.mock("vue-router", () => ({
  useRoute: vi.fn(() => ({
    query: {
      clusterId: "test-cluster-id",
    },
  })),
  useRouter: vi.fn(() => ({
    push: vi.fn(),
  })),
}));

describe("distributed-hybrid.vue", () => {
  let wrapper: VueWrapper<InstanceType<typeof DistributedHybrid>>;

  let pinia = createTestingPinia({ createSpy: vi.fn, stubActions: false });
  let trainingStore = useTrainingStore(pinia);

  setActivePinia(pinia);

  const defaultQuery: Record<string, string> = {
    projectId: "123",
    trainingName: "tr-1",
    distFramework: "PyTorch",
  };

  function factory(query = defaultQuery, replaceQuery = true): void {
    wrapper = shallowMount(DistributedHybrid, {
      global: {
        plugins: [Quasar, pinia],
        mocks: {
          $route: {
            query: replaceQuery ? query : { ...defaultQuery, ...query },
          },
          $q: mockQuasar,
          $router: {
            push: vi.fn(),
          },
        },
      },
    });
  }

  beforeEach(() => {
    pinia = createTestingPinia({ createSpy: vi.fn, stubActions: false });
    trainingStore = useTrainingStore(pinia);
    useClusterStore(pinia);
    useAppStore(pinia);
    useAuthStore(pinia);

    setActivePinia(pinia);

    factory();
  });

  vi.mocked(workloadService.getUnifiedWorkload).mockImplementation(async (): Promise<UnifiedWorkload> => {
    return {
      projectInfo: {
        name: "proj-3",
      },
    };
  });
  vi.mocked(workloadService.getDefaultWorkloadPriorityByType).mockImplementation(
    async (): Promise<WorkloadTypeConfig> => {
      return {
        categoryId: "708a14a6-15c9-4672-b0a5-0513b79a900a",
        priorityId: "d27f5c47-b24f-4903-9235-967d2bcae3e4",
        preemptibility: "preemptible",
        name: "DistributedWorkload",
        group: "run.ai",
        categoryName: "Train",
        priorityName: "low",
        updatedAt: "0001-01-01T00:00:00Z",
        updatedBy: "system",
      };
    },
  );
  vi.mocked(orgUnitService.getProject).mockImplementation(async (): Promise<Project> => {
    return {
      description: "",
      nodeTypes: {},
      id: "4500001",
      name: "p-1",
      clusterId: "8e19deb0-9786-4022-a924-5cf2cb16a417",
      parentId: "4500000",
      requestedNamespace: "",
      enforceRunaiScheduler: true,
      status: {
        namespace: "runai-p-1",
        phase: "Ready",
        phaseMessage: "",
        nodePoolQuotaStatuses: [
          {
            allocated: {},
            allocatedNonPreemptible: {},
            requested: {},
            nodePoolName: "default",
            nodePoolId: "100000",
          },
        ],
        quotaStatus: {
          allocated: {},
          allocatedNonPreemptible: {},
          requested: {},
        },
        additionalStatusData: {
          conditions: [
            {
              type: "NamespaceReady",
              status: "True",
              lastTransitionTime: "2025-07-09T08:49:19Z",
              lastProbeTime: null,
            },
            {
              type: "RoleBindingsReady",
              status: "True",
              lastTransitionTime: "2025-07-09T08:49:19Z",
              lastProbeTime: null,
            },
            {
              type: "ServiceMonitorReady",
              status: "True",
              lastTransitionTime: "2025-07-09T08:49:19Z",
              lastProbeTime: null,
            },
            {
              type: "SecretsReady",
              status: "True",
              lastTransitionTime: "2025-07-09T08:49:19Z",
              lastProbeTime: null,
            },
            {
              type: "PvcsReady",
              status: "True",
              lastTransitionTime: "2025-07-09T08:49:19Z",
              lastProbeTime: null,
            },
            {
              type: "ConfigMapsReady",
              status: "True",
              lastTransitionTime: "2025-07-09T08:49:19Z",
              lastProbeTime: null,
            },
            {
              type: "QueuesReady",
              status: "True",
              lastTransitionTime: "2025-07-09T08:49:19Z",
              lastProbeTime: null,
            },
          ],
        },
      },
      totalResources: {
        gpuQuota: 12,
      },
      createdAt: "2025-07-09T08:49:13.633408Z",
      updatedAt: "2025-07-09T08:49:13.633408Z",
      createdBy: "test@run.ai",
      parent: {
        id: "4500000",
        name: "default",
      },
      effective: {
        defaultNodePools: ["default"],
        nodeTypes: {},
      },
      overtimeData: {},
      resources: [
        {
          nodePool: {
            id: "100000",
            name: "default",
          },
          gpu: {
            deserved: 12,
            limit: -1,
            overQuotaWeight: 12,
          },
          priority: "MediumLow",
        },
      ],
    };
  });

  vi.mocked(dataSourceService.listStorageClassV2).mockImplementation(async (): Promise<StorageClassV2[]> => {
    return [];
  });

  vi.mocked(dataSourceService.listExistingConfigMaps).mockImplementation(async (): Promise<ConfigMapListInfo[]> => {
    return [];
  });
  vi.mocked(credentialService.listCredentials).mockImplementation(async (): Promise<CredentialsListResponseEntry[]> => {
    return [];
  });
  vi.mocked(credentialService.listExistingSecrets).mockImplementation(async (): Promise<SecretListInfo[]> => {
    return [];
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it("should render correctly", () => {
    expect(wrapper.html()).toContain("distributed-hybrid");
  });

  describe("life cycle - created", () => {
    it("should initiate details from query", async () => {
      expect(wrapper.vm.workerContext.hybridWorkloadModel.value.name).toBe("tr-1");
      expect(wrapper.vm.workerContext.hybridWorkloadModel.value.projectId).toBe("123");
      expect(wrapper.vm.workerContext.hybridWorkloadModel.value.spec.distributedFramework).toBe("PyTorch");
    });

    it("should initiate details from formState if it exists", async () => {
      const patchedState = {
        workloadModel: {
          projectId: "333",
          name: "from-store-tr",
          clusterId: "12312312",
          spec: { distributedFramework: "TensorFlow" as DistributedFramework },
        } as IUISuperWorkloadCreationRequest,
      };
      trainingStore.$patch({
        distributedWorkerFlatFormState: patchedState,
      });

      factory();

      expect(wrapper.vm.workerContext.hybridWorkloadModel.value.name).toEqual(patchedState.workloadModel.name);
      expect(wrapper.vm.workerContext.hybridWorkloadModel.value.projectId).toEqual(patchedState.workloadModel.projectId);
      expect(wrapper.vm.workerContext.hybridWorkloadModel.value.spec.distributedFramework).toEqual(
        patchedState.workloadModel.spec.distributedFramework,
      );
      expect(wrapper.vm.workerContext.hybridWorkloadModel.value.clusterId).toEqual(patchedState.workloadModel.clusterId);
    });

    it("should save the copied workloadId in the store if it was passed in query", async () => {
      factory({ fromCopyId: "some-id" }, false);
      expect(trainingStore.distFlatFromContext.fromCopyId).toBe("some-id");
    });

    it("should initiate the wizard steps correctly", async () => {
      expect(wrapper.vm.wizardLayout.steps.value).toEqual([
        { name: "Worker", title: "Workers", done: false },
        { name: "Master", title: "Master", done: false },
        { name: "Setup", title: "Workload setup", done: false },
      ]);
    });

    it("should set the current step properly if it was passed in the query", async () => {
      factory({ workloadFormName: "Master" }, false);
      await flushPromises();
      expect(wrapper.vm.wizardLayout.activeStep.value).toBe("Master");
    });

    it("should invoke both load secrets and sourceValuesOptions upon creation", async () => {
      factory();

      const secretsSpy = vi.spyOn(wrapper.vm, "loadSecrets");
      const sourceValuesOptionsSpy = vi.spyOn(wrapper.vm, "loadSourceValueOptions");

      await flushPromises();

      expect(secretsSpy).toHaveBeenCalled();
      expect(sourceValuesOptionsSpy).toHaveBeenCalled();
    });
  });

  describe("methods", () => {
    it("onSubmit - should omit copied preemptibility when cluster does not support it", async () => {
      const clusterStore = useClusterStore(pinia);
      vi.mocked(clusterStore.isClusterVersionSufficient).mockReturnValue(false);
      vi.mocked(trainingService.createDistributedFlat).mockResolvedValue({ workloadId: "created-id" } as never);

      wrapper.vm.workerContext.hybridWorkloadModel.value = workloadHybridUtil.setDistributedUIDefaults(
        {
          name: "copy-dist-test",
          projectId: "123",
          clusterId: "unsupported-cluster",
          masterSpec: {},
          spec: {
            image: "alpine",
            command: "sleep infinity",
            nodePools: ["default"],
            priorityClass: "low",
            preemptibility: "preemptible",
            distributedFramework: "PyTorch" as DistributedFramework,
          },
        } as IUISuperWorkloadCreationRequest,
        false,
        null,
        "2.23",
      );
      wrapper.vm.masterContext.hybridWorkloadModel.value.spec = workloadHybridUtil.setDistributedMasterSpecUIDefaults(
        {
          image: "alpine",
          command: "sleep infinity",
          nodePools: ["default"],
          priorityClass: "low",
          preemptibility: "preemptible",
        },
        false,
        null,
        "2.23",
      );

      await wrapper.vm.onSubmit();

      const request = vi.mocked(trainingService.createDistributedFlat).mock.calls[0][0];
      expect(request.spec).toBeDefined();
      expect(request.spec?.priorityClass).toBe("low");
      expect(request.spec?.preemptibility).toBeUndefined();
      expect(request.masterSpec?.preemptibility).toBeUndefined();
    });

    it("handleNextStep - should invoke this.Continue if masterSpec exists", async () => {
      const onContinueSpy = vi.spyOn(wrapper.vm, "onContinue");
      wrapper.vm.handleNextStep();
      expect(onContinueSpy).toHaveBeenCalled();
    });

    it("handleNextStep - should invoke onContinueToSetup if masterSpec does not exist", async () => {
      factory({ noMaster: "true" }, false);
      const onContinueToSetupSpy = vi.spyOn(wrapper.vm, "onContinueToSetup");
      wrapper.vm.handleNextStep();
      expect(onContinueToSetupSpy).toHaveBeenCalled();
    });

    it("initMasterModel - should set masterSpecSameAsWorker to undefined", async () => {
      wrapper.vm.initMasterModel();
      expect(wrapper.vm.workerContext.hybridWorkloadModel.value.masterSpecSameAsWorker).toBe(undefined);
    });

    it("updateMasterSpec - should update only the master spec", async () => {
      const updatedMasterSpec = {
        terminationGracePeriodSeconds: 10,
        autoDeletionTimeAfterCompletionSeconds: 10,
        backoffLimit: 10,
      };

      wrapper.vm.updateMasterSpec(updatedMasterSpec);

      // Verify master spec is updated
      expect(wrapper.vm.masterContext.hybridWorkloadModel.value.spec?.terminationGracePeriodSeconds).toBe(
        updatedMasterSpec.terminationGracePeriodSeconds,
      );
      expect(wrapper.vm.masterContext.hybridWorkloadModel.value.spec?.autoDeletionTimeAfterCompletionSeconds).toBe(
        updatedMasterSpec.autoDeletionTimeAfterCompletionSeconds,
      );
      expect(wrapper.vm.masterContext.hybridWorkloadModel.value.spec?.backoffLimit).toBe(updatedMasterSpec.backoffLimit);
    });

    it("redirectToPrevRoute - invoking should clear all form states", async () => {
      const patchedState = {
        workloadModel: {
          projectId: "333",
          name: "from-store-tr",
          clusterId: "12312312",
          spec: { distributedFramework: "TensorFlow" as DistributedFramework },
        } as IUISuperWorkloadCreationRequest,
      };

      trainingStore.$patch({
        distributedMasterFlatFormState: patchedState,
        distributedWorkerFlatFormState: patchedState,
        distFlatFromContext: { fromCopyId: "some-id", templateId: null },
      });

      const clearFormStatesSpy = vi.spyOn(wrapper.vm, "clearFormState");
      wrapper.vm.redirectToPrevRoute();
      expect(clearFormStatesSpy).toHaveBeenCalled();
      expect(trainingStore.distFlatFromContext.fromCopyId).toBe(null);
      expect(trainingStore.distributedMasterFlatFormState).toBe(null);
      expect(trainingStore.distributedWorkerFlatFormState).toBe(null);
    });

    it("onBack - should handle back button correctly", () => {
      wrapper.vm.onBack();
      expect(wrapper.vm.wizardLayout.activeStep.value).toBe("Worker");
      expect(wrapper.vm.wizardLayout.steps.value[0].done).toBe(false);
    });
  });
});
