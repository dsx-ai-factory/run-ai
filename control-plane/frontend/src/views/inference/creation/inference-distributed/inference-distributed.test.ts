/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect, vi, beforeEach } from "vitest";
import { mount, flushPromises, type VueWrapper } from "@vue/test-utils";
import { Quasar } from "quasar";
import { ref } from "vue";
import InferenceDistributed from "./inference-distributed.vue";
import { EDistributedInferenceType } from "./inference-distributed.model";
import { EServerType } from "@/models/inference.model";
import { UidGidSource, WorkloadDesiredPhase } from "@/swagger-models/workloads-client";
import { inferenceService } from "@/services/control-plane/inference.service/inference.service";

// eslint-disable-next-line @typescript-eslint/no-explicit-any
type AnyWrapper = VueWrapper<any>;

/**
 * Unit tests for the continueToWorkers method in inference-distributed.vue
 *
 * These tests mount the actual component and test the continueToWorkers method directly.
 */

// Mock stores
vi.mock("@/stores/app.store", () => ({
  useAppStore: () => ({
    isPageLoading: false,
    setPageLoading: vi.fn(),
    setFallback: vi.fn(),
  }),
}));

const mockSaveDistributedFormState = vi.fn();
const mockGetDistributedState = vi.fn(() => null);

vi.mock("@/stores/inference.store", () => ({
  useInferenceStore: () => ({
    getDistributedState: mockGetDistributedState,
    saveDistributedFormState: mockSaveDistributedFormState,
    clearDistributedFormState: vi.fn(),
    clearInferenceFormUIState: vi.fn(),
    clearPresetFormModel: vi.fn(),
  }),
}));

vi.mock("@/stores/cluster.store", () => ({
  useClusterStore: () => ({
    isClusterVersionSufficient: vi.fn(() => true),
    clusterList: [],
    setSelectedClusterById: vi.fn(),
  }),
}));

vi.mock("@/stores/auth.store", () => ({
  useAuthStore: () => ({ tenant: { id: 1 } }),
}));

vi.mock("@/stores/data-source.store", () => ({
  useDataSourceStore: () => ({}),
}));

vi.mock("@/stores/workload.store", () => ({
  useWorkloadStore: () => ({ setWorkloadAdded: vi.fn() }),
}));

// Mock services
vi.mock("@/services/control-plane/credential.service/credential.service", () => ({
  credentialService: { listSecrets: vi.fn(() => Promise.resolve({ entries: [] })) },
}));

vi.mock("@/services/control-plane/org-unit.service/org-unit.service", () => ({
  orgUnitService: { getProjectById: vi.fn(() => Promise.resolve({})) },
}));

vi.mock("@/services/infra/router.service/router.service", () => ({
  requestToLeave: vi.fn(() => Promise.resolve(true)),
}));

vi.mock("@/services/cluster/workload.service/workload.service", () => ({
  workloadService: { getDefaultPreemptibilityByType: vi.fn() },
}));

vi.mock("@/services/control-plane/inference.service/inference.service", () => ({
  inferenceService: { createDistributedInference: vi.fn() },
}));

// Mock composables
interface MockWorkloadModel {
  clusterId: string;
  projectId: string;
  name: string;
  spec: {
    leader?: Record<string, unknown>;
    worker?: Record<string, unknown>;
    nodePools?: string[];
    servingPort?: { protocol: string; container: number };
  };
}

const mockLeaderModel = ref<MockWorkloadModel>({
  clusterId: "cluster-1",
  projectId: "project-1",
  name: "my-workload",
  spec: {
    leader: {
      compute: { cpuCoreRequest: 4, gpuDevicesRequest: 2 },
      image: "leader-image:latest",
    },
    nodePools: ["pool-1", "pool-2"],
  },
});

const mockWorkerModel = ref<MockWorkloadModel>({
  clusterId: "old-cluster",
  projectId: "old-project",
  name: "old-name",
  spec: {
    worker: {
      compute: { cpuCoreRequest: 1, gpuDevicesRequest: 0 },
      image: "worker-image:v1",
    },
    nodePools: [],
  },
});

const mockCreateSectionsClearedStates = vi.fn();

vi.mock("@/composables/use-hybrid-templates.composable", () => ({
  useHybridTemplates: vi.fn((formType: string) => {
    if (formType === "distributedInferenceLeader") {
      return {
        hybridWorkloadModel: mockLeaderModel,
        selectedTemplates: ref({}),
        createSectionsClearedStates: vi.fn(),
        closeDrawer: vi.fn(),
      };
    }
    return {
      hybridWorkloadModel: mockWorkerModel,
      selectedTemplates: ref({}),
      createSectionsClearedStates: mockCreateSectionsClearedStates,
      closeDrawer: vi.fn(),
    };
  }),
}));

vi.mock("@/composables/helpers/use-hybrid-templates.mapper", () => ({
  EHybridFormTypes: {
    DistributedInferenceLeader: "distributedInferenceLeader",
    DistributedInferenceWorker: "distributedInferenceWorker",
  },
}));

const mockMarkStepComplete = vi.fn();
const mockNextStep = vi.fn();

vi.mock("@/composables/use-wizard-layout.composable", () => ({
  useWizardLayout: () => ({
    steps: ref([]),
    activeStep: ref("Leader"),
    markStepComplete: mockMarkStepComplete,
    markStepIncomplete: vi.fn(),
    nextStep: mockNextStep,
    previousStep: vi.fn(),
    goToStep: vi.fn(),
  }),
}));

vi.mock("@/composables/use-environment-variables-section.composable", () => ({
  useEnvironmentVariablesSection: () => ({
    environmentVariablesSourceValueOptions: ref([]),
    loadSourceValueOptions: vi.fn(() => Promise.resolve()),
  }),
}));

vi.mock("@/composables/use-workload-priorities.composable", () => ({
  useWorkloadPriorities: () => ({
    workloadPriorities: ref([]),
    loadWorkloadPriorities: vi.fn(() => Promise.resolve()),
    loadDefaultWorkloadPriority: vi.fn(() => Promise.resolve()),
    supportPriorityClass: false,
  }),
}));

// Mock vue-router
vi.mock("vue-router", () => ({
  useRoute: () => ({
    query: { clusterId: "cluster-1", projectId: "project-1" },
  }),
}));

describe("inference-distributed.vue", () => {
  describe("continueToWorkers", () => {
    let wrapper: AnyWrapper;

    const createWrapper = () => {
      return mount(InferenceDistributed, {
        global: {
          plugins: [Quasar],
          stubs: {
            "runai-wizard-layout": {
              template: "<div><slot /></div>",
            },
            "notification-alert-bar": true,
            "inference-distributed-leader-form": true,
            "inference-distributed-worker-form": true,
            "inference-distributed-workload-setup": true,
          },
          mocks: {
            $router: { push: vi.fn(), replace: vi.fn() },
            $route: { query: { clusterId: "cluster-1", projectId: "project-1" } },
            $q: { notify: vi.fn() },
          },
        },
      });
    };

    beforeEach(() => {
      vi.clearAllMocks();

      // Reset mock models
      mockLeaderModel.value = {
        clusterId: "cluster-1",
        projectId: "project-1",
        name: "my-workload",
        spec: {
          leader: {
            compute: { cpuCoreRequest: 4, gpuDevicesRequest: 2 },
            image: "leader-image:latest",
          },
          nodePools: ["pool-1", "pool-2"],
        },
      };

      mockWorkerModel.value = {
        clusterId: "old-cluster",
        projectId: "old-project",
        name: "old-name",
        spec: {
          worker: {
            compute: { cpuCoreRequest: 1, gpuDevicesRequest: 0 },
            image: "worker-image:v1",
          },
          nodePools: [],
        },
      };
    });

    it.each([
      [UidGidSource.FromIdpToken, UidGidSource.FromIdpToken],
      [UidGidSource.FromIdpToken, UidGidSource.Custom],
      [UidGidSource.Custom, UidGidSource.FromIdpToken],
    ])("submits leader=%s and worker=%s through the real serializer", async (leaderSource, workerSource) => {
      wrapper = createWrapper();
      await flushPromises();
      const identity = { runAsUid: 2001, runAsGid: 2002, supplementalGroups: "2003,2004" };
      mockLeaderModel.value.spec.servingPort = { protocol: "http", container: 8080 };
      mockLeaderModel.value.spec.leader = {
        ...mockLeaderModel.value.spec.leader,
        security: { uidGidSource: leaderSource, ...identity },
      };
      mockWorkerModel.value.spec.worker = {
        ...mockWorkerModel.value.spec.worker,
        security: { uidGidSource: workerSource, ...identity },
      };
      vi.mocked(inferenceService.createDistributedInference).mockResolvedValue({
        name: "created-inference",
        requestedName: "created-inference",
        workloadId: "inference-id",
        createdBy: "user",
        createdAt: "2026-09-01T00:00:00Z",
        desiredPhase: WorkloadDesiredPhase.Running,
        clusterId: "cluster-1",
        projectId: "project-1",
        spec: {},
      });
      const before = JSON.stringify([mockLeaderModel.value, mockWorkerModel.value]);

      await wrapper.vm.createInference();

      expect(inferenceService.createDistributedInference).toHaveBeenCalledOnce();
      const request = JSON.parse(
        JSON.stringify(vi.mocked(inferenceService.createDistributedInference).mock.calls[0][0]),
      );
      expect(request.spec.leader.security).toEqual({
        uidGidSource: leaderSource,
        ...(leaderSource === UidGidSource.FromIdpToken ? {} : identity),
      });
      expect(request.spec.worker.security).toEqual({
        uidGidSource: workerSource,
        ...(workerSource === UidGidSource.FromIdpToken ? {} : identity),
      });
      expect(JSON.stringify([mockLeaderModel.value, mockWorkerModel.value])).toBe(before);
      wrapper.unmount();
    });

    describe("parseServerType", () => {
      it("accepts a valid server type", () => {
        wrapper = createWrapper();

        expect(wrapper.vm.parseServerType(EServerType.Nim)).toBe(EServerType.Nim);
      });

      it("uses the first value when the query parameter is repeated", () => {
        wrapper = createWrapper();

        expect(wrapper.vm.parseServerType([EServerType.Nim, EServerType.Custom])).toBe(EServerType.Nim);
      });

      it("rejects an unknown server type", () => {
        wrapper = createWrapper();

        expect(wrapper.vm.parseServerType("unknown")).toBeUndefined();
      });
    });

    describe("when policy is null or undefined", () => {
      it("should copy leader data to worker when policy is null", async () => {
        wrapper = createWrapper();
        await flushPromises();

        // Set policy to null
        wrapper.vm.policy = null;

        await wrapper.vm.continueToWorkers();

        expect(wrapper.vm.hasPolicyMismatch).toBe(false);
        expect(mockWorkerModel.value.spec.worker).toEqual(mockLeaderModel.value.spec.leader);
      });

      it("should copy leader data to worker when policy is undefined", async () => {
        wrapper = createWrapper();
        await flushPromises();

        // Set policy to undefined
        wrapper.vm.policy = undefined;

        await wrapper.vm.continueToWorkers();

        expect(wrapper.vm.hasPolicyMismatch).toBe(false);
        expect(mockWorkerModel.value.spec.worker).toEqual(mockLeaderModel.value.spec.leader);
      });
    });

    describe("when no worker policy exists", () => {
      it("should copy leader data to worker when worker policy is missing", async () => {
        wrapper = createWrapper();
        await flushPromises();

        wrapper.vm.policy = {
          master: { defaults: { compute: { cpuCoreRequest: 2 } } },
        };

        await wrapper.vm.continueToWorkers();

        expect(wrapper.vm.hasPolicyMismatch).toBe(false);
        expect(mockWorkerModel.value.spec.worker).toEqual(mockLeaderModel.value.spec.leader);
      });

      it("should copy leader data to worker when worker has empty policy", async () => {
        wrapper = createWrapper();
        await flushPromises();

        wrapper.vm.policy = {
          master: { defaults: { compute: { cpuCoreRequest: 2 } } },
          worker: {},
        };

        await wrapper.vm.continueToWorkers();

        expect(wrapper.vm.hasPolicyMismatch).toBe(false);
        expect(mockWorkerModel.value.spec.worker).toEqual(mockLeaderModel.value.spec.leader);
      });
    });

    describe("when policies are equal", () => {
      it("should copy leader data to worker when defaults and rules are equal", async () => {
        wrapper = createWrapper();
        await flushPromises();

        wrapper.vm.policy = {
          worker: {
            defaults: {
              leader: { compute: { cpuCoreRequest: 2 } },
              worker: { compute: { cpuCoreRequest: 2 } },
            },
            rules: {
              leader: { compute: { cpuCoreRequest: { min: 1, max: 10 } } },
              worker: { compute: { cpuCoreRequest: { min: 1, max: 10 } } },
            },
          },
        };

        await wrapper.vm.continueToWorkers();

        expect(wrapper.vm.hasPolicyMismatch).toBe(false);
        expect(mockWorkerModel.value.spec.worker).toEqual(mockLeaderModel.value.spec.leader);
      });
    });

    describe("when copy flow (isCopyFlow is true)", () => {
      it("should NOT copy leader spec into worker spec (do not overwrite worker with leader)", async () => {
        wrapper = createWrapper();
        await flushPromises();

        wrapper.vm.policy = null;
        wrapper.vm.isCopyFlow = true;

        await wrapper.vm.continueToWorkers();

        expect(mockWorkerModel.value.spec.worker).not.toEqual(mockLeaderModel.value.spec.leader);
      });

      it("should still copy shared fields (clusterId, projectId, name, nodePools) in copy flow", async () => {
        wrapper = createWrapper();
        await flushPromises();

        wrapper.vm.isCopyFlow = true;

        await wrapper.vm.continueToWorkers();

        expect(mockWorkerModel.value.clusterId).toBe("cluster-1");
        expect(mockWorkerModel.value.projectId).toBe("project-1");
        expect(mockWorkerModel.value.name).toBe("my-workload");
        expect(mockWorkerModel.value.spec.nodePools).toEqual(["pool-1", "pool-2"]);
      });
    });

    describe("when policies are NOT equal", () => {
      it("should NOT copy spec.leader to spec.worker and set hasPolicyMismatch to true when defaults differ", async () => {
        wrapper = createWrapper();
        await flushPromises();

        wrapper.vm.policy = {
          worker: {
            defaults: {
              leader: { compute: { cpuCoreRequest: 2 } },
              worker: { compute: { cpuCoreRequest: 8 } },
            },
            rules: { leader: {}, worker: {} },
          },
        };

        await wrapper.vm.continueToWorkers();

        expect(wrapper.vm.hasPolicyMismatch).toBe(true);
        // spec.worker should NOT be copied from spec.leader
        expect(mockWorkerModel.value.spec.worker).not.toEqual(mockLeaderModel.value.spec.leader);
        // But shared fields ARE copied
        expect(mockWorkerModel.value.clusterId).toBe(mockLeaderModel.value.clusterId);
        expect(mockWorkerModel.value.spec.nodePools).toEqual(mockLeaderModel.value.spec.nodePools);
      });

      it("should NOT copy spec.leader to spec.worker and set hasPolicyMismatch to true when rules differ", async () => {
        wrapper = createWrapper();
        await flushPromises();

        wrapper.vm.policy = {
          worker: {
            defaults: { leader: {}, worker: {} },
            rules: {
              leader: { compute: { cpuCoreRequest: { min: 1, max: 10 } } },
              worker: { compute: { cpuCoreRequest: { min: 5, max: 50 } } },
            },
          },
        };

        await wrapper.vm.continueToWorkers();

        expect(wrapper.vm.hasPolicyMismatch).toBe(true);
        expect(mockWorkerModel.value.spec.worker).not.toEqual(mockLeaderModel.value.spec.leader);
      });
    });

    describe("shared fields always copied", () => {
      it("should always copy clusterId from leader to worker", async () => {
        wrapper = createWrapper();
        await flushPromises();

        wrapper.vm.policy = {
          worker: {
            defaults: {
              leader: { compute: { cpuCoreRequest: 2 } },
              worker: { compute: { cpuCoreRequest: 8 } },
            },
          },
        };

        await wrapper.vm.continueToWorkers();

        expect(mockWorkerModel.value.clusterId).toBe("cluster-1");
      });

      it("should always copy projectId from leader to worker", async () => {
        wrapper = createWrapper();
        await flushPromises();

        wrapper.vm.policy = {
          worker: {
            defaults: {
              leader: { compute: { cpuCoreRequest: 2 } },
              worker: { compute: { cpuCoreRequest: 8 } },
            },
          },
        };

        await wrapper.vm.continueToWorkers();

        expect(mockWorkerModel.value.projectId).toBe("project-1");
      });

      it("should always copy name from leader to worker", async () => {
        wrapper = createWrapper();
        await flushPromises();

        wrapper.vm.policy = {
          worker: {
            defaults: {
              leader: { compute: { cpuCoreRequest: 2 } },
              worker: { compute: { cpuCoreRequest: 8 } },
            },
          },
        };

        await wrapper.vm.continueToWorkers();

        expect(mockWorkerModel.value.name).toBe("my-workload");
      });

      it("should always copy nodePools from leader to worker", async () => {
        wrapper = createWrapper();
        await flushPromises();

        wrapper.vm.policy = {
          worker: {
            defaults: {
              leader: { compute: { cpuCoreRequest: 2 } },
              worker: { compute: { cpuCoreRequest: 8 } },
            },
          },
        };

        await wrapper.vm.continueToWorkers();

        expect(mockWorkerModel.value.spec.nodePools).toEqual(["pool-1", "pool-2"]);
      });
    });

    describe("side effects", () => {
      it("should always mark Leader step as complete", async () => {
        wrapper = createWrapper();
        await flushPromises();

        wrapper.vm.policy = null;

        await wrapper.vm.continueToWorkers();

        expect(mockMarkStepComplete).toHaveBeenCalledWith(EDistributedInferenceType.Leader);
      });

      it("should always call nextStep on wizard layout", async () => {
        wrapper = createWrapper();
        await flushPromises();

        wrapper.vm.policy = null;

        await wrapper.vm.continueToWorkers();

        expect(mockNextStep).toHaveBeenCalledTimes(1);
      });

      it("should call all side effects even when policy mismatch occurs", async () => {
        wrapper = createWrapper();
        await flushPromises();

        wrapper.vm.policy = {
          worker: {
            defaults: {
              leader: { compute: { cpuCoreRequest: 2 } },
              worker: { compute: { cpuCoreRequest: 8 } },
            },
          },
        };

        await wrapper.vm.continueToWorkers();

        expect(mockMarkStepComplete).toHaveBeenCalledWith(EDistributedInferenceType.Leader);
        expect(mockNextStep).toHaveBeenCalledTimes(1);
      });

      it("should save form state with isCopyFlow so copy flow is persisted across navigation", async () => {
        wrapper = createWrapper();
        await flushPromises();

        wrapper.vm.isCopyFlow = true;
        wrapper.vm.policy = null;

        await wrapper.vm.continueToWorkers();

        expect(mockSaveDistributedFormState).toHaveBeenCalledWith(expect.objectContaining({ isCopyFlow: true }));
      });
    });

    describe("deep copy verification", () => {
      it("should create a deep copy of leader data (not reference)", async () => {
        wrapper = createWrapper();
        await flushPromises();

        wrapper.vm.policy = null;

        await wrapper.vm.continueToWorkers();

        // Modify the original leader spec after copying
        mockLeaderModel.value.spec.leader = { compute: { cpuCoreRequest: 999 } };

        // Worker spec should not be affected
        expect(mockWorkerModel.value.spec.worker).not.toEqual(mockLeaderModel.value.spec.leader);
      });

      it("should copy entire workload model structure when shouldCopyLeaderToWorker is true", async () => {
        mockLeaderModel.value = {
          clusterId: "cluster-1",
          projectId: "project-1",
          name: "my-workload",
          spec: {
            leader: {
              compute: { cpuCoreRequest: 4, gpuDevicesRequest: 2 },
              image: "test-image",
              environment: { variables: [{ name: "VAR1", value: "value1" }] },
            },
            nodePools: ["pool-1"],
          },
        };

        wrapper = createWrapper();
        await flushPromises();

        wrapper.vm.policy = null;

        await wrapper.vm.continueToWorkers();

        expect(mockWorkerModel.value.spec.worker).toEqual({
          compute: { cpuCoreRequest: 4, gpuDevicesRequest: 2 },
          image: "test-image",
          environment: { variables: [{ name: "VAR1", value: "value1" }] },
        });
      });

      it("should always copy model structure even when policies mismatch", async () => {
        wrapper = createWrapper();
        await flushPromises();

        wrapper.vm.policy = {
          worker: {
            defaults: {
              leader: { compute: { cpuCoreRequest: 2 } },
              worker: { compute: { cpuCoreRequest: 8 } },
            },
          },
        };

        await wrapper.vm.continueToWorkers();

        // Model structure is copied (includes spec.leader from leader)
        expect(mockWorkerModel.value.spec.leader).toEqual(mockLeaderModel.value.spec.leader);
      });
    });
  });
});
